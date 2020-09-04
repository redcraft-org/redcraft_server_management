#!/bin/bash

set -e

export GOPATH=`pwd`

# Set vscode settings
if [ ! -f ".vscode/settings.json" ]; then
	mkdir -p .vscode
	echo -n "{
		\"go.gopath\": \"`pwd`\"
	}" > .vscode/settings.json
fi

go get github.com/joho/godotenv
go get gopkg.in/redis.v2
go get github.com/aws/aws-sdk-go/aws
go get github.com/aws/aws-sdk-go/service/s3

mkdir -p bin

GOOS=linux GOARCH=386 go build -o bin/rcsm_linux_386 src/main.go
GOOS=linux GOARCH=amd64 go build -o bin/rcsm_linux_amd64 src/main.go
GOOS=linux GOARCH=arm go build -o bin/rcsm_linux_arm src/main.go
GOOS=linux GOARCH=arm64 go build -o bin/rcsm_linux_arm64 src/main.go