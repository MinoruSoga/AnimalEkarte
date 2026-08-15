package reservation

import (
	"context"
	"fmt"
	"time"
	_ "time/tzdata" // Asia/Tokyo タイムゾーンデータを組み込む

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ReservationTypeOccupationRepository は職種紐付けの永続化インターフェース
type ReservationTypeOccupationRepository interface {
	// FindAll は Occupation を Preload して返す
	FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeOccupation, error)
	// FindByID は指定した occupation_id に対応する紐付けを1件返す
	FindByID(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) (*model.ReservationTypeOccupation, error)
	Create(ctx context.Context, o *model.ReservationTypeOccupation) error
	// Delete は物理削除（論理削除なし）
	Delete(ctx context.Context, clinicID, reservationTypeID, occupationID uint64) error
	// CountWorkingStaffByReservationTypeIDs は複数日分の出勤スタッフ数を1クエリでまとめて返す(G7-1: 日付ループN+1回避)。
	// 戻り値のキーは time.DateOnly 形式(JST)。シフトが無い日はキーとして存在しない(0扱い)。dates が空なら空map即返し。
	CountWorkingStaffByReservationTypeIDs(ctx context.Context, clinicID, reservationTypeID uint64, dates []time.Time) (map[string]int64, error)
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
		Preload("Occupation", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("reservation_type_id = ?", reservationTypeID).
		Order("id ASC").
		Find(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_occupations", fmt.Sprintf("clinic=%d type=%d", clinicID, reservationTypeID))
	}
	return results, nil
}

func (r *reservationTypeOccupationRepository) FindByID(
	ctx context.Context, clinicID, reservationTypeID, occupationID uint64,
) (*model.ReservationTypeOccupation, error) {
	var rto model.ReservationTypeOccupation
	err := r.db.WithContext(ctx).
		Preload("Occupation", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("reservation_type_id = ? AND occupation_id = ?", reservationTypeID, occupationID).
		First(&rto).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_occupation", "")
	}
	return &rto, nil
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
		Scopes(persistence.ClinicScope(clinicID)).
		Where("reservation_type_id = ? AND occupation_id = ?", reservationTypeID, occupationID).
		Delete(&model.ReservationTypeOccupation{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_type_occupation", fmt.Sprintf("type=%d occ=%d", reservationTypeID, occupationID))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_type_occupation", fmt.Sprintf("type=%d occ=%d", reservationTypeID, occupationID))
	}
	return nil
}

func (r *reservationTypeOccupationRepository) CountWorkingStaffByReservationTypeIDs(
	ctx context.Context, clinicID, reservationTypeID uint64, dates []time.Time,
) (map[string]int64, error) {
	result := make(map[string]int64, len(dates))
	if len(dates) == 0 {
		return result, nil
	}
	dateStrs := make([]string, len(dates))
	for i, d := range dates {
		dateStrs[i] = d.In(config.JST).Format(time.DateOnly)
	}
	type row struct {
		Date  string `gorm:"column:date"`
		Count int64  `gorm:"column:cnt"`
	}
	var rows []row
	err := r.db.WithContext(ctx).Raw(`
		SELECT se.date::text AS date, COUNT(DISTINCT se.staff_id) AS cnt
		FROM reservation_type_occupations rto
		JOIN staffs s ON s.occupation_id = rto.occupation_id AND s.deleted_at IS NULL
		JOIN shift_entries se ON se.staff_id = s.id
			AND se.clinic_id = ?
			AND se.date IN ?
			AND se.shift_type NOT IN ('off', 'paid_leave')
		WHERE rto.clinic_id = ?
		  AND rto.reservation_type_id = ?
		GROUP BY se.date
	`, clinicID, dateStrs, clinicID, reservationTypeID).Scan(&rows).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "count_working_staff_batch", fmt.Sprintf("type=%d", reservationTypeID))
	}
	for _, rw := range rows {
		result[rw.Date] = rw.Count
	}
	return result, nil
}
