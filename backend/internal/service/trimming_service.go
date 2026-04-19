// Package service provides business logic implementations for Trimming entity.
package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateTrimmingInput はトリミング予約作成の入力DTO（appointments ベース, BE-119）
type CreateTrimmingInput struct {
	ReservationTypeID uint64
	StartTime         time.Time
	EndTime           time.Time
	PetID             *uint64
	StaffID           *uint64 // appointments.doctor_id にマップ
	Status            model.ReservationStatus
	// トリミング詳細（appointment_trimming_details）
	CourseID        *uint64
	StyleRequest    string
	BodyWeight      *float64
	BWUnit          model.BodyWeightUnit
	BodyTemperature *float64
	UsedShampoo     string
	UsedRibbon      string
	Remarks         string
	StyleImage      string
	CompletedImage  string
	OptionIDs       []uint64
}

// UpdateTrimmingInput はトリミング予約部分更新の入力DTO。nil = 未送信フィールド。
// OptionIDs: nil = 変更なし、non-nil（空スライス含む）= 全置換
type UpdateTrimmingInput struct {
	StartTime       *time.Time
	EndTime         *time.Time
	PetID           *uint64
	StaffID         *uint64
	Status          *model.ReservationStatus
	CourseID        *uint64
	StyleRequest    *string
	BodyWeight      *float64
	BWUnit          *model.BodyWeightUnit
	BodyTemperature *float64
	UsedShampoo     *string
	UsedRibbon      *string
	Remarks         *string
	StyleImage      *string
	CompletedImage  *string
	OptionIDs       *[]uint64
}

// TrimmingService はトリミング管理のビジネスロジックインターフェース（BE-119）
type TrimmingService interface {
	List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	Create(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.Reservation, error)
	Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type trimmingService struct {
	reservation    repository.ReservationRepository
	trimmingDetail repository.AppointmentTrimmingDetailRepository
	transactor     repository.Transactor
}

func NewTrimmingService(
	reservation repository.ReservationRepository,
	trimmingDetail repository.AppointmentTrimmingDetailRepository,
	transactor repository.Transactor,
) TrimmingService {
	return &trimmingService{
		reservation:    reservation,
		trimmingDetail: trimmingDetail,
		transactor:     transactor,
	}
}

func (s *trimmingService) List(ctx context.Context, clinicID uint64, petID, ownerID *uint64, startDate, endDate *string, page, limit int) ([]model.Reservation, int64, error) {
	items, total, err := s.reservation.FindAllByCategory(ctx, clinicID, model.ReservationTypeCategoryTrimming, petID, ownerID, startDate, endDate, page, limit)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list trimming appointments")
	}
	return items, total, nil
}

func (s *trimmingService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error) {
	appt, err := s.reservation.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get trimming appointment")
	}
	// FindByID は汎用メソッドのため TrimmingDetail をプリロードしない。
	// 別途 AppointmentTrimmingDetailRepository から取得する。
	// NotFound は正常系（trimming_detail 未作成の予約もありうる）として無視し、
	// それ以外の DB エラー（接続断・タイムアウト等）は伝播させる。
	detail, detailErr := s.trimmingDetail.FindByAppointmentID(ctx, clinicID, id)
	if detailErr == nil {
		appt.TrimmingDetail = detail
	} else if !apperrors.IsNotFound(detailErr) {
		return nil, apperrors.Wrap(detailErr, "failed to get trimming detail")
	}
	return appt, nil
}

