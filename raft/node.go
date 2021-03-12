package raft

import (
	"context"
	pb "go.etcd.io/etcd/raft/raftpb"
	"errors"
)

type SnapshotStatus int

const (
	SnapshotFinish  SnapshotStatus = 1
	SnapshotFailure SnapshotStatus = 2
)

var (
	emptyState = pb.HardState{}

	// ErrStopped is returned by methods on Nodes that have been stopped.
	ErrStopped = errors.New("raft: stopped")
)

type Peer struct {
	ID      uint64
	Context []byte
}

// Node represents a node in a raft cluster.
type Node interface {

	// Propose proposes that data be appended to the log. Note that proposals can be lost without
	// notice, therefore it is user's job to ensure proposal retries.
	Propose(ctx context.Context, data []byte) error

	// Tick increments the internal logical clock for the Node by a single tick. Election
	// timeouts and heartbeat timeouts are in units of ticks.
	Tick()

	// Ready returns a channel that returns the current point-in-time state.
	// Users of the Node must call Advance after retrieving the state returned by Ready.
	//
	// NOTE: No committed entries from the next Ready may be applied until all committed entries
	// and snapshots from the previous one have finished.
	Ready() <-chan Ready

	// Advance notifies the Node that the application has saved progress up to the last Ready.
	// It prepares the node to return the next available Ready.
	//
	// The application should generally call Advance after it applies the entries in last Ready.
	//
	// However, as an optimization, the application may call Advance while it is applying the
	// commands. For example. when the last Ready contains a snapshot, the application might take
	// a long time to apply the snapshot data. To continue receiving Ready without blocking raft
	// progress, it can call Advance before finishing applying the last ready.
	Advance()
}

// IsEmptySnap returns true if the given Snapshot is empty.
func IsEmptySnap(sp pb.Snapshot) bool {
	return sp.Metadata.Index == 0
}


func isHardStateEqual(a, b pb.HardState) bool {
	return a.Term == b.Term && a.Vote == b.Vote && a.Commit == b.Commit
}

// IsEmptyHardState returns true if the given HardState is empty.
func IsEmptyHardState(st pb.HardState) bool {
	return isHardStateEqual(st, emptyState)
}

// StartNode returns a new Node given configuration and a list of raft peers.

func StartNode(c *Config, peers []Peer) Node {
	if len(peers) == 0 {
		panic("no peers given; use RestartNode instead")
	}
	rn, err := NewRawNode(c)
	if err != nil {
		panic(err)
	}
	rn.Bootstrap(peers)
	// TODO bootstrap
	n := newNode(rn)

	go n.run()
	return &n
}

// node is the canonical implementation of the Node interface
type node struct {
	propc      chan msgWithResult
	//recvc      chan pb.Message
	//confc      chan pb.ConfChangeV2
	//confstatec chan pb.ConfState
	readyc     chan Ready
	advancec   chan struct{}
	tickc      chan struct{}
	done       chan struct{}
	stop       chan struct{}
	//status     chan chan Status
	//
	rn *RawNode
}

func newNode(rn *RawNode) node {
	return node{
		propc:      make(chan msgWithResult),
		//recvc:      make(chan pb.Message),
		//confc:      make(chan pb.ConfChangeV2),
		//confstatec: make(chan pb.ConfState),
		readyc:     make(chan Ready),
		advancec:   make(chan struct{}),
		// make tickc a buffered chan, so raft node can buffer some ticks when the node
		// is busy processing raft messages. Raft node will resume process buffered
		// ticks when it becomes idle.
		tickc:  make(chan struct{}, 128),
		done:   make(chan struct{}),
		stop:   make(chan struct{}),
		//status: make(chan chan Status),
		rn:     rn,
	}
}

// MustSync returns true if the hard state and count of Raft entries indicate
// that a synchronous write to persistent storage is required.
func MustSync(st, prevst pb.HardState, entsnum int) bool {
	// Persistent state on all servers:
	// (Updated on stable storage before responding to RPCs)
	// currentTerm
	// votedFor
	// log entries[]
	return entsnum != 0 || st.Vote != prevst.Vote || st.Term != prevst.Term
}

type msgWithResult struct {
	m      pb.Message
	result chan error
}

