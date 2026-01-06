.PHONY: test test-tt test-ttt gen start stop

# Runs all tests in the internal directory
test:
	go test ./internal/... -v

# Runs all tests in the integration_tests directory
test-tt:
	go test ./integration_tests/... -v

# Runs all tests in the project
test-ttt:
	go test ./... -v

# Runs go generate ./...
gen:
	go generate ./...

# Runs docker compose up and builds/runs producer and consumer in background
start:
	docker compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Building and running producer..."
	@go run cmd/producer/main.go > /tmp/producer.log 2>&1 &
	@echo $$! > /tmp/producer.pid
	@echo "Building and running consumer..."
	@go run cmd/consumer/main.go > /tmp/consumer.log 2>&1 &
	@echo $$! > /tmp/consumer.pid
	@echo "Producer PID: $$(cat /tmp/producer.pid)"
	@echo "Consumer PID: $$(cat /tmp/consumer.pid)"
	@echo "Services started. Logs: /tmp/producer.log and /tmp/consumer.log"

# Stops producer, consumer and docker compose
stop:
	@if [ -f /tmp/producer.pid ]; then \
		kill $$(cat /tmp/producer.pid) 2>/dev/null || true; \
		rm /tmp/producer.pid; \
		echo "Producer stopped"; \
	fi
	@if [ -f /tmp/consumer.pid ]; then \
		kill $$(cat /tmp/consumer.pid) 2>/dev/null || true; \
		rm /tmp/consumer.pid; \
		echo "Consumer stopped"; \
	fi
	docker compose down

