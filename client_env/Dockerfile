FROM golang:1.25-alpine AS build

ENV CGO_ENABLED=0

RUN apk add upx

WORKDIR /app
COPY go.* ./
COPY cmd cmd
COPY internal internal
COPY pkg pkg
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o client \
    cmd/client/main.go && \
    upx --lzma -9 client

FROM archlinux:base-devel-20251019.0.436919 AS image-full

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Setup uv.
COPY --from=ghcr.io/astral-sh/uv:0.9.11 /uv /uvx /bin/
ENV UV_LINK_MODE=copy
ENV UV_SYSTEM_PYTHON=1

RUN pacman-key --refresh-keys && \
    pacman --noconfirm -Syu && \
    pacman --noconfirm -Sy \
    blas \
    curl \
    dbus \
    dnsutils \
    file \
    gcc \
    lapack \
    libjpeg-turbo \
    libmpc \
    libpng \
    mpfr \
    net-tools \
    openbsd-netcat \
    openssl \
    python \
    python-pip \
    python-virtualenv \
    uv \
    vim \
    wget \
    cryptominisat

ENV TINI_VERSION=v0.19.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /usr/bin/tini
RUN chmod +x /usr/bin/tini

COPY client_env/requirements.txt /requirements.txt
RUN --mount=type=cache,target=/root/.cache \
    uv pip install -r /requirements.txt

COPY --from=build /app/client /usr/local/bin/neo

COPY client_env/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /work
ENTRYPOINT ["/usr/bin/tini", "-s", "--", "/entrypoint.sh"]

FROM image-full AS image-full-sage

RUN pacman --noconfirm -Sy --overwrite '*' sagemath

FROM alpine:3.20 AS image-light

RUN apk add bash dbus

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

COPY --from=build /app/client /usr/local/bin/neo

COPY client_env/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /work
ENTRYPOINT ["/entrypoint.sh"]
