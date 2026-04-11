package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// DailyRecordRepository は日次記録のデータアクセスインターフェース
type DailyRecordRepository interface {
	ListByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error)
	FindByHospitalizationIDAndDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	GetOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	CreateVitalRecord(ctx context.Context, vr *model.VitalRecord) error
	CreateCareLog(ctx context.Context, cr *model.CareLog) error
	CreateStaffNote(ctx context.Context, sn *model.StaffNote) error
}

type dailyRecordRepository struct {
	db *gorm.DB
}

// NewDailyRecordRepository は DailyRecordRepository を初期化して返す
func NewDailyRecordRepository(db *gorm.DB) DailyRecordRepository {
	return &dailyRecordRepository{db: db}
}

func (r *dailyRecordRepository) ListByHospitalizationID(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error) {
	records := make([]model.DailyRecord, 0)
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND hospitalization_id = ?", clinicID, hospitalizationID).
		Order("date DESC").
		Preload("VitalRecords").
		Preload("CareLogs").
		Preload("StaffNotes").
		Find(&records).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "daily_record", "")
	}
	return records, nil
}

func (r *dailyRecordRepository) FindByHospitalizationIDAndDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	var record model.DailyRecord
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND hospitalization_id = ? AND date = ?", clinicID, hospitalizationID, date).
		Preload("VitalRecords").
		Preload("CareLogs").
		Preload("StaffNotes").
		First(&record).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "daily_record", fmt.Sprintf("%d/%s", hospitalizationID, date.Format("2006-01-02")))
	}
	return &record, nil
}

func (r *dailyRecordRepository) GetOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	record := model.DailyRecord{
		ClinicID:          clinicID,
		HospitalizationID: hospitalizationID,
		Date:              date,
	}
	result := r.db.WithContext(ctx).
		Where(model.DailyRecord{ClinicID: clinicID, HospitalizationID: hospitalizationID, Date: date}).
		FirstOrCreate(&record)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "daily_record", "")
	}
	return &record, nil
}

func (r *dailyRecordRepository) CreateVitalRecord(ctx context.Context, vr *model.VitalRecord) error {
	err := r.db.WithContext(ctx).Create(vr).Error
	if err != nil {
		return apperrors.FromGORM(err, "vital_record", "")
	}
	return nil
}

func (r *dailyRecordRepository) CreateCareLog(ctx context.Context, cr *model.CareLog) error {
	err := r.db.WithContext(ctx).Create(cr).Error
	if err != nil {
		return apperrors.FromGORM(err, "care_log_record", "")
	}
	return nil
}

func (r *dailyRecordRepository) CreateStaffNote(ctx context.Context, sn *model.StaffNote) error {
	err := r.db.WithContext(ctx).Create(sn).Error
	if err != nil {
		return apperrors.FromGORM(err, "staff_note_record", "")
	}
	return nil
}
