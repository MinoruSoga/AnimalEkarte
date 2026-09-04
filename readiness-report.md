---
generated: 2026-09-04
level: 4
level_name: Automated
score: 35
total: 36
stack: go-gin-react-typescript-docker
monorepo: true
apps: [backend, frontend]
previous:
  generated: 2026-09-04
  level: 2
  score: 24
caveats:
  - setup skill plugin still not installed; scored from readiness Phase 3 criteria
  - Pillar 6 300-line inventory deferred (project policy is soft 500 / hard 800); pre-commit enforces 800
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

**Project:** AnimalEkarte (Go/Gin + React 19/TypeScript, Docker Compose monorepo-like)
**Level:** 4 / 5 (Automated)
**Score:** 35 / 36 criteria passing
**Delta:** +11 since last report (was Level 2 / 24)

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
| backend | 5/5 | 2/3 | apperrors scoped green; `.claude/rules/tdd.md` present; >300-line files remain under project 800 policy |
| frontend | 5/5 | 2/3 | Full vitest: 529 files / 4195 tests passed (3 skipped); Prettier + ExportDefaultDeclaration eslint rule added |

## Passing

- ✓ Linter configured (golangci-lint + ESLint)
- ✓ Formatter configured (gofmt/goimports + Prettier `frontend/.prettierrc.json`)
- ✓ Lint-on-commit (`.githooks/pre-commit` + `.git/hooks` wrappers via `make setup-hooks`)
- ✓ No default exports rule (`ExportDefaultDeclaration` in `frontend/eslint.config.js`; tooling configs exempt)
- ✓ Test runners (go test / Vitest)
- ✓ Test colocation
- ✓ Coverage thresholds (coverage ratchets + CI)
- ✓ Tests pass (frontend full suite green after TreatmentItemSidePanel flake fix; backend scoped green)
- ✓ TDD rule file (`.claude/rules/tdd.md`, mirrored to `.agents/rules/`)
- ✓ Pre-commit runs lint/format/secrets
- ✓ Pre-push runs tests
- ✓ Secret scanning in pre-commit (gitleaks / docker / pattern fallback)
- ✓ File size limits in pre-commit (800 hard; matches project hooks)
- ✓ Smart test caching (`.test-passed` SHA skip on pre-push)
- ✓ CLAUDE.md / AGENTS.md
- ✓ Commands section
- ✓ Architecture section (`<!-- AUTO:architecture-dirs -->`)
- ✓ Critical Gotchas section
- ✓ Quality gates documented
- ✓ Code review checklist
- ✓ Auto-generated sections (`<!-- AUTO:* -->` + `scripts/generate-agent-doc-sections.sh`)
- ✓ No drift (vitest scoped example + migrations P10 reference fixed)
- ✓ Content quality
- ✓ Agent settings / allow / deny / path-scoped rules / enforcement hierarchy
- ✓ No obvious hardcoded secrets
- ✓ Consistent style
- ✓ `.env.example` / documented commands / lockfiles
- ✓ Agentic workflow + SessionStart validation

## Failing

- ✗ No source files over 300 lines — deferred by choice: project hard limit is 800 (soft 500). Pre-commit enforces 800. Mass split of ~130 backend + ~46 frontend files not done.

## Changes Since Last Report

- ↑ Now passing: formatter (Prettier), no-default-export rule, tests pass, TDD rule, all git-hook criteria, Commands, Critical Gotchas, AUTO markers, drift fixes
- ↓ Regressed: none
