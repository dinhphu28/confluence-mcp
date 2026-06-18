APP_NAME := confluence-mcp
VERSION ?= 0.1.0

OS := linux
ARCH := amd64

BUILD_DIR := build
DIST_DIR := dist

PACKAGE_NAME := $(APP_NAME)_$(VERSION)_$(OS)_$(ARCH)
PACKAGE_DIR := $(DIST_DIR)/$(PACKAGE_NAME)

TARBALL := $(DIST_DIR)/$(PACKAGE_NAME).tar.gz

.PHONY: build package release clean

build:
	mkdir -p $(BUILD_DIR)

	CGO_ENABLED=0 \
	GOOS=$(OS) \
	GOARCH=$(ARCH) \
	go build \
	-ldflags="-s -w -X main.version=$(VERSION)" \
	-o $(BUILD_DIR)/$(APP_NAME) .

package: build
	rm -rf $(PACKAGE_DIR)

	mkdir -p $(PACKAGE_DIR)

	cp $(BUILD_DIR)/$(APP_NAME) $(PACKAGE_DIR)/

	printf '%s\n' \
	'Run setup:' \
	'' \
	'./confluence-mcp setup' \
	> $(PACKAGE_DIR)/README.txt

	tar -C $(DIST_DIR) \
		-czf $(TARBALL) \
		$(PACKAGE_NAME)

	@echo ""
	@echo "Created:"
	@echo "  $(TARBALL)"

release: clean package

clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
