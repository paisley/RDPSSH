# Simple Makefile for building a Windows binary from Linux

BINARY_NAME := rdpssh.exe

# Target Windows 64-bit
GOOS   ?= windows
GOARCH ?= amd64

VERSION    ?= dev
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

GITHUB_USER ?= paisley
GITHUB_REPO ?= rdpssh

LDFLAGS := -ldflags "\
	-X 'main.AppName=RDPSSH' \
	-X 'main.AppVersion=$(VERSION)' \
	-X 'main.BuildCommit=$(COMMIT)' \
	-X 'main.BuildDate=$(BUILD_TIME)' \
	-X 'main.DocURL=https://github.com/$(GITHUB_USER)/$(GITHUB_REPO)' \
	-X 'main.IssueURL=https://github.com/$(GITHUB_USER)/$(GITHUB_REPO)/issues'"

.PHONY: all build clean test info

# Default target
all: build

# Build Windows binary
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -trimpath -o $(BINARY_NAME) .

# Remove build artifact
clean:
	rm -f $(BINARY_NAME)

# Run tests (on your Linux host)
test:
	go test ./...

# Show build info
info:
	@echo "RDPSSH Build Information"
	@echo "========================"
	@echo "Binary:      $(BINARY_NAME)"
	@echo "Version:     $(VERSION)"
	@echo "Commit:      $(COMMIT)"
	@echo "Build Time:  $(BUILD_TIME)"
	@echo "GitHub:      https://github.com/$(GITHUB_USER)/$(GITHUB_REPO)"
