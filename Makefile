# If you update this file, please follow
# https://suva.sh/posts/well-documented-makefiles

.DEFAULT_GOAL := help

PACKAGE = $(shell go list -m)
GIT_COMMIT_HASH = $(shell git rev-parse HEAD)
GIT_VERSION = $(shell git describe --tags --always --dirty)
BUILD_TIME = $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BINARY_NAME = ai-cli
LD_FLAGS = -s -w \
	-X '$(PACKAGE)/pkg/version.CommitHash=$(GIT_COMMIT_HASH)' \
	-X '$(PACKAGE)/pkg/version.Version=$(GIT_VERSION)' \
	-X '$(PACKAGE)/pkg/version.BuildTime=$(BUILD_TIME)' \
	-X '$(PACKAGE)/pkg/version.BinaryName=$(BINARY_NAME)'
COMMON_BUILD_ARGS = -ldflags "$(LD_FLAGS)"

GOLANGCI_LINT = $(shell pwd)/.work/tools/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.2.2

OSES = darwin linux windows
ARCHS = amd64 arm64

NPM_PACKAGE = npm-ai-cli

CLEAN_TARGETS :=
CLEAN_TARGETS += '$(BINARY_NAME)'
CLEAN_TARGETS += $(foreach os,$(OSES),$(foreach arch,$(ARCHS),$(BINARY_NAME)-$(os)-$(arch)$(if $(findstring windows,$(os)),.exe,)))
CLEAN_TARGETS += $(foreach os,$(OSES),$(foreach arch,$(ARCHS),./npm/$(NPM_PACKAGE)-$(os)-$(arch)/))
CLEAN_TARGETS += ./npm/$(NPM_PACKAGE)/.npmrc ./npm/$(NPM_PACKAGE)/LICENSE ./npm/$(NPM_PACKAGE)/package.json ./npm/$(NPM_PACKAGE)/README.md
CLEAN_TARGETS += $(BINARY_NAME)-darwin-*.tar.gz $(BINARY_NAME)-linux-*.tar.gz $(BINARY_NAME)-windows-*.zip


# GIT_TAG_VERSION should not append the -dirty flag
GIT_TAG_VERSION ?= $(shell echo $(shell git describe --tags --always) | sed 's/^v//')

# The help will print out all targets with their descriptions organized bellow their categories. The categories are represented by `##@` and the target descriptions by `##`.
# The awk commands is responsible to read the entire set of makefiles included in this invocation, looking for lines of the file as xyz: ## something, and then pretty-format the target and help. Then, if there's a line with ##@ something, that gets pretty-printed as a category.
# More info over the usage of ANSI control characters for terminal formatting: https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info over awk command: http://linuxcommand.org/lc3_adv_awk.php
#
# Notice that we have a little modification on the awk command to support slash in the recipe name:
# origin: /^[a-zA-Z_0-9-]+:.*?##/
# modified /^[a-zA-Z_0-9\/\.-]+:.*?##/
.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9\/\.-]+:.*?##/ { printf "  \033[36m%-22s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean up all build artifacts
	rm -rf $(CLEAN_TARGETS)

.PHONY: build
build: clean tidy format lint ## Build the project
	go build $(COMMON_BUILD_ARGS) -o $(BINARY_NAME)$(if $(findstring windows,$(shell go env GOOS)),.exe,) ./cmd/ai-cli

.PHONY: build-all-platforms
build-all-platforms: clean tidy format lint ## Build the project for all platforms
	$(foreach os,$(OSES),$(foreach arch,$(ARCHS), \
		GOOS=$(os) GOARCH=$(arch) go build $(COMMON_BUILD_ARGS) -o $(BINARY_NAME)-$(os)-$(arch)$(if $(findstring windows,$(os)),.exe,) ./cmd/ai-cli; \
	))

.PHONY: compress-binaries
compress-binaries: ## Compress binaries for distribution (.tar.gz for *nix, .zip for Windows)
	@echo "Compressing binaries..."
	@for os in darwin linux; do \
		for arch in $(ARCHS); do \
			if [ -f $(BINARY_NAME)-$$os-$$arch ]; then \
				tar -czf $(BINARY_NAME)-$$os-$$arch.tar.gz $(BINARY_NAME)-$$os-$$arch; \
				echo "Created $(BINARY_NAME)-$$os-$$arch.tar.gz"; \
			fi; \
		done; \
	done
	@for arch in $(ARCHS); do \
		if [ -f $(BINARY_NAME)-windows-$$arch.exe ]; then \
			zip -q $(BINARY_NAME)-windows-$$arch.zip $(BINARY_NAME)-windows-$$arch.exe; \
			echo "Created $(BINARY_NAME)-windows-$$arch.zip"; \
		fi; \
	done

.PHONY: test
test: ## Run the tests
	go test -count=1 -v ./...

.PHONY: format
format: ## Format the code
	go fmt ./...

.PHONY: tidy
tidy: ## Tidy up the go modules
	go mod tidy

