package paletteclient_test

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/HeaInSeo/NodePalette/pkg/paletteclient"
)

// helpers

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return string(b)
}

// goldenNodeVaultList is NodeVault's own GET /v1/catalog/certified-tools body, produced by
// catalogrest.toCertifiedToolItem + writeJSON at NodeVault f45d4a11bdfffd49ebe1e91841e666e5ba61017b
// (not hand-written, and not NodePalette's own marshal). certified_at is int64 UnixMilli.
const goldenNodeVaultList = "testdata/nodevault_f45d4a11_certified_tools.json"

func serveBytes(t *testing.T, body []byte) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/catalog/certified-tools" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("missing Accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(ts.Close)
	return ts
}

// Test 1 (M1/M5/M6/M9): NodeVault's real wire bytes decode; certified_at is the exact
// millisecond instant in UTC, unknown extra fields are ignored, and the zero-time epoch
// decodes without error.
func TestListCertifiedTools_NodeVaultGolden(t *testing.T) {
	body, err := os.ReadFile(goldenNodeVaultList)
	if err != nil {
		t.Fatal(err)
	}
	ts := serveBytes(t, body)

	got, err := paletteclient.NewWithAddr(ts.URL).ListCertifiedTools(context.Background())
	if err != nil {
		t.Fatalf("decode NodeVault golden: %v", err)
	}
	if len(got.Tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(got.Tools))
	}
	g := got.Tools[0]
	want := time.Date(2026, 9, 24, 10, 15, 42, 123000000, time.UTC)
	if !g.CertifiedAt.Equal(want) || g.CertifiedAt.Location() != time.UTC {
		t.Errorf("CertifiedAt: got %v, want %v (UTC)", g.CertifiedAt, want)
	}
	if g.CasHash != "sha256:1111111111111111111111111111111111111111111111111111111111111111" ||
		g.ToolName != "bwa" || g.PromotionStatus != "active" || g.ImageRef == "" || g.ValidationHash == "" {
		t.Errorf("fields not decoded: %+v", g)
	}
	if !got.Tools[1].CertifiedAt.Equal(time.Date(2026, 7, 31, 15, 0, 0, 0, time.UTC)) {
		t.Errorf("second CertifiedAt: got %v", got.Tools[1].CertifiedAt)
	}
	if !got.Tools[2].CertifiedAt.Equal(time.Time{}) {
		t.Errorf("zero epoch: got %v, want %v", got.Tools[2].CertifiedAt, time.Time{})
	}
}

// Test 2 (M7): certified_at in any form other than int64 UnixMilli is rejected, never
// accepted by accident — including the RFC3339 string NodePalette itself emits.
func TestListCertifiedTools_RejectsNonUnixMilliCertifiedAt(t *testing.T) {
	for name, value := range map[string]string{
		"rfc3339 string": `"2026-09-24T10:15:42Z"`,
		"fraction":       `1790244942123.5`,
		"null":           `null`,
		"bool":           `true`,
	} {
		t.Run(name, func(t *testing.T) {
			ts := serveBytes(t, []byte(`{"tools":[{"cas_hash":"h","promotion_status":"active","certified_at":`+value+`}]}`))
			_, err := paletteclient.NewWithAddr(ts.URL).ListCertifiedTools(context.Background())
			if err == nil || !strings.Contains(err.Error(), "certified_at") {
				t.Fatalf("certified_at=%s: err = %v, want certified_at decode error", value, err)
			}
			if !strings.Contains(err.Error(), ts.URL) {
				t.Errorf("error lost the upstream URL: %v", err)
			}
		})
	}
	t.Run("missing", func(t *testing.T) {
		ts := serveBytes(t, []byte(`{"tools":[{"cas_hash":"h","promotion_status":"active"}]}`))
		if _, err := paletteclient.NewWithAddr(ts.URL).ListCertifiedTools(context.Background()); err == nil {
			t.Fatal("missing certified_at accepted")
		}
	})
}

