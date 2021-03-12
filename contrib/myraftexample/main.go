package main

import (
	"go.etcd.io/etcd/raft"
	"go.etcd.io/etcd/raft/raftpb"
	"fmt"
	"time"
)

type raftNode struct {
	Node raft.Node
	done chan struct{}
}

func main() {

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()


	storage := raft.NewMemoryStorage()
	c := &raft.Config{
		ID:              0x01,
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}
	n := raft.StartNode(c, []raft.Peer{{ID: 0x01}})

	fmt.Println("======>start node done<======")

	s := &raftNode{
		Node: n,
		done: make(chan struct{}, 1),
	}

	for {
		select {
		case <-ticker.C:
			n.Tick()
		case rd := <-s.Node.Ready():
			//saveToStorage(rd.State, rd.Entries, rd.Snapshot)
			//send(rd.Messages)
			//if !raft.IsEmptySnap(rd.Snapshot) {
			//	processSnapshot(rd.Snapshot)
			//}
			for _, entry := range rd.CommittedEntries {
				process(entry)
				if entry.Type == raftpb.EntryConfChange {
					var cc raftpb.ConfChange
					cc.Unmarshal(entry.Data)
					s.Node.ApplyConfChange(cc)
				}
			}
			s.Node.Advance()
		case <-s.done:
			return
		}
	}

}

func process(e raftpb.Entry) {
	fmt.Printf("process entry e: [Term: %v, Index: %v, Type: %v, Data: %v]\n",
		e.Term, e.Index, e.Type.String(), string(e.Data))
}

func (rc *raftNode) Stop() {
	close(rc.done)
}