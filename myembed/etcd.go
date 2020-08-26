package myembed

import (
	"net"
	"context"
	"time"
	"go.etcd.io/etcd/myetcdserver/api/rafthttp"
	"go.etcd.io/etcd/myetcdserver"
	"fmt"
	"log"
	"go.etcd.io/etcd/pkg/types"
)

// etcd contains a running etcd server and its listeners
type Etcd struct {
	Peers 		[]*peerListener


	cfg 		Config
	stopc 		chan struct{}
	errc 		chan error

	Server 		*myetcdserver.EtcdServer
}

type peerListener struct {
	net.Listener
	serve func() error
	close func(context.Context) error
}

func StartEtcd(inCfg *Config) (e *Etcd, err error) {
	//serving := false
	e = &Etcd{cfg: *inCfg, stopc: make(chan struct{})}
	cfg := &e.cfg

	// TODO
	defer func() {

	}()

	if e.Peers, err = configurePeerListeners(cfg); err != nil {
		return e, err
	}
	// TODO configure client listeners

	var (
		urlsmap 	types.URLsMap
		token 		string
	)

	memberInitialized := true
	if !isMemberInitialized(cfg) {
		memberInitialized = false
		urlsmap, token, err = cfg.PeerURLsMapAndToken("etcd")
		if err != nil {
			return e, fmt.Errorf("error setting up initial cluster: %v", err)
		}
	}
	log.Printf("memberInitialized: %v\n", memberInitialized)

	srvcfg := myetcdserver.ServerConfig{
		Name: 			cfg.Name,
		PeerURLs: 		cfg.APUrls,
		DataDir: 		cfg.Dir,
		InitialPeerURLsMap: 	urlsmap,
		NewCluster: 		cfg.IsNewCluster(),
		InitialClusterToken: 	token,
		ElectionTicks:          cfg.ElectionTicks(),
		TickMs:                 cfg.TickMs,
	}

	e.Server, err = myetcdserver.NewServer(srvcfg)
	if err != nil {
		log.Printf("NewServer got err: %v\n", err)
		return e, err
	}

	log.Printf("e.Server: %v\n", e.Server)

	// TODO CheckInitialHashKV

	go e.Server.Start()

	return e, nil
}


func configurePeerListeners(cfg *Config) (peers []*peerListener, err error) {
	peers = make([]*peerListener, len(cfg.LPUrls))

	// TODO defer function
	defer func() {
		if err == nil {
			return
		}
		for i := range peers {
			if peers[i] != nil && peers[i].close != nil {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				peers[i].close(ctx)
				cancel()
			}
		}
	}()

	for i, u := range cfg.LPUrls {
		peers[i] = &peerListener{}
		peers[i].Listener, err = rafthttp.NewListener(u, nil)
		if err != nil {
			return nil, err
		}
		peers[i].close = func(context.Context) error {
			return peers[i].Listener.Close()
		}
	}
	return peers, nil
}



































