package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CashRegisterCloseRepository はレジ締めレコードのデータアクセスインターフェース
type CashRegisterCloseRepository interface {
	Create(ctx context.Context, c *model.CashRegisterClose) error
	FindAll(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error)
	FindByDateAndPeriod(ctx context.Context, clinicID uint64, date time.Time, period string) (*model.CashRegisterClose, error)
}

type cashRegisterCloseRepository struct{ db *gorm.DB }

// NewCashRegisterCloseRepository は CashRegisterCloseRepository を初期化して返す
func NewCashRegisterCloseRepository(db *gorm.DB) CashRegisterCloseRepository {
	return &cashRegisterCloseRepository{db: db}
}

func (r *cashRegisterCloseRepository) Create(ctx context.Context, c *model.CashRegisterClose) error {
	if err := r.db.WithContext(ctx).Create(c).Error; err != nil {
		return apperrors.FromGORM(err, "cash_register_close", "")
	}
	return nil
}

func (r *cashRegisterCloseRepository) FindAll(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error) {
	q := r.db.WithContext(ctx).
		Model(&model.CashRegisterClose{}).
		Where("clinic_id = ?", clinicID)
	if startDate != nil {
		q = q.Where("close_date >= ?", startDate)
	}
	if endDate != nil {
		q = q.Where("close_date <= ?", endDate)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to count cash register closes")
	}

	var cs []model.CashRegisterClose
	offset := (page - 1) * limit
	if err := q.Preload("ClosedByStaff").
		Order("close_date DESC, period ASC").
		Offset(offset).
		Limit(limit).
		Find(&cs).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "failed to find cash register closes")
	}
	return cs, total, nil
}

func (r *cashRegisterCloseRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error) {
	var c model.CashRegisterClose
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		Preload("ClosedByStaff").
		First(&c).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "cash_register_close", fmt.Sprintf("%d", id))
	}
	return &c, nil
}

func (r *cashRegisterCloseRepository) FindByDateAndPeriod(ctx context.Context, clinicID uint64, date time.Time, period string) (*model.CashRegisterClose, error) {
	var c model.CashRegisterClose
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND close_date = ? AND period = ?", clinicID, date, period).
		First(&c).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find cash register close")
	}
	return &c, nil
}
