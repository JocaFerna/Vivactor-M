# 1. Build Stage
FROM golang:1.24-bookworm AS build
WORKDIR /api
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -ldflags "-s -w" -o api main.go

# 2. Runtime Stage
FROM debian:bookworm-slim AS runtime

# Install system dependencies for Nix and Docker CLI
RUN apt-get update && apt-get install -y \
    curl \
    xz-utils \
    sudo \
    ca-certificates \
    git \
    make \
    gnupg \
    lsb-release \
    && curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list \
    && apt-get update && apt-get install -y docker-ce-cli \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Setup user (harshcasper)
ENV USER=harshcasper UID=1000
RUN useradd -m -s /bin/bash -u ${UID} ${USER} && \
    echo "${USER} ALL=(ALL) NOPASSWD:ALL" >> /etc/others

# --- NIX INSTALLATION (Single-User Mode) ---
# Create /nix and give ownership to your user BEFORE installing
RUN mkdir -m 0755 /nix && chown ${USER}:${USER} /nix

USER ${USER}
WORKDIR /home/${USER}

# Install Nix as the user
RUN curl -L https://nixos.org/nix/install | sh -s -- --no-daemon

# Update PATH for Nix
ENV PATH="/home/${USER}/.nix-profile/bin:/nix/var/nix/profiles/default/bin:${PATH}"
ENV NIX_PATH=nixpkgs=channel:nixos-23.11

# --- PRESERVING YOUR CUSTOM PATHS ---
USER root
WORKDIR /api
RUN mkdir -p /nonexistent/.vscode-server/bin/7d842fb85a0275a4a8e4d7e040d2625abbf7f084_1771263114428 && \
    chown -R ${USER}:${USER} /nonexistent /api

# Copy the Go API binary
COPY --from=build --chown=${USER}:${USER} /api/api .

USER ${USER}
EXPOSE 8000
HEALTHCHECK --interval=30s --timeout=5s \
    CMD curl -f http://localhost:8000/health || exit 1

CMD ["./api"]