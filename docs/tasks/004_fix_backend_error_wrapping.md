# Task: Fix Backend Error Wrapping

## Status
- [x] Extend `internal/errors` with `FromGORM` helper
- [x] Refactor all 40+ repositories to use `apperrors.FromGORM`
- [x] Refactor all 20+ services to use `apperrors.Wrap` and `WrapInvalidInput`
- [x] Verify handler layer consistency (RespondError & File splitting)

## Description
Backend error handling is now fully standardized around `internal/errors` across all layers (Repository, Service, Handler).

## Reference Pattern
### Repository Layer
```go
err := r.db.WithContext(ctx).First(&model, id).Error
if err != nil {
    return nil, apperrors.FromGORM(err, "resource", id)
}
```

### Service Layer
```go
if err := s.repo.Update(ctx, id, fields); err != nil {
    return nil, apperrors.Wrap(err, "failed to update resource")
}
```

### Handler Layer
```go
if err != nil {
    RespondError(c, err)
    return
}
```
