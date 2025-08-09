#!/usr/bin/env bash
set -e

CURR_DIR="$(cd "$(dirname "$0")" && pwd)"

# tidb/tikv/pd binaries should have been ready before call this script

# prepare test binary
go build -o ../script/bin/fuzz.test

# start to run test in docker
docker run --rm -it \
       -p 4000:4000 \
       -v "$CURR_DIR/docker/script":/root/script \
       -v "$CURR_DIR/../script/bin":/root/bin \
       -v "$CURR_DIR/docker/conf":/root/conf \
       -v "$CURR_DIR/docker/data":/root/data \
       -w /root \
       --entrypoint /root/script/entry.sh \
       rockylinux:9
