
# Delete everything first
kubectl delete -f k8s-default-namespace.yaml

# Delete PVCs to reset the data
kubectl delete pvc --all

# Wait a moment
sleep 2

# Reapply
kubectl apply -f k8s-default-namespace.yaml

# Watch the rollout
kubectl get pods -w