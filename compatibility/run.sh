#!/usr/bin/env bash
set -e
CURR_DIR="$(cd "$(dirname "$0")" && pwd)"

# tidb/tikv/pd binaries should have been ready before call this script

# prepare test binary
go build -o ../script/bin/compatibility.test
go build -o ../script/bin/oldAPI oldAPI/main.go

# start to run test in docker
docker run --rm -i \
       -v "$CURR_DIR/docker/script":/root/script \
       -v "$CURR_DIR/../script/bin":/root/bin \
       -v "$CURR_DIR/docker/conf":/root/conf \
       -v "$CURR_DIR/docker/data":/root/data \
       -w /root \
       --entrypoint /root/script/entry.sh \
       rockylinux:9


