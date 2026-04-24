# CODE-QUALITY-003: マスタ削除前 FK 依存チェック漏れ修正

## 概要

複数のマスタエンティティで、削除前の FK 依存チェック（`CountUsageByXxxID` → `apperrors.WrapConflict`）が未実装。  
プロジェクト規約（`.claude/CLAUDE.md` 1b節）に明示された必須実装が欠落している。

## 優先度

HIGH

## 影響ファイル

| エンティティ | Repository | Service | 状態 |
|---|---|---|---|
| `chief_complaint_type` | `chief_complaint_repository.go` | `chief_complaint_service.go` | CountUsageBy* なし |
| `inquiry_template` | `inquiry_template_repository.go` | `inquiry_template_service.go` | CountUsageBy* なし |
| `reservation_type` | `reservation_type_repository.go` | `reservation_type_service.go` | `ExistsByReservationTypeID`（bool型）で代替実装 |
| `cage` | `cage_repository.go` | `cage_service.go` | `ExistsByCageID`（bool型、clinicID なし）で代替実装 |

---

## 問題一覧

### 1. `chief_complaint_type` — 削除前依存チェック未実装

主訴タイプが `medical_records` 等に参照されている状態で削除しても、エラーなく実行される。  
DB 側で FK 違反が発生するか、参照先が孤立レコードになる。

**修正方針**:

```go
// chief_complaint_repository.go に追加
CountUsageByChiefComplaintTypeID(ctx context.Context, clinicID, id uint64) (int64, error)

// chief_complaint_service.go の Delete に追加
count, err := s.repo.CountUsageByChiefComplaintTypeID(ctx, clinicID, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check chief complaint type usage")
}
if count > 0 {
    return apperrors.WrapConflict("この主訴タイプは使用中のため削除できません")
}
```

---

### 2. `inquiry_template` — 削除前依存チェック未実装

LINE 予約などで使用中の問診テンプレートを削除できてしまう可能性がある。

**修正方針**:

```go
// inquiry_template_repository.go に追加
CountUsageByInquiryTemplateID(ctx context.Context, clinicID, id uint64) (int64, error)

// inquiry_template_service.go の Delete に追加
count, err := s.repo.CountUsageByInquiryTemplateID(ctx, clinicID, id)
if err != nil {
    return apperrors.Wrap(err, "failed to check inquiry template usage")
}
if count > 0 {
    return apperrors.WrapConflict("この問診テンプレートは使用中のため削除できません")
}
```

---

### 3. `reservation_type` — `ExistsByReservationTypeID`（bool型）を `CountUsageBy`（int64型）に統一

現在 `Delete` で `ExistsByReservationTypeID(ctx, clinicID, id) bool` を使用しているが、  
他の全マスタは `CountUsageByXxxID` で `int64` を返すパターンで統一されている。

**修正方針**: `ReservationTypeRepository` に `CountUsageByReservationTypeID` を追加し、  
Service で `bool` ではなく `int64 > 0` のチェックに統一。

```go
// reservation_type_repository.go
CountUsageByReservationTypeID(ctx context.Context, clinicID, id uint64) (int64, error)
```

---

### 4. `cage` — `ExistsByCageID` に `clinicID` なし（マルチテナント安全性）

`ExistsByCageID(ctx context.Context, id uint64) bool` の引数に `clinicID` がなく、  
他クリニックのケージが使用中かどうかを誤判定するリスクがある。

**修正方針**: `CountRecordsByCageID(ctx context.Context, clinicID, id uint64) (int64, error)` に変更し、  
クエリに `clinic_id` フィルタを追加。

---

## 規約参照

`.claude/CLAUDE.md` 1b節:
> マスタ削除時は必ず依存レコードの存在をチェックし、参照がある場合は `apperrors.WrapConflict(...)` で 409 を返す。

## テスト

各エンティティに対して:
- 使用中レコードを削除しようとした場合に 409 Conflict が返ることを検証
- 未使用レコードを削除できることを検証
