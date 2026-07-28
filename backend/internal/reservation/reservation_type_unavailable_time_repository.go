package reservation

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// ReservationTypeUnavailableTimeRepository は予約不可時間の永続化インターフェース
type ReservationTypeUnavailableTimeRepository interface {
	// FindAll は指定予約区分の予約不可時間を全件返す（LIFF・管理API 両方が使用）
	FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeUnavailableTime, error)
	Create(ctx context.Context, t *model.ReservationTypeUnavailableTime) error
	// Delete は id で物理削除する（論理削除なし）
	Delete(ctx context.Context, clinicID, id uint64) error
}

type reservationTypeUnavailableTimeRepository struct {
	db *gorm.DB
}

// New はリポジトリを初期化して返す
func NewReservationTypeUnavailableTimeRepository(db *gorm.DB) ReservationTypeUnavailableTimeRepository {
	return &reservationTypeUnavailableTimeRepository{db: db}
}

func (r *reservationTypeUnavailableTimeRepository) FindAll(
	ctx context.Context, clinicID, reservationTypeID uint64,
) ([]model.ReservationTypeUnavailableTime, error) {
	var results []model.ReservationTypeUnavailableTime
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("reservation_type_id = ?", reservationTypeID).
		Order("id ASC").
		Find(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_unavailable_times", fmt.Sprintf("clinic=%d type=%d", clinicID, reservationTypeID))
	}
	return results, nil
}

func (r *reservationTypeUnavailableTimeRepository) FindByID(
	ctx context.Context, clinicID, id uint64,
) (*model.ReservationTypeUnavailableTime, error) {
	var result model.ReservationTypeUnavailableTime
	err := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&result).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_unavailable_time", fmt.Sprintf("%d", id))
	}
	return &result, nil
}

func (r *reservationTypeUnavailableTimeRepository) Create(
	ctx context.Context, t *model.ReservationTypeUnavailableTime,
) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type_unavailable_time", "")
	}
	return nil
}

func (r *reservationTypeUnavailableTimeRepository) Delete(
	ctx context.Context, clinicID, id uint64,
) error {
	result := r.db.WithContext(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Delete(&model.ReservationTypeUnavailableTime{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "reservation_type_unavailable_time", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("reservation_type_unavailable_time", fmt.Sprintf("%d", id))
	}
	return nil
}
