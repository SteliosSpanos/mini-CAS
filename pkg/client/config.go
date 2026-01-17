package client

import (
	"fmt"
	"os"
)

type Config struct {
	ServerURL string
	AuthToken string
	CASDir    string
}

func NewClient(cfg Config) (Client, error) {
	if cfg.ServerURL != "" {
		return NewHTTPClient(cfg.ServerURL, cfg.AuthToken), nil
	}

	if cfg.CASDir == "" {
		return nil, fmt.Errorf("CASDir is required for local client")
	}

	return NewLocalClient(cfg.CASDir)
}

func NewClientFromEnv() (Client, error) {
	cfg := Config{
		ServerURL: os.Getenv("CAS_SERVER_URL"),
		AuthToken: os.Getenv("CAS_AUTH_TOKEN"),
		CASDir:    os.Getenv("CAS_DIR"),
	}

	if cfg.CASDir == "" && cfg.ServerURL == "" {
		cfg.CASDir = ".cas"
	}

	return NewClient(cfg)
}
