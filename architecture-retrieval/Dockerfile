# 1. Build Stage
FROM golang:1.24-bookworm AS build
WORKDIR /api
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -ldflags "-s -w" -o api main.go

# 2. Runtime Stage
FROM debian:bookworm-slim AS runtime

# Match host versions exactly
ENV DOCKER_VERSION=28.1.1
ENV COMPOSE_VERSION=v2.35.1

# Install system dependencies
# Added 'procps' here to provide the 'ps' command
RUN apt-get update && apt-get install -y \
    curl xz-utils sudo ca-certificates git gnupg lsb-release procps \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# 1. Install Docker CLI Static Binary (v28.1.1)
# This ensures the CLI inside the container matches your host daemon
RUN curl -fsSL "https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz" | tar -xzC /tmp \
    && mv /tmp/docker/docker /usr/local/bin/docker \
    && rm -rf /tmp/docker

# 2. Install Docker Compose V2 Plugin
# This fixes the "unknown command" error when running 'docker compose'
RUN mkdir -p /usr/local/lib/docker/cli-plugins/ \
    && curl -SL "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-linux-x86_64" -o /usr/local/lib/docker/cli-plugins/docker-compose \
    && chmod +x /usr/local/lib/docker/cli-plugins/docker-compose

# 3. Create a symlink for legacy "docker-compose" (optional but recommended)
RUN ln -s /usr/local/lib/docker/cli-plugins/docker-compose /usr/local/bin/docker-compose

# Setup user
ENV USER=harshcasper UID=1000
RUN useradd -m -s /bin/bash -u ${UID} ${USER} && \
    echo "${USER} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Nix Setup
RUN mkdir -m 0755 /nix && chown ${USER}:${USER} /nix
USER ${USER}
WORKDIR /home/${USER}

# Install Nix (No-daemon mode for container compatibility)
RUN curl -L https://nixos.org/nix/install | sh -s -- --no-daemon

# Paths for Nix environment
ENV PATH="/home/${USER}/.nix-profile/bin:/nix/var/nix/profiles/default/bin:${PATH}"
ENV NIX_PATH=nixpkgs=channel:nixos-23.11
RUN echo ". /home/${USER}/.nix-profile/etc/profile.d/nix.sh" >> ~/.bashrc

# App Setup
USER root
# Move Go binary to /usr/local/bin
COPY --from=build --chown=${USER}:${USER} /api/api /usr/local/bin/api
RUN chmod +x /usr/local/bin/api

# Create the app directory for runtime logs/data
WORKDIR /api
RUN chown -R ${USER}:${USER} /api

USER ${USER}
EXPOSE 8000

# Execute the Go binary
CMD ["api"]