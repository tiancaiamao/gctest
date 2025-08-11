#!/usr/bin/env bash
set -e

echo "fetch binary first..."
./script/fetch_binary.sh

declare -A tasks=(
    ["longtxn"]="./longtxn/run.sh"
    ["isolation"]="./isolation/run.sh"
    ["fuzz"]="./fuzz/run.sh"
    # ["compatibility"]="./compatibility/run.sh"
    # ["mockservice"]="./mockservice/run.sh"
)
declare -A pid_to_name

echo "now run tests"
for name in "${!tasks[@]}"; do
    "${tasks[$name]}" &
    pid=$!
    pid_to_name[$pid]="$name"
    echo "Started task '$name' (PID=$pid)"
done

echo "wait test to finish"
status=0
for pid in "${!pid_to_name[@]}"; do
    name="${pid_to_name[$pid]}"
    if wait "$pid"; then
        echo "✅ Task '$name' succeeded"
    else
        echo "❌ Task '$name' failed"
        status=1
    fi
done

exit $status
