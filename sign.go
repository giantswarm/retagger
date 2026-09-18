package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// `retagger sign <path>` signs the images a skopeo YAML file governs at their
// destination registry with cosign keyless signing (Sigstore: Fulcio
// certificate, Rekor transparency log, signature stored next to the image as an
// OCI referrer). The signing identity is the CircleCI job's OIDC identity, the
// same identity shape the architect orb signs Giant Swarm-built images with, so
// one keyless attestor admits mirrored and built images alike:
//
//	issuer:  https://oidc.circleci.com
//	subject: https://circleci.com/api/v2/projects/<project-id>/pipeline-definitions/<pipeline-definition-id>
//
// The command is idempotent: a digest that already verifies against this job's
// identity is skipped, so it runs after every copy (over the `.filtered` file)
// and, given the unfiltered file, as a sign-only pass over every tag the file
// governs.
const (
	// oidcIssuer is the issuer of the CircleCI OIDC tokens Fulcio federates with
	// (the root issuer, not the organisation-scoped one).
	oidcIssuer = "https://oidc.circleci.com"
	// sigstoreAudience is the audience Fulcio expects; the auto-injected
	// CIRCLE_OIDC_TOKEN_V2 carries a different one and is rejected.
	sigstoreAudience = "sigstore"
	// identityFormat is Fulcio's subject-alternative-name template for CircleCI
	// tokens (config/identity/config.yaml in sigstore/fulcio), filled from the
	// two token claims below.
	identityFormat            = "https://circleci.com/api/v2/projects/%s/pipeline-definitions/%s"
	claimProjectID            = "oidc.circleci.com/project-id"
	claimPipelineDefinitionID = "oidc.circleci.com/pipeline-definition-id"
	// oidcTokenMaxAge bounds how long one minted token is reused. CircleCI OIDC
	// tokens are valid for one hour; cosign requests a fresh Fulcio certificate
	// for every signature, so a long sign-only pass needs the token renewed.
	oidcTokenMaxAge = 30 * time.Minute
	// cosignAttempts is how often a cosign call is tried when it fails for a
	// transient reason (see transientErrorPattern). The backoff doubles from
	// five seconds, so six attempts wait up to 155 s in total: long enough for
	// the registry's per-minute throttling to pass when many jobs sign at once.
	cosignAttempts = 6
)

var (
	// transientErrorPattern matches transport failures and HTTP statuses of the
	// public Sigstore services and the registry that a retry with backoff
	// resolves. Auth and configuration errors (401, 403, 404) are not listed so
	// that a genuine misconfiguration fails fast. The architect orb's set plus
	// the registry's own throttling token: Azure Container Registry answers a
	// burst of referrers lookups with the error code TOOMANYREQUESTS.
	transientErrorPattern = regexp.MustCompile(`INTERNAL_ERROR|stream error|GOAWAY|connection reset|connection refused|broken pipe` +
		`|unexpected EOF|TLS handshake timeout|i/o timeout|context deadline exceeded|Client\.Timeout` +
		`|Internal Server Error|Bad Gateway|Service Unavailable|Gateway Timeout|Too Many Requests|TOOMANYREQUESTS|Request Timeout` +
		`|[Ss]tatus:? (408|425|429|5[0-9][0-9])`)
	// noSignaturePattern is cosign's answer when the registry lists no signature
	// for a digest. Right after a signature was pushed it is a read-after-write
	// race (the registry's referrers index catches up within seconds, and a
	// throttled listing looks the same), so the verify that follows a sign
	// retries it; before a sign it is the expected answer for an unsigned digest.
	noSignaturePattern = regexp.MustCompile(`no signatures found|no matching signatures`)
	// rekorDuplicatePattern matches Rekor's rejection of an entry it already
	// holds (HTTP 409): the signature was recorded, only the client's retry of
	// the upload was refused.
	rekorDuplicatePattern = regexp.MustCompile(`createLogEntryConflict|an equivalent entry already exists`)
	// manifestUnknownPattern matches a registry answering that a tag or
	// repository does not exist: the copy did not happen, there is nothing to sign.
	manifestUnknownPattern = regexp.MustCompile(`manifest unknown|name unknown|repository name not known|MANIFEST_UNKNOWN|NAME_UNKNOWN`)

	// backoff waits before the next attempt; a variable so tests run without sleeping.
	backoff = func(attempt int) { time.Sleep(time.Duration(5*(1<<(attempt-1))) * time.Second) }
)

