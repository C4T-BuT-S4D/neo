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
        upx --lzma -9 client

FROM ubuntu:noble AS image-full

ENV DEBIAN_FRONTEND=noninteractive \
    PIP_BREAK_SYSTEM_PACKAGES=1 \
    UV_SYSTEM_PYTHON=1

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN --mount=type=cache,sharing=private,target=/var/cache/apt \
    --mount=type=cache,sharing=private,target=/var/lib/apt \
        rm -f /etc/apt/apt.conf.d/docker-clean && \
        apt-get update && \
        apt-get install -y --no-install-recommends \
            build-essential \
            ca-certificates \
            curl \
            dbus \
            dnsutils \
            file \
            gcc \
            iputils-ping \
            libblas-dev \
            liblapack-dev \
            libjpeg-dev \
            libmpc-dev \
            libmpfr-dev \
            libpng-dev \
            netcat-openbsd \
            net-tools \
            openssl \
            python3 \
            python3-dev \
            python3-pip \
            python3-venv \
            tini \
            vim \
            wget && \
        rm -rf /var/lib/apt/lists/*

COPY --from=ghcr.io/astral-sh/uv:0.9.5 /uv /uvx /bin/

RUN python3 -m pip config set global.break-system-packages true

COPY client_env/requirements.txt /requirements.txt
RUN --mount=type=cache,target=/root/.cache/pip \
        uv pip install --system -r /requirements.txt

COPY --from=build /app/client /usr/local/bin/neo

COPY client_env/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /work
ENTRYPOINT ["/usr/bin/tini", "--", "/entrypoint.sh"]

FROM image-full AS image-full-sage

# Install Sage build dependencies
RUN --mount=type=cache,sharing=private,target=/var/cache/apt \
    --mount=type=cache,sharing=private,target=/var/lib/apt \
        apt-get update && \
        apt-get install -y --no-install-recommends \
            autoconf \
            automake \
            bc \
            binutils \
            bzip2 \
            g++ \
            gettext \
            git \
            gfortran \
            libtool \
            libreadline-dev \
            libssl-dev \
            libz-dev \
            m4 \
            make \
            patch \
            perl \
            pkg-config \
            tar \
            xz-utils && \
        rm -rf /var/lib/apt/lists/*

# Download and build the latest stable Sage from source
RUN cd /tmp && \
    git clone --depth 1 --branch master https://github.com/sagemath/sage.git && \
    cd sage && \
    git fetch --tags && \
    LATEST_TAG=$(git tag -l | grep -E '^[0-9]+\.[0-9]+(\.[0-9]+)?$' | sort -V | tail -n1) && \
    git checkout "$LATEST_TAG" && \
    ./configure && \
    make -j$(nproc) && \
    make install && \
    cd / && \
    rm -rf /tmp/sage

ENV PATH="/usr/local/sage:${PATH}"

FROM alpine:3.20 AS image-light

RUN apk add bash dbus tini

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

COPY --from=build /app/client /usr/local/bin/neo

COPY client_env/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

WORKDIR /work
ENTRYPOINT ["/sbin/tini", "--", "/entrypoint.sh"]
