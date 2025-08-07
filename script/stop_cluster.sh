#!/usr/bin/env bash
set -e

PID_FILE="./playground.pid"
if [[ -f "$PID_FILE" ]]; then
  PID=$(cat "$PID_FILE")
  echo "Stopping TiUP Playground (PID=$PID)..."
  kill "$PID"
  rm -f "$PID_FILE"
else
  echo "No running TiUP playground found."
fi
