#!/usr/bin/env bash
set -e

go test -c
cp longtxn.test ../script/bin/longtxn.test

docker run --rm -it -p 4000:4000  -v ../script/bin:/root/bin -v ../script/conf:/root/conf -v ./data:/root/data -w /root rockylinux:9 bash


../script/start_cluster.sh 3

# run test
go test -v -timeout 0 ./...

../script/stop_cluster.sh
