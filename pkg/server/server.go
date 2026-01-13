package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/SteliosSpanos/mini-CAS/pkg/catalog"
	"github.com/SteliosSpanos/mini-CAS/pkg/path"
)

type Config struct {
	Port         int
	Host         string
	AuthToken    string
	CORSOrigins  []string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	RepoPath     string
}

type Server struct {
	config     Config
	httpServer *http.Server
	catalog    *catalog.Catalog
	casDir     string
	logger     *log.Logger
}

func NewServer(config Config) (*Server, error) {
	repo, err := path.Open(config.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	cat := catalog.NewCatalog(repo.RootDir)
	if err := cat.Load(); err != nil {
		return nil, fmt.Errorf("failed to load catalog: %w", err)
	}

	logger := log.New(os.Stdout, "[CAS-SERVER]", log.LstdFlags)

	server := &Server{
		config:  config,
		catalog: cat,
		casDir:  repo.RootDir,
		logger:  logger,
	}

	return server, nil
}

func (s *Server) Start(ctx context.Context) error {
	handler := s.setupRoutes()

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      handler,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
	}

	s.logger.Printf("Starting server on %s:%d", s.config.Host, s.config.Port)

	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		s.logger.Println("Shutdown signal received, gracefully stopping server...")
		return s.Shutdown(context.Background())
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	s.logger.Println("Shutting down server...")
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	s.logger.Println("Server stopped")
	return nil
}
