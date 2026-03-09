#!/bin/sh
set -eu

VERSION="${1:-dev}"

go test ./...
tar czf "saga-${VERSION}.tar.gz" .
