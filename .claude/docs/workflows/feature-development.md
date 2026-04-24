# Feature Development Workflow

Standard workflow for implementing new features while maintaining rule compliance.

## Pre-Development

1. **Read Rules** (5 min)
   - Review relevant rule files from `.claude/rules/`:
     - `go-language.md` (backend work)
     - `typescript-react.md` (frontend work)
     - `database-design.md` (schema changes)
   - Check for forbidden patterns specific to your feature

2. **Reference Patterns** (5 min)
   - Backend: Check `memory/backend-patterns.md` for handler→service→repository pattern
   - Frontend: Check `memory/frontend-patterns.md` for memo/useCallback/useTransition patterns
   - Database: Check `memory/db-schema-reference.md` for multi-tenant (clinic_id) requirements

3. **Load Memory** (2 min)
   - Claude Code auto-loads memory files on session start

## Development Phase

### Backend Feature

```bash
# 1. Create migration (if schema change)
backend/migrations/XXX_feature_name.sql

# 2. Create GORM models
backend/internal/model/feature_model.go

# 3. Create handler (HTTP binding)
backend/internal/handler/feature_handler.go (with *_request.go, *_response.go)

# 4. Create service (business logic)
backend/internal/service/feature_service.go (with validators.go)

# 5. Create repository (GORM layer)
backend/internal/repository/feature_repository.go

# 6. Test
docker compose exec backend go test ./... -v

# 7. Lint (pre-commit hook auto-checks)
docker compose exec backend golangci-lint run ./...
```

**Rules to follow:**
- ✅ Context.Context as first parameter in all functions
- ✅ Sentinel errors in errors/errors.go
- ✅ fmt.Errorf("context: %w", err) for wrapping
- ✅ slog only in service layer
- ✅ PATCH: PointerInput + buildUpdateFields() → map[string]any
- ✅ Multi-tenant: clinic_id in WHERE clause always
- ❌ No Context.Background() in functions
- ❌ No *gin.Context in service layer
- ❌ No direct db.Create().ID usage

### Frontend Feature

```bash
# 1. Create API hooks
frontend/src/features/xxx/api/get-xxx.ts (getXxx() + useGetXxx())
frontend/src/features/xxx/api/create-xxx.ts (createXxx())
frontend/src/features/xxx/api/types.ts (request/response types)
frontend/src/features/xxx/api/transforms.ts (BackendXxx → Xxx)

# 2. Create hooks
frontend/src/features/xxx/hooks/use-xxx-form.ts (useTransition + handleSave)

# 3. Create components
frontend/src/features/xxx/components/XxxCard.tsx (memo + useCallback)
frontend/src/features/xxx/routes/XxxList.tsx (single feature page)

# 4. Create router lazy import
app/router.tsx (inline lazy with Promise.all)

# 5. Test
docker compose exec frontend pnpm test:run

# 6. Lint (pre-commit hook auto-checks)
docker compose exec frontend pnpm lint
```

**Rules to follow:**
- ✅ useTransition for complex forms (not useState + setIsPending)
- ✅ memo() + useCallback for form sections
- ✅ useDeferredValue for search filters
- ✅ Type safety: unknown + type guard (no any)
- ✅ Ternary operator for conditionals (not &&)
- ✅ Direct file imports (not barrel export *)
- ✅ models.ts → transforms.ts → feature types
- ❌ FC type, forwardRef
- ❌ Feature-to-feature imports
- ❌ localStorage for tokens
- ❌ dangerouslySetInnerHTML

### Database Feature

```bash
# 1. Schema planning
docs/ERD.md (update ERD)

# 2. Create migration
backend/migrations/XXX_add_feature_table.sql

# 3. Multi-tenant checklist
- [ ] clinic_id required column
- [ ] Composite index: (clinic_id, id)
- [ ] Soft delete: deleted_at column + partial index WHERE deleted_at IS NULL
- [ ] Foreign keys with ON DELETE CASCADE

# 4. GORM model
backend/internal/model/feature_model.go

# 5. Run migration
docker compose down && docker compose up --build

# 6. Verify schema
docker compose exec db psql -U ekarte_user -d ekarte_db -c "\\d feature_table"
```

**Rules to follow:**
- ✅ clinic_id in every table
- ✅ Composite index (clinic_id, id)
- ✅ Partial index WHERE deleted_at IS NULL
- ✅ Foreign keys required
- ✅ N+1 prevention: GORM Preload or JOIN
- ❌ SELECT * without clinic_id condition
- ❌ No single-column indexes

## Code Review Phase

1. **Self-Review** (10 min)
   - Read the rule files one more time
   - Run linters: `docker compose exec backend golangci-lint run ./...` + `docker compose exec frontend pnpm lint`

2. **Pre-Push Check** (auto)
   - `pre-bash-git-push-reminder.js` warns before git push
   - `post-edit-console-warn.js` detects console.log after edits

3. **Create PR**
   - Title: `feat(feature-name): brief description`
   - Body: Feature description, testing plan, screenshot (if UI change)

## Testing Phase

1. **Unit Tests**
   - Backend: `docker compose exec backend go test ./internal/service -v`
   - Frontend: `docker compose exec frontend pnpm test:run`

2. **Integration Tests** (if applicable)
   - Run full test suite before commit
   - Check test coverage: `go test ./... -cover`

3. **Manual Testing**
   - Backend: Test via Swagger UI or curl
   - Frontend: Test in browser, check console for errors

## Deployment Phase

```bash
# 1. Merge to develop
git checkout develop && git pull
git merge --no-ff feature/xxx

# 2. Run linters
docker compose exec backend golangci-lint run ./...
docker compose exec frontend pnpm lint

# 3. Push
git push origin develop
```

## Common Pitfalls

❌ **Backend**
- Using `context.Background()` instead of passing ctx
- Calling `db.Create()` then accessing `.ID` without error check
- Using `db.Model().Updates()` with direct struct (GORM zero-value bug)
- Writing slog in handler or repository instead of service

❌ **Frontend**
- Using `&&` for conditional render (0/empty string leaks)
- Storing JWT in localStorage (XSS vulnerability)
- Calling `useCallback` without including dependencies in array
- Importing from another feature instead of via app/pages/

❌ **Database**
- Missing clinic_id in WHERE clause
- Creating index on single column (should be (clinic_id, column))
- Forgetting soft delete partial index
- Not adding foreign key constraint

## Helpful Commands

```bash
# Check rule compliance
grep -r "dangerouslySetInnerHTML" frontend/src/  # Should be 0 results
grep -r "any" frontend/src/ | grep -v "string"   # Check for any type

# Run all linters
docker compose exec backend golangci-lint run ./...
docker compose exec frontend pnpm lint

# Run all tests
docker compose exec backend go test ./... -v
docker compose exec frontend pnpm test:run

# View git status
git status
git diff

# Check for secrets before commit
grep -r "SECRET\|API_KEY\|PASSWORD" . --include="*.go" --include="*.ts" --include="*.tsx"
```
