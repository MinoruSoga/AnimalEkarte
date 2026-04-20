# TASK-204: reservation_type_liff_service.go — Update でフィールドが空の場合に atLeastOneField エラーを返さない

## 優先度
High

## 対象ファイル
`backend/internal/service/reservation_type_liff_service.go`

## 問題概要
`Update` メソッドで `buildUpdateFields` の結果が空（更新フィールドなし）の場合に、
他の全ドメインが返す `apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)` を返さず、
現在の値をそのまま返す no-op 動作になっている。

```go
// 現状（NG）— 行103-111付近
fields := buildReservationTypeLiffUpdateFields(input)
if len(fields) == 0 {
    result, _ := s.repo.FindByID(ctx, clinicID, id)
    return result, nil  // 200 OK を返す（他ドメインと挙動が異なる）
}
```

他の全 service（cage, insurance, vaccine, exam_type 等）は同状況で:
```go
if len(fields) == 0 {
    return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
}
```
と 400 Bad Request を返している。

## 修正方針

```go
// あるべき姿（他ドメインと統一）
fields := buildReservationTypeLiffUpdateFields(input)
if len(fields) == 0 {
    return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
}
```

## 影響範囲
- フロントエンド側がフィールド空の PATCH リクエストを送った場合の挙動が変わる（200→400）
- クライアント側の挙動を事前に確認すること

## 完了条件
- [ ] フィールド空の場合に `apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)` を返すよう修正
- [ ] 対応するテストケース（空フィールドで 400 が返ること）を追加
- [ ] `go test ./backend/internal/...` がパス
