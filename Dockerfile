FROM golang:1.17-alpine3.14 as build

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd cmd
COPY internal internal
COPY lib lib
COPY pkg pkg
RUN CGO_ENABLED=0 go build -o neo_server cmd/server/main.go

FROM alpine:3.14

COPY --from=build /app/neo_server /neo_server

CMD ["/neo_server", "--config", "/config.yml"]
