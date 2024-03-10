# Variables
BIN_DIR := bin
BUILD_DIR := build

# Targets
.PHONY: build run

build:
	go build -o ./${BIN_DIR}/amigo ./amigo/

run:
	./${BIN_DIR}/amigo --directory ~/Workspace/qcscripts --extensions js,sjs,xqy


clean:
	rm -rf ${BIN_DIR}
