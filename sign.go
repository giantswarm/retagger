package main

import (
	"context"
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

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sirupsen/logrus"
)

// `retagger sign <path>` signs the images a skopeo YAML file or a renamed-images
// file governs at their destination registry with cosign keyless signing
// (Sigstore: Fulcio certificate, Rekor transparency log, signature stored next
// to the image as an OCI referrer). The signing identity is the CircleCI job's
// OIDC identity, the same identity shape the architect orb signs Giant
// Swarm-built images with, so one keyless attestor admits mirrored and built
// images alike:
//
//	issuer:  https://oidc.circleci.com
//	subject: https://circleci.com/api/v2/projects/<project-id>/pipeline-definitions/<pipeline-definition-id>
//
// Signing is idempotent: a digest that already verifies against this job's
// identity is skipped, and cosign's answer that a digest is unsigned is trusted
// only once the registry confirms it lists no signature bundle for it (see
// existingSignature). The skopeo path signs after every copy (over the
// `.filtered` file) and, given the unfiltered file, as a sign-only pass over
// every tag the file governs; `retagger run --sign` signs each renamed tag right
// after its copy, and `retagger sign` over a renamed-images file is the pass over
// every tag its rules govern, sharded by executor like `retagger run`.
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
	// signatureBundleArtifactType is the artifact type cosign v3 stores a
	// signature under: a Sigstore bundle, an OCI referrer of the signed digest.
	signatureBundleArtifactType = "application/vnd.dev.sigstore.bundle.v0.3+json"
	// registryTimeout bounds one referrers listing at the registry.
	registryTimeout = 2 * time.Minute
)

