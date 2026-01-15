package client

import (
	"errors"
	"fmt"
)

var (
	ErrBlobNotFound        = errors.New("blob not found")
	ErrEntryNotFound       = errors.New("catalog entry not found")
	ErrCatalogNotSupported = errors.New("catalog operations not supported")
	ErrInvalidHash         = errors.New("invalid hash format")
)

type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}
