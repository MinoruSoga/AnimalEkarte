# TASK-247: diagnosis_handler.go — uint64 クエリパラメータ解析のコードが重複

## 優先度
Medium

## 対象ファイル
- `backend/internal/handler/diagnosis_handler.go`

## 問題概要
`ListDiagnosisNames` と `ListDiagnosisNamesAll` の両メソッドで、
`type_id` クエリパラメータを `strconv.ParseUint` で解析するロジックが重複している。
また、他の handler は `parseIDParam(c, "id")` ヘルパーを使っており、
クエリパラメータ版の共通ヘルパーが存在しないため一貫性がない。

## 現状コード（行182〜188 および 209〜216 付近）

```go
// ListDiagnosisNames
var typeID *uint64
if s := c.Query("type_id"); s != "" {
    id, parseErr := strconv.ParseUint(s, 10, 64)
    if parseErr != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid type_id"))
        return
    }
    typeID = &id
}

// ListDiagnosisNamesAll（同一ロジックが重複）
var typeID *uint64
if s := c.Query("type_id"); s != "" {
    id, parseErr := strconv.ParseUint(s, 10, 64)
    if parseErr != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid type_id"))
        return
    }
    typeID = &id
}
```

## あるべき姿

`handler/` 共通ヘルパーとして `parseOptionalUint64Query` を追加し、重複を排除する。

```go
// handler/params.go または handler/handler.go に追加
func parseOptionalUint64Query(c *gin.Context, key string) (*uint64, bool) {
    s := c.Query(key)
    if s == "" {
        return nil, true
    }
    id, err := strconv.ParseUint(s, 10, 64)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput(fmt.Sprintf("invalid %s", key)))
        return nil, false
    }
    return &id, true
}

// 使用側
typeID, ok := parseOptionalUint64Query(c, "type_id")
if !ok {
    return
}
```

## 完了条件
- [ ] `parseOptionalUint64Query` ヘルパーを追加
- [ ] `ListDiagnosisNames` / `ListDiagnosisNamesAll` の重複コードをヘルパー呼び出しに変更
- [ ] `go test ./backend/internal/...` がパス
