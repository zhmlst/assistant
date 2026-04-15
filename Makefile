include common.mk

.PHONY: all
all: gen

$(GOPATH)/bin/buf:
	go install github.com/bufbuild/buf/cmd/buf@latest

$(GOPATH)/bin/protoc-gen-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

$(GOPATH)/bin/protoc-gen-go-grpc:
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: gen
gen: $(GOPATH)/bin/buf $(GOPATH)/bin/protoc-gen-go $(GOPATH)/bin/protoc-gen-go-grpc
	buf generate
	$(MAKE) -C conversation gen

.PHONY: up
up: all
	$(DOCKER) compose up --build -d

.PHONY: run
run: all
	$(DOCKER) compose up --build

.PHONY: down
down:
	$(DOCKER) compose down

.PHONY: install
install: all
	mkdir -p /opt/assistant
	cp -r . /opt/assistant
	cp assistant.service /etc/systemd/system/

.PHONY: uninstall
uninstall:
	systemctl disable --now assistant.service
	rm -f /etc/systemd/system/assistant.service
