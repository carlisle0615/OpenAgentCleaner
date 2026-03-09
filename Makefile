BINARY ?= oac
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
BUILDDIR ?= $(CURDIR)/bin
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || printf 'dev')
LDFLAGS ?= -s -w -X github.com/carlisle0615/OpenAgentCleaner/internal/cleaner.Version=$(VERSION)

.PHONY: build fmt install install-release test uninstall

fmt:
	gofmt -w main.go internal/cleaner/*.go

build:
	mkdir -p "$(BUILDDIR)"
	go build -ldflags "$(LDFLAGS)" -o "$(BUILDDIR)/$(BINARY)" .

test:
	go test ./...

install:
	PREFIX="$(PREFIX)" BINDIR="$(BINDIR)" BINARY="$(BINARY)" VERSION="$(VERSION)" LDFLAGS="$(LDFLAGS)" ./scripts/install-local.sh

install-release:
	PREFIX="$(PREFIX)" BINDIR="$(BINDIR)" BINARY="$(BINARY)" VERSION="$(VERSION)" ./install.sh

uninstall:
	PREFIX="$(PREFIX)" BINDIR="$(BINDIR)" BINARY="$(BINARY)" ./scripts/uninstall.sh
