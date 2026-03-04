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
	GetAllExaminations(ctx context.Context) ([]model.Examination, error)
	GetExaminationByID(ctx context.Context, id string) (*model.Examination, error)
	GetExaminationsByPetID(ctx context.Context, petID string) ([]model.Examination, error)
	GetExaminationsByOwnerID(ctx context.Context, ownerID string) ([]model.Examination, error)
	GetExaminationsByStatus(ctx context.Context, status string) ([]model.Examination, error)
	CreateExamination(ctx context.Context, exam *model.Examination) error
	UpdateExamination(ctx context.Context, exam *model.Examination) error
	DeleteExamination(ctx context.Context, id string) error
}

// examinationRepository 検査リポジトリ実装
type examinationRepository struct {
	db *gorm.DB
}

// NewExaminationRepository 新しい検査リポジトリを作成
func NewExaminationRepository(db *gorm.DB) ExaminationRepository {
	return &examinationRepository{db: db}
}

// GetAllExaminations 全ての検査を取得
func (r *examinationRepository) GetAllExaminations(ctx context.Context) ([]model.Examination, error) {
	var exams []model.Examination
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Preload("MedicalRecord").
		Order("examination_date DESC, created_at DESC").
		Find(&exams)

	if result.Error != nil {
		return nil, result.Error
	}

	return exams, nil
}

// GetExaminationByID IDで検査を取得
func (r *examinationRepository) GetExaminationByID(ctx context.Context, id string) (*model.Examination, error) {
	var exam model.Examination
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Preload("MedicalRecord").
		First(&exam, "id = ?", id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("examination with id %s not found", id)
		}
		return nil, apperrors.WrapInternal(result.Error, "failed to get examination")
	}

	return &exam, nil
}

// GetExaminationsByPetID ペットIDで検査を取得
func (r *examinationRepository) GetExaminationsByPetID(ctx context.Context, petID string) ([]model.Examination, error) {
	var exams []model.Examination
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Preload("MedicalRecord").
		Where("pet_id = ?", petID).
		Order("examination_date DESC, created_at DESC").
		Find(&exams)

	if result.Error != nil {
		return nil, result.Error
	}

	return exams, nil
}

// GetExaminationsByOwnerID 飼い主IDで検査を取得
func (r *examinationRepository) GetExaminationsByOwnerID(ctx context.Context, ownerID string) ([]model.Examination, error) {
	var exams []model.Examination
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Preload("MedicalRecord").
		Where("owner_id = ?", ownerID).
		Order("examination_date DESC, created_at DESC").
		Find(&exams)

	if result.Error != nil {
		return nil, result.Error
	}

	return exams, nil
}

// GetExaminationsByStatus ステータスで検査を取得
func (r *examinationRepository) GetExaminationsByStatus(ctx context.Context, status string) ([]model.Examination, error) {
	var exams []model.Examination
	result := r.db.WithContext(ctx).
		Preload("Pet").
		Preload("Owner").
		Preload("MedicalRecord").
		Where("status = ?", status).
		Order("examination_date DESC, created_at DESC").
		Find(&exams)

	if result.Error != nil {
		return nil, result.Error
	}

	return exams, nil
}

// CreateExamination 検査を作成
func (r *examinationRepository) CreateExamination(ctx context.Context, exam *model.Examination) error {
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

// UpdateExamination 検査を更新
func (r *examinationRepository) UpdateExamination(ctx context.Context, exam *model.Examination) error {
	result := r.db.WithContext(ctx).Save(exam)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// DeleteExamination 検査を削除
func (r *examinationRepository) DeleteExamination(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&model.Examination{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
