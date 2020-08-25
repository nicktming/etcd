package membership

import (
	"go.etcd.io/etcd/pkg/types"
	"time"
	"sort"
	"fmt"
	//"crypto/sha1"
	//"encoding/binary"
	"math/rand"
	//"crypto/sha1"
	//"encoding/binary"
)

type RaftAttributes struct {
	PeerURLs		[]string 	`json:"peerURLs"`
	// TODO isLearner
	IsLearner 		bool 		`json:"isLearner,omitempty"`
}

type Attributes struct {
	Name 			string 	`json:"name,omitempty"`
	ClientURLs		[]string 	`json:"clientURLs, omitempty"`
}

type Member struct {
	ID		types.ID	`json:"id"`
	RaftAttributes
	Attributes
}

func NewMember(name string, peerURLs types.URLs, clusterName string, now *time.Time) *Member {
	return newMember(name, peerURLs, clusterName, now, false)
}

func newMember(name string, peerURLs types.URLs, clusterName string, now *time.Time, isLearner bool)  *Member {
	m := &Member{
		RaftAttributes: RaftAttributes{
			PeerURLs: 	peerURLs.StringSlice(),
			IsLearner: 	isLearner,
		},
		Attributes: Attributes{Name: name},
	}

	var b []byte
	sort.Strings(m.PeerURLs)
	for _, p := range m.PeerURLs {
		b = append(b, []byte(p)...)
	}

	b = append(b, []byte(clusterName)...)
	if now != nil {
		b = append(b, []byte(fmt.Sprintf("%d", now.Unix()))...)
	}

	m.ID = types.ID(1)
	//hash := sha1.Sum(b)
	//m.ID = types.ID(binary.BigEndian.Uint64(hash[:8]))
	return m
}


func (m *Member) PickPeerURL() string {
	if len(m.PeerURLs) == 0 {
		panic("member should always have some peer url")
	}
	return m.PeerURLs[rand.Intn(len(m.PeerURLs))]
}


// MembersByID implements sort by ID interface
type MembersByID []*Member

func (ms MembersByID) Len() int           { return len(ms) }
func (ms MembersByID) Less(i, j int) bool { return ms[i].ID < ms[j].ID }
func (ms MembersByID) Swap(i, j int)      { ms[i], ms[j] = ms[j], ms[i] }

