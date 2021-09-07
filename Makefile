
VERSION                 ?= $(shell git describe --tags --always --dirty)
RELEASE_VERSION         ?= $(shell git describe --abbrev=0)
PROJECT_NAME			= lora-connector

LDFLAGS                	?= -X main.Version=$(VERSION) -w -s
GO_ENV                  = CGO_ENABLED=0

GO_BUILD                = $(GO_ENV) go build -ldflags "$(LDFLAGS)"
GO_TEST                 = $(GO_ENV) go test -cover -v

.PHONY: clean build test

$(PROJECT_NAME):
	$(GO_BUILD) -o $(PROJECT_NAME) ./service

build: $(PROJECT_NAME)

test:
	$(GO_TEST) ./...

clean: 
	rm -f ./$(PROJECT_NAME)

rebuild: clean build
