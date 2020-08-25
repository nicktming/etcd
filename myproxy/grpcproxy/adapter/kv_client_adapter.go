package adapter

import (
	pb "go.etcd.io/etcd/myetcdserver/etcdserverpb"
	grpc "google.golang.org/grpc"
	"context"
)

type kvs2kvc struct {
	kvs pb.KVServer
}

func KvServerToKvClient(kvs pb.KVServer) pb.KVClient {
	return &kvs2kvc{kvs}
}

func (s *kvs2kvc) Range(ctx context.Context, in *pb.RangeRequest, opts ...grpc.CallOption) (*pb.RangeResponse, error) {
	return s.kvs.Range(ctx, in)
}

func (s *kvs2kvc) Put(ctx context.Context, in *pb.PutRequest, opts ...grpc.CallOption) (*pb.PutResponse, error) {
	return s.kvs.Put(ctx, in)
}

func (s *kvs2kvc) DeleteRange(ctx context.Context, in *pb.DeleteRangeRequest, opts ...grpc.CallOption) (*pb.DeleteRangeResponse, error) {
	return s.kvs.DeleteRange(ctx, in)
}

func (s *kvs2kvc) Txn(ctx context.Context, in *pb.TxnRequest, opts ...grpc.CallOption) (*pb.TxnResponse, error) {
	return s.kvs.Txn(ctx, in)
}

func (s *kvs2kvc) Compact(ctx context.Context, in *pb.CompactionRequest, opts ...grpc.CallOption) (*pb.CompactionResponse, error) {
	return s.kvs.Compact(ctx, in)
}
