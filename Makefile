GOLANGCI_VERSION := v2.15.0

.PHONY: all test vet fmt fmt-check lint lint-install tidy-check check

all: check

test: ## Run unit tests
	go test -cover ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Run gofmt in place
	gofmt -s -w .

fmt-check: ## Fail if any file needs gofmt
	@echo "gofmt -s -l ."
	@out=$$(gofmt -s -l .); if [ -n "$$out" ]; then echo "gofmt needed:"; echo "$$out"; exit 1; fi

lint: ## Run golangci-lint
	golangci-lint run ./...

lint-install: ## Install pinned golangci-lint
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)

tidy-check: ## Fail if go mod tidy would change go.mod or go.sum
	go mod tidy -diff

check: fmt-check vet lint test tidy-check ## Run all checks
