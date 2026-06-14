#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${KIND_CLUSTER_NAME:-kind}"
MANIFEST="${K8S_MANIFEST:-k8s-default-namespace.yaml}"

services=(
  "user-service"
  "order-service"
  "notification-service"
)

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_command docker
require_command kind
require_command kubectl

if ! docker info >/dev/null 2>&1; then
  echo "Docker is not running. Start Docker Desktop and try again." >&2
  exit 1
fi

if ! kind get clusters | grep -qx "$CLUSTER_NAME"; then
  echo "kind cluster '$CLUSTER_NAME' was not found. Creating it..."
  kind create cluster --name "$CLUSTER_NAME"
fi

kubectl config use-context "kind-$CLUSTER_NAME"

for service in "${services[@]}"; do
  image="go-kube-${service}:latest"
  dockerfile="services/${service}/Dockerfile"

  echo "Building ${image}..."
  docker build -t "$image" -f "$dockerfile" .

  echo "Loading ${image} into kind cluster '${CLUSTER_NAME}'..."
  kind load docker-image "$image" --name "$CLUSTER_NAME"
done

echo "Applying Kubernetes manifest..."
kubectl apply -f "$MANIFEST"

echo "Waiting for workloads to roll out..."
kubectl rollout status deployment/postgres --timeout=180s
kubectl rollout status statefulset/zookeeper --timeout=240s
kubectl rollout status statefulset/kafka --timeout=300s
kubectl rollout status deployment/user-service --timeout=180s
kubectl rollout status deployment/order-service --timeout=180s
kubectl rollout status deployment/notification-service --timeout=180s

echo "Deployment complete."
kubectl get pods
kubectl get services

