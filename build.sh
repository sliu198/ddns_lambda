#!/usr/bin/env bash
export GOPATH=$(pwd)
cd src
GOOS=linux GOARCH=amd64 go build -o ../target/handler
cd ../target
zip handler.zip handler