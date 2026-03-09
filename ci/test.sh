#!/bin/sh
set -eu

gofmt -w $(find . -name '*.go' -not -path './.git/*')
git diff --exit-code
go test ./...
