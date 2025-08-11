#!/usr/bin/env bash
set -e

CURR_DIR="$(cd "$(dirname "$0")" && pwd)"

# start to run test in docker
docker run --rm -it \
       -p 4000:4000 \
       -v "$CURR_DIR/script":/root/script \
       -v "$CURR_DIR/../bin":/root/bin \
       -v "$CURR_DIR/conf":/root/conf \
       -v "$CURR_DIR/data":/root/data \
       -w /root \
       rockylinux:9 bash