.PHONY: golangci-lint
golangci-lint: ## Download and install golangci-lint if not already installed
ifeq ($(OS),Windows_NT)
	@echo "Skipping lint on Windows, delegating to CI/CD pipeline"
else
	@[ -f $(GOLANGCI_LINT) ] || { \
	set -e ;\
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell dirname $(GOLANGCI_LINT)) $(GOLANGCI_LINT_VERSION) ;\
	}
endif

.PHONY: lint
lint: golangci-lint ## Lint the code
ifeq ($(OS),Windows_NT)
	@echo "Skipping lint on Windows, delegating to CI/CD pipeline"
else
	$(GOLANGCI_LINT) run --verbose
endif

.PHONY: npm-copy-binaries
npm-copy-binaries: build-all-platforms ## Copy the binaries to each npm package
	$(foreach os,$(OSES),$(foreach arch,$(ARCHS), \
		EXECUTABLE=./$(BINARY_NAME)-$(os)-$(arch)$(if $(findstring windows,$(os)),.exe,); \
		NPM_EXECUTABLE=$(NPM_PACKAGE)-$(os)-$(arch)$(if $(findstring windows,$(os)),.exe,); \
		DIRNAME=$(NPM_PACKAGE)-$(os)-$(arch); \
		mkdir -p ./npm/$$DIRNAME/bin; \
		cp $$EXECUTABLE ./npm/$$DIRNAME/bin/$$NPM_EXECUTABLE; \
	))

MAIN_PACKAGE_JSON=./npm/$(NPM_PACKAGE)/package.json
.PHONY: npm-copy-project-files
npm-copy-project-files: npm-copy-binaries ## Copy the project files to the main npm package and generate all package.json files
	cp README.md LICENSE ./npm/$(NPM_PACKAGE)/
	@echo '{"name": "$(NPM_PACKAGE)",' > $(MAIN_PACKAGE_JSON)
	@echo '"version": "$(GIT_TAG_VERSION)",' >> $(MAIN_PACKAGE_JSON)
	@echo '"description": "AI CLI is a command line interface for AI services.",' >> $(MAIN_PACKAGE_JSON)
	@echo '"main": "./bin/index.js",' >> $(MAIN_PACKAGE_JSON)
	@echo '"bin": {"$(NPM_PACKAGE)": "bin/index.js"},' >> $(MAIN_PACKAGE_JSON)
	@echo '"optionalDependencies": {' >> $(MAIN_PACKAGE_JSON)
	@for os in $(OSES); do \
		for arch in $(ARCHS); do \
			if [ "$$os" = "$(lastword $(OSES))" ] && [ "$$arch" = "$(lastword $(ARCHS))" ]; then \
				echo "  \"$(NPM_PACKAGE)-$$os-$$arch\": \"$(GIT_TAG_VERSION)\""; \
			else \
				echo "  \"$(NPM_PACKAGE)-$$os-$$arch\": \"$(GIT_TAG_VERSION)\","; \
			fi \
		done; \
	done >> $(MAIN_PACKAGE_JSON)
	@echo '},' >> $(MAIN_PACKAGE_JSON)
	@echo '"repository": {"type": "git", "url": "git+https://github.com/manusa/ai-cli.git"}' >> $(MAIN_PACKAGE_JSON)
	@echo '}' >> $(MAIN_PACKAGE_JSON)
	$(foreach os,$(OSES),$(foreach arch,$(ARCHS), \
		OS_PACKAGE_JSON=./npm/$(NPM_PACKAGE)-$(os)-$(arch)/package.json; \
		echo '{"name": "$(NPM_PACKAGE)-$(os)-$(arch)",' > $$OS_PACKAGE_JSON; \
		echo '"version": "$(GIT_TAG_VERSION)",' >> $$OS_PACKAGE_JSON; \
		echo '"repository": {"type": "git", "url": "git+https://github.com/manusa/ai-cli.git"},' >> $$OS_PACKAGE_JSON; \
		echo '"os": ["$(os)"],' >> $$OS_PACKAGE_JSON; \
		NPM_ARCH="$(arch)"; \
		if [ "$$NPM_ARCH" = "amd64" ]; then NPM_ARCH="x64"; fi; \
		echo '"cpu": ["'$$NPM_ARCH'"]' >> $$OS_PACKAGE_JSON; \
		echo '}' >> $$OS_PACKAGE_JSON; \
	))

.PHONY: npm-publish
npm-publish: npm-copy-project-files ## Publish the npm packages
	$(foreach os,$(OSES),$(foreach arch,$(ARCHS), \
		DIRNAME="$(NPM_PACKAGE)-$(os)-$(arch)"; \
		cd npm/$$DIRNAME; \
		npm publish --tag latest; \
		cd ../..; \
	))
	cd npm/$(NPM_PACKAGE) && npm publish --tag latest

.PHONY: python-publish
python-publish: ## Publish the python packages
	cd ./python && \
	sed -i "s/version = \".*\"/version = \"$(GIT_TAG_VERSION)\"/" pyproject.toml && \
	uv build && \
	uv publish
