# Kubernetes Deployment Guide for go-kube

## Prerequisites

1. A running Kubernetes cluster (local: minikube, kind, Docker Desktop; cloud: EKS, GKE, AKS)
2. `kubectl` configured to communicate with your cluster
3. Docker images built and available (either in a registry or loaded locally)

## Building Docker Images

Build the three service images:

```bash
docker build -f services/user-service/Dockerfile -t go-kube-user-service:latest .
docker build -f services/order-service/Dockerfile -t go-kube-order-service:latest .
docker build -f services/notification-service/Dockerfile -t go-kube-notification-service:latest .
```

### For Local Clusters (minikube, kind, Docker Desktop)

If using a local cluster, load the images into the cluster:

**Minikube:**
```bash
minikube image load go-kube-user-service:latest
minikube image load go-kube-order-service:latest
minikube image load go-kube-notification-service:latest
```

**Kind:**
```bash
kind load docker-image go-kube-user-service:latest --name <cluster-name>
kind load docker-image go-kube-order-service:latest --name <cluster-name>
kind load docker-image go-kube-notification-service:latest --name <cluster-name>
```

**Docker Desktop Kubernetes:**
Images are automatically available to the cluster.

### For Remote Clusters (EKS, GKE, AKS)

Push images to a container registry:

```bash
# Tag images for your registry
docker tag go-kube-user-service:latest <registry>/go-kube-user-service:latest
docker tag go-kube-order-service:latest <registry>/go-kube-order-service:latest
docker tag go-kube-notification-service:latest <registry>/go-kube-notification-service:latest

# Push to registry
docker push <registry>/go-kube-user-service:latest
docker push <registry>/go-kube-order-service:latest
docker push <registry>/go-kube-notification-service:latest
```

Then update `k8s-manifest.yaml`:
- Change `imagePullPolicy: Never` to `imagePullPolicy: IfNotPresent`
- Update image references to use your registry URL

## Deployment

Deploy all resources to the cluster:

```bash
kubectl apply -f k8s-manifest.yaml
```

Verify the deployment:

```bash
# Check namespace
kubectl get ns go-kube

# Check all resources
kubectl get all -n go-kube

# Check pod status
kubectl get pods -n go-kube -w

# Check services
kubectl get svc -n go-kube

# Check persistent volumes
kubectl get pvc -n go-kube
```

Wait for all pods to be in `Running` and `Ready` state:

```bash
kubectl wait --for=condition=Ready pod -l app=postgres -n go-kube --timeout=300s
kubectl wait --for=condition=Ready pod -l app=zookeeper -n go-kube --timeout=300s
kubectl wait --for=condition=Ready pod -l app=kafka -n go-kube --timeout=300s
kubectl wait --for=condition=Ready pod -l app=user-service -n go-kube --timeout=300s
```

## Accessing Services

### Within the Cluster

Services are accessible using their DNS names:
- PostgreSQL: `postgres:5432`
- Zookeeper: `zookeeper:2181`
- Kafka: `kafka:9092`
- User Service: `user-service:50051`
- Order Service: `order-service:50052`
- Notification Service: Internal only

### From Your Local Machine

Port-forward services to localhost:

```bash
# PostgreSQL
kubectl port-forward -n go-kube svc/postgres 5432:5432 &

# Kafka
kubectl port-forward -n go-kube svc/kafka 9092:9092 &

# User Service
kubectl port-forward -n go-kube svc/user-service 50051:50051 &

# Order Service
kubectl port-forward -n go-kube svc/order-service 50052:50052 &
```

## Debugging

### View pod logs

```bash
# Follow logs from a specific service
kubectl logs -n go-kube -f deployment/user-service

# View logs from a specific pod
kubectl logs -n go-kube <pod-name>

# View logs from all pods in a deployment
kubectl logs -n go-kube -f deployment/postgres
```

### Describe pod for events

```bash
kubectl describe pod -n go-kube <pod-name>
```

### Execute commands in a pod

```bash
kubectl exec -n go-kube -it <pod-name> -- /bin/sh
```

### Check PostgreSQL databases

```bash
# Port-forward postgres first
kubectl port-forward -n go-kube svc/postgres 5432:5432

# From another terminal, connect with psql
psql -h localhost -U postgres -d user_db
```

## Configuration

### Updating Environment Variables

Edit `k8s-manifest.yaml` and modify the ConfigMap or Deployment environment variables, then:

```bash
kubectl apply -f k8s-manifest.yaml
```

### Updating Image Versions

Update the image tags in the Deployment specs (user-service, order-service, notification-service), then:

```bash
kubectl apply -f k8s-manifest.yaml
```

### Scaling Replicas

Increase or decrease replicas for any service:

```bash
kubectl scale deployment user-service -n go-kube --replicas=3
kubectl scale deployment order-service -n go-kube --replicas=3
kubectl scale deployment notification-service -n go-kube --replicas=3
```

## Cleanup

Remove all resources:

```bash
kubectl delete namespace go-kube
```

This will delete all deployments, services, PVCs, and ConfigMaps in the namespace.

## Storage Considerations

The manifest uses PersistentVolumeClaims for PostgreSQL, Kafka, and Zookeeper. In production:

1. Use StorageClasses backed by cloud storage (EBS for EKS, GCP Disks for GKE, etc.)
2. Configure automated backups for PostgreSQL data
3. Consider multi-replica setups for Kafka and Zookeeper
4. Implement PVC resizing policies

## Security Best Practices

1. Store the `postgres` password in a proper secret management system (HashiCorp Vault, AWS Secrets Manager, etc.)
2. Use network policies to restrict traffic between services
3. Add resource limits and requests for all containers (already included in the manifest)
4. Use RBAC for access control
5. Enable Pod Security Policies or Pod Security Standards
