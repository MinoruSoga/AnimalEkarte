# Backend — Go / Gin / GORM

## Stack

Go 1.25 / Gin / GORM / PostgreSQL 18

## Architecture

```
Handler → Service → Repository
```

- Handler: bind request, call service, convert response, return JSON
- Service: business logic, call repository, wrap errors
- Repository: GORM queries, convert GORM errors

**パッケージ分割規約（BE8）**: 新規ドメインの repository/service はフラット直下でなく `internal/<layer>/<domain>/` サブパッケージで作成する。詳細 = 各層 CLAUDE.md／規約正本 = `.claude/skills/go-package-conventions/SKILL.md`・計画 = `/BE-refactor.md`（対応後削除）。

## Error Handling (MANDATORY)

```go
// Repository
return nil, apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", id))

// Service
return nil, apperrors.Wrap(err, "failed to find vaccine")

// Handler
RespondError(c, err)  // NEVER c.JSON(http.StatusBadRequest, ...)
```

## Master Data Deletion Pattern

```go
// Service.Delete — always check FK references first
count, _ := s.repo.CountUsageByVaccineID(ctx, id)
if count > 0 {
    return apperrors.WrapConflict(fmt.Errorf("vaccine %d is in use", id))
}
```

## P1–P18 Compliance (MANDATORY)

Full rules in `.claude/refs/gin-architecture-compliance.md`.

| Layer | Patterns |
|-------|---------|
| Handler | P5, P6, P7, P12, P14, P15, P18 — see `internal/handler/CLAUDE.md` |
| Service | P1, P8, P10, P11, P13, P17 — see `internal/service/CLAUDE.md` |
| Repository | P2, P3, P4, P9, P16 — see `internal/repository/CLAUDE.md` |

## Prohibited Commands (must NOT auto-execute)

```bash
docker compose exec backend go test ./...
docker compose exec backend golangci-lint run ./...
docker compose exec backend gofmt -w ./...
docker compose exec backend go mod download
```
