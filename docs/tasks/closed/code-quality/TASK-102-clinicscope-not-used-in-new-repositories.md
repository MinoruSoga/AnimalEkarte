# TASK-102: `clinicScope` 未使用 — 新規 3 repository で直接 `WHERE clinic_id = ?` を記述

## 優先度

**Medium** — バックエンドルール違反。マルチテナント分離ロジックの一貫性が失われる。

---

## 概要

FEAT-368 で追加した `closing_special_period_repository.go`, `cash_register_close_repository.go`,
`clinic_settings_repository.go` の 3 ファイルが、`clinicScope(clinicID)` を使わず
`Where("clinic_id = ?", clinicID)` を直接記述している。

`backend/CLAUDE.md` には以下のように明記されている:

> ❌ 禁止: 手動で clinic_id を WHERE に記述  
> ✅ 必須: clinicScope を使用

他の全マスタ repository（`payment_method_master_repository.go` 等）は `clinicScope` を使用しており、
今回の3ファイルが例外的に不統一になっている。

---

## 問題箇所

### `repository/closing_special_period_repository.go`

全 6 メソッド（FindAll / FindByID / FindByDate / UpdateFields / Delete / HasOverlap）で
直接 `Where("clinic_id = ?", clinicID)` を記述。

### `repository/cash_register_close_repository.go`

FindAll / FindByID / FindByDateAndPeriod の 3 メソッドで直接記述。

### `repository/clinic_settings_repository.go`

Get メソッドで直接記述。

---

## 比較: 正しい実装（プロジェクト内参照実装）

```go
// ✅ repository/payment_method_master_repository.go:31-39
func (r *paymentMethodMasterRepository) FindAll(ctx context.Context, clinicID uint64) ([]model.PaymentMethodMaster, error) {
    var ms []model.PaymentMethodMaster
    err := r.db.WithContext(ctx).
        Scopes(clinicScope(clinicID)).
        Order("display_order ASC, id ASC").
        Find(&ms).Error
    ...
}
```

---

## 修正方針

各ファイルで `Where("clinic_id = ?", clinicID)` の箇所を `Scopes(clinicScope(clinicID))` に置き換える。
複合 WHERE 条件（例: `HasOverlap` の `start_date <= ? AND end_date >= ?`）は
`.Scopes(clinicScope(clinicID)).Where(残りの条件, ...)` の形式にチェーンする。

### 例: `closing_special_period_repository.go` FindAll

```go
// ❌ 修正前
err := r.db.WithContext(ctx).
    Where("clinic_id = ?", clinicID).
    Order("start_date ASC").
    Find(&periods).Error

// ✅ 修正後
err := r.db.WithContext(ctx).
    Scopes(clinicScope(clinicID)).
    Order("start_date ASC").
    Find(&periods).Error
```

### 例: `closing_special_period_repository.go` HasOverlap

```go
// ❌ 修正前
q := r.db.WithContext(ctx).
    Model(&model.ClosingSpecialPeriod{}).
    Where("clinic_id = ? AND start_date <= ? AND end_date >= ?", clinicID, endDate, startDate)

// ✅ 修正後
q := r.db.WithContext(ctx).
    Model(&model.ClosingSpecialPeriod{}).
    Scopes(clinicScope(clinicID)).
    Where("start_date <= ? AND end_date >= ?", endDate, startDate)
```

---

## 影響範囲

| ファイル | 対象メソッド | 状態 |
|---------|------------|------|
| `repository/closing_special_period_repository.go` | FindAll, FindByID, FindByDate, UpdateFields, Delete, HasOverlap | ❌ 直接記述 |
| `repository/cash_register_close_repository.go` | FindAll, FindByID, FindByDateAndPeriod | ❌ 直接記述 |
| `repository/clinic_settings_repository.go` | Get | ❌ 直接記述 |

---

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — マルチテナント: clinicScope（必須）

> ❌ 禁止: 手動で clinic_id を WHERE に記述  
> ✅ 必須: clinicScope を使用

### プロジェクト内参照実装

- `repository/payment_method_master_repository.go:33` — `Scopes(clinicScope(clinicID))` の正しい使用

---

## 関連ファイル

- `repository/base.go:12` — `clinicScope` の定義
