package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// AccountingRepository 会計リポジトリインターフェース
type AccountingRepository interface {
	GetAllAccounting(ctx context.Context) ([]model.Accounting, error)
	GetAccountingByID(ctx context.Context, id string) (*model.Accounting, error)
	GetAccountingByPetID(ctx context.Context, petID string) ([]model.Accounting, error)
	GetAccountingByOwnerID(ctx context.Context, ownerID string) ([]model.Accounting, error)
	GetAccountingByStatus(ctx context.Context, status string) ([]model.Accounting, error)
	CreateAccounting(ctx context.Context, acc *model.Accounting) error
	UpdateAccounting(ctx context.Context, acc *model.Accounting) error
	DeleteAccounting(ctx context.Context, id string) error
}

// accountingRepository 会計リポジトリ実装
type accountingRepository struct {
	db *gorm.DB
}

// NewAccountingRepository 新しい会計リポジトリを作成
func NewAccountingRepository(db *gorm.DB) AccountingRepository {
	return &accountingRepository{db: db}
}

// GetAllAccounting 全ての会計を取得
func (r *accountingRepository) GetAllAccounting(ctx context.Context) ([]model.Accounting, error) {
	var accs []model.Accounting
	result := r.db.WithContext(ctx).
		Preload("Items").
		Preload("PaymentInfo").
		Order("scheduled_date DESC, created_at DESC").
		Find(&accs)

	if result.Error != nil {
		return nil, result.Error
	}

	return accs, nil
}

// GetAccountingByID IDで会計を取得
func (r *accountingRepository) GetAccountingByID(ctx context.Context, id string) (*model.Accounting, error) {
	var acc model.Accounting
	result := r.db.WithContext(ctx).
		Preload("Items").
		Preload("PaymentInfo").
		First(&acc, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("accounting with id %s not found", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get accounting")
	}

	return &acc, nil
}

// GetAccountingByPetID ペットIDで会計を取得
func (r *accountingRepository) GetAccountingByPetID(ctx context.Context, petID string) ([]model.Accounting, error) {
	var accs []model.Accounting
	result := r.db.WithContext(ctx).
		Preload("Items").
		Preload("PaymentInfo").
		Where("pet_id = ?", petID).
		Order("scheduled_date DESC, created_at DESC").
		Find(&accs)

	if result.Error != nil {
		return nil, result.Error
	}

	return accs, nil
}

// GetAccountingByOwnerID 飼い主IDで会計を取得
func (r *accountingRepository) GetAccountingByOwnerID(ctx context.Context, ownerID string) ([]model.Accounting, error) {
	var accs []model.Accounting
	result := r.db.WithContext(ctx).
		Preload("Items").
		Preload("PaymentInfo").
		Where("owner_id = ?", ownerID).
		Order("scheduled_date DESC, created_at DESC").
		Find(&accs)

	if result.Error != nil {
		return nil, result.Error
	}

	return accs, nil
}

// GetAccountingByStatus ステータスで会計を取得
func (r *accountingRepository) GetAccountingByStatus(ctx context.Context, status string) ([]model.Accounting, error) {
	var accs []model.Accounting
	result := r.db.WithContext(ctx).
		Preload("Items").
		Preload("PaymentInfo").
		Where("status = ?", status).
		Order("scheduled_date DESC, created_at DESC").
		Find(&accs)

	if result.Error != nil {
		return nil, result.Error
	}

	return accs, nil
}

// CreateAccounting 会計を作成
func (r *accountingRepository) CreateAccounting(ctx context.Context, acc *model.Accounting) error {
	// Generate UUID if not set
	if acc.ID == uuid.Nil {
		acc.ID = uuid.New()
	}
	result := r.db.WithContext(ctx).Create(acc)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateAccounting 会計を更新
func (r *accountingRepository) UpdateAccounting(ctx context.Context, acc *model.Accounting) error {
	result := r.db.WithContext(ctx).Save(acc)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteAccounting 会計を削除
func (r *accountingRepository) DeleteAccounting(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Accounting{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
