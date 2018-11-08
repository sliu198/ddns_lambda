#!/usr/bin/env bash
export GOPATH=$(pwd)
rm -rf target || true
cd src
GOOS=linux GOARCH=amd64 go build -o ../target/handler
cd ../target
zip handler.zip handler