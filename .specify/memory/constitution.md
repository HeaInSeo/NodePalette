# NodePalette Constitution

<!--
  ②-form (D-12), authority revision AR-2026-08-17.1: this file does NOT own
  cross-repo invariants. It consumes the task Authority Snapshot and indexes
  only THIS repo's own enforced constraints. SoT for those is the rules
  themselves (CI / active branch ruleset), not this prose.
-->

## Cross-repo authority — verified revision-pinned repository mirror

Cross-repo platform meaning is selected by the external Authority Router. For
`AR-2026-08-17.1` the scoped authority chain is:

- platform invariants: `Platform Spec Wiki — CURRENT / 1. constitution`
- platform structure / responsibility / call direction:
  `Platform Spec Wiki — CURRENT / 2. architecture`
- repository-portable invariant mirror: `HeaInSeo/NodeVault` —
  `docs/PLATFORM_MASTER_DESIGN.md §4.1–§4.10`
- mirror verification record: `HeaInSeo/NodeVault` —
  `docs/AUTHORITY_MIRROR_VERIFICATION.md`

NodePalette does **not** treat NodeVault §4 as an independent platform
canonical. A task may consume the repository mirror for cross-repo invariant
meaning only when **all** of the following are true:

1. the task `Authority Snapshot` declares `AR-2026-08-17.1`;
2. the NodeVault verification record says `SYNC STATUS: VERIFIED`;
3. the mirror blob SHA matches the blob SHA recorded by that verification record;
4. every scoped/domain/component authority required by the NodePalette task is
   explicitly present in the task `Authority Snapshot`;
5. no semantic conflict with the current Authority Router/upstream authority has
   been detected.

If any condition is missing, `STALE`, `UNKNOWN`, mismatched, or conflicting, stop
with `AUTHORITY_CONFLICT`; do not choose a source by timestamp, filename, or
search rank. **Revision equality alone is not sufficient.**

The current repository verification record covers platform invariants only.
NodePalette work that depends on platform structure/ownership/call-direction or
the catalog/index read contract must carry the exact CURRENT architecture and
relevant scoped/component contract directly in the task `Authority Snapshot`.

Note: NodePalette **reads the NodeVault Index / catalog projection**. That read
contract is **cross-repo** — its authoritative home is the platform/component
contract layer selected by the current Authority Snapshot, not this constitution
or the invariant mirror. It is not declared or forked here.

## Process discipline (repo-operational — owned by this repo)

- **Deterministic gates are the guarantee.** Merge is decided by the active
  default-branch ruleset and its required status checks. LLM/agent review is
  **advisory**: a passing review never merges alone, and a failing required check
  is never overridden.
- **Spec-anchored change**; **test-first** (behavioral changes ship with tests
  that fail before / pass after; CI runs the race variant); **Builder/Critic
  separation** (read-only Critic pass before merge).
- **Local verify (where a local target exists):** `make lint test coverage-check`.
- **Branch protection**: active `main-branch-protection` ruleset on the default
  branch; PR required, review threads resolved, no force-push/deletion. Required
  checks are `Lint`, `Build`, `Unit Tests`, `K8s Data Plane Contract`,
  `Vulnerability Scan (govulncheck)`, `Analyze go`, and `Analyze actions`.

## Repo-local enforced constraints (derived index — NOT canonical)

> Derived index of THIS repo's own gates. Not canonical — SoT is the CI workflow
> plus the active branch ruleset. A check is marked IMPLEMENTED only when the
> active ruleset requires it and a failing result blocks merge.

- **golangci-lint** (IMPLEMENTED — required check `Lint`): lint gate.
- **build/vet** (IMPLEMENTED — required check `Build`): buildability and vet gate.
- **race tests** (IMPLEMENTED — required check `Unit Tests` runs `go test -race -cover`): concurrency safety.
- **coverage** (IMPLEMENTED — enforced inside required `Unit Tests` check via the 70% inline threshold): coverage threshold.
- **K8s Data Plane Contract** (IMPLEMENTED — required check of the same name): manifest/data-plane contract gate.
- **govulncheck** (IMPLEMENTED — required check `Vulnerability Scan (govulncheck)`): vulnerability scan; there is no local `make` target for this check.
- **CodeQL — Go** (IMPLEMENTED — required check `Analyze go`).
- **CodeQL — Actions** (IMPLEMENTED — required check `Analyze actions`).

NodePalette has **no gosec gate** (neither local target nor required CI check).
Actionlint and Module Drift may run in CI but are not listed as required checks by
the active ruleset, so they are not promoted to IMPLEMENTED merge gates here.

## §1.10 — "do not record what you did not observe"

**Authority: CURRENT platform invariant under `AR-2026-08-17.1`. Enforcement in
this repo: PROPOSED where no deterministic local gate exists.** NodePalette has
no deterministic rule that generally enforces this invariant today. The
platform invariant's authority status and this repo's local enforcement status
are separate axes.

## Governance

Cross-repo semantics cannot be amended by editing this constitution, a
repository mirror, or its verification record alone. They follow the task's
current Authority Snapshot; a new platform authority revision must be accepted
before repository mirrors are synchronized and independently re-verified.

**Version**: 2.2.0 | **Ratified**: 2026-08-02 | **Last Amended**: 2026-08-17
