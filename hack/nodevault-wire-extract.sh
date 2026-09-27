#!/usr/bin/env bash
# Prints NodeVault's certified-tools wire definitions — the CertifiedToolItem and
# ListCertifiedToolsResponse types and toCertifiedToolItem — from NodeVault's
# pkg/catalogrest/server.go. `make nodevault-wire-sync-check` diffs this output
# against the snapshot vendored next to NodePalette's golden, so a change to the
# owner's wire fails NodePalette CI instead of reaching production as a 502.
set -euo pipefail

src="${1:?usage: $0 <NodeVault pkg/catalogrest/server.go>}"

awk '
/^type CertifiedToolItem struct \{/ || /^type ListCertifiedToolsResponse struct \{/ || /^func toCertifiedToolItem\(/ {
	if (blocks++) print ""
	on = 1
}
on { print }
on && /^}/ { on = 0 }
' "$src"
