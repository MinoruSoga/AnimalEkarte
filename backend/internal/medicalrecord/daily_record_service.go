package medicalrecord

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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
	repo             DailyRecordRepository
	hospRepo         HospitalizationRepository
	ownerPets        sharedkernel.OwnerPetLinkVerifier
	staffs           clinicalStaffLocker
	staffAssignments clinicalStaffAssignmentLocker
	tx               Transactor
}

// NewDailyRecordService は owner/pet の clinic relation を検証する基本構成を返す。
// staff_id を受ける経路は WithRelationValidation で staff 依存も注入する。
func NewDailyRecordService(
	repo DailyRecordRepository,
	hospRepo HospitalizationRepository,
	ownerPets sharedkernel.OwnerPetLinkVerifier,
	tx Transactor,
) DailyRecordService {
	return &dailyRecordService{
		repo:      repo,
		hospRepo:  hospRepo,
		ownerPets: ownerPets,
		tx:        tx,
	}
}

// NewDailyRecordServiceWithRelationValidation は owner/pet と staff_id の active staff +
// clinic assignment 検証依存を注入する。親入院ロック・relation 検証・子保存は同じ
// tx callback 内で実行する。
func NewDailyRecordServiceWithRelationValidation(
	repo DailyRecordRepository,
	hospRepo HospitalizationRepository,
	ownerPets sharedkernel.OwnerPetLinkVerifier,
	staffs clinicalStaffLocker,
	staffAssignments clinicalStaffAssignmentLocker,
	tx Transactor,
) DailyRecordService {
	return &dailyRecordService{
		repo:             repo,
		hospRepo:         hospRepo,
		ownerPets:        ownerPets,
		staffs:           staffs,
		staffAssignments: staffAssignments,
		tx:               tx,
	}
}

// loadHospitalization は所有権検証を兼ねて親入院を返す（AUD-003 再利用・二重 FindByID を避ける）。
func (s *dailyRecordService) loadHospitalization(ctx context.Context, clinicID, hospitalizationID uint64) (*model.Hospitalization, error) {
	if s.hospRepo == nil {
		return nil, apperrors.WrapInternalServerError("daily record hospitalization dependency is required")
	}
	hosp, err := s.hospRepo.FindByID(ctx, clinicID, hospitalizationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to verify hospitalization ownership", "error", err)
		return nil, apperrors.Wrap(err, "failed to verify hospitalization ownership")
	}
	if err := validateDailyRecordHospitalization(hosp, clinicID, hospitalizationID); err != nil {
		return nil, err
	}
	if err := s.validateHospitalizationOwnerPet(ctx, clinicID, hosp); err != nil {
		return nil, err
	}
	return hosp, nil
}

