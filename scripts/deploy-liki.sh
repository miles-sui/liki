#!/bin/bash
# Liki 部署脚本 — 本地构建 Docker 镜像后上传到服务器部署
#
# 用法:
#   ./scripts/deploy-liki.sh        # 部署到两台服务器 (us + cn)
#   ./scripts/deploy-liki.sh us     # 仅海外
#   ./scripts/deploy-liki.sh cn     # 仅国内
#
# 环境变量:
#   LIKI_SERVER       海外服务器 IP (默认: 43.130.2.209, ubuntu)
#   LIKI_SERVER_CN    国内服务器 IP (默认: 120.79.194.247, root)
set -eo pipefail

SERVER="${LIKI_SERVER:-43.130.2.209}"
SERVER_CN="${LIKI_SERVER_CN:-120.79.194.247}"
SERVER_USER="${LIKI_SERVER_USER:-ubuntu}"
SERVER_USER_CN="${LIKI_SERVER_USER_CN:-root}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
IMAGE_FILE="/tmp/liki_images.tar.gz"

# 从 .env 读取各目标的域名和回调 URL
DOMAIN_US=$(/bin/grep -oP '^DOMAIN_US=\K.*' "$PROJECT_DIR/.env" 2>/dev/null || echo "")
DOMAIN_CN=$(/bin/grep -oP '^DOMAIN_CN=\K.*' "$PROJECT_DIR/.env" 2>/dev/null || echo "")
# 校验必须变量非空
fail_missing() {
  echo "ERROR: $1 is not set. Add it to $PROJECT_DIR/.env" >&2
  exit 1
}
[ -n "$DOMAIN_US" ] || fail_missing DOMAIN_US
[ -n "$DOMAIN_CN" ] || fail_missing DOMAIN_CN

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
	export BUILD_TIME=$(date '+%Y-%m-%d %H:%M:%S')
	node web/scripts/compile-vue-template.cjs
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -X 'main.BuildTime=$BUILD_TIME'" -o bin/liki ./cmd/liki/
	docker compose -f deploy/liki/docker-compose.yml build liki
	docker save liki:latest | gzip > "$IMAGE_FILE"

deploy() {
  local target="$1" server="$2" user="$3" mode="$4"

  echo ""
  echo "=========================================="
  echo "  Liki 部署 — $mode"
  echo "=========================================="
  echo "服务器: $user@$server"
  echo ""

  # 确定当前目标的域名
  local DOMAIN
  case "$target" in
    us) DOMAIN="$DOMAIN_US" ;;
    cn) DOMAIN="$DOMAIN_CN" ;;
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
  (cd "$TMP_WEB" && tar czf /tmp/liki_web.tar.gz .)
  rm -rf "$TMP_WEB" "$PROJECT_DIR/web/build.txt"
  # 配置 + 前端 + .env 打成一个包，一次 SCP
  if [ -f "$PROJECT_DIR/.env" ]; then
    tar czf /tmp/liki_configs.tar.gz \
      -C "$PROJECT_DIR/deploy/liki" docker-compose.yml Caddyfile \
      -C /tmp liki_web.tar.gz \
      -C "$PROJECT_DIR" .env
  else
    tar czf /tmp/liki_configs.tar.gz \
      -C "$PROJECT_DIR/deploy/liki" docker-compose.yml Caddyfile \
      -C /tmp liki_web.tar.gz
  fi
  $SCP_CMD "$IMAGE_FILE" /tmp/liki_configs.tar.gz "$user@$server:/tmp/"
  rm -f /tmp/liki_web.tar.gz /tmp/liki_configs.tar.gz

  echo "[3/4] 服务器部署..."
  # 通过命令行传入 DOMAIN 作为远程环境变量
  $SSH_CMD "$user@$server" "DOMAIN='$DOMAIN' bash -s" << 'ENDSSH'
	set -e
	docker ps >/dev/null 2>&1 && D="docker" || D="sudo -E docker"

	sudo mkdir -p /opt/liki
	sudo chown "$(whoami)" /opt/liki

	# 停止旧容器，释放端口
	if [ -f /opt/liki/docker-compose.yml ]; then
	  cd /opt/liki && $D compose down --timeout 10 2>&1 || true
	fi

	# 确保端口 80/443 释放（Docker 容器残留 + 非 Docker 进程）
	for port in 80 443; do
	  for i in $(seq 1 5); do
	    cid=$($D ps -q --filter "publish=$port" 2>/dev/null || true)
	    if [ -n "$cid" ]; then
	      echo "强制停止占用端口 $port 的容器: $cid"
	      $D stop "$cid" 2>/dev/null || true
	      $D rm -f "$cid" 2>/dev/null || true
	      sleep 1
	      continue
	    fi
	    pid=$(sudo lsof -ti :$port 2>/dev/null || true)
	    if [ -n "$pid" ]; then
	      echo "清除占用端口 $port 的旧进程: $pid"
	      sudo kill -9 $pid 2>/dev/null || true
	      sleep 1
	      continue
	    fi
	    break
	  done
	done

	# 解出配置包
	tar xzf /tmp/liki_configs.tar.gz -C /tmp/
mv /tmp/docker-compose.yml /opt/liki/docker-compose.yml
mv /tmp/Caddyfile /opt/liki/Caddyfile
[ -f /tmp/.env ] && mv /tmp/.env /opt/liki/.env
rm -f /tmp/liki_configs.tar.gz

# 将前端文件写入 Docker 命名卷 (liki_web_data)
if [ -f /tmp/liki_web.tar.gz ]; then
  $D volume create liki_web_data 2>/dev/null || true
  $D run --rm -i -v liki_web_data:/dest alpine sh -c "find /dest -mindepth 1 -delete 2>/dev/null; tar xzf - -C /dest" < /tmp/liki_web.tar.gz
  rm -f /tmp/liki_web.tar.gz
fi

# 确保 DB volume 对容器内的 liki 用户 (uid 1000) 可写
$D volume create liki_data 2>/dev/null || true
$D run --rm -v liki_data:/data alpine chown 1000:1000 /data

cd /opt/liki
$D load < /tmp/liki_images.tar.gz
$D compose up -d
$D image prune -f

echo "等待服务就绪..."
sleep 5
$D compose ps

rm -f /tmp/liki_images.tar.gz
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
