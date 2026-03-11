package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// VaccinationRepository ワクチン接種リポジトリインターフェース
type VaccinationRepository interface {
	GetAllVaccinations(ctx context.Context) ([]model.VaccinationRecord, error)
	GetVaccinationByID(ctx context.Context, id string) (*model.VaccinationRecord, error)
	GetVaccinationsByPetID(ctx context.Context, petID string) ([]model.VaccinationRecord, error)
	GetVaccinationsByOwnerID(ctx context.Context, ownerID string) ([]model.VaccinationRecord, error)
	CreateVaccination(ctx context.Context, vac *model.VaccinationRecord) error
	UpdateVaccination(ctx context.Context, vac *model.VaccinationRecord) error
	DeleteVaccination(ctx context.Context, id string) error
}

// vaccinationRepository ワクチン接種リポジトリ実装
type vaccinationRepository struct {
	db *gorm.DB
}

// NewVaccinationRepository 新しいワクチン接種リポジトリを作成
func NewVaccinationRepository(db *gorm.DB) VaccinationRepository {
	return &vaccinationRepository{db: db}
}

// GetAllVaccinations 全てのワクチン接種記録を取得
func (r *vaccinationRepository) GetAllVaccinations(ctx context.Context) ([]model.VaccinationRecord, error) {
	var vacs []model.VaccinationRecord
	result := r.db.WithContext(ctx).
		Order("date DESC, created_at DESC").
		Find(&vacs)

	if result.Error != nil {
		return nil, result.Error
	}

	return vacs, nil
}

// GetVaccinationByID IDでワクチン接種記録を取得
func (r *vaccinationRepository) GetVaccinationByID(ctx context.Context, id string) (*model.VaccinationRecord, error) {
	var vac model.VaccinationRecord
	result := r.db.WithContext(ctx).
		First(&vac, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("vaccination with id %s not found", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get vaccination")
	}

	return &vac, nil
}

// GetVaccinationsByPetID ペットIDでワクチン接種記録を取得
func (r *vaccinationRepository) GetVaccinationsByPetID(ctx context.Context, petID string) ([]model.VaccinationRecord, error) {
	var vacs []model.VaccinationRecord
	result := r.db.WithContext(ctx).
		Where("pet_id = ?", petID).
		Order("date DESC, created_at DESC").
		Find(&vacs)

	if result.Error != nil {
		return nil, result.Error
	}

	return vacs, nil
}

// GetVaccinationsByOwnerID 飼い主名でワクチン接種記録を取得
func (r *vaccinationRepository) GetVaccinationsByOwnerID(ctx context.Context, ownerID string) ([]model.VaccinationRecord, error) {
	var vacs []model.VaccinationRecord
	result := r.db.WithContext(ctx).
		Where("owner_name IN (SELECT owner_name FROM owners WHERE id = ?)", ownerID).
		Order("date DESC, created_at DESC").
		Find(&vacs)

	if result.Error != nil {
		return nil, result.Error
	}

	return vacs, nil
}

// CreateVaccination ワクチン接種記録を作成
func (r *vaccinationRepository) CreateVaccination(ctx context.Context, vac *model.VaccinationRecord) error {
	// Generate UUID if not set
	if vac.ID == uuid.Nil {
		vac.ID = uuid.New()
	}
	result := r.db.WithContext(ctx).Create(vac)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateVaccination ワクチン接種記録を更新
func (r *vaccinationRepository) UpdateVaccination(ctx context.Context, vac *model.VaccinationRecord) error {
	result := r.db.WithContext(ctx).Save(vac)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteVaccination ワクチン接種記録を削除
func (r *vaccinationRepository) DeleteVaccination(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.VaccinationRecord{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
