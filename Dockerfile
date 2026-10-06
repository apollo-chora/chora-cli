# syntax=docker/dockerfile:1.6
#
# chora-cli Dockerfile — standalone Go CLI.
#
# Build context = this repository. The CLI talks to the Chora gateway over
# HTTP; it listens on no ports and needs no database.

ARG GO_VERSION=1.26.6
ARG ALPINE_VERSION=3.23
ARG SERVICE_NAME=chora-cli
ARG GIT_SHA=unknown
ARG BUILD_TIME=unknown

############################
# Stage 1 — build
############################
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

ARG SERVICE_NAME
ARG GIT_SHA
ARG BUILD_TIME
ARG TARGETARCH

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY . .

RUN go mod download

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=${TARGETARCH}
RUN go build -trimpath \
      -ldflags "-s -w \
        -X main.version=${GIT_SHA}" \
      -o /out/chora \
      ./cmd/chora

############################
# Stage 2 — runtime
############################
FROM gcr.io/distroless/static-debian12:nonroot

ARG SERVICE_NAME
ARG GIT_SHA
ARG BUILD_TIME

LABEL org.opencontainers.image.title="${SERVICE_NAME}" \
      org.opencontainers.image.source="https://github.com/apollo-chora/chora-cli" \
      org.opencontainers.image.revision="${GIT_SHA}" \
      org.opencontainers.image.created="${BUILD_TIME}" \
      org.opencontainers.image.vendor="Chora Platform" \
      org.opencontainers.image.licenses="UNLICENSED" \
      io.chora.service="${SERVICE_NAME}" \
      io.chora.git-sha="${GIT_SHA}" \
      io.chora.build-time="${BUILD_TIME}"

WORKDIR /

COPY --from=builder /out/chora /chora

USER nonroot:nonroot
ENTRYPOINT ["/chora"]
