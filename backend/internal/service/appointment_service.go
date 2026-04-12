package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

type ReservationService interface {
	List(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Appointment, int64, error)
	GetByID(ctx context.Context, clinicID, id uint64) (*model.Appointment, error)
	Create(ctx context.Context, reservation *model.Appointment) error
	Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.Appointment, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type reservationService struct {
	repo repository.ReservationRepository
	db   *gorm.DB
}

func NewReservationService(repo repository.ReservationRepository, db *gorm.DB) ReservationService {
	return &reservationService{repo: repo, db: db}
}

func (s *reservationService) List(ctx context.Context, clinicID uint64, page, limit int, date *time.Time, status, source *string, petID, ownerID *uint64) ([]model.Appointment, int64, error) {
	items, total, err := s.repo.FindAll(ctx, clinicID, page, limit, date, status, source, petID, ownerID)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to list reservations")
	}
	return items, total, nil
}

func (s *reservationService) GetByID(ctx context.Context, clinicID, id uint64) (*model.Appointment, error) {
	result, err := s.repo.FindByID(ctx, clinicID, id)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get reservation")
	}
	return result, nil
}

func (s *reservationService) Create(ctx context.Context, reservation *model.Appointment) error {
	// BUG-034: end_time <= start_time の場合は 400 Bad Request
	if err := validateTimeRange(reservation.StartTime, reservation.EndTime); err != nil {
		return err
	}

	// SELECT FOR UPDATE + トランザクションで競合を防止
	// LINE予約・電子カルテ予約・管理者手動予約すべてで同一テーブルを使用するため、
	// アプリケーションレベルでの排他制御が必要
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := checkSlotConflict(ctx, tx, reservation.ClinicID, reservation.DoctorID, reservation.StartTime, reservation.EndTime, nil); err != nil {
			return err
		}
		if err := tx.Create(reservation).Error; err != nil {
			return apperrors.Wrap(err, "create reservation")
		}
		return nil
	})
	if err != nil {
		return apperrors.Wrap(err, "failed to create reservation")
	}

	slog.InfoContext(ctx, "reservation created",
		slog.Uint64("reservation_id", reservation.ID),
		slog.Uint64("clinic_id", reservation.ClinicID))
	return nil
}

// validateTimeRange は end_time > start_time を確認する共通バリデーション。
func validateTimeRange(startTime, endTime time.Time) error {
	if !endTime.After(startTime) {
		return apperrors.WrapInvalidInput("end_time must be after start_time")
	}
	return nil
}

// ptrToUint64 は *uint64 を uint64 に変換する（nil の場合は 0）
func ptrToUint64(p *uint64) uint64 {
	if p == nil {
		return 0
	}
	return *p
}

// checkSlotConflict は時間枠の空き・重複をチェックする（SELECT FOR UPDATE）。
//
//   - doctor_id 指定時 → 同一医師の重複のみチェック（別医師は許可）
//   - doctor_id nil 時 → その日の出勤医師数を上限として全予約件数をチェック
//
// excludeID が非 nil の場合、その予約 ID を競合対象から除外する（Update 時の自己競合防止）。
// 競合がある場合は apperrors.ErrConflict ラップエラーを返す。
func checkSlotConflict(ctx context.Context, tx *gorm.DB, clinicID uint64, doctorID *uint64, startTime, endTime time.Time, excludeID *uint64) error {
	if doctorID != nil {
		// 同一医師への重複チェック
		var existing []model.Appointment
		q := tx.Raw(`
			SELECT id FROM appointments
			WHERE clinic_id = ?
			  AND deleted_at IS NULL
			  AND status NOT IN ('cancelled')
			  AND start_time < ?
			  AND end_time > ?
			  AND doctor_id = ?
			  AND (? = 0 OR id != ?)
			FOR UPDATE`,
			clinicID, endTime, startTime, *doctorID,
			ptrToUint64(excludeID), ptrToUint64(excludeID),
		)
		if q.Scan(&existing); q.Error != nil {
			return apperrors.Wrap(q.Error, "lock reservations for conflict check")
		}
		if len(existing) > 0 {
			return apperrors.WrapConflict("この時間枠は既に予約が入っています")
		}
		return nil
	}

	// 医師未指定: その日の出勤医師数を上限として全予約件数をチェック
	var doctorCount int64
	cntQ := tx.Raw(`
		SELECT COUNT(DISTINCT se.staff_id)
		FROM shift_entries se
		JOIN staffs s ON s.id = se.staff_id
		WHERE se.clinic_id = ?
		  AND se.date = DATE(? AT TIME ZONE 'Asia/Tokyo')
		  AND se.shift_type NOT IN ('off', 'paid_leave')
		  AND s.staff_type = 'doctor'
		  AND s.is_active = true
		  AND s.deleted_at IS NULL`,
		clinicID, startTime,
	)
	if cntQ.Scan(&doctorCount); cntQ.Error != nil {
		return apperrors.Wrap(cntQ.Error, "count on-duty doctors")
	}

	var existing []model.Appointment
	q := tx.Raw(`
		SELECT id FROM appointments
		WHERE clinic_id = ?
		  AND deleted_at IS NULL
		  AND status NOT IN ('cancelled')
		  AND start_time < ?
		  AND end_time > ?
		  AND (? = 0 OR id != ?)
		FOR UPDATE`,
		clinicID, endTime, startTime,
		ptrToUint64(excludeID), ptrToUint64(excludeID),
	)
	if q.Scan(&existing); q.Error != nil {
		return apperrors.Wrap(q.Error, "lock reservations for capacity check")
	}

	if int64(len(existing)) >= doctorCount {
		return apperrors.WrapConflict("この時間枠は満員です（出勤医師数に達しています）")
	}
	return nil
}

