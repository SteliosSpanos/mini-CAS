package server

import (
	"net/http"
)

func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /blobs/{hash}", s.handleGetBlob)
	mux.HandleFunc("HEAD /blobs/{hash}", s.handleGetBlob)
	mux.HandleFunc("GET /blobs/{hash}/stat", s.handleStatBlob)
	mux.HandleFunc("GET /catalog", s.handleGetCatalog)
	mux.HandleFunc("POST /blobs", s.handlePostBlob)
	mux.HandleFunc("POST /catalog", s.handlePostCatalog)

	handler := Chain(mux,
		s.RecoveryMiddleware,
		s.LoggingMiddleware,
		s.CORSMiddleware,
		s.AuthMiddleware,
	)

	return handler
}
