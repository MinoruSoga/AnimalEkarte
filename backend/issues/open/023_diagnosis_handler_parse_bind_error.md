---
status: open
---

# diagnosis handler: parseBindError を使用してバリデーションエラーを人間可読に

## 背景

`diagnosis_handler.go` のリクエストバインドエラー処理では `err.Error()` をそのまま返している。
`pet_handler.go` では `parseBindError(err)` でユーザーフレンドリーなメッセージに変換している。

## 問題

```go
// 現在の diagnosis_handler.go
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    // ← 例: "Key: 'createDiagnosisCategoryRequest.Name' Error:Field validation for 'Name' failed on the 'required' tag"
    return
}
```

`err.Error()` はバリデーションライブラリの内部メッセージをそのまま露出するため、
フロントエンドに不適切なメッセージが表示される。

## 修正方針

```go
// pet_handler.go と同パターン
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
    return
}
```

`parseBindError` は `handler/response.go` に定義済みのヘルパー関数。
全4ハンドラ（CreateCategory / UpdateCategory / CreateName / UpdateName）に適用。

## 完了条件

- [ ] `CreateDiagnosisCategory` のバインドエラーを `parseBindError` に変更
- [ ] `UpdateDiagnosisCategory` のバインドエラーを `parseBindError` に変更
- [ ] `CreateDiagnosisName` のバインドエラーを `parseBindError` に変更
- [ ] `UpdateDiagnosisName` のバインドエラーを `parseBindError` に変更
