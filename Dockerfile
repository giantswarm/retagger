FROM gsoci.azurecr.io/giantswarm/skopeo:v1.15.0@sha256:85fc31993df6f0c8bbb553f2f105c12056ea846c133f341158c21d80f44312a9 AS skopeo-upstream

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
