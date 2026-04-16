# BUG-254: マルチテナント clinic_id 欠落 — クロスクリニック参照可能

## 概要

複数のドメインで Repository/Handler に `clinic_id` フィルタが欠落しており、
認証済みユーザーが他クリニックのデータを参照・更新・削除できる可能性がある。

## 脆弱性分類

- **CWE-639**: Authorization Bypass Through User-Controlled Key
- **OWASP**: A01:2021 Broken Access Control
- **影響**: 他クリニックの健診・治療・入院・請求データの読み書き

## 影響範囲

### Group A — checkup（健診）

| 対象 | 詳細 | 状態 |
|------|------|------|
| `repository/checkup_repository.go:78-88` | `FindByID` に clinic_id なし。`First(&checkup, id)` のみ | 未修正 |
| `handler/checkup_handler.go:16-28` | `ListCheckups` で `extractClinicID(c)` 未呼び出し | 未修正 |
| `handler/checkup_handler.go:82-133` | `UpdateCheckup` で `extractClinicID(c)` 未呼び出し | 未修正 |
| `handler/checkup_handler.go:135-154` | `DeleteCheckup` で `extractClinicID(c)` 未呼び出し | 未修正 |

### Group C — daily_record（入院日誌）

| 対象 | 詳細 | 状態 |
|------|------|------|
| `repository/daily_record_repository.go:33-98` | 全メソッドに clinic_id なし（`hospitalization_id` のみ） | 未修正 |
| `handler/daily_record_handler.go` | `extractClinicID(c)` 未呼び出し | 未修正 |

### Group C — care_plan_item（ケアプラン）

| 対象 | 詳細 | 状態 |
|------|------|------|
| `repository/care_plan_item_repository.go:31-88` | 全メソッドに clinic_id なし（`hospitalization_id` のみ） | 未修正 |
| `handler/care_plan_item_handler.go` | `extractClinicID(c)` 未呼び出し | 未修正 |

### Group C — treatment（治療）

| 対象 | 詳細 | 状態 |
|------|------|------|
| `repository/treatment_repository.go:38-93` | 全メソッドに clinic_id なし（`medical_record_id` のみ） | 未修正 |
| `handler/treatment_handler.go` | `extractClinicID` を呼ぶが値を下位層に渡していない（`_, ok :=` で破棄） | 未修正 |

### Group C — treatment_plan（治療計画）

| 対象 | 詳細 | 状態 |
|------|------|------|
| `repository/treatment_plan_repository.go:66-89` | `Update`/`Delete` に clinic_id なし（ID のみ） | 未修正 |

### Group B — billing_review（請求レビュー）

| 対象 | 詳細 | 状態 |
|------|------|------|
| `repository/billing_review_repository.go:29-38` | `FindByMedicalRecordID` に clinic_id なし | 未修正 |
| `handler/billing_review_handler.go` | `extractClinicID(c)` 未呼び出し | 未修正 |

### Group B — billing_item（請求明細）

| 対象 | 詳細 | 状態 |
|------|------|------|
| `repository/billing_item_repository.go:43-52` | `FindByBillingID` に clinic_id なし | 未修正 |
| `repository/billing_item_repository.go` | `UpdateBillingTotals` に clinic_id なし | 未修正 |

### Group D — reservation_staff（予約スタッフ）

| 対象 | 詳細 | 状態 |
|------|------|------|
| `repository/reservation_staff_repository.go:45-51` | `FindByID` に clinic_id なし | 未修正 |
| `repository/reservation_staff_repository.go:72-84` | `Update` に clinic_id なし | 未修正 |
| `repository/reservation_staff_repository.go:86-94` | `SoftDelete` に clinic_id なし | 未修正 |

### Group E — clinic（クリニック）

| 対象 | 詳細 | 状態 |
|------|------|------|
| `repository/clinic_repository.go:118-126` | `CountStaffByClinicID` が `staffs.clinic_id` を参照するが staffs テーブルに clinic_id カラムが存在しない | 未修正 |

## 修正方針

### パターン1: 直接 clinic_id フィルタ（テーブルに clinic_id がある場合）

```go
// checkup_repository.go — FindByID に clinicID 追加
func (r *checkupRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Checkup, error) {
    var checkup model.Checkup
    err := r.db.WithContext(ctx).
        Preload("CheckupType").Preload("Doctor").
        Where("id = ? AND clinic_id = ?", id, clinicID).
        First(&checkup).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "checkup", fmt.Sprintf("%d", id))
    }
    return &checkup, nil
}
```

### パターン2: JOIN で親テーブル経由の clinic_id 検証（子テーブルに clinic_id がない場合）

```go
// daily_record_repository.go — hospitalization 経由で clinic_id 検証
func (r *dailyRecordRepository) ListByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error) {
    var records []model.DailyRecord
    err := r.db.WithContext(ctx).
        Joins("JOIN hospitalizations ON hospitalizations.id = daily_records.hospitalization_id").
        Where("hospitalizations.clinic_id = ? AND daily_records.hospitalization_id = ?", clinicID, hospitalizationID).
        Order("record_date DESC").
        Find(&records).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "daily_record", "")
    }
    return records, nil
}
```

### パターン3: clinic_repository.go の CountStaffByClinicID 修正

```go
// staff_clinic_assignments テーブルを参照
func (r *clinicRepository) CountStaffByClinicID(ctx context.Context, clinicID uint64) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&model.StaffClinicAssignment{}).
        Where("clinic_id = ?", clinicID).
        Count(&count).Error
    if err != nil {
        return 0, apperrors.FromGORM(err, "staff_clinic_assignment", fmt.Sprintf("clinic_id=%d", clinicID))
    }
    return count, nil
}
```

## 準拠すべきプロジェクト規約

### `.claude/rules/database-design.md` — マルチテナント設計
> 全テーブルに `clinic_id` (マルチテナント)
> WHERE 句は `clinic_id` から開始

### `.claude/CLAUDE.md` — handler パターン
> Handler: `extractClinicID(c)` でテナント境界を確保

## 優先度

**Critical** — 認証済みユーザーが他クリニックのデータを読み書きできるセキュリティ脆弱性。

## 再検証結果（2026-04-10）

| 箇所 | ステータス |
|------|-----------|
| checkup_repository.FindByID | **STILL OPEN** — `First(&checkup, id)` のみ、clinic_id なし |
| checkup_handler ListCheckups/UpdateCheckup/DeleteCheckup | **STILL OPEN** — extractClinicID 未呼び出し |
| daily_record_repository 全メソッド | **STILL OPEN** — hospitalization_id のみ |
| care_plan_item_repository 全メソッド | **STILL OPEN** — hospitalization_id のみ |
| treatment_repository 全メソッド | **STILL OPEN** — medical_record_id のみ |
| treatment_plan_repository Update/Delete | **STILL OPEN** — id のみ |
| billing_review_repository.FindByMedicalRecordID | **STILL OPEN** — medical_record_id のみ |
| billing_review_handler 全メソッド | **STILL OPEN** — extractClinicID 未呼び出し |
| billing_item_repository.FindByBillingID | **STILL OPEN** — billing_id のみ |
| clinic_repository.CountStaffByClinicID | ~~**FIXED**~~ — ただし staffs テーブルに clinic_id カラムが存在しない問題は要確認 |

**9/10 件が未修正。**

## 関連チケット

- BUG-247: clinical_plan の clinic_id 欠落（修正済み）— 同種の問題
- BUG-253: 親チケット
