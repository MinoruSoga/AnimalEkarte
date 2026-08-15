# ADR-002: マルチテナント設計 — clinic_id 完全隔離

**Status**: Accepted
**Date**: 2026-04-01
**Deciders**: MinoruSoga

## Context

複数クリニックのデータを単一 PostgreSQL インスタンスで管理する。
テナント間のデータ漏洩は医療情報保護の観点から絶対に許容されない。

## Decision

すべてのテナント分割データテーブルに `clinic_id` を持たせ、認証済みidentityから決定したclinic scopeを、すべてのread/write/delete pathへ適用する。packageやlayerの名称には依存しない。

```go
// clinicScope は現在のGORM実装でWHERE clinic_id = ?を付与するhelperの一例
func (r *ownerRepository) FindByID(ctx context.Context, clinicID, ownerID uint) (*model.Owner, error) {
    return r.db.WithContext(ctx).Scopes(r.clinicScope(clinicID)).First(&model.Owner{}, ownerID)
}
```

GORM helperの使用有無だけで安全と判定しない。raw SQL、join、preload、count、bulk処理、background job、request由来FKのownershipを含め、全data pathをruntime isolation testで検証する。cross-tenant accessを可能にする変更はCRITICALとして拒否する。

## Consequences

**ポジティブ:**
- schema、query predicate、application ownership check、runtime testの多層防御で漏洩リスクを下げられる
- golangci-lint + `healthcare-reviewer` agent による静的検証で漏洩パターンを早期発見

**ネガティブ:**
- clinic-scoped operationは認証済みclinic identityを明示的に受け渡す必要がある
- system_admin等の横断queryは通常scopeと別の明示的なauthorization・audit・testが必要になる

## References

- [Backend Application Invariants](../../../.claude/refs/backend-application-invariants.md)
- [Go/Gin Backend Review](../../../.claude/refs/go-gin-backend-review.md)