// oidcTokenSource mints sigstore-audience OIDC tokens through the circleci CLI
// the job runs with (`circleci run oidc get`) and hands out one token for at
// most oidcTokenMaxAge before minting a new one.
type oidcTokenSource struct {
	mu       sync.Mutex
	token    string
	mintedAt time.Time
}

func (s *oidcTokenSource) Token() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.token != "" && time.Since(s.mintedAt) < oidcTokenMaxAge {
		return s.token, nil
	}
	c, stdout, stderr := command("circleci", "run", "oidc", "get", "--claims", fmt.Sprintf(`{"aud": %q}`, sigstoreAudience), "--root-issuer")
	if err := c.Run(); err != nil {
		return "", fmt.Errorf("error minting the OIDC token with 'circleci run oidc get': %w\n%s", err, stderr.String())
	}
	token := strings.TrimSpace(stdout.String())
	if token == "" {
		return "", errors.New("'circleci run oidc get' returned an empty token")
	}
	s.token = token
	s.mintedAt = time.Now()
	return token, nil
}

// identityFromToken derives the certificate identity Fulcio issues for a
// CircleCI OIDC token from the token's claims. The token is decoded, not
// verified: the identity is only used to recognise this job's own signatures,
// Fulcio does the verification when it issues the certificate.
func identityFromToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("token is not a JWT: expected 3 segments, got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
	if err != nil {
		return "", fmt.Errorf("error decoding the token payload: %w", err)
	}
	claims := map[string]any{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("error unmarshaling the token claims: %w", err)
	}
	projectID, _ := claims[claimProjectID].(string)
	pipelineDefinitionID, _ := claims[claimPipelineDefinitionID].(string)
	if projectID == "" || pipelineDefinitionID == "" {
		return "", fmt.Errorf("token carries no %q and %q claims; is this a CircleCI OIDC token?", claimProjectID, claimPipelineDefinitionID)
	}
	return fmt.Sprintf(identityFormat, projectID, pipelineDefinitionID), nil
}

// manifestDigest resolves a tag at the registry to the digest of the manifest
// (the index, for a multi-architecture image) the registry serves for it. The
// digest is computed over the raw manifest bytes, which is how the registry
// computes it. errManifestUnknown reports a tag or repository that does not exist.
var errManifestUnknown = errors.New("manifest unknown")

