# BUG-288: timeslot_engine.go の raw error が apperrors を通過しない

## 概要

`timeslot_engine.go` 内の `minutesSinceMidnight` 等が `fmt.Errorf` で返すエラーが、
`GenerateTimeSlots` → `liff_service.go:204` の経路で一度も `apperrors.Wrap` を通過せずに
サービス公開面に到達する。`RespondError` は `apperrors` 型でないエラーを 500 で返すため、
クライアントに適切なエラーメッセージが伝わらない。

## 現状コード

### `service/timeslot_engine.go:121-128`
```go
func GenerateTimeSlots(input TimeSlotsInput) ([]TimeSlot, error) {
    for _, staffInput := range input.Staffs {
        slots, err := generateForStaff(input, staffInput)
        if err != nil {
            return nil, err  // ← raw fmt.Errorf が素通り
        }
    }
}
```

### `service/liff_service.go:204`
```go
return GenerateTimeSlots(input)  // ← apperrors.Wrap なし
```

## 影響範囲

| 対象 | 詳細 |
|------|------|
| `service/timeslot_engine.go:127` | `GenerateTimeSlots` 内の naked return |
| `service/liff_service.go:204` | `GenerateTimeSlots` の戻り値を素通り |

## 修正方針

### `service/liff_service.go:204`
```go
result, err := GenerateTimeSlots(input)
if err != nil {
    return nil, apperrors.Wrap(err, "failed to generate time slots")
}
return result, nil
```

## 準拠すべきプロジェクト規約

### `.claude/rules/go-language.md` — Service エラーハンドリング
> Service: 内部エラーは `apperrors.Wrap(err, "message")` でラッピング

## 優先度

**High** — エラー発生時に500が返り、クライアントに原因が伝わらない。

## 関連チケット

- BUG-287: 第6回監査 親チケット
