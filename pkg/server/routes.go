package server

import (
	"net/http"
)

func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /blob/{hash}", s.handleGetBlob)
	mux.HandleFunc("HEAD /blob/{hash}", s.handleGetBlob)
	mux.HandleFunc("GET /blob/{hash}/stat", s.handleStatBlob)
	mux.HandleFunc("GET /catalog", s.handleGetCatalog)
	mux.HandleFunc("POST /blobs", s.handlePostBlob)

	handler := Chain(mux,
		s.RecoveryMiddleware,
		s.LoggingMiddleware,
		s.CORSMiddleware,
		s.AuthMiddleware,
	)

	return handler
}