func (s *dailyRecordService) List(ctx context.Context, clinicID, hospitalizationID uint64) ([]model.DailyRecord, error) {
	if err := validateDailyRecordIDs(clinicID, hospitalizationID); err != nil {
		return nil, err
	}
	if s.repo == nil || s.hospRepo == nil {
		return nil, apperrors.WrapInternalServerError("daily record persistence dependencies are required")
	}
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("daily record transaction dependency is required")
	}

	var result []model.DailyRecord
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		hospitalization, err := s.loadHospitalization(txCtx, clinicID, hospitalizationID)
		if err != nil {
			return err
		}
		result, err = s.repo.FindByHospitalizationID(txCtx, clinicID, hospitalizationID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to list daily record", "error", err)
			return apperrors.Wrap(err, "failed to list daily record")
		}
		for i := range result {
			if err := s.validateDailyRecordRead(txCtx, clinicID, hospitalization, &result[i], nil); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// GetByDate は指定日の日次記録を読み取り専用で返す（未存在は NotFound。FirstOrCreate しない）。
func (s *dailyRecordService) GetByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	if err := validateDailyRecordIDs(clinicID, hospitalizationID); err != nil {
		return nil, err
	}
	if s.repo == nil || s.hospRepo == nil {
		return nil, apperrors.WrapInternalServerError("daily record persistence dependencies are required")
	}
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("daily record transaction dependency is required")
	}

	var result *model.DailyRecord
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		hospitalization, err := s.loadHospitalization(txCtx, clinicID, hospitalizationID)
		if err != nil {
			return err
		}
		result, err = s.repo.FindByHospitalizationIDAndDate(txCtx, clinicID, hospitalizationID, date)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get daily record", "error", err)
			return apperrors.Wrap(err, "failed to get daily record")
		}
		return s.validateDailyRecordRead(txCtx, clinicID, hospitalization, result, &date)
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *dailyRecordService) FindOrCreateByDate(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time) (*model.DailyRecord, error) {
	if err := validateDailyRecordIDs(clinicID, hospitalizationID); err != nil {
		return nil, err
	}
	if s.repo == nil || s.hospRepo == nil {
		return nil, apperrors.WrapInternalServerError("daily record persistence dependencies are required")
	}
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("daily record transaction dependency is required")
	}
	var result *model.DailyRecord
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := s.lockHospitalizationForWrite(txCtx, clinicID, hospitalizationID); err != nil {
			return err
		}
		var err error
		result, err = s.repo.FindOrCreateByDate(txCtx, clinicID, hospitalizationID, date)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get or create daily record", "error", err)
			return apperrors.Wrap(err, "failed to get or create daily record")
		}
		return validateDailyRecordRelation(result, clinicID, hospitalizationID, date)
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *dailyRecordService) AddVitalRecord(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateVitalRecordInput) (*model.DailyRecord, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	if err := validateDailyRecordIDs(clinicID, hospitalizationID); err != nil {
		return nil, err
	}
	if s.repo == nil || s.hospRepo == nil {
		return nil, apperrors.WrapInternalServerError("daily record persistence dependencies are required")
	}
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("daily record transaction dependency is required")
	}

	var vitalID uint64
	var dailyID uint64
	var result *model.DailyRecord
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		hosp, err := s.lockHospitalizationForWrite(txCtx, clinicID, hospitalizationID)
		if err != nil {
			return err
		}
		if err := validateClinicalStaffReference(
			txCtx,
			clinicID,
			input.StaffID,
			s.staffs,
			s.staffAssignments,
		); err != nil {
			return err
		}
		daily, err := s.repo.FindOrCreateByDate(txCtx, clinicID, hospitalizationID, date)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get or create daily record", "error", err)
			return apperrors.Wrap(err, "failed to get or create daily record")
		}
		if err := validateDailyRecordRelation(daily, clinicID, hospitalizationID, date); err != nil {
			return err
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
		result, err = s.repo.FindByHospitalizationIDAndDate(txCtx, clinicID, hospitalizationID, date)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get daily record after adding vital", "error", err)
			return apperrors.Wrap(err, "failed to get daily record after adding vital")
		}
		return s.validateDailyRecordRead(txCtx, clinicID, hosp, result, &date)
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "vital record added",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("daily_record_id", dailyID),
		slog.Uint64("vital_record_id", vitalID))

	return result, nil
}

func (s *dailyRecordService) AddCareLog(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateCareLogInput) (*model.DailyRecord, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	if err := validateDailyRecordIDs(clinicID, hospitalizationID); err != nil {
		return nil, err
	}
	if s.repo == nil || s.hospRepo == nil {
		return nil, apperrors.WrapInternalServerError("daily record persistence dependencies are required")
	}
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("daily record transaction dependency is required")
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

	var dailyID uint64
	var careLogID uint64
	var result *model.DailyRecord
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		hospitalization, err := s.lockHospitalizationForWrite(txCtx, clinicID, hospitalizationID)
		if err != nil {
			return err
		}
		if err := validateClinicalStaffReference(
			txCtx,
			clinicID,
			input.StaffID,
			s.staffs,
			s.staffAssignments,
		); err != nil {
			return err
		}
		daily, err := s.repo.FindOrCreateByDate(txCtx, clinicID, hospitalizationID, date)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get or create daily record", "error", err)
			return apperrors.Wrap(err, "failed to get or create daily record")
		}
		if err := validateDailyRecordRelation(daily, clinicID, hospitalizationID, date); err != nil {
			return err
		}
		dailyID = daily.ID
		cr := &model.CareLog{
			ClinicID:      clinicID,
			DailyRecordID: dailyID,
			Time:          input.Time,
			Type:          careLogType,
			Status:        status,
			Value:         input.Value,
			StaffID:       input.StaffID,
			Notes:         input.Notes,
		}
		if err := s.repo.CreateCareLog(txCtx, cr); err != nil {
			slog.ErrorContext(txCtx, "failed to create care log record", "error", err)
			return apperrors.Wrap(err, "failed to create care log record")
		}
		careLogID = cr.ID
		result, err = s.repo.FindByHospitalizationIDAndDate(txCtx, clinicID, hospitalizationID, date)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get daily record after adding care log", "error", err)
			return apperrors.Wrap(err, "failed to get daily record after adding care log")
		}
		return s.validateDailyRecordRead(txCtx, clinicID, hospitalization, result, &date)
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "care log record added",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("daily_record_id", dailyID),
		slog.Uint64("care_log_id", careLogID))

	return result, nil
}

