# Casino Transaction System

A casino transaction management system consisting of two main components: producer and consumer.

## Architecture

### Producer

The producer generates transaction events (bets and wins) and sends them to Kafka. The application is configured through the `configs/producer-config.yaml` configuration file, where you can set event generation parameters:

- `initialBatchSize` - initial batch size
- `creationRPS` - number of events created per second
- `distinctUsers` - number of unique users
- `amountFrom` / `amountTo` - transaction amount range

### Consumer

The consumer receives messages from Kafka and performs two main functions:

1. **Message processing** - extracts transaction events from Kafka and saves them to PostgreSQL
2. **REST API** - provides HTTP API for querying transaction history with filtering support

The REST API is described in the `api/openapi.yaml` file and includes the following capabilities:
- Transaction search with filtering by user, transaction type, date, and amount
- Result pagination
- Health check endpoint

The consumer is configured through `configs/consumer-config.yaml`, where connection parameters for Kafka, PostgreSQL, and HTTP server are specified.

## Requirements

- Go 1.25+
- Docker and Docker Compose

## Running

### Using Makefile (Recommended)

The easiest way to start all services is using the Makefile:

```bash
make start
```

This command will:
1. Start infrastructure (Kafka, Zookeeper, PostgreSQL) using Docker Compose
2. Start the producer in the background
3. Start the consumer in the background

Logs are available at:
- Producer: `/tmp/producer.log`
- Consumer: `/tmp/consumer.log`

To stop all services:

```bash
make stop
```

### Manual Start

#### 1. Starting Infrastructure

Before running the applications, you need to start the infrastructure (Kafka and PostgreSQL) using Docker Compose:

```bash
docker compose up -d
```

This will start the following services:
- **Zookeeper** (port 2181) - for managing Kafka
- **Kafka** (port 9092) - message broker
- **PostgreSQL** (port 5432) - database for storing transactions

#### 2. Running Producer

```bash
go run cmd/producer/main.go
```

#### 3. Running Consumer

```bash
go run cmd/consumer/main.go
```

The consumer will start an HTTP server on port 9093 (configurable in `configs/consumer-config.yaml`).

## Testing

### Using Makefile (Recommended)

Before running tests, you need to generate mocks:

```bash
make gen
```

After generating mocks, you can run tests using Makefile commands:

```bash
make test        # Run tests in internal directory
make test-tt     # Run integration tests
make test-ttt    # Run all tests in the project
```

### Manual Testing

Testing uses `mockgen` from the `go.uber.org/mock` package. Before running tests, you need to generate mocks:

```bash
go generate ./...
```

After generating mocks, you can run tests:

```bash
go test ./...
```

## Configuration

Configuration files are located in the `configs/` directory:
- `producer-config.yaml` - producer settings
- `consumer-config.yaml` - consumer settings

If necessary, you can change connection parameters for Kafka, PostgreSQL, and other application settings.
