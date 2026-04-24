# BE-reorder-cages: ケージ並び替えエンドポイントの実装

## 概要

ケージマスタの DnD 並び替えに対応するため、`sort_order` を一括更新するエンドポイントを実装する。

## エンドポイント

```
PATCH /v1/masters/cages/reorder
```

## リクエスト

```json
{
  "ids": [3, 1, 5, 2, 4]
}
```

- `ids`: 新しい順序でケージ ID を並べた配列
- 配列の index が `sort_order` の値となる（0-indexed または 1-indexed はサービス層で決定）

## 処理内容

1. `ids` の順序に従い `cages.sort_order` を一括更新（GORM で `Updates` または raw SQL）
2. `clinic_id` フィルタを適用してマルチテナント安全性を確保
3. `ids` に含まれない ID が来た場合は `400 Bad Request`
4. 存在しない ID が含まれる場合は `404 Not Found`

## 参考実装

- `internal/handler/checkup_type_handler.go` — `ReorderCheckupTypes` ハンドラ
- `internal/service/checkup_type_service.go` — `ReorderCheckupTypes` サービス
- `internal/repository/checkup_type_repository.go` — `BulkUpdateSortOrder`

## 対応フロントエンド

- `frontend/src/features/master/api/cages.ts` — `useReorderCages`（実装済み）
- `frontend/src/features/master/routes/CageSettings.tsx` — DnD UI（実装済み）

## 優先度

Medium（フロントエンドは API エラー時に楽観的更新をロールバックするため、未実装でも動作する）
