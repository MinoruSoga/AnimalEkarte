# BE-083: バイタル記録 POST が 500 → pet_id 未セット & weight_unit ENUM 制約違反

**Status**: Open
**Priority**: High
**Affects**: internal/handler/vital_request.go, internal/handler/vital_handler.go, internal/handler/record_image_handler.go, internal/service/vital_service.go, internal/handler/vital_response.go
**Date Created**: 2026-03-29
**Related**: BUG-044（BUG-022 の根本原因も同一）

---

## Summary

`POST /api/v1/medical-records/:id/vitals` が HTTP 500 を返す原因が **2つ** 存在する。

### 原因 1: `PetID` がゼロ値 → FK 制約違反（旧 BUG-022 と同根）

`vitalService.Create()` が `model.VitalRecord` の `PetID` フィールドをセットしていない。
Go の zero value `uint64(0)` で INSERT されるため `vital_records_pet_id_fkey` 制約違反が発生する。

`PetID` は `medical_record.pet_id` から解決すべきだが、
`verifyMedicalRecordOwnership()` は medical record を取得後に破棄している（bool のみ返す）。

### 原因 2: `WeightUnit` が空文字 → ENUM 制約違反（BUG-044）

`createVitalRequest` / `CreateVitalInput` に `weight_unit` フィールドが存在しない。
GORM が空文字 `""` で INSERT するため `body_weight_unit` ENUM 制約違反が発生する。

---

## 実装手順

### 1. `record_image_handler.go` の `verifyMedicalRecordOwnership` を medical record 返却型に変更

```go
// 変更前: bool のみ返す
func (h *Handler) verifyMedicalRecordOwnership(c *gin.Context, clinicID, medicalRecordID uint64) bool {
    if _, err := h.svc.MedicalRecord.GetByID(c.Request.Context(), clinicID, medicalRecordID); err != nil {
        RespondError(c, err)
        return false
    }
    return true
}

// 変更後: *model.MedicalRecord も返す
func (h *Handler) verifyMedicalRecordOwnership(c *gin.Context, clinicID, medicalRecordID uint64) (*model.MedicalRecord, bool) {
    mr, err := h.svc.MedicalRecord.GetByID(c.Request.Context(), clinicID, medicalRecordID)
    if err != nil {
        RespondError(c, err)
        return nil, false
    }
    return mr, true
}
```

**注意**: `verifyMedicalRecordOwnership` を呼び出している全ハンドラーの呼び出し側を更新すること。
`_, ok := h.verifyMedicalRecordOwnership(...)` パターンで既存呼び出しを最小変更で対応可能。

### 2. `vital_handler.go` の CreateVital で PetID を取得・渡す

```go
func (h *Handler) CreateVital(c *gin.Context) {
    clinicID, ok := extractClinicID(c)
    if !ok {
        return
    }
    medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid medical record id"))
        return
    }
    mr, ok := h.verifyMedicalRecordOwnership(c, clinicID, medicalRecordID)
    if !ok {
        return
    }

    var req createVitalRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": parseBindError(err)})
        return
    }

    input := &service.CreateVitalInput{
        PetID:           mr.PetID,              // 追加: medical record から pet_id を解決
        RecordedAt:      req.RecordedAt,
        StaffID:         req.StaffID,
        Temperature:     req.Temperature,
        HeartRate:       req.HeartRate,
        RespirationRate: req.RespirationRate,
        Weight:          req.Weight,
        WeightUnit:      toBodyWeightUnit(req.WeightUnit),  // 追加
        Notes:           req.Notes,
    }
    // ...
}
```

### 3. `vital_request.go` に `weight_unit` を追加

```go
type createVitalRequest struct {
    RecordedAt      time.Time  `json:"recorded_at"  binding:"required"`
    StaffID         *uint64    `json:"staff_id"`
    Temperature     *float64   `json:"temperature"`
    HeartRate       *int       `json:"heart_rate"`
    RespirationRate *int       `json:"respiration_rate"`
    Weight          *float64   `json:"weight"`
    WeightUnit      *string    `json:"weight_unit"`   // 追加: "Kg" or "g"
    Notes           string     `json:"notes"`
}

type updateVitalRequest struct {
    RecordedAt      *time.Time `json:"recorded_at"`
    StaffID         *uint64    `json:"staff_id"`
    Temperature     *float64   `json:"temperature"`
    HeartRate       *int       `json:"heart_rate"`
    RespirationRate *int       `json:"respiration_rate"`
    Weight          *float64   `json:"weight"`
    WeightUnit      *string    `json:"weight_unit"`   // 追加
    Notes           *string    `json:"notes"`
}
```

