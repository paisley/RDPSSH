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

LDFLAGS := -ldflags "\
	-X 'main.AppName=RDPSSH' \
	-X 'main.AppVersion=$(VERSION)' \
	-X 'main.BuildCommit=$(COMMIT)' \
	-X 'main.BuildDate=$(BUILD_TIME)' \
	-X 'main.DocURL=https://github.com/$(GITHUB_USER)/$(GITHUB_REPO)' \
	-X 'main.IssueURL=https://github.com/$(GITHUB_USER)/$(GITHUB_REPO)/issues'"

# Normal build (shows a console window)
LDFLAGS := -ldflags "$(LDFLAGS_BASE)

# Release build: GUI subsystem (no console window)
GUI_LDFLAGS := -ldflags "-H=windowsgui $(LDFLAGS_BASE)"

.PHONY: all build clean release

# Default target
all: build

# Build dev/test binary
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) go build $(LDFLAGS) -trimpath -o $(BINARY_NAME) .

# Build release binart
release:
    GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) go build $(GUI_LDFLAGS) -trimpath -o $(BINARY_NAME) .

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
