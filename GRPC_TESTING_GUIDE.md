# gRPC Testing Guide for Go-Kube Microservices

## Service Ports Summary

| Service | Internal Port | NodePort | Access | Type |
|---------|---------------|----------|--------|------|
| **User Service** | 50051 | 30051 | `localhost:30051` | gRPC |
| **Order Service** | 50052 | 30052 | `localhost:30052` | gRPC |
| **Kafka Bootstrap** | 9092 | 30092 | `localhost:30092` | Broker |
| **PostgreSQL** | 5432 | - | Internal only | Database |
| **Zookeeper** | 2181 | - | Internal only | Cluster Manager |

---

## Prerequisites

Make sure you have `grpcurl` installed:
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Or via brew
brew install grpcurl
```

---

## User Service Testing (Port 30051)

### 1. List Available Services
```bash
grpcurl -plaintext localhost:30051 list
```

**Expected Output:**
```
user.v1.UserService
grpc.health.v1.Health
```

---

### 2. Describe UserService
```bash
grpcurl -plaintext localhost:30051 describe user.v1.UserService
```

---

### 3. List Users (Empty initially)
```bash
grpcurl -plaintext localhost:30051 user.v1.UserService/ListUsers
```

**Expected Output:**
```json
{
  "users": []
}
```

---

### 4. Create First User
```bash
grpcurl -plaintext -d '{
  "email": "john@example.com",
  "name": "John Doe"
}' localhost:30051 user.v1.UserService/CreateUser
```

**Expected Output:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john@example.com",
  "name": "John Doe",
  "created_at": "2026-06-11T10:45:00Z"
}
```

---

### 5. Create Second User
```bash
grpcurl -plaintext -d '{
  "email": "jane@example.com",
  "name": "Jane Smith"
}' localhost:30051 user.v1.UserService/CreateUser
```

---

### 6. Get User by ID
```bash
# Replace with actual ID from Create User response
grpcurl -plaintext -d '{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}' localhost:30051 user.v1.UserService/GetUser
```

