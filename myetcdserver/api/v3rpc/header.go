package v3rpc

import "github.com/nicktming/go/etcdserver"

type header struct {
	clusterID 	int64
	memberID 	int64
	//sg 		etcdserver
	rev 		func() int64
}

func newHeader(s *etcdserver.EtcdServer) header {
	return header {
		clusterID: 	int64(s.Cluster().ID()),
		memberID: 	int64(s.ID()),
		//sg: 		s,
		//rev: 		func() int64 {return }
	}
}
