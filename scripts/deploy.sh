#!/bin/bash

set -e

# 配置變量
ENVIRONMENT=${1:-development}
VERSION=${2:-latest}
NAMESPACE="warehouse-system"

echo "🚀 Deploying Warehouse Management System"
echo "Environment: $ENVIRONMENT"
echo "Version: $VERSION"

# 確保必要的工具已安裝
command -v kubectl >/dev/null 2>&1 || { echo "kubectl is required but not installed. Aborting." >&2; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "docker is required but not installed. Aborting." >&2; exit 1; }

# 函數：構建和推送 Docker 鏡像
build_and_push() {
    SERVICE_NAME=$1
    DOCKER_FILE_PATH=$2
    
    echo "📦 Building $SERVICE_NAME..."
    docker build -t warehouse/$SERVICE_NAME:$VERSION -f $DOCKER_FILE_PATH .
    
    if [ "$ENVIRONMENT" != "development" ]; then
        echo "📤 Pushing $SERVICE_NAME to registry..."
        docker push warehouse/$SERVICE_NAME:$VERSION
    fi
}

# 構建服務鏡像
echo "🔨 Building Docker images..."

build_and_push "inventory-service" "services/inventory-service/Dockerfile"
build_and_push "location-service" "services/location-service/Dockerfile"
build_and_push "realtime-service" "services/realtime-service/Dockerfile"
build_and_push "worker-app" "frontend/worker-app/Dockerfile"
build_and_push "admin-dashboard" "frontend/admin-dashboard/Dockerfile"

# 應用 Kubernetes 配置
echo "☸️  Applying Kubernetes configurations..."

# 創建命名空間
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# 應用基礎設施組件
kubectl apply -f infrastructure/kubernetes/postgres/ -n $NAMESPACE
kubectl apply -f infrastructure/kubernetes/mongodb/ -n $NAMESPACE
kubectl apply -f infrastructure/kubernetes/redis/ -n $NAMESPACE
kubectl apply -f infrastructure/kubernetes/kafka/ -n $NAMESPACE

# 等待基礎設施就緒
echo "⏳ Waiting for infrastructure to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app=mongodb -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app=redis -n $NAMESPACE --timeout=300s
kubectl wait --for=condition=ready pod -l app=kafka -n $NAMESPACE --timeout=300s

# 應用微服務
echo "🚀 Deploying microservices..."
kubectl apply -f infrastructure/kubernetes/inventory-service/ -n $NAMESPACE
kubectl apply -f infrastructure/kubernetes/location-service/ -n $NAMESPACE
kubectl apply -f infrastructure/kubernetes/realtime-service/ -n $NAMESPACE

# 應用前端應用
kubectl apply -f infrastructure/kubernetes/frontend/ -n $NAMESPACE

# 應用 API Gateway
kubectl apply -f infrastructure/kubernetes/api-gateway/ -n $NAMESPACE

# 應用監控
kubectl apply -f infrastructure/kubernetes/monitoring/ -n $NAMESPACE

# 等待部署完成
echo "⏳ Waiting for deployments to be ready..."
kubectl wait --for=condition=available deployment --all -n $NAMESPACE --timeout=600s

# 獲取服務訪問信息
echo "✅ Deployment completed successfully!"
echo ""
echo "📋 Service URLs:"
echo "API Gateway: $(kubectl get svc api-gateway -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')"
echo "Worker App: $(kubectl get svc worker-app -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')"
echo "Admin Dashboard: $(kubectl get svc admin-dashboard -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')"
echo "Grafana: $(kubectl get svc grafana -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')"
echo ""
echo "🔍 To check deployment status:"
echo "kubectl get pods -n $NAMESPACE"
echo ""
echo "📊 To view logs:"
echo "kubectl logs -f deployment/inventory-service -n $NAMESPACE"