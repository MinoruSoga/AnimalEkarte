# CODE-QUALITY-229: hospitalization_plan ドメイン品質問題（3層）

## 概要

`hospitalization_plan`（入院プラン）ドメインに handler / service / repository の
各層にまたがる品質問題が4件。うち repository の問題はマルチテナント分離違反。

---

## 問題1（HIGH）: repository — CountCarePlanItemsByPlanID に clinic_id フィルタなし（マルチテナント違反）

**ファイル:** `backend/internal/repository/hospitalization_plan_repository.go:87-96`

### 現状コード

```go
func (r *hospitalizationPlanRepository) CountCarePlanItemsByPlanID(
    ctx context.Context, clinicID, planID uint64,
) (int64, error) {
    var count int64
    if err := r.db.WithContext(ctx).
        Model(&model.CarePlanItem{}).
        Where("hospitalization_plan_id = ? AND deleted_at IS NULL", planID).
        Count(&count).Error; err != nil {
        return 0, apperrors.FromGORM(err, "care_plan_item", "")
    }
    return count, nil
}
```

### 問題

`clinicID` を引数で受け取っているにもかかわらず、クエリに `clinic_id` フィルタが入っていない。
別 clinic の `care_plan_items` も数えてしまうため、マルチテナント分離違反。

例: clinic A が `hospitalization_plan_id = 1` を持ち、
    clinic B も異なるプランで `hospitalization_plan_id = 1` を持つケースで、
    clinic A の Delete 前チェックが clinic B のデータまでカウントしてしまう。

### 比較（正しい実装: merchandise_item_repository.go:91）

```go
Joins("JOIN billings ON billings.id = billing_items.billing_id" +
    " AND billings.clinic_id = ? AND billings.deleted_at IS NULL", clinicID).
```

### 修正案

`care_plan_items` は直接 `clinic_id` を持つか、または `care_plans` 経由で確認する。
モデルを確認してから以下のいずれかで修正:

**パターンA: care_plan_items に clinic_id がある場合**
```go
Model(&model.CarePlanItem{}).
    Scopes(clinicScope(clinicID)).
    Where("hospitalization_plan_id = ? AND deleted_at IS NULL", planID).
    Count(&count)
```

**パターンB: JOIN で care_plans 経由**
```go
Model(&model.CarePlanItem{}).
    Joins("JOIN care_plans ON care_plans.id = care_plan_items.care_plan_id" +
        " AND care_plans.clinic_id = ? AND care_plans.deleted_at IS NULL", clinicID).
    Where("care_plan_items.hospitalization_plan_id = ? AND care_plan_items.deleted_at IS NULL", planID).
    Count(&count)
```

---

## 問題2（HIGH）: service — Update の input nil チェック + FindByID 欠落

**ファイル:** `backend/internal/service/hospitalization_plan_service.go:97-112`

### 現状コード

```go
func (s *hospitalizationPlanService) Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
    // ❌ input == nil チェックなし
    // ❌ FindByID 存在確認なし
    if err := validateOptionalName(input.Name); err != nil {  // input が nil だと panic
        return nil, err
    }
    // ...
}
```

### 修正案（merchandise_item_service.go 参照）

```go
func (s *hospitalizationPlanService) Update(ctx context.Context, clinicID, id uint64, input *UpdateHospitalizationPlanInput) (*model.HospitalizationPlan, error) {
    if input == nil {
        return nil, apperrors.WrapInvalidInput(ErrMsgInputNotNil)
    }
    if _, err := s.repo.FindByID(ctx, clinicID, id); err != nil {
        return nil, apperrors.Wrap(err, "failed to get hospitalization plan")
    }
    if err := validateOptionalName(input.Name); err != nil {
        return nil, err
    }
    // ...
}
```

---

## 問題3（MEDIUM）: handler — Create の Location ヘッダ欠落

**ファイル:** `backend/internal/handler/hospitalization_plan_handler.go:76`

### 現状コード

```go
c.JSON(http.StatusCreated, toHospitalizationPlanResponse(plan))
// ❌ c.Header("Location", ...) がない
```

### 修正案

```go
c.Header("Location", fmt.Sprintf("/v1/masters/hospitalization-plans/%d", plan.ID))
c.JSON(http.StatusCreated, toHospitalizationPlanResponse(plan))
```

他すべてのマスタハンドラは Location ヘッダを設定しており、hospitalization_plan のみ欠落している。

---

## 修正優先度

| 問題 | 優先度 | 理由 |
|------|--------|------|
| CountCarePlanItemsByPlanID の clinic_id フィルタ欠落 | **HIGH** | マルチテナント分離違反 |
| service Update nil/FindByID チェック欠落 | **HIGH** | panic リスク・404 応答不正 |
| handler Location ヘッダ欠落 | MEDIUM | REST 慣例違反（機能は動作する） |
