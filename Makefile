BIN := bin/bytree
MODULE := github.com/bytaesu/bytree
VERSION ?= dev
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install test clean

build:
	go build -ldflags '$(LDFLAGS)' -o $(BIN) ./cmd/bytree/

install:
	go install -ldflags '$(LDFLAGS)' ./cmd/bytree/

test:
	go test ./...

clean:
	rm -rf bin/ dist/
