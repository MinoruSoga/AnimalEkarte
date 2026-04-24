# Animal Ekarte - Veterinary Hospital Electronic Medical Record System

## 🎯 Engineering Mindset

**As a senior engineer, maintain these principles:**
- Flat Thinking: Remove social pleasantries. Direct feedback based on facts and logic
- Type Safety First: Prohibit `any` in both Go and TypeScript
- Architecture Adherence: Maintain handler → service → repository lightweight layering

---

## 📋 Project Overview

| Item | Details |
|------|---------|
| Frontend | React 19 / TypeScript 5.7 / Vite 6 / Tailwind CSS 4 / shadcn/ui |
| Backend | Go 1.25 / Gin / GORM |
| Database | PostgreSQL 18 (Docker: postgres:18-alpine) |

## 🔧 Mandatory Operational Rules

- **Docker Required**: npm/go commands prohibited locally. Use `docker compose exec frontend/backend` only
- **Branches**: Daily work on `main`. `main` → `staging` via PR. No direct `production` push

## 🚫 Auto-Execution Prohibited Commands

The following commands **must NOT be auto-executed by Claude Code**.
If execution is needed, inform the user with the command and have them run it manually.

### Build / Test / Quality Checks (large output)
- `docker compose exec backend go test ./...`
- `docker compose exec backend golangci-lint run ./...`
- `docker compose exec backend gofmt -w ./...`
- `docker compose exec frontend pnpm lint`
- `docker compose exec frontend ppnpm test:run`
- `docker compose exec frontend pnpm build`
- `docker compose exec frontend pnpm type-check`
- `make codegen`

### Docker Startup / Shutdown (large logs)
- `docker compose up` / `docker compose down`
- `docker compose restart`
- `docker compose logs` (streaming)
- `docker system prune`

### DB / Migration (high side effects)
- `make db` / DB reset commands
- `docker compose exec db psql ...` (direct SQL execution)

### Dependency Installation (verbose and slow)
- `docker compose exec frontend ppnpm install`
- `docker compose exec backend go mod download`

**Example response:**

```
Changes complete. Run this manually to verify:
$ docker compose exec backend go test ./internal/service/...
```

---

## ⚡ Context Loading Rules (Critical)

**Before starting work:**

1. Read the user's instructions
2. Determine work type
3. Read **only relevant files** from the table below (no full reads)
4. **Decide whether to enable `/think`** (see criteria below)

### `/think` Enablement Criteria

| Enable (complex, high cost) | Skip (simple, low cost) |
|--------------------------|------------------------|
| Architecture design, large refactors | File reading, searching, investigation |
| Mysterious bug investigation, debugging | Simple typo fixes, comment updates |
| Security design, vulnerability analysis | Known pattern implementation |
| Multi-layer design decisions | Answering questions, explanations |
| Technical selection with multiple trade-offs | Single file minor modifications |

**Principle**: When uncertain, **SKIP**. Extended Thinking has 3-5x token overhead. Enable only for clearly complex problems.

### Reference Files (`.claude/refs/`)

| File | Read When |
|------|-----------|
| `go-language.md` | Go code implementation/review |
| `gin-architecture-compliance.md` | Gin/GORM P1-P18 compliance check (handler/service/repository) |
| `error-handling.md` | Error handling implementation (Go/TS both) |
| `typescript-react.md` | Frontend implementation/review |
| `testing.md` | Test implementation |
| `api.md` | API design, endpoint additions |
| `naming-conventions.md` | DB/API/Go naming verification |
| `database-design.md` | DB design, migrations |
| `git-workflow.md` | Git operations, PR creation |
| `code-style.md` | Code convention verification |
| `performance-rules.md` | Performance optimization |
| `docker-rules.md` | Docker/infrastructure changes |
| `accessibility-rules.md` | Frontend UI implementation |
| `security.md` | Security-related changes |

---

## 🏗 Architecture (MANDATORY)

### Error Handling
- Repository: `apperrors.FromGORM(err, "resource", id)`
- Service: `apperrors.Wrap(err, "message")`
- Handler: `RespondError(c, err)` (direct `c.JSON(http.StatusBadRequest, ...)` prohibited)

### Master Data Deletion
Dependency check before deletion required. If references exist → `apperrors.WrapConflict(...)` returns 409.

### Backend Compliance (P1–P18)
18 fixed patterns enforced across all layers. See `gin-architecture-compliance.md`.
- **Handler**: P7(toXxxResponse), P12(ShouldBindJSON), P14(no direct repo), P15(201+Location), P18(toXxxResponse naming)
- **Service**: P1(FindByID first), P8(Wrap), P10(FK check), P11(slog.ErrorContext), P13(def order), P17(Input naming)
- **Repository**: P2(IS NULL count), P3(IS NULL preload), P4(clinicScope), P9(FromGORM), P16(method naming)
- **Routes**: P5(RequirePermission), P6("delete" on DELETE)

### Frontend (MANDATORY Patterns)
- Forms: `useActionState` + `<form action={formAction}>` + `SubmitButton`
- Conditional Render: `? (...) : null` (NOT `&&`)
- Feature Imports: Always via `index.ts` (NO deep imports)
- Styling: Use `C`, `STYLE` constants (NO hex color direct specification)