var (
	// transientErrorPattern matches transport failures and HTTP statuses of the
	// public Sigstore services and the registry that a retry with backoff
	// resolves. Auth and configuration errors (401, 403, 404) are not listed so
	// that a genuine misconfiguration fails fast. The architect orb's set plus
	// the registry's own throttling token (Azure Container Registry answers a
	// burst of referrers lookups with the error code TOOMANYREQUESTS) and the
	// race between concurrent cosign processes refreshing the shared TUF cache
	// ("failed to persist metadata"), which the next attempt does not hit.
	transientErrorPattern = regexp.MustCompile(`INTERNAL_ERROR|stream error|GOAWAY|connection reset|connection refused|broken pipe` +
		`|unexpected EOF|TLS handshake timeout|i/o timeout|context deadline exceeded|Client\.Timeout` +
		`|Internal Server Error|Bad Gateway|Service Unavailable|Gateway Timeout|Too Many Requests|TOOMANYREQUESTS|Request Timeout` +
		`|[Ss]tatus:? (408|425|429|5[0-9][0-9])` +
		`|failed to persist metadata`)
	// unsignedPattern is cosign's answer when it finds no signature for a digest.
	// Before a sign it is the expected answer for an unsigned digest, but also
	// what cosign says when its referrers listing failed (see existingSignature);
	// right after a sign it is a read-after-write race the verify waits out.
	unsignedPattern = regexp.MustCompile(`no signatures found`)
	// foreignSignaturePattern is cosign's answer when the digest carries
	// signatures, none of them issued to the expected identity.
	foreignSignaturePattern = regexp.MustCompile(`no matching (signatures|attestations)`)
	// noSignaturePattern is either answer: what the verify right after a sign
	// retries until the registry lists the new signature (its referrers index
	// catches up within seconds).
	noSignaturePattern = regexp.MustCompile(unsignedPattern.String() + `|` + foreignSignaturePattern.String())
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

// manifestDigest resolves a tag at the registry to the digest the registry
// addresses the manifest by: the sha256 of the raw manifest bytes, which is the
// registry's content address for schema 2 and OCI manifests, image indexes and
// OCI artifacts alike. (`skopeo inspect` without --raw picks a platform instance
// instead, which fails on an index without linux/amd64 and on an artifact such
// as a falco rules file.) A Docker schema 1 manifest is the exception, and it
// cannot be signed at all: cosign refuses its media type, so it is reported as
// errUnsignable. errManifestUnknown reports a tag or repository that does not exist.
var (
	errManifestUnknown = errors.New("manifest unknown")
	errUnsignable      = errors.New("schema 1 manifest, which cosign cannot sign")
)

func manifestDigest(image string) (string, error) {
	c, stdout, stderr := command("skopeo", "inspect", "--raw", "--retry-times", "3", dockerTransport+image)
	if err := c.Run(); err != nil {
		if manifestUnknownPattern.MatchString(stderr.String()) {
			return "", errManifestUnknown
		}
		return "", fmt.Errorf("error inspecting %q: %w\n%s", image, err, stderr.String())
	}
	var m struct {
		SchemaVersion int `json:"schemaVersion"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &m); err != nil {
		return "", fmt.Errorf("error parsing the manifest of %q: %w", image, err)
	}
	if m.SchemaVersion == 1 {
		return "", errUnsignable
	}
	sum := sha256.Sum256(stdout.Bytes())
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// retrying runs attempt until it succeeds, cosignAttempts times with backoff,
// as long as its error is transient: transientErrorPattern or, when given, also
// transient. what names the operation in the retry log line.
func retrying(what string, transient *regexp.Regexp, attempt func() error) error {
	var err error
	for n := 1; n <= cosignAttempts; n++ {
		if err = attempt(); err == nil {
			return nil
		}
		msg := err.Error()
		retry := transientErrorPattern.MatchString(msg) || (transient != nil && transient.MatchString(msg))
		if !retry || n == cosignAttempts {
			return err
		}
		logrus.Warnf("%s: transient error (attempt %d/%d), retrying", what, n, cosignAttempts)
		backoff(n)
	}
	return err
}

// cosign runs one cosign call, retrying failures that match transient (in
// addition to transientErrorPattern) with backoff. A Rekor duplicate-entry
// rejection counts as success: the entry exists.
func cosign(env []string, transient *regexp.Regexp, args ...string) error {
	return retrying("cosign "+args[0], transient, func() error {
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
		return fmt.Errorf("cosign %s: %w\n%s", strings.Join(args, " "), err, out)
	})
}

// hasSignatureBundle asks the registry whether it lists a signature bundle
// among the referrers of ref, a digest reference. The listing uses the
// credentials cosign uses (DOCKER_CONFIG) and retries the registry's
// throttling the way a cosign call does; a listing that fails for good is an
// error, never "unsigned".
func hasSignatureBundle(ref string) (bool, error) {
	digest, err := name.NewDigest(ref)
	if err != nil {
		return false, err
	}
	var found bool
	err = retrying("referrers of "+ref, nil, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), registryTimeout)
		defer cancel()
		index, err := remote.Referrers(digest, remote.WithContext(ctx), remote.WithAuthFromKeychain(authn.DefaultKeychain),
			remote.WithFilter("artifactType", signatureBundleArtifactType))
		if err != nil {
			return fmt.Errorf("listing the referrers of %q: %w", ref, err)
		}
		manifest, err := index.IndexManifest()
		if err != nil {
			return fmt.Errorf("reading the referrers of %q: %w", ref, err)
		}
		found = len(manifest.Manifests) > 0
		return nil
	})
	return found, err
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
	outcomeUnsignable
	outcomeFailed
)

// existingSignature reports whether ref already carries a signature issued to
// identity. false with a nil error means the digest is unsigned, or signed by
// another identity only, and is to be signed; an error means the question could
// not be answered, which is not a reason to sign.
//
// cosign's "no signatures found" is trusted only once the registry confirms it
// lists no signature bundle for the digest: cosign detects the bundle format
// with a referrers listing and, when that listing fails (the registry
// throttling the per-identity referrers lookups of a busy pass), falls back
// silently to the legacy signature tag and reports the digest unsigned. Signing
// it again then adds a duplicate signature and Rekor entry.
func existingSignature(ref, identity string) (bool, error) {
	for attempt := 1; ; attempt++ {
		err := cosignVerify(ref, identity, false)
		if err == nil {
			return true, nil
		}
		logrus.Debugf("%s: not verified: %v", ref, err)
		switch {
		case foreignSignaturePattern.MatchString(err.Error()):
			return false, nil
		case !unsignedPattern.MatchString(err.Error()):
			return false, fmt.Errorf("checking %q for an existing signature: %w", ref, err)
		}
		signed, err := hasSignatureBundle(ref)
		if err != nil || !signed {
			return false, err
		}
		if attempt == cosignAttempts {
			return false, fmt.Errorf("the registry lists a signature bundle for %q but cosign verify found none in %d attempts", ref, attempt)
		}
		logrus.Warnf("cosign verify %s: the registry lists a signature bundle but cosign found none, a throttled lookup (attempt %d/%d), retrying", ref, attempt, cosignAttempts)
		backoff(attempt)
	}
}

// signTag signs one destination tag unless its digest already verifies against
// identity. It returns outcomeMissing when the tag is not at the registry (the
// copy did not happen), outcomeUnsignable for a manifest cosign cannot sign, and
// outcomeFailed with the error otherwise.
func signTag(image, identity string, tokens *oidcTokenSource) (signOutcome, error) {
	digest, err := manifestDigest(image)
	switch {
	case errors.Is(err, errManifestUnknown):
		return outcomeMissing, nil
	case errors.Is(err, errUnsignable):
		return outcomeUnsignable, nil
	case err != nil:
		return outcomeFailed, err
	}
	repository := image[:strings.LastIndex(image, ":")]
	ref := repository + "@" + digest

	switch signed, err := existingSignature(ref, identity); {
	case err != nil:
		return outcomeFailed, err
	case signed:
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

// signer signs destination references under the job's identity and keeps the
// tally of outcomes for the summary line. It is safe for concurrent use.
type signer struct {
	identity string
	tokens   *oidcTokenSource
	logger   *logrus.Entry
	mu       sync.Mutex
	counts   [5]int
}

// newSigner mints the job's OIDC token, derives the identity Fulcio issues for
// it and logs that identity once.
func newSigner(logger *logrus.Entry) (*signer, error) {
	tokens := &oidcTokenSource{}
	token, err := tokens.Token()
	if err != nil {
		return nil, err
	}
	identity, err := identityFromToken(token)
	if err != nil {
		return nil, err
	}
	logger.Infof("Signing as %q (issuer %q)", identity, oidcIssuer)
	return &signer{identity: identity, tokens: tokens, logger: logger}, nil
}

// Sign signs one destination reference (registry/namespace/name:tag) unless it
// already carries the job's signature, logs the outcome and counts it.
func (s *signer) Sign(image string) signOutcome {
	outcome, err := signTag(image, s.identity, s.tokens)
	s.mu.Lock()
	s.counts[outcome]++
	s.mu.Unlock()
	switch outcome {
	case outcomeSigned:
		s.logger.Infof("signed %q", image)
	case outcomeSkipped:
		s.logger.Debugf("already signed %q", image)
	case outcomeMissing:
		s.logger.Warnf("not at the registry, nothing to sign: %q", image)
	case outcomeUnsignable:
		s.logger.Warnf("cannot be signed (%v): %q", errUnsignable, image)
	case outcomeFailed:
		s.logger.Errorf("failed to sign %q: %v", image, err)
	}
	return outcome
}

// SignAll signs images with workers concurrent signers.
func (s *signer) SignAll(images []string, workers int) {
	var wg sync.WaitGroup
	queue := make(chan string)
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for image := range queue {
				s.Sign(image)
			}
		}()
	}
	for _, image := range images {
		queue <- image
	}
	close(queue)
	wg.Wait()
}

// Failed is the number of references whose signing or verification failed.
func (s *signer) Failed() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.counts[outcomeFailed]
}

// Summary is the one-line tally of the outcomes so far.
func (s *signer) Summary() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fmt.Sprintf("Done: %d signed, %d already signed, %d not at the registry, %d unsignable (schema 1), %d failed",
		s.counts[outcomeSigned], s.counts[outcomeSkipped], s.counts[outcomeMissing], s.counts[outcomeUnsignable], s.counts[outcomeFailed])
}

// governedReferences lists the references a file governs at flagRegistry,
// sorted: for a skopeo YAML file the tags `skopeo sync --dry-run` reports, the
// way `retagger filter` does; for a renamed-images file the destinations its
// rules select from the source registries, the way `retagger run` does, for
// this executor's share of the rules. A rule whose source cannot be listed is
// reported in the returned error while the other rules' references are still
// returned.
func governedReferences(filePath string) ([]string, error) {
	renamed, err := isRenamedImagesFile(filePath)
	if err != nil {
		return nil, err
	}
	var references []string
	var errs []error
	if renamed {
		renamedImages, err := loadRenamedImages(filePath)
		if err != nil {
			return nil, err
		}
		for i, image := range renamedImages {
			if !executorOwns(i) {
				continue
			}
			if err := image.Validate(); err != nil {
				errs = append(errs, fmt.Errorf("%q: %w", image.Image, err))
				continue
			}
			imageReferences, err := image.DestinationReferences(flagRegistry)
			if err != nil {
				errs = append(errs, fmt.Errorf("%q: %w", image.Image, err))
				continue
			}
			references = append(references, imageReferences...)
		}
	} else {
		tagsPerImage, err := skopeoSyncTags(filePath)
		if err != nil {
			return nil, err
		}
		for image, tags := range tagsPerImage {
			for _, tag := range tags {
				references = append(references, fmt.Sprintf("%s/%s:%s", flagRegistry, imageBaseName(image), tag))
			}
		}
	}
	sort.Strings(references)
	return references, errors.Join(errs...)
}

// commandSign is invoked when `retagger sign <path>` is called.
//
// It lists the references the file governs at the destination registry
// (governedReferences), resolves each and signs every digest that does not yet
// carry this job's signature. Tags absent from the registry and manifests
// cosign cannot sign are reported and skipped; a rule that could not be
// resolved, a signing failure or a verification failure fails the command once
// every reference has been attempted.
func commandSign(filePath string) {
	if filePath == "" {
		logrus.Fatal("You need to specify filepath: 'retagger sign <path>'")
	}
	if flagSignWorkers < 1 {
		logrus.Fatalf("%q cannot be lower than 1", "sign-workers")
	}
	validateExecutorFlags()
	logger := logrus.WithField("file", filePath)

	signer, err := newSigner(logger)
	if err != nil {
		logger.Fatal(err)
	}

	images, listErr := governedReferences(filePath)
	if listErr != nil {
		logger.Errorf("some rules could not be resolved: %v", listErr)
	}
	logger.Infof("Found %d image tags to sign at %q", len(images), flagRegistry)

	signer.SignAll(images, flagSignWorkers)

	logger.Info(signer.Summary())
	if signer.Failed() > 0 || listErr != nil {
		logger.Fatalf("Signing ended with %d errors", signer.Failed()+len(multiErrors(listErr)))
	}
}

// multiErrors unwraps an errors.Join result into its parts (none for nil).
func multiErrors(err error) []error {
	if err == nil {
		return nil
	}
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		return joined.Unwrap()
	}
	return []error{err}
}
