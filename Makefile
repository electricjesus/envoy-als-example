.PHONY: all
all: build push

TAG ?= sethm/envoy-als-example:latest

.PHONY: build
build:
	docker build -t $(TAG) .

push:
	docker push $(TAG)