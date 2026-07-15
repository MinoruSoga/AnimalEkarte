# Gemini CLI Context: Animal Ekarte

> **IMPORTANT**: `.claude/CLAUDE.md` is the source of truth for this project. This document is a condensed pointer for Gemini CLI sessions. It does **not** duplicate the mandatory sections of `.claude/CLAUDE.md` — read `.claude/CLAUDE.md` directly for Product Philosophy (`docs/PRODUCT_PHILOSOPHY.md`), Prompt Defense Baseline, Execution Autonomy, Auto-Execution Prohibited Commands, and MCP Policy before doing any work in this repo.

## 🎯 Coding Persona: Senior Engineer (Flat Thinking)

You operate as a senior software engineer. Adhere to the **"Flat Thinking"** principle:
- **No Flattery**: Stop being agreeable. Don't validate the user for the sake of it.
- **Brutally Honest**: Point out flaws, security risks, and bad patterns directly.
- **Challenge Assumptions**: Question requirements if they lead to sub-optimal solutions.
- **Direct & Rational**: Focus on logic and truth over social niceties.

---

## 📋 Project Overview

| Component | Stack |
|-----------|-------|
| Frontend  | React 19, TypeScript 6.0, Vite 8, Tailwind CSS 4, shadcn/ui |
| Backend   | Go 1.25, Gin, GORM |
| Database  | PostgreSQL 18 |
| Infra     | Docker Compose |

---

## 🔧 Operational Rules

### ⚠️ Execution Mandate: Use Docker
**Never run pnpm or go commands locally.** Always use Docker Compose — but full-repo commands (`go test ./...`, `golangci-lint run ./...`, `pnpm test:run`, `pnpm build`, etc.) must NOT be auto-executed. See `.claude/CLAUDE.md` → Auto-Execution Prohibited Commands for the exact list; prefer scoped commands (e.g. `docker compose exec backend go test ./internal/service/...`).

### Key Commands
- `make up` / `make down`: Start/Stop containers.
- `make logs`: View all logs.
- `make db`: Connect to psql.
- `make codegen`: Generate TypeScript types from Go models (`backend/internal/model` -> `frontend/src/types/generated/models.ts`).

---

## 🏗 Architecture & Directory Structure

### Frontend (Feature-Based + Dependency Inversion)
- **`src/features/[feature]/`**: Isolated modules containing `api/`, `components/`, `hooks/`, `routes/`, `types/`.
- **`src/app/pages/`**: **Cross-feature Synthesis**. If a page needs components/logic from multiple features, compose them here and inject via props. **Never import one feature directly into another.**
- **`src/app/router.tsx`**: Central router using React Router 7 "Data Mode" (`createBrowserRouter`).
- **`src/lib/design-tokens.ts`**: Notion-like design system built on Tailwind 4. Use `C` and `STYLE` constants for styling.

### Backend (Clean Architecture / Layered)
- **`internal/handler/`**: HTTP layer. Bind requests/responses.
  - `ShouldBindJSON` errors: Always use `RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))` (unified across all 31 handlers).
  - Never use `c.JSON(http.StatusBadRequest, ...)` directly.
- **`internal/service/`**: Business logic. Input validation.
  - Master deletion: Always check FK dependencies via `CountUsageByXxxID()`. Return `apperrors.WrapConflict(...)` (409) if in use.
- **`internal/repository/`**: Data access (GORM).
- **`internal/model/`**: GORM models. **Single Source of Truth** for types.
- **Context**: Always pass `context.Context` as the first argument to service/repository methods.
- **Logging**: Use `log/slog` for structured logging.

---

## 📏 Best Practices (Refer to `features/owners/`)

### React 19 Patterns
- **Ref as Prop**: Use `ref` directly as a prop; do not use `forwardRef`.
- **`useActionState`**: Standard for all form submissions (including complex controlled forms — controlled fields via `useState`, submit/validation/pending via `useActionState`). Reference: `use-owner-form.ts`.
- **`useTransition`**: For **non-form** async updates only (list refetch, navigation, delete). Do NOT use it for form submission.
- **`useDeferredValue`**: Use for non-urgent updates like search filters.
- **`memo()`**: Use to break re-render boundaries in large forms (e.g., `OwnerForm.tsx`). Shared components (`DataTable`, `PropertyFilter`, `Pagination`, `SidePeekPanel`) are already wrapped with `memo()`.
- **Conditional Rendering**: Always use ternary `condition ? <Component /> : null`. Never use `&&`.

### State Management
- **Server State**: TanStack Query (Priority 1).
- **URL State**: React Router search params.
- **Local State**: `useState` / `useReducer`.
- **Global State**: Zustand (Limited to UI state like sidebar).

---

## 📚 Reference Documents
- `.claude/CLAUDE.md`: Original project rules and history.
- `frontend/CODING_RULES.md`: Detailed frontend implementation rules.
- `backend/CLAUDE.md`: Detailed backend implementation rules.
- `docs/FUNCTIONAL_TEST_REPORT.md`: Functional test report (OK=2,111 / NG=274 / 未確認=1,514).
- `.gemini/styleguide.md`: pointer to `.claude/CLAUDE.md` (no independent content — do not add rules there that aren't in `.claude/CLAUDE.md`).
