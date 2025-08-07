#!/usr/bin/env bash
set -e

N=3

for i in $(seq 0 $((N-1))); do
  name=pd-${i}
  data_dir=./data/${name}
  pid_file=${data_dir}/pd.pid

  if [[ -f "$pid_file" ]]; then
    pid=$(cat "$pid_file")
    echo "Stopping $name (pid=$pid)..."
    kill "$pid" || true
    rm -f "$pid_file"
  else
    echo "PID file not found for $name"
  fi
done
