#!/bin/bash
# Deploy mingli server to api.tokflux.com
#
# Usage:
#   DOMAIN=api.tokflux.com ./deploy-mingli.sh
set -eo pipefail

SERVER="api.tokflux.com"
SERVER_USER="root"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
DOMAIN="${DOMAIN:-api.tokflux.com}"
IMAGE_FILE="/tmp/mingli_images.tar.gz"

echo "=========================================="
echo "  Mingli Server 部署 — $DOMAIN"
echo "=========================================="
echo ""

if ssh -o ConnectTimeout=5 -o BatchMode=yes -o StrictHostKeyChecking=accept-new "$SERVER_USER@$SERVER" "echo ok" >/dev/null 2>&1; then
    SSH_CMD="ssh"
    SCP_CMD="scp"
    echo "SSH 密钥认证已就绪"
else
    read -rsp "请输入服务器密码: " SSHPASS
    echo ""
    SSH_CMD="sshpass -e ssh -o StrictHostKeyChecking=accept-new"
    SCP_CMD="sshpass -e scp -o StrictHostKeyChecking=accept-new"
    export SSHPASS
fi

echo ""
echo "[1/4] 本地构建 Docker 镜像..."

docker compose -f deploy/mingli/docker-compose.yml build

echo ""
echo "[2/4] 保存并上传镜像..."
docker save 25types-mingli:latest 25types-caddy:latest | gzip > "$IMAGE_FILE"
echo "镜像文件: $IMAGE_FILE ($(ls -lh $IMAGE_FILE | awk '{print $5}'))"

$SCP_CMD "$IMAGE_FILE" "$SERVER_USER@$SERVER:/tmp/"
$SCP_CMD "$PROJECT_DIR/deploy/mingli/docker-compose.yml" "$SERVER_USER@$SERVER:/tmp/docker-compose.mingli.yml"
$SCP_CMD "$PROJECT_DIR/deploy/caddy/Caddyfile.mingli" "$SERVER_USER@$SERVER:/tmp/Caddyfile.mingli"

echo ""
echo "[3/4] 服务器加载镜像并启动..."
$SSH_CMD "$SERVER_USER@$SERVER" DOMAIN="$DOMAIN" CF_API_TOKEN="${CF_API_TOKEN:-}" 'bash -s' << 'ENDSSH'
set -e

# root user, no sudo needed

# Stop any previous mingli deployment
cd /opt/mingli 2>/dev/null && docker compose down 2>/dev/null || true
sleep 2

# Stop old containers on ports 80/443
for p in 80 443; do
  CIDS=$(docker ps -q --filter "publish=$p" 2>/dev/null || true)
  for cid in $CIDS; do
    NAME=$(docker inspect --format '{{.Name}}' "$cid" 2>/dev/null | sed 's,^/,,')
    echo "Stopping container: $NAME (port $p)"
    docker stop "$cid" && docker rm "$cid" || true
  done
done

mkdir -p /opt/mingli
mv /tmp/docker-compose.mingli.yml /opt/mingli/docker-compose.yml
mv /tmp/Caddyfile.mingli /opt/mingli/Caddyfile

cd /opt/mingli

echo "加载镜像..."
docker load < /tmp/mingli_images.tar.gz

echo "启动服务 (DOMAIN=$DOMAIN)..."
DOMAIN=$DOMAIN CF_API_TOKEN=$CF_API_TOKEN docker compose up -d

echo "等待就绪..."
READY=0
for i in $(seq 1 30); do
    STATUS=$(docker inspect --format='{{.State.Health.Status}}' mingli-mingli-1 2>/dev/null || echo "starting")
    if [ "$STATUS" = "healthy" ]; then
        echo "Mingli 后端已就绪 ($((i*2))s)"
        READY=1
        break
    elif [ "$STATUS" = "unhealthy" ]; then
        echo "错误: 健康检查失败，查看日志："
        docker compose logs --tail=30
        exit 1
    fi
    sleep 2
done
if [ "$READY" = "0" ]; then
    echo "错误: 后端在 60s 内未能就绪"
    docker compose logs --tail=30
    exit 1
fi

sleep 3
echo "服务状态："
docker compose ps

rm -f /tmp/mingli_images.tar.gz
ENDSSH

rm -f "$IMAGE_FILE"
unset SSHPASS 2>/dev/null || true

echo ""
echo "[4/4] 健康检查 (等待 TLS + 服务就绪，最长 60s)..."
READY=0
for i in $(seq 1 20); do
    if curl -sf --connect-timeout 5 "https://$DOMAIN/api/health" >/dev/null 2>&1; then
        echo "Mingli API 健康检查通过 (HTTPS $DOMAIN, $((i*3))s)"
        READY=1
        break
    fi
    printf "."
    sleep 3
done
if [ "$READY" = "0" ]; then
    echo ""
    echo "API 健康检查失败 — 服务可能仍在启动中"
fi

echo ""
echo "部署完成。查看日志: ssh $SERVER_USER@$SERVER 'cd /opt/mingli && docker compose logs -f'"
