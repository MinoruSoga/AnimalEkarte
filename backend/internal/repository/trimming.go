package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TrimmingRepository トリミングリポジトリインターフェース
type TrimmingRepository interface {
	GetAllTrimmings(ctx context.Context) ([]model.TrimmingRecord, error)
	GetTrimmingByID(ctx context.Context, id string) (*model.TrimmingRecord, error)
	GetTrimmingsByPetID(ctx context.Context, petID string) ([]model.TrimmingRecord, error)
	GetTrimmingsByOwnerID(ctx context.Context, ownerID string) ([]model.TrimmingRecord, error)
	GetTrimmingsByStatus(ctx context.Context, status string) ([]model.TrimmingRecord, error)
	CreateTrimming(ctx context.Context, trim *model.TrimmingRecord) error
	UpdateTrimming(ctx context.Context, trim *model.TrimmingRecord) error
	DeleteTrimming(ctx context.Context, id string) error
}

// trimmingRepository トリミングリポジトリ実装
type trimmingRepository struct {
	db *gorm.DB
}

// NewTrimmingRepository 新しいトリミングリポジトリを作成
func NewTrimmingRepository(db *gorm.DB) TrimmingRepository {
	return &trimmingRepository{db: db}
}

// GetAllTrimmings 全てのトリミングを取得
func (r *trimmingRepository) GetAllTrimmings(ctx context.Context) ([]model.TrimmingRecord, error) {
	var trims []model.TrimmingRecord
	result := r.db.WithContext(ctx).
		Preload("Options").
		Order("date DESC, created_at DESC").
		Find(&trims)

	if result.Error != nil {
		return nil, result.Error
	}

	return trims, nil
}

// GetTrimmingByID IDでトリミングを取得
func (r *trimmingRepository) GetTrimmingByID(ctx context.Context, id string) (*model.TrimmingRecord, error) {
	var trim model.TrimmingRecord
	result := r.db.WithContext(ctx).
		Preload("Options").
		First(&trim, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("trimming", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get trimming")
	}

	return &trim, nil
}

// GetTrimmingsByPetID ペットIDでトリミングを取得
func (r *trimmingRepository) GetTrimmingsByPetID(ctx context.Context, petID string) ([]model.TrimmingRecord, error) {
	var trims []model.TrimmingRecord
	result := r.db.WithContext(ctx).
		Preload("Options").
		Where("pet_id = ?", petID).
		Order("date DESC, created_at DESC").
		Find(&trims)

	if result.Error != nil {
		return nil, result.Error
	}

	return trims, nil
}

// GetTrimmingsByOwnerID 飼い主名でトリミングを取得
func (r *trimmingRepository) GetTrimmingsByOwnerID(ctx context.Context, ownerID string) ([]model.TrimmingRecord, error) {
	var trims []model.TrimmingRecord
	result := r.db.WithContext(ctx).
		Preload("Options").
		Where("pet_id IN (SELECT id FROM pets WHERE owner_id = ?)", ownerID).
		Order("date DESC, created_at DESC").
		Find(&trims)

	if result.Error != nil {
		return nil, result.Error
	}

	return trims, nil
}

// GetTrimmingsByStatus ステータスでトリミングを取得
func (r *trimmingRepository) GetTrimmingsByStatus(ctx context.Context, status string) ([]model.TrimmingRecord, error) {
	var trims []model.TrimmingRecord
	result := r.db.WithContext(ctx).
		Preload("Options").
		Where("status = ?", status).
		Order("date DESC, created_at DESC").
		Find(&trims)

	if result.Error != nil {
		return nil, result.Error
	}

	return trims, nil
}

// CreateTrimming トリミングを作成
func (r *trimmingRepository) CreateTrimming(ctx context.Context, trim *model.TrimmingRecord) error {
	// Generate UUID if not set
	if trim.ID == uuid.Nil {
		trim.ID = uuid.New()
	}
	result := r.db.WithContext(ctx).Create(trim)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateTrimming トリミングを更新
func (r *trimmingRepository) UpdateTrimming(ctx context.Context, trim *model.TrimmingRecord) error {
	result := r.db.WithContext(ctx).Save(trim)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteTrimming トリミングを削除
func (r *trimmingRepository) DeleteTrimming(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.TrimmingRecord{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
