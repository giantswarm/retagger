ARG GOLANG_VERSION=1.18

FROM cimg/go:${GOLANG_VERSION} AS skopeo-builder
COPY skopeo /skopeo
WORKDIR /skopeo
ENV CGO_ENABLED=0
RUN DISABLE_DOCS=1 make BUILDTAGS=containers_image_openpgp GO_DYN_FLAGS=''

FROM quay.io/skopeo/stable:v1 AS skopeo-upstream
# RUN find / -name 'containers'

FROM cimg/go:${GOLANG_VERSION}
USER root
COPY --from=skopeo-upstream /etc/containers/* /etc/containers/
COPY --from=skopeo-upstream /usr/share/containers/* /usr/share/containers/
COPY --from=skopeo-upstream /var/lib/containers/* /var/lib/containers/
COPY --from=skopeo-builder /skopeo/bin/skopeo /usr/bin/
COPY retagger retagger
USER circleci
