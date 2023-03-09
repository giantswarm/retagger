[![CircleCI](https://circleci.com/gh/giantswarm/retagger.svg?style=shield)](https://circleci.com/gh/giantswarm/retagger)

# retagger

> A tool to handle the retagging of third party docker images and make them
  available in own registries.

## Building retagger

Based on [CircleCI golang image](https://hub.docker.com/r/cimg/go) and [skopeo](https://github.com/containers/skopeo).

Your best bet is building a docker container using the `Dockerfile` contained in this repository:
```
docker build -t retagger:latest .
```
The Dockerfile uses a Golang container to build a static binary of `skopeo`
(following [this doc](https://github.com/containers/skopeo/blob/main/install.md#building-a-static-binary)).
Then it copies the binary to a `cimg/go`-based container, which is an official
CircleCI's runner with Golang installed. The resulting image is used in the
`.circleci/config`.

You can test resulting image by running:
```
docker run --rm -it retagger:latest skopeo --help
```

Check `skopeo` version as well:
```
~ docker run --rm -it retagger-test:latest skopeo --version
skopeo version 1.11.1-dev
```

## Updating versions

> **Important:** Make sure Golang versions in the Dockerfile match. Just in case.

- **CircleCI runner**, also Golang version  - update `GOLANG_VERSION` version in the [Dockerfile](/Dockerfile).
- **Skopeo** - update `skopeo` [subtree](/skopeo). You might want to update
               `GOLANG_VERSION` as well to match [the upstream](https://github.com/containers/skopeo/blob/main/go.mod#L3).
