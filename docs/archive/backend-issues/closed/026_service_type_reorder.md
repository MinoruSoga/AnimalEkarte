# [BE-026] 予約区分マスタ reorder エンドポイント追加

## 概要

予約区分マスタ（service_types）に並び替えエンドポイントを追加する。
フロントエンドの D&D による sort_order 変更をサポートする。

## エンドポイント

```
PATCH /v1/masters/service-types/reorder
```

## リクエスト

```json
{ "ids": [3, 1, 2] }
```

- `ids`: `uint64` の配列。配列の順序が新しい `sort_order` になる。

## レスポンス

```
204 No Content
```

## 実装参照

`diagnosis_category_handler.go` の `ReorderDiagnosisCategories` と同パターン。

- handler: `PATCH /v1/masters/service-types/reorder`
- service: `ReorderServiceTypes(ctx, ids []uint64) error`
- repository: `BulkUpdateSortOrder(ctx, ids []uint64) error`

## 注意事項

- `ids` に含まれない service_type の sort_order は変更しない
- clinic_id フィルタを必ず適用すること（マルチテナント）
- フロントエンドは BE 実装前は reorder mutation を空関数で代替中

## ステータス

- [ ] 実装
- [ ] テスト
- [ ] レビュー
