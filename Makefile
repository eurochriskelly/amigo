# Variables
BIN_DIR := bin
BUILD_DIR := build

# Targets
.PHONY: build run

build:
	go build -o ./${BIN_DIR}/amigo ./amigo/

run:
	./${BIN_DIR}/amigo --directory ./test --extensions js,sjs,xqy --port 9292

clean:
	rm -rf ${BIN_DIR}

# Local commands
serve:
	bash .private/serve.sh
