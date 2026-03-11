package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ClinicRepository クリニックリポジトリインターフェース
type ClinicRepository interface {
	GetAllClinics(ctx context.Context) ([]model.Clinic, error)
	GetClinicByID(ctx context.Context, id string) (*model.Clinic, error)
	CreateClinic(ctx context.Context, clinic *model.Clinic) error
	UpdateClinic(ctx context.Context, clinic *model.Clinic) error
	DeleteClinic(ctx context.Context, id string) error
}

// clinicRepository クリニックリポジトリ実装
type clinicRepository struct {
	db *gorm.DB
}

// NewClinicRepository 新しいクリニックリポジトリを作成
func NewClinicRepository(db *gorm.DB) ClinicRepository {
	return &clinicRepository{db: db}
}

// GetAllClinics 全てのクリニックを取得
func (r *clinicRepository) GetAllClinics(ctx context.Context) ([]model.Clinic, error) {
	var clinics []model.Clinic
	result := r.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&clinics)

	if result.Error != nil {
		return nil, result.Error
	}

	return clinics, nil
}

// GetClinicByID IDでクリニックを取得
func (r *clinicRepository) GetClinicByID(ctx context.Context, id string) (*model.Clinic, error) {
	var clinic model.Clinic
	result := r.db.WithContext(ctx).
		First(&clinic, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("clinic", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get clinic")
	}

	return &clinic, nil
}

// CreateClinic クリニックを作成
func (r *clinicRepository) CreateClinic(ctx context.Context, clinic *model.Clinic) error {
	// Generate UUID if not set
	if clinic.ID == uuid.Nil {
		clinic.ID = uuid.New()
	}
	result := r.db.WithContext(ctx).Create(clinic)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateClinic クリニックを更新
func (r *clinicRepository) UpdateClinic(ctx context.Context, clinic *model.Clinic) error {
	result := r.db.WithContext(ctx).Save(clinic)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteClinic クリニックを削除
func (r *clinicRepository) DeleteClinic(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Clinic{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
