#!/usr/bin/env bash
set -e

../script/start_cluster.sh

# run test
go run main.go

../script/stop_cluster.sh
