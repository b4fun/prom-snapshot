# Release version
RELEASE ?= latest
# Image URL to use all building/pushing image targets
IMG ?= b4fun/frpcontroller:${RELEASE}

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: prom-snapshot-server

prom-snapshot-server: fmt vet
	go build -o bin/prom-snapshot-server ./cmd/server

fmt:
	go fmt ./...

vet:
	go vet ./...