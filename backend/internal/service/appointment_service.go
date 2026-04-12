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

// errNoDoctorsOnDuty は当日の出勤医師が 0 人のためスロット予約不可を示すセンチネルエラー。
// apperrors.ErrConflict をラップしているため RespondError で 409 になる。
// LINE パスでは reservation_validators.go が errors.Is でこれを識別し RedirectStep: 4 を返す。
var errNoDoctorsOnDuty = fmt.Errorf("本日は医師が出勤していないため予約できません: %w", apperrors.ErrConflict)

// ptrToUint64 は *uint64 を uint64 に変換する（nil の場合は 0）
func ptrToUint64(p *uint64) uint64 {
	if p == nil {
		return 0
	}
	return *p
}

// checkDoctorSlotConflict は特定医師の時間枠重複をチェックする（SELECT FOR UPDATE）。
func checkDoctorSlotConflict(ctx context.Context, tx *gorm.DB, clinicID, doctorID uint64, startTime, endTime time.Time, excludeID *uint64) error {
	var existing []model.Appointment
	q := tx.WithContext(ctx).Raw(`
		SELECT id FROM appointments
		WHERE clinic_id = ?
		  AND deleted_at IS NULL
		  AND status NOT IN ('cancelled')
		  AND start_time < ?
		  AND end_time > ?
		  AND doctor_id = ?
		  AND (? = 0 OR id != ?)
		FOR UPDATE`,
		clinicID, endTime, startTime, doctorID,
		ptrToUint64(excludeID), ptrToUint64(excludeID),
	)
	if err := q.Scan(&existing).Error; err != nil {
		return apperrors.Wrap(err, "lock reservations for conflict check")
	}
	if len(existing) > 0 {
		return apperrors.WrapConflict("この時間枠は既に予約が入っています")
	}
	return nil
}

// checkCapacitySlotConflict は出勤医師数を上限として時間枠の空き確認をする（SELECT FOR UPDATE）。
// 出勤医師が 0 人の場合は errNoDoctorsOnDuty を返す（LINE パスで RedirectStep を分岐するため）。
func checkCapacitySlotConflict(ctx context.Context, tx *gorm.DB, clinicID uint64, startTime, endTime time.Time, excludeID *uint64) error {
	var doctorCount int64
	cntQ := tx.WithContext(ctx).Raw(`
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
	if err := cntQ.Scan(&doctorCount).Error; err != nil {
		return apperrors.Wrap(err, "count on-duty doctors")
	}
	if doctorCount == 0 {
		return errNoDoctorsOnDuty
	}

	var existing []model.Appointment
	q := tx.WithContext(ctx).Raw(`
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
	if err := q.Scan(&existing).Error; err != nil {
		return apperrors.Wrap(err, "lock reservations for capacity check")
	}
	if int64(len(existing)) >= doctorCount {
		return apperrors.WrapConflict("この時間枠は満員です（出勤医師数に達しています）")
	}
	return nil
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
		return checkDoctorSlotConflict(ctx, tx, clinicID, *doctorID, startTime, endTime, excludeID)
	}
	return checkCapacitySlotConflict(ctx, tx, clinicID, startTime, endTime, excludeID)
}

// resolveUpdateParams は現在の予約と更新入力から、競合チェックに使用する時刻・医師 ID を確定する。
// 未指定フィールドは現在値を維持する。DoctorID=0 は NULL（医師未指定）として扱う。
func resolveUpdateParams(current model.Appointment, input *UpdateReservationInput) (start, end time.Time, doctorID *uint64) {
	start = current.StartTime
	if input.StartTime != nil {
		start = *input.StartTime
	}
	end = current.EndTime
	if input.EndTime != nil {
		end = *input.EndTime
	}
	doctorID = current.DoctorID
	if input.DoctorID != nil {
		if *input.DoctorID == 0 {
			doctorID = nil // 0 は「医師未指定」として NULL 扱い
		} else {
			doctorID = input.DoctorID
		}
	}
	return start, end, doctorID
}

// updateWithConflictCheck は SELECT FOR UPDATE + トランザクション内で競合チェック + 予約更新を実行する。
// 時刻・医師変更がある場合にのみ呼び出す。
func (s *reservationService) updateWithConflictCheck(ctx context.Context, clinicID, id uint64, fields map[string]any, input *UpdateReservationInput) (*model.Appointment, error) {
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

		resolvedStart, resolvedEnd, resolvedDoctorID := resolveUpdateParams(current, input)

		if input.StartTime != nil || input.EndTime != nil {
			if err := validateTimeRange(resolvedStart, resolvedEnd); err != nil {
				return err
			}
		}

		if err := checkSlotConflict(ctx, tx, clinicID, resolvedDoctorID, resolvedStart, resolvedEnd, &id); err != nil {
			return err
		}

		res := tx.Model(&model.Appointment{}).
			Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, id).
			Updates(fields)
		if res.Error != nil {
			return apperrors.FromGORM(res.Error, "reservation", fmt.Sprintf("%d", id))
		}
		if res.RowsAffected == 0 {
			return apperrors.WrapNotFound("reservation", fmt.Sprintf("%d", id))
		}

		var updated model.Appointment
		if err := tx.Where("clinic_id = ? AND id = ? AND deleted_at IS NULL", clinicID, id).
			First(&updated).Error; err != nil {
			return apperrors.FromGORM(err, "reservation", fmt.Sprintf("%d", id))
		}
		result = &updated
		return nil
	})
	return result, err
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
	result, err := s.updateWithConflictCheck(ctx, clinicID, id, fields, input)
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
		if *input.DoctorID == 0 {
			fields["doctor_id"] = nil // 0 は「医師未指定に変更」として NULL に設定
		} else {
			fields["doctor_id"] = *input.DoctorID
		}
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
