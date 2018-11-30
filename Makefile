TARGET = kubernetes-oomkill-exporter
GOTARGET = github.com/sapcc/$(TARGET)
BUILDMNT = /go/src/$(GOTARGET)
REGISTRY ?= sapcc
VERSION ?= 0.1.0
IMAGE = $(REGISTRY)/$(BIN)
BUILD_IMAGE ?= golang:1.11-alpine3.8
DOCKER ?= docker
DIR := ${CURDIR}

ifneq ($(VERBOSE),)
VERBOSE_FLAG = -v
endif
TESTARGS ?= $(VERBOSE_FLAG) -timeout 60s
TEST_PKGS ?= $(GOTARGET)/...
TEST = CGO_ENABLED=0 go test $(TEST_PKGS) $(TESTARGS)
VET_PKGS ?= $(GOTARGET)/...
VET = CGO_ENABLED=0 go vet $(VET_PKGS)

DOCKER_BUILD ?= $(DOCKER) run --rm -v $(DIR):$(BUILDMNT) -w $(BUILDMNT) $(BUILD_IMAGE) /bin/sh -c

all: container

container:
	$(DOCKER_BUILD) 'go build'
	$(DOCKER) build -t $(REGISTRY)/$(TARGET):latest -t $(REGISTRY)/$(TARGET):$(VERSION) .

push:
	$(DOCKER) push $(REGISTRY)/$(TARGET):latest
	$(DOCKER) push $(REGISTRY)/$(TARGET):$(VERSION)

test:
	$(DOCKER_BUILD) '$(TEST)'

vet:
	$(DOCKER_BUILD) '$(VET)'

.PHONY: all local container push

clean:
	rm -f $(TARGET)
	$(DOCKER) rmi $(REGISTRY)/$(TARGET):latest
	$(DOCKER) rmi $(REGISTRY)/$(TARGET):$(VERSION)
