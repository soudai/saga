.PHONY: build test fmt

build:
	mkdir -p bin && go build -o bin/saga ./cmd/saga

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')