func (s *dailyRecordService) AddStaffNote(ctx context.Context, clinicID, hospitalizationID uint64, date time.Time, input *CreateStaffNoteInput) (*model.DailyRecord, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	if err := validateDailyRecordIDs(clinicID, hospitalizationID); err != nil {
		return nil, err
	}
	if s.repo == nil || s.hospRepo == nil {
		return nil, apperrors.WrapInternalServerError("daily record persistence dependencies are required")
	}
	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("daily record transaction dependency is required")
	}

	var dailyID uint64
	var staffNoteID uint64
	var result *model.DailyRecord
	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		hospitalization, err := s.lockHospitalizationForWrite(txCtx, clinicID, hospitalizationID)
		if err != nil {
			return err
		}
		if err := validateClinicalStaffReference(
			txCtx,
			clinicID,
			input.StaffID,
			s.staffs,
			s.staffAssignments,
		); err != nil {
			return err
		}
		daily, err := s.repo.FindOrCreateByDate(txCtx, clinicID, hospitalizationID, date)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get or create daily record", "error", err)
			return apperrors.Wrap(err, "failed to get or create daily record")
		}
		if err := validateDailyRecordRelation(daily, clinicID, hospitalizationID, date); err != nil {
			return err
		}
		dailyID = daily.ID
		sn := &model.StaffNote{
			DailyRecordID: dailyID,
			Time:          input.Time,
			Content:       input.Content,
			StaffID:       input.StaffID,
		}
		if err := s.repo.CreateStaffNote(txCtx, sn); err != nil {
			slog.ErrorContext(txCtx, "failed to create staff note record", "error", err)
			return apperrors.Wrap(err, "failed to create staff note record")
		}
		staffNoteID = sn.ID
		result, err = s.repo.FindByHospitalizationIDAndDate(txCtx, clinicID, hospitalizationID, date)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get daily record after adding staff note", "error", err)
			return apperrors.Wrap(err, "failed to get daily record after adding staff note")
		}
		return s.validateDailyRecordRead(txCtx, clinicID, hospitalization, result, &date)
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "staff note record added",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("hospitalization_id", hospitalizationID),
		slog.Uint64("daily_record_id", dailyID),
		slog.Uint64("staff_note_id", staffNoteID))

	return result, nil
}

func validateDailyRecordIDs(clinicID, hospitalizationID uint64) error {
	if clinicID == 0 || hospitalizationID == 0 {
		return apperrors.WrapInvalidInput("clinic_id and hospitalization_id are required")
	}
	return nil
}

func validateDailyRecordHospitalization(
	hospitalization *model.Hospitalization,
	clinicID, hospitalizationID uint64,
) error {
	if hospitalization == nil ||
		hospitalization.ID != hospitalizationID ||
		hospitalization.ClinicID != clinicID ||
		hospitalization.PetID == 0 {
		return apperrors.WrapNotFound("hospitalization", "relation")
	}
	return nil
}

func (s *dailyRecordService) lockHospitalizationForWrite(
	ctx context.Context,
	clinicID, hospitalizationID uint64,
) (*model.Hospitalization, error) {
	if s.hospRepo == nil {
		return nil, apperrors.WrapInternalServerError("daily record hospitalization dependency is required")
	}
	hospitalization, err := s.hospRepo.LockByIDForUpdate(ctx, clinicID, hospitalizationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to lock hospitalization for daily record write", "error", err)
		return nil, apperrors.Wrap(err, "failed to verify hospitalization ownership")
	}
	if err := validateDailyRecordHospitalization(hospitalization, clinicID, hospitalizationID); err != nil {
		return nil, err
	}
	if err := s.validateHospitalizationOwnerPet(ctx, clinicID, hospitalization); err != nil {
		return nil, err
	}
	return hospitalization, nil
}

