package myetcdserver


import (
	"github.com/gogo/protobuf/proto"
	//"go.etcd.io/etcd/pkg/traceutil"
)

type applyResult struct {
	resp 		proto.Message
	err 		error

	physc 		<-chan struct{}
	//trace 		*traceutil.Trace
}