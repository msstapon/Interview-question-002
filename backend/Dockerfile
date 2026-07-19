# syntax=docker/dockerfile:1.7
ARG GO_VERSION=1.24

# ─────────────── builder ───────────────
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/keys ./cmd/keys \
 && mkdir -p /out/secrets

# ─────────────── runtime (distroless, non-root) ───────────────
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder /out/api /app/api
COPY --from=builder /out/migrate /app/migrate
COPY --from=builder /out/keys /app/keys
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Pre-create /app/secrets owned by the nonroot uid (65532) so a named volume mounted
# here initialises with writable ownership — keygen can then write the JWT keypair.
COPY --from=builder --chown=65532:65532 /out/secrets /app/secrets

ENV TZ=UTC
EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/api"]
