# CODE-QUALITY-217: service Update メソッドの input nil チェック欠落（CODE-QUALITY-209 追加対象）

## 概要

CODE-QUALITY-209 で起票した「service Update メソッドの `input == nil` チェック欠落」の
追加対象。209 起票後のスキャンで新たに3サービスが欠落していることが判明した。

## 該当サービス

| ファイル | メソッド | 状態 |
|---------|---------|------|
| `backend/internal/service/inquiry_template_service.go` | `Update` | `input == nil` チェックなし |
| `backend/internal/service/trimming_option_service.go` | `Update` | `input == nil` チェックなし |
| `backend/internal/service/reservation_type_group_service.go` | `Update` | `input == nil` チェックなし |

## 問題点

CODE-QUALITY-209 と同様。`Update` の最初で `input == nil` をチェックしないと、
呼び出し元がバグで `nil` を渡した場合に nil pointer dereference でパニックする。

標準パターン（CODE-QUALITY-209 修正済みサービスの例）:
```go
func (s *xxxService) Update(ctx context.Context, clinicID, id uint64, input *UpdateXxxInput) (*model.Xxx, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)  // ← 必須
    }
    // ...
}
```

## 修正方針

各サービスの `Update` メソッド先頭（`FindByID` 呼び出しの前）に追加:

```go
if input == nil {
    return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
}
```

`ErrMsgInputNotNil` は `backend/internal/service/validators.go` に定義済み。

## 修正対象の現状確認が必要な箇所

スキャン時点での grep 結果に基づく。修正前に各ファイルの Update メソッド冒頭を
確認してから追加すること。

## 関連チケット

- CODE-QUALITY-209 — 同問題の先行チケット（9サービス対象）。本チケット完了後に
  CODE-QUALITY-209 の対象リストに追記し、まとめてクローズすることを推奨。

## 優先度

MEDIUM — nil pointer dereference は runtime panic を引き起こすが、
handler が `ShouldBindJSON` でバインドに成功していれば nil にはならないため
実害リスクは低い。ただし防御的プログラミングとして統一すべき。
