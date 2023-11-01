FROM golang:1.21 as build

ENV CGO_ENABLED=0

WORKDIR /app
COPY go.* ./
COPY cmd cmd
COPY internal internal
COPY proto/go proto/go
COPY pkg pkg
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
        go build \
            -ldflags="-s -w" \
            -o client \
            cmd/client/main.go

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
        go build \
            -ldflags="-s -w" \
            -o reaper \
            cmd/reaper/main.go

FROM ubuntu:jammy as image-full

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG DEBIAN_FRONTEND=noninteractive
RUN --mount=type=cache,sharing=private,target=/var/cache/apt \
        apt-get update && \
        apt-get install -y \
            build-essential \
            dbus \
            dnsutils \
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
            vim && \
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

FROM image-full as image-full-sage

RUN --mount=type=cache,sharing=private,target=/var/cache/apt \
        apt-get update && \
        apt-get install -y \
            sagemath && \
        rm -rf /var/lib/apt/lists/*

FROM alpine:3.16 as image-light

RUN apk add bash dbus

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

COPY --from=build /app/client /usr/local/bin/neo
COPY --from=build /app/reaper /usr/local/bin/reaper

COPY client_env/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /work
ENTRYPOINT ["/entrypoint.sh"]