func (s *reservationService) Update(ctx context.Context, clinicID, id uint64, input *UpdateReservationInput) (*model.Appointment, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	fields := buildReservationUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}

	// 時刻・医師の変更がある場合のみ競合チェックが必要
	needsConflictCheck := input.StartTime != nil || input.EndTime != nil || input.DoctorID != nil

	if !needsConflictCheck {
		// 時刻・医師変更なし: トランザクション不要。リポジトリ経由で直接更新
		updated, err := s.repo.UpdateFields(ctx, clinicID, id, fields)
		if err != nil {
			return nil, apperrors.Wrap(err, "failed to update reservation")
		}
		slog.InfoContext(ctx, "reservation updated",
			slog.Uint64("reservation_id", id),
			slog.Uint64("clinic_id", clinicID))
		return updated, nil
	}

	// 時刻・医師変更あり: SELECT FOR UPDATE + トランザクションで競合を防止
	var result *model.Appointment
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 現在の予約を行ロックで取得（競合チェックの基準値として使用）
		var current model.Appointment
		if err := tx.Raw(
			`SELECT * FROM appointments WHERE clinic_id = ? AND id = ? AND deleted_at IS NULL FOR UPDATE`,
			clinicID, id,
		).Scan(&current).Error; err != nil {
			return apperrors.Wrap(err, "lock current reservation")
		}
		if current.ID == 0 {
			return apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
		}

		// 変更後の時刻・医師を確定（未指定フィールドは現在値を維持）
		resolvedStart := current.StartTime
		if input.StartTime != nil {
			resolvedStart = *input.StartTime
		}
		resolvedEnd := current.EndTime
		if input.EndTime != nil {
			resolvedEnd = *input.EndTime
		}
		resolvedDoctorID := current.DoctorID
		if input.DoctorID != nil {
			resolvedDoctorID = input.DoctorID
		}

		// 時刻が変更される場合のみ時刻バリデーション
		if input.StartTime != nil || input.EndTime != nil {
			if err := validateTimeRange(resolvedStart, resolvedEnd); err != nil {
				return err
			}
		}

		// 自分自身を除外して競合チェック
		if err := checkSlotConflict(ctx, tx, clinicID, resolvedDoctorID, resolvedStart, resolvedEnd, &id); err != nil {
			return err
		}

		// フィールドを更新
		res := tx.Model(&model.Appointment{}).
			Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, id).
			Updates(fields)
		if res.Error != nil {
			return apperrors.FromGORM(res.Error, "reservation", fmt.Sprintf("%d", id))
		}
		if res.RowsAffected == 0 {
			return apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
		}

		// 更新後レコードを取得
		var updated model.Appointment
		if err := tx.Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, id).
			First(&updated).Error; err != nil {
			return apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
		}
		result = &updated
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to update reservation")
	}

	slog.InfoContext(ctx, "reservation updated",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID))
	return result, nil
}

// UpdateReservationInput は予約更新のサービス入力 DTO
type UpdateReservationInput struct {
	StartTime         *time.Time
	EndTime           *time.Time
	OwnerID           *uint64
	PetID             *uint64
	VisitType         *model.VisitType
	ReservationTypeID *uint64
	DoctorID          *uint64
	IsDesignated      *bool
	Status            *model.ReservationStatus
	Notes             *string
}

func buildReservationUpdateFields(input *UpdateReservationInput) map[string]any {
	fields := make(map[string]any)
	if input.StartTime != nil {
		fields["start_time"] = *input.StartTime
	}
	if input.EndTime != nil {
		fields["end_time"] = *input.EndTime
	}
	if input.OwnerID != nil {
		fields["owner_id"] = *input.OwnerID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.VisitType != nil {
		fields["visit_type"] = *input.VisitType
	}
	if input.ReservationTypeID != nil {
		fields["reservation_type_id"] = *input.ReservationTypeID
	}
	if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.IsDesignated != nil {
		fields["is_designated"] = *input.IsDesignated
	}
	if input.Status != nil {
		fields["status"] = *input.Status
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	return fields
}

func (s *reservationService) Delete(ctx context.Context, clinicID, id uint64) error {
	count, err := s.repo.CountMedicalRecordsByReservationID(ctx, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to check reservation dependencies")
	}
	if count > 0 {
		return apperrors.WrapConflict("この予約にはカルテが紐付いているため削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, id); err != nil {
		return apperrors.Wrap(err, "failed to delete reservation")
	}
	slog.InfoContext(ctx, "reservation deleted",
		slog.Uint64("reservation_id", id),
		slog.Uint64("clinic_id", clinicID))
	return nil
}
