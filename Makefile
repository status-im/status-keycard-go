.PHONY: build-lib

BUILD=./build

build-lib:
	mkdir -p $(BUILD)/libkeycard
	@echo "Building static library..."
	go build -buildmode=c-shared -o $(BUILD)/libkeycard/libkeycard.so .
	@echo "Static library built:"
	@ls -la $(BUILD)/libkeycard/*
