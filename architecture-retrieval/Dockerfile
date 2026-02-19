# 1. Build Stage
FROM golang:1.24-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /api
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -ldflags "-s -w" -o api main.go

# 2. Use a modern, secure runtime image
FROM alpine:3.19 AS runtime

# FIX: Add dependencies required for VS Code Server to run on Alpine
RUN apk add --no-cache libstdc++ gcompat ca-certificates

WORKDIR /api

# Setup user for security
ENV USER=harshcasper UID=1000
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/home/${USER}" \
    --shell "/bin/sh" \
    --uid "${UID}" \
    "${USER}"


USER root

RUN mkdir -p /nonexistent && chown -R ${USER}:${USER} /nonexistent

RUN mkdir -p /nonexistent/.vscode-server/bin/7d842fb85a0275a4a8e4d7e040d2625abbf7f084_1771263114428 && chown -R ${USER}:${USER} /nonexistent/.vscode-server

# Ensure both the app directory and the new home directory belong to the user
RUN chown -R ${USER}:${USER} /api /home/${USER}

# Copy the binary from the build stage
COPY --from=build --chown=${USER}:${USER} /api/api .

USER ${USER}:${USER}

HEALTHCHECK --interval=30s --timeout=5s \
    CMD wget -nv -t1 --spider 'http://localhost:8000/health' || exit 1

EXPOSE 8000

CMD ["./api"]