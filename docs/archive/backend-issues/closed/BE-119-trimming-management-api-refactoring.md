# BE-119: カルテ側トリミング管理API — appointments ベースに全面書き換え

**Status**: Open
**Priority**: High
**Affects**: /v1/clinics/{id}/trimmings — handler, service, repository
**Date Created**: 2026-04-16
**Related**: TASK-002, BE-118（前提）, FE-253

## Summary

BE-118 で `trimming_records` を廃止し `appointment_trimming_details` に移行したため、
カルテ側トリミング管理 API（`/v1/clinics/{id}/trimmings`）を
`appointments` + `appointment_trimming_details` の合成ベースに全面書き換えする。
API パスとレスポンス構造はフロントエンドとの契約変更を最小化するよう設計する。

## 現状のコード

```go
// backend/internal/service/trimming_service.go:53-59
type TrimmingService interface {
    List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.TrimmingRecord, int64, error)
    GetByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error)
    Create(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.TrimmingRecord, error)
    Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.TrimmingRecord, error)
    Delete(ctx context.Context, clinicID, id uint64) error
}
// ※ 戻り値が TrimmingRecord → BE-118 で廃止

// backend/internal/repository/trimming_repository.go:13-20
type TrimmingRepository interface {
    FindAll(ctx context.Context, ...) ([]model.TrimmingRecord, int64, error)
    FindByID(ctx context.Context, clinicID, id uint64) (*model.TrimmingRecord, error)
    Create(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error
    Update(ctx context.Context, clinicID uint64, trimming *model.TrimmingRecord) error
    Delete(ctx context.Context, clinicID, id uint64) error
    SetOptions(ctx context.Context, recordID uint64, optionIDs []uint64) error
}
// ※ trimming_records テーブルに依存 → BE-118 で廃止

// backend/internal/repository/repositories.go:26
Trimming TrimmingRepository  // ← 削除対象

// backend/internal/service/service.go:19
Trimming TrimmingService     // ← インターフェース維持、実装を差し替え

// backend/internal/handler/trimming_request.go:6-22
type createTrimmingRequest struct {
    Date           *time.Time `json:"date"`        // ← DATE型。変更: start_time/end_time に
    PetID          *uint64    `json:"pet_id"`
    StaffID        *uint64    `json:"staff_id"`
    CourseID       *uint64    `json:"course_id"`
    Status         string     `json:"status"`
    StyleRequest   string     `json:"style_request"`
    BW             *float64   `json:"bw"`
    BWUnit         string     `json:"bw_unit"`
    BT             *float64   `json:"bt"`
    UsedShampoo    string     `json:"used_shampoo"`
    UsedRibbon     string     `json:"used_ribbon"`
    Remarks        string     `json:"remarks"`
    StyleImage     string     `json:"style_image"`
    CompletedImage string     `json:"completed_image"`
    OptionIDs      []uint64   `json:"option_ids"`
}

// backend/internal/handler/trimming_response.go:20-43
type trimmingResponse struct {
    ID             uint64    `json:"id"`
    Date           time.Time `json:"date"`          // ← DATE型。変更: start_time/end_time に
    Status         string    `json:"status"`        // trimming_status → reservation_status
    // ... 他フィールド
}

func toTrimmingResponse(t *model.TrimmingRecord) trimmingResponse { ... }
// ※ 引数が TrimmingRecord → Appointment に変更
```

## 必要な変更

### 1. `backend/internal/repository/` — TrimmingRepository 廃止 + AppointmentTrimmingDetailRepository 追加

#### 1-a. `trimming_repository.go` — 完全削除

ファイルごと削除する。`TrimmingCourseRepository`, `TrimmingOptionRepository` は別ファイルで存在するため影響なし。

#### 1-b. `appointment_trimming_detail_repository.go` — 新規作成

