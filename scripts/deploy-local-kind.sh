#!/usr/bin/env bash
set -euo pipefail

# Default to the standard kind cluster name. You can override it with:
# KIND_CLUSTER_NAME=my-cluster make deploy-local
CLUSTER_NAME="${KIND_CLUSTER_NAME:-kind}"

# Kubernetes manifest applied after the service images are built and loaded.
# Override with K8S_MANIFEST=some-file.yaml if you want to deploy another file.
MANIFEST="${K8S_MANIFEST:-k8s-default-namespace.yaml}"

# These names match the service folders and Docker image names used by the
# Kubernetes manifest: go-kube-<service>:latest.
services=(
  "user-service"
  "order-service"
  "notification-service"
)

# Stop early with a clear message if a required local tool is missing.
require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_command docker
require_command kind
require_command kubectl

# Docker Desktop must be running because we build the images locally before
# loading them into the kind cluster.
if ! docker info >/dev/null 2>&1; then
  echo "Docker is not running. Start Docker Desktop and try again." >&2
  exit 1
fi

# Create the local kind cluster on first deploy. If it already exists, reuse it.
if ! kind get clusters | grep -qx "$CLUSTER_NAME"; then
  echo "kind cluster '$CLUSTER_NAME' was not found. Creating it..."
  kind create cluster --name "$CLUSTER_NAME"
fi

# Point kubectl at the kind cluster so the apply/rollout commands go to the
# local cluster instead of any other Kubernetes context on your machine.
kubectl config use-context "kind-$CLUSTER_NAME"

# kind nodes cannot see Docker Desktop images automatically. Build each service
# image locally, then copy it into the kind cluster's container runtime.
for service in "${services[@]}"; do
  image="go-kube-${service}:latest"
  dockerfile="services/${service}/Dockerfile"

  echo "Building ${image}..."
  docker build -t "$image" -f "$dockerfile" .

  echo "Loading ${image} into kind cluster '${CLUSTER_NAME}'..."
  kind load docker-image "$image" --name "$CLUSTER_NAME"
done

# Apply Postgres, Kafka, Zookeeper, and the three Go service deployments.
echo "Applying Kubernetes manifest..."
kubectl apply -f "$MANIFEST"

# Wait until Kubernetes confirms each workload has rolled out. This makes the
# script fail fast if a pod cannot start, pull an image, or pass readiness.
echo "Waiting for workloads to roll out..."
kubectl rollout status deployment/postgres --timeout=180s
kubectl rollout status statefulset/zookeeper --timeout=240s
kubectl rollout status statefulset/kafka --timeout=300s
kubectl rollout status deployment/user-service --timeout=180s
kubectl rollout status deployment/order-service --timeout=180s
kubectl rollout status deployment/notification-service --timeout=180s

# Print a short final status summary so you can see pod and service state.
echo "Deployment complete."
kubectl get pods
kubectl get services
