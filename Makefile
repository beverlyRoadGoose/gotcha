include project.mk

.PHONY: build test

build:
	go build -v ./...
test:
	go test ./... -coverprofile=coverage.txt -covermode=atomic
