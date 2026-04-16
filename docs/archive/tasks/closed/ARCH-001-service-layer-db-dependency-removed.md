# ARCH-001: サービス層 *gorm.DB 依存の除去 ✅ 完了

## 概要

サービス層が `*gorm.DB` を直接保持していたアーキテクチャ違反を修正した。
Clean Architecture では repository 層のみが ORM に依存すべきであり、service 層は
インフラ詳細（GORM）を知ってはならない。

## 修正前の問題

| ファイル | 違反内容 |
|--------|---------|
| `appointment_service.go` | `db *gorm.DB` フィールド保持、`s.db.WithContext(ctx).Transaction(...)` 直接使用 |
| `appointment_admin_service.go` | `db *gorm.DB` フィールド保持、`s.db.WithContext(ctx).Transaction(...)` 直接使用 |
| `reservation_validators.go` | `db *gorm.DB` フィールド保持、`v.db.WithContext(ctx).Transaction(...)` 直接使用 |
| `liff_service.go` | `NewLiffService(..., db *gorm.DB, ...)` で受け取り Validators に渡す |

## 修正内容

### 1. `repository/transactor.go` （新規）

```go
type Transactor interface {
    WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

サービス層はこのインターフェース経由でトランザクションを開始する。
`gormTransactor` はトランザクション開始時に `*gorm.DB` を context に格納する。

### 2. `repository/base.go` — `dbOrTx` ヘルパー追加

```go
func dbOrTx(ctx context.Context, db *gorm.DB) *gorm.DB {
    if tx := txFromContext(ctx); tx != nil {
        return tx.WithContext(ctx)
    }
    return db.WithContext(ctx)
}
```

Repository メソッドはすべて `r.db.WithContext(ctx)` を `dbOrTx(ctx, r.db)` に置換。
`Transactor.WithTx` が渡したコンテキストを受け取った場合、自動的に同一 tx を使用する。

### 3. `repository/appointment_repository.go` — インターフェース拡張

新規メソッド（すべて `dbOrTx` でコンテキストの tx を自動使用）:

| メソッド | 用途 |
|---------|------|
| `LockAndFindByID` | `SELECT ... FOR UPDATE` で行ロック取得（競合チェック用） |
| `HasDoctorConflict` | 医師の時間枠重複チェック（`FOR UPDATE`） |
| `CountOnDutyDoctors` | 当日出勤医師数（shift_entries JOIN）|
| `CountConflicts` | 全体競合予約数（`FOR UPDATE`）|
| `CountByCustomerAndDateRange` | 日次・月次制限チェック用件数 |
| `CountByDateAndSource` | 確認番号生成用件数 |

### 4. サービス層の変更

**`appointment_service.go`**:
- `db *gorm.DB` → `tx repository.Transactor`
- `checkSlotConflict(ctx, tx *gorm.DB, ...)` → `checkSlotConflict(ctx, repo ReservationRepository, ...)`
- `Create`, `updateWithConflictCheck` が `s.tx.WithTx(ctx, ...)` を使用

**`appointment_admin_service.go`**:
- `db *gorm.DB` → `resRepo repository.ReservationRepository, tx repository.Transactor`
- `Create` が `s.tx.WithTx(ctx, ...)` を使用

**`reservation_validators.go`**:
- `db *gorm.DB` → `tx repository.Transactor, repo repository.ReservationRepository`
- `ValidateAndCreate` が `v.tx.WithTx(ctx, ...)` を使用
- 日次・月次制限チェックを `repo.CountByCustomerAndDateRange` に委譲
- `generateConfirmationNumber` が `repo.CountByDateAndSource` を使用

**`liff_service.go`**:
- `db *gorm.DB` → `tx repository.Transactor, reservationRepo repository.ReservationRepository`
- `NewReservationValidators(tx, reservationRepo)` に変更

### 5. DI 配線（`service.go`）

```go
tx := repository.NewTransactor(repos.DB())

Reservation:      NewReservationService(repos.Reservation, tx),
ReservationAdmin: NewReservationAdminService(repos.ReservationAdmin, repos.Reservation, tx),
Liff:             NewLiffService(..., tx, repos.Reservation, notifier),
```

## 動作保証

- `go build ./...` → EXIT:0 ✅
- `go test ./internal/service/...` → PASS ✅

## アーキテクチャ上の意義

```
Before:
  service.Create() → s.db.Transaction() → raw SQL inside closure

After:
  service.Create() → s.tx.WithTx(ctx, func(ctx) {
                          s.repo.CheckDoctorConflict(ctx)  ← ctx に tx が入っている
                          s.repo.Create(ctx)               ← 同一 tx を自動使用
                      })
```

service 層は GORM を知らない。test では `Transactor` を mock 化することで
DB 不要のユニットテストが可能になる。
