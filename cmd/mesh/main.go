package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	meshapi "github.com/GoreeCloud/goreecloud-mesh/internal/api"
	"github.com/GoreeCloud/goreecloud-mesh/internal/contracts"
	"github.com/GoreeCloud/goreecloud-mesh/internal/mesh"
	"github.com/GoreeCloud/goreecloud-mesh/internal/store"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8787", "HTTP listen address")
	statePath := flag.String("state", "./mesh-state.json", "durable Mesh state path; empty disables persistence")
	attestationPath := flag.String("source-attestations", "./mesh-source-attestations.json", "durable source-attestation state path; empty disables persistence")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	state, err := store.New(*statePath)
	if err != nil {
		logger.Error("load state", "error", err)
		os.Exit(1)
	}

	attestationRegistry, err := contracts.NewPersistentSourceAttestationRegistry(*attestationPath)
	if err != nil {
		logger.Error("load source attestations", "error", err)
		os.Exit(1)
	}

	contractRegistry := contracts.NewRegistry()
	handler := meshapi.NewWithAttestations(mesh.New(state), contractRegistry, attestationRegistry, logger)
	server := &http.Server{
		Addr:              *listen,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("mesh starting", "listen", *listen, "state", *statePath, "source_attestations", *attestationPath)
		errCh <- server.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown", "error", err)
			os.Exit(1)
		}
		logger.Info("mesh stopped")
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			logger.Error("server", "error", err)
			os.Exit(1)
		}
	}
}
