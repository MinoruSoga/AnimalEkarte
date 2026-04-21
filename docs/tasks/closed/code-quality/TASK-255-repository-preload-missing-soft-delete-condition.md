# TASK-255: repository — マスタエンティティの Preload に deleted_at IS NULL 条件が欠落

## 優先度
High

## 対象ファイル
- `backend/internal/repository/appointment_admin_repository.go`（行39, 40, 58, 59, 60, 96, 97, 111, 112）
- `backend/internal/repository/hospitalization_repository.go`（Cage/Doctor Preload）
- `backend/internal/repository/examination_repository.go`（ExaminationType/Doctor Preload）
- `backend/internal/repository/checkup_repository.go`（Doctor Preload）
- `backend/internal/repository/reservation_repository.go`（ReservationType/Doctor Preload）

## 問題概要
複数のリポジトリで、ソフトデリートを持つ**マスタエンティティ**の `Preload` に
`deleted_at IS NULL` 条件が設定されていない。

論理削除されたマスタ（予約コース・スタッフ等）が関連データとして読み込まれ、
API レスポンスに含まれてしまう。

## 正しい参照実装（reservation_type_repository.go）

```go
// ✅ 正しい: 条件付き Preload
Preload("Group", "deleted_at IS NULL")
```

## 問題のある実装例（appointment_admin_repository.go）

```go
// ❌ 条件なし（論理削除済みデータも読み込まれる）
Preload("ReservationType").  // 行39: ReservationType はソフトデリートあり
Preload("Doctor").           // 行40: Doctor(Staff) はソフトデリートあり
Preload("LineCustomer").
```

## 対象マスタエンティティ（ソフトデリートあり）

| エンティティ | ソフトデリート | 主要 Preload 箇所 |
|------------|-------------|-----------------|
| `ReservationType` | ✅ DeletedAt | appointment_admin, reservation |
| `Doctor`（= Staff） | ✅ DeletedAt | appointment_admin, hospitalization, examination, checkup |
| `Cage` | ✅ DeletedAt | hospitalization |
| `ExaminationType` | ✅ DeletedAt | examination |

## あるべき姿

```go
// appointment_admin_repository.go
Preload("ReservationType", "deleted_at IS NULL").
Preload("Doctor", "deleted_at IS NULL").

// hospitalization_repository.go
Preload("Cage", "deleted_at IS NULL").
Preload("Doctor", "deleted_at IS NULL").

// examination_repository.go
Preload("ExaminationType", "deleted_at IS NULL").
Preload("Doctor", "deleted_at IS NULL").
```

## 確認事項
- `LineCustomer`・`CreatedByStaff` 等は論理削除の有無を確認してから条件追加を判断する
- GORM の `Preload("Assoc", "deleted_at IS NULL")` は第2引数に WHERE 条件文字列を渡す形式

## 完了条件
- [ ] 上記リポジトリの各 Preload にソフトデリートあり確認済みのマスタエンティティへの条件追加
- [ ] `go test ./backend/internal/...` がパス
