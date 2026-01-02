# Test

## Casino Transaction Management System

### Overview

You are tasked with building a simple transaction management system for a casino. The system will track user transactions related to their bets and wins. The transactions will be processed asynchronously via a message system, and the data will be stored in a relational database. You will also need to expose an API to allow querying of transaction data.

---

## Key Components

### 1. Message System

- **Choice**: Kafka or RabbitMQ
- **Purpose**: The system will receive transaction data (bet/win events) as messages, which need to be processed and saved in the database.

### 2. Database

- **Choice**: PostgreSQL or MySQL
- **Purpose**: Store the transaction data

**Transaction fields:**
- `user_id` - The ID of the user making the transaction
- `transaction_type` - Either "bet" or "win"
- `amount` - The amount of money for the transaction
- `timestamp` - The time the transaction occurred

### 3. Transaction API

- **Purpose**: Allow clients to query transaction data
- **Features**:
  - Query for a single user or all transactions
  - Support filtering by transaction type (bet, win, or all transactions)

---

## Requirements

### 1. Message Consumer

- ✅ Create a message consumer that listens for messages (bet/win transactions) from the chosen message system (Kafka or RabbitMQ)
- ✅ The consumer must process the messages asynchronously and store the transaction details in the chosen database

### 2. Database

- ✅ Set up the database schema to store the transaction data

### 3. API

- ✅ Implement the API in Go
- ✅ The API must allow users to query their transaction history
- ✅ Support filtering by `transaction_type` (bet, win, or all)
- ✅ Ensure the API returns the transactions in JSON format

### 4. Testing

- ✅ Write unit and integration tests for all components
- ✅ Test coverage should be at least **85%**

### 5. Documentation

- ✅ Provide a README file with any relevant instructions

---

## Submission

- Please submit the source code (how you prefer) along with the README file

---

## Implementation Notes

### Technology Stack Used

- **Language**: Go
- **Message System**: Apache Kafka (using segmentio/kafka-go)
- **Database**: PostgreSQL (using sqlx)
- **API Protocol**: REST with OpenAPI 3.0 specification
- **Serialization**: Protocol Buffers (protobuf)
- **Testing**: testify/assert and testify/mock

### Project Structure

```
casino-transactions-system/
├── api/                          # API specifications
│   ├── openapi.yaml             # OpenAPI 3.0 specification
│   ├── transaction-event.proto  # Protobuf definition
│   └── transaction-event.pb.go  # Generated protobuf code
├── cmd/                          # Application entry points
│   ├── consumer/                # Consumer application
│   └── producer/                # Producer application (for testing)
├── configs/                      # Configuration files
│   ├── consumer-config.yaml
│   └── producer-config.yaml
├── internal/                     # Internal packages
│   ├── app/                     # Application layer
│   ├── config/                  # Configuration handling
│   ├── entity/                  # Domain entities
│   ├── handlers/                # HTTP handlers
│   ├── infrastructure/          # Infrastructure layer
│   │   ├── kafka/              # Kafka adapter
│   │   └── repositories/       # Database repositories
│   ├── services/                # Business logic services
│   └── transformer/             # DTO transformers
├── migrations/                   # Database migrations
│   └── postgresql/
│       └── V0001__Initial_schema.sql
├── go.mod
├── go.sum
├── README.md
└── task.md                      # This file
```

### Key Features Implemented

#### Message Consumer
- Asynchronous message processing from Kafka
- Protobuf deserialization
- Batch storage in PostgreSQL
- Error handling and logging

#### Database
- PostgreSQL schema with indexes
- Support for complex queries with filtering
- Transaction support for batch operations
- Connection pooling

#### API
- REST API with POST /transactions endpoint
- Filtering by:
  - User ID
  - Transaction type (bet/win/all)
  - Amount range
  - Date range
- Pagination support (limit/offset)
- JSON response format
- Health check endpoint

#### Testing
- Unit tests for all services
- Mock-based testing for isolation
- Test coverage > 85%
- Table-driven tests

### Running the Application

See README.md for detailed instructions on:
- Setting up the development environment
- Running Kafka and PostgreSQL
- Starting the consumer
- Running tests
- API usage examples

