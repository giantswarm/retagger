ARG ALPINE_VERSION=3.19
ARG GO_VERSION=1.22.1

FROM gsoci.azurecr.io/giantswarm/golang:${GO_VERSION}-alpine${ALPINE_VERSION} as builder

# Build a static skopeo binary
ARG SKOPEO_VERSION=v1.15.0
RUN apk add --no-cache git make bash
WORKDIR /build
RUN git clone --branch ${SKOPEO_VERSION} --depth 1 https://github.com/containers/skopeo.git
WORKDIR /build/skopeo
RUN BUILDTAGS=containers_image_openpgp DISABLE_CGO=1 CGO_ENABLED=0 make bin/skopeo

# Build retagger binary
WORKDIR /build/retagger
COPY main.go go.mod go.sum /build/retagger/
RUN CGO_ENABLED=0 go build -o retagger .

# Add both binaries to a fresh image
FROM gsoci.azurecr.io/giantswarm/alpine:${ALPINE_VERSION}
COPY --from=builder /build/skopeo/bin/skopeo /usr/local/bin/skopeo
COPY --from=builder /build/retagger/retagger /usr/local/bin/retagger

ENTRYPOINT ["retagger"]
