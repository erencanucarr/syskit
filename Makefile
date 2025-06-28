# Simple helper makefile

BINARY=syskit
PREFIX?=/usr/local
BINDIR:=$(PREFIX)/bin

build:
	go build -o $(BINARY)

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY)-linux-amd64

build-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(BINARY)-linux-arm64

build-all: build-linux build-arm64

install: build
	install -Dm755 $(BINARY) $(BINDIR)/$(BINARY)
	@echo "Installed $(BINARY) to $(BINDIR)"

uninstall:
	rm -f $(BINDIR)/$(BINARY)

.PHONY: build install uninstall
