package raft

import pb "go.etcd.io/etcd/raft/raftpb"

func limitSize(ents []pb.Entry, maxSize uint64) []pb.Entry {
	if len(ents) == 0 {
		return ents
	}
	size := ents[0].Size()
	var limit int
	for limit = 1; limit < len(ents); limit++ {
		size += ents[limit].Size()
		if uint64(size) > maxSize {
			break
		}
	}
	return ents[:limit]
}


func min(a, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

func assertConfStatesEquivalent(l Logger, cs1, cs2 pb.ConfState) {
	err := cs1.Equivalent(cs2)
	if err == nil {
		return
	}
	l.Panic(err)
}


