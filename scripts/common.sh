#!/bin/bash

set -e

export GOPATH=`pwd`

# Set vscode gopath settings
if [ ! -f ".vscode/settings.json" ]; then
	mkdir -p .vscode
	echo -n "{
		\"go.gopath\": \"`pwd`\"
	}" > .vscode/settings.json
fi

# Libraries
echo "Checking libraries..."
go get github.com/joho/godotenv
go get gopkg.in/redis.v2
go get github.com/aws/aws-sdk-go/aws
go get github.com/aws/aws-sdk-go/service/s3

# Dev deps
echo "Checking dev deps..."
go get github.com/ramya-rao-a/go-outline
go get github.com/mdempsky/gocode
go get github.com/uudashr/gopkgs/v2/cmd/gopkgs
go get github.com/acroca/go-symbols
go get golang.org/x/tools/cmd/guru
go get golang.org/x/tools/cmd/gorename
go get github.com/cweill/gotests/...
go get github.com/fatih/gomodifytags
go get github.com/josharian/impl
go get github.com/davidrjenni/reftools/cmd/fillstruct
go get github.com/haya14busa/goplay/cmd/goplay
go get github.com/godoctor/godoctor
go get github.com/go-delve/delve/cmd/dlv
go get github.com/stamblerre/gocode
go get github.com/rogpeppe/godef
go get golang.org/x/tools/cmd/goimports
go get golang.org/x/lint/golint
