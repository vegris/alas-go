.PHONY: build run

build:
	go build

run: build
	env $$(cat .env | xargs) ./kiwi

test:
	env $$(cat .env | xargs) go test ./...

