---
description: Error handling standards (Go Sentinel, HTTP Status, user messages)
alwaysApply: true
globs: ["backend/**/*.go", "frontend/src/**/*.{ts,tsx}"]
---

# Error Handling Rules

Standard error handling conventions.

## Core Rules

### 1. Go Error Flow (Repository → Service → Handler)

```go
// errors/errors.go - Define Sentinels
var (
  ErrNotFound       = errors.New("not found")
  ErrConflict       = errors.New("conflict")
  ErrInvalidInput   = errors.New("invalid input")
)

// repository/owner_repository.go - Convert GORM errors
func (r *OwnerRepository) GetByID(ctx context.Context, id uint) (*model.Owner, error) {
  var owner model.Owner
  if err := r.db.WithContext(ctx).First(&owner, id).Error; err != nil {
    // ✅ MANDATE: Use FromGORM in Repository
    return nil, apperrors.FromGORM(err, "owner", fmt.Sprintf("%d", id))
  }
  return &owner, nil
}

// service/owner_service.go - Wrap errors
func (s *OwnerService) GetOwner(ctx context.Context, id uint) (*model.Owner, error) {
  owner, err := s.repo.GetByID(ctx, id)
  if err != nil {
    // ✅ MANDATE: Use Wrap in Service
    return nil, apperrors.Wrap(err, "failed to get owner")
  }
  return owner, nil
}
```

### 2. Frontend Error Handling (handleApiError)

Call `handleApiError` in all `catch` blocks.

```typescript
// ✅ MANDATE: Use handleApiError in all catch blocks
try {
  await api.updateOwner(id, data);
} catch (error) {
  handleApiError(error, "owner update");
}
```

### 3. HTTP Status Mapping

```go
// handler/response.go
func RespondError(c *gin.Context, err error) {
  code := http.StatusInternalServerError
  message := "Internal server error"

  switch {
  case errors.Is(err, apperrors.ErrNotFound):
    code = http.StatusNotFound
    message = "Resource not found"
  case errors.Is(err, apperrors.ErrConflict):
    code = http.StatusConflict
    message = "Resource already exists"
  case errors.Is(err, apperrors.ErrInvalidInput):
    code = http.StatusBadRequest
    message = "Invalid input"
  }

  slog.ErrorContext(c.Request.Context(), "error", "message", message, "error", err)

  c.JSON(code, gin.H{
    "code": errorCode(err),
    "message": message,
    "timestamp": time.Now(),
  })
}
```

### 4. React Error Boundary

```typescript
// components/errors/ErrorBoundary.tsx
export class ErrorBoundary extends React.Component<Props, State> {
  state: State = { hasError: false, error: null };

  static getDerivedStateFromError(error: Error) {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return <ErrorFallback error={this.state.error} />;
    }
    return this.props.children;
  }
}
```

### 5. React Query Error Handling

```typescript
// hooks/use-owners.ts
export function useGetOwners(clinicID: number) {
  return useQuery({
    queryKey: ['owners', clinicID],
    queryFn: async () => {
      const response = await axios.get(`/api/owners?clinic_id=${clinicID}`);
      return response.data;
    },
    retry: 1,
  });
}
```

## Checklist

- [ ] Repository: Use `apperrors.FromGORM(err, "resource", id)`
- [ ] Service: Use `apperrors.Wrap(err, "context")`
- [ ] Frontend: Use `handleApiError(error, "context")` in all `catch` blocks
- [ ] HTTP Status mapping (RespondError)
- [ ] Logging: Structured with slog.ErrorContext
- [ ] React Error Boundary implemented
- [ ] React Query retry configured
- [ ] No console.error in production code
