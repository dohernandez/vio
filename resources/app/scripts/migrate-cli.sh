#!/usr/bin/env bash

[ -z "$GO" ] && GO=go

# detecting GOPATH and removing trailing "/" if any
GOPATH="$(go env GOPATH)"
GOPATH=${GOPATH%/}

# adding GOBIN to PATH
[[ ":$PATH:" != *"$GOPATH/bin"* ]] && PATH=$PATH:"$GOPATH"/bin

# checking if migrate is available
if ! command -v migrate > /dev/null ; then \
    echo ">> installing migrate"; \
    PLATFORM=$(uname | tr '[:upper:]' '[:lower:]')
    curl -sL https://github.com/golang-migrate/migrate/releases/download/v4.5.0/migrate."$PLATFORM"-amd64.tar.gz | tar xvz -C /tmp/ \
      && mv /tmp/migrate."$PLATFORM"-amd64 "$GOPATH"/bin/migrate
fi
