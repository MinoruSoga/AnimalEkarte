// Package dailyrecord owns daily_records (+ vital_records/care_logs/staff_notes writes)
// data access (BE8-4 batch6 — leaf domain).
package medicalrecord

// Moved from internal/repository/dailyrecord (BE9-2D ⑤ Batch A roll-up・BE8-4 batch6 subpackage).
// generic Repository/New は entity-specific 名へ改名（①先例）— 外部は facade alias 経由で不変。

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository is the data access interface for daily records.
type DailyRecordRepository interface {
	FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error)
	FindByHospitalizationIDAndDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	FindOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	CreateVitalRecord(ctx context.Context, vr *model.VitalRecord) error
	CreateCareLog(ctx context.Context, cr *model.CareLog) error
	CreateStaffNote(ctx context.Context, sn *model.StaffNote) error
}

type dailyRecordRepository struct {
	db *gorm.DB
}

// New constructs a Repository.
func NewDailyRecordRepository(db *gorm.DB) DailyRecordRepository {
	return &dailyRecordRepository{db: db}
}

func (r *dailyRecordRepository) FindByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error) {
	records := make([]model.DailyRecord, 0)
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN hospitalizations ON hospitalizations.id = daily_records.hospitalization_id AND hospitalizations.clinic_id = daily_records.clinic_id").
		Where("daily_records.clinic_id = ? AND daily_records.hospitalization_id = ?", clinicID, hospitalizationID).
		Order("daily_records.date DESC").
		Preload("VitalRecords", "clinic_id = ?", clinicID).
		Preload("CareLogs", "clinic_id = ?", clinicID).
		Preload("StaffNotes").
		Find(&records).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "daily_record", "")
	}
	return records, nil
}

func (r *dailyRecordRepository) FindByHospitalizationIDAndDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	var record model.DailyRecord
	err := persistence.DBOrTx(ctx, r.db).
		Joins("JOIN hospitalizations ON hospitalizations.id = daily_records.hospitalization_id AND hospitalizations.clinic_id = daily_records.clinic_id").
		Where("daily_records.clinic_id = ? AND daily_records.hospitalization_id = ? AND daily_records.date = ?", clinicID, hospitalizationID, date).
		Preload("VitalRecords", "clinic_id = ?", clinicID).
		Preload("CareLogs", "clinic_id = ?", clinicID).
		Preload("StaffNotes").
		First(&record).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "daily_record", fmt.Sprintf("%d/%s", hospitalizationID, date.Format(time.DateOnly)))
	}
	return &record, nil
}

func (r *dailyRecordRepository) FindOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	record := model.DailyRecord{
		ClinicID:          clinicID,
		HospitalizationID: hospitalizationID,
		Date:              date,
	}
	result := persistence.DBOrTx(ctx, r.db).
		Where(model.DailyRecord{ClinicID: clinicID, HospitalizationID: hospitalizationID, Date: date}).
		FirstOrCreate(&record)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "daily_record", "")
	}
	return &record, nil
}

func (r *dailyRecordRepository) CreateVitalRecord(ctx context.Context, vr *model.VitalRecord) error {
	err := persistence.DBOrTx(ctx, r.db).Create(vr).Error
	if err != nil {
		return apperrors.FromGORM(err, "vital_record", "")
	}
	return nil
}

func (r *dailyRecordRepository) CreateCareLog(ctx context.Context, cr *model.CareLog) error {
	err := persistence.DBOrTx(ctx, r.db).Create(cr).Error
	if err != nil {
		return apperrors.FromGORM(err, "care_log", "")
	}
	return nil
}

func (r *dailyRecordRepository) CreateStaffNote(ctx context.Context, sn *model.StaffNote) error {
	err := persistence.DBOrTx(ctx, r.db).Create(sn).Error
	if err != nil {
		return apperrors.FromGORM(err, "staff_note", "")
	}
	return nil
}
