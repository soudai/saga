.PHONY: build test fmt

build:
	mkdir -p bin && go build -o bin/sg ./cmd/sg

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')
