# BUG-274: Reservation テスト mock の `source` パラメータ欠落

## 概要

`ReservationRepository.FindAll` と `ReservationService.List` のインターフェースは `source *string` パラメータを持つが、
テスト mock の function field がこのパラメータを省略しており、`source` フィルタの検証が不可能。

## 影響範囲

| ファイル | 行 | 問題 |
|----------|-----|------|
| `service/reservation_service_test.go` | 17 | `findAllFn` の型定義に `source *string` 欠落 |
| `service/reservation_service_test.go` | 26 | `FindAll` メソッドが `source` を `findAllFn` に渡さない |
| `handler/reservation_handler_test.go` | 24 | `listFn` の型定義に `source *string` 欠落 |
| `handler/reservation_handler_test.go` | 32 | `List` メソッドが `source` を `listFn` に渡さない |

## 修正方針

### `reservation_service_test.go`
```go
// BEFORE
findAllFn func(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status *string, petID, ownerID *uint64) (...)

// AFTER
findAllFn func(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) (...)
```

### `reservation_handler_test.go`
```go
// BEFORE
listFn func(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status *string, petID, ownerID *uint64) (...)

// AFTER
listFn func(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) (...)
```

## 優先度

**High** — `source` フィルタ（LINE予約 vs 通常予約の区別）の動作がテストで保証されていない。

## 関連チケット

- BUG-270: 第4回監査 親チケット
