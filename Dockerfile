FROM golang:1.18 AS skopeo-builder
COPY skopeo /skopeo
WORKDIR /skopeo
ENV CGO_ENABLED=0
RUN DISABLE_DOCS=1 make BUILDTAGS=containers_image_openpgp GO_DYN_FLAGS=''

FROM cimg/go:1.18.0
USER root
COPY --from=skopeo-builder /skopeo/bin/skopeo /usr/bin/
USER circleci
