// Package server implements NodePalette's read-only REST API for pipeline
// builders. It proxies data from NodeVault's catalog endpoints and exposes
// a tool palette suitable for DagEdit and similar pipeline-building clients.
package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/HeaInSeo/NodePalette/pkg/paletteclient"
)

// Server is the NodePalette HTTP server.
type Server struct {
	client *paletteclient.Client
	mux    *http.ServeMux
}

// New creates a Server backed by the given NodeVault client.
func New(client *paletteclient.Client) *Server {
	s := &Server{client: client, mux: http.NewServeMux()}
	s.mux.HandleFunc("/v1/palette/tools", s.handleListTools)
	s.mux.HandleFunc("/v1/palette/tools/", s.handleGetTool)
	s.mux.HandleFunc("/healthz", s.handleHealthz)
	return s
}

// Handler returns the HTTP handler for the server.
func (s *Server) Handler() http.Handler { return s.mux }

// PaletteTool is the canonical NodePalette wire type for a tool entry.
// It combines certified-tool metadata with registration details.
type PaletteTool struct {
	CasHash         string    `json:"cas_hash"`
	ToolName        string    `json:"tool_name"`
	Version         string    `json:"version"`
	StableRef       string    `json:"stable_ref"`
	ImageDigest     string    `json:"image_digest"`
	ImageRef        string    `json:"image_ref"`
	DisplayLabel    string    `json:"display_label"`
	DisplayCategory string    `json:"display_category"`
	PromotionStatus string    `json:"promotion_status"`
	CertifiedAt     time.Time `json:"certified_at"`
	ValidationHash  string    `json:"validation_hash"`
}

// ListToolsResponse is the top-level envelope for GET /v1/palette/tools.
type ListToolsResponse struct {
	Tools []PaletteTool `json:"tools"`
	Total int           `json:"total"`
}

// handleListTools serves GET /v1/palette/tools — returns all active certified tools.
func (s *Server) handleListTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	resp, err := s.client.ListCertifiedTools(ctx)
	if err != nil {
		slog.Error("failed to fetch certified tools from NodeVault", "err", err)
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}

	tools := make([]PaletteTool, 0, len(resp.Tools))
	for _, t := range resp.Tools {
		if t.PromotionStatus != "active" {
			continue
		}
		tools = append(tools, PaletteTool{
			CasHash:         t.CasHash,
			ToolName:        t.ToolName,
			Version:         t.Version,
			StableRef:       t.StableRef,
			ImageDigest:     t.ImageDigest,
			ImageRef:        t.ImageRef,
			DisplayLabel:    t.DisplayLabel,
			DisplayCategory: t.DisplayCategory,
			PromotionStatus: t.PromotionStatus,
			CertifiedAt:     t.CertifiedAt,
			ValidationHash:  t.ValidationHash,
		})
	}

	writeJSON(w, ListToolsResponse{Tools: tools, Total: len(tools)})
}

// handleGetTool serves GET /v1/palette/tools/{cas_hash}.
func (s *Server) handleGetTool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	casHash := strings.TrimPrefix(r.URL.Path, "/v1/palette/tools/")
	if casHash == "" {
		http.Error(w, "missing cas_hash", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	resp, err := s.client.ListCertifiedTools(ctx)
	if err != nil {
		slog.Error("failed to fetch certified tools from NodeVault", "err", err)
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}

	for _, t := range resp.Tools {
		if t.CasHash == casHash {
			writeJSON(w, PaletteTool{
				CasHash:         t.CasHash,
				ToolName:        t.ToolName,
				Version:         t.Version,
				StableRef:       t.StableRef,
				ImageDigest:     t.ImageDigest,
				ImageRef:        t.ImageRef,
				DisplayLabel:    t.DisplayLabel,
				DisplayCategory: t.DisplayCategory,
				PromotionStatus: t.PromotionStatus,
				CertifiedAt:     t.CertifiedAt,
				ValidationHash:  t.ValidationHash,
			})
			return
		}
	}
	http.Error(w, "not found", http.StatusNotFound)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "err", err)
	}
}
