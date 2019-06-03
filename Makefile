TARGET = kubernetes-oomkill-exporter
GOTARGET = github.com/sapcc/$(TARGET)
REGISTRY ?= sapcc
VERSION ?= 0.3.0
IMAGE = $(REGISTRY)/$(BIN)
DOCKER ?= docker

all: container

container:
	$(DOCKER) build -t $(REGISTRY)/$(TARGET):latest -t $(REGISTRY)/$(TARGET):$(VERSION) .

push:
	$(DOCKER) push $(REGISTRY)/$(TARGET):latest
	$(DOCKER) push $(REGISTRY)/$(TARGET):$(VERSION)

.PHONY: all local container push

clean:
	rm -f $(TARGET)
	$(DOCKER) rmi $(REGISTRY)/$(TARGET):latest
	$(DOCKER) rmi $(REGISTRY)/$(TARGET):$(VERSION)
