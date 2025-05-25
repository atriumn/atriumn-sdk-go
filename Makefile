.PHONY: test test-verbose lint fmt clean tag help

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.1.0")
NEXT_VERSION ?= $(shell echo $(VERSION) | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g')
NEXT_MINOR_VERSION ?= $(shell echo $(VERSION) | awk -F. '{$$2 = $$2 + 1; $$3 = 0;} 1' | sed 's/ /./g')
NEXT_MAJOR_VERSION ?= $(shell echo $(VERSION) | awk -F. '{$$1 = substr($$1,2) + 1; $$2 = 0; $$3 = 0;} 1' | sed 's/ /./g' | sed 's/^/v/')

help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	go test ./...

test-verbose: ## Run tests with verbose output
	go test -v ./...

test-coverage: ## Run tests with coverage reporting
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

test-all: test ## Run all tests (alias for test target)

lint: ## Run linters
	golangci-lint run ./...

fmt: ## Run gofmt on all files
	find . -name "*.go" -not -path "./vendor/*" | xargs gofmt -s -w

clean: ## Clean build artifacts
	rm -rf coverage.out

tag-patch: ## Tag a new patch version (vX.Y.Z -> vX.Y.Z+1)
	@echo "Current version: $(VERSION), new version will be: $(NEXT_VERSION)"
	@read -p "Are you sure? [y/N] " confirm && [ $$confirm = "y" ] || exit 1
	@git tag -a $(NEXT_VERSION) -m "Release $(NEXT_VERSION)"
	@echo "Tagged $(NEXT_VERSION). Run 'git push origin $(NEXT_VERSION)' to push the tag."

tag-minor: ## Tag a new minor version (vX.Y.Z -> vX.Y+1.0)
	@echo "Current version: $(VERSION), new version will be: $(NEXT_MINOR_VERSION)"
	@read -p "Are you sure? [y/N] " confirm && [ $$confirm = "y" ] || exit 1
	@git tag -a $(NEXT_MINOR_VERSION) -m "Release $(NEXT_MINOR_VERSION)"
	@echo "Tagged $(NEXT_MINOR_VERSION). Run 'git push origin $(NEXT_MINOR_VERSION)' to push the tag."

tag-major: ## Tag a new major version (vX.Y.Z -> vX+1.0.0)
	@echo "Current version: $(VERSION), new version will be: $(NEXT_MAJOR_VERSION)"
	@read -p "Are you sure? [y/N] " confirm && [ $$confirm = "y" ] || exit 1
	@git tag -a $(NEXT_MAJOR_VERSION) -m "Release $(NEXT_MAJOR_VERSION)"
	@echo "Tagged $(NEXT_MAJOR_VERSION). Run 'git push origin $(NEXT_MAJOR_VERSION)' to push the tag."

release: ## Commit, tag and push a new release (specify VERSION=vX.Y.Z)
	@[ "$(VERSION)" != "" ] || ( echo "Usage: make release VERSION=vX.Y.Z"; exit 1 )
	@echo "Releasing version $(VERSION)"
	@git commit -am "Release $(VERSION)"
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin main
	@git push origin $(VERSION)
	@echo "Released $(VERSION)" 