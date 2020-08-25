package v3rpc

import (
	"github.com/nicktming/go/etcdserver"
	pb "github.com/nicktming/go/etcdserver/etcdserverpb"
)

type kvServer struct {
	// header

	kv 		etcdserver.RaftKV

	maxTxnOps 	uint
}

func NewKVServer(s *etcdserver.EtcdServer) pb.KVServer {
	return &kvServer{kv: s}
}
