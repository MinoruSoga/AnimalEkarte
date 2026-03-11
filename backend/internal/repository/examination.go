package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// ExaminationRepository 検査リポジトリインターフェース
type ExaminationRepository interface {
	GetAllExaminationRecords(ctx context.Context) ([]model.ExaminationRecord, error)
	GetExaminationRecordByID(ctx context.Context, id string) (*model.ExaminationRecord, error)
	GetExaminationRecordsByMedicalRecordID(ctx context.Context, medicalRecordID string) ([]model.ExaminationRecord, error)
	GetExaminationRecordsByPetID(ctx context.Context, petID string) ([]model.ExaminationRecord, error)
	GetExaminationRecordsByOwnerID(ctx context.Context, ownerID string) ([]model.ExaminationRecord, error)
	CreateExaminationRecord(ctx context.Context, exam *model.ExaminationRecord) error
	UpdateExaminationRecord(ctx context.Context, exam *model.ExaminationRecord) error
	DeleteExaminationRecord(ctx context.Context, id string) error
}

// examinationRepository 検査リポジトリ実装
type examinationRepository struct {
	db *gorm.DB
}

// NewExaminationRepository 新しい検査リポジトリを作成
func NewExaminationRepository(db *gorm.DB) ExaminationRepository {
	return &examinationRepository{db: db}
}

// GetAllExaminationRecords 全ての検査記録を取得
func (r *examinationRepository) GetAllExaminationRecords(ctx context.Context) ([]model.ExaminationRecord, error) {
	var exams []model.ExaminationRecord
	result := r.db.WithContext(ctx).
		Preload("Items").
		Order("date DESC, created_at DESC").
		Find(&exams)

	if result.Error != nil {
		return nil, result.Error
	}

	return exams, nil
}

// GetExaminationRecordByID IDで検査記録を取得
func (r *examinationRepository) GetExaminationRecordByID(ctx context.Context, id string) (*model.ExaminationRecord, error) {
	var exam model.ExaminationRecord
	result := r.db.WithContext(ctx).
		Preload("Items").
		First(&exam, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("examination record", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get examination record")
	}

	return &exam, nil
}

// GetExaminationRecordsByMedicalRecordID カルテIDで検査記録を取得
func (r *examinationRepository) GetExaminationRecordsByMedicalRecordID(ctx context.Context, medicalRecordID string) ([]model.ExaminationRecord, error) {
	var exams []model.ExaminationRecord
	result := r.db.WithContext(ctx).
		Preload("Items").
		Where("medical_record_id = ?", medicalRecordID).
		Order("date DESC, created_at DESC").
		Find(&exams)

	if result.Error != nil {
		return nil, result.Error
	}

	return exams, nil
}

// GetExaminationRecordsByPetID ペットIDで検査記録を取得
func (r *examinationRepository) GetExaminationRecordsByPetID(ctx context.Context, petID string) ([]model.ExaminationRecord, error) {
	var exams []model.ExaminationRecord
	result := r.db.WithContext(ctx).
		Preload("Items").
		Where("pet_id = ?", petID).
		Order("date DESC, created_at DESC").
		Find(&exams)

	if result.Error != nil {
		return nil, result.Error
	}

	return exams, nil
}

// GetExaminationRecordsByOwnerID 飼い主IDで検査記録を取得（pets経由）
func (r *examinationRepository) GetExaminationRecordsByOwnerID(ctx context.Context, ownerID string) ([]model.ExaminationRecord, error) {
	var exams []model.ExaminationRecord
	result := r.db.WithContext(ctx).
		Preload("Items").
		Where("pet_id IN (SELECT id FROM pets WHERE owner_id = ?)", ownerID).
		Order("date DESC, created_at DESC").
		Find(&exams)

	if result.Error != nil {
		return nil, result.Error
	}

	return exams, nil
}

// CreateExaminationRecord 検査記録を作成
func (r *examinationRepository) CreateExaminationRecord(ctx context.Context, exam *model.ExaminationRecord) error {
	// Generate UUID if not set
	if exam.ID == uuid.Nil {
		exam.ID = uuid.New()
	}
	result := r.db.WithContext(ctx).Create(exam)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// UpdateExaminationRecord 検査記録を更新
func (r *examinationRepository) UpdateExaminationRecord(ctx context.Context, exam *model.ExaminationRecord) error {
	result := r.db.WithContext(ctx).Save(exam)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteExaminationRecord 検査記録を削除
func (r *examinationRepository) DeleteExaminationRecord(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.ExaminationRecord{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
