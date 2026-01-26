package merkle

import "errors"

var (
	ErrEmptyLeaves      = errors.New("cannot build tree from empty leaf list")
	ErrTreeNotBuilt     = errors.New("tree has not been built")
	ErrIndexOutOfBounds = errors.New("leaf index out of bounds")
)
