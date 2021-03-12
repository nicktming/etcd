package raft

import "fmt"

//import "gopkg.in/cheggaaa/pb.v1"

//import

// RawNode is a thread-unsafe Node.
// The methods of this struct correspond to the methods of Node and are described
// more fully there.
type RawNode struct {
	raft       *raft
	//prevSoftSt *SoftState
	//prevHardSt pb.HardState
}

// NewRawNode instantiates a RawNode from the given configuration.
//
// See Bootstrap() for bootstrapping an initial state; this replaces the former
// 'peers' argument to this method (with identical behavior). However, It is
// recommended that instead of calling Bootstrap, applications bootstrap their
// state manually by setting up a Storage that has a first index > 1 and which
// stores the desired ConfState as its InitialState.
func NewRawNode(config *Config) (*RawNode, error) {
	r := newRaft(config)
	rn := &RawNode{
		raft: r,
	}
	//rn.prevSoftSt = r.softState()
	//rn.prevHardSt = r.hardState()
	return rn, nil
}

// HasReady called when RawNode user need to check if any Ready pending.
// Checking logic in this method should be consistent with Ready.containsUpdates().
func (rn *RawNode) HasReady() bool {
	r := rn.raft

	ready := len(r.msgs) > 0 || len(r.raftLog.unstableEntries()) > 0 || r.raftLog.hasNextEnts()

	//fmt.Printf("HasReady: %v\n", ready)

	return ready
}

// readyWithoutAccept returns a Ready. This is a read-only operation, i.e. there
// is no obligation that the Ready must be handled.
func (rn *RawNode) readyWithoutAccept() Ready {
	//return newReady(rn.raft, rn.prevSoftSt, rn.prevHardSt)
	return newReady(rn.raft)
}


// acceptReady is called when the consumer of the RawNode has decided to go
// ahead and handle a Ready. Nothing must alter the state of the RawNode between
// this call and the prior call to Ready().
func (rn *RawNode) acceptReady(rd Ready) {
	//if rd.SoftState != nil {
	//	rn.prevSoftSt = rd.SoftState
	//}
	//if len(rd.ReadStates) != 0 {
	//	rn.raft.readStates = nil
	//}
	rn.raft.msgs = nil
}

// Advance notifies the RawNode that the application has applied and saved progress in the
// last Ready results.
func (rn *RawNode) Advance(rd Ready) {
	//if !IsEmptyHardState(rd.HardState) {
	//	rn.prevHardSt = rd.HardState
	//}
	rn.raft.advance(rd)

	fmt.Println("advance raft done.")
}

func (rn *RawNode) Tick() {
	rn.raft.tick()
}























































