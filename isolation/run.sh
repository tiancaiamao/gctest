#!/usr/bin/env bash
set -e

../script/start_cluster.sh

../script/bin/tidb-server -config conf/tidb1.toml > tidb1.log &
../script/bin/tidb-server -config conf/tidb2.toml > tidb2.log &
../script/bin/tidb-server -config conf/tidb3.toml > tidb3.log &
../script/bin/tidb-server -config conf/tidb4.toml > tidb4.log &
../script/bin/tidb-server -config conf/tidb5.toml > tidb5.log &
../script/bin/tidb-server -config conf/tidb6.toml > tidb6.log &

# run test
go test -v ./...

pkill -f tidb-server
../script/stop_cluster.sh
