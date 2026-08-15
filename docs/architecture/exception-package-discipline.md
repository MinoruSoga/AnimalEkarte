# Exception package discipline

> **Purpose**: Limit growth of non-domain “exception” packages and keep intentional exceptions from becoming a second architecture (ARCH-A8).
> **Scope**: Policy + machine pins. Does not delete existing cross-cutting packages.
> **Related**: [ADR-006](adr/006-backend-domain-package-boundaries.md), [boundary map](be9-2a-boundary-map.md) §4 keep-tier, [composition-root-conventions](composition-root-conventions.md), [cross-domain-orchestration-catalog](cross-domain-orchestration-catalog.md), `package_boundary_gate_test.go`, `exception_package_discipline_lint_test.go`.

## What counts as an exception package

| Kind | Examples | Default stance |
|------|----------|----------------|
| Cutover / tooling | `csvimport` | **cmd-only** consumers; not a general app write API |
| Narrow capability domain | `identitylink` | Allowed as domain; **no** owner/pet Go import |
| HTTP/shared kernel | `httpapi`, `sharedkernel`, `persistence`, `apperrors` | Prefer existing; do not invent parallel helpers |
| Cross-cutting runtime | `audit`, `middleware`, `authjwt`, `timeutil` | Expand only with real multi-consumer need |
| Forbidden buckets | `common`, `util`, `utils`, `shared`, `helpers` as package names | **Blocked** by package boundary gate (C4) |

## Rules

### A8-1 — `csvimport` stays cutover/tooling

- Production importers of `internal/csvimport` are limited to **cmd tools** (`cmd/csv-import`, `cmd/seed-export`, `cmd/csv-import-failure-rehearsal`, …).
- Do **not** import `csvimport` from `internal/<domain>` or wire it into the online API as a generic multi-table write API.
- Need for app-facing bulk import → design a domain-owned use case (or ADR), not “expose csvimport”.

### A8-2 — No convenient cross-domain write exceptions

- New multi-domain writes go through owner typed intents / ambient tx / documented orchestration ([orchestration catalog](cross-domain-orchestration-catalog.md)).
- A new permanent exception package or dual write path requires **ADR-level** justification (write owner, tx, recovery, deletion condition).

### A8-3 — `identitylink` does not import owner/pet

- Keep identity grouping isolated: no `internal/owner` or `internal/pet` production imports.
- Enforced by domain import allowlist (`identitylink` → `httpapi` only among domains) and explicit discipline lint.

### A8-4 — No bucket packages

- Do not create `internal/common`, `util`, `utils`, `shared`, `helpers` (or domain subpackages with those names).
- Enforced by `TestPackageBoundaryGate` (C4).

### A8-5 — Extract cross-cutting only at 2+ real consumers

- Do not invent `sharedkernel` helpers “for the next feature”.
- Second real production consumer + shared change risk → then name and extract.
- Single-consumer helpers stay in the owner domain (or `cmd`).

## Adding a new top-level `internal/` package

1. Prefer an existing domain or keep-tier package.  
2. If new top-level is required: update `acceptedTopLevelPackages` in `package_boundary_gate_test.go` **and** ADR/boundary map in the same PR.  
3. If it is a domain: add `domainPackages` + `domainImportAllowlist` edges.  
4. If it is an exception: document here why it is not a domain and what consumers are allowed.

## Machine gates

| Rule | Gate |
|------|------|
| A8-1 csvimport cmd-only | `lintscan/exception_package_discipline_lint_test.go` |
| A8-3 identitylink ↛ owner/pet | same + `domain_import_allowlist_lint_test.go` |
| A8-4 no bucket packages | `lintscan/package_boundary_gate_test.go` C4 |
| Unapproved top-level package | package boundary C1 |
| Retired layer resurrection | package boundary C2 |

## PR checklist

- [ ] No new `common`/`util` package  
- [ ] No domain import of `csvimport`  
- [ ] No identitylink → owner/pet import  
- [ ] Cross-domain write catalog/ADR updated if exception claimed  
- [ ] `go test ./internal/lintscan/ -run 'ExceptionPackage|PackageBoundary|DomainImport' -count=1` (Docker)
