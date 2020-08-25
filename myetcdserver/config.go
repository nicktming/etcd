package myetcdserver

import (
	"go.etcd.io/etcd/pkg/types"
	"path/filepath"
	"time"
)

type ServerConfig struct {
	Name 			string
	ClientURLs		types.URLs
	PeerURLs		types.URLs
	DataDir 		string

	InitialPeerURLsMap	types.URLsMap

	NewCluster 		bool
	InitialClusterToken 	string
	ElectionTicks 		int


	TickMs        uint
}


func (c *ServerConfig) MemberDir() string {
	return filepath.Join(c.DataDir, "member")
}

func (c *ServerConfig) WALDir() string {
	return filepath.Join(c.MemberDir(), "wal")
}

func (c *ServerConfig) SnapDir() string {
	return filepath.Join(c.MemberDir(), "snap")
}

func (c *ServerConfig) backendPath() string {
	return filepath.Join(c.SnapDir(), "db")
}

func (c *ServerConfig) peerDialTimeout() time.Duration {
	// 1s for queue wait and election timeout
	return time.Second + time.Duration(c.ElectionTicks*int(c.TickMs))*time.Millisecond
}
