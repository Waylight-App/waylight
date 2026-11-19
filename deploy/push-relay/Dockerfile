# Image of the push relay (ADR 0025, docs/push-relay.md). Build from the
# repository root: docker build -f deploy/push-relay/Dockerfile .
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/push-relay ./cmd/push-relay \
    && mkdir -p /out/data

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/push-relay /push-relay
COPY --from=build --chown=65532:65532 /out/data /data
ENV PUSH_RELAY_ADDR=:8080 \
    PUSH_RELAY_DB=/data/push-relay.db
EXPOSE 8080
VOLUME ["/data"]
USER nonroot:nonroot
ENTRYPOINT ["/push-relay"]
