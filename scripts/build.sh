#!/bin/bash

set -e

source $(dirname "$0")/common.sh

mkdir -p bin

echo "Compiling for linux/386"
GOOS=linux GOARCH=386 go build -o bin/rcsm_linux_386  src/main.go
echo "Compiling for linux/amd64"
GOOS=linux GOARCH=amd64 go build -o bin/rcsm_linux_amd64  src/main.go
echo "Compiling for linux/arm"
GOOS=linux GOARCH=arm go build -o bin/rcsm_linux_arm  src/main.go
echo "Compiling for linux/arm64"
GOOS=linux GOARCH=arm64 go build -o bin/rcsm_linux_arm64  src/main.go