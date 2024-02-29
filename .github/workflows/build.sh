#!/bin/sh

set -ex

go generate ./...

GOOS=${OS}
[ "$GOOS" = "macos" ] && GOOS=darwin
GOARCH="${ARCH}" GOOS="$GOOS" go build -tags netgo -o cftest ./cmd/cftest

mkdir -p build/cftest
cp README.md build/cftest
cp cftest build/cftest/cftest

cd build
tar -czf "cftest-${OS}-${ARCH}-${REF}.tgz" cftest
