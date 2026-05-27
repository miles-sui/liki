#!/bin/bash
# 25types 部署脚本 — 本地构建 Docker 镜像后上传到服务器部署
#
# 用法:
#   ./deploy.sh
set -eo pipefail

SERVER="43.130.2.209"
SERVER_USER="ubuntu"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
IMAGE_FILE="/tmp/25types_images.tar.gz"

echo "=========================================="
echo "  25types 部署脚本"
echo "=========================================="
echo ""
echo "服务器: $SERVER_USER@$SERVER"
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
echo "[1/6] 构建前端..."
cd "$PROJECT_DIR/web"
MINGLI_API_HOST=${MINGLI_API_HOST:-https://api.tokflux.com} npm run build
cd "$PROJECT_DIR"

echo ""
echo "[2/6] 本地构建 Docker 镜像..."

# Backend from source
docker compose -f deploy/app/docker-compose.prod.yml --env-file .env build backend

# Caddy with Cloudflare DNS plugin
docker compose -f deploy/app/docker-compose.prod.yml --env-file .env build caddy

echo ""
echo "[3/6] 保存并上传镜像..."
IMAGES="25types-backend:latest 25types-caddy:latest"
docker save $IMAGES | gzip > "$IMAGE_FILE"
echo "镜像文件: $IMAGE_FILE ($(ls -lh $IMAGE_FILE | awk '{print $5}'))"

$SCP_CMD "$IMAGE_FILE" "$SERVER_USER@$SERVER:/tmp/"
$SCP_CMD "$PROJECT_DIR/deploy/app/docker-compose.prod.yml" "$SERVER_USER@$SERVER:/tmp/docker-compose.prod.yml"
$SCP_CMD "$PROJECT_DIR/deploy/caddy/Caddyfile" "$SERVER_USER@$SERVER:/tmp/Caddyfile"
$SCP_CMD "$PROJECT_DIR/deploy/caddy/routes" "$SERVER_USER@$SERVER:/tmp/routes"
$SCP_CMD "$PROJECT_DIR/deploy/caddy/static_routes" "$SERVER_USER@$SERVER:/tmp/static_routes"

# Upload .env if present locally
if [ -f "$PROJECT_DIR/.env" ]; then
    $SCP_CMD "$PROJECT_DIR/.env" "$SERVER_USER@$SERVER:/tmp/25types.env"
    echo ".env 已上传"
else
    echo "注意: 本地无 .env 文件，服务器将使用已有配置"
fi

echo ""
echo "[4/6] 服务器加载镜像并启动..."
$SSH_CMD "$SERVER_USER@$SERVER" 'bash -s' << 'ENDSSH'
set -e

# Use sudo for docker commands if needed
docker ps >/dev/null 2>&1 && D="docker" || D="sudo docker"

# Stop any previous 25types deployment
cd /opt/25types 2>/dev/null && $D compose down 2>/dev/null || true
sleep 2  # wait for ports to release

# Stop old containers on ports 80/443 (but not other projects' containers)
for p in 80 443; do
  CIDS=$($D ps -q --filter "publish=$p" 2>/dev/null || true)
  for cid in $CIDS; do
    NAME=$($D inspect --format '{{.Name}}' "$cid" 2>/dev/null | sed 's,^/,,')
    echo "Stopping container: $NAME (port $p)"
    $D stop "$cid" && $D rm "$cid" || true
  done
done

# Preserve database across redeployments (files owned by root: use sudo)
TMP_DIR="/tmp/25types_tmp_$$"
mkdir -p "$TMP_DIR"
if [ -f /opt/25types/db/25types.db ]; then
    cp /opt/25types/db/25types.db "$TMP_DIR/25types.db"
fi
if [ -f /opt/25types/.env ]; then
    cp /opt/25types/.env "$TMP_DIR/.env"
fi

sudo mv /opt/25types /opt/25types.bak 2>/dev/null || true
sudo mkdir -p /opt/25types
sudo mv /tmp/docker-compose.prod.yml /opt/25types/docker-compose.yml
sudo mv /tmp/Caddyfile /opt/25types/Caddyfile
sudo mv /tmp/routes /opt/25types/routes
sudo mv /tmp/static_routes /opt/25types/static_routes
if [ -f /tmp/25types.env ]; then
    sudo mv /tmp/25types.env /opt/25types/.env
elif [ -f "$TMP_DIR/.env" ]; then
    sudo cp "$TMP_DIR/.env" /opt/25types/.env
fi
sudo mkdir -p /opt/25types/db
if [ -f "$TMP_DIR/25types.db" ]; then
    sudo mv "$TMP_DIR/25types.db" /opt/25types/db/25types.db
fi
sudo chown -R ubuntu:ubuntu /opt/25types
sudo chown -R 65534:65534 /opt/25types/db
# Ensure container user (65534) can write: directory needs execute+write, file needs write
sudo chmod 755 /opt/25types/db
if [ -f /opt/25types/db/25types.db ]; then
    sudo chmod 644 /opt/25types/db/25types.db
fi
# Verify write access before starting Docker
if [ -f /opt/25types/db/25types.db ]; then
    if sudo -u nobody test -w /opt/25types/db/25types.db; then
        echo "数据库文件权限检查通过 (uid 65534 可写)"
    else
        echo "错误: 数据库文件不可写 (uid 65534)，检查文件权限："
        ls -la /opt/25types/db/
        exit 1
    fi
else
    # First deploy — directory must be writable so server can create the DB
    if sudo -u nobody test -w /opt/25types/db; then
        echo "数据库目录权限检查通过 (uid 65534 可写)"
    else
        echo "错误: 数据库目录不可写 (uid 65534)，检查目录权限："
        ls -la /opt/25types/
        exit 1
    fi
fi
rm -rf "$TMP_DIR"

cd /opt/25types

echo "加载镜像..."
$D load < /tmp/25types_images.tar.gz

echo "启动服务..."
$D compose up -d

echo "等待后端就绪..."
READY=0
for i in $(seq 1 30); do
    # Check Docker's own health status first
    STATUS=$($D inspect --format='{{.State.Health.Status}}' 25types-backend-1 2>/dev/null || echo "starting")
    if [ "$STATUS" = "healthy" ]; then
        echo "后端已就绪 ($((i*2))s)"
        READY=1
        break
    elif [ "$STATUS" = "unhealthy" ]; then
        echo "错误: Docker 健康检查失败，查看日志："
        $D compose logs --tail=30
        exit 1
    fi
    sleep 2
done
if [ "$READY" = "0" ]; then
    echo "错误: 后端在 60s 内未能就绪，当前状态："
    $D inspect --format='{{.State.Health.Status}}' 25types-backend-1 2>/dev/null || echo "unknown"
    $D compose logs --tail=30
    exit 1
fi

echo "等待 Caddy 启动..."
sleep 3

echo "服务状态："
$D compose ps

sudo rm -rf /opt/25types.bak
rm -f /tmp/25types_images.tar.gz
ENDSSH

rm -f "$IMAGE_FILE"
unset SSHPASS 2>/dev/null || true

echo ""
echo "[5/6] 健康检查..."
# Backend port 8080 is internal (expose, not ports) — only Caddy (80/443) is published.
# Use --connect-timeout to avoid hanging on unreachable ports.
if curl -sf --connect-timeout 5 "https://25types.com/api/health" >/dev/null 2>&1; then
    echo "API 健康检查通过 (HTTPS 25types.com)"
elif curl -sfk --connect-timeout 5 "https://$SERVER/api/health" >/dev/null 2>&1; then
    echo "API 健康检查通过 (HTTPS $SERVER)"
else
    echo "API 健康检查失败 — 服务可能仍在启动中"
fi

echo ""
echo "[6/6] 烟雾测试..."
if [ -f "$SCRIPT_DIR/smoke.sh" ]; then
    "$SCRIPT_DIR/smoke.sh" "https://25types.com" || echo "烟雾测试有失败项，详见上方输出"
else
    echo "smoke.sh 未找到，跳过"
fi

echo ""
echo "部署完成。查看日志: ssh $SERVER_USER@$SERVER 'cd /opt/25types && docker compose logs -f'"
