package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type MedicalRecordRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.MedicalRecord, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
	FindByRecordNo(ctx context.Context, clinicID uint64, recordNo string) (*model.MedicalRecord, error)
	Create(ctx context.Context, record *model.MedicalRecord) error
	Update(ctx context.Context, record *model.MedicalRecord) error
	Delete(ctx context.Context, clinicID, id uint64) error
}

type medicalRecordRepository struct {
	db *gorm.DB
}

func NewMedicalRecordRepository(db *gorm.DB) MedicalRecordRepository {
	return &medicalRecordRepository{db: db}
}

func (r *medicalRecordRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, page, limit int) ([]model.MedicalRecord, int64, error) {
	records := make([]model.MedicalRecord, 0)
	var total int64

	q := r.db.WithContext(ctx).Model(&model.MedicalRecord{}).Where("clinic_id = ?", clinicID)
	if petID != nil {
		q = q.Where("pet_id = ?", *petID)
	}
	if ownerID != nil {
		q = q.Where("owner_id = ?", *ownerID)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "count medical records")
	}
	if err := q.Offset((page - 1) * limit).Limit(limit).Order("date DESC, created_at DESC").
		Preload("Owner").Preload("Pet.AnimalSpecies").Preload("Doctor").Preload("Inquiry").Preload("Billing").
		Find(&records).Error; err != nil {
		return nil, 0, apperrors.Wrap(err, "find medical records")
	}
	return records, total, nil
}

func (r *medicalRecordRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error) {
	var record model.MedicalRecord
	if err := r.db.WithContext(ctx).
		Preload("Treatments").
		Preload("Vitals").
		Preload("Doctor").
		Preload("Owner").
		Preload("Pet.AnimalSpecies").
		First(&record, "id = ? AND clinic_id = ?", id, clinicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", id))
		}
		return nil, apperrors.Wrap(err, "find medical record by id")
	}
	return &record, nil
}

func (r *medicalRecordRepository) FindByRecordNo(ctx context.Context, clinicID uint64, recordNo string) (*model.MedicalRecord, error) {
	var record model.MedicalRecord
	if err := r.db.WithContext(ctx).
		Preload("Treatments").
		Preload("Vitals").
		Preload("Doctor").
		First(&record, "record_no = ? AND clinic_id = ?", recordNo, clinicID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("medical_record", recordNo)
		}
		return nil, apperrors.Wrap(err, "find medical record by record_no")
	}
	return &record, nil
}

func (r *medicalRecordRepository) Create(ctx context.Context, record *model.MedicalRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		if isUniqueConstraintErr(err) {
			return apperrors.WrapAlreadyExists("medical_record", record.RecordNo)
		}
		return apperrors.Wrap(err, "create medical record")
	}
	return nil
}

// buildMedicalRecordUpdateFields は MedicalRecord から非ゼロ値/非nilフィールドのみを map に変換する。
// GORM の zero-value スキップ問題を回避し、PATCH セマンティクスを実現する。
func buildMedicalRecordUpdateFields(record *model.MedicalRecord) map[string]any {
	fields := make(map[string]any)
	if !record.Date.IsZero() {
		fields["date"] = record.Date
	}
	if record.Status != "" {
		fields["status"] = record.Status
	}
	if record.OwnerID != nil {
		fields["owner_id"] = record.OwnerID
	}
	if record.PetID != nil {
		fields["pet_id"] = record.PetID
	}
	if record.DoctorID != nil {
		fields["doctor_id"] = record.DoctorID
	}
	if record.ReservationAppointmentID != nil {
		fields["reservation_appointment_id"] = record.ReservationAppointmentID
	}
	return fields
}

func (r *medicalRecordRepository) Update(ctx context.Context, record *model.MedicalRecord) error {
	result := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Where("id = ? AND clinic_id = ?", record.ID, record.ClinicID).
		Updates(buildMedicalRecordUpdateFields(record))
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "update medical record")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", record.ID))
	}
	return nil
}

func (r *medicalRecordRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&model.MedicalRecord{}, "id = ? AND clinic_id = ?", id, clinicID)
	if result.Error != nil {
		return apperrors.Wrap(result.Error, "delete medical record")
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", id))
	}
	return nil
}