// Test 3: HTTP 500 → error returned
func TestListCertifiedTools_500Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := paletteclient.NewWithAddr(ts.URL)
	_, err := c.ListCertifiedTools(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP 500") {
		t.Errorf("error message should mention HTTP 500, got: %v", err)
	}
}

// Test 4: connection failure → error returned (closed server)
func TestListCertifiedTools_ConnectionFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	addr := ts.URL
	ts.Close() // close immediately so connection is refused

	c := paletteclient.NewWithAddr(addr)
	_, err := c.ListCertifiedTools(context.Background())
	if err == nil {
		t.Fatal("expected error for connection failure, got nil")
	}
}

// Test 5: invalid JSON → unmarshal error
func TestListCertifiedTools_InvalidJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html>not json</html>"))
	}))
	defer ts.Close()

	c := paletteclient.NewWithAddr(ts.URL)
	_, err := c.ListCertifiedTools(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid JSON response, got nil")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("error message should mention decode, got: %v", err)
	}
}

// Test 6: trailing slash in baseURL is trimmed — no double slashes.
func TestNewWithAddr_TrailingSlash(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(paletteclient.ListCertifiedToolsResponse{})
	}))
	defer ts.Close()

	// NewWithAddr with trailing slash — path must not contain double slashes
	c := paletteclient.NewWithAddr(ts.URL + "/")
	_, err := c.ListCertifiedTools(context.Background())
	if err != nil {
		t.Fatalf("unexpected error with trailing slash URL: %v", err)
	}
	if gotPath != "/v1/catalog/certified-tools" {
		t.Errorf("expected path /v1/catalog/certified-tools, got %q", gotPath)
	}
}

// Test 7: broken body (connection closed mid-stream) → error from io.ReadAll
func TestListCertifiedTools_BrokenBody(t *testing.T) {
	// Use a raw TCP listener to send a partial HTTP response then abruptly close.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	go func() {
		conn, acceptErr := ln.Accept()
		if acceptErr != nil {
			return
		}
		// Read the HTTP request (consume it so the client sends it)
		br := bufio.NewReader(conn)
		for {
			line, _ := br.ReadString('\n')
			if line == "\r\n" || line == "" {
				break
			}
		}
		// Write a partial HTTP/1.1 response with Content-Length larger than what we send
		_, _ = conn.Write([]byte(
			"HTTP/1.1 200 OK\r\n" +
				"Content-Type: application/json\r\n" +
				"Content-Length: 1000\r\n" +
				"\r\n" +
				"{", // send only 1 byte of the promised 1000-byte body
		))
		_ = conn.Close() // abrupt close
	}()

	addr := "http://" + ln.Addr().String()
	c := paletteclient.NewWithAddr(addr)
	_, err = c.ListCertifiedTools(context.Background())
	if err == nil {
		t.Fatal("expected error for broken body, got nil")
	}
}

// Test 8: empty tools list (null vs empty array) — valid JSON with empty tools
func TestListCertifiedTools_EmptyList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tools":[]}`))
	}))
	defer ts.Close()

	c := paletteclient.NewWithAddr(ts.URL)
	got, err := c.ListCertifiedTools(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got.Tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(got.Tools))
	}
}

// Test 9: context cancellation → error propagated
func TestListCertifiedTools_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	c := paletteclient.NewWithAddr(ts.URL)
	_, err := c.ListCertifiedTools(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context, got nil")
	}
}

// Test 10: New() uses env var NODEVAULT_API_ADDR (backward compat)
func TestNew_UsesEnvVar(t *testing.T) {
	// Just verify New() creates a client without panicking;
	// the env var is read at construction time.
	t.Setenv("NODEVAULT_API_ADDR", "http://localhost:9999")
	c := paletteclient.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
}

// Test 11: New() defaults when env var is empty
func TestNew_DefaultAddr(t *testing.T) {
	t.Setenv("NODEVAULT_API_ADDR", "")
	c := paletteclient.New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNewWithAddr_EmptyUsesDefault(t *testing.T) {
	c := paletteclient.NewWithAddr("   ")
	if c == nil {
		t.Fatal("NewWithAddr returned nil")
	}
}

// Suppress unused import warning for mustMarshal
var _ = mustMarshal
