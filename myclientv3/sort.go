package clientv3

type SortTarget int
type SortOrder int

const (
	SortNone SortOrder = iota
	SortAscend
	SortDescend
)

const (
	SortByKey SortTarget = iota
	SortByVersion
	SortByCreateRevision
	SortByModRevision
	SortByValue
)

type SortOption struct {
	Target SortTarget
	Order  SortOrder
}
