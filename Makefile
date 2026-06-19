APP_NAME := confluence-mcp
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

.PHONY: build package release release-linux release-windows release-all clean

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

	printf '%s\n' \
	'Run setup:' \
	'' \
	'  $(BIN_NAME) setup' \
	> $(PACKAGE_DIR)/README.txt

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

release-all: clean release-linux release-windows

clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
