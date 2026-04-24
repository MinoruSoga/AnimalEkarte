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
- `docker compose exec frontend pnpm test:run`
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
- `docker compose exec frontend pnpm install`
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

## 🏗 Architecture (Layer-specific CLAUDE.md)

Layer-specific rules are documented close to the code:

| Directory | Rules |
|-----------|-------|
| `backend/` | Error handling, P1-P18 overview, build commands |
| `backend/internal/handler/` | P5, P6, P7, P12, P14, P15, P18 |
| `backend/internal/service/` | P1, P8, P10, P11, P13, P17 |
| `backend/internal/repository/` | P2, P3, P4 (clinicScope), P9, P16 |
| `backend/migrations/` | Migration naming, clinic_id, CASCADE DELETE禁止 |
| `frontend/` | React 19 patterns, design tokens, build commands |
| `frontend/src/features/` | Feature Indexing, index.ts structure |
| `frontend/src/hooks/` | Shared global hooks — placement rules, React hook rules |

## 📚 refs/ との使い分け

| 種別 | 場所 | 目的 |
|------|------|------|
| 各ディレクトリ CLAUDE.md | コードの隣 | 編集時に常時ロード。簡潔なルールサマリー |
| `.claude/refs/*.md` | `.claude/refs/` | 詳細リファレンス。スキャンプロンプト・完全仕様 |

**原則**: ディレクトリ CLAUDE.md で日常的なルールを把握する。
P1-P18 の完全スキャンや網羅的な確認が必要な時だけ `refs/gin-architecture-compliance.md` を読む。
`refs/` は削除しない — ディレクトリ CLAUDE.md の圧縮サマリーと相補的に機能する。
