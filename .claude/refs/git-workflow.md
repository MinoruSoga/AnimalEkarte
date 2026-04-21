---
description: Git workflow standards (branch strategy, commit messages)
alwaysApply: true
---

# Git Workflow Rules

GitHub/Git standard workflow.

## Core Rules

### 1. Branch Strategy

```
production  (production/no direct push)
  ↑ --no-ff merge (release only)
staging     (staging/no direct push, CI/CD → stg.noah-karte.com)
  ↑ PR merge
main        (development/daily work)
```

**Rules:**
- `main`: **Development branch**. Daily work and commits directly here
- `staging`: Staging deployment. Create PR from `main` to `staging` (no direct push)
- `production`: Production (no direct push). Only --no-ff merge from `staging`)

**Daily workflow:**
```bash
# Work on main
git checkout main
git pull origin main
# Make changes/commit
git add ...
git commit -m "fix(xxx): ..."
git push origin main

# To promote to staging
# Create PR: main → staging
gh pr create --base staging --title "..."
```

### 2. Commit Message Format

```
<type>(<scope>): <subject>

<body>

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
```

**type:**
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Refactoring (no behavior change)
- `test`: Test additions/modifications
- `docs`: Documentation
- `ci`: CI/CD configuration
- `chore`: Build, dependency updates etc.
- `perf`: Performance optimization

**Example:**
```
feat(owners): add owner export functionality

- Implement CSV export for owner list
- Add ExportService.ExportOwners() function
- Integrate export button in UI

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
```

### 3. Pull Request (PR) Rules

Base branch is **`staging`** (main → staging).

```markdown
## Summary
- Feature description (1-3 lines)

## Test Plan
- [ ] Verified locally
- [ ] Unit tests pass
- [ ] Integration tests pass

🤖 Generated with Claude Code
```

**Conditions:**
- 1 Approval required for merge to `staging`
- All checks green
- Conflicts resolved

### 4. Pre-commit Checks

```bash
# Lint check
docker compose exec backend golangci-lint run ./...
docker compose exec frontend npm run lint

# Run tests
docker compose exec backend go test ./... -v
docker compose exec frontend npm run test:run
```

### 5. Merge Rules

```bash
# main → staging (via PR on GitHub)
# GitHub merge: Squash or Merge commit

# staging → production (release)
git checkout production
git pull origin production
git merge --no-ff staging -m "Release vX.Y.Z"
git tag vX.Y.Z
git push origin production --tags
```

## Checklist

- [ ] Daily work: commit on `main` branch
- [ ] PR: target `staging` branch
- [ ] Commit: format `type(scope): subject`
- [ ] Message body: explain why, context
- [ ] Co-Authored-By: Claude signature (AI-generated)
- [ ] Tests: all passing
- [ ] Lint: golangci-lint, npm run lint pass
- [ ] PR description: Summary + Test Plan

## Prohibited

- Direct `production` push ❌
- Direct `staging` push ❌ (always via PR from `main`)
- `git push --force` ❌ (shared branches)
- Force push after commit without context ❌
- Large binary/log files in commits ❌
- Secrets (API key, password) in commits ❌
