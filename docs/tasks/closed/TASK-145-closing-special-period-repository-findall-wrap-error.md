# TASK-145: [Low] closing_special_period_repository.go — FindAll で apperrors.Wrap を使用（FromGORM に統一すべき）

## 優先度
**Low**（エラーハンドリング規約の統一）

## 対象ファイル
- `backend/internal/repository/closing_special_period_repository.go`

## 問題

Repository レイヤーでの GORM エラーハンドリング規約:
> **Repository**: GORM エラーは必ず `apperrors.FromGORM(err, "resource", id)` で変換。

`FindAll`（L33〜43）は一覧取得で特定 ID がないため `FromGORM` の id 引数に空文字 `""` を
渡すパターン（例: `accounting_repository.go` の `FindAll` で `apperrors.FromGORM(err, "billing", "")`）が正しい。

現状は `apperrors.Wrap` が使われており、GORM エラーの種別情報（`ErrRecordNotFound` 等）が
`FromGORM` で適切にマッピングされずに握りつぶされる可能性がある。

## 現状コード（closing_special_period_repository.go L33〜43）

```go
func (r *closingSpecialPeriodRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
    var periods []model.ClosingSpecialPeriod
    err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Order("start_date ASC").
        Find(&periods).Error
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to find closing special periods")  // ← 違反
    }
    return periods, nil
}
```

## 修正後コード

```go
func (r *closingSpecialPeriodRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.ClosingSpecialPeriod, error) {
    var periods []model.ClosingSpecialPeriod
    err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Order("start_date ASC").
        Find(&periods).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "closing_special_period", "")  // ← FromGORM に統一
    }
    return periods, nil
}
```

## 補足

`Find` は `ErrRecordNotFound` を返さない（0件の場合は空スライス）ため、
`errors.Is(err, gorm.ErrRecordNotFound)` の特別ハンドリングは不要。
`FromGORM` に変更するだけでよい。
