package merkle

type Tree struct {
	Root     *Node
	Leaves   []*Node
	HashFunc func(data []byte) string
}

func NewTree(hashFunc func(data []byte) string) *Tree {
	return &Tree{
		Root:     nil,
		Leaves:   nil,
		HashFunc: hashFunc,
	}
}
