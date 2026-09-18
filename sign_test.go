package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
