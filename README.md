# retagger

[![CircleCI](https://circleci.com/gh/giantswarm/retagger.svg?style=shield)](https://circleci.com/gh/giantswarm/retagger)

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
    registry-intl.cn-shanghai.aliyuncs.com/giantswarm/bar:first-version

By relying on SHAs for pulling we make sure to get exactly the image version we
expect.

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
```

What the attributes mean:

- `name`: The image name to be used with `docker pull`, without a tag
- `comment` (optional): Allows to add helpful information to the entry
- `tags`: List of image versions
- `tags[].sha`: The SHA describing the version to pull from the source registry
- `tags[].tag`: The image tag to apply in the target registry

## Adding an image

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

Please provide a PR with the change. Once merged into master, CI will execute
`retagger` to push any new images.

**Note**: To keep execution speedy, when adding a new version, please remove older versions (tags) that are no longer used from the configuration.

## Background and details

- `retagger` checks if an image and tag is already available in the target
registry, to avoid unnecessary pushing.

- `retagger` currently only supports public images as a source.

- In Quay, a repository must exist for the image before retagger can push an image.

The `retagger` works inside a CI build. On merges to master, the binary is executed. The workflow is to add a new image / sha tag pair in a PR, review it, and then merge. The `retagger` will take care that the image is handled. Users will still need to create repositories in the registry.

### Running

The environment variables `REGISTRY`, `REGISTRY_ORGANISATION`, `REGISTRY_USERNAME`, and `REGISTRY_PASSWORD` need to be set.

Executing

```
./retagger
```

will iterate through the defined images, pull them from a public registry, and push them to the specified private registry.

### How to build/test

The standard way!

```nohighlight
go test
go build
```
