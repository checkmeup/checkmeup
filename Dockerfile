# ── Stage 1: Build Vue frontend ────────────────────────────────────────────
FROM oven/bun:1.3-alpine AS frontend
WORKDIR /app

COPY apps/web/package.json ./
RUN bun install --ignore-scripts

COPY apps/web .
RUN bun run build

# ── Stage 2: Build Go API ───────────────────────────────────────────────────
FROM golang:1.25-bookworm AS backend
WORKDIR /app

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /api ./cmd/api

# ── Stage 3: Runtime ────────────────────────────────────────────────────────
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=backend /api /api
COPY --from=frontend /app/dist /static
COPY apps/api/migrations /migrations

ENV MIGRATIONS_DIR=/migrations \
    STATIC_DIR=/static \
    PORT=8080

EXPOSE 8080
CMD ["/api"]
