# Name of the binary
BINARY_NAME=schedulerx

# Location of the main Go file
MAIN=src/main.go

# Default target: build and run
all: build run

# Build the Go application
build:
	go build -o $(BINARY_NAME) $(MAIN)

# Run the built binary
run:
	./$(BINARY_NAME)

# Run directly using 'go run'
gorun:
	go run $(MAIN)

# Run with Air for hot reload
dev:
	air

# Install Air for hot reloading
install-air:
	go install github.com/cosmtrek/air@latest

# Clean up binary
clean:
	rm -f $(BINARY_NAME)

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod tidy

# Generate Air config if it doesn't exist
init-air:
	@if [ ! -f .air.toml ]; then \
		air init; \
	fi
