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
	"github.com/GoreeCloud/goreecloud-mesh/internal/governance"
	"github.com/GoreeCloud/goreecloud-mesh/internal/mesh"
	"github.com/GoreeCloud/goreecloud-mesh/internal/platformregistry"
	"github.com/GoreeCloud/goreecloud-mesh/internal/store"
	"github.com/GoreeCloud/goreecloud-mesh/internal/trust"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:8787", "HTTP listen address")
	statePath := flag.String("state", "./mesh-state.json", "durable Mesh state path; empty disables persistence")
	platformRegistryPath := flag.String("platform-registry", "./mesh-platform-registry.json", "durable authority-preserving platform registry state path; empty disables persistence")
	attestationPath := flag.String("source-attestations", "./mesh-source-attestations.json", "durable source-attestation state path; empty disables persistence")
	runtimeEvidencePath := flag.String("runtime-evidence", "./mesh-runtime-evidence.json", "durable runtime contract evidence path; empty disables persistence")
	evidenceEnvelopePath := flag.String("evidence-envelopes", "./mesh-evidence-envelopes.json", "durable producer-authoritative evidence envelope path; empty disables persistence")
	recoveryEvidencePath := flag.String("everkeep-recovery-evidence", "./mesh-everkeep-recovery-evidence.json", "durable Everkeep recovery evidence path; empty disables persistence")
	identityJWKSURL := flag.String("identity-jwks-url", "", "GoreeCloud Identity JWKS endpoint; empty keeps authenticated APIs fail-closed")
	identityIssuer := flag.String("identity-issuer", trust.DefaultIdentityIssuer, "required GoreeCloud Identity token issuer")
	identityAudience := flag.String("identity-audience", trust.DefaultIdentityAudience, "required GoreeCloud Identity token audience")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	state, err := store.New(*statePath)
	if err != nil {
		logger.Error("load state", "error", err)
		os.Exit(1)
	}

	platformRegistry, err := platformregistry.NewPersistent(*platformRegistryPath)
	if err != nil {
		logger.Error("load platform registry", "error", err)
		os.Exit(1)
	}

	attestationRegistry, err := contracts.NewPersistentSourceAttestationRegistry(*attestationPath)
	if err != nil {
		logger.Error("load source attestations", "error", err)
		os.Exit(1)
	}

	contractRegistry, err := contracts.NewPersistentRegistry(*runtimeEvidencePath)
	if err != nil {
		logger.Error("load runtime evidence", "error", err)
		os.Exit(1)
	}

	evidenceEnvelopeRegistry, err := contracts.NewPersistentEvidenceEnvelopeRegistry(*evidenceEnvelopePath)
	if err != nil {
		logger.Error("load evidence envelopes", "error", err)
		os.Exit(1)
	}

	recoveryRegistry, err := governance.NewPersistentRecoveryRegistry(*recoveryEvidencePath, time.Now().UTC())
	if err != nil {
		logger.Error("load Everkeep recovery evidence", "error", err)
		os.Exit(1)
	}

	var verifier trust.Verifier
	if *identityJWKSURL != "" {
		identityVerifier := trust.NewIdentityJWTVerifier(*identityJWKSURL)
		identityVerifier.Issuer = *identityIssuer
		identityVerifier.Audience = *identityAudience
		verifier = identityVerifier
		logger.Info("GoreeCloud Identity verifier configured", "issuer", *identityIssuer, "audience", *identityAudience, "jwks_url", *identityJWKSURL)
	} else {
		logger.Warn("GoreeCloud Identity verifier is not configured; authenticated Mesh APIs will fail closed")
	}

	baseHandler := meshapi.NewAuthorizedWithRecoveryAndEvidence(mesh.New(state), contractRegistry, attestationRegistry, recoveryRegistry, evidenceEnvelopeRegistry, verifier, logger)
	handler := meshapi.WithPlatformRegistry(baseHandler, platformRegistry, verifier)
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
		logger.Info("mesh starting", "listen", *listen, "state", *statePath, "platform_registry", *platformRegistryPath, "source_attestations", *attestationPath, "runtime_evidence", *runtimeEvidencePath, "evidence_envelopes", *evidenceEnvelopePath, "everkeep_recovery_evidence", *recoveryEvidencePath)
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
