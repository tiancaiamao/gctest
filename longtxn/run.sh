# docker get latest image
docker pull us-docker.pkg.dev/pingcap-testing-account/hub/tikv/pd/image:master-next-gen_linux_amd64
docker pull us-docker.pkg.dev/pingcap-testing-account/hub/tikv/tikv/image:release-nextgen-20250815-67eb240-next-gen_linux_amd64
docker pull us-docker.pkg.dev/pingcap-testing-account/hub/pingcap/tidb/images/tidb-server:master-327a22d-next-gen_linux_amd64

# take out tidb/tikv/pd binary from the docker image
docker run --rm -v $(pwd)/bin:/pd-bin --entrypoint /bin/sh us-docker.pkg.dev/pingcap-testing-account/hub/tikv/pd/image:master-next-gen_linux_amd64 -c 'cp /pd-server /pd-bin/pd-server'
docker run --rm -v $(pwd)/bin:/tikv-bin --entrypoint /bin/sh us-docker.pkg.dev/pingcap-testing-account/hub/tikv/tikv/image:release-nextgen-20250815-67eb240-next-gen_linux_amd64 -c 'cp /tikv-server /tikv-bin/tikv-server'
docker run --rm -v $(pwd)/bin:/tidb-bin --entrypoint /bin/sh us-docker.pkg.dev/pingcap-testing-account/hub/pingcap/tidb/images/tidb-server:master-327a22d-next-gen_linux_amd64 -c 'cp /tidb-server /tidb-bin/tidb-server'

# start local cluster
tiup playground --host 0.0.0.0 --tiflash 0 --pd.binpath=$(pwd)/bin/pd-server --kv.binpath=$(pwd)/bin/tikv-server --db.binpath=$(pwd)/bin/tidb-server --pd.config conf/pd.toml --kv.config conf/tikv.toml  --db.config conf/tidb.toml

# wait until tidb online?
# run the test
go test ./...
