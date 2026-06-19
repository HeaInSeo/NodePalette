// NodePalette — read-only REST service that exposes the certified tool palette
// to pipeline-building clients (DagEdit, etc.).
//
// Data source: NodeVault catalog REST API (NODEVAULT_API_ADDR).
// Serves:      GET /v1/palette/tools
//
//	GET /v1/palette/tools/{cas_hash}
//	GET /healthz
package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/HeaInSeo/NodePalette/pkg/paletteclient"
	"github.com/HeaInSeo/NodePalette/pkg/server"
)

const defaultVaultAPIAddr = "http://nodevault.nodevault-system.svc:8082"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	addr := listenAddr()

	slog.Info("NodePalette starting",
		"listen", addr,
		"vault_api", vaultAPIAddr(),
	)

	client := paletteclient.NewWithAddr(vaultAPIAddr())
	srv := server.New(client)

	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      srv.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		<-ctx.Done()
		stop()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			log.Printf("NodePalette graceful shutdown error: %v", err)
		}
	}()

	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("NodePalette server exited: %v", err)
	}
	slog.Info("NodePalette stopped")
}

// listenAddr returns the address NodePalette listens on.
// It reads NODEPALETTE_ADDR and trims whitespace; defaults to ":8083".
func listenAddr() string {
	addr := strings.TrimSpace(os.Getenv("NODEPALETTE_ADDR"))
	if addr == "" {
		addr = ":8083"
	}
	return addr
}

// vaultAPIAddr returns the NodeVault catalog API base URL.
// It reads NODEVAULT_API_ADDR and strips surrounding whitespace.
func vaultAPIAddr() string {
	addr := strings.TrimSpace(os.Getenv("NODEVAULT_API_ADDR"))
	if addr == "" {
		return defaultVaultAPIAddr
	}
	return addr
}
