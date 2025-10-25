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

FROM archlinux:base-devel-20241110.0.278197 AS image-full

ENV PIP_BREAK_SYSTEM_PACKAGES=1 UV_BREAK_SYSTEM_PACKAGES=1

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN --mount=type=cache,sharing=private,target=/var/cache/pacman \
        pacman-key --refresh-keys && \
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
            wget

RUN python -m pip config set global.break-system-packages true

COPY client_env/requirements.txt /requirements.txt
RUN --mount=type=cache,target=/root/.cache/pip \
        uv pip install --system -r /requirements.txt

COPY --from=build /app/client /usr/local/bin/neo
COPY --from=build /app/reaper /usr/local/bin/reaper

COPY client_env/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /work
ENTRYPOINT ["/entrypoint.sh"]

FROM image-full AS image-full-sage

RUN --mount=type=cache,sharing=private,target=/var/cache/pacman \
        pacman --noconfirm -Sy --overwrite '*' \
            sagemath

FROM alpine:3.20 AS image-light

RUN apk add bash dbus

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

COPY --from=build /app/client /usr/local/bin/neo
COPY --from=build /app/reaper /usr/local/bin/reaper

COPY client_env/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /work
ENTRYPOINT ["/entrypoint.sh"]
