package api

import (
	"go.etcd.io/etcd/pkg/types"
	"github.com/coreos/go-semver/semver"
	"go.etcd.io/etcd/myetcdserver/api/membership"
)

type Cluster interface {
	// ID returns the cluster ID
	ID() types.ID
	// ClientURLs returns an aggregate set of all URLs on which this
	// cluster is listening for client requests
	ClientURLs() []string
	// Members returns a slice of members sorted by their ID
	Members() []*membership.Member
	// Member retrieves a particular member based on ID, or nil if the
	// member does not exist in the cluster
	Member(id types.ID) *membership.Member
	// Version is the cluster-wide minimum major.minor version.
	Version() *semver.Version
}

