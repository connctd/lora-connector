
VERSION                 ?= $(shell git describe --tags --always --dirty)
RELEASE_VERSION         ?= $(shell git describe --abbrev=0)
PROJECT_NAME			= lora-connector

LDFLAGS                	?= -X main.Version=$(VERSION) -w -s
GO_ENV                  = CGO_ENABLED=0

GO_BUILD                = $(GO_ENV) go build -ldflags "$(LDFLAGS)"
GO_TEST                 = $(GO_ENV) go test -cover -v

GCR_PROJECT_ID 			?= molten-mariner-162315
GCR_IMAGE 				?= eu.gcr.io/$(GCR_PROJECT_ID)/connctd/$(PROJECT_NAME)

.PHONY: clean build test docker docker/push

$(PROJECT_NAME): test
	$(GO_BUILD) -o $(PROJECT_NAME) ./service

$(PROJECT_NAME)_linux_amd64: test
	GOARCH=amd64 GOOS=linux $(GO_BUILD) -o $(PROJECT_NAME)_linux_amd64 ./service

build: $(PROJECT_NAME)

test:
	$(GO_TEST) ./...

clean: 
	rm -f ./$(PROJECT_NAME)

docker: $(PROJECT_NAME)_linux_amd64
	docker build \
		--file Dockerfile \
		--rm \
		--tag "$(GCR_IMAGE):$(VERSION)" .

docker/push: docker
	docker push "$(GCR_IMAGE):$(VERSION)"

rebuild: clean build
