#!/usr/bin/env bash
set -e

# docker get latest image
docker pull us-docker.pkg.dev/pingcap-testing-account/hub/tikv/pd/image:master-next-gen_linux_amd64
# docker pull us-docker.pkg.dev/pingcap-testing-account/hub/tikv/tikv/image:release-nextgen-20250815-67eb240-next-gen_linux_amd64
# docker pull us-docker.pkg.dev/pingcap-testing-account/hub/pingcap/tidb/images/tidb-server:master-327a22d-next-gen_linux_amd64

# take out tidb/tikv/pd binary from the docker image
docker run --rm -v $(pwd)/bin:/pd-bin --entrypoint /bin/sh us-docker.pkg.dev/pingcap-testing-account/hub/tikv/pd/image:master-next-gen_linux_amd64 -c 'cp /pd-server /pd-bin/pd-server'

# start pd server
N=3
BASE_PORT=2379
CLUSTER=""
for i in $(seq 0 $((N-1))); do
  name=pd-${i}
  peer_port=$((2380 + i * 2))
  CLUSTER+="${name}=http://127.0.0.1:${peer_port},"
done
CLUSTER=${CLUSTER%,}  # remove trailing comma


 # ./bin/pd-server --name=pd-0 --config=./conf/pd.toml --data-dir=./data --peer-urls=http://0.0.0.0:2380 --client-urls=http://0.0.0.0:2379 --log-file=./data/pd.log

for i in $(seq 0 $((N-1))); do
  name=pd-${i}
  peer_port=$((2380 + i * 2))
  client_port=$((2379 + i * 2))
  data_dir=./data/${name}
  mkdir -p "$data_dir"

  echo "Starting $name..."

  ./bin/pd-server \
    --name=${name} \
    --config=./conf/pd.toml \
    --data-dir=${data_dir} \
    --peer-urls=http://127.0.0.1:${peer_port} \
    --client-urls=http://127.0.0.1:${client_port} \
    --advertise-client-urls=http://127.0.0.1:${client_port} \
    --advertise-peer-urls=http://127.0.0.1:${peer_port} \
    --initial-cluster=${CLUSTER} \
    --log-file=${data_dir}/pd.log > ${data_dir}/stdout.log 2>&1 &
  
  echo $! > ${data_dir}/pd.pid
done

# wait pd server online?
echo "Waiting for PD cluster leader..."
while ! curl -sf http://127.0.0.1:2379/pd/api/v1/leader >/dev/null; do
  sleep 1
done
echo "PD cluster is online!"

# run test
go run main.go

./stop_pd_cluster.sh
