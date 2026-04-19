# TASK-011: billing_item / examination_repository の JOIN 条件に deleted_at IS NULL を追加

## 概要

`billing_item_repository.go` と `examination_repository.go` の JOIN クエリで、JOIN 先テーブルの `deleted_at IS NULL` 条件が欠落している。論理削除済みの請求・検査レコードへの参照が誤って取得・カウントされるリスクがある。

## 優先度

HIGH（データ整合性）

## 影響ファイル

| ファイル | 行 | 問題 |
|---------|-----|------|
| `backend/internal/repository/billing_item_repository.go` | L34 | `billings` JOIN に `billings.deleted_at IS NULL` 欠落 |
| `backend/internal/repository/billing_item_repository.go` | L46 | `billings` JOIN に `billings.deleted_at IS NULL` 欠落 |
| `backend/internal/repository/examination_repository.go` | L117 | `exams` JOIN に `exams.deleted_at IS NULL` 欠落 |

## 規約違反

`.claude/rules/database-design.md` および `.claude/CLAUDE.md`:
> JOIN 先テーブルの `deleted_at IS NULL` が JOIN 条件に含まれていなければならない

## 修正方針

### billing_item_repository.go

```go
// FindByID (L34)
Joins("JOIN billings ON billings.id = billing_items.billing_id"+
    " AND billings.clinic_id = ?"+
    " AND billings.deleted_at IS NULL", clinicID)

// FindByBillingID (L46)
Joins("JOIN billings ON billings.id = billing_items.billing_id"+
    " AND billings.deleted_at IS NULL")
```

### examination_repository.go

```go
// CountItemsByExamID (L117)
Joins("JOIN exams ON exam_results.exam_id = exams.id"+
    " AND exams.deleted_at IS NULL")
```

## 影響範囲

- `billing_item_repository.go L34`: 請求明細の取得 API — 論理削除済み請求の明細が含まれる
- `billing_item_repository.go L46`: 請求IDによる明細一覧 — 論理削除済み請求への参照が通過する
- `examination_repository.go L117`: FK 依存チェック（`CountItemsByExamID`）— 論理削除済みの検査に紐づく結果が削除可否判定に影響

## テスト

- 請求を論理削除した後、その明細が取得されないことを確認
- 検査を論理削除した後、`CountItemsByExamID` が 0 を返すことを確認