**Expected Output:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "john@example.com",
  "name": "John Doe",
  "created_at": "2026-06-11T10:45:00Z"
}
```

---

### 7. List All Users
```bash
grpcurl -plaintext localhost:30051 user.v1.UserService/ListUsers
```

**Expected Output:**
```json
{
  "users": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "email": "john@example.com",
      "name": "John Doe",
      "created_at": "2026-06-11T10:45:00Z"
    },
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "email": "jane@example.com",
      "name": "Jane Smith",
      "created_at": "2026-06-11T10:46:00Z"
    }
  ]
}
```

---

## Order Service Testing (Port 30052)

### 1. List Available Services
```bash
grpcurl -plaintext localhost:30052 list
```

**Expected Output:**
```
order.v1.OrderService
grpc.health.v1.Health
```

---

### 2. Describe OrderService
```bash
grpcurl -plaintext localhost:30052 describe order.v1.OrderService
```

---

### 3. List Orders (Empty initially)
```bash
grpcurl -plaintext localhost:30052 order.v1.OrderService/ListOrders
```

**Expected Output:**
```json
{
  "orders": []
}
```

---

### 4. Create Order
```bash
# Use a valid user ID from User Service
grpcurl -plaintext -d '{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 99.99,
  "items": "Laptop, Mouse"
}' localhost:30052 order.v1.OrderService/CreateOrder
```

**Expected Output:**
```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": "99.99",
  "items": "Laptop, Mouse",
  "status": "PENDING",
  "created_at": "2026-06-11T10:47:00Z"
}
```

**Note:** This also publishes an `order.created` event to Kafka, which the Notification Service will consume.

---

### 5. Get Order by ID
```bash
grpcurl -plaintext -d '{
  "id": "660e8400-e29b-41d4-a716-446655440001"
}' localhost:30052 order.v1.OrderService/GetOrder
```

---

### 6. List All Orders
```bash
grpcurl -plaintext localhost:30052 order.v1.OrderService/ListOrders
```

---

## Kafka Testing (Port 30092)

### 1. Test Kafka Broker Connectivity
```bash
# Using nc (netcat)
nc -zv localhost 30092
```

**Expected Output:**
```
Connection to localhost port 30092 [tcp/*] succeeded!
```

---

### 2. Check Kafka Topics
```bash
# Port-forward to Kafka first (or use from container)
kubectl exec -it kafka-0 -- kafka-topics.sh --list --bootstrap-server kafka:9092
```

**Expected Topics:**
```
order.created
order.updated
```

---

### 3. View Topic Configuration
```bash
kubectl exec -it kafka-0 -- kafka-topics.sh --describe --bootstrap-server kafka:9092 --topic order.created
```

---

### 4. Monitor Kafka Messages
```bash
# Consume messages from order.created topic (from beginning)
kubectl exec -it kafka-0 -- kafka-console-consumer.sh \
  --bootstrap-server kafka:9092 \
  --topic order.created \
  --from-beginning
```

---

## PostgreSQL Testing

### 1. Connect to PostgreSQL
```bash
# Port-forward first
kubectl port-forward svc/postgres 5432:5432

# In another terminal
psql -U postgres -h localhost -d user_db
```

### 2. Query Users Table
```sql
SELECT * FROM users;
```

### 3. Query Orders Table
```sql
SELECT * FROM orders;
```

### 4. Query Notifications Table
```sql
SELECT * FROM notifications;
```

---

## Complete Test Workflow

### Step 1: Start Port-Forward for Database (Optional)
```bash
kubectl port-forward svc/postgres 5432:5432 &
```

### Step 2: Create Users
```bash
USER1=$(grpcurl -plaintext -d '{"email":"alice@example.com","name":"Alice"}' localhost:30051 user.v1.UserService/CreateUser | jq -r '.id')
echo "Created User 1: $USER1"

USER2=$(grpcurl -plaintext -d '{"email":"bob@example.com","name":"Bob"}' localhost:30051 user.v1.UserService/CreateUser | jq -r '.id')
echo "Created User 2: $USER2"
```

### Step 3: List All Users
```bash
grpcurl -plaintext localhost:30051 user.v1.UserService/ListUsers | jq '.'
```

### Step 4: Create Orders
```bash
ORDER1=$(grpcurl -plaintext -d "{\"user_id\":\"$USER1\",\"amount\":49.99,\"items\":\"Book\"}" localhost:30052 order.v1.OrderService/CreateOrder | jq -r '.id')
echo "Created Order 1: $ORDER1"

ORDER2=$(grpcurl -plaintext -d "{\"user_id\":\"$USER2\",\"amount\":199.99,\"items\":\"Monitor, Keyboard\"}" localhost:30052 order.v1.OrderService/CreateOrder | jq -r '.id')
echo "Created Order 2: $ORDER2"
```

### Step 5: List All Orders
```bash
grpcurl -plaintext localhost:30052 order.v1.OrderService/ListOrders | jq '.'
```

### Step 6: Monitor Kafka Events
```bash
kubectl logs -f deployment/notification-service
```

### Step 7: Verify in Database
```bash
psql -U postgres -h localhost -d notification_db -c "SELECT * FROM notifications;"
```

---

## Troubleshooting

### Service not responding
```bash
# Check if pod is running
kubectl get pods

# Check logs
kubectl logs deployment/user-service
kubectl logs deployment/order-service
kubectl logs deployment/notification-service

# Check service is exposed
kubectl get svc | grep user-service
```

### Connection refused
```bash
# Verify service is accessible
kubectl exec -it postgres-0 -- nc -zv user-service 50051

# Check if NodePort is open
kubectl describe svc user-service
```

### Proto reflection not working
```bash
# Re-apply K8s manifests
kubectl delete -f k8s-default-namespace.yaml
kubectl apply -f k8s-default-namespace.yaml

# Rebuild images
docker-compose build
```

---

## Quick Commands Reference

```bash
# List all services and their ports
kubectl get svc

# Watch all pods
kubectl get pods -w

# Get pod details
kubectl describe pod user-service-<pod-id>

# Tail logs from all services
kubectl logs -f deployment/user-service
kubectl logs -f deployment/order-service
kubectl logs -f deployment/notification-service
kubectl logs -f statefulset/kafka

# Port-forward if needed
kubectl port-forward svc/postgres 5432:5432
```

---

## Notes

- All services support **gRPC reflection**, so you can use `grpcurl -plaintext <host>:<port> list` to discover services
- Order creation automatically publishes events to Kafka
- Notification Service consumes Kafka events and stores them in the database
- All data is persisted across pod restarts (except ephemeral containers)
