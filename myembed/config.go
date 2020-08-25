package myembed

import (
	"net/url"
	"log"
	"go.etcd.io/etcd/pkg/types"
)


const (
	DefaultName		= "default"
	ClusterStateFlagNew 	= "new"

	DefaultListenPeerURLs 	= "http://localhost:2380"
)

var (
	DefaultInitialAdvertisePeerURLs = "http://localhost:2380"
)

type Config struct {
	Name 			string 		`json:"name"`
	Dir 			string 		`json:"data-dir"`
	WalDir 			string 		`json:"wal-dir"`

	InitialCluster		string 		`json:"initial-cluster"`
	APUrls			[]url.URL
	LPUrls			[]url.URL
	InitialClusterToken   	string 		`json:"initial-cluster-token"`
	ClusterState          	string 		`json:"initial-cluster-state"`


	// TickMs is the number of milliseconds between heartbeat ticks.
	// TODO: decouple tickMs and heartbeat tick (current heartbeat tick = 1).
	// make ticks a cluster wide configuration.
	TickMs     uint `json:"heartbeat-interval"`
	ElectionMs uint `json:"election-timeout"`

}

func NewConfig() *Config {
	lpurl, _ := url.Parse(DefaultListenPeerURLs)
	apurl, _ := url.Parse(DefaultInitialAdvertisePeerURLs)

	cfg := &Config{
		Name:			DefaultName,

		APUrls: 		[]url.URL{*apurl},
		LPUrls: 		[]url.URL{*lpurl},
		ClusterState: 		ClusterStateFlagNew,
		InitialClusterToken: 	"etcd-cluster",
		TickMs:                     100,
		ElectionMs:                 1000,

	}

	cfg.InitialCluster = cfg.InitialClusterFromName(cfg.Name)
	log.Printf("got cfg.InitialCluster: %v\n", cfg.InitialCluster)
	return cfg
}

// PeerURLsMapAndToken sets up an initial peer URLsMap and cluster token for bootstrap or discovery.
func (cfg *Config) PeerURLsMapAndToken(which string) (urlsmap types.URLsMap, token string, err error) {
	token = cfg.InitialClusterToken
	switch {
	//case cfg.Durl != "":
	//	urlsmap = types.URLsMap{}
	//	// If using discovery, generate a temporary cluster based on
	//	// self's advertised peer URLs
	//	urlsmap[cfg.Name] = cfg.APUrls
	//	token = cfg.Durl
	//
	//case cfg.DNSCluster != "":
	//	clusterStrs, cerr := cfg.GetDNSClusterNames()
	//	lg := cfg.logger
	//	if cerr != nil {
	//		if lg != nil {
	//			lg.Warn("failed to resolve during SRV discovery", zap.Error(cerr))
	//		} else {
	//			plog.Errorf("couldn't resolve during SRV discovery (%v)", cerr)
	//		}
	//		return nil, "", cerr
	//	}
	//	for _, s := range clusterStrs {
	//		if lg != nil {
	//			lg.Info("got bootstrap from DNS for etcd-server", zap.String("node", s))
	//		} else {
	//			plog.Noticef("got bootstrap from DNS for etcd-server at %s", s)
	//		}
	//	}
	//	clusterStr := strings.Join(clusterStrs, ",")
	//	if strings.Contains(clusterStr, "https://") && cfg.PeerTLSInfo.TrustedCAFile == "" {
	//		cfg.PeerTLSInfo.ServerName = cfg.DNSCluster
	//	}
	//	urlsmap, err = types.NewURLsMap(clusterStr)
	//	// only etcd member must belong to the discovered cluster.
	//	// proxy does not need to belong to the discovered cluster.
	//	if which == "etcd" {
	//		if _, ok := urlsmap[cfg.Name]; !ok {
	//			return nil, "", fmt.Errorf("cannot find local etcd member %q in SRV records", cfg.Name)
	//		}
	//	}

	default:
		// We're statically configured, and cluster has appropriately been set.
		urlsmap, err = types.NewURLsMap(cfg.InitialCluster)
	}
	return urlsmap, token, err
}

func (cfg Config) ElectionTicks() int { return int(cfg.ElectionMs / cfg.TickMs) }

func (cfg Config) IsNewCluster() bool { return cfg.ClusterState == ClusterStateFlagNew }

func (cfg Config) InitialClusterFromName(name string) (ret string) {
	if len(cfg.APUrls) == 0 {
		return ""
	}
	n := name
	if n == "" {
		n = DefaultName
	}
	for i := range cfg.APUrls {
		ret = ret + "," + n + "=" + cfg.APUrls[i].String()
	}
	return ret[1:]
}