func (s *trimmingService) Create(ctx context.Context, clinicID uint64, input *CreateTrimmingInput) (*model.Reservation, error) {
	status := model.ReservationStatusPending
	if input.Status != "" {
		status = input.Status
	}
	bwUnit := model.BodyWeightUnitKg
	if input.BWUnit != "" {
		bwUnit = input.BWUnit
	}

	var apptID uint64
	// appointment → trimming_detail → options の3書き込みを単一トランザクションで実行する。
	// 中間でエラーが発生した場合はロールバックされ、孤立レコードは残らない。
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		appt := &model.Reservation{
			ClinicID:          clinicID,
			ReservationTypeID: input.ReservationTypeID,
			StartTime:         input.StartTime,
			EndTime:           input.EndTime,
			PetID:             input.PetID,
			DoctorID:          input.StaffID,
			Status:            status,
			Source:            model.ReservationSourceManual,
		}
		if err := s.reservation.Create(txCtx, appt); err != nil {
			return apperrors.Wrap(err, "failed to create trimming appointment")
		}

		detail := &model.AppointmentTrimmingDetail{
			ClinicID:        clinicID,
			AppointmentID:   appt.ID,
			CourseID:        input.CourseID,
			StyleRequest:    input.StyleRequest,
			BodyWeight:      input.BodyWeight,
			BWUnit:          bwUnit,
			BodyTemperature: input.BodyTemperature,
			UsedShampoo:     input.UsedShampoo,
			UsedRibbon:      input.UsedRibbon,
			Remarks:         input.Remarks,
			StyleImage:      input.StyleImage,
			CompletedImage:  input.CompletedImage,
		}
		if err := s.trimmingDetail.Create(txCtx, detail); err != nil {
			return apperrors.Wrap(err, "failed to create trimming detail")
		}
		if len(input.OptionIDs) > 0 {
			if err := s.trimmingDetail.SetOptions(txCtx, appt.ID, input.OptionIDs); err != nil {
				return apperrors.Wrap(err, "failed to set trimming options")
			}
		}

		apptID = appt.ID
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "trimming appointment created",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("appointment_id", apptID))

	return s.GetByID(ctx, clinicID, apptID)
}

func (s *trimmingService) Update(ctx context.Context, clinicID, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error) {
	if _, err := s.reservation.FindByID(ctx, clinicID, id); err != nil {
		return nil, apperrors.Wrap(err, "failed to get trimming appointment")
	}

	apptFields := map[string]any{}
	if input.StartTime != nil {
		apptFields["start_time"] = *input.StartTime
	}
	if input.EndTime != nil {
		apptFields["end_time"] = *input.EndTime
	}
	if input.PetID != nil {
		apptFields["pet_id"] = *input.PetID
	}
	if input.StaffID != nil {
		apptFields["doctor_id"] = *input.StaffID
	}
	if input.Status != nil {
		apptFields["status"] = *input.Status
	}

	// appointment → trimming_detail → options の更新を単一トランザクションで実行する。
	// appointments が更新済みで trimming_detail が失敗すると不整合が生じるため、
	// Create と対称なアトミック性を保証する。
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if len(apptFields) > 0 {
			if _, err := s.reservation.UpdateFields(txCtx, clinicID, id, apptFields); err != nil {
				return apperrors.Wrap(err, "failed to update trimming appointment")
			}
		}

		// trimming detail の更新
		detail, err := s.trimmingDetail.FindByAppointmentID(txCtx, clinicID, id)
		if err != nil {
			return apperrors.Wrap(err, "failed to get trimming detail for update")
		}
		if input.CourseID != nil {
			detail.CourseID = input.CourseID
		}
		if input.StyleRequest != nil {
			detail.StyleRequest = *input.StyleRequest
		}
		if input.BodyWeight != nil {
			detail.BodyWeight = input.BodyWeight
		}
		if input.BWUnit != nil {
			detail.BWUnit = *input.BWUnit
		}
		if input.BodyTemperature != nil {
			detail.BodyTemperature = input.BodyTemperature
		}
		if input.UsedShampoo != nil {
			detail.UsedShampoo = *input.UsedShampoo
		}
		if input.UsedRibbon != nil {
			detail.UsedRibbon = *input.UsedRibbon
		}
		if input.Remarks != nil {
			detail.Remarks = *input.Remarks
		}
		if input.StyleImage != nil {
			detail.StyleImage = *input.StyleImage
		}
		if input.CompletedImage != nil {
			detail.CompletedImage = *input.CompletedImage
		}
		if err := s.trimmingDetail.Update(txCtx, detail); err != nil {
			return apperrors.Wrap(err, "failed to update trimming detail")
		}
		if input.OptionIDs != nil {
			if err := s.trimmingDetail.SetOptions(txCtx, id, *input.OptionIDs); err != nil {
				return apperrors.Wrap(err, "failed to set trimming options")
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "trimming appointment updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("appointment_id", id))

	return s.GetByID(ctx, clinicID, id)
}

func (s *trimmingService) Delete(ctx context.Context, clinicID, id uint64) error {
	if err := s.reservation.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete trimming appointment")
	}
	slog.InfoContext(ctx, "trimming appointment deleted",
		slog.Uint64("appointment_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
