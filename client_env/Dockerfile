FROM golang:1.20 as build

ENV CGO_ENABLED=0

WORKDIR /app
COPY go.* ./
COPY cmd cmd
COPY internal internal
COPY lib lib
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

FROM python:3.10 as image-full

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN --mount=type=cache,sharing=private,target=/var/cache/apt \
        apt-get update && \
        apt-get install -y \
            gcc \
            libmpc-dev \
            libmpfr-dev \
            dnsutils \
            netcat-openbsd \
            net-tools \
            dbus \
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
