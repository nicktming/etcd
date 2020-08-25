package membership

import (
	"go.etcd.io/etcd/pkg/types"
	"go.uber.org/zap"
	"fmt"
	"sync"
	"encoding/binary"
	"crypto/sha1"
	"sort"
	"log"
	"github.com/coreos/go-semver/semver"
	"go.etcd.io/etcd/mvcc/backend"
	"go.etcd.io/etcd/raft"
)

type RaftCluster struct {
	lg 		*zap.Logger
	localID 	types.ID
	cid 		types.ID

	be 		backend.Backend

	token 		string

	members 	map[types.ID]*Member
	removed 	map[types.ID]bool

	version    	*semver.Version

	sync.Mutex
}

func (c *RaftCluster) MemberIDs() []types.ID {
	c.Lock()
	defer c.Unlock()
	var ids []types.ID
	for _, m := range c.members {
		ids = append(ids, m.ID)
	}
	sort.Sort(types.IDSlice(ids))
	return ids
}

func (c *RaftCluster) Member(id types.ID) *Member {
	c.Lock()
	defer c.Unlock()
	return c.members[id].Clone()
}

func (c *RaftCluster) ID() types.ID { return c.cid }

func (c *RaftCluster) SetID(localID, cid types.ID) {
	c.localID = localID
	c.cid = cid
}

func (c *RaftCluster) Version() *semver.Version {
	c.Lock()
	defer c.Unlock()
	if c.version == nil {
		return nil
	}
	return semver.Must(semver.NewVersion(c.version.String()))
}

func (c *RaftCluster) ClientURLs() []string {
	c.Lock()
	defer c.Unlock()
	urls := make([]string, 0)
	for _, p := range c.members {
		urls = append(urls, p.ClientURLs...)
	}
	sort.Strings(urls)
	return urls
}

func (c *RaftCluster) Members() []*Member {
	c.Lock()
	defer c.Unlock()
	var ms MembersByID
	for _, m := range c.members {
		ms = append(ms, m.Clone())
	}
	sort.Sort(ms)
	return []*Member(ms)
}

// MemberByName returns a Member with the given name if exists.
// If more than one member has the given name, it will panic.
func (c *RaftCluster) MemberByName(name string) *Member {
	c.Lock()
	defer c.Unlock()
	var memb *Member
	for _, m := range c.members {
		if m.Name == name {
			if memb != nil {
				if c.lg != nil {
					c.lg.Panic("two member with same name found", zap.String("name", name))
				} else {
					log.Panicf("two members with the given name %q exist", name)
				}
			}
			memb = m
		}
	}
	return memb.Clone()
}


func (m *Member) Clone() *Member {
	if m == nil {
		return nil
	}
	mm := &Member{
		ID: m.ID,
		RaftAttributes: RaftAttributes{
			IsLearner: m.IsLearner,
		},
		Attributes: Attributes{
			Name: m.Name,
		},
	}
	if m.PeerURLs != nil {
		mm.PeerURLs = make([]string, len(m.PeerURLs))
		copy(mm.PeerURLs, m.PeerURLs)
	}
	if m.ClientURLs != nil {
		mm.ClientURLs = make([]string, len(m.ClientURLs))
		copy(mm.ClientURLs, m.ClientURLs)
	}
	return mm
}

func (c *RaftCluster) genID() {
	mIDs := c.MemberIDs()
	b := make([]byte, 8*len(mIDs))
	for i, id := range mIDs {
		binary.BigEndian.PutUint64(b[8*i:], uint64(id))
	}
	hash := sha1.Sum(b)
	c.cid = types.ID(binary.BigEndian.Uint64(hash[:8]))
}


func (c *RaftCluster) SetBackend(be backend.Backend) {
	c.be = be
	// TODO mustCreate
}


func NewClusterFromURLsMap(lg *zap.Logger, token string, urlsmap types.URLsMap) (*RaftCluster, error) {
	c := NewCluster(lg, token)
	for name, urls := range urlsmap {
		m := NewMember(name, urls, token, nil)
		if _, ok := c.members[m.ID]; ok {
			return nil, fmt.Errorf("member exists with identical ID %v", m)
		}
		if uint64(m.ID) == raft.None {
			return nil, fmt.Errorf("cannot use %x as member id", raft.None)
		}
		c.members[m.ID] = m
	}
	c.genID()
	return c, nil
}

func NewCluster(lg *zap.Logger, token string) *RaftCluster {
	return &RaftCluster{
		lg:		lg,
		token: 		token,
		members:	make(map[types.ID]*Member),
		removed:	make(map[types.ID]bool),
	}
}