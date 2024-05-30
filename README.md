[![CircleCI](https://dl.circleci.com/status-badge/img/gh/giantswarm/retagger/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/giantswarm/retagger/tree/main)

# retagger

> A tool to handle the retagging of third party docker images and make them
  available in own registries.

## What does retagger do, exactly?

`retagger` is first and foremost a CircleCI worfklow that runs every day at 21:30
UTC and on every merge to master branch. It utilizes [skopeo][skopeo] and
[custom golang code](main.go) to take upstream docker images, customize them if
necessary, and push them to Giant Swarm's container registries: `quay.io` and
`giantswarm-registry.cn-shanghai.cr.aliyuncs.com`. It is capable of working
with `v1`, `v2`, and `OCI` registries, as well as retagging multi-architecture
images.

> 💡Please note it **is not responsible** for pushing images to neither
`docker.io/giantswarm`, nor `azurecr.io/giantswarm` container registries.

## How to add your image to the job

You've come to the right place. Pick one of the below methods. For both methods first make sure
the repository exists in the desired container registry.

### Plain copy

You do **not** need any customizations. Great!
1. Find a `skopeo-*.yaml` file in [images](images/) matching your upstream
   container registry's name. Create a new one, if necessary.
2. Add a tag, SHA, or a semantic version constraint for your image. Refer to
   [Skopeo](#skopeo) section or existing files for format definition.
3. Create the registry for the image on [quay.io](https://quay.io/organization/giantswarm).
4. If you haven't created a new file, that's it. You're set. Otherwise continue
   following the steps.
5. Open [CircleCI config][ciconf] and add your file to both `retag-registry`
   steps under `matrix.parameters.images_file`.

### Modify upstream Dockerfile

You need to modify the upstream image before it's pushed to Giant Swarm registries.
1. Find the [customized-images.yaml][custom] and add your image to the list.
   Please refer to [Customized Images](#customized-images) section to see the
   options.

### Manual copy

You need to copy a few tags, it's a one-off situation. You can use `docker
pull/docker tag/docker push` combination or the below `skopeo` snippet:

```bash
$ skopeo sync --src docker --dest docker --all --keep-going crossplane/crossplane:v1.11.0 docker.io/giantswarm/
```

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

### Customized images

Customized images are represented as an array of `CustomImage` objects. Please
see the below definition:

```golang
type CustomImage struct {
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
	// DockerfileExtras is a list of additional Dockerfile statements you want to
	// append to the upstream Dockerfile. (optional)
	// Example: ["RUN apk add -y bash"]
	DockerfileExtras []string `yaml:"dockerfile_extras,omitempty"`
	// AddTagSuffix is an extra string to append to the tag.
	// Example: "giantswarm", the tag would become "<tag>-giantswarm"
	AddTagSuffix string `yaml:"add_tag_suffix,omitempty"`
	// OverrideRepoName allows user to rewrite the name of the image entirely.
	// Example: "alpinegit", so "alpine" would become
	// "quay.io/giantswarm/alpinegit"
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
[custom]: images/customized-images.yaml