```go
package repository

import (
    "context"

    "gorm.io/gorm"

    apperrors "github.com/animal-ekarte/backend/internal/errors"
    "github.com/animal-ekarte/backend/internal/model"
)

// AppointmentTrimmingDetailRepository はトリミング詳細の永続化を担う
type AppointmentTrimmingDetailRepository interface {
    FindByAppointmentID(ctx context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error)
    Create(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
    Update(ctx context.Context, detail *model.AppointmentTrimmingDetail) error
    SetOptions(ctx context.Context, appointmentID uint64, optionIDs []uint64) error
}

type appointmentTrimmingDetailRepository struct {
    db *gorm.DB
}

func NewAppointmentTrimmingDetailRepository(db *gorm.DB) AppointmentTrimmingDetailRepository {
    return &appointmentTrimmingDetailRepository{db: db}
}

func (r *appointmentTrimmingDetailRepository) FindByAppointmentID(ctx context.Context, clinicID, appointmentID uint64) (*model.AppointmentTrimmingDetail, error) {
    var detail model.AppointmentTrimmingDetail
    err := r.db.WithContext(ctx).
        Preload("Course").
        Preload("Options").
        Where("clinic_id = ? AND appointment_id = ?", clinicID, appointmentID).
        First(&detail).Error
    if err != nil {
        return nil, apperrors.FromGORM(err, "appointment_trimming_detail", fmt.Sprintf("%d", appointmentID))
    }
    return &detail, nil
}

func (r *appointmentTrimmingDetailRepository) Create(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
    if err := r.db.WithContext(ctx).Create(detail).Error; err != nil {
        return apperrors.Wrap(err, "failed to create appointment_trimming_detail")
    }
    return nil
}

func (r *appointmentTrimmingDetailRepository) Update(ctx context.Context, detail *model.AppointmentTrimmingDetail) error {
    if err := r.db.WithContext(ctx).Save(detail).Error; err != nil {
        return apperrors.Wrap(err, "failed to update appointment_trimming_detail")
    }
    return nil
}

// SetOptions はトリミングオプションを全置換する（トランザクション内）
func (r *appointmentTrimmingDetailRepository) SetOptions(ctx context.Context, appointmentID uint64, optionIDs []uint64) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        if err := tx.Unscoped().
            Where("appointment_id = ?", appointmentID).
            Delete(&model.AppointmentTrimmingOption{}).Error; err != nil {
            return apperrors.Wrap(err, "failed to clear trimming options")
        }
        if len(optionIDs) == 0 {
            return nil
        }
        opts := make([]model.AppointmentTrimmingOption, 0, len(optionIDs))
        for i, oid := range optionIDs {
            opts = append(opts, model.AppointmentTrimmingOption{
                AppointmentID: appointmentID,
                OptionID:      oid,
                SortOrder:     i,
            })
        }
        return apperrors.Wrap(tx.Create(&opts).Error, "failed to create trimming options")
    })
}
```

#### 1-c. `repositories.go` — Trimming 削除、AppointmentTrimmingDetail 追加

変更箇所:

```go
// 削除
Trimming TrimmingRepository  // line 26

// 追加
AppointmentTrimmingDetail AppointmentTrimmingDetailRepository

// 削除
Trimming: NewTrimmingRepository(db),  // line 92

// 追加
AppointmentTrimmingDetail: NewAppointmentTrimmingDetailRepository(db),
```

#### 1-d. `appointment_repository.go` — `FindAllByCategory` メソッド追加

既存の `FindAll` は変更せず、トリミング専用のフィルタメソッドを追加する:

