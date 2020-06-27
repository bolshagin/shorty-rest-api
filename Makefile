.PHONY: build
build:
	go build -v ./cmd/apiserver

.PHONY: test
test:
	go test -count=1 -v -race -timeout 30s ./...

.DEFAULT_GOAL := build
