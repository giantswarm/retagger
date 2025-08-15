ARG ALPINE_VERSION=3.19
ARG GO_VERSION=1.22.1

FROM gsoci.azurecr.io/giantswarm/golang:${GO_VERSION}-alpine${ALPINE_VERSION} as builder

RUN apk add --no-cache git make bash curl

# Build a static skopeo binary
ARG SKOPEO_VERSION=v1.20.0
WORKDIR /build
RUN git clone --branch ${SKOPEO_VERSION} --depth 1 https://github.com/containers/skopeo.git
WORKDIR /build/skopeo
RUN BUILDTAGS=containers_image_openpgp DISABLE_CGO=1 CGO_ENABLED=0 make bin/skopeo

# Build retagger binary
WORKDIR /build/retagger
COPY main.go go.mod go.sum /build/retagger/
RUN CGO_ENABLED=0 go build -o retagger .

# Fetch docker binary
WORKDIR /build/docker
ARG DOCKER_VERSION=25.0.5
RUN curl -O https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz && tar -xvf docker-${DOCKER_VERSION}.tgz

FROM gsoci.azurecr.io/giantswarm/skopeo:v1.15.0 as skopeo

# Add all binaries to a fresh image
FROM gsoci.azurecr.io/giantswarm/alpine:${ALPINE_VERSION}

# We need bash for CircleCI script execution
RUN apk add --no-cache bash

COPY --from=builder /build/skopeo/bin/skopeo /usr/local/bin/skopeo
COPY --from=builder /build/retagger/retagger /usr/local/bin/retagger
COPY --from=builder /build/docker/docker/docker /usr/local/bin/docker

# Copy trust policies
COPY --from=skopeo /etc/containers /etc/containers

ENTRYPOINT ["retagger"]
