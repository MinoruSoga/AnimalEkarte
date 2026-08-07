# Composition root conventions

> **Purpose**: Keep `backend/cmd/api` a thin composition root (ADR-006). Prevent a second `service.Services` / `repository.Repositories` god aggregator.
> **Scope**: Wiring rules, evaluation of domain `Application`/`Dependencies`, and regression gates. Not a Clean Architecture revival.
> **Related**: [ADR-006](adr/006-backend-domain-package-boundaries.md), [boundary map](be9-2a-boundary-map.md), `backend/cmd/api/composition_*.go`, `route_composition_smoke_test.go`, `composition_root_conventions_lint_test.go`.

## Rules (ARCH-A5)

### A5-1 — New wiring lands in domain composition files

| Do | Don't |
|----|--------|
| Add or extend `composition_<domain>.go` (and domain-local `*_repositories.go` / `*_services.go` when needed) | Dump new domain constructors into `main.go` |
| Prefer domain package constructors (`lstep.NewApplication`, `reservation.New…`) | Re-open central `internal/service` / `internal/repository` |
| Pass **narrow** cross-domain deps (interfaces / typed deps structs) | Pass a kitchen-sink “all services” bag |

`runtimeComposition` in `composition_runtime.go` may **assemble** domain compositions; it must not re-implement domain business logic.

### A5-2 — `main` stays bootstrap-only

`main.go` / `run()` own process lifecycle only:

- config, logger, timezone, DB open/close  
- cipher / storage init  
- hand off to `prepareRuntimeExecution` / server runner  

Do **not** grow `main` with repository fields, service graphs, or route tables. Route registration stays on `runtimeComposition.registerRoutes` (and domain handlers).

### A5-3 — Domain `Application` / `Dependencies` (evaluation)

| Domain | Pattern today | Recommendation |
|--------|----------------|----------------|
| **lstep** | Typed `Dependencies` + `NewApplication` + small exported `Application` surface | **Keep as the template** for domains whose owned graph is mostly self-contained and called from composition with a fixed dep set |
| **medicalrecord** | `composition_medicalrecord*.go` (~600 LOC total) with `medicalRecordCompositionDependencies` | **Do not force** a full `medicalrecord.NewApplication` yet. Cross-domain fan-in (reservation intents, billing discharge, staff, inventory co-tx, audit adapters) is still composition-visible. Prefer further **file split inside cmd/api composition_*** and domain-side constructors for new features |
| **billing / reservation / staff / …** | Domain-scoped composition structs in `cmd/api` | Keep; migrate to domain `NewApplication` only when a domain gains a stable, testable owned graph and composition becomes pure dep injection |

**When to promote a domain to lstep-style `NewApplication`**

1. Owned repositories/services construct from `*gorm.DB` + a small typed deps struct.  
2. Cross-domain needs are interfaces declared on the **consumer** (or a short explicit deps list).  
3. `cmd/api` change for a feature is “fill deps + register routes” without new business branches.  
4. No consumer-0 root facade is introduced to “make DI easier”.

### A5-4 — Route composition regression

Primary monitor: `backend/cmd/api/route_composition_smoke_test.go`.

- Builds full `newRuntimeComposition` and registers routes.  
- Pins **exact** unique Method+Path count (update the constant + changelog comment when routes intentionally change).  
- Samples critical public/protected surfaces.

When adding routes: update OpenAPI drift allowlists / `docs/api.yaml` per existing apicontract gates **and** bump the smoke count with a one-line reason.

### A5-5 — No consumer-0 root facades

Do **not** reintroduce:

- package-global `Services` / `Repositories` aggregator types in `cmd/api` or `internal/`  
- “root facade” types with zero production consumers used only to group constructors  
- thin re-exports that re-create `internal/handler|service|repository`  

Compatibility facades, if ever needed, must be temporary, thin delegates, and have a deletion condition (ADR-006).

## PR checklist (composition / wiring)

- [ ] Diff is mostly `composition_<domain>.go` or domain constructor, not `main.go`  
- [ ] No new god `Services` / `Repositories` type  
- [ ] Cross-domain write appears in [cross-domain orchestration catalog](cross-domain-orchestration-catalog.md) if multi-package  
- [ ] `route_composition_smoke_test` count updated if routes changed  
- [ ] Lint: `go test ./cmd/api/ -run CompositionRootConventions -count=1` (Docker)

## Machine gates

| Gate | Location |
|------|----------|
| Composition conventions lint | `cmd/api/composition_root_conventions_lint_test.go` |
| Route surface pin | `cmd/api/route_composition_smoke_test.go` |
| Retired layer packages | `internal/lintscan/package_boundary_gate_test.go` |
| Domain import allowlist | `internal/lintscan/domain_import_allowlist_lint_test.go` |
