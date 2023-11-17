SHELL := /bin/bash

IMAGE := ghcr.io/c4t-but-s4d/neo_env:latest
TARGET := image-full
CONTAINER_NAME := neo_env

NEED_COMMANDS := curl wget dig nc file nslookup ifconfig python3 pip3 vim
NEED_PACKAGES := pymongo pymysql psycopg2 redis z3 secrets checklib requests pwn numpy bs4 hashpumpy dnslib regex lxml gmpy2 sympy grequests websocket

.PHONY: lint-go
lint-go:
	golangci-lint run -v --config .golangci.yml

.PHONY: lint-proto
lint-proto:
	cd proto && buf lint

.PHONY: lint
lint: lint-go lint-proto

.PHONY: goimports
goimports:
	gofancyimports fix --local github.com/c4t-but-s4d/neo -w $(shell find . -type f -name '*.go' -not -path "./proto/*")

.PHONY: test
test:
	go test -race -timeout 1m ./...

.PHONY: validate
validate: lint test

.PHONY: proto
proto:
	cd proto && buf generate
	cd front && ./add_ts_ignore.sh

.PHONY: test-cov
test-cov:
	go test -race -timeout 1m -coverprofile=coverage.txt -covermode=atomic ./...

.PHONY: build-image
build-image:
	docker build -t "${IMAGE}" --target "${TARGET}" -f client_env/Dockerfile .

.PHONY: test-image
test-image:
	@for cmd in $(NEED_COMMANDS) ; do \
  		echo -n "checking for command $$cmd... "; \
		if docker run --rm --entrypoint /bin/bash "${IMAGE}" which "$$cmd" >/dev/null; then \
			echo "ok"; \
		else \
			echo "Command $$cmd not found in image"; \
			exit 1; \
		fi \
	done

	@for pkg in $(NEED_PACKAGES) ; do \
  		echo -n "checking for python package $$pkg... "; \
		if docker run --rm --entrypoint /bin/bash "${IMAGE}" -c "python3 -c 'import $$pkg'" >/dev/null; then \
			echo "ok"; \
		else \
			echo "Command $$cmd not found in image"; \
			exit 1; \
		fi \
	done

.PHONY: push-image
push-image:
	docker push "${IMAGE}"

.PHONY: prepare-image
prepare-image: build-image test-image

.PHONY: release-image
release-image: prepare-image push-image

.PHONY: cleanup-release-all
cleanup-release-all: cleanup-release
	rm -rf dist

# To run before & after
.PHONY: cleanup-release
cleanup-release:
	rm -f exploits/.keep
	rmdir exploits || :

.PHONY: setup-release
setup-release: cleanup-release-all
	@echo "[*] Preparing aux dirs"
	mkdir exploits
	touch exploits/.keep

.PHONY: release-dry-run
release-dry-run:
	goreleaser --snapshot --skip-publish --clean

.PHONY: test-release
test-release: setup-release release-dry-run cleanup-release
