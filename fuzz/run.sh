
# docker get latest image
docker pull us-docker.pkg.dev/pingcap-testing-account/hub/tikv/pd/image:master-next-gen_linux_amd64
# docker pull us-docker.pkg.dev/pingcap-testing-account/hub/tikv/tikv/image:release-nextgen-20250815-67eb240-next-gen_linux_amd64
# docker pull us-docker.pkg.dev/pingcap-testing-account/hub/pingcap/tidb/images/tidb-server:master-327a22d-next-gen_linux_amd64

# take out tidb/tikv/pd binary from the docker image
docker run --rm -v $(pwd)/bin:/pd-bin --entrypoint /bin/sh us-docker.pkg.dev/pingcap-testing-account/hub/tikv/pd/image:master-next-gen_linux_amd64 -c 'cp /pd-server /pd-bin/pd-server'

# start pd server
 ./bin/pd-server --name=pd-0 --config=./conf/pd.toml --data-dir=./data --peer-urls=http://0.0.0.0:2380 --client-urls=http://0.0.0.0:2379 --log-file=./data/pd.log

# wait pd server online?

# run the test
go run main.go
