BINARY ?= oac
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
BUILDDIR ?= $(CURDIR)/bin

.PHONY: build fmt install test uninstall

fmt:
	gofmt -w main.go internal/cleaner/*.go

build:
	mkdir -p "$(BUILDDIR)"
	go build -o "$(BUILDDIR)/$(BINARY)" .

test:
	go test ./...

install:
	PREFIX="$(PREFIX)" BINDIR="$(BINDIR)" BINARY="$(BINARY)" ./scripts/install.sh

uninstall:
	PREFIX="$(PREFIX)" BINDIR="$(BINDIR)" BINARY="$(BINARY)" ./scripts/uninstall.sh
