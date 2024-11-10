FROM golang:1.23-alpine AS build

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
        go build \
            -trimpath \
            -ldflags="-s -w" \
            -o reaper \
            cmd/reaper/main.go && \
        upx --lzma -9 client && \
        upx --lzma -9 reaper

FROM ubuntu:24.04 AS image-full

ENV PIP_BREAK_SYSTEM_PACKAGES=1

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG DEBIAN_FRONTEND=noninteractive
RUN --mount=type=cache,sharing=private,target=/var/cache/apt \
        apt-get update && \
        apt-get install -y \
            build-essential \
            curl \
            dbus \
            dnsutils \
            file \
            gcc \
            libblas-dev \
            libjpeg8-dev \
            liblapack-dev \
            libmpc-dev \
            libmpfr-dev \
            libpng-dev \
            libssl-dev \
            net-tools \
            netcat-openbsd \
            python-is-python3 \
            python3 \
            python3-dev \
            python3-pip \
            python3-venv \
            vim \
            wget && \
        rm -rf /var/lib/apt/lists/*

COPY client_env/requirements.txt /requirements.txt
RUN --mount=type=cache,target=/root/.cache/pip \
        pip install -r /requirements.txt

COPY --from=build /app/client /usr/local/bin/neo
COPY --from=build /app/reaper /usr/local/bin/reaper

COPY client_env/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /work
ENTRYPOINT ["/entrypoint.sh"]

FROM image-full AS image-full-sage

RUN --mount=type=cache,sharing=private,target=/var/cache/apt \
        apt-get update && \
        apt-get install -y \
            sagemath && \
        rm -rf /var/lib/apt/lists/*

FROM alpine:3.20 AS image-light

RUN apk add bash dbus

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

COPY --from=build /app/client /usr/local/bin/neo
COPY --from=build /app/reaper /usr/local/bin/reaper

COPY client_env/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /work
ENTRYPOINT ["/entrypoint.sh"]