func (n *node) run() {
	var propc chan msgWithResult
	var readyc chan Ready
	var advancec chan struct{}
	var rd Ready
	r := n.rn.raft

	lead := None

	for {
		if advancec != nil {
			readyc = nil
		} else if n.rn.HasReady() {
			rd = n.rn.readyWithoutAccept()
			readyc = n.readyc
		}

		if lead != r.lead {
			if r.hasLeader() {
				if lead == None {
					r.logger.Infof("raft.node: %x elected leader %x at term %d", r.id, r.lead, r.Term)
				} else {
					r.logger.Infof("raft.node: %x changed leader from %x to %x at term %d", r.id, lead, r.lead, r.Term)
				}
				propc = n.propc
			} else {
				r.logger.Infof("raft.node: %x lost leader %x at term %d", r.id, lead, r.Term)
				propc = nil
			}
			lead = r.lead
		}

		select {
		case pm := <- propc:
			m := pm.m
			m.From = r.id
			err := r.Step(m)
			if pm.result != nil {
				pm.result <- err
				close(pm.result)
			}
		case <-n.tickc:
			n.rn.Tick()
		case readyc <- rd:
			n.rn.acceptReady(rd)
			advancec = n.advancec
		case <-advancec:
			n.rn.Advance(rd)
			rd = Ready{}
			advancec = nil
		}
	}
}

func (n *node) Ready() <-chan Ready {
	return n.readyc
}

func (n *node) Advance() {
	select {
	case n.advancec <- struct {}{}:
	case <- n.done:
	}
}

//func newReady(r *raft, prevSoftSt *SoftState, prevHardSt pb.HardState) Ready {
func newReady(r *raft) Ready {
	rd := Ready{
		Entries:          r.raftLog.unstableEntries(),
		CommittedEntries: r.raftLog.nextEnts(),
		Messages:         r.msgs,
	}
	//if softSt := r.softState(); !softSt.equal(prevSoftSt) {
	//	rd.SoftState = softSt
	//}
	//if hardSt := r.hardState(); !isHardStateEqual(hardSt, prevHardSt) {
	//	rd.HardState = hardSt
	//}
	//if r.raftLog.unstable.snapshot != nil {
	//	rd.Snapshot = *r.raftLog.unstable.snapshot
	//}
	//if len(r.readStates) != 0 {
	//	rd.ReadStates = r.readStates
	//}
	//rd.MustSync = MustSync(r.hardState(), prevHardSt, len(rd.Entries))
	return rd
}


// Ready encapsulates the entries and messages that are ready to read,
// be saved to stable storage, committed or sent to other peers.
// All fields in Ready are read-only.
type Ready struct {
	// CommittedEntries specifies entries to be committed to a
	// store/state-machine. These have previously been committed to stable
	// store.
	CommittedEntries []pb.Entry

	// Messages specifies outbound messages to be sent AFTER Entries are
	// committed to stable storage.
	// If it contains a MsgSnap message, the application MUST report back to raft
	// when the snapshot has been received or has failed by calling ReportSnapshot.
	Messages []pb.Message

	// Entries specifies entries to be saved to stable storage BEFORE
	// Messages are sent.
	Entries []pb.Entry
}


func (rd Ready) appliedCursor() uint64 {
	if n := len(rd.CommittedEntries); n > 0 {
		return rd.CommittedEntries[n - 1].Index
	}
	// TODO snapshot
	return 0
}

func (n *node) Tick() {
	select {
	case n.tickc <- struct {}{}:
	case <- n.done:
	default:
		n.rn.raft.logger.Warningf("%x (leader %v) A tick missed to fire. Node blocks too long!", n.rn.raft.id, n.rn.raft.id == n.rn.raft.lead)
	}
}

func (n *node) Propose(ctx context.Context, data []byte) error {
	return n.stepWait(ctx, pb.Message{Type: pb.MsgProp, Entries: []pb.Entry{{Data: data}}})
}

func (n *node) stepWait(ctx context.Context, m pb.Message) error {
	return n.stepWithWaitOption(ctx, m, true)
}

// Step advances the state machine using msgs. The ctx.Err() will be returned,
// if any.
func (n *node) stepWithWaitOption(ctx context.Context, m pb.Message, wait bool) error {
	if m.Type != pb.MsgProp {
		// TODO
	}
	ch := n.propc
	pm := msgWithResult{m: m}
	if wait {
		pm.result = make(chan error, 1)
	}
	select {
	case ch <- pm:
		if !wait {
			return nil
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-n.done:
		return ErrStopped
	}
	select {
	case err := <-pm.result:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-n.done:
		return ErrStopped
	}
	return nil
}






































