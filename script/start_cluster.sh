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
docker run --rm -v "$BIN_DIR":/pd-bin --entrypoint /bin/sh us-docker.pkg.dev/pingcap-testing-account/hub/tikv/pd/image:master-next-gen_linux_amd64 -c 'cp /pd-server /pd-bin/pd-server'
docker run --rm -v "$BIN_DIR":/tikv-bin --entrypoint /bin/sh us-docker.pkg.dev/pingcap-testing-account/hub/tikv/tikv/image:release-nextgen-20250815-67eb240-next-gen_linux_amd64 -c 'cp /tikv-server /tikv-bin/tikv-server'
docker run --rm -v "$BIN_DIR":/tidb-bin --entrypoint /bin/sh us-docker.pkg.dev/pingcap-testing-account/hub/pingcap/tidb/images/tidb-server:master-327a22d-next-gen_linux_amd64 -c 'cp /tidb-server /tidb-bin/tidb-server'

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

# Wait for TiDB to be online
echo "Waiting for TiDB to be online..."
while ! nc -z 127.0.0.1 4000; do
  sleep 1
done
echo "✅ TiDB is online!"
