FROM debian:stable-slim

# Install necessary packages for development (non-interactive)
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
    git gnupg openssh-client curl make zsh unzip ca-certificates \
    && curl -fsSL https://deb.nodesource.com/setup_24.x | bash - \
    && apt-get update \
    && apt-get install -y --no-install-recommends nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user for development
RUN useradd --create-home --user-group --shell /bin/zsh dev
USER dev
WORKDIR /home/dev

# Install Bun for the dev user
ENV BUN_INSTALL=/home/dev/.bun
ENV PATH="${BUN_INSTALL}/bin:${PATH}"
RUN curl -fsSL https://bun.sh/install | bash \
	&& ln -sf "${BUN_INSTALL}/bin/bun" "${BUN_INSTALL}/bin/bunx"

# Install Oh My Zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" || true

# Install global development tools
RUN bun install -g @anthropic-ai/claude-code oxfmt oxlint turbo
RUN turbo telemetry disable