### 4. `vital_service.go` の `CreateVitalInput` / `UpdateVitalInput` に追加

```go
type CreateVitalInput struct {
    PetID           uint64                 // 追加: 必須
    RecordedAt      time.Time
    StaffID         *uint64
    Temperature     *float64
    HeartRate       *int
    RespirationRate *int
    Weight          *float64
    WeightUnit      *model.BodyWeightUnit  // 追加
    Notes           string
}

type UpdateVitalInput struct {
    RecordedAt      *time.Time
    StaffID         *uint64
    Temperature     *float64
    HeartRate       *int
    RespirationRate *int
    Weight          *float64
    WeightUnit      *model.BodyWeightUnit  // 追加
    Notes           *string
}
```

### 5. `vital_service.go` の Create で PetID と WeightUnit をセット

```go
func (s *vitalService) Create(ctx context.Context, medicalRecordID uint64, input *CreateVitalInput) (*model.VitalRecord, error) {
    vital := &model.VitalRecord{
        PetID:           input.PetID,                          // 追加
        MedicalRecordID: &medicalRecordID,
        RecordedAt:      input.RecordedAt,
        StaffID:         input.StaffID,
        Temperature:     input.Temperature,
        HeartRate:       input.HeartRate,
        RespirationRate: input.RespirationRate,
        Weight:          input.Weight,
        WeightUnit:      weightUnitOrDefault(input.WeightUnit), // 追加
        Notes:           input.Notes,
    }
    // ...
}

func weightUnitOrDefault(u *model.BodyWeightUnit) model.BodyWeightUnit {
    if u != nil {
        return *u
    }
    return model.BodyWeightUnitKg
}
```

Update の `buildVitalUpdateFields()` にも `weight_unit` を追加する。

### 6. `vital_response.go` に WeightUnit を追加

```go
type vitalResponse struct {
    ID              string                `json:"id"`
    MedicalRecordID *string               `json:"medical_record_id,omitempty"`
    RecordedAt      time.Time             `json:"recorded_at"`
    StaffID         *string               `json:"staff_id,omitempty"`
    Temperature     *float64              `json:"temperature,omitempty"`
    HeartRate       *int                  `json:"heart_rate,omitempty"`
    RespirationRate *int                  `json:"respiration_rate,omitempty"`
    Weight          *float64              `json:"weight,omitempty"`
    WeightUnit      model.BodyWeightUnit  `json:"weight_unit"`   // 追加
    Notes           string                `json:"notes"`
    CreatedAt       time.Time             `json:"created_at"`
}
```

`toVitalResponse()` に `WeightUnit: v.WeightUnit` を追加。

### 7. `toBodyWeightUnit` ヘルパー（`vital_handler.go`）

```go
func toBodyWeightUnit(s *string) *model.BodyWeightUnit {
    if s == nil {
        return nil
    }
    u := model.BodyWeightUnit(*s)
    return &u
}
```

---

## 確認コマンド

```bash
make reset
docker compose exec backend go build ./...
docker compose exec backend go test ./... -v
```

---

## 受入条件

- [ ] `POST /api/v1/medical-records/:id/vitals` → 201 返る（500 解消）
- [ ] レスポンスに `weight_unit` フィールドが含まれる
- [ ] `weight_unit` 未送信時は `"Kg"` デフォルト適用
- [ ] `weight_unit: "g"` 送信時は `"g"` が保存される
- [ ] `verifyMedicalRecordOwnership` の戻り値変更により既存ハンドラーがコンパイルエラーなし
- [ ] `docker compose exec backend go build ./...` 成功
- [ ] `docker compose exec backend go test ./... -v` 成功
