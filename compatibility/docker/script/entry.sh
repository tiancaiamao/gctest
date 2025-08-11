#!/usr/bin/env bash
set -e

# this is the file of docker entry

# start the cluster first
/root/script/start.sh &

# Wait for TiDB to be online or TiUP crash
echo "Waiting for TiDB (port 4000) to be online..."
MAX_WAIT=60   # 最多等待时间，单位秒
WAITED=0

while ! (echo > /dev/tcp/127.0.0.1/4000) >/dev/null 2>&1; do
    # 如果端口已开放，表示启动成功
    # 超时控制（可选）
    if [ "$WAITED" -ge "$MAX_WAIT" ]; then
	echo "❌ Timeout waiting for TiDB to start."
	exit 1
    fi
    echo "Waiting... (${WAITED}s)"
    sleep 1
    WAITED=$((WAITED + 1))
done

echo "✅ TiDB is online."

/root/bin/compatibility.test > /root/data/compatibility.test.log
