SHELL := /bin/bash

.PHONY: build run

SETUP_ENV := set -a && source .env && set +a

build:
	go build

run: build
	$(SETUP_ENV) && ./orcrist

test:
	$(SETUP_ENV) && go test ./...

