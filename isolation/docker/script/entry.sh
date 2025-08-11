#!/usr/bin/env bash
set -e

# this is the file of docker entry

# start the cluster first
/root/script/start.sh

# run test
/root/bin/isolation.test -test.timeout 0 > /root/data/isolation.test.log
