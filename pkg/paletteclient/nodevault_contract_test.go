package paletteclient_test

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/HeaInSeo/NodePalette/pkg/paletteclient"
)

// nodeVaultWireSnapshot is NodeVault's CertifiedToolItem / ListCertifiedToolsResponse /
// toCertifiedToolItem source at the same NodeVault commit as goldenNodeVaultList.
// CI job "NodeVault Wire Sync" (make nodevault-wire-sync-check) diffs it against
// NodeVault main, and TestNodeVaultWireSnapshot_MatchesGolden ties it to the golden.
const nodeVaultWireSnapshot = "testdata/nodevault_f45d4a11_catalogrest_wire.go.txt"

// rawGoldenTools returns the golden's tool objects exactly as NodeVault sent them.
func rawGoldenTools(t *testing.T) []map[string]json.RawMessage {
	t.Helper()
	body, err := os.ReadFile(goldenNodeVaultList)
	if err != nil {
		t.Fatal(err)
	}
	var env struct {
		Tools []map[string]json.RawMessage `json:"tools"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Tools) == 0 {
		t.Fatal("golden has no tools")
	}
	return env.Tools
}

func rawString(t *testing.T, tool map[string]json.RawMessage, key string) string {
	t.Helper()
	v, ok := tool[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		t.Fatalf("golden %s is not a string: %s", key, v)
	}
	return s
}

// The authority fields NodeVault owns (cas_hash, stable_ref, image_digest, ...) reach
// NodePalette byte-for-byte: any normalisation or re-derivation fails here.
func TestListCertifiedTools_NodeVaultGoldenPassThrough(t *testing.T) {
	raw := rawGoldenTools(t)
	body, err := os.ReadFile(goldenNodeVaultList)
	if err != nil {
		t.Fatal(err)
	}
	got, err := paletteclient.NewWithAddr(serveBytes(t, body).URL).ListCertifiedTools(context.Background())
	if err != nil {
		t.Fatalf("decode NodeVault golden: %v", err)
	}
	if len(got.Tools) != len(raw) {
		t.Fatalf("decoded %d tools, golden has %d", len(got.Tools), len(raw))
	}
	for i, tool := range got.Tools {
		for key, value := range map[string]string{
			"cas_hash":         tool.CasHash,
			"stable_ref":       tool.StableRef,
			"image_digest":     tool.ImageDigest,
			"image_ref":        tool.ImageRef,
			"tool_name":        tool.ToolName,
			"version":          tool.Version,
			"promotion_status": tool.PromotionStatus,
			"validation_hash":  tool.ValidationHash,
		} {
			if want := rawString(t, raw[i], key); value != want {
				t.Errorf("tools[%d].%s = %q, NodeVault sent %q", i, key, value, want)
			}
		}
		for _, key := range []string{"cas_hash", "stable_ref", "image_digest"} {
			if rawString(t, raw[i], key) == "" {
				t.Errorf("golden tools[%d].%s is empty; the pass-through assert would be vacuous", i, key)
			}
		}
	}
}

// wireField is one JSON field of NodeVault's CertifiedToolItem as declared in the snapshot.
type wireField struct {
	goType    string
	omitempty bool
}

func snapshotWireFields(t *testing.T) map[string]wireField {
	t.Helper()
	src, err := os.ReadFile(nodeVaultWireSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	f, err := parser.ParseFile(token.NewFileSet(), nodeVaultWireSnapshot,
		append([]byte("package catalogrest\n\n"), src...), parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse NodeVault wire snapshot: %v", err)
	}
	fields := map[string]wireField{}
	ast.Inspect(f, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok || ts.Name.Name != "CertifiedToolItem" {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			t.Fatal("CertifiedToolItem is not a struct in the snapshot")
		}
		for _, field := range st.Fields.List {
			if field.Tag == nil {
				continue
			}
			tag, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				t.Fatal(err)
			}
			name, opts, _ := strings.Cut(reflect.StructTag(tag).Get("json"), ",")
			fields[name] = wireField{goType: types.ExprString(field.Type), omitempty: strings.Contains(opts, "omitempty")}
		}
		return false
	})
	if len(fields) == 0 {
		t.Fatal("no CertifiedToolItem fields found in the snapshot")
	}
	return fields
}

// The vendored golden and the vendored owner-source snapshot describe the same NodeVault
// wire: every golden key is a declared field, every non-omitempty field is present, and
// certified_at is declared int64 and sent as a JSON integer. Updating one without the
// other fails here; the NodeVault Wire Sync CI job catches the owner moving on.
func TestNodeVaultWireSnapshot_MatchesGolden(t *testing.T) {
	fields := snapshotWireFields(t)
	if got := fields["certified_at"]; got.goType != "int64" || got.omitempty {
		t.Fatalf("snapshot certified_at = %+v, want non-omitempty int64 (UnixMilli)", got)
	}
	for i, tool := range rawGoldenTools(t) {
		for key := range tool {
			if _, ok := fields[key]; !ok {
				t.Errorf("golden tools[%d] has %q, which NodeVault's CertifiedToolItem does not declare", i, key)
			}
		}
		for name, field := range fields {
			if _, ok := tool[name]; !ok && !field.omitempty {
				t.Errorf("golden tools[%d] lacks non-omitempty field %q", i, name)
			}
		}
		var ms int64
		if err := json.Unmarshal(tool["certified_at"], &ms); err != nil {
			t.Errorf("golden tools[%d].certified_at = %s is not a JSON integer: %v", i, tool["certified_at"], err)
		}
	}
}
