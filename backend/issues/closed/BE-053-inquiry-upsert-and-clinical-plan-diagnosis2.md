# BE-053: カルテ問診 upsert API 追加 + clinical-plan handler diagnosis_2 フィールド公開

**Status**: Closed
**Closed At**: 2026-03-23

## クローズ情報

- **変更ファイル**:
  - `backend/internal/repository/inquiry_repository.go` — 新規作成（InquiryRepository + UpsertByMedicalRecordID）
  - `backend/internal/service/inquiry_service.go` — 新規作成（InquiryService + Upsert）
  - `backend/internal/handler/inquiry_handler.go` — 新規作成（UpdateInquiry + RegisterInquiryRoutes）
  - `backend/internal/handler/medical_record_handler.go` — RegisterInquiryRoutes 呼び出し追加
  - `backend/internal/handler/clinical_plan_request.go` — diagnosis_2_category_id / diagnosis_2_name_id 追加
  - `backend/internal/handler/clinical_plan_handler.go` — Diagnosis2CategoryID / Diagnosis2NameID をサービスに渡す
  - `backend/internal/repository/repositories.go` — Inquiry フィールド + NewInquiryRepository 追加
  - `backend/internal/service/service.go` — Inquiry フィールド + NewInquiryService 追加

---

**Status**: Open
**Priority**: High
**Affects**: カルテ詳細画面の「保存」機能（問診タブ・診察プランタブ）
**Date Created**: 2026-03-23
**Related**: BUG-008, FE-001

## Summary

BUG-008 のバックエンド修正。2点の実装が必要：
1. `PATCH /v1/medical-records/:id/inquiries` エンドポイントを新規実装（upsert）
2. `clinical_plan_request.go` に `diagnosis_2_category_id` / `diagnosis_2_name_id` を追加

## 現状のコード

### 1. inquiry エンドポイントが存在しない

```go
// backend/internal/handler/medical_record_handler.go:306-321
func (h *Handler) RegisterMedicalRecordRoutes(rg *gin.RouterGroup) {
    records := rg.Group("/medical-records")
    // ... 他のルート
    h.RegisterTreatmentPlanMedicalRecordRoutes(records)
    h.RegisterClinicalPlanRoutes(records)
    // ↑ RegisterInquiryRoutes が存在しない
}

// backend/internal/handler/ に inquiry_handler.go が存在しない
// backend/internal/service/ に inquiry_service.go が存在しない（inquiry_template_service.go はある）
// backend/internal/repository/ に inquiry_repository.go が存在しない（inquiry_template_repository.go はある）
```

### 2. clinical-plan handler が diagnosis_2 フィールドを受け付けない

```go
// backend/internal/handler/clinical_plan_request.go:1-9
type updateClinicalPlanRequest struct {
    PhysicalExam        *string `json:"physical_exam"`
    DiagnosisCategoryID *uint64 `json:"diagnosis_category_id"`
    DiagnosisNameID     *uint64 `json:"diagnosis_name_id"`
    DiagnosisDetails    *string `json:"diagnosis_details"`
    TreatmentPolicy     *string `json:"treatment_policy"`
    // Diagnosis2CategoryID と Diagnosis2NameID が存在しない
}

// backend/internal/handler/clinical_plan_handler.go:42-48
input := &service.UpdateClinicalPlanInput{
    PhysicalExam:        req.PhysicalExam,
    DiagnosisCategoryID: req.DiagnosisCategoryID,
    DiagnosisNameID:     req.DiagnosisNameID,
    DiagnosisDetails:    req.DiagnosisDetails,
    TreatmentPolicy:     req.TreatmentPolicy,
    // Diagnosis2CategoryID, Diagnosis2NameID が渡されていない
}

// service 側は既に対応済み:
// backend/internal/service/clinical_plan_service.go:17-18
// Diagnosis2CategoryID *uint64
// Diagnosis2NameID     *uint64
```

### 3. Inquiry モデル（参照）

```go
// backend/internal/model/inquiry.go:28-53
type Inquiry struct {
    ID                       uint64            `gorm:"primaryKey;autoIncrement" json:"id"`
    MedicalRecordID          uint64            `gorm:"not null"                 json:"medical_record_id"`
    ChiefComplaintCategoryID *uint64           `                                json:"chief_complaint_category_id,omitempty"`
    ChiefComplaint           string            `gorm:"default:''"               json:"chief_complaint"`
    History                  string            `gorm:"default:''"               json:"history"`
    Notes                    string            `gorm:"default:''"               json:"notes"`
    // ... 他フィールド
}
```

## 必要な変更

