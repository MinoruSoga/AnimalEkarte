package medicalrecord

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
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
	Time    string // "HH:MM:SS"（TIME 列は string 規約）
	Type    string
	Status  string
	Value   string
	StaffID *uint64
	Notes   string
}

// CreateStaffNoteInput はスタッフメモ記録作成のサービス入力DTO
type CreateStaffNoteInput struct {
	Time    string // "HH:MM:SS"（TIME 列は string 規約）
	Content string
	StaffID *uint64
}

// DailyRecordService は日次記録のビジネスロジックインターフェース
type DailyRecordService interface {
	List(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error)
	GetByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	FindOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error)
	AddVitalRecord(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateVitalRecordInput) (*model.DailyRecord, error)
	AddCareLog(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateCareLogInput) (*model.DailyRecord, error)
	AddStaffNote(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateStaffNoteInput) (*model.DailyRecord, error)
}

type dailyRecordService struct {
	repo     DailyRecordRepository
	hospRepo HospitalizationRepository
	tx       Transactor
}

// NewDailyRecordService は DailyRecordService を初期化して返す
func NewDailyRecordService(repo DailyRecordRepository, hospRepo HospitalizationRepository, tx Transactor) DailyRecordService {
	return &dailyRecordService{repo: repo, hospRepo: hospRepo, tx: tx}
}

// verifyHospitalizationOwnership は親入院が caller の clinic に属することを確認する。
// 異院 hospitalization_id での DailyRecord 読取/書込を NotFound で拒否する（AUD-003）。
func (s *dailyRecordService) verifyHospitalizationOwnership(ctx context.Context, clinicID, hospitalizationID uint64) error {
	_, err := s.loadHospitalization(ctx, clinicID, hospitalizationID)
	return err
}

// loadHospitalization は所有権検証を兼ねて親入院を返す（AUD-003 再利用・二重 FindByID を避ける）。
func (s *dailyRecordService) loadHospitalization(ctx context.Context, clinicID, hospitalizationID uint64) (*model.Hospitalization, error) {
	hosp, err := s.hospRepo.FindByID(ctx, clinicID, hospitalizationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to verify hospitalization ownership", "error", err)
		return nil, apperrors.Wrap(err, "failed to verify hospitalization ownership")
	}
	return hosp, nil
}

func (s *dailyRecordService) List(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error) {
	if err := s.verifyHospitalizationOwnership(ctx, clinicID, hospitalizationID); err != nil {
		return nil, err
	}
	result, err := s.repo.FindByHospitalizationID(ctx, clinicID, hospitalizationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list daily record", "error", err)
		return nil, apperrors.Wrap(err, "failed to list daily record")
	}
	return result, nil
}

// GetByDate は指定日の日次記録を読み取り専用で返す（未存在は NotFound。FirstOrCreate しない）。
func (s *dailyRecordService) GetByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	if err := s.verifyHospitalizationOwnership(ctx, clinicID, hospitalizationID); err != nil {
		return nil, err
	}
	result, err := s.repo.FindByHospitalizationIDAndDate(ctx, clinicID, hospitalizationID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get daily record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get daily record")
	}
	return result, nil
}

func (s *dailyRecordService) FindOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	if err := s.verifyHospitalizationOwnership(ctx, clinicID, hospitalizationID); err != nil {
		return nil, err
	}
	result, err := s.repo.FindOrCreateByDate(ctx, clinicID, hospitalizationID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get or create daily record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get or create daily record")
	}
	return result, nil
}

func (s *dailyRecordService) AddVitalRecord(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateVitalRecordInput) (*model.DailyRecord, error) {
	hosp, err := s.loadHospitalization(ctx, clinicID, hospitalizationID)
	if err != nil {
		return nil, err
	}

	var vitalID uint64
	var dailyID uint64
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		daily, err := s.repo.FindOrCreateByDate(txCtx, clinicID, hospitalizationID, date)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get or create daily record", "error", err)
			return apperrors.Wrap(err, "failed to get or create daily record")
		}
		dailyID = daily.ID
		vr := &model.VitalRecord{
			ClinicID:        clinicID,
			PetID:           hosp.PetID, // AUD-006: 親入院から必須 PetID を解決
			DailyRecordID:   &dailyID,
			RecordedAt:      input.Time,
			Temperature:     input.Temperature,
			HeartRate:       input.HeartRate,
			RespirationRate: input.RespirationRate,
			Weight:          input.Weight,
			Notes:           input.Notes,
			StaffID:         input.StaffID,
		}
		if err := s.repo.CreateVitalRecord(txCtx, vr); err != nil {
			slog.ErrorContext(txCtx, "failed to create vital record", "error", err)
			return apperrors.Wrap(err, "failed to create vital record")
		}
		vitalID = vr.ID
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "vital record added",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("daily_record_id", dailyID),
		slog.Uint64("vital_record_id", vitalID))

	result, err := s.repo.FindByHospitalizationIDAndDate(ctx, clinicID, hospitalizationID, date)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get daily record after adding vital", "error", err)
		return nil, apperrors.Wrap(err, "failed to get daily record after adding vital")
	}
	return result, nil
}

func (s *dailyRecordService) AddCareLog(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateCareLogInput) (*model.DailyRecord, error) {
	if err := s.verifyHospitalizationOwnership(ctx, clinicID, hospitalizationID); err != nil {
		return nil, err
	}

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
	if err := s.verifyHospitalizationOwnership(ctx, clinicID, hospitalizationID); err != nil {
		return nil, err
	}
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
