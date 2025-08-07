#!/usr/bin/env bash
set -e

PD_COUNT="${1:-1}"  # 默认 1 个 PD，传 3 启 3 个
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"
CONF_DIR="$SCRIPT_DIR/conf"

# Pull images
echo "Pulling docker images..."
docker pull us-docker.pkg.dev/pingcap-testing-account/hub/tikv/pd/image:master-next-gen_linux_amd64
docker pull us-docker.pkg.dev/pingcap-testing-account/hub/tikv/tikv/image:release-nextgen-20250815-67eb240-next-gen_linux_amd64
docker pull us-docker.pkg.dev/pingcap-testing-account/hub/pingcap/tidb/images/tidb-server:master-327a22d-next-gen_linux_amd64

# Extract binaries
mkdir -p "$BIN_DIR"

echo "Extracting binaries..."
CID=$(docker create us-docker.pkg.dev/pingcap-testing-account/hub/tikv/pd/image:master-next-gen_linux_amd64)
docker cp "$CID:/pd-server" $BIN_DIR/pd-server
docker rm "$CID"

CID=$(docker create us-docker.pkg.dev/pingcap-testing-account/hub/tikv/tikv/image:release-nextgen-20250815-67eb240-next-gen_linux_amd64)
docker cp "$CID:/tikv-server" $BIN_DIR/tikv-server
docker rm "$CID"

CID=$(docker create us-docker.pkg.dev/pingcap-testing-account/hub/pingcap/tidb/images/tidb-server:master-327a22d-next-gen_linux_amd64)
docker cp "$CID:/tidb-server" $BIN_DIR/tidb-server
docker rm "$CID"


# docker run --rm -v "$BIN_DIR":/pd-bin --entrypoint /bin/sh us-docker.pkg.dev/pingcap-testing-account/hub/tikv/pd/image:master-next-gen_linux_amd64 -c 'cp /pd-server /pd-bin/pd-server'
# docker run --rm -v "$BIN_DIR":/tikv-bin --entrypoint /bin/sh us-docker.pkg.dev/pingcap-testing-account/hub/tikv/tikv/image:release-nextgen-20250815-67eb240-next-gen_linux_amd64 -c 'cp /tikv-server /tikv-bin/tikv-server'
# docker run --rm -v "$BIN_DIR":/tidb-bin --entrypoint /bin/sh us-docker.pkg.dev/pingcap-testing-account/hub/pingcap/tidb/images/tidb-server:master-327a22d-next-gen_linux_amd64 -c 'cp /tidb-server /tidb-bin/tidb-server'

# Start TiUP playground
echo "Starting TiUP playground with PD count = $PD_COUNT..."
tiup playground \
  --host 0.0.0.0 \
  --pd $PD_COUNT \
  --tiflash 0 \
  --pd.binpath="$BIN_DIR/pd-server" \
  --kv.binpath="$BIN_DIR/tikv-server" \
  --db.binpath="$BIN_DIR/tidb-server" \
  --pd.config="$CONF_DIR/pd.toml" \
  --kv.config="$CONF_DIR/tikv.toml" \
  --db.config="$CONF_DIR/tidb.toml" > playground.log 2>&1 &

TIUP_PID=$!
echo $TIUP_PID > playground.pid
echo "TiUP Playground PID: $TIUP_PID"

# Wait for TiDB to be online or TiUP crash
echo "Waiting for TiDB (port 4000) to be online..."
MAX_WAIT=60   # 最多等待时间，单位秒
WAITED=0

while true; do
  # 如果端口已开放，表示启动成功
  if nc -z 127.0.0.1 4000; then
    echo "✅ TiDB is online."
    break
  fi

  # 如果 TiUP playground 进程已退出，说明启动失败
  if ! kill -0 "$TIUP_PID" 2>/dev/null; then
    echo "❌ TiUP playground process exited early. Check playground.log for details. Last log lines:"
    tail -n 20 playground.log
    exit 1
  fi

  # 超时控制（可选）
  if [ "$WAITED" -ge "$MAX_WAIT" ]; then
    echo "❌ Timeout waiting for TiDB to start."
    exit 1
  fi

  sleep 1
  WAITED=$((WAITED + 1))
done
