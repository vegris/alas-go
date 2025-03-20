.PHONY: format check

format:
	go fmt ./...

check:
	@echo "Checking formatting"
	@test -z "`gofmt -l .`" || (echo "Some files are not formatted" && exit 1)
	go vet ./...
