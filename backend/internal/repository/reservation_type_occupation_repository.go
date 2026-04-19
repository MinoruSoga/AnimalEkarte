package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// jstLoc は起動時に一度だけロードする JST タイムゾーン。
// tzdata 欠落時は起動時に panic して早期発見できる（_ エラー握りつぶし禁止の規約に準拠）。
var jstLoc = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(fmt.Sprintf("failed to load JST timezone: %v", err))
	}
	return loc
}()

// ReservationTypeOccupationRepository は職種紐付けの永続化インターフェース
type ReservationTypeOccupationRepository interface {
	// FindAll は Occupation を Preload して返す
	FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error)
	Create(ctx context.Context, o *model.ReservationTypeOccupation) error
	// Delete は物理削除（論理削除なし）
	Delete(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error
	// CountWorkingStaff は指定日に対応職種のスタッフが何人出勤しているかを返す（LIFF 専用）
	// shift_type が 'off' / 'paid_leave' 以外のスタッフのみカウント
	CountWorkingStaff(ctx context.Context, clinicID, reservationTypeID uint64, date time.Time) (int64, error)
}

type reservationTypeOccupationRepository struct {
	db *gorm.DB
}

// NewReservationTypeOccupationRepository はリポジトリを初期化して返す
func NewReservationTypeOccupationRepository(db *gorm.DB) ReservationTypeOccupationRepository {
	return &reservationTypeOccupationRepository{db: db}
}

func (r *reservationTypeOccupationRepository) FindAll(
	ctx context.Context, clinicID, reservationTypeID uint64,
) ([]model.ReservationTypeOccupation, error) {
	var results []model.ReservationTypeOccupation
	err := r.db.WithContext(ctx).
		Preload("Occupation").
		Where("clinic_id = ? AND reservation_type_id = ?", clinicID, reservationTypeID).
		Order("id ASC").
		Find(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_occupations", fmt.Sprintf("clinic=%d type=%d", clinicID, reservationTypeID))
	}
	return results, nil
}

func (r *reservationTypeOccupationRepository) Create(
	ctx context.Context, o *model.ReservationTypeOccupation,
) error {
	if err := r.db.WithContext(ctx).Create(o).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type_occupation", "")
	}
	return nil
}

func (r *reservationTypeOccupationRepository) Delete(
	ctx context.Context, clinicID, reservationTypeID, occupationID uint64,
) error {
	result := r.db.WithContext(ctx).
		Where("clinic_id = ? AND reservation_type_id = ? AND occupation_id = ?", clinicID, reservationTypeID, occupationID).
		Delete(&model.ReservationTypeOccupation{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_type_occupation", fmt.Sprintf("type=%d occ=%d", reservationTypeID, occupationID))
	}
	if result.RowsAffected == 0 {
		return apperrors.FromGORM(gorm.ErrRecordNotFound, "reservation_type_occupation", fmt.Sprintf("type=%d occ=%d", reservationTypeID, occupationID))
	}
	return nil
}

func (r *reservationTypeOccupationRepository) CountWorkingStaff(
	ctx context.Context, clinicID, reservationTypeID uint64, date time.Time,
) (int64, error) {
	// JST 日付文字列で shift_entries.date と比較する
	dateStr := date.In(jstLoc).Format("2006-01-02")
	var count int64
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(DISTINCT se.staff_id)
		FROM reservation_type_occupations rto
		JOIN staffs s ON s.occupation_id = rto.occupation_id AND s.deleted_at IS NULL
		JOIN shift_entries se ON se.staff_id = s.id
			AND se.clinic_id = ?
			AND se.date = ?
			AND se.shift_type NOT IN ('off', 'paid_leave')
		WHERE rto.clinic_id = ?
		  AND rto.reservation_type_id = ?
	`, clinicID, dateStr, clinicID, reservationTypeID).Scan(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "count_working_staff", fmt.Sprintf("type=%d date=%s", reservationTypeID, dateStr))
	}
	return count, nil
}
