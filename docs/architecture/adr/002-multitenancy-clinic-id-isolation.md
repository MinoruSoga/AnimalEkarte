# ADR-002: マルチテナント設計 — clinic_id 完全隔離

**Status**: Accepted
**Date**: 2026-04-01
**Deciders**: MinoruSoga

## Context

複数クリニックのデータを単一 PostgreSQL インスタンスで管理する。
テナント間のデータ漏洩は医療情報保護の観点から絶対に許容されない。

## Decision

すべてのテナント分割データテーブルに `clinic_id` カラムを追加し、GORM の `clinicScope` を全 repository メソッドに強制適用する。

```go
// clinicScope は全 SELECT / UPDATE / DELETE に WHERE clinic_id = ? を自動付与する
func (r *ownerRepository) FindByID(ctx context.Context, clinicID, ownerID uint) (*model.Owner, error) {
    return r.db.WithContext(ctx).Scopes(r.clinicScope(clinicID)).First(&model.Owner{}, ownerID)
}
```

Repository 層の P4 規約: `clinicScope` なしのデータアクセスはコードレビューで CRITICAL として拒否する。

## Consequences

**ポジティブ:**
- テナント隔離が infrastructure 層で保証されるため、service 層でのミスが影響しない
- golangci-lint + `healthcare-reviewer` agent による静的検証で漏洩パターンを早期発見

**ネガティブ:**
- すべての repository メソッドで clinicID 引数が必要になり、インタフェースが冗長になる
- system_admin 権限での全クリニック横断クエリは clinicScope を意図的にバイパスする必要があり、レビュー負荷が高い

## References

- [backend/internal/repository/CLAUDE.md](../../../backend/internal/repository/CLAUDE.md)
- [.claude/refs/gin-architecture-compliance.md](../../../.claude/refs/gin-architecture-compliance.md) (P4 規約)
