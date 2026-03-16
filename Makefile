VERSION=0.4.0
BINARY=terraform-provider-wayscloud
GOBIN=$(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN=$(shell go env GOPATH)/bin
endif

OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)
INSTALL_DIR=~/.terraform.d/plugins/registry.terraform.io/wayscloud/wayscloud/$(VERSION)/$(OS_ARCH)

.PHONY: build test testacc install clean fmt vet lint

build:
	go build -o $(BINARY) .

test:
	go test -count=1 -race ./...

testacc:
	TF_ACC=1 go test -v -count=1 -timeout 60m ./internal/provider/

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/

clean:
	rm -f $(BINARY)

fmt:
	gofmt -s -w .

vet:
	go vet ./...

lint: fmt vet
	@echo "Lint complete"
