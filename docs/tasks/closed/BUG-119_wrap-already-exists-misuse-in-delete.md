# BUG-119: 削除時 FK 参照エラーで WrapAlreadyExists 誤用 — トーストに不正形式メッセージ表示

## 概要

3 つのマスタサービスで、FK 参照による削除禁止エラーを返す際に
`WrapConflict` ではなく `WrapAlreadyExists` を誤用している。
その結果、フロントエンドのトーストに
`"service_type 'この項目は予約データで使用中のため削除できません' already exists"`
のような機械的な文字列が表示されてしまう。

## 実確認 (ローカル確認: 2026-04-01)

`DELETE /api/v1/masters/service-types/1` → HTTP 409 だが、レスポンスボディが:
```json
{"error":"service_type 'この項目は予約データで使用中のため削除できません' already exists"}
```
UI トーストに上記の不正形式のメッセージがそのまま表示される。

## 誤用箇所 (backend)

| ファイル | 行 | 誤用 |
|---------|-----|------|
| `internal/service/service_type_service.go:123` | `WrapAlreadyExists("service_type", "この項目は予約データで使用中のため削除できません")` | ❌ |
| `internal/service/staff_service.go:165,172` | `WrapAlreadyExists("staff", "このスタッフはシフト・予約データで使用中のため削除できません")` | ❌ |
| `internal/service/insurance_service.go:59` | `WrapAlreadyExists("insurance", "この保険はペット情報で使用中のため削除できません")` | ❌ |

## 修正方法

`WrapAlreadyExists(entity, msg)` → `WrapConflict(msg)` に置き換えるだけ。

```go
// ❌ Before
return apperrors.WrapAlreadyExists("service_type", "この項目は予約データで使用中のため削除できません")

// ✅ After
return apperrors.WrapConflict("この項目は予約データで使用中のため削除できません")
```

## 正常動作しているエンドポイント (参考)

`job_title_service`, `medicine_service`, `trimming_master_service` 等は
`WrapConflict` を正しく使用しており、クリーンなメッセージを返している。

## 影響

- FE トーストに不正形式のメッセージが表示されてユーザーが混乱
- `/settings/service-type`, `/settings/staff`, `/settings/insurance` の削除フロー

## 関連

- BUG-106 (削除エラーメッセージ generic 問題) — FE 側は修正済み、BE 側の残件
