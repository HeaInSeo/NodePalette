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
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/HeaInSeo/NodePalette/pkg/paletteclient"
	"github.com/HeaInSeo/NodePalette/pkg/server"
)

func main() {
	addr := os.Getenv("NODEPALETTE_ADDR")
	if addr == "" {
		addr = ":8083"
	}

	slog.Info("NodePalette starting",
		"listen", addr,
		"vault_api", vaultAPIAddr(),
	)

	client := paletteclient.New()
	srv := server.New(client)

	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      srv.Handler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if err := httpSrv.ListenAndServe(); err != nil {
		log.Fatalf("NodePalette server exited: %v", err)
	}
}

func vaultAPIAddr() string {
	v := os.Getenv("NODEVAULT_API_ADDR")
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, v)
}
