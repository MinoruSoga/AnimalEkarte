# Troubleshooting Guide

## Common Issues and Solutions

### Backend (Go) Issues

#### Context.Background() in function
**Error**: `context.Background() must not be called in service/handler functions`

**Fix**:
```go
// ❌ Wrong
func (s *Service) Create() (*Model, error) {
  ctx := context.Background()
  return s.repo.Create(ctx, ...)
}

// ✅ Correct
func (s *Service) Create(ctx context.Context, ...) (*Model, error) {
  return s.repo.Create(ctx, ...)
}
```

#### GORM PATCH zeros out fields
**Error**: `PATCH request sets fields to zero instead of updating them`

**Fix**:
```go
// ❌ Wrong
input := UpdateOwnerInput{Name: "", Email: ""}  // Zero values!
s.repo.UpdateFields(ctx, id, input)

// ✅ Correct
type UpdateOwnerInput struct {
  Name  *string `json:"name"`
  Email *string `json:"email"`
}
input := UpdateOwnerInput{Name: ptr("new name")}
fields := buildUpdateFields(input)  // Only includes non-nil fields
s.repo.UpdateFields(ctx, id, fields)
```

#### N+1 Query Problem
**Error**: `Database is slow, lots of queries in logs`

**Fix**:
```go
// ❌ Wrong - triggers N queries
var owners []Owner
db.Where("clinic_id = ?", clinicID).Find(&owners)
for _, owner := range owners {
  db.Find(&owner.Pets, "owner_id = ?", owner.ID)  // N additional queries!
}

// ✅ Correct - single query
var owners []Owner
db.Preload("Pets").Where("clinic_id = ?", clinicID).Find(&owners)
```

#### slog in handler/repository
**Error**: `Structured logging not found at service layer`

**Fix**:
```go
// ❌ Wrong
func (h *OwnerHandler) Create(c *gin.Context) {
  slog.Info("creating owner", "name", req.Name)
}

// ✅ Correct
func (h *OwnerHandler) Create(c *gin.Context) {
  owner, err := h.service.Create(c.Request.Context(), input)
}

func (s *OwnerService) Create(ctx context.Context, input CreateOwnerInput) (*Owner, error) {
  slog.InfoContext(ctx, "creating owner", "name", input.Name)
}
```

#### Foreign key constraint violation
**Error**: `INSERT violates foreign key constraint`

**Fix**:
- Ensure parent record exists before inserting child
- Check clinic_id references valid clinic record
- Verify id types match (both uint64, etc.)

### Frontend (React) Issues

#### Conditional render shows 0 or empty string
**Error**: `Page shows "0" or empty component on screen`

**Fix**:
```typescript
// ❌ Wrong
{items.length && <List items={items} />}  // Shows 0 when items.length = 0
{isLoaded && <Content />}                  // Shows false/true sometimes

// ✅ Correct
{items.length > 0 ? <List items={items} /> : null}
{isLoaded ? <Content /> : null}
```

#### memo() not preventing re-renders
**Error**: `Component still re-renders even with memo()`

**Fix**:
```typescript
// ❌ Wrong
const handleChange = (value) => setData(value);  // New function each render!
<OwnerSection onChange={handleChange} />

// ✅ Correct
const handleChange = useCallback((value) => setData(value), [setData]);
const OwnerSection = memo(function OwnerSection({ onChange }: Props) { ... });
<OwnerSection onChange={handleChange} />
```

#### useTransition not working with complex form
**Error**: `Form still freezes or allows double-submission`

**Fix**:
```typescript
// ❌ Wrong
const [isPending, setIsPending] = useState(false);
const handleSave = async () => {
  setIsPending(true);
  try { await save(); } finally { setIsPending(false); }  // Can forget finally!
}

// ✅ Correct
const [isPending, startTransition] = useTransition();
const handleSave = () => {
  startTransition(async () => {
    await save();  // Automatically manages isPending
  });
}
```

#### TypeScript any type
**Error**: `Type checking disabled, missed bugs at runtime`

**Fix**:
```typescript
// ❌ Wrong
const handleData = (data: any) => {
  return data.name.toUpperCase();  // Could crash if data is null!
}

// ✅ Correct
const handleData = (data: unknown) => {
  if (data && typeof data === 'object' && 'name' in data) {
    return (data as { name: string }).name.toUpperCase();
  }
  return '';
}
```

#### localStorage token vulnerability
**Error**: `JWT token visible in DevTools, stolen by XSS attack`