```go
// ReservationRepository インターフェースに追加
FindAllByCategory(ctx context.Context, clinicID uint64, category string, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Appointment, int64, error)

// 実装
func (r *reservationRepository) FindAllByCategory(ctx context.Context, clinicID uint64, category string, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Appointment, int64, error) {
    var appointments []model.Appointment
    var total int64

    q := r.db.WithContext(ctx).
        Joins("JOIN reservation_types ON reservation_types.id = appointments.reservation_type_id").
        Where("appointments.clinic_id = ?", clinicID).
        Where("reservation_types.category = ?", category).
        Where("appointments.deleted_at IS NULL").
        Preload("Pet.Owner").
        Preload("Pet.AnimalSpecies").
        Preload("Doctor").
        Preload("ReservationType").
        Preload("TrimmingDetail.Course").
        Preload("TrimmingDetail.Options")

    if petID != nil {
        q = q.Where("appointments.pet_id = ?", *petID)
    }
    if ownerID != nil {
        q = q.Joins("JOIN pets ON pets.id = appointments.pet_id AND pets.deleted_at IS NULL").
            Where("pets.owner_id = ?", *ownerID)
    }
    if startDate != nil {
        q = q.Where("DATE(appointments.start_time AT TIME ZONE 'Asia/Tokyo') >= ?", *startDate)
    }
    if endDate != nil {
        q = q.Where("DATE(appointments.start_time AT TIME ZONE 'Asia/Tokyo') <= ?", *endDate)
    }

    if err := q.Model(&model.Appointment{}).Count(&total).Error; err != nil {
        return nil, 0, apperrors.Wrap(err, "failed to count trimming appointments")
    }
    if err := q.Order("appointments.start_time DESC").
        Offset((page - 1) * limit).Limit(limit).
        Find(&appointments).Error; err != nil {
        return nil, 0, apperrors.Wrap(err, "failed to list trimming appointments")
    }
    return appointments, total, nil
}
```

**注意**: `ownerID` でフィルタする場合、`pets` の JOIN が重複する可能性がある（既に `Preload("Pet")` がある）。
JOIN の別名を使うか、サブクエリで解決すること。実装時に EXPLAIN ANALYZE で確認すること。

### 2. `backend/internal/service/trimming_service.go` — 完全書き換え

