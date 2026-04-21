# TASK-221: reservation_type_liff_response.go — category フィールドが欠落（管理側 Response には存在）

## 優先度
Medium

## 対象ファイル
- `backend/internal/handler/reservation_type_liff_response.go`
- `backend/internal/handler/reservation_type_liff_handler.go`

## 問題概要
`reservationTypeLiffResponse` に `category` フィールドが存在しない。
管理側の `reservationTypeResponse` には `Category model.ReservationTypeCategory \`json:"category"\`` が含まれている。

`category` は予約コースを `general`（通常）/ `trimming`（トリミング）等に分類するフィールドであり、
LIFF 側でトリミング専用コースを除外するフィルタなどに使用できる情報。

## あるべき姿

```go
type reservationTypeLiffResponse struct {
    // 既存フィールド
    ID              uint64 `json:"id"`
    Name            string `json:"name"`
    // ...
    // 追加
    Category        string `json:"category"`
}

func toReservationTypeLiffResponse(r model.ReservationType) reservationTypeLiffResponse {
    return reservationTypeLiffResponse{
        // 既存マッピング
        Category: string(r.Category),  // 追加
    }
}
```

## 確認事項
- フロントエンド（LIFF 側）で `category` が必要かどうかを仕様確認すること
- 不要であれば起票を保留してよい

## 完了条件
- [ ] `reservationTypeLiffResponse` に `Category string \`json:"category"\`` を追加
- [ ] `toReservationTypeLiffResponse` で `Category` をマッピング
- [ ] `go test ./backend/internal/...` がパス
