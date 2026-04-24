# TASK-251: reservation_type_liff_service.go — buildFunc が interface 定義より後ろに定義されている

## 優先度
Medium

## 対象ファイル
- `backend/internal/service/reservation_type_liff_service.go`

## 問題概要
TASK-222（permission_group_service.go / reservation_type_group_service.go）と同様のパターン。
`buildReservationTypeLiffUpdateFields` が `ReservationTypeLiffService` インターフェース定義よりも
後ろ（行163付近）に配置されている。

プロジェクト規約の正しい定義順序:
```
1. const (colXxx = "xxx")
2. buildXxxUpdateFields()
3. type {Entity}Service interface
4. type {entity}Service struct
5. func New{Entity}Service(...)
6. func methods...
```

## 現状コード

```go
// 行13付近: インターフェースが先
type ReservationTypeLiffService interface {
    // ...
}

// 行163付近: buildFunc が後ろに定義（❌ 規約違反）
func buildReservationTypeLiffUpdateFields(input *UpdateReservationTypeLiffInput) map[string]any {
    // ...
}
```

## あるべき姿

```go
// const/buildFunc を先に定義
const (
    colReservationTypeLiffName  = "name"
    colReservationTypeLiffColor = "color"
    // ...
)

func buildReservationTypeLiffUpdateFields(input *UpdateReservationTypeLiffInput) map[string]any {
    // ...
}

// その後にインターフェース
type ReservationTypeLiffService interface {
    // ...
}
```

## 完了条件
- [ ] `buildReservationTypeLiffUpdateFields` をインターフェース定義より前に移動
- [ ] 関連する `const` 宣言も合わせて移動（あれば）
- [ ] ロジックの変更なし
- [ ] `go test ./backend/internal/...` がパス
