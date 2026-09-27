package server_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/HeaInSeo/NodePalette/pkg/paletteclient"
	"github.com/HeaInSeo/NodePalette/pkg/server"
)

// These tests run the real paletteclient against NodeVault's own wire bytes (golden
// generated from NodeVault f45d4a11 catalogrest), not NodePalette's own marshal.

const goldenNodeVaultList = "../paletteclient/testdata/nodevault_f45d4a11_certified_tools.json"

// fakeNodeVault serves body with status on /v1/catalog/certified-tools.
func fakeNodeVault(t *testing.T, status int, body []byte) *httptest.Server {
	t.Helper()
	nv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/catalog/certified-tools" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(nv.Close)
	return nv
}

func paletteOver(t *testing.T, nv *httptest.Server) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(server.New(paletteclient.NewWithAddr(nv.URL)).Handler())
	t.Cleanup(ts.Close)
	return ts
}

func golden(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(goldenNodeVaultList)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// M1/M3/M5: a non-empty NodeVault list is served (200), only active tools are kept, and
// certified_at keeps the exact millisecond instant as an RFC3339 string outbound.
func TestNodeVaultGolden_ListTools(t *testing.T) {
	ts := paletteOver(t, fakeNodeVault(t, http.StatusOK, golden(t)))

	resp, err := http.Get(ts.URL + "/v1/palette/tools")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for a non-empty NodeVault catalog, got %d", resp.StatusCode)
	}
	var raw struct {
		Tools []map[string]any `json:"tools"`
		Total int              `json:"total"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		t.Fatal(err)
	}
	if raw.Total != 2 || len(raw.Tools) != 2 {
		t.Fatalf("expected the 2 active tools, got total=%d tools=%v", raw.Total, raw.Tools)
	}
	if got := raw.Tools[0]["certified_at"]; got != "2026-09-24T10:15:42.123Z" {
		t.Errorf("outbound certified_at = %#v, want RFC3339 \"2026-09-24T10:15:42.123Z\"", got)
	}
	if got := raw.Tools[1]["certified_at"]; got != "0001-01-01T00:00:00Z" {
		t.Errorf("zero-epoch outbound certified_at = %#v, want \"0001-01-01T00:00:00Z\"", got)
	}
}

// M4: an active cas_hash from the NodeVault list is served with the same instant.
func TestNodeVaultGolden_GetTool(t *testing.T) {
	ts := paletteOver(t, fakeNodeVault(t, http.StatusOK, golden(t)))

	resp, err := http.Get(ts.URL + "/v1/palette/tools/sha256:1111111111111111111111111111111111111111111111111111111111111111")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var tool map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&tool); err != nil {
		t.Fatal(err)
	}
	if tool["tool_name"] != "bwa" || tool["certified_at"] != "2026-09-24T10:15:42.123Z" {
		t.Errorf("tool = %v", tool)
	}
}

// M7/M8: an RFC3339 (legacy/mock) certified_at, malformed JSON, or a non-2xx upstream
// keeps the existing 502 boundary on both endpoints.
func TestNodeVaultGolden_BadUpstreamIs502(t *testing.T) {
	legacy := strings.Replace(string(golden(t)), `"certified_at":1790244942123`, `"certified_at":"2026-09-24T10:15:42.123Z"`, 1)
	if legacy == string(golden(t)) {
		t.Fatal("fixture substitution failed")
	}
	cases := map[string]*httptest.Server{
		"rfc3339 certified_at": fakeNodeVault(t, http.StatusOK, []byte(legacy)),
		"malformed json":       fakeNodeVault(t, http.StatusOK, []byte(`{"tools":[`)),
		"upstream 500":         fakeNodeVault(t, http.StatusInternalServerError, []byte(`index error`)),
	}
	for name, nv := range cases {
		ts := paletteOver(t, nv)
		for _, path := range []string{"/v1/palette/tools", "/v1/palette/tools/sha256:1111111111111111111111111111111111111111111111111111111111111111"} {
			resp, err := http.Get(ts.URL + path)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusBadGateway {
				t.Errorf("%s %s: status %d, want 502", name, path, resp.StatusCode)
			}
		}
	}
}
