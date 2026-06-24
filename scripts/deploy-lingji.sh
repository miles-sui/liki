#!/bin/bash
# LingJi 部署脚本 — 本地构建 Docker 镜像后上传到服务器部署
#
# 用法:
#   ./scripts/deploy-lingji.sh        # 部署到两台服务器 (us + cn)
#   ./scripts/deploy-lingji.sh us     # 仅海外
#   ./scripts/deploy-lingji.sh cn     # 仅国内
#
# 环境变量:
#   LINGJI_SERVER       海外服务器 IP (默认: 43.130.2.209, ubuntu)
#   LINGJI_SERVER_CN    国内服务器 IP (默认: 120.79.194.247, root)
set -eo pipefail

SERVER="${LINGJI_SERVER:-43.130.2.209}"
SERVER_CN="${LINGJI_SERVER_CN:-120.79.194.247}"
SERVER_USER="${LINGJI_SERVER_USER:-ubuntu}"
SERVER_USER_CN="${LINGJI_SERVER_USER_CN:-root}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
IMAGE_FILE="/tmp/lingji_images.tar.gz"

# 从 .env 读取各目标的域名和回调 URL
DOMAIN_US=$(/bin/grep -oP '^DOMAIN_US=\K.*' "$PROJECT_DIR/.env" 2>/dev/null || echo "")
DOMAIN_CN=$(/bin/grep -oP '^DOMAIN_CN=\K.*' "$PROJECT_DIR/.env" 2>/dev/null || echo "")
RETURN_URL_US=$(/bin/grep -oP '^RETURN_URL_US=\K.*' "$PROJECT_DIR/.env" 2>/dev/null || echo "")
RETURN_URL_CN=$(/bin/grep -oP '^RETURN_URL_CN=\K.*' "$PROJECT_DIR/.env" 2>/dev/null || echo "")

# 校验必须变量非空
fail_missing() {
  echo "ERROR: $1 is not set. Add it to $PROJECT_DIR/.env" >&2
  exit 1
}
[ -n "$DOMAIN_US" ] || fail_missing DOMAIN_US
[ -n "$DOMAIN_CN" ] || fail_missing DOMAIN_CN
[ -n "$RETURN_URL_US" ] || fail_missing RETURN_URL_US
[ -n "$RETURN_URL_CN" ] || fail_missing RETURN_URL_CN

if ! command -v docker &>/dev/null; then
  echo "ERROR: docker not found. Install Docker and try again."
  exit 1
fi

TARGET="${1:-all}"
case "$TARGET" in
  all) TARGETS="us cn" ;;
  us)  TARGETS="us"    ;;
  cn)  TARGETS="cn"    ;;
  *)   echo "Usage: $0 [us|cn]"; exit 1 ;;
esac

	echo "[1/4] 构建 + 导出镜像..."
	export DOMAIN="$DOMAIN_US"
	export RETURN_URL="$RETURN_URL_US"
	export BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
	node web/scripts/compile-vue-template.cjs
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X 'main.BuildTime=$BUILD_TIME'" -o bin/lingji ./cmd/lingji/
	docker compose -f deploy/lingji/docker-compose.yml build lingji
	docker save lingji:latest | gzip > "$IMAGE_FILE"

