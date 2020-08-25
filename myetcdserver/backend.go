package myetcdserver

import "go.etcd.io/etcd/mvcc/backend"

func newBackend(cfg ServerConfig) backend.Backend {
	bcfg := backend.DefaultBackendConfig()
	bcfg.Path = cfg.backendPath()

	return backend.New(bcfg)
}
