# Simple Makefile for building a Windows binary from Linux

BINARY_NAME := rdpssh.exe

# Target Windows 64-bit
GOOS   ?= windows
GOARCH ?= amd64
CGO_ENABLED ?= 1
CC := x86_64-w64-mingw32-gcc


VERSION    ?= dev
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

GITHUB_USER ?= paisley
GITHUB_REPO ?= rdpssh
LDFLAGS_BASE := -X main.AppName=RDPSSH \
	-X 'main.AppVersion=$(VERSION)' \
	-X 'main.BuildCommit=$(COMMIT)' \
	-X 'main.BuildDate=$(BUILD_TIME)' \
	-X 'main.DocURL=https://github.com/$(GITHUB_USER)/$(GITHUB_REPO)' \
	-X 'main.IssueURL=https://github.com/$(GITHUB_USER)/$(GITHUB_REPO)/issues'

# Normal build (shows a console window)
LDFLAGS := -ldflags "$(LDFLAGS_BASE)

BUILD_DIR := release
ZIP_NAME := rdpssh-$(VERSION)-windows-amd64.zip

.PHONY: all build clean release release-zip

# Default target
all: build

# Build dev/test binary
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) go build -ldflags $(LDFLAGS) -trimpath -o $(BINARY_NAME) .

# Build release binary
release:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) go build -ldflags "-H=windowsgui $(LDFLAGS_BASE)" -trimpath -o $(BINARY_NAME) .

# Archive release
release-zip: release
	mkdir -p $(BUILD_DIR)
	zip $(BUILD_DIR)/$(ZIP_NAME) $(BINARY_NAME) LICENSE README.md

# Remove build artifact
clean:
	rm -f $(BINARY_NAME)

# Show build info
info:
	@echo "RDPSSH Build Information"
	@echo "========================"
	@echo "Binary:      $(BINARY_NAME)"
	@echo "Version:     $(VERSION)"
	@echo "Commit:      $(COMMIT)"
	@echo "Build Time:  $(BUILD_TIME)"
	@echo "GitHub:      https://github.com/$(GITHUB_USER)/$(GITHUB_REPO)"
