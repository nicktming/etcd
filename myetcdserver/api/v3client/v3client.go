package v3client

import (
	"context"
	"go.etcd.io/etcd/myproxy/grpcproxy/adapter"
	"go.etcd.io/etcd/myetcdserver"
	clientv3 "go.etcd.io/etcd/myclientv3"
)

func New(s *myetcdserver.EtcdServer) *clientv3.Client {
	c := clientv3.NewCtxClient(context.Background())

	kvc := adapter.KvServerToKvClient(s)

	c.KV = clientv3.NewKVFromKVClient(kvc, c)
	return c
}