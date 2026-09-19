package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// fakeJWT builds an unsigned JWT with the given claims; identityFromToken
// decodes, it does not verify.
func fakeJWT(t *testing.T, claims map[string]any) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	return header + "." + base64.RawURLEncoding.EncodeToString(payload) + ".signature"
}

func TestIdentityFromToken(t *testing.T) {
	token := fakeJWT(t, map[string]any{
		"aud":                          "sigstore",
		"oidc.circleci.com/project-id": "0a65bbac-fae8-4aca-bc8c-23d51903006b",
		"oidc.circleci.com/pipeline-definition-id": "b8ddcac6-0699-5c43-a349-7906ac0d4a36",
	})
	identity, err := identityFromToken(token)
	if err != nil {
		t.Fatal(err)
	}
	want := "https://circleci.com/api/v2/projects/0a65bbac-fae8-4aca-bc8c-23d51903006b/pipeline-definitions/b8ddcac6-0699-5c43-a349-7906ac0d4a36"
	if identity != want {
		t.Fatalf("identity = %q, want %q", identity, want)
	}
}

func TestIdentityFromTokenRejectsForeignTokens(t *testing.T) {
	for name, token := range map[string]string{
		"not a JWT":          "CCIPAT_not_a_jwt",
		"no CircleCI claims": fakeJWT(t, map[string]any{"aud": "sigstore", "sub": "repo:example/example"}),
		"missing definition": fakeJWT(t, map[string]any{"oidc.circleci.com/project-id": "0a65bbac"}),
	} {
		if _, err := identityFromToken(token); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestParseSkopeoSyncOutput(t *testing.T) {
	output := strings.Join([]string{
		`time="2026-09-18T20:30:49+02:00" level=info msg="Would have copied image ref 1/3" from="docker://kserve/storage-initializer:v0.20.0" to="docker://dry-run.invalid/storage-initializer:v0.20.0"`,
		`time="2026-09-18T20:30:49+02:00" level=info msg="Would have copied image ref 2/3" from="docker://kserve/storage-initializer:v0.20.1" to="docker://dry-run.invalid/storage-initializer:v0.20.1"`,
		`time="2026-09-18T20:30:49+02:00" level=info msg="Would have copied image ref 3/3" from="docker://quay.io/brancz/kube-rbac-proxy:v0.22.1" to="docker://dry-run.invalid/kube-rbac-proxy:v0.22.1"`,
		`time="2026-09-18T20:30:49+02:00" level=info msg="Would have synced 3 images from 2 sources"`,
	}, "\n")
	got := parseSkopeoSyncOutput(output)
	if len(got) != 2 {
		t.Fatalf("got %d images, want 2: %v", len(got), got)
	}
	if tags := got["kserve/storage-initializer"]; len(tags) != 2 || tags[0] != "v0.20.0" || tags[1] != "v0.20.1" {
		t.Errorf("storage-initializer tags = %v", tags)
	}
	if tags := got["quay.io/brancz/kube-rbac-proxy"]; len(tags) != 1 || tags[0] != "v0.22.1" {
		t.Errorf("kube-rbac-proxy tags = %v", tags)
	}
	if len(parseSkopeoSyncOutput(`level=info msg="Would have synced 0 images from 0 sources"`)) != 0 {
		t.Error("an empty sync must yield no images")
	}
}

func TestErrorPatterns(t *testing.T) {
	transient := []string{
		"Error: signing bundle: error signing bundle: stream error: stream ID 1; INTERNAL_ERROR; received from peer",
		"Error: PUT https://gsoci.azurecr.io/v2/giantswarm/x/manifests/sha256-...: unexpected status code 503 Service Unavailable",
		"error: [POST /api/v1/log/entries][429] createLogEntry default &{Code:429 Message:Too Many Requests}",
		"Error: signing [gsoci.azurecr.io/giantswarm/x@sha256:1a3a]: signing digest: failed to upload manifest: GET https://gsoci.azurecr.io/v2/giantswarm/x/referrers/sha256:1a3a: TOOMANYREQUESTS",
		"Error: GET https://gsoci.azurecr.io/v2/giantswarm/x/referrers/sha256:1a3a: TOOMANYREQUESTS: Identity x exceeded the per-identity ListReferrers rate limit of 1000 requests in a 60 second window.",
		"Error: setting trusted material: getting trusted root from TUF for new bundle verification: error getting live trusted root: failed to create TUF client failed to load metadata: tuf refresh failed: failed to persist metadata",
	}
	for _, s := range transient {
		if !transientErrorPattern.MatchString(s) {
			t.Errorf("expected transient: %q", s)
		}
	}
	permanent := []string{
		"Error: no matching signatures: none of the expected identities matched what was in the certificate",
		"Error: GET https://gsoci.azurecr.io/v2/giantswarm/x/manifests/v1: UNAUTHORIZED: authentication required",
		"Error: new-bundle timeout policy: sha256:5001234",
	}
	for _, s := range permanent {
		if transientErrorPattern.MatchString(s) {
			t.Errorf("expected permanent: %q", s)
		}
	}
	if transientErrorPattern.MatchString("Error: no signatures found") || !noSignaturePattern.MatchString("Error: no signatures found") {
		t.Error("an unsigned digest is not a transient error, but the read-after-write pattern must recognise it")
	}
	if !unsignedPattern.MatchString("Error: no signatures found") || foreignSignaturePattern.MatchString("Error: no signatures found") {
		t.Error("\"no signatures found\" is the unsigned answer, not a foreign signature")
	}
	foreign := `Error: no matching attestations: failed to verify certificate identity: no matching CertificateIdentity found, last error: expected SAN value "https://circleci.com/api/v2/projects/p/pipeline-definitions/d", got "https://circleci.com/api/v2/projects/q/pipeline-definitions/e"`
	if !foreignSignaturePattern.MatchString(foreign) || unsignedPattern.MatchString(foreign) || !noSignaturePattern.MatchString(foreign) {
		t.Error("another identity's signature must be recognised as foreign, and retried by the verify after a sign")
	}
	if !rekorDuplicatePattern.MatchString("error: [POST /api/v1/log/entries][409] createLogEntryConflict  &{Code:409 Message:an equivalent entry already exists in the transparency log}") {
		t.Error("Rekor 409 not recognised")
	}
	if !manifestUnknownPattern.MatchString(`reading manifest v0.20.0 in gsoci.azurecr.io/giantswarm/agent: manifest unknown`) {
		t.Error("missing tag not recognised")
	}
	if !manifestUnknownPattern.MatchString(`requested access to the resource is denied: NAME_UNKNOWN: repository "giantswarm/agent" is not found`) {
		t.Error("missing repository not recognised")
	}
}

// fakeCosign puts a `cosign` script on PATH that fails with message for the
// first failures calls and succeeds afterwards, and returns the file that
// counts the calls.
func fakeCosign(t *testing.T, failures int, message string) string {
	t.Helper()
	dir := t.TempDir()
	calls := filepath.Join(dir, "calls")
	script := fmt.Sprintf(`#!/bin/sh
n=$(cat %q 2>/dev/null || echo 0)
n=$((n + 1))
echo "$n" > %q
if [ "$n" -le %d ]; then
  echo %q >&2
  exit 1
fi
`, calls, calls, failures, message)
	if err := os.WriteFile(filepath.Join(dir, "cosign"), []byte(script), 0o700); err != nil { // #nosec G306 -- an executable test fixture
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return calls
}

func callCount(t *testing.T, calls string) int {
	t.Helper()
	b, err := os.ReadFile(calls) // #nosec G304 -- the test reads the counter file it created
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if _, err := fmt.Sscan(string(b), &n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestCosignRetriesTransientErrorsOnly(t *testing.T) {
	backoff = func(int) {}

	calls := fakeCosign(t, 2, "stream error: stream ID 1; INTERNAL_ERROR; received from peer")
	if err := cosign(nil, nil, "sign", "--yes", "example.invalid/img@sha256:0"); err != nil {
		t.Fatalf("transient failures must be retried: %v", err)
	}
	if n := callCount(t, calls); n != 3 {
		t.Errorf("cosign called %d times, want 3", n)
	}

	calls = fakeCosign(t, 1, "UNAUTHORIZED: authentication required")
	if err := cosign(nil, nil, "sign", "--yes", "example.invalid/img@sha256:0"); err == nil {
		t.Fatal("a permanent failure must not be retried into success")
	}
	if n := callCount(t, calls); n != 1 {
		t.Errorf("cosign called %d times, want 1", n)
	}

	calls = fakeCosign(t, cosignAttempts+1, "connection reset by peer")
	if err := cosign(nil, nil, "sign", "--yes", "example.invalid/img@sha256:0"); err == nil {
		t.Fatal("exhausted attempts must fail")
	}
	if n := callCount(t, calls); n != cosignAttempts {
		t.Errorf("cosign called %d times, want %d", n, cosignAttempts)
	}

	fakeCosign(t, 1, "createLogEntryConflict: an equivalent entry already exists in the transparency log")
	if err := cosign(nil, nil, "sign", "--yes", "example.invalid/img@sha256:0"); err != nil {
		t.Fatalf("a Rekor duplicate is a recorded signature: %v", err)
	}
}

func TestVerifyAfterSignRetriesTheReadAfterWriteRace(t *testing.T) {
	backoff = func(int) {}
	identity := "https://circleci.com/api/v2/projects/p/pipeline-definitions/d"

	// Before a sign, "no signatures found" is the answer for an unsigned digest: one call, no retry.
	calls := fakeCosign(t, 1, "Error: no signatures found")
	if err := cosignVerify("example.invalid/img@sha256:0", identity, false); err == nil {
		t.Fatal("an unsigned digest must not verify")
	}
	if n := callCount(t, calls); n != 1 {
		t.Errorf("verify before sign called cosign %d times, want 1", n)
	}

	// Right after a sign the same answer is the registry catching up: retried until it verifies.
	calls = fakeCosign(t, 2, "Error: no signatures found")
	if err := cosignVerify("example.invalid/img@sha256:0", identity, true); err != nil {
		t.Fatalf("the verify after a sign must outlast the referrers index: %v", err)
	}
	if n := callCount(t, calls); n != 3 {
		t.Errorf("verify after sign called cosign %d times, want 3", n)
	}
}

// fakeSkopeo puts a `skopeo` script on PATH whose `inspect --raw` prints manifest.
func fakeSkopeo(t *testing.T, manifest string) {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\ncase \"$*\" in *--raw*) printf '%s' \"$MANIFEST\";; *) exit 1;; esac\n"
	if err := os.WriteFile(filepath.Join(dir, "skopeo"), []byte(script), 0o700); err != nil { // #nosec G306 -- an executable test fixture
		t.Fatal(err)
	}
	t.Setenv("MANIFEST", manifest)
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestManifestDigestIsTheHashOfTheRawManifest(t *testing.T) {
	// An OCI index without a linux/amd64 instance: content-addressed like any other manifest.
	index := `{"schemaVersion":2,"mediaType":"application/vnd.oci.image.index.v1+json","manifests":[{"mediaType":"application/vnd.oci.image.manifest.v1+json","digest":"sha256:a","size":1,"platform":{"os":"linux","architecture":"arm64"}}]}`
	fakeSkopeo(t, index)
	sum := sha256.Sum256([]byte(index))
	digest, err := manifestDigest("example.invalid/ns/img:v1")
	if err != nil || digest != "sha256:"+hex.EncodeToString(sum[:]) {
		t.Fatalf("digest = %q, err = %v", digest, err)
	}
}

func TestSchema1ManifestsAreUnsignable(t *testing.T) {
	fakeSkopeo(t, `{"schemaVersion":1,"name":"giantswarm/etcd","tag":"v3.3","fsLayers":[],"history":[],"signatures":[]}`)
	if _, err := manifestDigest("example.invalid/ns/etcd:v3.3"); !errors.Is(err, errUnsignable) {
		t.Fatalf("err = %v, want errUnsignable", err)
	}
	outcome, err := signTag("example.invalid/ns/etcd:v3.3", "https://circleci.com/api/v2/projects/p/pipeline-definitions/d", &oidcTokenSource{})
	if err != nil || outcome != outcomeUnsignable {
		t.Fatalf("outcome = %v, err = %v; want unsignable, nil", outcome, err)
	}
}

func TestSignTagReportsMissingImages(t *testing.T) {
	dir := t.TempDir()
	script := "#!/bin/sh\necho 'reading manifest v0.20.0 in example.invalid/ns/agent: manifest unknown' >&2\nexit 1\n"
	if err := os.WriteFile(filepath.Join(dir, "skopeo"), []byte(script), 0o700); err != nil { // #nosec G306 -- an executable test fixture
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	outcome, err := signTag("example.invalid/ns/agent:v0.20.0", "https://circleci.com/api/v2/projects/p/pipeline-definitions/d", &oidcTokenSource{})
	if err != nil || outcome != outcomeMissing {
		t.Fatalf("outcome = %v, err = %v; want missing, nil", outcome, err)
	}
	if _, err := manifestDigest("example.invalid/ns/agent:v0.20.0"); !errors.Is(err, errManifestUnknown) {
		t.Fatalf("err = %v, want errManifestUnknown", err)
	}
}

// fakeRegistry is an in-memory registry with the referrers API behind a
// handler that answers the first throttled referrers listings the way the
// registry throttles a busy pass, and counts the listings it saw. DOCKER_CONFIG
// points at an empty directory so the machine's credentials stay out of the test.
type fakeRegistry struct {
	host      string
	listings  atomic.Int64
	throttled atomic.Int64
}

func newFakeRegistry(t *testing.T, throttled int) *fakeRegistry {
	t.Helper()
	f := &fakeRegistry{}
	f.throttled.Store(int64(throttled))
	backend := registry.New(registry.WithReferrersSupport(true))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/referrers/") {
			f.listings.Add(1)
			if f.throttled.Add(-1) >= 0 {
				http.Error(w, `{"errors":[{"code":"TOOMANYREQUESTS","message":"Identity x exceeded the per-identity ListReferrers rate limit of 1000 requests in a 60 second window."}]}`, http.StatusTooManyRequests)
				return
			}
		}
		backend.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("DOCKER_CONFIG", t.TempDir())
	f.host = strings.TrimPrefix(srv.URL, "http://")
	return f
}

// rawManifest is a manifest to push as it is.
type rawManifest []byte

func (m rawManifest) RawManifest() ([]byte, error) { return m, nil }

// image pushes manifest as repository:tag and returns the tag's reference.
func (f *fakeRegistry) image(t *testing.T, repository, tag, manifest string) string {
	t.Helper()
	ref, err := name.NewTag(f.host + "/" + repository + ":" + tag)
	if err != nil {
		t.Fatal(err)
	}
	if err := remote.Put(ref, rawManifest(manifest)); err != nil {
		t.Fatal(err)
	}
	return ref.String()
}

// referrer pushes a referrer of digest into repository of the shape cosign v3
// stores a signature in when artifactType is signatureBundleArtifactType.
func (f *fakeRegistry) referrer(t *testing.T, repository, digest, artifactType string) {
	t.Helper()
	subject, err := v1.NewHash(digest)
	if err != nil {
		t.Fatal(err)
	}
	img := mutate.ConfigMediaType(mutate.MediaType(empty.Image, types.OCIManifestSchema1), types.MediaType(artifactType))
	img, err = mutate.AppendLayers(img, static.NewLayer([]byte("{}"), types.MediaType(artifactType)))
	if err != nil {
		t.Fatal(err)
	}
	img = mutate.Subject(img, v1.Descriptor{MediaType: types.OCIManifestSchema1, Digest: subject, Size: 1}).(v1.Image)
	repo, err := name.NewRepository(f.host + "/" + repository)
	if err != nil {
		t.Fatal(err)
	}
	referrerDigest, err := img.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if err := remote.Write(repo.Digest(referrerDigest.String()), img); err != nil {
		t.Fatal(err)
	}
}

func digestOf(manifest string) string {
	sum := sha256.Sum256([]byte(manifest))
	return "sha256:" + hex.EncodeToString(sum[:])
}

const testManifest = `{"schemaVersion":2,"mediaType":"application/vnd.oci.image.manifest.v1+json","config":{"mediaType":"application/vnd.oci.image.config.v1+json","digest":"sha256:0","size":1},"layers":[]}`

func TestHasSignatureBundle(t *testing.T) {
	backoff = func(int) {}
	digest := digestOf(testManifest)

	reg := newFakeRegistry(t, 0)
	reg.image(t, "ns/img", "v1", testManifest)
	ref := reg.host + "/ns/img@" + digest
	if signed, err := hasSignatureBundle(ref); err != nil || signed {
		t.Fatalf("no referrers: signed = %v, err = %v; want false, nil", signed, err)
	}
	reg.referrer(t, "ns/img", digest, "application/vnd.example.sbom")
	if signed, err := hasSignatureBundle(ref); err != nil || signed {
		t.Fatalf("a referrer of another artifact type: signed = %v, err = %v; want false, nil", signed, err)
	}
	reg.referrer(t, "ns/img", digest, signatureBundleArtifactType)
	if signed, err := hasSignatureBundle(ref); err != nil || !signed {
		t.Fatalf("a signature bundle: signed = %v, err = %v; want true, nil", signed, err)
	}

	// The registry's throttling of the listing is retried, not read as "unsigned".
	throttled := newFakeRegistry(t, 2)
	throttled.image(t, "ns/img", "v1", testManifest)
	throttled.referrer(t, "ns/img", digest, signatureBundleArtifactType)
	if signed, err := hasSignatureBundle(throttled.host + "/ns/img@" + digest); err != nil || !signed {
		t.Fatalf("throttled listing: signed = %v, err = %v; want true, nil", signed, err)
	}
	if n := throttled.listings.Load(); n < 3 {
		t.Errorf("the throttled listing was tried %d times, want at least 3", n)
	}
}

// TestSignTagTrustsAnUnsignedAnswerOnlyWithTheRegistry pins the pre-check:
// cosign's "no signatures found" leads to a sign only when the registry lists
// no signature bundle; when it lists one, the verify is retried; another
// identity's signature is signed over without asking the registry; a verify
// that cannot be completed fails the tag instead of signing it.
func TestSignTagTrustsAnUnsignedAnswerOnlyWithTheRegistry(t *testing.T) {
	backoff = func(int) {}
	identity := "https://circleci.com/api/v2/projects/p/pipeline-definitions/d"
	tokens := &oidcTokenSource{token: "token", mintedAt: time.Now()} // #nosec G101 -- a test fixture, no token is minted
	digest := digestOf(testManifest)

	for _, tc := range []struct {
		name         string
		verifyError  string
		bundle       bool
		wantOutcome  signOutcome
		wantErr      bool
		wantCosign   int
		wantListings int64
	}{
		{"unsigned: the registry lists no bundle, the digest is signed", "Error: no signatures found", false, outcomeSigned, false, 3, 1},
		{"a throttled lookup: the registry lists a bundle, the verify is retried", "Error: no signatures found", true, outcomeSkipped, false, 2, 1},
		{"another identity's signature: signed over without asking the registry", "Error: no matching attestations: failed to verify certificate identity", true, outcomeSigned, false, 3, 0},
		{"a verify that cannot be completed is not a reason to sign", "Error: GET https://example.invalid/v2/ns/img/manifests/sha256:0: UNAUTHORIZED: authentication required", true, outcomeFailed, true, 1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reg := newFakeRegistry(t, 0)
			image := reg.image(t, "ns/img", "v1", testManifest)
			if tc.bundle {
				reg.referrer(t, "ns/img", digest, signatureBundleArtifactType)
			}
			fakeSkopeo(t, testManifest)
			calls := fakeCosign(t, 1, tc.verifyError)
			reg.listings.Store(0) // pushing the fixtures lists the referrers too

			outcome, err := signTag(image, identity, tokens)
			if outcome != tc.wantOutcome || (err != nil) != tc.wantErr {
				t.Fatalf("outcome = %v, err = %v; want %v, error %v", outcome, err, tc.wantOutcome, tc.wantErr)
			}
			if n := callCount(t, calls); n != tc.wantCosign {
				t.Errorf("cosign called %d times, want %d", n, tc.wantCosign)
			}
			if n := reg.listings.Load(); n != tc.wantListings {
				t.Errorf("the registry's referrers were listed %d times, want %d", n, tc.wantListings)
			}
		})
	}
}
