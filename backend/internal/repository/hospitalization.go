package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// HospitalizationRepository 入院/ホテルリポジトリインターフェース
type HospitalizationRepository interface {
	GetAllHospitalizations(ctx context.Context) ([]model.Hospitalization, error)
	GetHospitalizationByID(ctx context.Context, id string) (*model.Hospitalization, error)
	GetHospitalizationsByPetID(ctx context.Context, petID string) ([]model.Hospitalization, error)
	GetHospitalizationsByOwnerID(ctx context.Context, ownerID string) ([]model.Hospitalization, error)
	GetHospitalizationsByStatus(ctx context.Context, status string) ([]model.Hospitalization, error)
	CreateHospitalization(ctx context.Context, hosp *model.Hospitalization) error
	UpdateHospitalization(ctx context.Context, hosp *model.Hospitalization) error
	DeleteHospitalization(ctx context.Context, id string) error
}

// hospitalizationRepository 入院/ホテルリポジトリ実装
type hospitalizationRepository struct {
	db *gorm.DB
}

// NewHospitalizationRepository 新しい入院/ホテルリポジトリを作成
func NewHospitalizationRepository(db *gorm.DB) HospitalizationRepository {
	return &hospitalizationRepository{db: db}
}

// GetAllHospitalizations 全ての入院/ホテルを取得
func (r *hospitalizationRepository) GetAllHospitalizations(ctx context.Context) ([]model.Hospitalization, error) {
	var hosps []model.Hospitalization
	result := r.db.WithContext(ctx).
		Order("start_date DESC, created_at DESC").
		Find(&hosps)

	if result.Error != nil {
		return nil, result.Error
	}

	return hosps, nil
}

// GetHospitalizationByID IDで入院/ホテルを取得
func (r *hospitalizationRepository) GetHospitalizationByID(ctx context.Context, id string) (*model.Hospitalization, error) {
	var hosp model.Hospitalization
	result := r.db.WithContext(ctx).
		Preload("CarePlanItems").
		Preload("DailyRecords").
		Preload("TreatmentPlans").
		First(&hosp, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("hospitalization", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get hospitalization")
	}

	return &hosp, nil
}

// GetHospitalizationsByPetID ペットIDで入院/ホテルを取得
func (r *hospitalizationRepository) GetHospitalizationsByPetID(ctx context.Context, petID string) ([]model.Hospitalization, error) {
	var hosps []model.Hospitalization
	result := r.db.WithContext(ctx).
		Where("pet_id = ?", petID).
		Order("start_date DESC, created_at DESC").
		Find(&hosps)

	if result.Error != nil {
		return nil, result.Error
	}

	return hosps, nil
}

// GetHospitalizationsByOwnerID 飼い主IDで入院/ホテルを取得
func (r *hospitalizationRepository) GetHospitalizationsByOwnerID(ctx context.Context, ownerID string) ([]model.Hospitalization, error) {
	var hosps []model.Hospitalization
	result := r.db.WithContext(ctx).
		Where("owner_id = ?", ownerID).
		Order("start_date DESC, created_at DESC").
		Find(&hosps)

	if result.Error != nil {
		return nil, result.Error
	}

	return hosps, nil
}

// GetHospitalizationsByStatus ステータスで入院/ホテルを取得
func (r *hospitalizationRepository) GetHospitalizationsByStatus(ctx context.Context, status string) ([]model.Hospitalization, error) {
	var hosps []model.Hospitalization
	result := r.db.WithContext(ctx).
		Where("status = ?", status).
		Order("start_date DESC, created_at DESC").
		Find(&hosps)

	if result.Error != nil {
		return nil, result.Error
	}

	return hosps, nil
}

// CreateHospitalization 入院/ホテルを作成
func (r *hospitalizationRepository) CreateHospitalization(ctx context.Context, hosp *model.Hospitalization) error {
	// Generate UUID if not set
	if hosp.ID == uuid.Nil {
		hosp.ID = uuid.New()
	}
	result := r.db.WithContext(ctx).Create(hosp)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateHospitalization 入院/ホテルを更新
func (r *hospitalizationRepository) UpdateHospitalization(ctx context.Context, hosp *model.Hospitalization) error {
	result := r.db.WithContext(ctx).Save(hosp)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteHospitalization 入院/ホテルを削除
func (r *hospitalizationRepository) DeleteHospitalization(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Hospitalization{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
