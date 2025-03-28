SHELL := /bin/bash

.PHONY: test format check

SETUP_ENV := set -a && source ../.env && set +a

test:
	$(SETUP_ENV) && go test -v -count 1

format:
	go fmt ./...

check:
	@echo "Checking formatting"
	@test -z "`gofmt -l .`" || (echo "Some files are not formatted" && exit 1)
	go vet ./...

