# gctest

## About

- longtxn: long run tidb session should block GC
- isolation: GC on each keyspace works independently
- fuzz: GC API fuzz test
- compatibility: compatibility test
- mockservice: mock the external service to call GC API

## How to

Run test on GCP:

```
gcloud deployment-manager deployments create test-mkl-gc --config test-mkl-gc.yaml
```

Destroy instances:

```
gcloud deployment-manager deployments delete test-mkl-gc
```

Check execution result:

```
gcloud compute instances get-serial-port-output test-mkl-gc
```
