package myetcdserver

import (
	"context"
	pb "go.etcd.io/etcd/myetcdserver/etcdserverpb"
	//"go.etcd.io/etcd/pkg/traceutil"
	"time"
	"github.com/gogo/protobuf/proto"
)
type RaftKV interface {
	Range(ctx context.Context, r *pb.RangeRequest) (*pb.RangeResponse, error)
	Put(ctx context.Context, r *pb.PutRequest) (*pb.PutResponse, error)
	DeleteRange(ctx context.Context, r *pb.DeleteRangeRequest) (*pb.DeleteRangeResponse, error)
	Txn(ctx context.Context, r *pb.TxnRequest) (*pb.TxnResponse, error)
	Compact(ctx context.Context, r *pb.CompactionRequest) (*pb.CompactionResponse, error)
}

func (s *EtcdServer) Compact(ctx context.Context, r *pb.CompactionRequest) (*pb.CompactionResponse, error) {
	return nil, nil
}

func (s *EtcdServer) Txn(ctx context.Context, r *pb.TxnRequest) (*pb.TxnResponse, error) {
	return nil, nil
}

func (s *EtcdServer) DeleteRange(ctx context.Context, r *pb.DeleteRangeRequest) (*pb.DeleteRangeResponse, error) {
	return nil, nil
}

func (s *EtcdServer) Range(ctx context.Context, r *pb.RangeRequest) (*pb.RangeResponse, error) {
	return nil, nil
}

func (s *EtcdServer) Put(ctx context.Context, r *pb.PutRequest) (*pb.PutResponse, error) {
	ctx = context.WithValue(ctx, "startTime", time.Now())
	resp, err := s.raftRequest(ctx, pb.InternalRaftRequest{Put: r})
	if err != nil {
		return nil, err
	}
	return resp.(*pb.PutResponse), nil
}

func (s *EtcdServer) raftRequest(ctx context.Context, r pb.InternalRaftRequest) (proto.Message, error) {
	return s.raftRequestOnce(ctx, r)
}

func (s *EtcdServer) raftRequestOnce(ctx context.Context, r pb.InternalRaftRequest) (proto.Message, error) {
	result, err := s.processInternalRaftRequestOnce(ctx, r)
	if err != nil {
		return nil, err
	}
	if result.err != nil {
		return nil, result.err
	}
	//if startTime, ok := ctx.Value(traceutil.StartTimeKey).(time.Time); ok && result.trace != nil {
	//	applyStart := result.trace.GetStartTime()
	//	// The trace object is created in apply. Here reset the start time to trace
	//	// the raft request time by the difference between the request start time
	//	// and apply start time
	//	result.trace.SetStartTime(startTime)
	//	result.trace.InsertStep(0, applyStart, "process raft request")
	//	//result.trace.LogIfLong(traceThreshold)
	//}
	return result.resp, nil
}

func (s *EtcdServer) processInternalRaftRequestOnce(ctx context.Context, r pb.InternalRaftRequest) (*applyResult, error) {
	// TODO check
	//ai := s.getAppliedIndex()
	//ci := s.getCommittedIndex()
	//if ci > ai+maxGapBetweenApplyAndCommitIndex {
	//	return nil, ErrTooManyRequests
	//}

	//r.Header = &pb.RequestHeader{
	//	ID: s.reqIDGen.Next(),
	//}

	//authInfo, err := s.AuthInfoFromCtx(ctx)
	//if err != nil {
	//	return nil, err
	//}
	//if authInfo != nil {
	//	r.Header.Username = authInfo.Username
	//	r.Header.AuthRevision = authInfo.Revision
	//}

	data, err := r.Marshal()
	if err != nil {
		return nil, err
	}

	//if len(data) > int(s.Cfg.MaxRequestBytes) {
	//	return nil, ErrRequestTooLarge
	//}

	id := r.ID
	//if id == 0 {
	//	id = r.Header.ID
	//}
	ch := s.w.Register(id)

	cctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	defer cancel()

	//start := time.Now()
	err = s.r.Propose(cctx, data)
	if err != nil {
		//proposalsFailed.Inc()
		s.w.Trigger(id, nil) // GC wait
		return nil, err
	}
	//proposalsPending.Inc()
	//defer proposalsPending.Dec()

	select {
	case x := <-ch:
		return x.(*applyResult), nil
	case <-cctx.Done():
	//proposalsFailed.Inc()
		s.w.Trigger(id, nil) // GC wait
	//return nil, s.parseProposeCtxErr(cctx.Err(), start)
		return nil, cctx.Err()
	//case <-s.done:
	//	return nil, ErrStopped
	}
}

