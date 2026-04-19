# BUG-405: RespondError が ErrForbidden のカスタムメッセージを無視する

## 概要
`response.go` の `RespondError` 関数で `ErrForbidden` の case が `apperrors.AppError.Message` を参照せず、ハードコードの `"forbidden"` 文字列を返す。`ErrUnauthorized` は正しく `AppError.Message` を抽出しているのに `ErrForbidden` だけが異なる実装になっており、`WrapForbidden("詳細メッセージ")` で設定したカスタムメッセージがクライアントに届かない。

## 再現手順
1. 権限のないユーザーが編集エンドポイントを呼び出す
2. **結果**: `{"error": "forbidden"}` が返る（カスタムメッセージが失われる）
3. `ErrUnauthorized` と比較: `{"error": "ログインが必要です"}` のように詳細メッセージが返る

## 現状コード

### `backend/internal/handler/response.go:60-61`（問題箇所）
```go
case errors.Is(err, apperrors.ErrForbidden):
    c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})  // ← カスタムメッセージ無視
```

### 比較: 正しい実装（ErrUnauthorized — 同ファイル内）
```go
case errors.Is(err, apperrors.ErrUnauthorized):
    var appErr *apperrors.AppError
    msg := "unauthorized"
    if errors.As(err, &appErr) {
        msg = appErr.Message  // ← AppError.Message を参照 ✅
    }
    c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
```

## 影響範囲

| 対象 | 変更内容 |
|------|---------|
| `backend/internal/handler/response.go:60-61` | `ErrForbidden` case を `ErrUnauthorized` と同パターンに修正 |
| `backend/internal/handler/response.go` | `ErrNotImplemented` case も同様の問題がある可能性（確認して修正） |

## 修正方針

### `response.go:60-61` — ErrForbidden を AppError.Message 対応に修正
```go
case errors.Is(err, apperrors.ErrForbidden):
    var appErr *apperrors.AppError
    msg := "forbidden"
    if errors.As(err, &appErr) && appErr.Message != "" {
        msg = appErr.Message
    }
    c.JSON(http.StatusForbidden, gin.H{"error": msg})
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — エラー処理の統一
> handler: `RespondError(c, err)` で統一レスポンス。エラーの詳細はユーザーに適切に返すこと。

## 優先度
**Medium** — 権限エラー時にユーザーへ適切なメッセージが届かない。フロントエンドが `"forbidden"` をキャッチしても詳細理由を表示できない。UX と可観測性の問題。

## 関連チケット
なし

## 関連ファイル
- `backend/internal/handler/response.go:60-61` — 修正対象
