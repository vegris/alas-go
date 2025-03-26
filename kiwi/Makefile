SHELL := /bin/bash

.PHONY: build run test format check

SETUP_ENV := set -a && source ../.env && set +a

build:
	go build

run: build
	$(SETUP_ENV) && ./kiwi

test:
	$(SETUP_ENV) && go test ./...

format:
	go fmt ./...

check:
	@echo "Checking formatting"
	@test -z "`gofmt -l .`" || (echo "Some files are not formatted" && exit 1)
	go vet ./...
