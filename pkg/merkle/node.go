package merkle

type Node struct {
	Hash   string
	Left   *Node
	Right  *Node
	isLeaf bool
	Index  int
}

func NewLeafNode(hash string, index int) *Node {
	return &Node{
		Hash:   hash,
		Left:   nil,
		Right:  nil,
		isLeaf: true,
		Index:  index,
	}
}

func NewInternalNode(left, right *Node, hashFunc func([]byte) string) *Node {
	return &Node{
		Hash:   hashPair(left.Hash, right.Hash, hashFunc),
		Left:   left,
		Right:  right,
		isLeaf: false,
		Index:  0,
	}
}

func hashPair(left, right string, hashFunc func([]byte) string) string {
	combined := left + right
	return hashFunc([]byte(combined))
}
