include ../common.mk

.PHONY: all
all: image

.PHONY: image
image:
	$(DOCKER) build -f Dockerfile -t assistant-inference ..
