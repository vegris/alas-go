SHELL := /bin/bash

.PHONY: up down build run

SETUP_ENV := set -a && source .env && set +a

up:
	docker compose up --wait

down:
	docker compose down

build:
	go build

run: build
	$(SETUP_ENV) && ./kiwi

test:
	$(SETUP_ENV) && go test ./...

