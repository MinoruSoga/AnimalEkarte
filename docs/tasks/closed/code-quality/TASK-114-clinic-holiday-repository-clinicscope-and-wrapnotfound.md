# TASK-114: `clinic_holiday_repository.go` — `clinicScope` 未使用 + `FromGORM(ErrRecordNotFound)` アンチパターン

## 優先度

**Medium** — テナント分離実装の不統一 + エラーハンドリングのアンチパターン混在。

---

## 概要

`clinic_holiday_repository.go` に 2 つの問題が混在している。

1. `Upsert`（行 47-48）で `clinicScope` を使わず `Where("clinic_id = ?")` を直接記述
2. `Delete`（行 80）で `gorm.ErrRecordNotFound` を `apperrors.FromGORM` にハードコードして渡している

---

## 問題箇所

### 問題 1: `repository/clinic_holiday_repository.go:47-48` — `clinicScope` 未使用

```go
// ❌ clinic_id を直接 WHERE に記述（clinicScope を使っていない）
err := r.db.WithContext(ctx).
    Where("clinic_id = ? AND date = ?", holiday.ClinicID, holiday.Date).
    First(&existing).Error
```

`clinic_holidays` テーブルは直接 `clinic_id` を持つため、`clinicScope` を使うのがプロジェクト規約。
同ファイル内の `FindByYearMonth`（行 29-30）および `Delete`（行 72-73）では `Scopes(clinicScope(clinicID))` を正しく使用しており、同一ファイル内で不統一になっている。

### 問題 2: `repository/clinic_holiday_repository.go:80` — `FromGORM(gorm.ErrRecordNotFound, ...)` のアンチパターン

```go
// ❌ gorm.ErrRecordNotFound をハードコードして FromGORM に渡している
if result.RowsAffected == 0 {
    return apperrors.FromGORM(gorm.ErrRecordNotFound, "clinic_holiday", date.Format("2006-01-02"))
}
```

`apperrors.FromGORM` は **GORM から返されたエラーを変換する** 関数であり、`gorm.ErrRecordNotFound` をハードコードして渡す使い方は想定外。
RowsAffected == 0 の場合は `apperrors.WrapNotFound` を使うのがプロジェクト規約（他リポジトリの参照実装）。

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ 同ファイル FindByYearMonth（行 29-30） — clinicScope を正しく使用
r.db.WithContext(ctx).
    Scopes(clinicScope(clinicID)).
    Order("date ASC")

// ✅ 同ファイル Delete（行 72-73） — clinicScope を正しく使用
r.db.WithContext(ctx).
    Scopes(clinicScope(clinicID)).
    Where("date = ?", date.Format("2006-01-02")).
    Delete(&model.ClinicHoliday{})

// ✅ repository/cage_repository.go — RowsAffected == 0 に WrapNotFound を使用
if result.RowsAffected == 0 {
    return apperrors.WrapNotFound("cage", fmt.Sprintf("%d", id))
}
```

---

## 修正方針

### 問題 1: `Upsert` の clinic_id 条件

`Upsert` メソッドには `clinicID` パラメータがないため、`holiday.ClinicID` を使う現状も問題。
メソッドシグネチャに `clinicID uint64` を追加するか、`holiday.ClinicID` を `clinicScope` に渡す形に変更する。

```go
// ✅ 修正後（clinicID をシグネチャに追加する場合）
func (r *clinicHolidayRepository) Upsert(ctx context.Context, clinicID uint64, holiday *model.ClinicHoliday) (*model.ClinicHoliday, error) {
    var existing model.ClinicHoliday
    err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Where("date = ?", holiday.Date).
        First(&existing).Error
```

または既存シグネチャを維持したまま `holiday.ClinicID` を clinicScope に利用:

```go
// ✅ 修正後（シグネチャ変更なし）
err := r.db.WithContext(ctx).
    Scopes(clinicScope(holiday.ClinicID)).
    Where("date = ?", holiday.Date).
    First(&existing).Error
```

### 問題 2: `Delete` の RowsAffected == 0 処理

```go
// ✅ 修正後
if result.RowsAffected == 0 {
    return apperrors.WrapNotFound("clinic_holiday", date.Format("2006-01-02"))
}
```

---

## 影響範囲

| ファイル | 行 | 状態 |
|---------|---|------|
| `repository/clinic_holiday_repository.go:47-48` | Upsert の clinic_id 条件 | ❌ `clinicScope` 未使用 |
| `repository/clinic_holiday_repository.go:80` | Delete の NotFound 判定 | ❌ `FromGORM(ErrRecordNotFound)` アンチパターン |

---

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — マルチテナント: clinicScope（必須）

> プライマリテーブルが直接 `clinic_id` を持つ場合は、`clinicScope` を使用する。
> ❌ 禁止: 手動で clinic_id を WHERE に記述
> ✅ 必須: clinicScope を使用

### プロジェクト内参照実装

- `repository/clinic_holiday_repository.go:29-30, 72-73` — 同ファイル内で clinicScope を正しく使用済み
- `repository/cage_repository.go` — RowsAffected == 0 に WrapNotFound を使用
