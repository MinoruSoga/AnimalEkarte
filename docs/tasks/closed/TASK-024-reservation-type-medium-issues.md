# TASK-024: reservation_type 系 MEDIUM 問題 5件

## 優先度

MEDIUM（ただし MEDIUM-5 は実行時パニックリスクあり、優先対応を推奨）

---

## 問題 1: jstTimeLocation() の _ エラー握りつぶし（パニックリスク）

### ファイル
`backend/internal/repository/reservation_type_occupation_repository.go:15-17`

### 問題
```go
func jstTimeLocation() *time.Location {
    loc, _ := time.LoadLocation("Asia/Tokyo") // エラー無視
    return loc
}
```
tzdata が欠落した環境で `loc` が nil になり、後続の `date.In(nil)` でパニックが発生する。`_ = err` 禁止（code-style.md）にも違反。

### 修正案
```go
// パッケージ init で一度だけロード。起動時にエラーが出るので早期発見できる
var jstLoc = func() *time.Location {
    loc, err := time.LoadLocation("Asia/Tokyo")
    if err != nil {
        panic(fmt.Sprintf("failed to load JST timezone: %v", err))
    }
    return loc
}()
```

---

## 問題 2: ハンドラ層でデフォルト値を補完している

### ファイル
`backend/internal/handler/reservation_type_handler.go:59-66`

### 問題
```go
reservationVisible := true
if req.ReservationVisible != nil {
    reservationVisible = *req.ReservationVisible
}
durationMinutes := 15
if req.DurationMinutes != nil {
    durationMinutes = *req.DurationMinutes
}
```
デフォルト値適用はビジネスロジック（service の責務）。handler はリクエスト→Input DTO 変換のみ行うべき。

### 修正案
`CreateReservationTypeInput` の `DurationMinutes` と `ReservationVisible` をポインタ型にして service 内でデフォルト補完する:
```go
// service/reservation_type_service.go
func (s *reservationTypeService) Create(ctx context.Context, clinicID uint64, input CreateReservationTypeInput) (...) {
    durationMinutes := 15
    if input.DurationMinutes != nil {
        durationMinutes = *input.DurationMinutes
    }
    reservationVisible := true
    if input.ReservationVisible != nil {
        reservationVisible = *input.ReservationVisible
    }
    // ...
}
```

---

## 問題 3: デフォルトカラーコードのハードコード重複

### ファイル
`backend/internal/service/reservation_type_group_service.go:63-66`

### 問題
```go
color := input.Color
if color == "" {
    color = "#3B82F6"  // DB default と二重管理
}
```
`reservation_type_groups` テーブルには `gorm:"default:'#3B82F6'"` が設定されているため、INSERT 時は DB に任せれば済む。service 層でのハードコードは管理箇所が二重になる。

### 修正案
```go
// DB default に任せるため service 側の補完を削除
// Input が空文字の場合は GORM の default が機能するよう color フィールドをポインタ型にするか
// または定数で一元管理
const defaultGroupColor = "#3B82F6"
```

---

## 問題 4: reservation_type_group_repository の RowsAffected チェック非対称

### ファイル
`backend/internal/repository/reservation_type_group_repository.go:75-78`

### 問題
`reservation_type_repository.Update` は RowsAffected=0 のとき Count で再確認（論理削除済みレコードのフォールスポジティブ防止）してから `WrapNotFound` を返すが、`reservation_type_group_repository.Update` は Count チェックなしで即 `WrapNotFound` を返す。動作結果は同じでも実装の非対称性がメンテナンス負荷を増す。

### 修正案
`reservation_type_repository.go` の Update パターンを参照実装として `reservation_type_group_repository.go` に適用する。

---

## 問題 5: ReservationTypeService インターフェースが 12 メソッドで肥大

### ファイル
`backend/internal/service/reservation_type_service.go:140-155`

### 問題
12 メソッドを1インターフェースに集約している。規約では 3〜5 メソッドが推奨。

### 修正案
```go
// 分割案
type ReservationTypeService interface {
    List(ctx, clinicID, ...)
    GetByID(ctx, clinicID, id)
    Create(ctx, clinicID, input)
    Update(ctx, clinicID, id, input)
    Delete(ctx, clinicID, id)
    Reorder(ctx, clinicID, ids)
}

type ReservationTypeUnavailableTimeService interface {
    ListUnavailableTimes(...)
    CreateUnavailableTime(...)
    DeleteUnavailableTime(...)
}

type ReservationTypeOccupationService interface {
    ListOccupations(...)
    LinkOccupation(...)
    UnlinkOccupation(...)
}
```
