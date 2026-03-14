package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// DailyRecordRepository は日次記録のデータアクセスインターフェース
type DailyRecordRepository interface {
	ListByHospitalizationID(ctx context.Context, hospitalizationID uint64) ([]model.DailyRecord, error)
	FindByHospitalizationIDAndDate(ctx context.Context, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	GetOrCreateByDate(ctx context.Context, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	CreateVitalRecord(ctx context.Context, vr *model.VitalRecord) error
	CreateCareLogRecord(ctx context.Context, cr *model.CareLogRecord) error
	CreateStaffNoteRecord(ctx context.Context, sn *model.StaffNoteRecord) error
}

type dailyRecordRepository struct {
	db *gorm.DB
}

// NewDailyRecordRepository は DailyRecordRepository を初期化して返す
func NewDailyRecordRepository(db *gorm.DB) DailyRecordRepository {
	return &dailyRecordRepository{db: db}
}

func (r *dailyRecordRepository) ListByHospitalizationID(ctx context.Context, hospitalizationID uint64) ([]model.DailyRecord, error) {
	records := make([]model.DailyRecord, 0)
	if err := r.db.WithContext(ctx).
		Where("hospitalization_id = ?", hospitalizationID).
		Order("date DESC").
		Preload("VitalRecords").
		Preload("CareLogRecords").
		Preload("StaffNoteRecords").
		Find(&records).Error; err != nil {
		return nil, apperrors.Wrap(err, "list daily records")
	}
	return records, nil
}

func (r *dailyRecordRepository) FindByHospitalizationIDAndDate(ctx context.Context, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	var record model.DailyRecord
	if err := r.db.WithContext(ctx).
		Where("hospitalization_id = ? AND date = ?", hospitalizationID, date).
		Preload("VitalRecords").
		Preload("CareLogRecords").
		Preload("StaffNoteRecords").
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.WrapNotFound("daily_record", fmt.Sprintf("%d/%s", hospitalizationID, date.Format("2006-01-02")))
		}
		return nil, apperrors.Wrap(err, "find daily record by date")
	}
	return &record, nil
}

func (r *dailyRecordRepository) GetOrCreateByDate(ctx context.Context, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	record := model.DailyRecord{
		HospitalizationID: hospitalizationID,
		Date:              date,
	}
	result := r.db.WithContext(ctx).
		Where(model.DailyRecord{HospitalizationID: hospitalizationID, Date: date}).
		FirstOrCreate(&record)
	if result.Error != nil {
		return nil, apperrors.Wrap(result.Error, "get or create daily record")
	}
	return &record, nil
}

func (r *dailyRecordRepository) CreateVitalRecord(ctx context.Context, vr *model.VitalRecord) error {
	if err := r.db.WithContext(ctx).Create(vr).Error; err != nil {
		return apperrors.Wrap(err, "create vital record")
	}
	return nil
}

func (r *dailyRecordRepository) CreateCareLogRecord(ctx context.Context, cr *model.CareLogRecord) error {
	if err := r.db.WithContext(ctx).Create(cr).Error; err != nil {
		return apperrors.Wrap(err, "create care log record")
	}
	return nil
}

func (r *dailyRecordRepository) CreateStaffNoteRecord(ctx context.Context, sn *model.StaffNoteRecord) error {
	if err := r.db.WithContext(ctx).Create(sn).Error; err != nil {
		return apperrors.Wrap(err, "create staff note record")
	}
	return nil
}
