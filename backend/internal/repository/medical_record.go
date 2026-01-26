package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MedicalRecordRepository カルテリポジトリインターフェース
type MedicalRecordRepository interface {
	GetAllMedicalRecords(ctx context.Context) ([]model.MedicalRecord, error)
	GetMedicalRecordByID(ctx context.Context, id string) (*model.MedicalRecord, error)
	GetMedicalRecordsByPetID(ctx context.Context, petID string) ([]model.MedicalRecord, error)
	GetMedicalRecordsByOwnerID(ctx context.Context, ownerID string) ([]model.MedicalRecord, error)
	CreateMedicalRecord(ctx context.Context, record *model.MedicalRecord) error
	UpdateMedicalRecord(ctx context.Context, record *model.MedicalRecord) error
	DeleteMedicalRecord(ctx context.Context, id string) error
}

// medicalRecordRepository カルテリポジトリ実装
type medicalRecordRepository struct {
	db *gorm.DB
}

// NewMedicalRecordRepository 新しいカルテリポジトリを作成
func NewMedicalRecordRepository(db *gorm.DB) MedicalRecordRepository {
	return &medicalRecordRepository{db: db}
}

// GetAllMedicalRecords 全てのカルテを取得
func (r *medicalRecordRepository) GetAllMedicalRecords(ctx context.Context) ([]model.MedicalRecord, error) {
	var records []model.MedicalRecord
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Order("visit_date DESC, created_at DESC").
		Find(&records)

	if result.Error != nil {
		return nil, result.Error
	}

	return records, nil
}

// GetMedicalRecordByID IDでカルテを取得
func (r *medicalRecordRepository) GetMedicalRecordByID(ctx context.Context, id string) (*model.MedicalRecord, error) {
	var record model.MedicalRecord
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		First(&record, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("medical record with id %s not found", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get medical record")
	}

	return &record, nil
}

// GetMedicalRecordsByPetID ペットIDでカルテを取得
func (r *medicalRecordRepository) GetMedicalRecordsByPetID(ctx context.Context, petID string) ([]model.MedicalRecord, error) {
	var records []model.MedicalRecord
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Where("pet_id = ?", petID).
		Order("visit_date DESC, created_at DESC").
		Find(&records)

	if result.Error != nil {
		return nil, result.Error
	}

	return records, nil
}

// GetMedicalRecordsByOwnerID 飼い主IDでカルテを取得
func (r *medicalRecordRepository) GetMedicalRecordsByOwnerID(ctx context.Context, ownerID string) ([]model.MedicalRecord, error) {
	var records []model.MedicalRecord
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Where("owner_id = ?", ownerID).
		Order("visit_date DESC, created_at DESC").
		Find(&records)

	if result.Error != nil {
		return nil, result.Error
	}

	return records, nil
}

// CreateMedicalRecord カルテを作成
func (r *medicalRecordRepository) CreateMedicalRecord(ctx context.Context, record *model.MedicalRecord) error {
	result := r.db.WithContext(ctx).Create(record)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateMedicalRecord カルテを更新
func (r *medicalRecordRepository) UpdateMedicalRecord(ctx context.Context, record *model.MedicalRecord) error {
	result := r.db.WithContext(ctx).Save(record)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteMedicalRecord カルテを削除
func (r *medicalRecordRepository) DeleteMedicalRecord(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.MedicalRecord{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
