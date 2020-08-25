package myetcdserver

import (
	"go.etcd.io/etcd/pkg/types"
	"go.etcd.io/etcd/pkg/pbutil"
	pb "go.etcd.io/etcd/myetcdserver/etcdserverpb"
	"log"
	"encoding/json"
	"go.uber.org/zap"
	"sync"
	"time"
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/wal"
	"go.etcd.io/etcd/myetcdserver/api/membership"
	"go.etcd.io/etcd/raft/raftpb"
	"go.etcd.io/etcd/myetcdserver/api/rafthttp"
)


func startNode(cfg ServerConfig, cl *membership.RaftCluster, ids []types.ID) (id types.ID, n raft.Node, s *raft.MemoryStorage, w *wal.WAL) {
	var err error
	member := cl.MemberByName(cfg.Name)
	metadata := pbutil.MustMarshal(
		&pb.Metadata{
			NodeID: 	uint64(member.ID),
			ClusterID: 	uint64(cl.ID()),
		},
	)
	if w, err = wal.Create(nil, cfg.WALDir(), metadata); err != nil {
		log.Panicf("create wal error: %v", err)
	}
	peers := make([]raft.Peer, len(ids))
	for i, id := range ids {
		var ctx []byte
		ctx, err = json.Marshal((*cl).Member(id))
		if err != nil {
			log.Panicf("marshal member should never fail: %v", err)
		}
		peers[i] = raft.Peer{ID: uint64(id), Context: ctx}
	}
	id = member.ID
	log.Printf("starting member %s in cluster %s", id, cl.ID())

	s = raft.NewMemoryStorage()
	c := &raft.Config{
		ID: 			uint64(id),
		ElectionTick: 		cfg.ElectionTicks,
		HeartbeatTick: 		1,
		Storage: 		s,
		MaxInflightMsgs: 	10,
		// TODO other flag
	}
	if len(peers) == 0 {
		// TODO restart
	} else {
		n = raft.StartNode(c, peers)
	}

	// TODO status
	return id, n, s, w
}

type raftNode struct {
	lg 		*zap.Logger
	tickMu		*sync.Mutex
	raftNodeConfig

	ticker 		*time.Ticker

	applyc 		chan apply
}

type raftNodeConfig struct {
	lg 		*zap.Logger

	isIDRemoved 	func(id uint64) bool
	raft.Node

	raftStorage 	*raft.MemoryStorage
	heartbeat 	time.Duration

	transport 	rafthttp.Transporter
}

type apply struct {
	entries 	[]raftpb.Entry
	snapshot 	raftpb.Snapshot
	notifyc 	chan struct{}
}

func newRaftNode(cfg raftNodeConfig) *raftNode {
	r := &raftNode{
		tickMu:			new(sync.Mutex),
		raftNodeConfig:		cfg,
		applyc:			make(chan apply),
	}
	if r.heartbeat == 0 {
		r.ticker = &time.Ticker{}
	} else {
		r.ticker = time.NewTicker(r.heartbeat)
	}
	return r
}

func (r *raftNode) tick() {
	r.tickMu.Lock()
	r.Tick()
	r.tickMu.Unlock()
}

// TODO raftReadyHandler
func (r *raftNode) start() {
	go func() {
		for {
			select {
			case <- r.ticker.C:
				r.tick()
			case rd := <-r.Ready():
			// TODO softstate
			// TODO ReadState

				log.Printf("From rd: %v\n", rd)

				notifyc := make(chan struct{}, 1)
				ap := apply {
					entries: 	rd.CommittedEntries,
					// TODO snapshot
					notifyc: 	notifyc,
				}
			// TODO update commited entries

					select {
					case r.applyc <- ap:
					}
			// TODO
				r.Advance()
			}
		}
	}()
}

func (r *raftNode) apply() chan apply {
	return r.applyc
}






































