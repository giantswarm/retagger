[![CircleCI](https://dl.circleci.com/status-badge/img/gh/giantswarm/retagger/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/giantswarm/retagger/tree/main)

# retagger

> A tool to handle the retagging of third-party docker images and make them
  available in their own registries.

## What does retagger do, exactly?

`retagger` is first and foremost a CircleCI workflow that runs every day at 21:30
UTC and on every merge to the main branch. It utilizes [skopeo][skopeo] and
[custom golang code](main.go) to take upstream docker images, rename them if
necessary, and push them to Giant Swarm's container registries: `gsoci.azurecr.io` and
`giantswarm-registry.cn-shanghai.cr.aliyuncs.com`. It is capable of working
with `v1`, `v2`, and `OCI` registries, as well as retagging multi-architecture
images.

> 💡Please note it **is not responsible** for pushing images to neither
`docker.io/giantswarm`, nor `azurecr.io/giantswarm` container registries.

Every image it copies to `gsoci.azurecr.io` is signed after the copy, see
[Signed images](#signed-images).

## How to add your image to the job

You've come to the right place. Pick one of the following methods. For both methods, ensure that the repository exists in the desired container registry first.

### Plain copy

You do **not** need any customizations. Great!
1. Find a `skopeo-*.yaml` file in [images](images/) matching your upstream
   container registry's name. Create a new one, if necessary.
2. Add a tag, SHA, or a semantic version constraint for your image. Refer to
   [Skopeo](#skopeo) section or existing files for format definition.
3. If you haven't created a new file, that's it. You're set. Otherwise, continue
   following the steps.
4. Open [CircleCI config][ciconf] and add your file to both `retag-registry`
   steps under `matrix.parameters.images_file`.

### Manual copy

You need to copy a few tags, it's a one-off situation. You can use `docker
pull/docker tag/docker push` combination or the below `skopeo` snippet:

```bash
$ skopeo sync --src docker --dest docker --all --keep-going crossplane/crossplane:v1.11.0 docker.io/giantswarm/
```

## Signed images

Every tag retagger copies to `gsoci.azurecr.io/giantswarm/` is signed right after
the copy with [cosign](https://github.com/sigstore/cosign) keyless signing: a
short-lived certificate from Fulcio for the CircleCI job's OIDC identity, the
signature recorded in the Rekor transparency log and stored next to the image as
an OCI referrer (cosign v3's bundle format, the same the architect orb uses for
images Giant Swarm builds). What is signed is the digest of the manifest the
registry serves for the tag, the image index for a multi-architecture image.

The signatures carry this identity:

| | |
|---|---|
| Issuer | `https://oidc.circleci.com` |
| Subject | `https://circleci.com/api/v2/projects/0a65bbac-fae8-4aca-bc8c-23d51903006b/pipeline-definitions/b8ddcac6-0699-5c43-a349-7906ac0d4a36` |

The subject is the CircleCI pipeline definition of this repository, the shape
Fulcio issues for every CircleCI job. Images built by the architect orb carry
the same shape with their own project, so one keyless attestor admits mirrored
and built images alike: issuer `https://oidc.circleci.com`, subject matching
`^https://circleci\.com/api/v2/projects/[a-f0-9-]+/pipeline-definitions/[a-f0-9-]+$`.

Verify a mirrored image (cosign v3 or newer):

```bash
cosign verify \
  --certificate-oidc-issuer https://oidc.circleci.com \
  --certificate-identity-regexp '^https://circleci\.com/api/v2/projects/[a-f0-9-]+/pipeline-definitions/[a-f0-9-]+$' \
  gsoci.azurecr.io/giantswarm/storage-initializer:v0.20.0
```

Admit mirrored and built images with one Kyverno `verifyImages` attestor:

```yaml
attestors:
  - entries:
      - keyless:
          issuer: https://oidc.circleci.com
          subjectRegExp: ^https://circleci\.com/api/v2/projects/[a-f0-9-]+/pipeline-definitions/[a-f0-9-]+$
          rekor:
            url: https://rekor.sigstore.dev
```

`retagger sign <skopeo yaml>` does the signing (see [`sign.go`](sign.go)): it lists
the tags the file governs the way `retagger filter` does, resolves each at the
registry and signs every digest that does not already verify against the job's
own identity, so it is idempotent. The `retag-registry` job runs it over the
`.filtered` file, the tags the run copied. The pipeline parameter `sign-all`
runs it over the unfiltered files instead, every tag they govern: that is the
one-off pass for tags mirrored before signing existed, and the repair of a run
whose signing failed. Trigger it on `main`:

```bash
curl -X POST -H "Circle-Token: $CIRCLE_TOKEN" -H "Content-Type: application/json" \
  -d '{"branch": "main", "parameters": {"sign-all": true}}' \
  https://circleci.com/api/v2/project/gh/giantswarm/retagger/pipeline
```

The copies in the Aliyun registry are not signed; the images renamed through
`retagger run` ([renamed images](#renamed-images)) are not signed yet either.
Mirrors in the Docker schema 1 manifest format (a few images from before 2019,
`etcd:v3.3` for one) cannot carry a cosign signature at all; `retagger sign`
reports them as unsignable and moves on.

## Image list formats

### Skopeo

The basic file format looks as follows:
`images/skopeo-registry-example-com.yaml`
```yaml
registry.example.com:
    images:
        redis:
            - "1.0"
            - "2.0"
            - "sha256:0000000000000000000000000000000011111111111111111111111111111111"
    images-by-semver:
        alpine: ">= 3.17"
```

The full specification is available in [upstream skopeo-sync docs][skopeo-sync
docs]. Semantic version constraint documentation is available in
[Masterminds/semver docs][masterminds docs].

### Custom image builds

Custom container image builds were moved to: https://github.com/giantswarm/custom-container-images.

### Renamed images

Renamed images are represented as an array of `RenamedImages` objects. Please see the definition below:

```golang
type RenamedImage struct {
	// Image is the full name of the image to pull.
	// Example: "alpine", "docker.io/giantswarm/app-operator", or
	// "ghcr.io/fluxcd/kustomize-controller"
	Image string `yaml:"image"`
	// TagOrPattern is used to filter image tags. All tags matching the pattern
	// will be retagged. Required if SHA is specified.
	// Example: "v1.[234].*" or ".*-stable"
	TagOrPattern string `yaml:"tag_or_pattern,omitempty"`
	// SHA is used to filter image tags. If SHA is specified, it will take
	// precedence over TagOrPattern. However TagOrPattern is still required!
	// Example: 234cb88d3020898631af0ccbbcca9a66ae7306ecd30c9720690858c1b007d2a0
	SHA string `yaml:"sha,omitempty"`
	// Semver is used to filter image tags by semantic version constraints. All
	// tags satisfying the constraint will be retagged.
	Semver string `yaml:"semver,omitempty"`
	// Filter is a regexp pattern used to extract a part of the tag for Semver
	// comparison. First matched group will be supplied for semver comparison.
	// Example:
	//   Filter: "(.+)-alpine"  ->  Image tag: "3.12-alpine" -> Comparison: "3.12>=3.10"
	//   Semver: ">= 3.10"          Extracted group: "3.12"
	Filter string `yaml:"filter,omitempty"`
	// AddTagSuffix is an extra string to append to the tag.
	// Example: "giantswarm", the tag would become "<tag>-giantswarm"
	AddTagSuffix string `yaml:"add_tag_suffix,omitempty"`
	// OverrideRepoName allows user to rewrite the name of the image entirely.
	// Example: "alpinegit", so "alpine" would become
	// "gsoci.azurecr.io/giantswarm/alpinegit". A slash keeps a nested path:
	// "kagent/controller" becomes "gsoci.azurecr.io/giantswarm/kagent/controller".
	OverrideRepoName string `yaml:"override_repo_name,omitempty"`
	// StripSemverPrefix removes the initial 'v' in 'v1.2.3' if enabled. Works
	// only when Semver is defined.
	StripSemverPrefix bool `yaml:"strip_semver_prefix,omitempty"`
}
```

## Contributing

Please refer to [CONTRIBUTING.md](CONTRIBUTING.md).

[skopeo]: https://github.com/containers/skopeo
[skopeo-sync docs]: https://github.com/kubasobon/skopeo/blob/semver/docs/skopeo-sync.1.md#yaml-file-content-used-source-for---src-yaml
[masterminds docs]: https://github.com/Masterminds/semver/tree/v3.2.0#basic-comparisons

[ciconf]: .circleci/config.yml
[renamed]: images/renamed-images.yaml