### 1. `backend/internal/repository/inquiry_repository.go` を新規作成

```go
package repository

import (
    "context"
    "github.com/animal-ekarte/backend/internal/model"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

type InquiryRepository interface {
    UpsertByMedicalRecordID(ctx context.Context, inquiry *model.Inquiry) (*model.Inquiry, error)
    GetByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (*model.Inquiry, error)
}

type inquiryRepository struct {
    db *gorm.DB
}

func NewInquiryRepository(db *gorm.DB) InquiryRepository {
    return &inquiryRepository{db: db}
}

func (r *inquiryRepository) UpsertByMedicalRecordID(ctx context.Context, inquiry *model.Inquiry) (*model.Inquiry, error) {
    result := r.db.WithContext(ctx).
        Where(model.Inquiry{MedicalRecordID: inquiry.MedicalRecordID}).
        Assign(inquiry).
        FirstOrCreate(inquiry)
    if result.Error != nil {
        return nil, result.Error
    }
    return inquiry, nil
}

func (r *inquiryRepository) GetByMedicalRecordID(ctx context.Context, medicalRecordID uint64) (*model.Inquiry, error) {
    var inquiry model.Inquiry
    err := r.db.WithContext(ctx).
        Where("medical_record_id = ?", medicalRecordID).
        First(&inquiry).Error
    return &inquiry, err
}
```

### 2. `backend/internal/service/inquiry_service.go` を新規作成

```go
package service

import (
    "context"
    "fmt"
    "log/slog"
    "github.com/animal-ekarte/backend/internal/model"
    "github.com/animal-ekarte/backend/internal/repository"
)

type UpsertInquiryInput struct {
    MedicalRecordID          uint64
    ChiefComplaintCategoryID *uint64
    ChiefComplaint           *string
    Notes                    *string
}

type InquiryService interface {
    Upsert(ctx context.Context, input UpsertInquiryInput) (*model.Inquiry, error)
}

type inquiryService struct {
    repo repository.InquiryRepository
}

func NewInquiryService(repo repository.InquiryRepository) InquiryService {
    return &inquiryService{repo: repo}
}

func (s *inquiryService) Upsert(ctx context.Context, input UpsertInquiryInput) (*model.Inquiry, error) {
    inquiry := &model.Inquiry{
        MedicalRecordID: input.MedicalRecordID,
    }
    if input.ChiefComplaintCategoryID != nil {
        inquiry.ChiefComplaintCategoryID = input.ChiefComplaintCategoryID
    }
    if input.ChiefComplaint != nil {
        inquiry.ChiefComplaint = *input.ChiefComplaint
    }
    if input.Notes != nil {
        inquiry.Notes = *input.Notes
    }

    result, err := s.repo.UpsertByMedicalRecordID(ctx, inquiry)
    if err != nil {
        slog.ErrorContext(ctx, "failed to upsert inquiry",
            slog.Uint64("medical_record_id", input.MedicalRecordID),
            slog.String("error", err.Error()))
        return nil, fmt.Errorf("failed to upsert inquiry: %w", err)
    }
    slog.InfoContext(ctx, "inquiry upserted", slog.Uint64("medical_record_id", input.MedicalRecordID))
    return result, nil
}
```

### 3. `backend/internal/handler/inquiry_handler.go` を新規作成

```go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"

    apperrors "github.com/animal-ekarte/backend/internal/errors"
    "github.com/animal-ekarte/backend/internal/service"
)

type updateInquiryRequest struct {
    ChiefComplaint           *string `json:"chief_complaint"`
    ChiefComplaintCategoryID *uint64 `json:"chief_complaint_category_id"`
    Notes                    *string `json:"notes"`
}

func (h *Handler) UpdateInquiry(c *gin.Context) {
    medicalRecordID, err := strconv.ParseUint(c.Param("id"), 10, 64)
    if err != nil {
        RespondError(c, apperrors.WrapInvalidInput("invalid id"))
        return
    }
    var req updateInquiryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        RespondError(c, apperrors.WrapInvalidInput(parseBindError(err)))
        return
    }
    inquiry, err := h.svc.Inquiry.Upsert(c.Request.Context(), service.UpsertInquiryInput{
        MedicalRecordID:          medicalRecordID,
        ChiefComplaintCategoryID: req.ChiefComplaintCategoryID,
        ChiefComplaint:           req.ChiefComplaint,
        Notes:                    req.Notes,
    })
    if err != nil {
        RespondError(c, err)
        return
    }
    c.JSON(http.StatusOK, inquiry)
}

func (h *Handler) RegisterInquiryRoutes(rg *gin.RouterGroup) {
    rg.PATCH("/:id/inquiries", h.UpdateInquiry)
}
```

