package etcd

import "github.com/coreos/etcd/clientv3"

type Datastore struct {
	Cli     *clientv3.Client
	Kv      clientv3.KV
	Lease   clientv3.Lease
	Watcher clientv3.Watcher
}

func NewDatastore(cli *clientv3.Client, kv clientv3.KV, lease clientv3.Lease, watcher clientv3.Watcher) *Datastore {
	return &Datastore{
		Cli:     cli,
		Kv:      kv,
		Lease:   lease,
		Watcher: watcher,
	}
}
