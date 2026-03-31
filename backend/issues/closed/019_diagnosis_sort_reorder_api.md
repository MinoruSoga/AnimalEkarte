---
status: open
---

# 診断病名マスタ: 並び替え (reorder) API エンドポイントの実装

## 背景

Figmaデザインの診断病名マスタページ（`/settings/diagnosis`）の各行に
ドラッグハンドル（⠿）が表示されており、行の並び替えを想定したUIとなっている。

`DiagnosisCategory` / `DiagnosisName` モデルには既に `sort_order int` カラムが存在するが、
一括並び替えに対応した専用 API エンドポイントがない。

## 問題

現在の PATCH エンドポイントは1件ずつ `sort_order` を更新する設計。
ドラッグ&ドロップで複数行の順序を一度に変更する場合、N回のAPIコールが必要になり非効率。

## 修正方針

### 追加エンドポイント（2本）

```
PATCH /v1/masters/diagnosis-categories/reorder
PATCH /v1/masters/diagnosis-names/reorder
```

### リクエストボディ

```json
{
  "ids": ["3", "1", "5", "2", "4"]
}
```

`ids` の順番が新しい `sort_order` になる（index 0 → sort_order 1）。

### レスポンス

```json
{ "message": "reordered" }
```

### 実装方針

- `clinic_id` でスコープ制限（マルチテナント考慮）
- トランザクション内で全件の `sort_order` を一括更新
- `ids` に存在しないIDが含まれる場合は 400 Bad Request

## 完了条件

- [ ] `PATCH /v1/masters/diagnosis-categories/reorder` 実装
- [ ] `PATCH /v1/masters/diagnosis-names/reorder` 実装
- [ ] フロントエンド: `@dnd-kit/core` 等でドラッグ&ドロップ実装
- [ ] フロントエンド: ドロップ完了時に reorder API を呼び出す
