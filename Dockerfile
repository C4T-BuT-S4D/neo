FROM golang:1.21-alpine as build

ENV CGO_ENABLED=0

WORKDIR /app
COPY go.* ./
COPY cmd cmd
COPY internal internal
COPY pkg pkg
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
        go build \
            -ldflags="-w -s" \
            -o neo_server \
            cmd/server/main.go

FROM node:20-slim AS front-base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

COPY front /app
WORKDIR /app

FROM front-base AS front-build
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile
RUN pnpm run build

FROM alpine

WORKDIR /app
COPY --from=build /app/neo_server neo_server
COPY --from=front-build /app/dist front/dist

CMD ["./neo_server", "--config", "/config.yml"]