```go
package service

// CreateTrimmingInput はトリミング予約作成の入力
type CreateTrimmingInput struct {
    ReservationTypeID uint64                 // category='trimming' の予約区分
    StartTime         time.Time
    EndTime           time.Time
    PetID             *uint64
    StaffID           *uint64                // appointments.doctor_id に対応
    Status            model.ReservationStatus // デフォルト: pending
    // 以下はappointment_trimming_detailsに入る
    CourseID          *uint64
    StyleRequest      string
    BodyWeight        *float64
    BWUnit            model.BodyWeightUnit   // デフォルト: Kg
    BodyTemperature   *float64
    UsedShampoo       string
    UsedRibbon        string
    Remarks           string
    StyleImage        string
    CompletedImage    string
    OptionIDs         []uint64
}

// UpdateTrimmingInput はトリミング予約更新の入力（nil = 変更なし）
type UpdateTrimmingInput struct {
    StartTime         *time.Time
    EndTime           *time.Time
    PetID             *uint64
    StaffID           *uint64
    Status            *model.ReservationStatus
    // trimming_detail フィールド
    CourseID          *uint64
    StyleRequest      *string
    BodyWeight        **float64
    BWUnit            *model.BodyWeightUnit
    BodyTemperature   **float64
    UsedShampoo       *string
    UsedRibbon        *string
    Remarks           *string
    StyleImage        *string
    CompletedImage    *string
    OptionIDs         *[]uint64              // nil=変更なし、non-nil（空スライス含む）=全置換
}

// TrimmingService はトリミング管理のビジネスロジック（appointments + trimming_details を合成）
type TrimmingService interface {
    List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Appointment, int64, error)
    GetByID(ctx context.Context, clinicID, id uint64) (*model.Appointment, error)
    Create(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.Appointment, error)
    Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.Appointment, error)
    Delete(ctx context.Context, clinicID, id uint64) error
}

type trimmingService struct {
    reservationRepo  repository.ReservationRepository
    trimmingRepo     repository.AppointmentTrimmingDetailRepository
    reservationTypes repository.ReservationTypeRepository  // category 検証用
}

func NewTrimmingService(
    reservationRepo repository.ReservationRepository,
    trimmingRepo repository.AppointmentTrimmingDetailRepository,
    reservationTypes repository.ReservationTypeRepository,
) TrimmingService {
    return &trimmingService{
        reservationRepo:  reservationRepo,
        trimmingRepo:     trimmingRepo,
        reservationTypes: reservationTypes,
    }
}

// List は trimming カテゴリの appointment を一覧取得する
func (s *trimmingService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Appointment, int64, error) {
    return s.reservationRepo.FindAllByCategory(ctx, clinicID, string(model.ReservationTypeCategoryTrimming), petID, ownerID, startDate, endDate, page, limit)
}

// GetByID は指定 appointment を取得し category='trimming' を検証する
func (s *trimmingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Appointment, error) {
    appt, err := s.reservationRepo.FindByID(ctx, clinicID, id)
    if err != nil {
        return nil, err
    }
    if appt.ReservationType == nil || appt.ReservationType.Category != model.ReservationTypeCategoryTrimming {
        return nil, apperrors.ErrNotFound
    }
    return appt, nil
}

// Create は appointment + trimming_detail を同一トランザクションで作成する
func (s *trimmingService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.Appointment, error) {
    // ReservationType の category 検証
    rt, err := s.reservationTypes.FindByID(ctx, clinicID, input.ReservationTypeID)
    if err != nil {
        return nil, err
    }
    if rt.Category != model.ReservationTypeCategoryTrimming {
        return nil, apperrors.WrapInvalidInput("指定された予約区分はトリミングカテゴリではありません")
    }

    if input.Status == "" {
        input.Status = model.ReservationStatusPending
    }
    if input.BWUnit == "" {
        input.BWUnit = model.BodyWeightUnitKg
    }

    appt := &model.Appointment{
        ClinicID:          clinicID,
        ReservationTypeID: input.ReservationTypeID,
        StartTime:         input.StartTime,
        EndTime:           input.EndTime,
        PetID:             input.PetID,
        DoctorID:          input.StaffID,
        Status:            input.Status,
        Source:            model.ReservationSourceManual,
    }
    if err := s.reservationRepo.Create(ctx, appt); err != nil {
        return nil, err
    }

    detail := &model.AppointmentTrimmingDetail{
        ClinicID:        clinicID,
        AppointmentID:   appt.ID,
        CourseID:        input.CourseID,
        StyleRequest:    input.StyleRequest,
        BodyWeight:      input.BodyWeight,
        BWUnit:          input.BWUnit,
        BodyTemperature: input.BodyTemperature,
        UsedShampoo:     input.UsedShampoo,
        UsedRibbon:      input.UsedRibbon,
        Remarks:         input.Remarks,
        StyleImage:      input.StyleImage,
        CompletedImage:  input.CompletedImage,
    }
    if err := s.trimmingRepo.Create(ctx, detail); err != nil {
        return nil, err
    }
    if len(input.OptionIDs) > 0 {
        if err := s.trimmingRepo.SetOptions(ctx, appt.ID, input.OptionIDs); err != nil {
            return nil, err
        }
    }

    // Reload with all relations
    return s.reservationRepo.FindByID(ctx, clinicID, appt.ID)
}

// Update は appointment と trimming_detail を更新する
func (s *trimmingService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.Appointment, error) {
    // category 検証を兼ねた GetByID
    appt, err := s.GetByID(ctx, clinicID, id)
    if err != nil {
        return nil, err
    }

    // appointment フィールド更新
    apptFields := map[string]any{}
    if input.StartTime != nil { apptFields["start_time"] = *input.StartTime }
    if input.EndTime != nil   { apptFields["end_time"] = *input.EndTime }
    if input.PetID != nil     { apptFields["pet_id"] = *input.PetID }
    if input.StaffID != nil   { apptFields["doctor_id"] = *input.StaffID }
    if input.Status != nil    { apptFields["status"] = *input.Status }
    if len(apptFields) > 0 {
        if _, err := s.reservationRepo.UpdateFields(ctx, clinicID, id, apptFields); err != nil {
            return nil, err
        }
    }

    // trimming_detail 更新（既存detailがない場合は新規作成）
    detail := appt.TrimmingDetail
    if detail == nil {
        detail = &model.AppointmentTrimmingDetail{
            ClinicID:      clinicID,
            AppointmentID: id,
        }
    }
    if input.CourseID != nil        { detail.CourseID = input.CourseID }
    if input.StyleRequest != nil    { detail.StyleRequest = *input.StyleRequest }
    if input.BodyWeight != nil      { detail.BodyWeight = *input.BodyWeight }
    if input.BWUnit != nil          { detail.BWUnit = *input.BWUnit }
    if input.BodyTemperature != nil { detail.BodyTemperature = *input.BodyTemperature }
    if input.UsedShampoo != nil     { detail.UsedShampoo = *input.UsedShampoo }
    if input.UsedRibbon != nil      { detail.UsedRibbon = *input.UsedRibbon }
    if input.Remarks != nil         { detail.Remarks = *input.Remarks }
    if input.StyleImage != nil      { detail.StyleImage = *input.StyleImage }
    if input.CompletedImage != nil  { detail.CompletedImage = *input.CompletedImage }

    if appt.TrimmingDetail == nil {
        if err := s.trimmingRepo.Create(ctx, detail); err != nil {
            return nil, err
        }
    } else {
        if err := s.trimmingRepo.Update(ctx, detail); err != nil {
            return nil, err
        }
    }

    if input.OptionIDs != nil {
        if err := s.trimmingRepo.SetOptions(ctx, id, *input.OptionIDs); err != nil {
            return nil, err
        }
    }

    return s.reservationRepo.FindByID(ctx, clinicID, id)
}

// Delete は appointment を論理削除する（appointment_trimming_details は CASCADE で削除）
func (s *trimmingService) Delete(ctx context.Context, clinicID, id uint64) error {
    if _, err := s.GetByID(ctx, clinicID, id); err != nil {
        return err
    }
    return s.reservationRepo.Delete(ctx, clinicID, id)
}
```

