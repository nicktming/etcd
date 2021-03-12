package raft

import (
	"errors"
	pb "go.etcd.io/etcd/raft/raftpb"
)

// Bootstrap initializes the RawNode for first use by appending configuration
// changes for the supplied peers. This method returns an error if the Storage
// is nonempty.
//
// It is recommended that instead of calling this method, applications bootstrap
// their state manually by setting up a Storage that has a first index > 1 and
// which stores the desired ConfState as its InitialState.
func (rn *RawNode) Bootstrap(peers []Peer) error {
	if len(peers) == 0 {
		return errors.New("must provide at least one peer to Bootstrap")
	}

	rn.raft.becomeFollower(1, None)
	ents := make([]pb.Entry, len(peers))
	for i, peer := range peers {
		cc := pb.ConfChange{Type: pb.ConfChangeAddNode, NodeID: peer.ID, Context: peer.Context}
		data, err := cc.Marshal()
		if err != nil {
			return err
		}

		ents[i] = pb.Entry{Type: pb.EntryConfChange, Term: 1, Index: uint64(i + 1), Data: data}
	}
	rn.raft.raftLog.append(ents...)

	rn.raft.raftLog.committed = uint64(len(ents))
	for _, peer := range peers {
		rn.raft.applyConfChange(pb.ConfChange{NodeID: peer.ID, Type: pb.ConfChangeAddNode}.AsV2())
	}
	return nil
}
