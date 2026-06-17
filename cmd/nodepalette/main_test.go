package main

import (
	"strings"
	"testing"
)

// Test 1: vaultAPIAddr strips newlines
func TestVaultAPIAddr_StripsNewlines(t *testing.T) {
	t.Setenv("NODEVAULT_API_ADDR", "http://localhost:8082\n")
	got := vaultAPIAddr()
	if strings.Contains(got, "\n") {
		t.Errorf("vaultAPIAddr() still contains newline: %q", got)
	}
	if got != "http://localhost:8082" {
		t.Errorf("vaultAPIAddr() = %q, want %q", got, "http://localhost:8082")
	}
}

// Test 2: vaultAPIAddr strips carriage returns
func TestVaultAPIAddr_StripsCarriageReturns(t *testing.T) {
	t.Setenv("NODEVAULT_API_ADDR", "http://localhost:8082\r\n")
	got := vaultAPIAddr()
	if strings.ContainsAny(got, "\r\n") {
		t.Errorf("vaultAPIAddr() still contains CR/LF: %q", got)
	}
	if got != "http://localhost:8082" {
		t.Errorf("vaultAPIAddr() = %q, want %q", got, "http://localhost:8082")
	}
}

// Test 3: vaultAPIAddr returns empty string when env not set
func TestVaultAPIAddr_EmptyWhenNotSet(t *testing.T) {
	t.Setenv("NODEVAULT_API_ADDR", "")
	got := vaultAPIAddr()
	if got != "" {
		t.Errorf("vaultAPIAddr() = %q, want empty string", got)
	}
}

// Test 4: vaultAPIAddr strips surrounding whitespace (tabs, spaces)
func TestVaultAPIAddr_StripsWhitespace(t *testing.T) {
	t.Setenv("NODEVAULT_API_ADDR", "  http://localhost:8082  ")
	got := vaultAPIAddr()
	if got != "http://localhost:8082" {
		t.Errorf("vaultAPIAddr() = %q, want %q", got, "http://localhost:8082")
	}
}

// Test 5: listenAddr returns default when env not set
func TestListenAddr_Default(t *testing.T) {
	t.Setenv("NODEPALETTE_ADDR", "")
	got := listenAddr()
	if got != ":8083" {
		t.Errorf("listenAddr() = %q, want %q", got, ":8083")
	}
}

// Test 6: listenAddr uses env var when set
func TestListenAddr_FromEnv(t *testing.T) {
	t.Setenv("NODEPALETTE_ADDR", ":9090")
	got := listenAddr()
	if got != ":9090" {
		t.Errorf("listenAddr() = %q, want %q", got, ":9090")
	}
}

// Test 7: listenAddr trims whitespace from env var
func TestListenAddr_TrimsWhitespace(t *testing.T) {
	t.Setenv("NODEPALETTE_ADDR", "  :9091  ")
	got := listenAddr()
	if got != ":9091" {
		t.Errorf("listenAddr() = %q, want %q", got, ":9091")
	}
}

// Test 8: listenAddr trims newlines (similar to vaultAPIAddr)
func TestListenAddr_TrimsNewline(t *testing.T) {
	t.Setenv("NODEPALETTE_ADDR", ":9092\n")
	got := listenAddr()
	if got != ":9092" {
		t.Errorf("listenAddr() = %q, want %q", got, ":9092")
	}
}
