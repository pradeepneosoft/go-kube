# go-kube

Small Go microservices demo for learning Docker and Kubernetes later.

## Architecture

```
┌─────────────────┐     gRPC      ┌─────────────────┐
│  order-service  │──────────────▶│  user-service   │
│   :50052        │               │   :50051        │
└────────┬────────┘               └────────┬────────┘
         │ publish                          │ Postgres
         │ orders.created                   │ user_db
         ▼                                    ▼
    ┌─────────┐                        ┌───────────┐
    │  Kafka  │───────────────────────▶│ Postgres  │
    └─────────┘   consume              └───────────┘
         │
         ▼
┌─────────────────────┐
│ notification-service│
└─────────┬───────────┘
          │ Postgres
          ▼
     notification_db
```

### Services

| Service | Role | Protocol |
|---------|------|----------|
| **user-service** | User CRUD | gRPC + Postgres |
| **order-service** | Order CRUD, validates user via gRPC, publishes events | gRPC + Postgres + Kafka |
| **notification-service** | Stores notifications when orders are created | Kafka consumer + Postgres |

## Project layout

```
go-kube/
├── gen/proto/              # Generated gRPC code
├── pkg/                    # Shared helpers (kafka, postgres, events)
├── proto/                  # Protobuf definitions
└── services/
    ├── user-service/
    ├── order-service/
    └── notification-service/
```

## Prerequisites

Run these locally before starting the services:

1. **Go 1.22+**
2. **PostgreSQL** — three databases (one per service)
3. **Kafka** — broker on `localhost:9092`

### Create databases

```sql
CREATE DATABASE user_db;
CREATE DATABASE order_db;
CREATE DATABASE notification_db;
```

Each database needs the `pgcrypto` extension for UUID generation:

```sql
\c user_db
CREATE EXTENSION IF NOT EXISTS pgcrypto;

\c order_db
CREATE EXTENSION IF NOT EXISTS pgcrypto;

\c notification_db
CREATE EXTENSION IF NOT EXISTS pgcrypto;
```

Migrations run automatically when each service starts.

## Environment variables

### user-service

```bash
export USER_SERVICE_DATABASE_URL="postgres://postgres:postgres@localhost:5432/user_db?sslmode=disable"
export USER_SERVICE_GRPC_PORT="50051"
```

### order-service

```bash
export ORDER_SERVICE_DATABASE_URL="postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable"
export ORDER_SERVICE_GRPC_PORT="50052"
export USER_SERVICE_URL="localhost:50051"
export KAFKA_BROKERS="localhost:9092"
export KAFKA_ORDERS_TOPIC="orders.created"
```

### notification-service

```bash
export NOTIFICATION_SERVICE_DATABASE_URL="postgres://postgres:postgres@localhost:5432/notification_db?sslmode=disable"
export KAFKA_BROKERS="localhost:9092"
export KAFKA_ORDERS_TOPIC="orders.created"
export KAFKA_NOTIFICATION_GROUP="notification-service"
```

## Run locally

From the project root:

```bash
make tidy
make build

# Terminal 1
make run-user

# Terminal 2
make run-order

# Terminal 3
make run-notification
```

## Test with grpcurl

Install [grpcurl](https://github.com/fullstorydev/grpcurl), then:

```bash
# Create a user
grpcurl -plaintext -d '{"email":"alice@example.com","name":"Alice"}' \
  localhost:50051 user.v1.UserService/CreateUser

# Create an order (replace USER_ID)
grpcurl -plaintext -d '{"user_id":"USER_ID","product_name":"Go Book","quantity":1}' \
  localhost:50052 order.v1.OrderService/CreateOrder
```

After creating an order, check notification-service logs — it should store a notification in `notification_db`.

## Next steps

- Docker Compose for Postgres + Kafka
- Dockerfiles per service
- Kubernetes manifests (Deployments, Services, ConfigMaps)
