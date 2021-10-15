.PHONY: build-lib

BUILD_PATH=$(realpath ./build)

ifeq ($(OS),Windows_NT)     # is Windows_NT on XP, 2000, 7, Vista, 10...
 detected_OS := Windows
else
 detected_OS := $(strip $(shell uname))
endif

ifeq ($(detected_OS),Darwin)
 LIB_EXT := dylib
 ifeq ("$(shell sysctl -nq hw.optional.arm64)","1")
  # Building on M1 is still not supported, so in the meantime we crosscompile to amd64
  CGOFLAGS=CGO_ENABLED=1 GOOS=darwin GOARCH=amd64
 endif
else ifeq ($(detected_OS),Windows)
 LIB_EXT:= dll
 LIBKEYCARD_EXT := dll
else
 LIB_EXT := so
endif

build-lib:
	mkdir -p $(BUILD_PATH)/libkeycard
	@echo "Building static library..."
	cd shared && \
		$(CGOFLAGS) go build -buildmode=c-shared -o $(BUILD_PATH)/libkeycard/libkeycard.$(LIB_EXT) .
	@echo "Static library built:"
	@ls -la $(BUILD_PATH)/libkeycard/*

build-example-shared: build-lib
	mkdir -p $(BUILD_PATH)
	@echo "Building example-c..."
	cd examples/example-shared && \
		go build -o $(BUILD_PATH)/example-shared

run-example-shared: build-example-shared
		LD_LIBRARY_PATH=$(BUILD_PATH)/libkeycard $(BUILD_PATH)/example-shared

build-example-go: build-lib
	mkdir -p $(BUILD_PATH)
	@echo "Building example-c..."
	cd examples/example-go && \
		go build -o $(BUILD_PATH)/example-go

run-example-go: build-example-go
		$(BUILD_PATH)/example-go
