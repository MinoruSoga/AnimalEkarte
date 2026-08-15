package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// CashRegisterCloseRepository はレジ締めレコードのデータアクセスインターフェース。
// W-013 FINAL B: Create / CreateAdjustment のみ。Update・Delete・soft-delete 再開は持たない（append-only）。
type CashRegisterCloseRepository interface {
	Create(ctx context.Context, c *model.CashRegisterClose) error
	// CreateAdjustment は締め後訂正台帳へ 1 行追記する。ambient tx があれば参加する（fail-closed）。
	CreateAdjustment(ctx context.Context, adj *model.CashRegisterCloseAdjustment) error
	FindAll(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.CashRegisterClose, error)
	FindByDateAndPeriod(ctx context.Context, clinicID uint64, date time.Time, period string) (*model.CashRegisterClose, error)
	// #115: 指定日に1件以上のレジ締めレコードが存在するか確認する。
	HasCloseOnDate(ctx context.Context, clinicID uint64, date time.Time) (bool, error)
}

type cashRegisterCloseRepository struct{ db *gorm.DB }

// NewCashRegisterCloseRepository は CashRegisterCloseRepository を初期化して返す
func NewCashRegisterCloseRepository(db *gorm.DB) CashRegisterCloseRepository {
	return &cashRegisterCloseRepository{db: db}
}

func (r *cashRegisterCloseRepository) Create(ctx context.Context, c *model.CashRegisterClose) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(c).Error; err != nil {
		return apperrors.FromGORM(err, "cash_register_close", "")
	}
	return nil
}

func (r *cashRegisterCloseRepository) CreateAdjustment(ctx context.Context, adj *model.CashRegisterCloseAdjustment) error {
	if adj == nil {
		return apperrors.WrapInvalidInput("cash register close adjustment is required")
	}
	if adj.Reason == "" {
		return apperrors.WrapInvalidInput("cash register close adjustment reason is required")
	}
	if err := persistence.DBOrTx(ctx, r.db).Create(adj).Error; err != nil {
		return apperrors.FromGORM(err, "cash_register_close_adjustment", "")
	}
	return nil
}

func (r *cashRegisterCloseRepository) FindAll(ctx context.Context, clinicID uint64, startDate, endDate *time.Time, page, limit int) ([]model.CashRegisterClose, int64, error) {
	q := persistence.DBOrTx(ctx, r.db).
		Model(&model.CashRegisterClose{}).
		Scopes(persistence.ClinicScope(clinicID))
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
	if err := q.Preload("ClosedByStaff", "deleted_at IS NULL").
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
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Preload("ClosedByStaff", "deleted_at IS NULL").
		First(&c).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "cash_register_close", fmt.Sprintf("%d", id))
	}
	return &c, nil
}

func (r *cashRegisterCloseRepository) HasCloseOnDate(ctx context.Context, clinicID uint64, date time.Time) (bool, error) {
	var count int64
	err := persistence.DBOrTx(ctx, r.db).
		Model(&model.CashRegisterClose{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("close_date = ?", date.Format(time.DateOnly)).
		Count(&count).Error
	if err != nil {
		return false, apperrors.Wrap(err, "failed to check close on date")
	}
	return count > 0, nil
}

func (r *cashRegisterCloseRepository) FindByDateAndPeriod(ctx context.Context, clinicID uint64, date time.Time, period string) (*model.CashRegisterClose, error) {
	var c model.CashRegisterClose
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("close_date = ? AND period = ?", date, period).
		First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find cash register close")
	}
	return &c, nil
}
