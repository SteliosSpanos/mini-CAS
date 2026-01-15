package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
	"github.com/SteliosSpanos/mini-CAS/pkg/storage"
)

type LocalClient struct {
	casDir  string
	catalog *catalog.Catalog
	mu      sync.RWMutex
}

func NewLocalClient(casDir string) (*LocalClient, error) {
	cat := catalog.NewCatalog(casDir)

	if err := cat.Load(); err != nil {
		return nil, fmt.Errorf("failed to load catalog: %w", err)
	}

	return &LocalClient{
		casDir:  casDir,
		catalog: cat,
	}, nil
}

func (c *LocalClient) Upload(ctx context.Context, reader io.Reader) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	hash, err := storage.WriteBlobStream(c.casDir, reader)
	if err != nil {
		return "", fmt.Errorf("upload failed: %w", err)
	}

	return hash, nil
}

func (c *LocalClient) Download(ctx context.Context, hash string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(hash) != 64 {
		return nil, ErrInvalidHash
	}

	reader, err := storage.OpenBlob(c.casDir, hash)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "not found") {
			return nil, ErrBlobNotFound
		}
		return nil, fmt.Errorf("download failed: %w", err)
	}

	return reader, nil
}
