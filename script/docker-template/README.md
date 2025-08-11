A docker template to start a cluster in one command:

## Usage

Put pd-server, tikv-server, tidb-server into the bin directory.

```
docker run --rm -it \
       -p 4000:4000 \
       -v "$CURR_DIR/script":/root/script \
       -v "$CURR_DIR/bin":/root/bin \
       -v "$CURR_DIR/conf":/root/conf \
       -v "$CURR_DIR/data":/root/data \
       -w /root \
       --entrypoint /root/script/entry.sh \
       rockylinux:9
```

If you want to run test using this script, copy it to your directory, edit the `script/entry.sh` to run test script.

## How it works

Using docker as a isolated environment for testing.

-p expose the 4000 port, not a must for testing
-v maps local docker/{script|bin|conf|data} to the docker filesystem /root/{script|bin|conf|data}
-w set the working directory to /root
--entrypoint specifies what to run, here it is indeed `docker/script/entry.sh`

Those files are mapping to /root/*

```
$ tree
.
├── bin
├── conf
│   ├── pd-0.toml
│   ├── tidb-0.toml
│   └── tikv-0.toml
├── data
├── README.md
└── script
    ├── entry.sh
    └── start.sh
```

`script/start.sh` is the script to start all the services.
`script/entry.sh` is the entrypoint and it calls start.sh to start a cluster.

## Debug

```
docker run --rm -it \
       -p 4000:4000 \
       -v "$CURR_DIR/script":/root/script \
       -v "$CURR_DIR/bin":/root/bin \
       -v "$CURR_DIR/conf":/root/conf \
       -v "$CURR_DIR/data":/root/data \
       -w /root \
       rockylinux:9 bash
	   
# start cluster manually
cd /root/script
./start.sh &
```
