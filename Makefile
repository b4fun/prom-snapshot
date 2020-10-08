# Release version
RELEASE ?= latest
# Docker image repo
REPO ?= b4fun

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: prom-snapshot-server snapshot-sidecar

prom-snapshot-server: fmt vet
	go build -o bin/prom-snapshot-server ./cmd/server

snapshot-sidecar: fmt vet
	go build -o bin/snapshot-sidecar

fmt:
	go fmt ./...

vet:
	go vet ./...

docker-build-prom-snapshot-server:
	GOOS=linux GOARCH=amd64 go build -o docker_bin/prom-snapshot-server ./cmd/server
	docker build docker_bin \
		-f ./cmd/server/Dockerfile \
		-t ${REPO}/prom-snapshot-server:${RELEASE}

docker-build-snapshot-sidecar:
	GOOS=linux GOARCH=amd64 go build -o docker_bin/snapshot-sidecar ./cmd/snapshot-sidecar
	docker build docker_bin \
		-f ./cmd/snapshot-sidecar/Dockerfile \
		-t ${REPO}/prom-snapshot-sidecar:${RELEASE}

docker-push-snapshot-sidecar:
	docker push ${REPO}/prom-snapshot-sidecar:${RELEASE}

docker-push-prom-snapshot-server:
	docker push ${REPO}/prom-snapshot-server:${RELEASE}

docker-build: docker-build-prom-snapshot-server docker-build-snapshot-sidecar

docker-push: docker-push-prom-snapshot-server docker-push-snapshot-sidecar