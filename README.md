# store

./etcd --advertise-client-urls=http://0.0.0.0:2379 --listen-client-urls=http://0.0.0.0:2379

./etcdctl --endpoints=localhost:2379 watch /stream/system/subscribe --prefix --rev=1

./etcdctl --endpoints=localhost:2379 get / --prefix
