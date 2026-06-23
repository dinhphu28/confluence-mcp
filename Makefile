APP_NAME := atlassian-mcp
VERSION ?= 0.1.0

OS ?= linux
ARCH ?= amd64

BUILD_DIR := build
DIST_DIR := dist

ifeq ($(OS),windows)
	EXT := .exe
	ARCHIVE_EXT := zip
else
	EXT :=
	ARCHIVE_EXT := tar.gz
endif

BIN_NAME := $(APP_NAME)$(EXT)
PACKAGE_NAME := $(APP_NAME)_$(VERSION)_$(OS)_$(ARCH)
PACKAGE_DIR := $(DIST_DIR)/$(PACKAGE_NAME)
ARCHIVE := $(DIST_DIR)/$(PACKAGE_NAME).$(ARCHIVE_EXT)

# Platform-specific install hint shipped as README.txt inside the package.
ifeq ($(OS),windows)
README_TXT := Atlassian MCP (Windows)\n\nOpen PowerShell as Administrator, then run:\n\n  .\\$(BIN_NAME) setup\n
else
README_TXT := Atlassian MCP\n\nRun setup:\n\n  ./$(BIN_NAME) setup\n
endif

.PHONY: build package release release-linux release-windows release-all publish clean

build:
	mkdir -p $(BUILD_DIR)/$(OS)_$(ARCH)

	CGO_ENABLED=0 \
	GOOS=$(OS) \
	GOARCH=$(ARCH) \
	go build \
	-ldflags="-s -w -X main.version=$(VERSION)" \
	-o $(BUILD_DIR)/$(OS)_$(ARCH)/$(BIN_NAME) ./cmd/$(APP_NAME)

package: build
	rm -rf $(PACKAGE_DIR)

	mkdir -p $(PACKAGE_DIR)

	cp $(BUILD_DIR)/$(OS)_$(ARCH)/$(BIN_NAME) $(PACKAGE_DIR)/

	printf '%b' '$(README_TXT)' > $(PACKAGE_DIR)/README.txt

ifeq ($(ARCHIVE_EXT),zip)
	cd $(DIST_DIR) && zip -qr $(PACKAGE_NAME).zip $(PACKAGE_NAME)
else
	tar -C $(DIST_DIR) -czf $(ARCHIVE) $(PACKAGE_NAME)
endif

	@echo ""
	@echo "Created:"
	@echo "  $(ARCHIVE)"

release: clean package

release-linux:
	$(MAKE) package OS=linux ARCH=amd64

release-windows:
	$(MAKE) package OS=windows ARCH=amd64

release-all: clean
	$(MAKE) package OS=linux   ARCH=amd64
	$(MAKE) package OS=linux   ARCH=arm64
	$(MAKE) package OS=windows ARCH=amd64
	$(MAKE) package OS=darwin  ARCH=amd64
	$(MAKE) package OS=darwin  ARCH=arm64

# Build both platform archives and publish them as a GitHub release tagged
# v$(VERSION). Requires the `gh` CLI, authenticated, with the commit pushed.
publish: release-all
	@command -v gh >/dev/null || { echo "gh CLI not found; install from https://cli.github.com"; exit 1; }
	gh release create v$(VERSION) \
		$(DIST_DIR)/$(APP_NAME)_$(VERSION)_*.tar.gz \
		$(DIST_DIR)/$(APP_NAME)_$(VERSION)_*.zip \
		--title "v$(VERSION)" \
		--generate-notes

clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
