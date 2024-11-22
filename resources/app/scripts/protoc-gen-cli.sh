#!/usr/bin/env bash

[ -z "$GO" ] && GO=go

# detecting GOPATH and removing trailing "/" if any
GOPATH="$(go env GOPATH)"
GOPATH=${GOPATH%/}

# adding GOBIN to PATH
[[ ":$PATH:" != *"$GOPATH/bin"* ]] && PATH=$PATH:"$GOPATH"/bin

# checking if protoc-gen-go is available
if ! command -v protoc > /dev/null ; then \
    echo ">> installing protoc-gen-go... "; \
    $GO install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26;
    echo ">> installing protoc-gen-go-grpc... "; \
    $GO install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1;
fi

# checking if protoc-gen-grpc-gateway is available
if ! command -v protoc-gen-grpc-gateway > /dev/null ; then \
    echo ">> installing protoc-gen-grpc-gateway... "; \
    $GO install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway;
fi

# checking if protoc-gen-openapiv2 is available
if ! command -v protoc-gen-openapiv2 > /dev/null ; then \
    echo ">> installing protoc-gen-openapiv2... "; \
    $GO install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2;
fi

# checking if protoc-gen-openapiv2 is available
if ! command -v protoc-gen-openapiv2 > /dev/null ; then \
    echo ">> installing protoc-gen-openapiv2... "; \
    $GO install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2;
fi
