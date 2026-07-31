package server

import (
	"net/http/httptest"
	"testing"
)

// TestWriteJSON_SerializationFailure exercises the real writeJSON directly
// (white-box, package server) rather than reimplementing its marshal-error
// branch, so a regression in the actual 500-on-marshal-failure path would be
// caught. chan int is not JSON-marshalable, forcing json.Marshal to fail.
func TestWriteJSON_SerializationFailure(t *testing.T) {
	w := httptest.NewRecorder()
	writeJSON(w, map[string]any{"bad": make(chan int)})

	if w.Code != 500 {
		t.Fatalf("writeJSON status = %d, want 500 on marshal failure", w.Code)
	}
	if body := w.Body.String(); body == "" {
		t.Fatal("writeJSON wrote no body on marshal failure")
	}
}
