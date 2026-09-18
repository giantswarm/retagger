# Every pin below is kept current by Renovate: the FROM tags and the skopeo ARG
# (used in a FROM) by its Dockerfile manager, the other ARGs by the org preset's
# comment manager, which reads the annotation line above each of them.
ARG SKOPEO_VERSION=v1.19.0

FROM gsoci.azurecr.io/giantswarm/golang:1.25.0-alpine3.22 AS builder

RUN apk add --no-cache git make bash curl

# Build a static skopeo binary from the release whose trust policies the final
# image copies below (the same SKOPEO_VERSION).
WORKDIR /build
ARG SKOPEO_VERSION
RUN git clone --branch "${SKOPEO_VERSION}" --depth 1 https://github.com/containers/skopeo.git
WORKDIR /build/skopeo
RUN BUILDTAGS=containers_image_openpgp DISABLE_CGO=1 CGO_ENABLED=0 make bin/skopeo

# Build retagger binary
WORKDIR /build/retagger
COPY *.go go.mod go.sum /build/retagger/
RUN CGO_ENABLED=0 go build -o retagger .

# Fetch docker binary
WORKDIR /build/docker
# renovate: datasource=docker depName=docker
ARG DOCKER_VERSION=25.0.5
RUN curl -O https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz && tar -xvf docker-${DOCKER_VERSION}.tgz

# Fetch cosign, which `retagger sign` signs the mirrored images with. The same
# version the architect image carries, so mirrored and built images are signed alike.
WORKDIR /build/cosign
# renovate: datasource=github-releases depName=sigstore/cosign
ARG COSIGN_VERSION=v3.1.3
RUN curl -sSLO https://github.com/sigstore/cosign/releases/download/${COSIGN_VERSION}/cosign-linux-amd64 && \
    curl -sSLO https://github.com/sigstore/cosign/releases/download/${COSIGN_VERSION}/cosign_checksums.txt && \
    grep ' cosign-linux-amd64$' cosign_checksums.txt | sha256sum -c - && \
    install -m 0755 cosign-linux-amd64 cosign

FROM gsoci.azurecr.io/giantswarm/skopeo:${SKOPEO_VERSION} AS skopeo

# Add all binaries to a fresh image
FROM gsoci.azurecr.io/giantswarm/alpine:3.22.1

# We need bash for CircleCI script execution
RUN apk add --no-cache bash

COPY --from=builder /build/skopeo/bin/skopeo /usr/local/bin/skopeo
COPY --from=builder /build/retagger/retagger /usr/local/bin/retagger
COPY --from=builder /build/docker/docker/docker /usr/local/bin/docker
COPY --from=builder /build/cosign/cosign /usr/local/bin/cosign

# Copy trust policies
COPY --from=skopeo /etc/containers /etc/containers

ENTRYPOINT ["retagger"]
