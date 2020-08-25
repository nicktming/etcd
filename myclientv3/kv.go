package clientv3

import (
	pb "go.etcd.io/etcd/myetcdserver/etcdserverpb"
	"google.golang.org/grpc"

	"context"
)

type (
	CompactResponse pb.CompactionResponse
	PutResponse     pb.PutResponse
	GetResponse     pb.RangeResponse
	DeleteResponse  pb.DeleteRangeResponse
	TxnResponse     pb.TxnResponse
)

type KV interface {
	Put(ctx context.Context, key, val string) (*PutResponse, error)


	Do(ctx context.Context, op Op) (OpResponse, error)
}

type OpResponse struct {
	put 		*PutResponse
	get 		*GetResponse
	del 		*DeleteResponse
	txn 		*TxnResponse
}

func (op OpResponse) Put() *PutResponse {return op.put}
func (op OpResponse) Get() *GetResponse {return op.get}
func (op OpResponse) Del() *DeleteResponse {return op.del}
func (op OpResponse) Txn() *TxnResponse {return op.txn}

func (resp *PutResponse) OpResponse() OpResponse {
	return OpResponse{put: resp}
}
func (resp *GetResponse) OpResponse() OpResponse {
	return OpResponse{get: resp}
}
func (resp *DeleteResponse) OpResponse() OpResponse {
	return OpResponse{del: resp}
}
func (resp *TxnResponse) OpResponse() OpResponse {
	return OpResponse{txn: resp}
}

type kv struct {
	remote 		pb.KVClient
	callOpts 	[]grpc.CallOption
}

func (kv *kv) Put(ctx context.Context, key, val string) (*PutResponse, error) {
	r, err := kv.Do(ctx, OpPut(key, val))
	return r.put, err 
}

func NewKVFromKVClient(remote pb.KVClient, c *Client) KV {
	api := &kv{remote: remote}
	//if c != nil {
	//
	//}
	return api
}

func (kv *kv) Do(ctx context.Context, op Op) (OpResponse, error) {
	var err error
	switch op.t {
	case tPut:
		var resp *pb.PutResponse
		r := &pb.PutRequest{Key: op.key, Value: op.val, PrevKv: op.prevKV, IgnoreValue: op.ignoreValue, IgnoreLease: op.ignoreLease}
		resp, err = kv.remote.Put(ctx, r, kv.callOpts...)
		if err == nil {
			return OpResponse{put: (*PutResponse)(resp)}, nil
		}
	default:
		panic("Unknown op")
	}
	return OpResponse{}, err
}















































