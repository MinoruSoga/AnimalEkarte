package medicalrecord

// Moved from internal/repository (BE9-2D ⑦ Batch A). 旧 package-private helper は repohelpers
// 同等物へ置換（同一述語/ambient-tx参加）。外部は internal/repository の facade alias 経由で不変。

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
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
	if err := persistence.DBOrTx(ctx, r.db).Create(addendum).Error; err != nil {
		return apperrors.FromGORM(err, "medical_record_addendum", "")
	}
	return nil
}

func (r *medicalRecordAddendumRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecordAddendum, error) {
	var addendum model.MedicalRecordAddendum
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN medical_records ON medical_records.id = medical_record_addenda.medical_record_id AND medical_records.clinic_id = medical_record_addenda.clinic_id AND medical_records.deleted_at IS NULL").
		Where("medical_record_addenda.clinic_id = ? AND medical_record_addenda.id = ?", clinicID, id).
		First(&addendum).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_addendum", fmt.Sprintf("%d", id))
	}
	return &addendum, nil
}

func (r *medicalRecordAddendumRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error) {
	var addenda []*model.MedicalRecordAddendum
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN medical_records ON medical_records.id = medical_record_addenda.medical_record_id AND medical_records.clinic_id = medical_record_addenda.clinic_id AND medical_records.deleted_at IS NULL").
		Where("medical_record_addenda.clinic_id = ? AND medical_record_addenda.medical_record_id = ?", clinicID, medicalRecordID).
		Order("medical_record_addenda.created_at ASC").
		Find(&addenda).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "medical_record_addendum", fmt.Sprintf("medical_record_id=%d", medicalRecordID))
	}
	return addenda, nil
}
