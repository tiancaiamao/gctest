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
