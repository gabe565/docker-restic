ARG RESTIC_IMAGE=restic/restic
ARG RESTIC_TAG=0.19.1

FROM ghcr.io/bdd/runitor:v1.4.1-alpine@sha256:6615cb301e31efeab3f6084b66972bd66da3fe484356e2c912a3814229b02079 AS runitor

FROM ghcr.io/gabe565/moreutils:0.6.1@sha256:ae74fb352762709efdb79a67d1b51ab78d6fd04b2e39eb6d53e218674e7dc4e5 AS moreutils

FROM golang:1.27.4-alpine AS go-deps
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY internal internal

FROM go-deps AS wrapper
COPY cmd/restic-wrapper cmd/restic-wrapper
RUN go build -ldflags='-w -s' -trimpath ./cmd/restic-wrapper

FROM go-deps AS dumpdb
COPY cmd/dumpdb cmd/dumpdb
RUN go build -ldflags='-w -s' -trimpath ./cmd/dumpdb


FROM $RESTIC_IMAGE:$RESTIC_TAG AS restic
RUN test -f /usr/bin/restic

RUN apk add --no-cache bash postgresql-client mariadb-client mongodb-tools sqlite

COPY --from=runitor /usr/local/bin/runitor /usr/bin/runitor
COPY --from=moreutils /usr/bin/ts /usr/bin/ts
COPY --from=wrapper /app/restic-wrapper /usr/local/bin/restic
COPY --from=dumpdb /app/dumpdb /usr/bin/dumpdb
COPY rootfs /

ENV KUBECONFIG=/.kube/config
ENV RESTIC_GROUP_BY=tags
ENV RESTIC_CACHE_DIR=/cache
