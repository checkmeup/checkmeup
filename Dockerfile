# ── Stage 1: Build Vue frontend (native arch — output is platform-independent JS/CSS) ─
FROM --platform=${BUILDPLATFORM} oven/bun:1.3-alpine AS frontend
ARG APP_VERSION=dev
WORKDIR /app

COPY apps/web/package.json ./
RUN bun install --ignore-scripts

COPY apps/web .
ENV VITE_APP_VERSION=$APP_VERSION
RUN ./node_modules/.bin/vite build

# ── Stage 2: Build Go API (native arch + cross-compile target) ──────────────
FROM --platform=${BUILDPLATFORM} golang:1.25-bookworm AS backend
ARG APP_VERSION=dev
WORKDIR /app

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-X main.version=${APP_VERSION}" -o /api ./cmd/api

# ── Stage 3: Runtime (amd64 — matches server) ───────────────────────────────
FROM debian:bookworm-slim
ARG APP_VERSION=dev
ARG VCS_REF=dev
ARG BUILD_DATE=unknown
LABEL org.opencontainers.image.source="https://github.com/checkmeup/checkmeup" \
      org.opencontainers.image.title="checkmeup" \
      org.opencontainers.image.description="Cron, uptime, SSL, domain expiry, and port (TCP) monitors with execution logs, Telegram alerts, and white-label status pages for agencies." \
      org.opencontainers.image.url="https://checkmeup.net" \
      org.opencontainers.image.licenses="BUSL-1.1" \
      org.opencontainers.image.vendor="Andrew Molyuk" \
      org.opencontainers.image.version="${APP_VERSION}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.created="${BUILD_DATE}"

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=backend /api /api
COPY --from=frontend /app/dist /static
COPY apps/api/migrations /migrations

ENV MIGRATIONS_DIR=/migrations \
    STATIC_DIR=/static \
    PORT=8080

RUN groupadd --gid 1001 checkmeup \
    && useradd --uid 1001 --gid checkmeup --no-create-home --shell /usr/sbin/nologin checkmeup
USER checkmeup

EXPOSE 8080
CMD ["/api"]
