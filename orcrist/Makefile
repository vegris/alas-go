SHELL := /bin/bash

.PHONY: build run format check

SETUP_ENV := set -a && source ../.env && set +a

build:
	go build

run: build
	$(SETUP_ENV) && ./orcrist

format:
	go fmt ./...

check:
	@echo "Checking formatting"
	@test -z "`gofmt -l .`" || (echo "Some files are not formatted" && exit 1)
	go vet ./...
