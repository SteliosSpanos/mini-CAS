package client

import (
	"context"
	"io"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
)

type Client interface {
	BlobOperations
	CatalogOperations
	io.Closer
}

type BlobOperations interface {
	Upload(ctx context.Context, reader io.Reader) (string, error)
	Download(ctx context.Context, hash string) (io.ReadCloser, error)
	Stat(ctx context.Context, hash string) (BlobInfo, error)
	Exists(ctx context.Context, hash string) (bool, error)
}

type CatalogOperations interface {
	GetCatalog(ctx context.Context) ([]catalog.Entry, error)
	GetEntry(ctx context.Context, filepath string) (catalog.Entry, error)
	AddEntry(ctx context.Context, entry catalog.Entry) error
	SaveCatalog(ctx context.Context) error
}

type BlobInfo struct {
	Hash   string
	Size   int64
	Exists bool
}
