BINDIR := bin

.PHONY: all server loader clean

all: server loader

server:
	go build -o $(BINDIR)/spotter-server ./cmd/server

loader:
	go build -o $(BINDIR)/spotter-loader ./cmd/loader

clean:
	rm -rf $(BINDIR)

format:
	@echo "Formatting Go code..."
	gofmt -s -w .
