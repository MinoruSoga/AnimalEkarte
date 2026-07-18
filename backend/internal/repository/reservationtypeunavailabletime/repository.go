// Package reservationtypeunavailabletime owns reservation_type_unavailable_times data
// access (BE8-4 Wave D — "reservation type children (slot/unavailable)" per
// repository/CLAUDE.md's next cut order).
package reservationtypeunavailabletime

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// Repository は予約不可時間の永続化インターフェース
type Repository interface {
	// FindAll は指定予約区分の予約不可時間を全件返す（LIFF・管理API 両方が使用）
	FindAll(ctx context.Context, clinicID, reservationTypeID uint64) ([]model.ReservationTypeUnavailableTime, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.ReservationTypeUnavailableTime, error)
	Create(ctx context.Context, t *model.ReservationTypeUnavailableTime) error
	// Delete は id で物理削除する（論理削除なし）
	Delete(ctx context.Context, clinicID, id uint64) error
}

type repository struct {
	db *gorm.DB
}

// New はリポジトリを初期化して返す
func New(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) FindAll(
	ctx context.Context, clinicID, reservationTypeID uint64,
) ([]model.ReservationTypeUnavailableTime, error) {
	var results []model.ReservationTypeUnavailableTime
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("reservation_type_id = ?", reservationTypeID).
		Order("id ASC").
		Find(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_unavailable_times", fmt.Sprintf("clinic=%d type=%d", clinicID, reservationTypeID))
	}
	return results, nil
}

func (r *repository) FindByID(
	ctx context.Context, clinicID, id uint64,
) (*model.ReservationTypeUnavailableTime, error) {
	var result model.ReservationTypeUnavailableTime
	err := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
		Where("id = ?", id).
		First(&result).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "reservation_type_unavailable_time", fmt.Sprintf("%d", id))
	}
	return &result, nil
}

func (r *repository) Create(
	ctx context.Context, t *model.ReservationTypeUnavailableTime,
) error {
	if err := r.db.WithContext(ctx).Create(t).Error; err != nil {
		return apperrors.FromGORM(err, "reservation_type_unavailable_time", "")
	}
	return nil
}

func (r *repository) Delete(
	ctx context.Context, clinicID, id uint64,
) error {
	result := r.db.WithContext(ctx).
		Scopes(repohelpers.ClinicScope(clinicID)).
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
