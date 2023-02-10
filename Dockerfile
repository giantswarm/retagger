ARG GOLANG_VERSION=1.18

FROM golang:${GOLANG_VERSION} AS skopeo-builder
COPY skopeo /skopeo
WORKDIR /skopeo
ENV CGO_ENABLED=0
RUN DISABLE_DOCS=1 make BUILDTAGS=containers_image_openpgp GO_DYN_FLAGS=''

FROM cimg/go:${GOLANG_VERSION}
USER root
COPY --from=skopeo-builder /skopeo/bin/skopeo /usr/bin/
USER circleci
