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

func (t *Tree) Build(leafHashes []string) error {
	if len(leafHashes) == 0 {
		return ErrEmptyLeaves
	}

	currentLevel := make([]*Node, len(leafHashes))
	t.Leaves = make([]*Node, len(leafHashes))

	for i, hash := range leafHashes {
		node := NewLeafNode(hash, i)
		currentLevel[i] = node
		t.Leaves[i] = node
	}

	for len(currentLevel) > 1 {
		if len(currentLevel)%2 == 1 {
			currentLevel = append(currentLevel, currentLevel[len(currentLevel)-1])
		}

		nextLevel := make([]*Node, len(currentLevel)/2)

		for i := 0; i < len(currentLevel); i += 2 {
			parent := NewInternalNode(currentLevel[i], currentLevel[i+1], t.HashFunc)
			nextLevel[i/2] = parent
		}

		currentLevel = nextLevel
	}

	t.Root = currentLevel[0]
	return nil
}

func (t *Tree) RootHash() (string, error) {
	if t.Root == nil {
		return "", ErrTreeNotBuilt
	}
	return t.Root.Hash, nil
}
