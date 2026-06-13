package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

// CreateVitalRecordInput はバイタル記録作成のサービス入力DTO
type CreateVitalRecordInput struct {
	Time            time.Time
	Temperature     *float64
	HeartRate       *int
	RespirationRate *int
	Weight          *float64
	Notes           string
	StaffID         *uint64
}

// CreateCareLogInput はケアログ記録作成のサービス入力DTO
type CreateCareLogInput struct {
	Time    time.Time
	Type    string
	Status  string
	Value   string
	StaffID *uint64
	Notes   string
}

// CreateStaffNoteInput はスタッフメモ記録作成のサービス入力DTO
type CreateStaffNoteInput struct {
	Time    time.Time
	Content string
	StaffID *uint64
}

// DailyRecordService は日次記録のビジネスロジックインターフェース
type DailyRecordService interface {
	List(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error)
	FindOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	AddVitalRecord(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateVitalRecordInput) (*model.DailyRecord, error)
	AddCareLog(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateCareLogInput) (*model.DailyRecord, error)
	AddStaffNote(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateStaffNoteInput) (*model.DailyRecord, error)
}

type dailyRecordService struct {
	repo repository.DailyRecordRepository
}

// NewDailyRecordService は DailyRecordService を初期化して返す
func NewDailyRecordService(repo repository.DailyRecordRepository) DailyRecordService {
	return &dailyRecordService{repo: repo}
}

func (s *dailyRecordService) List(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error) {
	result, err := s.repo.FindByHospitalizationID(ctx, clinicID, hospitalizationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list daily record", "error", err)
		return nil, apperrors.Wrap(err, "failed to list daily record")
	}
	return result, nil
}

func (s *dailyRecordService) FindOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	result, err := s.repo.FindOrCreateByDate(ctx, clinicID, hospitalizationID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get or create daily record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get or create daily record")
	}
	return result, nil
}

func (s *dailyRecordService) AddVitalRecord(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateVitalRecordInput) (*model.DailyRecord, error) {
	daily, err := s.repo.FindOrCreateByDate(ctx, clinicID, hospitalizationID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get or create daily record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get or create daily record")
	}

	dailyID := daily.ID
	vr := &model.VitalRecord{
		ClinicID:        clinicID,
		DailyRecordID:   &dailyID,
		RecordedAt:      input.Time,
		Temperature:     input.Temperature,
		HeartRate:       input.HeartRate,
		RespirationRate: input.RespirationRate,
		Weight:          input.Weight,
		Notes:           input.Notes,
		StaffID:         input.StaffID,
	}
	if err := s.repo.CreateVitalRecord(ctx, vr); err != nil {
		slog.ErrorContext(ctx, "failed to create vital record", "error", err)
		return nil, apperrors.Wrap(err, "failed to create vital record")
	}

	slog.InfoContext(ctx, "vital record added",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("daily_record_id", daily.ID),
		slog.Uint64("vital_record_id", vr.ID))

	result, err := s.repo.FindByHospitalizationIDAndDate(ctx, clinicID, hospitalizationID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get daily record after adding vital", "error", err)
		return nil, apperrors.Wrap(err, "failed to get daily record after adding vital")
	}
	return result, nil
}

func (s *dailyRecordService) AddCareLog(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateCareLogInput) (*model.DailyRecord, error) {
	careLogType := model.CareLogType(input.Type)
	switch careLogType {
	case model.CareLogTypeFood, model.CareLogTypeExcretion, model.CareLogTypeMedicine,
		model.CareLogTypeTreatment, model.CareLogTypeOther:
		// valid
	default:
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("invalid care log type: %s", input.Type))
	}

	status := model.CareLogStatusCompleted
	if input.Status != "" {
		careLogStatus := model.CareLogStatus(input.Status)
		switch careLogStatus {
		case model.CareLogStatusCompleted, model.CareLogStatusPartial, model.CareLogStatusSkipped:
			status = careLogStatus
		default:
			return nil, apperrors.WrapInvalidInput(fmt.Sprintf("invalid care log status: %s", input.Status))
		}
	}

	daily, err := s.repo.FindOrCreateByDate(ctx, clinicID, hospitalizationID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get or create daily record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get or create daily record")
	}

	cr := &model.CareLog{
		ClinicID:      clinicID,
		DailyRecordID: daily.ID,
		Time:          input.Time,
		Type:          careLogType,
		Status:        status,
		Value:         input.Value,
		StaffID:       input.StaffID,
		Notes:         input.Notes,
	}
	if err := s.repo.CreateCareLog(ctx, cr); err != nil {
		slog.ErrorContext(ctx, "failed to create care log record", "error", err)
		return nil, apperrors.Wrap(err, "failed to create care log record")
	}

	slog.InfoContext(ctx, "care log record added",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("daily_record_id", daily.ID),
		slog.Uint64("care_log_id", cr.ID))

	result, err := s.repo.FindByHospitalizationIDAndDate(ctx, clinicID, hospitalizationID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get daily record after adding care log", "error", err)
		return nil, apperrors.Wrap(err, "failed to get daily record after adding care log")
	}
	return result, nil
}

func (s *dailyRecordService) AddStaffNote(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateStaffNoteInput) (*model.DailyRecord, error) {
	daily, err := s.repo.FindOrCreateByDate(ctx, clinicID, hospitalizationID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get or create daily record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get or create daily record")
	}

	sn := &model.StaffNote{
		DailyRecordID: daily.ID,
		Time:          input.Time,
		Content:       input.Content,
		StaffID:       input.StaffID,
	}
	if err := s.repo.CreateStaffNote(ctx, sn); err != nil {
		slog.ErrorContext(ctx, "failed to create staff note record", "error", err)
		return nil, apperrors.Wrap(err, "failed to create staff note record")
	}

	slog.InfoContext(ctx, "staff note record added",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("daily_record_id", daily.ID),
		slog.Uint64("staff_note_id", sn.ID))

	result, err := s.repo.FindByHospitalizationIDAndDate(ctx, clinicID, hospitalizationID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get daily record after adding staff note", "error", err)
		return nil, apperrors.Wrap(err, "failed to get daily record after adding staff note")
	}
	return result, nil
}
