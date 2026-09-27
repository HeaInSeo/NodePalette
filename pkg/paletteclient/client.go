// Package paletteclient provides a read-only HTTP client for NodeVault's
// catalog REST API. NodePalette calls these endpoints to serve the tool
// palette to pipeline-building clients (DagEdit, etc.).
//
// Default endpoint: NODEVAULT_API_ADDR (default http://nodevault.nodevault-system.svc:8082)
package paletteclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultVaultAPIAddr   = "http://nodevault.nodevault-system.svc:8082"
	certifiedToolsPath    = "/v1/catalog/certified-tools"
	defaultRequestTimeout = 10 * time.Second
	maxResponseBodyBytes  = 1 << 20
)

// Client queries NodeVault's catalog REST API.
type Client struct {
	baseURL string
	http    *http.Client
}

// New creates a Client. The base URL is read from NODEVAULT_API_ADDR.
func New() *Client {
	addr := os.Getenv("NODEVAULT_API_ADDR")
	if addr == "" {
		addr = defaultVaultAPIAddr
	}
	return &Client{
		baseURL: addr,
		http:    &http.Client{Timeout: defaultRequestTimeout},
	}
}

// NewWithAddr creates a Client with an explicit base URL (no env var reading).
// Useful for testing and for wiring the address at the call site.
func NewWithAddr(addr string) *Client {
	addr = strings.TrimRight(strings.TrimSpace(addr), "/")
	if addr == "" {
		addr = defaultVaultAPIAddr
	}
	return &Client{
		baseURL: addr,
		http:    &http.Client{Timeout: defaultRequestTimeout},
	}
}

// CertifiedTool is a single entry from NodeVault's certified tool catalog.
//
// NodeVault owns the wire format: certified_at is an int64 Unix time in
// milliseconds (catalogrest.CertifiedToolItem). It is decoded into a UTC
// time.Time; any other form (RFC3339 string, fraction, null, missing) is
// rejected, so the caller surfaces an upstream error instead of a wrong instant.
type CertifiedTool struct {
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

// UnmarshalJSON decodes NodeVault's CertifiedToolItem wire format.
func (t *CertifiedTool) UnmarshalJSON(data []byte) error {
	type wire CertifiedTool
	aux := struct {
		*wire
		CertifiedAt json.RawMessage `json:"certified_at"`
	}{wire: (*wire)(t)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(aux.CertifiedAt) == 0 || string(aux.CertifiedAt) == "null" {
		return fmt.Errorf("paletteclient: certified_at missing for %q", t.CasHash)
	}
	var ms int64
	if err := json.Unmarshal(aux.CertifiedAt, &ms); err != nil {
		return fmt.Errorf("paletteclient: certified_at for %q must be int64 Unix milliseconds: %w", t.CasHash, err)
	}
	t.CertifiedAt = time.UnixMilli(ms).UTC()
	return nil
}

// ListCertifiedToolsResponse is the JSON envelope from NodeVault.
type ListCertifiedToolsResponse struct {
	Tools []CertifiedTool `json:"tools"`
}

// ListCertifiedTools returns all active certified tools from NodeVault.
func (c *Client) ListCertifiedTools(ctx context.Context) (*ListCertifiedToolsResponse, error) {
	var resp ListCertifiedToolsResponse
	if err := c.get(ctx, c.baseURL+certifiedToolsPath, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) get(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("paletteclient: new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("paletteclient: GET %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes))
	if readErr != nil {
		return fmt.Errorf("paletteclient: read response from %s: %w", url, readErr)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("paletteclient: GET %s: HTTP %d: %s", url, resp.StatusCode, body)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("paletteclient: decode response from %s: %w", url, err)
	}
	return nil
}
