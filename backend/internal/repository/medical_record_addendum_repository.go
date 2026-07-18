package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

type MedicalRecordAddendumRepository interface {
	Create(ctx context.Context, addendum *model.MedicalRecordAddendum) error
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecordAddendum, error)
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error)
}

type medicalRecordAddendumRepository struct {
	db *gorm.DB
}

func NewMedicalRecordAddendumRepository(db *gorm.DB) MedicalRecordAddendumRepository {
	return &medicalRecordAddendumRepository{db: db}
}

func (r *medicalRecordAddendumRepository) Create(ctx context.Context, addendum *model.MedicalRecordAddendum) error {
	if err := dbOrTx(ctx, r.db).Create(addendum).Error; err != nil {
		return apperrors.FromGORM(err, "medical_record_addendum", "")
	}
	return nil
}

func (r *medicalRecordAddendumRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecordAddendum, error) {
	var addendum model.MedicalRecordAddendum
	err := dbOrTx(ctx, r.db).
		Scopes(clinicScope(clinicID)).
		Where("id = ?", id).
		First(&addendum).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_addendum", fmt.Sprintf("%d", id))
	}
	return &addendum, nil
}

func (r *medicalRecordAddendumRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error) {
	var addenda []*model.MedicalRecordAddendum
	err := dbOrTx(ctx, r.db).
		Scopes(clinicScope(clinicID)).
		Where("medical_record_id = ?", medicalRecordID).
		Order("created_at ASC").
		Find(&addenda).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_addendum", fmt.Sprintf("medical_record_id=%d", medicalRecordID))
	}
	return addenda, nil
}
