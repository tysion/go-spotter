BINDIR := bin

.PHONY: all server loader bot clean

all: server loader bot

server:
	go build -o $(BINDIR)/spotter-server ./cmd/server

loader:
	go build -o $(BINDIR)/spotter-loader ./cmd/loader

bot:
	go build -o $(BINDIR)/spotter-bot ./cmd/bot

clean:
	rm -rf $(BINDIR)

format:
	@echo "Formatting Go code..."
	gofmt -s -w .
