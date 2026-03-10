#!/bin/sh
set -eu

VERSION="${1:-dev}"
DIST_DIR="dist"
PACKAGE_DIR="${DIST_DIR}/sg-${VERSION}"
ARCHIVE_PATH="${DIST_DIR}/sg-${VERSION}.tar.gz"

go test ./...
rm -rf "${PACKAGE_DIR}"
mkdir -p "${PACKAGE_DIR}"

go build -o "${PACKAGE_DIR}/sg" ./cmd/sg
cp README.md CHANGELOG.md "${PACKAGE_DIR}/"

tar czf "${ARCHIVE_PATH}" -C "${DIST_DIR}" "sg-${VERSION}"
