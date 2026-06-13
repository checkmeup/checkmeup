FROM debian:stable-slim

ARG GO_VERSION=1.25.11
ARG GOLANGCI_LINT_VERSION=v2.12.2

# System packages + Go (multi-arch)
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
       git=1:2.47.3-0+deb13u1 \
       gnupg=2.4.7-21+deb13u1 \
       openssh-client=1:10.0p1-7+deb13u4 \
       curl=8.14.1-2+deb13u3 \
       make=4.4.1-2 \
       zsh=5.9-8+b23 \
       unzip=6.0-29 \
       ca-certificates=20250419 \
    && curl -fsSL https://deb.nodesource.com/setup_24.x | bash - \
    && apt-get install -y --no-install-recommends nodejs=24.16.0-1nodesource1 \
    && GOARCH=$(dpkg --print-architecture) \
    && curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz" | tar -C /usr/local -xz \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV PATH="/usr/local/go/bin:${PATH}"

# Non-root dev user
RUN useradd --create-home --user-group --shell /bin/zsh dev
USER dev
WORKDIR /home/dev

ENV GOPATH="/home/dev/go"
ENV BUN_INSTALL="/home/dev/.bun"
ENV PATH="${GOPATH}/bin:${BUN_INSTALL}/bin:${PATH}"

# Bun
RUN curl -fsSL https://bun.sh/install | bash \
    && ln -sf "${BUN_INSTALL}/bin/bun" "${BUN_INSTALL}/bin/bunx"

# Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" || true

# JS global tools
RUN bun install -g @anthropic-ai/claude-code oxfmt oxlint turbo \
    && turbo telemetry disable

# Go tools
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION} \
    && go install github.com/air-verse/air@latest \
    && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest \
    && go install github.com/pressly/goose/v3/cmd/goose@latest \
    && go install github.com/jandelgado/gcov2lcov@latest
