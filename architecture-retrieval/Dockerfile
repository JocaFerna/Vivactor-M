# 1. Build Stage
FROM golang:1.24-bookworm AS build
WORKDIR /api
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -ldflags "-s -w" -o api main.go

# 2. Runtime Stage
FROM debian:bookworm-slim AS runtime

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl xz-utils sudo ca-certificates git make gnupg lsb-release \
    && curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list \
    && apt-get update && apt-get install -y docker-ce-cli \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Setup user
ENV USER=harshcasper UID=1000
RUN useradd -m -s /bin/bash -u ${UID} ${USER} && \
    echo "${USER} ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Nix Setup
RUN mkdir -m 0755 /nix && chown ${USER}:${USER} /nix
USER ${USER}
WORKDIR /home/${USER}
RUN curl -L https://nixos.org/nix/install | sh -s -- --no-daemon

# Paths
ENV PATH="/home/${USER}/.nix-profile/bin:/nix/var/nix/profiles/default/bin:${PATH}"
ENV NIX_PATH=nixpkgs=channel:nixos-23.11
RUN echo ". /home/${USER}/.nix-profile/etc/profile.d/nix.sh" >> ~/.bashrc

# App Setup
USER root
# Move binary to /usr/local/bin so volumes don't overwrite it
COPY --from=build --chown=${USER}:${USER} /api/api /usr/local/bin/api
RUN chmod +x /usr/local/bin/api

# Create the app directory and set permissions
WORKDIR /api
RUN chown -R ${USER}:${USER} /api

USER ${USER}
EXPOSE 8000
# Execute the binary from the system path
CMD ["api"]