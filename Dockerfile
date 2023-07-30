FROM golang:1.20-alpine as build

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
            -ldflags="-w -s" \
            -o neo_server \
            cmd/server/main.go

FROM alpine

COPY --from=build /app/neo_server /neo_server

CMD ["/neo_server", "--config", "/config.yml"]
