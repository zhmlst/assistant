include ../common.mk

.PHONY: all
all: image

$(GOPATH)/bin/stringer:
	go install golang.org/x/tools/cmd/stringer@latest

.PHONY: gen
gen: $(GOPATH)/bin/stringer
	go generate ./...
	gofmt -w .
	go mod tidy

$(GOPATH)/bin/grpcui:
	go install github.com/fullstorydev/grpcui/cmd/grpcui@latest

.PHONY: ui
ui: $(GOPATH)/bin/grpcui
	grpcui -plaintext localhost:50050

.PHONY: image
image:
	$(DOCKER) build -f Dockerfile -t assistant-conversation ..
