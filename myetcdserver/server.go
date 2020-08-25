package myetcdserver


import (
	"go.etcd.io/etcd/pkg/types"
	"go.etcd.io/etcd/pkg/fileutil"
	"fmt"
	//"github.com/nicktming/go/etcdserver/api/snap"

	"time"
	"context"
	"log"
	"go.etcd.io/etcd/pkg/wait"
	"go.etcd.io/etcd/myetcdserver/api"
	"go.etcd.io/etcd/mvcc/backend"
	"go.etcd.io/etcd/raft/raftpb"
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/wal"
	"go.etcd.io/etcd/myetcdserver/api/membership"
	"go.etcd.io/etcd/myetcdserver/api/rafthttp"
)

type EtcdServer struct {
	id 		types.ID
	Cfg 		ServerConfig
	cluster 	*membership.RaftCluster
	r 		raftNode
	be         	backend.Backend

	w 		wait.Wait
}

type etcdProgress struct {
	confState 	raftpb.ConfState
	snapi 		uint64
	appliedt 	uint64
	appliedi 	uint64
}

func (s *EtcdServer) Process(ctx context.Context, m raftpb.Message) error {
	// TODO isIDRemoved
	return s.r.Step(ctx, m)
}

func (s *EtcdServer) Cluster() api.Cluster { return s.cluster }

func (s *EtcdServer) ID() types.ID { return s.id }

func (s *EtcdServer) Start() {
	s.start()
}

func (s *EtcdServer) start() {
	s.w = wait.New()
	s.run()
}

func (s *EtcdServer) run() {
	// TODO snapshot

	log.Printf("s.r: %v\n", s.r)
	sn, err := s.r.raftStorage.Snapshot()
	if err != nil {
		panic("failed to get snapshot")
	}

	s.r.start()

	ep := etcdProgress{
		confState: 	sn.Metadata.ConfState,
		snapi:		sn.Metadata.Index,
		appliedt:	sn.Metadata.Term,
		appliedi: 	sn.Metadata.Index,
	}


	for {
		select {
		case ap := <-s.r.apply():
			s.applyAll(&ep, &ap)
		}
	}
}

func (s *EtcdServer) apply(
es []raftpb.Entry, confState *raftpb.ConfState) (appliedt uint64, appliedi uint64, shouldStop bool) {

	for i := range es {
		e := es[i]
		switch e.Type {
		case raftpb.EntryNormal:
			log.Println("got type: raftpb.EntryNormal")
		case raftpb.EntryConfChange:
			log.Println("got type: raftpb.EntryConfChange")
		default:
			panic("entry type should either EntryNormal or EntryConfChange")
		}
		appliedi, appliedt = e.Index, e.Term
	}

	return appliedt, appliedi, shouldStop
}

func (s *EtcdServer) applyEntries(ep *etcdProgress, apply *apply) {

	if len(apply.entries) == 0 {
		return
	}

	firsti := apply.entries[0].Index
	if firsti > ep.appliedi + 1 {
		log.Panicf("firsti:%v should be < ep.applied(%v) + 1\n", firsti, ep.appliedi)
	}

	var ents []raftpb.Entry
	if ep.appliedi + 1 - firsti < uint64(len(apply.entries)) {
		ents = apply.entries[ep.appliedi + 1 - firsti : ]
	}

	if len(ents) == 0 {
		return
	}

	// TODO apply to blotdb

	var shouldstop bool
	if ep.appliedt, ep.appliedi, shouldstop = s.apply(ents, &ep.confState); shouldstop {
		log.Println("should stop this node")
	}

}


func (s *EtcdServer) applyAll(ep *etcdProgress, apply *apply) {

	s.applyEntries(ep, apply)

	<- apply.notifyc
}

func (s *EtcdServer) IsIDRemoved(id uint64) bool {
	return false
}
func (s *EtcdServer) ReportUnreachable(id uint64) {}
func (s *EtcdServer) ReportSnapshot(id uint64, status raft.SnapshotStatus) {}

func NewServer(cfg ServerConfig) (srv *EtcdServer, err error) {
	var (
		//w 	*wal.WAL
		id 	types.ID
		n 	raft.Node
		s 	*raft.MemoryStorage
		cl 	*membership.RaftCluster
	)

	if terr := fileutil.TouchDirAll(cfg.DataDir); terr != nil {
		return nil, fmt.Errorf("cannot access data directory: %v", terr)
	}

	haveWal := wal.Exist(cfg.WALDir())

	if err = fileutil.TouchDirAll(cfg.SnapDir()); err != nil {
		return nil, fmt.Errorf("cannot create snapshot directory error: %v", err)
	}

	//ss := snap.New(cfg.SnapDir())

	//bepath := cfg.backendPath()
	//beExist := fileutil.Exist(bepath)
	be := newBackend(cfg)

	defer func() {
		if err != nil {
			be.Close()
		}
	}()

	if !haveWal && cfg.NewCluster {
		cl, err = membership.NewClusterFromURLsMap(nil, cfg.InitialClusterToken, cfg.InitialPeerURLsMap)
		if err != nil {
			log.Printf("return nil etcdserver and err: %v\n", err)
			return nil, err
		}

		//m := cl.MemberByName(cfg.Name)

		// TODO isMemberBootstrapped and should discover

		cl.SetBackend(be)
		id, n, s, _ = startNode(cfg, cl, cl.MemberIDs())
		cl.SetID(id, cl.ID())

	} else {
		panic("this should not happen")
	}



	heartbeat := time.Duration(cfg.TickMs) * time.Millisecond
	srv = &EtcdServer{
		id: 			id,
		Cfg:			cfg,
		r: 			*newRaftNode(raftNodeConfig{
			Node: 		n,
			heartbeat: 	heartbeat,
			raftStorage: 	s,
			// TODO new storage
		}),
		cluster:		cl,
	}

	srv.be = be
	// TODO lease

	// TODO kv must be done

	// TODO defer close function



	tr := &rafthttp.Transport{
		ID:		id,
		URLs:		cfg.PeerURLs,
		Raft:		srv,
		DialTimeout: 	cfg.peerDialTimeout(),
	}

	if err = tr.Start(); err != nil {
		return nil, err
	}

	for _, m := range cl.Members() {
		if m.ID != id {
			tr.AddPeer(m.ID, m.PeerURLs)
		}
	}

	srv.r.transport = tr


	return srv, nil
}