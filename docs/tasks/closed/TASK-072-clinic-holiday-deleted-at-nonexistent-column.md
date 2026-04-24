# TASK-072: clinic_holiday_repository — deleted_at IS NULL で存在しないカラム参照

## 優先度

HIGH

---

## 概要

`clinic_holiday_repository.go` の `FindByYearMonth` メソッドが `WHERE deleted_at IS NULL` を
明示的に付加しているが、`clinic_holidays` テーブルに `deleted_at` カラムは存在しない。
PostgreSQL は実行時に `ERROR: column "deleted_at" does not exist` を返す。

---

## 問題箇所

### backend/internal/repository/clinic_holiday_repository.go（L27-43）

```go
func (r *clinicHolidayRepository) FindByYearMonth(...) {
    var holidays []model.ClinicHoliday
    q := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Where("deleted_at IS NULL").   // ❌ 存在しないカラム参照
        Order("date ASC")
    ...
}
```

### backend/migrations/001_init.sql（L1316-1324）

```sql
CREATE TABLE clinic_holidays (
    id         BIGSERIAL   PRIMARY KEY,
    clinic_id  bigint      NOT NULL REFERENCES clinics(id) ON DELETE RESTRICT,
    date       date        NOT NULL,
    reason     text        NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT uk_clinic_holidays_clinic_date UNIQUE (clinic_id, date)
);
-- deleted_at カラムなし
```

---

## 原因

`clinic_holidays` テーブルは soft delete を採用しておらず、`deleted_at` カラムは存在しない。
また、`ClinicHoliday` モデル（`backend/internal/model/staff.go`）も `gorm.DeletedAt` フィールドを持っていない。

Delete は物理削除（`DELETE FROM clinic_holidays WHERE ...`）が正しい動作であり、
`WHERE deleted_at IS NULL` は不要かつ有害。

---

## 修正方針

```go
// ✅ 修正後: deleted_at IS NULL を削除
func (r *clinicHolidayRepository) FindByYearMonth(...) {
    var holidays []model.ClinicHoliday
    q := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Order("date ASC")   // ← deleted_at IS NULL を削除

    if yearMonth != "" {
        q = q.Where("date >= ?::date AND date < (?::date + INTERVAL '1 month')",
            yearMonth+"-01", yearMonth+"-01")
    }

    if err := q.Find(&holidays).Error; err != nil {
        return nil, apperrors.FromGORM(err, "clinic_holiday", yearMonth)
    }
    return holidays, nil
}
```

---

## 修正ファイル

| ファイル | 修正内容 |
|---------|---------|
| `clinic_holiday_repository.go` | `Where("deleted_at IS NULL")` 行を削除（L31） |
