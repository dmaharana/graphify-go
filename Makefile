# Graphify-Go Makefile

BINARY_NAME=graphify-go
LDFLAGS=-ldflags="-s -w"

.PHONY: all build clean tidy run

all: tidy build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) main.go

tidy:
	go mod tidy

clean:
	rm -f $(BINARY_NAME)
	rm -f graph.json

run: build
	./$(BINARY_NAME) scan .

query: build
	./$(BINARY_NAME) query "$(q)" --graph graph.json