func manifestDigest(image string) (string, error) {
	c, stdout, stderr := command("skopeo", "inspect", "--raw", "--retry-times", "3", dockerTransport+image)
	if err := c.Run(); err != nil {
		if manifestUnknownPattern.MatchString(stderr.String()) {
			return "", errManifestUnknown
		}
		return "", fmt.Errorf("error inspecting %q: %w\n%s", image, err, stderr.String())
	}
	sum := sha256.Sum256(stdout.Bytes())
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// cosign runs one cosign call, retrying failures that match transient (in
// addition to transientErrorPattern) with backoff. A Rekor duplicate-entry
// rejection counts as success: the entry exists.
func cosign(env []string, transient *regexp.Regexp, args ...string) error {
	var lastErr error
	for attempt := 1; attempt <= cosignAttempts; attempt++ {
		c, _, stderr := command("cosign", args...)
		c.Env = append(os.Environ(), env...)
		err := c.Run()
		if err == nil {
			return nil
		}
		out := stderr.String()
		if rekorDuplicatePattern.MatchString(out) {
			logrus.Debugf("cosign %s: Rekor already holds this entry", strings.Join(args, " "))
			return nil
		}
		lastErr = fmt.Errorf("cosign %s: %w\n%s", strings.Join(args, " "), err, out)
		retry := transientErrorPattern.MatchString(out) || (transient != nil && transient.MatchString(out))
		if !retry || attempt == cosignAttempts {
			return lastErr
		}
		logrus.Warnf("cosign %s: transient error (attempt %d/%d), retrying", args[0], attempt, cosignAttempts)
		backoff(attempt)
	}
	return lastErr
}

// cosignVerify checks that ref carries a valid keyless signature issued to
// identity. justSigned makes a "no signatures found" answer a retried
// read-after-write race instead of the final word.
func cosignVerify(ref, identity string, justSigned bool) error {
	var transient *regexp.Regexp
	if justSigned {
		transient = noSignaturePattern
	}
	return cosign(nil, transient, "verify", "--certificate-oidc-issuer", oidcIssuer, "--certificate-identity", identity, ref)
}

// cosignSign signs ref with cosign keyless signing under the OIDC token; cosign
// reads the token from SIGSTORE_ID_TOKEN.
func cosignSign(ref string, tokens *oidcTokenSource) error {
	token, err := tokens.Token()
	if err != nil {
		return err
	}
	return cosign([]string{"SIGSTORE_ID_TOKEN=" + token}, nil, "sign", "--yes", ref)
}

// signOutcome is what happened to one image tag.
type signOutcome int

const (
	outcomeSigned signOutcome = iota
	outcomeSkipped
	outcomeMissing
	outcomeFailed
)

// signTag signs one destination tag unless its digest already verifies against
// identity. It returns outcomeMissing when the tag is not at the registry (the
// copy did not happen) and outcomeFailed with the error otherwise.
func signTag(image, identity string, tokens *oidcTokenSource) (signOutcome, error) {
	digest, err := manifestDigest(image)
	if errors.Is(err, errManifestUnknown) {
		return outcomeMissing, nil
	}
	if err != nil {
		return outcomeFailed, err
	}
	repository := image[:strings.LastIndex(image, ":")]
	ref := repository + "@" + digest

	if err := cosignVerify(ref, identity, false); err == nil {
		return outcomeSkipped, nil
	}
	if err := cosignSign(ref, tokens); err != nil {
		return outcomeFailed, err
	}
	if err := cosignVerify(ref, identity, true); err != nil {
		return outcomeFailed, fmt.Errorf("signed %q but the signature does not verify: %w", ref, err)
	}
	return outcomeSigned, nil
}

// commandSign is invoked when `retagger sign <path>` is called.
//
// It lists the tags the skopeo YAML file governs (the way `retagger filter`
// does), resolves each at the destination registry and signs every digest that
// does not yet carry this job's signature. Tags absent from the registry are
// reported and skipped; a signing or verification failure fails the command
// once every tag has been attempted.
func commandSign(filePath string) {
	if filePath == "" {
		logrus.Fatal("You need to specify filepath: 'retagger sign <path>'")
	}
	if flagSignWorkers < 1 {
		logrus.Fatalf("%q cannot be lower than 1", "sign-workers")
	}
	logger := logrus.WithField("file", filePath)

	tokens := &oidcTokenSource{}
	token, err := tokens.Token()
	if err != nil {
		logger.Fatal(err)
	}
	identity, err := identityFromToken(token)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("Signing as %q (issuer %q)", identity, oidcIssuer)

	tagsPerImage, err := skopeoSyncTags(filePath)
	if err != nil {
		logger.Fatal(err)
	}
	var images []string
	for image, tags := range tagsPerImage {
		for _, tag := range tags {
			images = append(images, fmt.Sprintf("%s/%s:%s", flagRegistry, imageBaseName(image), tag))
		}
	}
	sort.Strings(images)
	logger.Infof("Found %d image tags to sign at %q", len(images), flagRegistry)

	var (
		counts [4]int
		mu     sync.Mutex
		wg     sync.WaitGroup
		queue  = make(chan string)
	)
	for w := 0; w < flagSignWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for image := range queue {
				outcome, err := signTag(image, identity, tokens)
				mu.Lock()
				counts[outcome]++
				mu.Unlock()
				switch outcome {
				case outcomeSigned:
					logger.Infof("signed %q", image)
				case outcomeSkipped:
					logger.Debugf("already signed %q", image)
				case outcomeMissing:
					logger.Warnf("not at the registry, nothing to sign: %q", image)
				case outcomeFailed:
					logger.Errorf("failed to sign %q: %v", image, err)
				}
			}
		}()
	}
	for _, image := range images {
		queue <- image
	}
	close(queue)
	wg.Wait()

	logger.Infof("Done: %d signed, %d already signed, %d not at the registry, %d failed",
		counts[outcomeSigned], counts[outcomeSkipped], counts[outcomeMissing], counts[outcomeFailed])
	if counts[outcomeFailed] > 0 {
		logger.Fatalf("Signing ended with %d errors", counts[outcomeFailed])
	}
}