func (s *dailyRecordService) validateHospitalizationOwnerPet(
	ctx context.Context,
	clinicID uint64,
	hospitalization *model.Hospitalization,
) error {
	if s.ownerPets == nil {
		return apperrors.WrapInternalServerError(
			"daily record owner and pet validation dependency is required",
		)
	}
	ownerID := hospitalization.OwnerID
	petID := hospitalization.PetID
	if err := s.ownerPets.AssertOwnerInClinic(ctx, clinicID, ownerID); err != nil {
		return apperrors.Wrap(err, "failed to verify hospitalization owner")
	}
	if _, err := s.ownerPets.FindPetOwnerInClinic(ctx, clinicID, petID); err != nil {
		return apperrors.Wrap(err, "failed to verify hospitalization pet")
	}
	return nil
}

func (s *dailyRecordService) validateDailyRecordRead(
	ctx context.Context,
	clinicID uint64,
	hospitalization *model.Hospitalization,
	daily *model.DailyRecord,
	expectedDate *time.Time,
) error {
	if hospitalization == nil {
		return apperrors.WrapNotFound("hospitalization", "relation")
	}
	if daily == nil ||
		daily.ID == 0 ||
		daily.ClinicID != clinicID ||
		daily.HospitalizationID != hospitalization.ID {
		return apperrors.WrapNotFound("daily_record", "relation")
	}
	if expectedDate != nil &&
		daily.Date.Format(time.DateOnly) != expectedDate.Format(time.DateOnly) {
		return apperrors.WrapNotFound("daily_record", "relation")
	}

	for i := range daily.VitalRecords {
		vital := &daily.VitalRecords[i]
		if vital.ClinicID != clinicID ||
			vital.PetID != hospitalization.PetID ||
			vital.DailyRecordID == nil ||
			*vital.DailyRecordID != daily.ID {
			return apperrors.WrapNotFound("daily_record", "vital relation")
		}
		if err := s.validateNestedStaff(ctx, clinicID, vital.StaffID, vital.Staff); err != nil {
			return err
		}
	}
	for i := range daily.CareLogs {
		careLog := &daily.CareLogs[i]
		if careLog.ClinicID != clinicID || careLog.DailyRecordID != daily.ID {
			return apperrors.WrapNotFound("daily_record", "care log relation")
		}
		if err := s.validateNestedStaff(ctx, clinicID, careLog.StaffID, careLog.Staff); err != nil {
			return err
		}
	}
	for i := range daily.StaffNotes {
		note := &daily.StaffNotes[i]
		if note.DailyRecordID != daily.ID {
			return apperrors.WrapNotFound("daily_record", "staff note relation")
		}
		if err := s.validateNestedStaff(ctx, clinicID, note.StaffID, note.Staff); err != nil {
			return err
		}
	}
	return nil
}

func (s *dailyRecordService) validateNestedStaff(
	ctx context.Context,
	clinicID uint64,
	staffID *uint64,
	staff *model.Staff,
) error {
	if staffID == nil {
		if staff != nil {
			return apperrors.WrapNotFound("staff", "nested relation")
		}
		return nil
	}
	if staff != nil && (staff.ID != *staffID || !staff.IsActive) {
		return apperrors.WrapNotFound("staff", "nested relation")
	}
	return validateClinicalStaffReference(
		ctx,
		clinicID,
		staffID,
		s.staffs,
		s.staffAssignments,
	)
}

func validateDailyRecordRelation(
	daily *model.DailyRecord,
	clinicID, hospitalizationID uint64,
	date time.Time,
) error {
	if daily == nil ||
		daily.ID == 0 ||
		daily.ClinicID != clinicID ||
		daily.HospitalizationID != hospitalizationID ||
		daily.Date.Format(time.DateOnly) != date.Format(time.DateOnly) {
		return apperrors.WrapNotFound("daily_record", "relation")
	}
	return nil
}
