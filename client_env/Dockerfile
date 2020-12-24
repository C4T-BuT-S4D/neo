FROM golang:1.15 as build

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o client cmd/client/main.go

FROM python:3.7

RUN apt-get update && \
    apt-get install -y --no-install-recommends python-z3

COPY client_env/requirements.txt /requirements.txt
RUN pip install -r /requirements.txt

COPY --from=build /app/client /neo

WORKDIR /sploits
ENTRYPOINT ["/neo", "--config", "/work/config.yml"]
