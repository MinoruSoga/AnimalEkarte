# Pattern References

Quick links to coding pattern implementations. Refer to memory files for detailed examples and explanations.

## Go Patterns

**File**: `.claude/projects/.../memory/backend-patterns.md`

### Core Patterns

1. **Handler → Service → Repository** - Three-layer HTTP API architecture
   - Handler binds JSON, converts to service.Input
   - Service validates and orchestrates business logic
   - Repository manages GORM data access

2. **PATCH (Partial Update)**
   - Use pointer types (*string, *int) to distinguish nil from zero value
   - `buildXxxUpdateFields()` returns map[string]any
   - Pass to `UpdateFields(ctx, id, fields)` to avoid GORM zero-value bug

3. **Error Handling (Sentinel + Wrap)**
   - Define errors in `errors/errors.go` (ErrNotFound, ErrConflict, etc.)
   - Wrap with context: `fmt.Errorf("failed to create: %w", err)`
   - Handle in handler via `RespondError(c, err)` with HTTP status mapping

4. **Concurrent Operations (errgroup)**
   - Use `golang.org/x/sync/errgroup` for parallel operations
   - `g.Go(func() error { ... })` for each operation
   - `g.Wait()` to collect errors

5. **Validation Pattern**
   - Implement `(Input) Validate() error` on service input structs
   - Return []ValidationError with Field + Message
   - Return sentinel error (ErrInvalidInput)

6. **Context Timeout Management**
   - Always accept ctx as first parameter
   - Use `context.WithTimeout()` for deadline management
   - Check `errors.Is(err, context.DeadlineExceeded)` for timeout handling

## React 19 Patterns

**File**: `.claude/projects/.../memory/frontend-patterns.md`

### Core Patterns

1. **useTransition for Forms**
   - Complex forms: use useTransition for pending state
   - `const [isPending, startTransition] = useTransition()`
   - `startTransition(async () => { await saveOwner(formData) })`
   - Prevents double-submission, shows loading state

2. **memo() + useCallback Composition**
   - Break large forms into memoized sections
   - Each section receives useCallback handlers as props
   - Prevents unnecessary re-renders of child components
   - Example: OwnerInfoSection, PetTableRow, MembershipTypeButtons

3. **useDeferredValue for Search**
   - Maintain responsive search input
   - Defer expensive filtering: `const deferred = useDeferredValue(searchTerm)`
   - Use in useMemo to filter results
   - Allows input to update immediately while filtering lags behind

4. **React Query Hooks**
   - `useGetXxx()` for read operations with caching
   - `useCreateXxx()` for mutations with error handling
   - Configure staleTime and gcTime for optimal caching
   - Invalidate cache on mutations

5. **Router Loader Pattern**
   - Load data before component renders: `lazy: async () => { loader: ... }`
   - Use `Promise.all()` for parallel data fetching
   - Direct axios calls (not queryClient.prefetchQuery)
   - Throw errors that trigger error boundary

6. **Type Safety**
   - Never use `any` type
   - Use `unknown` + type guards: `if (typeof x === 'object') { return x as Type }`
   - Derive types from models.ts using Omit/Partial
   - transforms.ts for BackendXxx → Xxx conversion

7. **Error Boundary**
   - `getDerivedStateFromError()` for error capture
   - `componentDidCatch()` for logging
   - Show fallback UI instead of blank screen

## Database Patterns

**File**: `.claude/projects/.../memory/db-schema-reference.md`

### Core Patterns

1. **Multi-Tenant Isolation (clinic_id)**
   - Every table must have clinic_id column
   - ALWAYS include clinic_id in WHERE clause: `WHERE clinic_id = ? AND id = ?`
   - Don't SELECT without clinic_id filter (data leak risk)

2. **Composite Index Strategy**
   - (clinic_id, id) for basic lookups
   - (clinic_id, status) for status filters
   - (clinic_id, created_at DESC) for timeline queries
   - clinic_id always comes first

3. **Soft Delete with Partial Index**
   - deleted_at timestamp column in all tables
   - Partial index: `WHERE deleted_at IS NULL` for active records
   - Include in UNIQUE constraints: `uk_email UNIQUE (clinic_id, email) WHERE deleted_at IS NULL`
   - Filter in queries: `WHERE deleted_at IS NULL`

4. **N+1 Query Prevention**
   - GORM Preload: `.Preload("Pets").Where(...).Find(&owners)`
   - Nested Preload: `.Preload("Pets.MedicalRecords")`
   - Specify columns for performance: `.Select("id", "name")`
   - Use JOIN for complex filtering

5. **Foreign Key with Cascade**
   - Define constraints: `FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE CASCADE`
   - GORM relations: `HasMany`, `BelongsTo`
   - Index foreign key columns for join performance

## Security Patterns

**File**: `.claude/projects/.../memory/security-checklist.md`

### Core Patterns

1. **HTTP Security Headers**
   - X-Content-Type-Options: nosniff
   - X-Frame-Options: DENY (or SAMEORIGIN)
   - Strict-Transport-Security: HSTS
   - Content-Security-Policy: restrict inline scripts

2. **Authentication (JWT + httpOnly Cookies)**
   - Store JWT in httpOnly cookie (not localStorage)
   - Set SameSite=Lax for CSRF protection
   - Include withCredentials: true in axios requests
   - No token in response body

3. **Input Validation**
   - Client: UI validation for UX
   - Server: Always re-validate (never trust client)
   - Check type, format, length, whitelist values
   - Return 400 Bad Request for invalid input

4. **SQL Injection Prevention**
   - Use parameterized queries: `WHERE email = ?`
   - GORM automatically handles this
   - Never construct SQL with string interpolation

5. **XSS Prevention**
   - Never use dangerouslySetInnerHTML
   - DOMPurify for user-generated content
   - React auto-escapes by default
   - CSP headers prevent inline script execution

6. **Password Security**
   - Use bcrypt or argon2 (not plaintext)
   - Salt automatically included
   - Compare time-constant: `bcrypt.CompareHashAndPassword()`
   - Never log passwords

## Performance Patterns

**File**: `.claude/projects/.../memory/security-checklist.md` and `.claude/rules/performance-rules.md`

### Go Performance

1. **Memory Allocation**
   - Pre-allocate slices: `make([]Owner, 0, 100)` with capacity
   - Use bytes.Buffer for string concatenation
   - Profile with pprof: `go tool pprof http://localhost:8080/debug/pprof`

2. **Database Performance**
   - EXPLAIN ANALYZE to verify index usage
   - Avoid Seq Scan (should be Index Scan)
   - Create indexes on frequent WHERE columns
   - Preload related data (prevent N+1)

### React Performance

1. **Bundle Size**
   - Code split with React Router lazy()
   - Dynamic import for heavy modals
   - Check bundle with `pnpm build`
   - Target: < 200KB JS, < 50KB CSS

2. **Render Optimization**
   - memo() prevents unnecessary re-renders
   - useCallback prevents handler recreation
   - useMemo for expensive calculations
   - Profile with React DevTools Profiler

