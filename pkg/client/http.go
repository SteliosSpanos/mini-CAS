package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient struct {
	baseURL   string
	authToken string
	client    *http.Client
}

func NewHTTPClient(baseURL, authToken string) *HTTPClient {
	return &HTTPClient{
		baseURL:   baseURL,
		authToken: authToken,
		client: &http.Client{
			Timeout: 0,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (c *HTTPClient) Upload(ctx context.Context, reader io.Reader) (string, error) {
	url := c.baseURL + "/blobs"

	req, err := http.NewRequestWithContext(ctx, "POST", url, reader)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", &HTTPError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	var result struct {
		Hash string `json:"hash"`
		Size int64  `json:"size"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Hash, nil
}

func (c *HTTPClient) Download(ctx context.Context, hash string) (io.ReadCloser, error) {
	if len(hash) != 64 {
		return nil, ErrInvalidHash
	}

	url := fmt.Sprintf("%s/blob/%s", c.baseURL, hash)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer"+c.authToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download request failed: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, ErrBlobNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		resp.Body.Close()
		return nil, &HTTPError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	return resp.Body, nil
}

func (c *HTTPClient) Stat(ctx context.Context, hash string) (BlobInfo, error) {
	if len(hash) != 64 {
		return BlobInfo{}, ErrInvalidHash
	}

	url := fmt.Sprintf("%s/blob/%s/stat", c.baseURL, hash)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return BlobInfo{}, fmt.Errorf("failed to create request: %w", err)
	}

	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return BlobInfo{}, fmt.Errorf("stat request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return BlobInfo{}, &HTTPError{StatusCode: resp.StatusCode, Message: string(body)}
	}

	var result struct {
		Hash   string `json:"hash"`
		Size   int64  `json:"size"`
		Exists bool   `json:"exists"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return BlobInfo{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return BlobInfo{
		Hash:   result.Hash,
		Size:   result.Size,
		Exists: result.Exists,
	}, nil
}

func (c *HTTPClient) Exists(ctx context.Context, hash string) (bool, error) {
	if len(hash) != 64 {
		return false, ErrInvalidHash
	}

	url := fmt.Sprintf("%s/blob/%s", c.baseURL, hash)

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("exists request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	return false, &HTTPError{StatusCode: resp.StatusCode, Message: "unexpected status"}
}