### 3. `backend/internal/service/service.go` — NewTrimmingService 引数変更

```go
// Before:
Trimming: NewTrimmingService(repos.Trimming),

// After:
Trimming: NewTrimmingService(repos.Reservation, repos.AppointmentTrimmingDetail, repos.ReservationType),
```

### 4. `backend/internal/handler/trimming_request.go` — 全面書き換え

```go
package handler

import "time"

// createTrimmingRequest はトリミング作成のバインド struct
type createTrimmingRequest struct {
    ReservationTypeID uint64     `json:"reservation_type_id" binding:"required"` // category='trimming' のID
    StartTime         time.Time  `json:"start_time"          binding:"required"`  // ISO 8601
    EndTime           time.Time  `json:"end_time"            binding:"required"`
    PetID             *uint64    `json:"pet_id"`
    StaffID           *uint64    `json:"staff_id"`
    Status            string     `json:"status"`
    // trimming detail
    CourseID       *uint64  `json:"course_id"`
    StyleRequest   string   `json:"style_request"`
    BW             *float64 `json:"bw"`
    BWUnit         string   `json:"bw_unit"`
    BT             *float64 `json:"bt"`
    UsedShampoo    string   `json:"used_shampoo"`
    UsedRibbon     string   `json:"used_ribbon"`
    Remarks        string   `json:"remarks"`
    StyleImage     string   `json:"style_image"`
    CompletedImage string   `json:"completed_image"`
    OptionIDs      []uint64 `json:"option_ids"`
}

// updateTrimmingRequest はトリミング更新のバインド struct
type updateTrimmingRequest struct {
    StartTime  *time.Time `json:"start_time"`
    EndTime    *time.Time `json:"end_time"`
    PetID      *uint64    `json:"pet_id"`
    StaffID    *uint64    `json:"staff_id"`
    Status     *string    `json:"status"`
    // trimming detail
    CourseID       *uint64   `json:"course_id"`
    StyleRequest   *string   `json:"style_request"`
    BW             **float64 `json:"bw"`
    BWUnit         *string   `json:"bw_unit"`
    BT             **float64 `json:"bt"`
    UsedShampoo    *string   `json:"used_shampoo"`
    UsedRibbon     *string   `json:"used_ribbon"`
    Remarks        *string   `json:"remarks"`
    StyleImage     *string   `json:"style_image"`
    CompletedImage *string   `json:"completed_image"`
    // nil = 変更なし、non-nil（空スライス含む）= 全置換
    OptionIDs *[]uint64 `json:"option_ids"`
}
```