### 4. `backend/internal/handler/medical_record_handler.go` に `RegisterInquiryRoutes` を追加

```go
// Before（medical_record_handler.go:306-321）:
func (h *Handler) RegisterMedicalRecordRoutes(rg *gin.RouterGroup) {
    records := rg.Group("/medical-records")
    // ... 既存ルート
    h.RegisterClinicalPlanRoutes(records)
    h.RegisterCheckupRoutes(records)
}

// After（末尾に追加）:
func (h *Handler) RegisterMedicalRecordRoutes(rg *gin.RouterGroup) {
    records := rg.Group("/medical-records")
    // ... 既存ルート
    h.RegisterClinicalPlanRoutes(records)
    h.RegisterCheckupRoutes(records)
    h.RegisterInquiryRoutes(records)  // ← 追加
}
```

### 5. DI配線: `backend/cmd/api/main.go` に Inquiry を追加

```go
// InquiryRepository と InquiryService を svc に追加（既存の同パターンを参照）
inquiryRepo := repository.NewInquiryRepository(db)
inquiryService := service.NewInquiryService(inquiryRepo)
// h.svc.Inquiry に注入（Handler.svc 構造体への追加も必要）
```

### 6. `backend/internal/handler/clinical_plan_request.go` に diagnosis_2 フィールドを追加

```go
// Before:
type updateClinicalPlanRequest struct {
    PhysicalExam        *string `json:"physical_exam"`
    DiagnosisCategoryID *uint64 `json:"diagnosis_category_id"`
    DiagnosisNameID     *uint64 `json:"diagnosis_name_id"`
    DiagnosisDetails    *string `json:"diagnosis_details"`
    TreatmentPolicy     *string `json:"treatment_policy"`
}

// After:
type updateClinicalPlanRequest struct {
    PhysicalExam         *string `json:"physical_exam"`
    DiagnosisCategoryID  *uint64 `json:"diagnosis_category_id"`
    DiagnosisNameID      *uint64 `json:"diagnosis_name_id"`
    DiagnosisDetails     *string `json:"diagnosis_details"`
    TreatmentPolicy      *string `json:"treatment_policy"`
    Diagnosis2CategoryID *uint64 `json:"diagnosis_2_category_id"` // ← 追加
    Diagnosis2NameID     *uint64 `json:"diagnosis_2_name_id"`     // ← 追加
}
```

### 7. `backend/internal/handler/clinical_plan_handler.go` UpdateClinicalPlan に diagnosis_2 を渡す

```go
// Before（clinical_plan_handler.go:42-48）:
input := &service.UpdateClinicalPlanInput{
    PhysicalExam:        req.PhysicalExam,
    DiagnosisCategoryID: req.DiagnosisCategoryID,
    DiagnosisNameID:     req.DiagnosisNameID,
    DiagnosisDetails:    req.DiagnosisDetails,
    TreatmentPolicy:     req.TreatmentPolicy,
}

// After:
input := &service.UpdateClinicalPlanInput{
    PhysicalExam:         req.PhysicalExam,
    DiagnosisCategoryID:  req.DiagnosisCategoryID,
    DiagnosisNameID:      req.DiagnosisNameID,
    DiagnosisDetails:     req.DiagnosisDetails,
    TreatmentPolicy:      req.TreatmentPolicy,
    Diagnosis2CategoryID: req.Diagnosis2CategoryID, // ← 追加
    Diagnosis2NameID:     req.Diagnosis2NameID,     // ← 追加
}
```

## API レスポンス形式

```json
// PATCH /v1/medical-records/:id/inquiries
// 200 OK
{
  "id": 1,
  "medical_record_id": 42,
  "chief_complaint_category_id": null,
  "chief_complaint": "元気がない",
  "notes": "昨日から食欲がない",
  "created_at": "2026-03-23T10:00:00Z",
  "updated_at": "2026-03-23T10:05:00Z"
}
```

## フロントエンド影響

- FE-001 で `treatment-plans.ts` の URL と フィールド名を修正する必要がある
- inquiry upsert は新しいエンドポイントのため FE 変更不要（既存 URL と一致）

## 完了条件

- [ ] `PATCH /v1/medical-records/:id/inquiries` が 200 を返す（レコードなし → INSERT、あり → UPDATE）
- [ ] `PATCH /v1/medical-records/:id/clinical-plan` に `diagnosis_2_category_id` を送ると正しく保存される
- [ ] 既存のカルテ一覧・詳細 API が壊れていない（回帰なし）
- [ ] `docker compose exec backend go test ./... -v` が通る
