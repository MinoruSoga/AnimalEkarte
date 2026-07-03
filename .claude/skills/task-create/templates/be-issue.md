# BE-XXX: イシュータイトル

**Status**: Open
**Priority**: High / Medium / Low
**Affects**: 影響する機能・コンポーネント
**Date Created**: YYYY-MM-DD
**Related**: TASK-XXX, FE-XXX（関連イシュー）

## Summary

1-2行で問題・実装内容を説明。

## 現状のコード

**実際のコードを読んで** 現在の実装を記載（推測禁止）。

```go
// backend/internal/model/xxx.go:行番号
// 現在のコード（関連部分のみ抜粋）
```

## 必要な変更

### 1. DB マイグレーション（該当する場合）

```sql
-- backend/migrations/0NN_<変更内容>.sql（最終番号+1 で新規採番。適用済み migration の編集は禁止 — checksum mismatch で STG db_reset が必要になる）
ALTER TABLE xxx ADD COLUMN yyy TYPE;
```

### 2. Model 変更

```go
// backend/internal/model/xxx.go
// Before → After のコード差分
```

### 3. Repository 変更

```go
// backend/internal/repository/xxx_repository.go
// 追加・修正するメソッド
```

### 4. Service 変更

```go
// backend/internal/service/xxx_service.go
// 追加・修正するメソッド
```

### 5. Handler 変更

```go
// backend/internal/handler/xxx_handler.go
// 追加・修正するメソッド
```

### 6. Request/Response 変更（該当する場合）

```go
// backend/internal/handler/xxx_request.go
// backend/internal/handler/xxx_response.go
```

## API レスポンス形式（該当する場合）

```json
{
  "data": { ... }
}
```

## フロントエンド影響

- `make codegen` で `models.ts` が更新される
- FE-XXX で対応が必要

## 完了条件

- [ ] DB マイグレーション適用
- [ ] モデル変更 + `make codegen`
- [ ] 3層（handler → service → repository）実装
- [ ] 既存テストが通る
- [ ] API レスポンスが期待通り