### 5. `backend/internal/handler/trimming_response.go` — 全面書き換え

```go
package handler

import (
    "time"
    "github.com/animal-ekarte/backend/internal/model"
)

// trimmingResponse はトリミング管理API の統一レスポンス
type trimmingResponse struct {
    // --- appointment フィールド ---
    ID                uint64    `json:"id"`
    ClinicID          uint64    `json:"clinic_id"`
    ReservationTypeID uint64    `json:"reservation_type_id"`
    StartTime         time.Time `json:"start_time"`
    EndTime           time.Time `json:"end_time"`
    PetID             *uint64   `json:"pet_id,omitempty"`
    StaffID           *uint64   `json:"staff_id,omitempty"`  // = doctor_id
    Status            string    `json:"status"`              // reservation_status
    Source            string    `json:"source"`
    // --- trimming_detail フィールド ---
    CourseID       *uint64  `json:"course_id,omitempty"`
    StyleRequest   string   `json:"style_request"`
    BW             *float64 `json:"bw,omitempty"`
    BWUnit         string   `json:"bw_unit"`
    BT             *float64 `json:"bt,omitempty"`
    UsedShampoo    string   `json:"used_shampoo"`
    UsedRibbon     string   `json:"used_ribbon"`
    Remarks        string   `json:"remarks"`
    StyleImage     string   `json:"style_image"`
    CompletedImage string   `json:"completed_image"`
    // ---
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    // --- Relations ---
    Pet     *petSummaryResponse            `json:"pet,omitempty"`
    Staff   *staffSummaryResponse          `json:"staff,omitempty"`
    Course  *trimmingCourseSummaryResponse `json:"course,omitempty"`
    Options []trimmingOptionSummaryResponse `json:"options"`
}

func toTrimmingResponse(appt *model.Appointment) trimmingResponse {
    options := make([]trimmingOptionSummaryResponse, 0)
    var courseID *uint64
    var styleRequest, usedShampoo, usedRibbon, remarks, styleImage, completedImage string
    var bw *float64
    var bwUnit string
    var bt *float64
    var course *trimmingCourseSummaryResponse

    if appt.TrimmingDetail != nil {
        d := appt.TrimmingDetail
        courseID = d.CourseID
        styleRequest = d.StyleRequest
        bw = d.BodyWeight
        bwUnit = string(d.BWUnit)
        bt = d.BodyTemperature
        usedShampoo = d.UsedShampoo
        usedRibbon = d.UsedRibbon
        remarks = d.Remarks
        styleImage = d.StyleImage
        completedImage = d.CompletedImage
        for i := range d.Options {
            options = append(options, trimmingOptionSummaryResponse{
                ID:   d.Options[i].ID,
                Name: d.Options[i].Name,
            })
        }
        if d.Course != nil {
            var price int64
            if d.Course.Price != nil {
                price = *d.Course.Price
            }
            course = &trimmingCourseSummaryResponse{
                ID: d.Course.ID, Name: d.Course.Name, Price: price,
            }
        }
    }

    return trimmingResponse{
        ID:                appt.ID,
        ClinicID:          appt.ClinicID,
        ReservationTypeID: appt.ReservationTypeID,
        StartTime:         appt.StartTime,
        EndTime:           appt.EndTime,
        PetID:             appt.PetID,
        StaffID:           appt.DoctorID,
        Status:            string(appt.Status),
        Source:            string(appt.Source),
        CourseID:          courseID,
        StyleRequest:      styleRequest,
        BW:                bw,
        BWUnit:            bwUnit,
        BT:                bt,
        UsedShampoo:       usedShampoo,
        UsedRibbon:        usedRibbon,
        Remarks:           remarks,
        StyleImage:        styleImage,
        CompletedImage:    completedImage,
        CreatedAt:         appt.CreatedAt,
        UpdatedAt:         appt.UpdatedAt,
        Pet:               toPetSummary(appt.Pet),
        Staff:             toStaffSummary(appt.Doctor),
        Course:            course,
        Options:           options,
    }
}
```

