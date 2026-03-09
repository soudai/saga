.PHONY: build test fmt

build:
	go build ./cmd/saga

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')
