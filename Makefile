BINARY := hr
GO := go

.PHONY: build test lint fmt vet clean setup-hooks

build:
	$(GO) build -o $(BINARY) .

test:
	$(GO) test ./... -v

lint:
	golangci-lint run

fmt:
	$(GO) fmt ./...
	goimports -w .

vet:
	$(GO) vet ./...

clean:
	rm -f $(BINARY)

setup-hooks:
	git config core.hooksPath .githooks
	@echo "Git hooks configured to use .githooks/"
