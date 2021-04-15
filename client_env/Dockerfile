FROM golang:1.16 as build

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -ldflags="-s -w" -o client cmd/client/main.go

FROM python:3.9

ENV PYTHONUNBUFFERED 1
ENV PWNLIB_NOTERM 1

RUN apt-get update && apt-get install -y \
        dnsutils \
        netcat \
        net-tools \
    && rm -rf /var/lib/apt/lists/*

COPY client_env/requirements.txt /requirements.txt
RUN pip install -r /requirements.txt

COPY --from=build /app/client /usr/local/bin/neo

WORKDIR /work
ENTRYPOINT ["/bin/bash", "-c"]
