#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

rm -rf artifacts/*

export CGO_ENABLED=0
export GOARCH="${ARCH}"
export GOOS="${OS}"

# build linux
env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o artifacts/autonetpol main.go