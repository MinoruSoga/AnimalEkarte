# BUG-413: chief_complaint_service の削除前依存チェックに clinic_id フィルタ欠落

## 概要

`chief_complaint_service.go` の `Delete` 内で `inquiryRepo.CountByChiefComplaintTypeID()` を
呼び出す際に `clinicID` を渡していない。また repository 側も WHERE 句に `clinic_id` を
含まないため、**別クリニックの問診レコード数がカウントされてしまう**。

マルチテナント環境では、自クリニックで未使用のマスタを削除しようとした際に
他クリニックのデータが「使用中」と判定され、削除できないという誤動作が発生する可能性がある。
逆に共有データがある場合の参照整合性が保証されない。

## 問題箇所

```go
// chief_complaint_service.go:118
count, err := s.inquiryRepo.CountByChiefComplaintTypeID(ctx, id)
// ↑ clinicID を渡していない（第2引数に clinicID が必要）
```

```go
// inquiry_repository.go:98-103
func (r *inquiryRepository) CountByChiefComplaintTypeID(ctx context.Context, categoryID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.Inquiry{}).
        Where("chief_complaint_type_id = ?", categoryID).  // ← clinic_id フィルタなし
        Count(&count).Error
    return count, apperrors.FromGORM(err, "inquiry", fmt.Sprintf("%d", categoryID))
}
```

## 修正方針

### 1. `inquiry_repository.go` — メソッドシグネチャに clinicID を追加

```go
// Interface 変更
CountByChiefComplaintTypeID(ctx context.Context, clinicID, categoryID uint64) (int64, error)

// 実装変更
func (r *inquiryRepository) CountByChiefComplaintTypeID(ctx context.Context, clinicID, categoryID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.Inquiry{}).
        Where("clinic_id = ? AND chief_complaint_type_id = ?", clinicID, categoryID).  // 追加
        Count(&count).Error
    return count, apperrors.FromGORM(err, "inquiry", fmt.Sprintf("%d", categoryID))
}
```

### 2. `chief_complaint_service.go` — clinicID を渡す

```go
count, err := s.inquiryRepo.CountByChiefComplaintTypeID(ctx, clinicID, id)
```

### 3. InquiryRepository interface も同様に更新

## 影響ファイル

- `backend/internal/service/chief_complaint_service.go` — 行 118
- `backend/internal/repository/inquiry_repository.go` — 行 98-103（シグネチャ + WHERE句）
- `backend/internal/repository/interfaces.go`（または同等のファイル）— InquiryRepository interface

## 優先度

**High** — マルチテナントデータ分離の欠陥。別クリニックのデータが削除判定に影響する。

## 関連規約

- `.claude/rules/database-design.md` — マルチテナント設計（clinic_id 必須）
- `.claude/CLAUDE.md` — 「マスタ削除の FK 依存チェック (MANDATORY)」
