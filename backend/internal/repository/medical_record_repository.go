package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type MedicalRecordRepository interface {
	FindAll(ctx context.Context, petID *uuid.UUID, page, limit int) ([]model.MedicalRecord, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.MedicalRecord, error)
	FindByRecordNo(ctx context.Context, recordNo string) (*model.MedicalRecord, error)
	Create(ctx context.Context, record *model.MedicalRecord) error
	Update(ctx context.Context, record *model.MedicalRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type medicalRecordRepository struct {
	db *gorm.DB
}

func NewMedicalRecordRepository(db *gorm.DB) MedicalRecordRepository {
	return &medicalRecordRepository{db: db}
}

func (r *medicalRecordRepository) FindAll(ctx context.Context, petID *uuid.UUID, page, limit int) ([]model.MedicalRecord, int64, error) {
	var records []model.MedicalRecord
	var total int64

	q := r.db.WithContext(ctx).Model(&model.MedicalRecord{})
	if petID != nil {
		q = q.Where("pet_id = ?", petID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count medical records")
	}
	if err := q.Offset((page-1)*limit).Limit(limit).Order("date DESC, created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find medical records")
	}
	return records, total, nil
}

func (r *medicalRecordRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.MedicalRecord, error) {
	var record model.MedicalRecord
	if err := r.db.WithContext(ctx).
		Preload("TreatmentItems").
		Preload("VitalEntries").
		Preload("Doctor").
		Preload("Owner").
		Preload("Pet").
		Preload("Diagnosis1Category").
		Preload("Diagnosis1Name").
		Preload("Diagnosis2Category").
		Preload("Diagnosis2Name").
		First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("medical_record", id.String())
		}
		return nil, apperrors.Wrap(err, "find medical record by id")
	}
	return &record, nil
}

func (r *medicalRecordRepository) FindByRecordNo(ctx context.Context, recordNo string) (*model.MedicalRecord, error) {
	var record model.MedicalRecord
	if err := r.db.WithContext(ctx).
		Preload("TreatmentItems").
		Preload("VitalEntries").
		Preload("Doctor").
		First(&record, "record_no = ?", recordNo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("medical_record", recordNo)
		}
		return nil, apperrors.Wrap(err, "find medical record by record_no")
	}
	return &record, nil
}

func (r *medicalRecordRepository) Create(ctx context.Context, record *model.MedicalRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return apperrors.Wrap(err, "create medical record")
	}
	return nil
}

func (r *medicalRecordRepository) Update(ctx context.Context, record *model.MedicalRecord) error {
	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		return apperrors.Wrap(err, "update medical record")
	}
	return nil
}

func (r *medicalRecordRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&model.MedicalRecord{}, "id = ?", id).Error; err != nil {
		return apperrors.Wrap(err, "delete medical record")
	}
	return nil
}
