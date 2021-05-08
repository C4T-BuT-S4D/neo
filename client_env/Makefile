SHELL := /bin/bash

ACCOUNT := c4tbuts4d
IMAGE := c4tbuts4d/neo_env:latest
CONTAINER_NAME := neo_env

NEED_COMMANDS := curl wget dig nc file nslookup ifconfig python3 pip3
NEED_PACKAGES := pymongo pymysql psycopg2 redis z3 secrets checklib requests pwn numpy bs4 hashpumpy dnslib regex lxml gmpy2 sympy

.PHONY: test
test:
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

.PHONY: build
build:
	@echo "Make sure you're logged in as ${ACCOUNT}"
	cd .. && docker build -t "${IMAGE}" -f client_env/Dockerfile .

.PHONY: push
push:
	docker push "${IMAGE}"

.PHONY: prepare
prepare: build test

.PHONY: release
release: prepare push
