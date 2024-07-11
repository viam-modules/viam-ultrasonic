SOURCE_OS ?= $(shell uname -s | tr '[:upper:]' '[:lower:]')
SOURCE_ARCH ?= $(shell uname -m)
TARGET_OS ?= $(SOURCE_OS)
TARGET_ARCH ?= $(SOURCE_ARCH)
BIN_OUTPUT_PATH = bin/$(TARGET_OS)-$(TARGET_ARCH)
TOOL_BIN = bin/gotools/$(shell uname -s)-$(shell uname -m)
UNAME_S ?= $(shell uname -s)

ifeq ($(TARGET_OS),linux)
	CGO_ENABLED = 1
	CGO_LDFLAGS := -l:libjpeg.a
endif

.PHONY: build clean gofmt lint update-rdk test

build:
	rm -f $(BIN_OUTPUT_PATH)/ultrasonic-module
	CGO_ENABLED=$(CGO_ENABLED) CGO_LDFLAGS="$(CGO_LDFLAGS)" go build -o $(BIN_OUTPUT_PATH)/ultrasonic-module main.go

# bin/ultrasonic-module is the expected entrypoint specified in meta.json
module.tar.gz: build
	cp $(BIN_OUTPUT_PATH)/ultrasonic-module bin/ultrasonic-module
	tar czf module.tar.gz bin/ultrasonic-module
	rm bin/ultrasonic-module

clean:
	rm -rf $(BIN_OUTPUT_PATH)/ultrasonic-module $(BIN_OUTPUT_PATH)/module.tar.gz ultrasonic-module

gofmt:
	gofmt -w -s .

lint: gofmt
	go mod tidy

update-rdk:
	go get go.viam.com/rdk@latest
	go mod tidy

test:
	go test ./...