### 6. `backend/internal/handler/trimming_handler.go` — 引数型変更

`CreateTrimming`, `UpdateTrimming`, `ListTrimmings`, `GetTrimming` の各ハンドラで:
- `req.Date` → `req.StartTime` / `req.EndTime`
- `req.Status` → `model.ReservationStatus(req.Status)`（TrimmingStatus の参照を削除）
- `toTrimmingResponse(record)` の引数型が `*model.TrimmingRecord` → `*model.Appointment` に変わるため、コンパイルエラーが消えるだけで呼び出し自体は変更なし

`CreateTrimming` の input 組み立て:
```go
input := &service.CreateTrimmingInput{
    ReservationTypeID: req.ReservationTypeID,
    StartTime:         req.StartTime,
    EndTime:           req.EndTime,
    PetID:             req.PetID,
    StaffID:           req.StaffID,
    CourseID:          req.CourseID,
    StyleRequest:      req.StyleRequest,
    BodyWeight:        req.BW,
    BodyTemperature:   req.BT,
    UsedShampoo:       req.UsedShampoo,
    UsedRibbon:        req.UsedRibbon,
    Remarks:           req.Remarks,
    StyleImage:        req.StyleImage,
    CompletedImage:    req.CompletedImage,
    OptionIDs:         req.OptionIDs,
}
if req.Status != "" {
    input.Status = model.ReservationStatus(req.Status)
}
if req.BWUnit != "" {
    input.BWUnit = model.BodyWeightUnit(req.BWUnit)
}
```

## API レスポンス変更点（フロントエンドへの影響）

| フィールド | 変更前 | 変更後 | 影響 |
|-----------|--------|--------|------|
| `date` | `"2026-05-01"` (DATE) | 廃止 | FE-253 で削除 |
| `start_time` | なし | `"2026-05-01T09:00:00Z"` (ISO8601) | FE-253 で追加 |
| `end_time` | なし | `"2026-05-01T10:00:00Z"` | FE-253 で追加 |
| `status` | `"reserved"` / `"in_progress"` / `"completed"` | `"pending"` / `"confirmed"` / `"in_consultation"` / `"completed"` / `"cancelled"` | FE-253でマッピング変更 |
| `reservation_type_id` | なし（course_id のみ） | 追加 | FE-253 で追加 |
| `source` | なし | `"manual"` / `"line"` | FE-253 で追加 |

## 完了条件

- [ ] `trimming_repository.go` が削除されている
- [ ] `appointment_trimming_detail_repository.go` が新規作成されている
- [ ] `repositories.go` の `Trimming TrimmingRepository` が `AppointmentTrimmingDetail AppointmentTrimmingDetailRepository` に変更されている
- [ ] `appointment_repository.go` に `FindAllByCategory` メソッドが追加されている
- [ ] `trimming_service.go` の `TrimmingService` インターフェースが `[]model.Appointment` を返すよう変更されている
- [ ] `service.go` の `NewTrimmingService` 引数が更新されている
- [ ] `trimming_request.go` が `start_time` / `end_time` ベースに変更されている
- [ ] `trimming_response.go` が `*model.Appointment` を受け取るよう変更されている
- [ ] `docker compose exec backend go build ./...` が通る
- [ ] `docker compose exec backend go test ./...` が通る（既存テストの更新を含む）
- [ ] `GET /v1/clinics/{id}/trimmings` が appointments ベースで動作する
- [ ] `POST /v1/clinics/{id}/trimmings` が appointment + trimming_detail を同時作成する
- [ ] `PATCH /v1/clinics/{id}/trimmings/:id` が appointment + trimming_detail を更新する
- [ ] `DELETE /v1/clinics/{id}/trimmings/:id` が appointment を論理削除する
