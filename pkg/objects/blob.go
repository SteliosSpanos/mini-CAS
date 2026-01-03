package objects

import (
	"crypto/sha1"
	"fmt"
)

type Blob struct {
	Data []byte
}

func NewBlob(data []byte) *Blob {
	return &Blob{Data: data}
}

func Hash(blob Blob) string {
	hash := sha1.Sum(blob.Data)

	return fmt.Sprintf("%x", hash)
}
