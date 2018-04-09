# retagger

[![CircleCI](https://circleci.com/gh/giantswarm/retagger.svg?style=shield)](https://circleci.com/gh/giantswarm/retagger)

`retagger` is a tool for managing retagging third party images.

## Getting Project

Clone the git repository: https://github.com/giantswarm/retagger.git

### How to build/test

The standard way!

```
go build
```
```
go test
```

## About

`retagger` is used to move all images required by Giant Swarm to our registry, while maintaining integrity of images - that is, we should be able to determine where all images came from. The alternative is manually retagging and pushing images, which means we lose a degree of accountability.

A list of images (see `images.go`) is maintained, with a list of sha hashes and tag pairs. The `retagger` goes through each image, and then each sha hash and tag pair. The image is pulled by the sha hash, to overcome issues where a specific tag is rewritten, retagged with the tag, and pushed to our registry.

The image name is also rewritten. For example, `quay.io/coreos/hyperkube` becomes `quay.io/giantswarm/hyperkube`, `prom/prometheus` becomes `quay.io/giantswarm/prometheus`.

We attempt to pull the image from our registry first, to avoid unnecessary pulling of other images.

We currently only support public images, both for pulling and pushing.

The `retagger` works inside a CI build. On merges to master, the binary is executed. The workflow is to add a new image / sha tag pair in a PR, review it, and then merge. The `retagger` will take care that the image is handled. Users will still need to create repositories in the registry.

### Tags

An image's sha hash can be found like so:

```
$ docker pull quay.io/coreos/hyperkube:v1.5.2_coreos.2 | grep sha | awk -F ':' '{print $3}'
5ff22b5c65d5b93aa948b79028dc136a22cda2f049283103f10bd45650b47312
```

With tags, use the tag from the third party image. For example, the above image should be tagged `v1.5.2_coreos.2`. In cases where the tag has been updated, add an ID to the tag. For example, if the tag `v1.5.2_coreos.2` was updated, the second tag would be `v1.5.2_coreos.2-2`, the third would be `v1.5.2_coroes.2-3`, and so on.

### Running

The environment variables `REGISTRY`, `REGISTRY_ORGANISATION`, `REGISTRY_USERNAME` and `REGISTRY_PASSWORD` need to be set.

Executing
```
./retagger
```
will iterate through the defined images, pull them from a public registry, and push them to the specified private registry.

Note: This project is used internally at Giant Swarm, hence the hardcoded image descriptions.
Feel free to raise an issue if you have thoughts on re-using.
