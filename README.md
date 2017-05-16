# retagger

[![CircleCI](https://circleci.com/gh/giantswarm/retagger.svg?style=shield)](https://circleci.com/gh/giantswarm/retagger)

`retagger` is a tool for managing retagging third party images.

`retagger` is used to move all images required by Giant Swarm to our registry, while maintaining integrity of images - that is, we should be able to determine where all images came from. The alternative is manually retagging and pushing images, which means we lose a degree of accountability.

A list of images (see `images.go`) is maintained, with a list of sha hashes and tag pairs. The `retagger` goes through each image, and then each sha hash and tag pair. The image is pulled by the sha hash, to overcome issues where a specific tag is rewritten, retagged with the tag, and pushed to our registry.

The image name is also rewritten. For example, `quay.io/coreos/hyperkube` becomes `quay.io/giantswarm/hyperkube`, `prom/prometheus` becomes `quay.io/giantswarm/prometheus`.

We attempt to pull the image from our registry first, to avoid unnecessary pulling of other images.

We currently only support public images, both for pulling and pushing.

The `retagger` works inside a CI build. On merges to master, the binary is executed. The workflow is to add a new image / sha tag pair in a PR, review it, and then merge. The `retagger` will take care that the image is handled. Users will still need to create repositories in the registry.