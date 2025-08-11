#!/usr/bin/env bash
set -e

# this is the file of docker entry

# start the cluster first
/root/script/start.sh

/root/bin/longtxn.test -test.timeout 0 > /root/data/longtxn.test.log
