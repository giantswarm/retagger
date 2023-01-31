[![CircleCI](https://circleci.com/gh/giantswarm/retagger.svg?style=shield)](https://circleci.com/gh/giantswarm/retagger)

# retagger

A tool to handle the retagging of third party docker images and make them
available in own registries.

Basically, this automates the following for a bunch of images and tags:

```nohighlight

docker pull \
  [[<registry>/]<namespace>/]<image>@sha256:<sha>

docker tag \
  <sha> \
  <my_registry>/<my_organization>/<image>:<my_tag>

docker push \
  <my_registry>/<my_organization>/<image>:<my_tag>
```

So, for example:

    gcr.io/foo/bar:first-version

will become

    quay.io/giantswarm/bar:first-version
    giantswarm-registry.cn-shanghai.cr.aliyuncs.com/giantswarm/bar:first-version

Images can be defined using their SHA or using a tag pattern.

If using a SHA, the exact image version is pulled from the source repository, pushed to Quay, and tagged with the given
tag.

If using a pattern, all tags in the original repository are compared against the pattern.
If a matching tag is found, the image at that tag is pulled and retagged in Quay with the same tag
(e.g. `alpine:9.99` matches pattern `9.[0-9]*` --> retagged as `quay.io/giantswarm/alpine:9.99`).

Besides avoiding manual work, we do this for accountability and reproducability.
Changes in the image configuration are tracked in the repo. `retagger` execution
happens in CI and logs are available.

## Image configuration file

In `images.yaml` you can see the image configuration used by Giant Swarm,
with one entry for (almost) every third party image we rely on.
An image entry has one or several sub-entries for the versions we need.
Here is an example entry:

```yaml
- name: fluent/fluent-bit
  comment: see https://hub.docker.com/r/fluent/fluent-bit/ for details
  tags:
  - sha: 6861ee5ea81fbbacf8802c622d9305930b341acc0800bbf30205b4d74ce2b486
    tag: "0.14.6"
    customImages:
    - tagSuffix: giantswarm
      dockerfileOptions:
      - EXPOSE 1053
- name: registry.k8s.io/hyperkube
  tags:
  - sha: 29590ae7991517b214a9239349ee1cc803a22f2a36279612a52caa2bc8673ff0
    tag: v1.16.3
  patterns:
    - pattern: '>= v1.17.0'       # Match any v1.17.x
```

What the attributes mean:

- `name`: The image name to be used with `docker pull`, without a tag.
- `comment` (optional): Allows adding helpful information to the entry.
- `overrideRepoName` (optional): If set, use this repository name on Quay instead of the original repo name.
- `tags`: List of image versions.
- `tags[].sha`: The SHA describing the version to pull from the source registry.
- `tags[].tag`: The image tag to apply in the target registry.
- `tags[].customImages[]`: Custom images definition with original tag as base.
- `tags[].customImages[].tagSuffix`: Tag suffix for custom image build.
- `tags[].customImages[].dockerfileOptions[]`: The list of Dockerfile options, used to override base image
- `patterns[]`: List of patterns. New tags matching one of these patterns will be automatically retagged.
- `patterns[].pattern`: Valid semver condition to match tags.
- `patterns[].filter`: Regex to filter tags before validating semver. If empty, filtering is not done. 
- `patterns[].customImages[]`: Custom images as explained above.

An image may define both `tags` and `patterns`.
A `pattern` may also include all optional features of a `tag`, such as a `tagSuffix` or `dockerfileOptions`.
The `filter` MUST include the `<version>` capture group, e.g.: `(?P<version>.*)-alpine`. This way, the version is extracted from the tag and validated against SemVer as usual.

## Adding an image

Images can be added by SHA or by pattern. It is preferable to use a SHA whenever possible as this avoids tagging
unnecessary images and guarantees a certain known image. Patterns should be used when automation is in place to handle
new images, and not simply used as a convenience.

**Note:** Images in the `images.yaml` file need to be sorted alphabetically, otherwise CI will stay red!

### By SHA

To add an image to the configuration file `images.yaml`, find out the SHA of the
image like this (where `image:tag` has to be replaced with the real image name
and tag):

```nohighlight
$ docker pull image:tag | grep sha | awk -F ':' '{print $3}'
```

As a target tag, please use the original tag.

In cases where the original image has been updated but the tag stayed the same,
add a version counter to the tag. For example, if the tag `v1.5.2` was updated,
the target tag should be `v1.5.2-2`, `v1.5.2-3`, and so on.

### By Pattern

**Warning:** Run `retagger` locally with `--dry-run` _before_ pushing your pattern to the repository to make sure it
does what you want.

It is also possible to watch and automatically retag new tagged releases in the upstream repository.
To do this, specify a pattern for the image in the `images.yaml` configuration file.
Each pattern must be a valid semver constraint (you can read about it [here](https://github.com/Masterminds/semver)),
and should match as little as possible to avoid retagging huge numbers of useless images.

_Hint:_ The `v` infront of a version is optional - so `v1.0.0` and `1.0.0` behave the same.

For example, at the time of this writing:

```yaml
- name: registry.k8s.io/hyperkube
  patterns:
    - pattern: '>= v1.17.0'       # Match any v1.17.x
```

### Execution

Please provide a PR with the change. Once merged into master, CI will execute
`retagger` to push any new images.

**Note**: To keep execution speedy, when adding a new version, please remove older versions (tags) that are no longer
used from the configuration.

## Background and details

- `retagger` checks if an image and tag are already available in the target
  registry, to avoid unnecessary pushing.

- `retagger` currently only supports public images as a source.

- In Quay, a repository must exist for the image before retagger can push an image.

The `retagger` works inside a CI build. On merges to master, the binary is executed. The workflow is to add a new image
/ sha tag pair or pattern in a PR, review it, and then merge. The `retagger` will take care that the image is handled.
Users will still need to create repositories in the registry.

### Usage

The environment variables `REGISTRY`, `REGISTRY_ORGANISATION`, `REGISTRY_USERNAME`, and `REGISTRY_PASSWORD` need to be
set (or passed as arguments).

Executing

```console
./retagger
```

will iterate through the defined images, pull them from a public registry, and push them to the specified private
registry.

#### Options

```console
Usage:
  retagger [flags]
  retagger [command]

Available Commands:
  help        Help about any command
  version     Prints version information.

Flags:
      --dry-run               if set, will list jobs but not run them
  -f, --file string           retagger config file to use (default "images.yaml")
  -h, --help                  help for retagger
  -r, --host string           Registry hostname (e.g. quay.io)
  -o, --organization string   organization to tag images for (default "giantswarm")
  -p, --password string       password to authenticate against registry
  -u, --username string       username to authenticate against registry

Use "retagger [command] --help" for more information about a command.
```

### How to build/test

The standard way!

```nohighlight
go test
go build
```