deploy() {
  local target="$1" server="$2" user="$3" mode="$4"

  echo ""
  echo "=========================================="
  echo "  LingJi 部署 — $mode"
  echo "=========================================="
  echo "服务器: $user@$server"
  echo ""

  # 确定当前目标的域名和回调 URL
  local DOMAIN RETURN_URL
  case "$target" in
    us) DOMAIN="$DOMAIN_US" RETURN_URL="$RETURN_URL_US" ;;
    cn) DOMAIN="$DOMAIN_CN" RETURN_URL="$RETURN_URL_CN" ;;
  esac

  if ssh -o ConnectTimeout=5 -o BatchMode=yes -o StrictHostKeyChecking=accept-new "$user@$server" "echo ok" >/dev/null 2>&1; then
      SSH_CMD="ssh"
      SCP_CMD="scp"
  else
      read -rsp "请输入 $mode 服务器密码: " SSHPASS
      echo ""
      SSH_CMD="sshpass -e ssh -o StrictHostKeyChecking=accept-new"
      SCP_CMD="sshpass -e scp -o StrictHostKeyChecking=accept-new"
      export SSHPASS
  fi

  echo "[2/4] 打包 + 上传..."
  BUILD_TS=$(date '+%s')
  date '+%Y-%m-%d %H:%M:%S CST' > "$PROJECT_DIR/web/build.txt"
  TMP_WEB=$(mktemp -d)
  cp -r "$PROJECT_DIR/web"/* "$TMP_WEB/"
	  # Merge wiki GEO pages into /wiki/
	  if [ -d "$PROJECT_DIR/../liki_wiki/docs" ]; then
	    mkdir -p "$TMP_WEB/wiki"
	    cp -r "$PROJECT_DIR/../liki_wiki/docs"/* "$TMP_WEB/wiki/"
	  fi
  find "$TMP_WEB" -name '*.html' -exec sed -i "s|\(src=\"[/]\?js/[^\"]*\)|\1?v=$BUILD_TS|g; s|\(href=\"[/]\?css/[^\"]*\)|\1?v=$BUILD_TS|g" {} +
	find "$TMP_WEB" -name '*.html' -exec sed -i "s/BUILD_TIME_PLACEHOLDER/$BUILD_TIME/g" {} +
  date '+%Y-%m-%d %H:%M:%S CST' > "$TMP_WEB/build.txt"
  (cd "$TMP_WEB" && tar czf /tmp/lingji_web.tar.gz .)
  rm -rf "$TMP_WEB" "$PROJECT_DIR/web/build.txt"
  # 配置 + 前端 + .env 打成一个包，一次 SCP
  if [ -f "$PROJECT_DIR/.env" ]; then
    tar czf /tmp/lingji_configs.tar.gz \
      -C "$PROJECT_DIR/deploy/lingji" docker-compose.yml Caddyfile \
      -C /tmp lingji_web.tar.gz \
      -C "$PROJECT_DIR" .env
  else
    tar czf /tmp/lingji_configs.tar.gz \
      -C "$PROJECT_DIR/deploy/lingji" docker-compose.yml Caddyfile \
      -C /tmp lingji_web.tar.gz
  fi
  $SCP_CMD "$IMAGE_FILE" /tmp/lingji_configs.tar.gz "$user@$server:/tmp/"
  rm -f /tmp/lingji_web.tar.gz /tmp/lingji_configs.tar.gz

  echo "[3/4] 服务器部署..."
  # 通过命令行传入 DOMAIN 和 RETURN_URL 作为远程环境变量
  $SSH_CMD "$user@$server" "DOMAIN='$DOMAIN' RETURN_URL='$RETURN_URL' bash -s" << 'ENDSSH'
set -e
docker ps >/dev/null 2>&1 && D="docker" || D="sudo docker"

# Kill legacy processes on port 80/443
for port in 80 443; do
  pid=$(sudo lsof -ti :$port 2>/dev/null || true)
  if [ -n "$pid" ]; then
    echo "清除占用端口 $port 的旧进程: $pid"
    sudo kill -9 $pid 2>/dev/null || true
  fi
done

sudo mkdir -p /opt/lingji
sudo chown "$(whoami)" /opt/lingji
cd /opt/lingji && $D compose down 2>/dev/null || true

# 解出配置包
tar xzf /tmp/lingji_configs.tar.gz -C /tmp/
mv /tmp/docker-compose.yml /opt/lingji/docker-compose.yml
mv /tmp/Caddyfile /opt/lingji/Caddyfile
[ -f /tmp/.env ] && mv /tmp/.env /opt/lingji/.env
rm -f /tmp/lingji_configs.tar.gz

# 将前端文件写入 Docker 命名卷 (lingji_web_data)
if [ -f /tmp/lingji_web.tar.gz ]; then
  $D volume create lingji_web_data 2>/dev/null || true
  $D run --rm -i -v lingji_web_data:/dest alpine sh -c "find /dest -mindepth 1 -delete 2>/dev/null; tar xzf - -C /dest" < /tmp/lingji_web.tar.gz
  rm -f /tmp/lingji_web.tar.gz
fi

cd /opt/lingji
$D load < /tmp/lingji_images.tar.gz
$D compose up -d
$D image prune -f

echo "等待服务就绪..."
sleep 5
$D compose ps

rm -f /tmp/lingji_images.tar.gz
ENDSSH

  echo ""
  echo "[4/4] 健康检查 (https://${DOMAIN}/api/health)..."
  for i in $(seq 1 6); do
    st=$(curl -sf -o /dev/null -w "%{http_code}" --connect-timeout 5 "https://${DOMAIN}/api/health" 2>&1 || true)
    if [ "$st" = "200" ]; then
      echo "API 健康检查通过 ($mode)"
      break
    fi
    if [ "$i" -eq 6 ]; then
      echo "API 健康检查失败 ($mode), HTTP $st" >&2
      exit 1
    fi
    echo "  HTTP $st, 重试 ($i/6)..."
    sleep 5
  done

  unset SSHPASS
  echo ""
  echo "部署完成 ($mode)。"
}

for t in $TARGETS; do
  case "$t" in
    us) deploy "us" "$SERVER"    "$SERVER_USER"    "海外" ;;
    cn) deploy "cn" "$SERVER_CN" "$SERVER_USER_CN" "国内" ;;
  esac
done
rm -f "$IMAGE_FILE"
