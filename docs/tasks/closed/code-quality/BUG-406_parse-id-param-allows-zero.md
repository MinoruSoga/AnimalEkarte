# BUG-406: parseIDParam が id=0 を有効値として通過させる

## 概要
`response.go` の `parseIDParam` は URL パスパラメータを `uint64` に変換するが、変換成功時に `id == 0` のチェックがない。`/masters/medicines/0` のようなリクエストがバリデーションを通過してリポジトリ層まで到達し、DB クエリ `WHERE id = 0 AND clinic_id = ?` が発行される。ID として 0 は DB 上存在しないため最終的には 404 になるが、早期拒否がなく不要な DB クエリが発生する。

## 再現手順
1. `GET /v1/masters/medicines/0` にリクエストを送信
2. **結果**: 400 ではなく 404 が返る（DB クエリが発行された後）
3. **期待**: parseIDParam で即座に 400 Bad Request が返る

## 現状コード

### `backend/internal/handler/response.go:268-285`（問題箇所）
```go
func parseIDParam(c *gin.Context, key string) (uint64, bool) {
    s := c.Param(key)
    if s == "" {
        RespondError(c, apperrors.WrapInvalidInput("missing "+key))
        return 0, false
    }
    id, err := strconv.ParseUint(s, 10, 64)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid "+key))
        return 0, false
    }
    return id, true  // ← id == 0 のチェックなし
}
```

## 修正方針

### `response.go:parseIDParam` — id=0 チェック追加
```go
func parseIDParam(c *gin.Context, key string) (uint64, bool) {
    s := c.Param(key)
    if s == "" {
        RespondError(c, apperrors.WrapInvalidInput("missing "+key))
        return 0, false
    }
    id, err := strconv.ParseUint(s, 10, 64)
    if err != nil || id == 0 {  // ← id == 0 チェック追加
        RespondError(c, apperrors.WrapInvalidInput("invalid "+key))
        return 0, false
    }
    return id, true
}
```

## 優先度
**Low** — 機能上の実害は少ない（最終的に 404 が返る）が、不要な DB クエリの発行と、ログに 0 番 ID への誤ったアクセスが記録される問題がある。防衛的プログラミングの観点から修正すべき。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/response.go:268-285` — 修正対象
