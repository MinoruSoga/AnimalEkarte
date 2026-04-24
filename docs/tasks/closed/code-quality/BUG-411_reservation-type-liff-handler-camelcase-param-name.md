# BUG-411: reservation_type_liff_handler のパスパラメータ名が camelCase（命名規則違反）

## 概要

`reservation_type_liff_handler.go` 内の `parseIDParam` 呼び出しで、パスパラメータ名に
`"typeId"` という camelCase を使用している。プロジェクト規約（`.claude/rules/naming-conventions.md`）
では API パスは `kebab-case` / `snake_case` 統一のため違反。

## 問題箇所

```go
// reservation_type_liff_handler.go:63
id, ok := parseIDParam(c, "typeId")   // ← camelCase 違反

// 同様に行 98, 115, 138 も "typeId" 使用
```

## 他ファイルとの比較

```go
// exam_type_handler.go:36（標準パターン）
id, ok := parseIDParam(c, "id")   // ← lowercase

// medicine_handler.go:43
id, ok := parseIDParam(c, "id")
```

## 修正方針

1. `reservation_type_liff_routes.go` のルーティング定義を確認し、パラメータ名を特定
2. ルーティングのパラメータ名を `:id` に統一（`/:typeId` → `/:id`）
3. `reservation_type_liff_handler.go` の4箇所を `parseIDParam(c, "id")` に変更

## 影響ファイル

- `backend/internal/handler/reservation_type_liff_handler.go` — 行 63, 98, 115, 138
- `backend/internal/handler/reservation_type_liff_routes.go` — ルーティング定義

## 優先度

**Low** — 命名規則違反。動作に影響はないが、一貫性確保のため修正すべき。

## 参考

- `.claude/rules/naming-conventions.md` — API パス命名規則
- `BUG-LINE-002`（done）— 類似の typeId/courseId 不統一問題（LINE 側で修正済み）
