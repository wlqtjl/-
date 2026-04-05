GO ?= go
BINARY := wozai

.PHONY: build run test lint clean vet

build:
	$(GO) build -ldflags="-s -w" -o $(BINARY) ./cmd/wozai

run:
	$(GO) run ./cmd/wozai

test:
	$(GO) test ./... -v -cover

vet:
	$(GO) vet ./...

lint: vet

clean:
	rm -f $(BINARY)
