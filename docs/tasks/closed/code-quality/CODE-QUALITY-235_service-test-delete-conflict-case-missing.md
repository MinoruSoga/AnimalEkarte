# CODE-QUALITY-235: service Delete テスト — 409 Conflict ケース欠落（複数ドメイン）

## 概要

CODE-QUALITY-231（cage/insurance/occupation）と同じパターンで、
さらに複数ドメインの Delete テストが FK 依存チェック（409 Conflict）ケースを欠いている。

---

## 対象ファイルと欠落ケース

### 1. diagnosis_service_test.go

**ファイル:** `backend/internal/service/diagnosis_service_test.go:805`

**欠落ケース:** `TestDiagnosisNameService_Delete` — clinical_plans で参照されている場合の 409 Conflict

```go
{
    name: "使用中の診断名は削除できない",
    setup: func(m *mockDiagnosisNameRepository) {
        m.On("FindByID", ...).Return(&model.DiagnosisName{ID: 1}, nil)
        m.On("CountClinicalPlansByDiagnosisNameID", ..., uint64(1)).Return(int64(2), nil)
    },
    wantErr:      true,
    wantErrType:  apperrors.ErrConflict,
},
```

### 2. reservation_type_service_test.go

**ファイル:** `backend/internal/service/reservation_type_service_test.go:318`

**欠落ケース:** `TestReservationTypeService_Delete` — 予約に使用されている場合の 409 Conflict

```go
{
    name: "使用中の予約種別は削除できない",
    setup: func(m *mockReservationTypeRepository) {
        m.On("FindByID", ...).Return(&model.ReservationType{ID: 1}, nil)
        m.On("CountUsageByReservationTypeID", ..., uint64(1)).Return(int64(3), nil)
    },
    wantErr:     true,
    wantErrType: apperrors.ErrConflict,
},
```

### 3. hospitalization_plan_service_test.go

**ファイル:** `backend/internal/service/hospitalization_plan_service_test.go:271-310`

**欠落ケース:** `TestHospitalizationPlanService_Delete` — ケアプランアイテムが存在する場合の 409 Conflict

```go
{
    name: "使用中の入院プランは削除できない",
    setup: func(m *mockHospitalizationPlanRepository) {
        m.On("FindByID", ...).Return(&model.HospitalizationPlan{ID: 1}, nil)
        m.On("CountCarePlanItemsByPlanID", ..., uint64(1)).Return(int64(5), nil)
    },
    wantErr:     true,
    wantErrType: apperrors.ErrConflict,
},
```

### 4. payment_method_master_service_test.go

**ファイル:** `backend/internal/service/payment_method_master_service_test.go:238-310`

**欠落ケース:** `TestPaymentMethodMasterService_Delete` — 会計で使用中の場合の 409 Conflict

```go
{
    name: "使用中の支払方法は削除できない",
    setup: func(m *mockPaymentMethodMasterRepository) {
        m.On("FindByID", ...).Return(&model.PaymentMethodMaster{ID: 1}, nil)
        m.On("CountUsageByPaymentMethodID", ..., uint64(1)).Return(int64(10), nil)
    },
    wantErr:     true,
    wantErrType: apperrors.ErrConflict,
},
```

### 5. inquiry_template — ハンドラテストの 409 ケース欠落

**ファイル:** `backend/internal/handler/inquiry_template_handler_test.go:406-468`

**欠落ケース:** `TestInquiryTemplateHandler_Delete` — テンプレートに質問が存在する場合の 409 Conflict

```go
{
    name: "使用中のテンプレートは削除できない",
    setup: func(m *mockInquiryTemplateService) {
        m.On("Delete", ...).Return(apperrors.WrapConflict("..."))
    },
    wantStatus: http.StatusConflict,
},
```

---

## 参考（正しい実装例: medicine_service_test.go）

```go
{
    name: "使用中の薬品は削除できない",
    setup: func(m *mockMedicineRepository) {
        m.On("FindByID", ...).Return(&model.Medicine{ID: 1}, nil)
        m.On("CountUsageByMedicineID", ..., uint64(1)).Return(int64(1), nil)
    },
    wantErr:     true,
    wantErrType: apperrors.ErrConflict,
},
```

---

## 優先度

MEDIUM — 実装バグではなくテストカバレッジ欠落。
FK 依存チェックのリグレッションを検知できないリスクあり。

> CODE-QUALITY-231 はすでに cage/insurance/occupation をカバーしている。
> 本チケットは diagnosis/reservation_type/hospitalization_plan/payment_method_master/inquiry_template の追加分。
