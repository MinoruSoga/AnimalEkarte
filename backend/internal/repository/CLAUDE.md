# Repository Layer — P2 / P3 / P4 / P9 / P16

## P2: deleted_at IS NULL in CountUsage (MANDATORY)

```go
// ✅
err := r.db.WithContext(ctx).Model(&model.Vaccination{}).
    Where("vaccine_id = ? AND deleted_at IS NULL", vaccineID).Count(&count).Error

// ❌
Where("vaccine_id = ?", vaccineID)  // deleted_at IS NULL なし
```

## P3: Preload with deleted_at IS NULL (MANDATORY)

ソフトデリート対象エンティティのすべての `Preload` に条件を付ける。

```go
// ✅
db.Preload("ReservationType", "deleted_at IS NULL").
   Preload("Doctor", "deleted_at IS NULL").Find(&reservations)

// ❌
db.Preload("ReservationType").Preload("Doctor").Find(&reservations)
```

**注意**: `Preload("Doctor")` は `Staff` モデルへのエイリアス — 対象。

対象 42 エンティティ（抜粋）: `Account`, `Billing`, `Cage`, `Checkup`, `Consultation`,
`Examination`, `Hospitalization`, `MedicalRecord`, `Medicine`, `Owner`, `Pet`,
`Reservation`, `Staff`, `Treatment`, `Vaccination`, `Vaccine` など。
完全なリストは `.claude/refs/gin-architecture-compliance.md` P3 を参照。

## P3.1: clinic-scoped マスタの Preload は clinic_id 述語必須 (MANDATORY — クロステナント read 漏洩防止)

clinic_id を持つ「マスタ/区分」を FK 値で `Preload` する場合、`deleted_at IS NULL` だけでなく
**`clinic_id` 述語を必ず付ける**。base クエリが clinic-scoped でも、FK 値（例: `vaccination.vaccine_id`）が
別クリニックのマスタを指すと（write 側の FK 検証漏れ・過去データ汚染 #124/#125）、clinic_id 述語の無い
Preload は別クリニックのマスタ名/価格を応答に混入させる（IDOR / read 漏洩）。

```go
// ✅ 単一クリニック
Preload("Vaccine", "clinic_id = ? AND deleted_at IS NULL", clinicID)
// ✅ 拠点横断 (#86: clinicIDs は handler で所属検証済み)
Preload("ReservationType", "clinic_id IN ? AND deleted_at IS NULL", clinicIDs)

// ❌ clinic_id 述語なし → 別クリニックのマスタが混入する
Preload("Vaccine", "deleted_at IS NULL")
```

対象マスタ: `Vaccine` / `Medicine` / `Procedure` / `Consultation` / `ReservationType`(+`.Group`) /
`TrimmingCourse` / `TrimmingOption`(`TrimmingDetail.Course/Options`) / `Cage` / `Insurance` /
`ExaminationType` / `CheckupType` / `DiagnosisType`(+`Names`)/`DiagnosisName` など。
`AnimalSpecies`・`ManualArticle` はグローバルマスタ（clinic_id 無し）のため対象外。

**例外: Staff(`Doctor`/`EnteredByStaff`/`PaidByStaff`/`ClosedByStaff`/`CreatedByStaff` 等)**
staff は `staff_clinic_assignments` による**多医院所属**（`staffs.clinic_id` は主所属のみ）。
`staffs.clinic_id = ?` 単純スコープは共有スタッフを誤って隠す。さらに既往カルテ等の**履歴 preload**を
スコープすると、退職/再配属したスタッフの担当医名が過去記録から消える回帰を生む。
- **履歴系 preload（medical_record/vaccination/examination/hospitalization/checkup/billing/refund 等の Doctor/Staff）は意図的に scope しない**（漏洩は staff 名のみで低 severity・write 隔離 72e8887c で通常到達不可）。
- **現在/未来データの reservation の `Doctor`/`CreatedByStaff` のみ** assignment-EXISTS でスコープ可
  （`staffAssignedToClinicsCond`・`reservation_repository.go`）。`staffs.clinic_id` でなく
  `staff_clinic_assignments` の所属で判定し多医院所属を尊重する。

## P4: clinicScope on Update/Upsert (MANDATORY — 最重要)

UPDATE/UPSERT に `Scopes(clinicScope(clinicID))` を必ず付ける。

```go
// ✅
err := r.db.WithContext(ctx).Model(&model.Vaccine{}).
    Scopes(clinicScope(clinicID)).
    Where("id = ?", id).Updates(fields).Error

// ❌ クロスクリニックデータ更新リスク
r.db.Model(&model.Vaccine{}).Where("id = ?", id).Updates(fields)
```

**例外（clinicScope 不要）**: `clinic_repository.go`, `company_repository.go`,
`account_repository.go`, `password_reset_token_repository.go`, `audit_repository.go`

## P9: apperrors.FromGORM on GORM errors (MANDATORY)

```go
// ✅
if err := r.db.WithContext(ctx).First(&vaccine, id).Error; err != nil {
    return nil, apperrors.FromGORM(err, "vaccine", fmt.Sprintf("%d", id))
}

// ❌
if err := r.db.First(&vaccine, id).Error; err != nil {
    return nil, err
}
```

## P16: Method naming conventions (MANDATORY)

```
FindAll / FindByClinicID  ← 一覧（GetAll, List, Fetch は違反）
FindByID                  ← 単件（GetByID, Get, Find は違反）
Create / Update / Delete  ← 標準
CountBy{Xxx}              ← カウント
CountUsageBy{Xxx}         ← 使用数カウント
```

```go
// ✅
type VaccineRepository interface {
    FindAll(ctx context.Context, clinicID uint64) ([]*model.Vaccine, error)
    FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccine, error)
    Create(ctx context.Context, vaccine *model.Vaccine) error
    Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Vaccine, error)
    Delete(ctx context.Context, clinicID, id uint64) error
}

// ❌
GetAll(...)  GetByID(...)  List(...)  Fetch(...)
```
