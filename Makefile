# ttt - Tiny Task Tool Makefile

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -X github.com/yostos/tiny-task-tool/internal/cli.Version=$(VERSION)

.PHONY: build install test lint clean

build:
	go build -ldflags "$(LDFLAGS)" -o ttt .

install: build
	cp ttt ~/bin/

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
