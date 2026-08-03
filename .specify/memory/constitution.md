# NodePalette Constitution

<!--
  ②-form (D-12): this file does NOT own cross-repo invariants. It references the
  platform canonical constitution and indexes only THIS repo's own enforced
  constraints. SoT for those is the rules themselves (Makefile gates / CI), not
  this prose.
-->

## Cross-repo invariants live in the platform canonical (NodeVault §4)

Cross-repo invariants — reproducibility, `casHash`, `stableRef`, the artifact
dual-axis (`lifecycle_phase` / `integrity_health`), the sori boundary, and the
image-build / ResolveRecipe rules — are owned solely by the platform canonical:
**`github.com/HeaInSeo/NodeVault` — `docs/PLATFORM_MASTER_DESIGN.md` §4**
(immutable architecture decisions). This document does not restate or fork
them; on any conflict, §4 wins.

Note: NodePalette **reads the NodeVault Index / catalog projection**. That read
contract is **cross-repo** — its canonical home is the platform contract layer
(platform-canonical-owned per NodeVault §4), not this constitution. It is not declared or forked here.

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

**Status: PROPOSED (not enforced in this repo).** §1.10 is a cross-repo
rule (not yet part of NodeVault §4); NodePalette has **no deterministic rule** enforcing
it today. Marked PROPOSED, not IMPLEMENTED, until such a gate exists.

**Version**: 1.0.0 | **Ratified**: 2026-08-02 | **Last Amended**: 2026-08-02
