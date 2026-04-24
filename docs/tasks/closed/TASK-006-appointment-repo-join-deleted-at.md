# TASK-006: appointment_repository の JOIN 条件に reservation_types.deleted_at IS NULL を追加

## 概要

`appointment_repository.go` のカテゴリ検索クエリで、`reservation_types` テーブルへの JOIN 条件に `deleted_at IS NULL` が含まれていない。論理削除済みの予約区分に紐づいた予約が、カテゴリフィルタの検索結果に混入する可能性がある。

## 優先度

HIGH（データ整合性リスク）

## 影響ファイル

| ファイル | 問題箇所 |
|---------|---------|
| `backend/internal/repository/appointment_repository.go` | L297-299（JOIN 条件の `deleted_at` 漏れ） |

## 規約違反

`.claude/rules/database-design.md`:
> JOIN 先テーブルの `deleted_at IS NULL` が JOIN 条件に含まれていなければならない。

## 現状コード

```go
// appointment_repository.go L297-299（現状）
Joins("JOIN reservation_types ON reservation_types.id = appointments.reservation_type_id").
Where("reservation_types.category = ?", category)
```

## 修正方針

```go
// 修正後
Joins("JOIN reservation_types"+
    " ON reservation_types.id = appointments.reservation_type_id"+
    " AND reservation_types.deleted_at IS NULL").
Where("reservation_types.category = ?", category)
```

## 影響範囲

- `ListByCategory` 相当のメソッド（カテゴリ別予約一覧 API）
- 検索結果に論理削除済み予約区分の予約が含まれることで、UI 上の表示件数・フィルタ結果が不正になる可能性がある

## テスト

- 論理削除済みの `reservation_type` に紐づく予約が、カテゴリ検索に含まれないことを確認するテストケースを追加
- `appointment_repository_test.go` の既存テーブルに以下ケースを追加:
  - `"deleted reservation_type is excluded from category search"`
