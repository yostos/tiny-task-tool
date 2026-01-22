# ttt - Tiny Task Tool Makefile

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X github.com/yostos/tiny-task-tool/internal/cli.Version=$(VERSION)

# Installation directory
# Default: $GOPATH/bin or $HOME/go/bin
# Override with: make install PREFIX=/usr/local
ifdef PREFIX
    INSTALL_DIR := $(PREFIX)/bin
else ifdef GOPATH
    INSTALL_DIR := $(GOPATH)/bin
else
    INSTALL_DIR := $(HOME)/go/bin
endif

.PHONY: build install test lint clean

build:
	go build -ldflags "$(LDFLAGS)" -o ttt .

install: build
	@mkdir -p $(INSTALL_DIR)
	cp ttt $(INSTALL_DIR)/
	@echo "Installed ttt to $(INSTALL_DIR)/ttt"

test:
	go test ./...

lint:
	golangci-lint run

check: test lint

clean:
	rm -f ttt

# Show current version
version:
	@echo $(VERSION)
