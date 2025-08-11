#!/usr/bin/env bash
set -e

echo "fetch binary first..."
./script/fetch_binary.sh

declare -A task_dirs
task_dirs["longtxn"]="longtxn"
task_dirs["isolation"]="isolation"
task_dirs["fuzz"]="fuzz"
declare -A pid_to_name

echo "now run tests"
for name in "${!task_dirs[@]}"; do
    dir="${task_dirs[$name]}"
    if [[ -z "$dir" ]]; then
        echo "Error: No directory configured for task '$name'"
        exit 1
    fi
    (
        cd "$dir" || exit 1
        ./run.sh > run.log
    ) &
    pid=$!
    pid_to_name[$pid]="$name"
    echo "Started task '$name' in $dir (PID=$pid)"
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
