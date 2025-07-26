#!/bin/bash

set -e

echo "🛠️  Setting up development environment..."

# 檢查 Docker 和 Docker Compose
command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Aborting." >&2; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose is required but not installed. Aborting." >&2; exit 1; }

# 創建必要的目錄
mkdir -p logs/{inventory,location,realtime}
mkdir -p data/{postgres,mongodb,redis}

# 生成開發環境配置
echo "📝 Generating development configuration..."

# 創建 .env 文件
cat > .env << EOF
# Database
POSTGRES_DB=warehouse
POSTGRES_USER=admin
POSTGRES_PASSWORD=devpassword123

# Redis
REDIS_PASSWORD=

# Kafka
KAFKA_BROKERS=localhost:9092

# JWT
JWT_SECRET=dev-secret-key-change-