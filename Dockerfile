FROM quay.io/skopeo/stable:v1.15.0@sha256:3ddd5a84d11b8ea4447e8f6ec5e6a749832642724e041837b7de98f2c7f62927 AS skopeo-upstream

FROM cimg/go:1.21
USER root
COPY --from=skopeo-upstream /etc/containers/* /etc/containers/
COPY --from=skopeo-upstream /usr/share/containers/* /usr/share/containers/
COPY --from=skopeo-upstream /var/lib/containers/* /var/lib/containers/
COPY --from=skopeo-upstream /usr/bin/skopeo /usr/bin/
RUN mkdir -p /run/containers && \
    chown -R circleci:circleci /run/containers
COPY retagger retagger
USER circleci
