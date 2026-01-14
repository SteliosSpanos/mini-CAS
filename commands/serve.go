package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/SteliosSpanos/mini-CAS/pkg/server"
)

func Serve(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)

	port := fs.Int("port", getEnvInt("CAS_PORT", 8080), "HTTP port")
	host := fs.String("host", getEnv("CAS_HOST", "0.0.0.0"), "Bind address")
	authToken := fs.String("auth-token", getEnv("CAS_AUTH_TOKEN", ""), "Bearer token for write operations (optional)")
	corsOrigins := fs.String("cors-origins", getEnv("CAS_CORS_ORIGINS", "*"), "Comma-seperated CORS origins")

	fs.Parse(args)

	config := server.Config{
		Port:         *port,
		Host:         *host,
		AuthToken:    *authToken,
		CORSOrigins:  parseCORSOrigins(*corsOrigins),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		RepoPath:     ".",
	}

	srv, err := server.NewServer(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal...")
		cancel()
	}()

	if err := srv.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func parseCORSOrigins(origins string) []string {
	if origins == "" {
		return []string{}
	}

	parts := strings.Split(origins, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result

}
