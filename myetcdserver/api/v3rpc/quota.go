package v3rpc

import (
	pb "github.com/nicktming/go/etcdserver/etcdserverpb"
)

type quotaKVServer struct {
	pb.KVServer
}

