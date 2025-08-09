#!/usr/bin/env bash
set -e

BIN_DIR=/root/bin
DATA_DIR=/root/data
CONF_DIR=/root/conf

mkdir -p $DATA_DIR/pd-0
$BIN_DIR/pd-server --name=pd-0 --config=$CONF_DIR/pd-0.toml --data-dir=$DATA_DIR/pd-0/data --peer-urls=http://127.0.0.1:2380 --advertise-peer-urls=http://127.0.0.1:2380 --client-urls=http://127.0.0.1:2379 --advertise-client-urls=http://127.0.0.1:2379 --log-file=$DATA_DIR/pd-0/pd.log --initial-cluster=pd-0=http://127.0.0.1:2380 > $DATA_DIR/pd-0/pd_stderr 2>&1 &

sleep 1;

mkdir -p $DATA_DIR/tikv-0
$BIN_DIR/tikv-server --addr=127.0.0.1:20160 --advertise-addr=127.0.0.1:20160 --status-addr=127.0.0.1:20180 --pd-endpoints=http://127.0.0.1:2379 --config=$CONF_DIR/tikv-0.toml --data-dir=$DATA_DIR/tikv-0/data --log-file=$DATA_DIR/tikv-0/tikv.log  > $DATA_DIR/tikv-0/tikv_stderr 2>&1 &

sleep 2;

mkdir -p $DATA_DIR/tidb-0
$BIN_DIR/tidb-server -P 4000 --store=tikv --host=0.0.0.0 --status=10080 --path=127.0.0.1:2379 --log-file=$DATA_DIR/tidb-0/tidb.log --config=$CONF_DIR/tidb-0.toml > $DATA_DIR/tidb-0/tidb_stderr 2>&1 &

