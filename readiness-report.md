---
generated: 2026-09-04
level: 4
level_name: Automated
score: 35
total: 36
stack: go-gin-react-typescript-docker
monorepo: true
apps: [backend, frontend]
caveats:
  - setup skill skipped by request; scored from readiness Phase 3 pillars only
  - install not re-run; lockfiles + Docker/Make path treated as reproducible
pillars:
  style-validation: { pass: 4, total: 4 }
  testing: { pass: 5, total: 5 }
  git-hooks: { pass: 5, total: 5 }
  documentation: { pass: 9, total: 9 }
  agent-config: { pass: 5, total: 5 }
  code-quality: { pass: 2, total: 3 }
  dev-environment: { pass: 3, total: 3 }
  agentic-workflow: { pass: 2, total: 2 }
per_app:
  backend:
    testing: { pass: 5, total: 5 }
    code-quality: { pass: 2, total: 3 }
  frontend:
    testing: { pass: 5, total: 5 }
    code-quality: { pass: 2, total: 3 }
---

# Harness Readiness Report

**Project:** AnimalEkarte (Go/Gin + React 19/TypeScript, Docker Compose)
**Level:** 4 / 5 (Automated)
**Score:** 35 / 36 criteria passing
**Delta:** 0 since last report (still Level 4 / 35)

## Pillar Scores

Style & Validation    ██████ 4/4
Testing               ██████ 5/5
Git Hooks             ██████ 5/5
Documentation         ██████ 9/9
Agent Configuration   ██████ 5/5
Code Quality          ████░░ 2/3
Dev Environment       ██████ 3/3
Agentic Workflow      ██████ 2/2

## Monorepo Breakdown

| Package | Testing | Code Quality | Notes |
|---|---|---|---|
| backend | 5/5 | 2/3 | `go test ./internal/apperrors/...` green; 130 non-test `.go` files >300 lines |
| frontend | 5/5 | 2/3 | TreatmentItemSidePanel vitest green; 46 non-generated `src` files >300 lines |

## Passing

- ✓ Linter (`backend/.golangci.yml` + `make lint`; `frontend/eslint.config.js` + `make lint-front`)
- ✓ Formatter (golangci gofmt/goimports; `frontend/.prettierrc.json` + `format`/`format:check`)
- ✓ Lint-on-commit (`.githooks/pre-commit` via `make setup-hooks`; wrappers at `.git/hooks/pre-commit`)
- ✓ No default exports (`ExportDefaultDeclaration` in eslint; vite/playwright/`*.d.ts` exempt)
- ✓ Test runners (`make test` / Vitest `make test-front`)
- ✓ Test colocation (`*_test.go`, `*.test.ts(x)`)
- ✓ Coverage thresholds (`backend/.coverage-baseline`, `frontend/.coverage-baseline`)
- ✓ Tests pass (scoped backend + frontend samples green this run)
- ✓ TDD rule (`.claude/rules/tdd.md` with `paths:` frontmatter)
- ✓ Pre-commit covers lint/format/secrets (`.git/hooks/pre-commit` → `.githooks/pre-commit`)
- ✓ Pre-push runs tests (`.git/hooks/pre-push` → `.githooks/pre-push`)
- ✓ Secret scanning wired (`gitleaks protect --staged` / docker / pattern fallback)
- ✓ File size limits wired (`.githooks/lib/check-file-sizes.sh`, hard 800)
- ✓ Smart test caching (`.test-passed` SHA skip)
- ✓ CLAUDE.md / AGENTS.md
- ✓ Commands section (`.claude/CLAUDE.md`)
- ✓ Architecture section + `<!-- AUTO:architecture-dirs -->`
- ✓ Critical Gotchas section
- ✓ Quality gates (800-line hard / soft 500 documented + hooks)
- ✓ Code review checklist (`CODING_RULES` / review refs / review command)
- ✓ Auto-generated sections (`scripts/generate-agent-doc-sections.sh`)
- ✓ No drift (architecture dirs on disk; vitest scoped command corrected)
- ✓ Content quality (project-specific, actionable)
- ✓ `.claude/settings.json` with allow/deny; path-scoped rules; enforcement hierarchy
- ✓ No obvious hardcoded secrets (pattern scan clean on backend + `frontend/src`)
- ✓ Consistent style (domain packages + ESLint mechanical guards)
- ✓ `.env.example`; documented Make/Docker commands; lockfiles present
- ✓ Agentic workflow (skills/commands/agents) + SessionStart (`session-init.sh`)

## Failing

- ✗ No source files over 300 lines — backend ~130 non-test `.go` and frontend ~46 non-generated `src` files exceed 300. Project policy intentionally uses soft 500 / hard 800 (pre-commit enforces 800).

## Changes Since Last Report

- ↑ Now passing: none (already at 35/36)
- ↓ Regressed: none
