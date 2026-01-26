package merkle

type Proof struct {
	LeafHash  string
	LeafIndex int
	Siblings  []string
	RootHash  string
}

func (t *Tree) GenerateProof(index int) (*Proof, error) {
	if t.Root == nil {
		return nil, ErrTreeNotBuilt
	}

	if index < 0 || index >= len(t.Leaves) {
		return nil, ErrIndexOutOfBounds
	}

	siblings := []string{}
	currentIndex := index
	currentLevel := t.Leaves

	for len(currentLevel) > 1 {
		if len(currentLevel)%2 == 1 {
			currentLevel = append(currentLevel, currentLevel[len(currentLevel)-1])
		}

		var siblingIndex int
		if currentIndex%2 == 0 {
			siblingIndex = currentIndex + 1
		} else {
			siblingIndex = currentIndex - 1
		}

		siblings = append(siblings, currentLevel[siblingIndex].Hash)

		nextLevel := make([]*Node, len(currentLevel)/2)
		for i := 0; i < len(currentLevel); i += 2 {
			nextLevel[i/2] = NewInternalNode(currentLevel[i], currentLevel[i+1], t.HashFunc)
		}

		currentLevel = nextLevel
		currentIndex = currentIndex / 2
	}

	return &Proof{
		LeafHash:  t.Leaves[index].Hash,
		LeafIndex: index,
		Siblings:  siblings,
		RootHash:  t.Root.Hash,
	}, nil
}
