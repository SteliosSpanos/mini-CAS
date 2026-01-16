package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

func (c *LocalClient) Stat(ctx context.Context, hash string) (BlobInfo, error) {
	if err := ctx.Err(); err != nil {
		return BlobInfo{}, err
	}

	if len(hash) != 64 {
		return BlobInfo{}, ErrInvalidHash
	}

	blobPath := filepath.Join(c.casDir, "storage", hash[:2], hash[2:4], hash)

	info, err := os.Stat(blobPath)
	if err != nil {
		if os.IsNotExist(err) {
			return BlobInfo{Hash: hash, Exists: false}, nil
		}
		return BlobInfo{}, fmt.Errorf("stat failed: %w", err)
	}

	return BlobInfo{
		Hash:   hash,
		Size:   info.Size(),
		Exists: true,
	}, nil
}

func (c *LocalClient) Exists(ctx context.Context, hash string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	if len(hash) != 64 {
		return false, ErrInvalidHash
	}

	blobPath := filepath.Join(c.casDir, hash[:2], hash[2:4], hash)

	_, err := os.Stat(blobPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("stat failed: %w", err)
	}

	return true, nil
}

func (c *LocalClient) GetCatalog(ctx context.Context) ([]catalog.Entry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if err := c.catalog.Load(); err != nil {
		return nil, fmt.Errorf("failed to reload catalog: %w", err)
	}

	return c.catalog.ListEntries(), nil
}

func (c *LocalClient) GetEntry(ctx context.Context, filepath string) (catalog.Entry, error) {
	if err := ctx.Err(); err != nil {
		return catalog.Entry{}, err
	}

	c.mu.RLock()
	defer c.mu.Unlock()

	if err := c.catalog.Load(); err != nil {
		return catalog.Entry{}, fmt.Errorf("failed to relaod catalog: %w", err)
	}

	entry, err := c.catalog.GetEntry(filepath)
	if err != nil {
		return catalog.Entry{}, ErrEntryNotFound
	}

	return entry, nil
}

func (c *LocalClient) AddEntry(ctx context.Context, entry catalog.Entry) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.catalog.AddEntry(entry)
	return nil
}

func (c *LocalClient) SaveCatalog(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.catalog.Save(); err != nil {
		return fmt.Errorf("failed to save catalog: %w", err)
	}

	return nil
}

func (c *LocalClient) Close() error {
	return nil
}
