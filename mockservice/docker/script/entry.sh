#!/usr/bin/env bash
set -e

# this is the file of docker entry

# start the cluster first
/root/script/start.sh

# change it to something else
/root/bin/mockservice.test > /root/data/mockservice.test.log