**Fix**:
```typescript
// ❌ Wrong
localStorage.setItem('token', jwt);
axios.defaults.headers.common['Authorization'] = jwt;

// ✅ Correct
// Server sets httpOnly cookie automatically
// Frontend sends with: axios({ withCredentials: true })
const axiosInstance = axios.create({
  withCredentials: true  // Includes httpOnly cookies
});
```

#### Feature imports crossing boundary
**Error**: `features/owners imports features/pets components`

**Fix**:
```typescript
// ❌ Wrong
// features/owners/routes/OwnersList.tsx
import { PetCard } from '@/features/pets/components/PetCard';

// ✅ Correct
// app/pages/OwnerWithPetsPage.tsx
import { OwnersList } from '@/features/owners/routes/OwnersList';
import { PetCard } from '@/features/pets/components/PetCard';
// Pass PetCard as prop to OwnersList (dependency inversion)
<OwnersList petComponent={PetCard} />
```

### Database Issues

#### clinic_id missing from WHERE clause
**Error**: `Data leak: user sees data from other clinics`

**Fix**:
```go
// ❌ Wrong - security issue!
var owners []Owner
db.Where("id = ?", ownerID).Find(&owners)

// ✅ Correct
var owners []Owner
db.Where("clinic_id = ?", clinicID).
  Where("id = ?", ownerID).
  Find(&owners)
```

#### Index not being used
**Error**: `Query is slow, logs show Seq Scan instead of Index Scan`

**Fix**:
```bash
# Check current query plan
docker compose exec db psql -U ekarte_user -d ekarte_db
EXPLAIN ANALYZE SELECT * FROM owners WHERE clinic_id = 1 AND id = 100;

# If Seq Scan, create index
CREATE INDEX idx_owners_clinic_id ON owners(clinic_id, id);
```

#### Soft delete returning deleted records
**Error**: `SELECT returns deleted records (deleted_at is not null)`

**Fix**:
```go
// ❌ Wrong - includes soft-deleted records
var owners []Owner
db.Where("clinic_id = ?", clinicID).Find(&owners)

// ✅ Correct
var owners []Owner
db.Where("clinic_id = ?", clinicID).
  Where("deleted_at IS NULL").
  Find(&owners)

// ✅ Or use GORM scope
func ActiveRecords(db *gorm.DB) *gorm.DB {
  return db.Where("deleted_at IS NULL")
}
var owners []Owner
db.Scopes(ActiveRecords).Where("clinic_id = ?", clinicID).Find(&owners)
```

### Docker Issues

#### npm/go commands fail locally
**Error**: `bash: npm: command not found` (when running `npm install` locally)

**Fix**:
```bash
# ❌ Wrong
npm install
npm run build
go test ./...

# ✅ Correct
docker compose exec frontend npm install
docker compose exec frontend npm run build
docker compose exec backend go test ./...
```

#### Docker image build cache not working
**Error**: `Every build takes 5 minutes, code changes trigger full rebuild`

**Fix**: Check Dockerfile layer ordering in `.claude/rules/docker-rules.md`
```dockerfile
# ❌ Wrong - code changes trigger dependency re-download
FROM golang:1.25
COPY . .
RUN go mod download

# ✅ Correct - dependencies cached separately
FROM golang:1.25
COPY go.mod go.sum ./
RUN go mod download
COPY . .
```

### Git/Commit Issues

#### Pre-commit hook blocks commit
**Error**: `git commit fails with lint/secrets errors`

**Fix**:
1. Read error message carefully
2. Run linter locally: `docker compose exec backend golangci-lint run ./...`
3. Fix violations
4. Re-commit: `git commit -m "message"`

#### Secrets committed to git
**Error**: `API keys visible in git history`

**Fix**:
1. Move secrets to `.env.local` (add to `.gitignore`)
2. Use environment variables in code
3. Use `git-secrets` or `.git/hooks/pre-commit.sh` to prevent future commits
4. Rotate compromised secrets immediately

## Getting Help

1. **Check the rules**: `.claude/rules/` - detailed explanations of violations
2. **Reference patterns**: `.claude/docs/patterns/README.md` - implementation examples
3. **Read memories**: `.claude/projects/.../memory/` - 6 knowledge files
4. **Check recent commits**: `git log --oneline -10` - see similar implementations
5. **Run linters**: `docker compose exec backend golangci-lint run ./...`

## When All Else Fails

1. Save your changes: `git stash`
2. Check git status: `git status`
3. Look at diffs: `git diff HEAD~1`
4. Rebuild containers: `make clean && make up`
5. Read error messages carefully (not just the first line)
6. Try with verbose flags: `golangci-lint run ./... -v`
