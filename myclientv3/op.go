package clientv3

type opType int

const (
	// A default Op has opType 0, which is invalid.
	tRange opType = iota + 1
	tPut
	tDeleteRange
	tTxn
)

var noPrefixEnd = []byte{0}

// Op represents an Operation that kv can execute.
type Op struct {
	t   opType
	key []byte
	end []byte

	// for range
	limit        int64
	sort         *SortOption
	serializable bool
	keysOnly     bool
	countOnly    bool
	minModRev    int64
	maxModRev    int64
	minCreateRev int64
	maxCreateRev int64

	// for range, watch
	rev int64

	// for watch, put, delete
	prevKV bool

	// for watch
	// fragmentation should be disabled by default
	// if true, split watch events when total exceeds
	// "--max-request-bytes" flag value + 512-byte
	fragment bool

	// for put
	ignoreValue bool
	ignoreLease bool

	// progressNotify is for progress updates.
	progressNotify bool
	// createdNotify is for created event
	createdNotify bool
	// filters for watchers
	filterPut    bool
	filterDelete bool

	// for put
	val     []byte
	//leaseID LeaseID

	// txn
	cmps    []Cmp
	thenOps []Op
	elseOps []Op
}

// OpPut returns "put" operation based on given key-value and operation options.
func OpPut(key, val string, opts ...OpOption) Op {
	ret := Op{t: tPut, key: []byte(key), val: []byte(val)}
	ret.applyOpts(opts)
	switch {
	case ret.end != nil:
		panic("unexpected range in put")
	case ret.limit != 0:
		panic("unexpected limit in put")
	case ret.rev != 0:
		panic("unexpected revision in put")
	case ret.sort != nil:
		panic("unexpected sort in put")
	case ret.serializable:
		panic("unexpected serializable in put")
	case ret.countOnly:
		panic("unexpected countOnly in put")
	case ret.minModRev != 0, ret.maxModRev != 0:
		panic("unexpected mod revision filter in put")
	case ret.minCreateRev != 0, ret.maxCreateRev != 0:
		panic("unexpected create revision filter in put")
	case ret.filterDelete, ret.filterPut:
		panic("unexpected filter in put")
	case ret.createdNotify:
		panic("unexpected createdNotify in put")
	}
	return ret
}

func (op *Op) applyOpts(opts []OpOption) {
	for _, opt := range opts {
		opt(op)
	}
}

// OpOption configures Operations like Get, Put, Delete.
type OpOption func(*Op)

func getPrefix(key []byte) []byte {
	end := make([]byte, len(key))
	copy(end, key)
	for i := len(end) - 1; i >= 0; i-- {
		if end[i] < 0xff {
			end[i] = end[i] + 1
			end = end[:i+1]
			return end
		}
	}
	// next prefix does not exist (e.g., 0xffff);
	// default to WithFromKey policy
	return noPrefixEnd
}