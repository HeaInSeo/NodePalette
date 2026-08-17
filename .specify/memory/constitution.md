# NodePalette Constitution

<!--
  ②-form (D-12), authority revision AR-2026-08-17.1: this file does NOT own
  cross-repo invariants. It consumes the task Authority Snapshot and indexes
  only THIS repo's own enforced constraints. SoT for those is the rules
  themselves (Makefile gates / CI), not this prose.
-->

## Cross-repo authority — revision-pinned repository mirror

Cross-repo platform meaning is selected by the external Authority Router. For
`AR-2026-08-17.1` the scoped authority chain is:

- platform invariants: `Platform Spec Wiki — CURRENT / 1. constitution`
- platform structure / responsibility / call direction:
  `Platform Spec Wiki — CURRENT / 2. architecture`
- repository-portable mirror: `HeaInSeo/NodeVault` —
  `docs/PLATFORM_MASTER_DESIGN.md` at the same authority revision

NodePalette does **not** treat NodeVault §4 as an independent platform
canonical. A task may consume that repository mirror only when its `Authority
Snapshot` declares `AR-2026-08-17.1`. Missing/mismatched/conflicting snapshots
must stop with `AUTHORITY_CONFLICT`; do not choose a source by timestamp,
filename, or search rank.

Note: NodePalette **reads the NodeVault Index / catalog projection**. That read
contract is **cross-repo** — its authoritative home is the platform contract
layer selected by the current Authority Snapshot, not this constitution. It is
not declared or forked here.

## Process discipline (repo-operational — owned by this repo)

- **Deterministic gates are the guarantee.** Merge is decided by deterministic
  checks (tests, coverage, golangci-lint). LLM/agent review is **advisory**: a
  passing review never merges alone, a failing gate is never overridden.
- **Spec-anchored change**; **test-first** (behavioral changes ship with tests
  that fail before / pass after; CI runs the race variant); **Builder/Critic
  separation** (read-only Critic pass before merge).
- **Local verify (before a PR):** `make lint test coverage-check`.
- **Branch protection**: `main` lands via PR with required checks; no direct
  pushes.

## Repo-local enforced constraints (derived index — NOT canonical)

> Derived index of THIS repo's own gates. Not canonical — SoT is the gate itself.

- **golangci-lint** (IMPLEMENTED — required check "Lint" via `golangci-lint-action` with `--config=.golangci.yml`; not `make`): lint gate.
- **race tests** (IMPLEMENTED — required check "Unit Tests" runs `go test -race -cover` inline; not `make`): concurrency safety.
- **coverage** (IMPLEMENTED — `ci.yml` inline 70% gate via `go tool cover`): coverage threshold enforced in CI
  (70%).

CI-only (no `make` target): **govulncheck** (vulnerability scan, `ci.yml`) and
**CodeQL** (`codeql.yml`) run as GitHub Actions gates but have no local `make`
target. NodePalette has **no gosec gate** (neither `make` nor CI).

## §1.10 — "do not record what you did not observe"

**Authority: CURRENT platform invariant under `AR-2026-08-17.1`. Enforcement in
this repo: PROPOSED where no deterministic local gate exists.** NodePalette has
no deterministic rule that generally enforces this invariant today. The
platform invariant's authority status and this repo's local enforcement status
are separate axes.

**Version**: 2.0.0 | **Ratified**: 2026-08-02 | **Last Amended**: 2026-08-17
