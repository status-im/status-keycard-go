.PHONY: build-lib

BUILD=./build

ifeq ($(OS),Windows_NT)     # is Windows_NT on XP, 2000, 7, Vista, 10...
 detected_OS := Windows
else
 detected_OS := $(strip $(shell uname))
endif

ifeq ($(detected_OS),Darwin)
 LIB_EXT := dylib
else ifeq ($(detected_OS),Windows)
 LIB_EXT:= dll
 LIBKEYCARD_EXT := dll
else
 LIB_EXT := so
endif

build-lib:
	mkdir -p $(BUILD)/libkeycard
	@echo "Building static library..."
	go build -buildmode=c-shared -o $(BUILD)/libkeycard/libkeycard.$(LIB_EXT) .
	@echo "Static library built:"
	@ls -la $(BUILD)/libkeycard/*
