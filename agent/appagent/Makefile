.PHONY: build run clean

# Build the RabbitMQ worker
build:
	go build -o bin/agent cmd/agent/main.go

# Run the RabbitMQ worker
run:
	./bin/agent --config=config.yaml

# Clean build artifacts
clean:
	rm -rf bin/
