#!/usr/bin/env bash
set -e

BIN_DIR=/root/bin
DATA_DIR=/root/data
CONF_DIR=/root/conf

rm -rf $DATA_DIR/*

mkdir -p $DATA_DIR/pd-0
$BIN_DIR/pd-server --name=pd-0 --config=$CONF_DIR/pd-0.toml --data-dir=$DATA_DIR/pd-0/data --peer-urls=http://127.0.0.1:2380 --advertise-peer-urls=http://127.0.0.1:2380 --client-urls=http://127.0.0.1:2379 --advertise-client-urls=http://127.0.0.1:2379 --log-file=$DATA_DIR/pd-0/pd.log --initial-cluster=pd-0=http://127.0.0.1:2380 > $DATA_DIR/pd-0/pd_stderr 2>&1 &

sleep 1;

mkdir -p $DATA_DIR/tikv-0
$BIN_DIR/tikv-server --addr=127.0.0.1:20160 --advertise-addr=127.0.0.1:20160 --status-addr=127.0.0.1:20180 --pd-endpoints=http://127.0.0.1:2379 --config=$CONF_DIR/tikv-0.toml --data-dir=$DATA_DIR/tikv-0/data --log-file=$DATA_DIR/tikv-0/tikv.log  > $DATA_DIR/tikv-0/tikv_stderr 2>&1 &

sleep 2;

mkdir -p $DATA_DIR/tidb-0
$BIN_DIR/tidb-server --store=tikv --host=0.0.0.0 --path=127.0.0.1:2379 --log-file=$DATA_DIR/tidb-0/tidb.log --config=$CONF_DIR/tidb-0.toml > $DATA_DIR/tidb-0/tidb_stderr 2>&1 &

sleep 1;

mkdir -p $DATA_DIR/tidb-1
$BIN_DIR/tidb-server --store=tikv --host=0.0.0.0 --path=127.0.0.1:2379 --log-file=$DATA_DIR/tidb-1/tidb.log --config=$CONF_DIR/tidb-1.toml > $DATA_DIR/tidb-1/tidb_stderr 2>&1 &

mkdir -p $DATA_DIR/tidb-2
$BIN_DIR/tidb-server --store=tikv --host=0.0.0.0 --path=127.0.0.1:2379 --log-file=$DATA_DIR/tidb-2/tidb.log --config=$CONF_DIR/tidb-2.toml > $DATA_DIR/tidb-2/tidb_stderr 2>&1 &

mkdir -p $DATA_DIR/tidb-3
$BIN_DIR/tidb-server --store=tikv --host=0.0.0.0 --path=127.0.0.1:2379 --log-file=$DATA_DIR/tidb-3/tidb.log --config=$CONF_DIR/tidb-3.toml > $DATA_DIR/tidb-3/tidb_stderr 2>&1 &

mkdir -p $DATA_DIR/tidb-4
$BIN_DIR/tidb-server --store=tikv --host=0.0.0.0 --path=127.0.0.1:2379 --log-file=$DATA_DIR/tidb-4/tidb.log --config=$CONF_DIR/tidb-4.toml > $DATA_DIR/tidb-4/tidb_stderr 2>&1 &

mkdir -p $DATA_DIR/tidb-5
$BIN_DIR/tidb-server --store=tikv --host=0.0.0.0 --path=127.0.0.1:2379 --log-file=$DATA_DIR/tidb-5/tidb.log --config=$CONF_DIR/tidb-5.toml > $DATA_DIR/tidb-5/tidb_stderr 2>&1 &

mkdir -p $DATA_DIR/tidb-6
$BIN_DIR/tidb-server --store=tikv --host=0.0.0.0 --path=127.0.0.1:2379 --log-file=$DATA_DIR/tidb-6/tidb.log --config=$CONF_DIR/tidb-6.toml > $DATA_DIR/tidb-6/tidb_stderr 2>&1 &

sleep 3;

# Wait for TiDB to be online or TiUP crash
echo "Waiting for TiDB to be online..."
MAX_WAIT=60   # 最多等待时间，单位秒
WAITED=0

while ! (echo > /dev/tcp/127.0.0.1/4001) >/dev/null 2>&1; do
    # 如果端口已开放，表示启动成功
    # 超时控制（可选）
    if [ "$WAITED" -ge "$MAX_WAIT" ]; then
	echo "❌ Timeout waiting for TiDB1 to start."
	exit 1
    fi
    echo "Waiting... (${WAITED}s)"
    sleep 1
    WAITED=$((WAITED + 1))
done

while ! (echo > /dev/tcp/127.0.0.1/4003) >/dev/null 2>&1; do
    # 如果端口已开放，表示启动成功
    # 超时控制（可选）
    if [ "$WAITED" -ge "$MAX_WAIT" ]; then
	echo "❌ Timeout waiting for TiDB3 to start."
	exit 1
    fi
    echo "Waiting... (${WAITED}s)"
    sleep 1
    WAITED=$((WAITED + 1))
done

while ! (echo > /dev/tcp/127.0.0.1/4005) >/dev/null 2>&1; do
    # 如果端口已开放，表示启动成功
    # 超时控制（可选）
    if [ "$WAITED" -ge "$MAX_WAIT" ]; then
	echo "❌ Timeout waiting for TiDB5 to start."
	exit 1
    fi
    echo "Waiting... (${WAITED}s)"
    sleep 1
    WAITED=$((WAITED + 1))
done

echo "✅ TiDB is online."
