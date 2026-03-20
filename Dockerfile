ARG RESTIC_IMAGE=restic/restic
ARG RESTIC_TAG=0.18.1@sha256:39d9072fb5651c80d75c7a811612eb60b4c06b32ffe87c2e9f3c7222e1797e76

FROM ghcr.io/bdd/runitor:v1.4.1-alpine@sha256:2b8314456908065f5d79cd25c7c310ddecd52995cea9d2a3526a4b0b19b4dc94 AS runitor

FROM ghcr.io/gabe565/moreutils:0.6.0@sha256:50aea6b946b78e2988f800964c60832b7059f82e5da7d1b1833b10a66b6d57bd AS moreutils

FROM $RESTIC_IMAGE:$RESTIC_TAG AS restic

RUN apk add --no-cache postgresql-client mariadb-client mongodb-tools sqlite

COPY --from=runitor /usr/local/bin/runitor /usr/bin/runitor
COPY --from=moreutils /usr/bin/ts /usr/bin/ts
COPY rootfs /

ENV KUBECONFIG=/.kube/config
ENV RESTIC_GROUP_BY=tags
