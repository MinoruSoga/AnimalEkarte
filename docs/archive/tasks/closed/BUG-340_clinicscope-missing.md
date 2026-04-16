# BUG-340: clinicScope を使わず `clinic_id = ?` を WHERE に直書き（appointment_repository）

## 概要

`appointment_repository.go` の `CountByDateAndSource` メソッドで、同ファイル内の他メソッドが `Scopes(clinicScope(clinicID))` を使用しているにもかかわらず、
このメソッドだけ `Where("clinic_id = ? AND ...")` と直接 `clinic_id` フィルタを混在させている。
一貫性の欠如はマルチテナント境界の見落としリスクを高める。

## 再現手順

コードレビュー上の問題であり、実行時バグはない。ただし将来の変更で `clinic_id` 条件を外した場合、マルチテナント漏洩が発生しうる。

## 現状コード

### `backend/internal/repository/appointment_repository.go:269-280`
```go
func (r *reservationRepository) CountByDateAndSource(ctx context.Context, clinicID uint64, date time.Time, source model.ReservationSource) (int64, error) {
	var count int64
	dateStr := date.Format("2006-01-02")
	err := dbOrTx(ctx, r.db).Model(&model.Appointment{}).
		Where("clinic_id = ? AND DATE(start_time) = ? AND source = ?",
			clinicID, dateStr, source).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.Wrap(err, "count reservations by date and source")
	}
	return count, nil
}
```

### 比較: 正しい実装（同ファイル内の他メソッド）
```go
// appointment_repository.go — FindAll 等
r.db.WithContext(ctx).Scopes(clinicScope(clinicID)).Where("...").Find(&appointments)
```

## 修正方針

### `backend/internal/repository/appointment_repository.go:272-275`
```go
err := dbOrTx(ctx, r.db).Model(&model.Appointment{}).
    Scopes(clinicScope(clinicID)).
    Where("DATE(start_time) = ? AND source = ?", dateStr, source).
    Count(&count).Error
```

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — マルチテナント: clinicScope（必須）
> プライマリテーブルが直接 `clinic_id` を持つ場合は、`clinicScope` を使用する。
>
> ❌ 禁止: 手動で clinic_id を WHERE に記述
> ✅ 必須: clinicScope を使用

`appointments` テーブルは直接 `clinic_id` を持つため `clinicScope` が適用対象。

### プロジェクト内参照実装

同ファイル `appointment_repository.go` の `FindAll`、`FindByID`、`CountConflicts` 等はすべて `Scopes(clinicScope(clinicID))` を使用している。

## 優先度

**Medium** — 現時点でセキュリティ実害はないが、ファイル内で `clinicScope` の使用が統一されていないコードの一貫性問題。

## 関連ファイル

- `backend/internal/repository/appointment_repository.go:269-280`
- `backend/internal/repository/repository.go`（clinicScope 定義）
