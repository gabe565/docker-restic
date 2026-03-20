ARG RESTIC_IMAGE=restic/restic
ARG RESTIC_TAG=0.18.1

FROM ghcr.io/bdd/runitor:v1.4.1-alpine@sha256:2b8314456908065f5d79cd25c7c310ddecd52995cea9d2a3526a4b0b19b4dc94 AS runitor

FROM ghcr.io/gabe565/moreutils:0.6.1@sha256:ae74fb352762709efdb79a67d1b51ab78d6fd04b2e39eb6d53e218674e7dc4e5 AS moreutils

FROM $RESTIC_IMAGE:$RESTIC_TAG AS restic
RUN test -f /usr/bin/restic

RUN apk add --no-cache postgresql-client mariadb-client mongodb-tools sqlite

COPY --from=runitor /usr/local/bin/runitor /usr/bin/runitor
COPY --from=moreutils /usr/bin/ts /usr/bin/ts
COPY rootfs /

ENV KUBECONFIG=/.kube/config
ENV RESTIC_GROUP_BY=tags
