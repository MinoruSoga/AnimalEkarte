# BUG-256: Service 層で apperrors.Wrap なし naked return（第2波）

## 概要

BUG-249（第1回監査）で修正した箇所以外に、20+ サービスで Repository 呼び出しのエラーを
`apperrors.Wrap` なしで直接返している。List/GetByID のパススルーと、
Create/Update/Delete 末尾の `return s.repo.FindByID(...)` の両方が対象。

## 影響範囲

### Group A — Core Medical

| ファイル | メソッド | 行番号 |
|---------|---------|--------|
| `service/medical_record_service.go` | List, GetByID, CountByPetID | :92 等 |
| `service/vital_service.go` | List | :56 |
| `service/diagnosis_service.go` | Category.List/GetByID/Update末尾, Name.List/ListByCategory/GetByID | :89,:93,:124,:192,:196,:200 |
| `service/examination_service.go` | List, GetByID | :30,:34 |
| `service/exam_type_service.go` | List, GetByID, Create, Delete, Reorder | :31,:34,:37,:59,:66 |
| `service/checkup_service.go` | List, Create末尾, Update末尾 | :63,:93,:115 |
| `service/checkup_type_service.go` | List, GetByID, Create, Delete, Reorder | :33,:36,:38,:61,:68 |
| `service/consultation_service.go` | List, GetByID, Reorder | :33,:36,:79 |
| `service/vaccination_service.go` | List, GetByID, Create, Delete | :30,:33,:41,:124 |
| `service/vaccine_service.go` | List, GetByID, Create, Delete, Reorder | :31,:33,:36,:103,:110 |

### Group B — Financial

| ファイル | メソッド | 行番号 |
|---------|---------|--------|
| `service/inventory_service.go` | List, GetByID, Create | :30,:34,:38 |
| `service/refund_service.go` | ListByBillingID | :74 |
| `service/insurance_service.go` | Reorder | :80 |

### Group C — Hospital Operations

| ファイル | メソッド | 行番号 |
|---------|---------|--------|
| `service/procedure_service.go` | List, GetByID, Create, Delete, Reorder | :33,:36,:39,:64,:71 |
| `service/trimming_service.go` | List, GetByID, Create末尾, Update末尾, Delete | :69,:73,:110,:168,:172 |
| `service/treatment_plan_service.go` | ListByMedicalRecord, ListByHospitalization, Create末尾, Update末尾 | :53,:57,:87,:99 |
| `service/chief_complaint_category_service.go` | List, GetByID, Update末尾, Delete | :42,:46,:70,:82 |

### Group D — Reservation

| ファイル | メソッド | 行番号 |
|---------|---------|--------|
| `service/reservation_service.go` | Create内トランザクションエラー | :80 |
| `service/reservation_staff_service.go` | Delete内 GetByID | :119 |
| `service/reservation_setting_service.go` | Upsert末尾 FindByClinicID | :103 |

### Group E — Admin/System

| ファイル | メソッド | 行番号 |
|---------|---------|--------|
| `service/animal_species_service.go` | List, GetByID, Update末尾, Delete, Reorder | :57,:61,:89,:100,:107 |
| `service/occupation_service.go` | List, GetByID, Update末尾, Delete, Reorder | :42,:46,:70,:81,:85 |
| `service/inquiry_template_service.go` | List, GetByID, Update末尾, Delete | :42,:46,:70,:74 |
| `service/permission_group_service.go` | List, GetByID, Update末尾 | :44,:48,:73 |
| `service/company_service.go` | Get, Update末尾 | :44,:57 |

## 修正方針

一括パターン:

```go
// Before
func (s *xxxService) List(ctx context.Context, clinicID uint64) ([]model.Xxx, error) {
    return s.repo.FindAll(ctx, clinicID)
}

// After
func (s *xxxService) List(ctx context.Context, clinicID uint64) ([]model.Xxx, error) {
    items, err := s.repo.FindAll(ctx, clinicID)
    if err != nil {
        return nil, apperrors.Wrap(err, "failed to list xxx")
    }
    return items, nil
}
```

## 優先度

**High** — エラーコンテキストが失われ、デバッグ困難になる。

## 関連チケット

- BUG-249: 同種の問題（第1回監査で一部修正済み）
- BUG-253: 親チケット
