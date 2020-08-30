# Release version
RELEASE ?= latest
# Image URL to use all building/pushing image targets
IMG ?= b4fun/prom-snapshot-server:${RELEASE}

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

docker-build:
	GOOS=linux GOARCH=amd64 go build -o docker_bin/prom-snapshot-server-amd64 ./cmd/server
	docker build . -t ${IMG}

docker-push:
	docker push ${IMG}