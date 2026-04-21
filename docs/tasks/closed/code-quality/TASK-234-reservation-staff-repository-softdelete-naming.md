# TASK-234: reservation_staff_repository.go — SoftDelete メソッド名が他リポジトリの Delete と不統一

## 優先度
Low

## 対象ファイル
- `backend/internal/repository/reservation_staff_repository.go`

## 問題概要
`reservationStaffRepository` のインターフェースおよび実装メソッドが `SoftDelete` という名前を使っているが、
他の全リポジトリは `Delete` で統一されている。

GORM の論理削除（`deleted_at` フィールドによるソフトデリート）は全リポジトリで共通の仕組みであり、
メソッド名で "Soft" を明示する必要はない。

## 現状コード（行20, 95）

```go
// interface
SoftDelete(ctx context.Context, id uint64) error

// 実装
func (r *reservationStaffRepository) SoftDelete(ctx context.Context, id uint64) error {
```

## 比較（他リポジトリ）

```go
// vaccine_repository.go
Delete(ctx context.Context, clinicID, id uint64) error

// procedure_repository.go
Delete(ctx context.Context, clinicID, id uint64) error

// payment_method_master_repository.go
Delete(ctx context.Context, clinicID, id uint64) error
```

## あるべき姿

```go
// interface
Delete(ctx context.Context, clinicID, id uint64) error

// 実装
func (r *reservationStaffRepository) Delete(ctx context.Context, clinicID, id uint64) error {
```

呼び出し元サービスも `SoftDelete` → `Delete` に変更する。

## 完了条件
- [ ] `SoftDelete` → `Delete` にリネーム（interface + 実装 + 呼び出し元 service）
- [ ] シグネチャも `clinicID` 引数を追加して他と統一（clinicID スコープが未適用の場合は合わせて修正）
- [ ] `go test ./backend/internal/...` がパス
