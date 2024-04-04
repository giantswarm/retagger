FROM gsoci.azurecr.io/giantswarm/golang:1.22.1-alpine3.19 as builder

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

# Add both binraries to a fresh image
FROM gsoci.azurecr.io/giantswarm/alpine:3.19
COPY --from=builder /build/skopeo/bin/skopeo /usr/local/bin/skopeo
COPY --from=builder /build/retagger/retagger /usr/local/bin/retagger

ENTRYPOINT ["retagger"]
