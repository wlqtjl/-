GO := ~/go/bin/go
BINARY := wozai

.PHONY: build run test lint clean migrate-up migrate-down

build:
	$(GO) build -ldflags="-s -w" -o $(BINARY) ./cmd/wozai

run:
	$(GO) run ./cmd/wozai

test:
	$(GO) test ./... -v -cover

clean:
	rm -f $(BINARY)
