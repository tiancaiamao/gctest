#!/usr/bin/env bash
set -e

../script/start_cluster.sh 3

# run test
go test -v -timeout 0 ./...

../script/stop_cluster.sh
