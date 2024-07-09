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

build:
	rm -f $(BIN_OUTPUT_PATH)/ultrasonic-module
	CGO_ENABLED=$(CGO_ENABLED) CGO_LDFLAGS="$(CGO_LDFLAGS)" go build -o $(BIN_OUTPUT_PATH)/ultrasonic-module main.go

module.tar.gz: build
	rm -f $(BIN_OUTPUT_PATH)/module.tar.gz
	tar czf $(BIN_OUTPUT_PATH)/module.tar.gz $(BIN_OUTPUT_PATH)/ultrasonic-module

setup:
	if [ "$(UNAME_S)" = "Linux" ]; then \
		sudo apt install -y libjpeg-dev pkg-config; \
	fi

clean:
	rm -rf $(BIN_OUTPUT_PATH)/ultrasonic-module $(BIN_OUTPUT_PATH)/module.tar.gz ultrasonic-module

gofmt:
	gofmt -w -s .

lint: gofmt
	go mod tidy

update-rdk:
	go get go.viam.com/rdk@latest
	go mod tidy
