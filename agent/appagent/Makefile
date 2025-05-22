.PHONY: build-worker build-publisher run-worker run-publisher clean

# Build the RabbitMQ worker
build-worker:
	go build -o bin/rabbitmq-worker cmd/rabbitmq-worker/main.go

# Build the task publisher
build-publisher:
	go build -o bin/task-publisher cmd/task-publisher/main.go

# Build all
build: build-worker build-publisher

# Run the RabbitMQ worker
run-worker:
	./bin/rabbitmq-worker --config=config.yaml

# Run the task publisher with session task
run-publisher-session:
	./bin/task-publisher --config=config.yaml --type=session --source=$(SOURCE) --url=$(URL)

# Run the task publisher with book task
run-publisher-book:
	./bin/task-publisher --config=config.yaml --type=book --source=$(SOURCE) --url=$(URL)

# Run the task publisher with chapter task
run-publisher-chapter:
	./bin/task-publisher --config=config.yaml --type=chapter --source=$(SOURCE) --url=$(URL)

# Run the task publisher with session task for sangtacviet
run-publisher-stv-session:
	./bin/task-publisher --config=config.yaml --type=session --source=sangtacviet --url=$(URL)

# Run the task publisher with book task for sangtacviet
run-publisher-stv-book:
	./bin/task-publisher --config=config.yaml --type=book --source=sangtacviet --url=$(URL)

# Run the task publisher with chapter task for sangtacviet
run-publisher-stv-chapter:
	./bin/task-publisher --config=config.yaml --type=chapter --source=sangtacviet --url=$(URL)

# Clean build artifacts
clean:
	rm -rf bin/
