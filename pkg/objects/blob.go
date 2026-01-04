package objects

import (
	"crypto/sha256"
	"fmt"
)

type Blob struct {
	Data []byte
}

func NewBlob(data []byte) *Blob {
	return &Blob{Data: data}
}

func Hash(blob Blob) string {
	hash := sha256.Sum256(blob.Data)

	return fmt.Sprintf("%x", hash)
}
