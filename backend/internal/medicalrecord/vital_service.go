package medicalrecord

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CreateVitalInput はバイタル作成の入力DTO（HTTP非依存）
type CreateVitalInput struct {
	ClinicID        uint64 // 監査ログ用
	PetID           uint64
	RecordedAt      time.Time
	StaffID         *uint64
	Temperature     *float64
	HeartRate       *int
	RespirationRate *int
	Weight          *float64
	WeightUnit      *model.BodyWeightUnit
	Notes           string
}

// UpdateVitalInput はバイタル更新の入力DTO（nil = 未送信フィールド）
type UpdateVitalInput struct {
	RecordedAt      *time.Time
	StaffID         *uint64
	Temperature     *float64
	HeartRate       *int
	RespirationRate *int
	Weight          *float64
	WeightUnit      *model.BodyWeightUnit
	Notes           *string
	ActorID         *uint64 // 監査ログ用: 操作スタッフ ID（nil = システム）
}

// VitalService はバイタル記録のビジネスロジックインターフェース
type VitalService interface {
	List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error)
	Create(ctx context.Context, medicalRecordID uint64, input *CreateVitalInput) (*model.VitalRecord, error)
	Update(ctx context.Context, clinicID, medicalRecordID, vitalID uint64, input *UpdateVitalInput) (*model.VitalRecord, error)
	Delete(ctx context.Context, clinicID, medicalRecordID, vitalID uint64) error
}

type vitalService struct {
	repo              VitalRepository
	medicalRecordRepo medicalRecordLocker
	auditTx           AuditTxLogger
	relations         ClinicalRelationVerifier
	staffs            clinicalStaffLocker
	staffAssignments  clinicalStaffAssignmentLocker
	transactor        Transactor
}

// NewVitalService はVitalServiceを初期化して返す。transactor は BE-refactor.md X-11
// （確定と子書込の競合防止）のため、子書込を LockByIDForUpdate の行ロックと同一トランザクションに
// 収める目的で注入する。auditTx は BUG-015 の同一 tx fail-closed 監査用。
func NewVitalService(repo VitalRepository, medicalRecordRepo medicalRecordLocker, auditTx AuditTxLogger, transactor Transactor) VitalService {
	return &vitalService{repo: repo, medicalRecordRepo: medicalRecordRepo, auditTx: auditTx, transactor: transactor}
}

// NewVitalServiceWithRelationValidation は request 由来 staff_id の active staff + clinic
// assignment 検証依存を注入する。検証と保存は transactor の同一トランザクションで実行する。
func NewVitalServiceWithRelationValidation(
	repo VitalRepository,
	medicalRecordRepo medicalRecordLocker,
	auditTx AuditTxLogger,
	relations ClinicalRelationVerifier,
	staffs clinicalStaffLocker,
	staffAssignments clinicalStaffAssignmentLocker,
	transactor Transactor,
) VitalService {
	return &vitalService{
		repo:              repo,
		medicalRecordRepo: medicalRecordRepo,
		auditTx:           auditTx,
		relations:         relations,
		staffs:            staffs,
		staffAssignments:  staffAssignments,
		transactor:        transactor,
	}
}

func (s *vitalService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.VitalRecord, error) {
	items, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list vital records")
	}
	return items, nil
}

func (s *vitalService) Create(ctx context.Context, medicalRecordID uint64, input *CreateVitalInput) (*model.VitalRecord, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	if input.ClinicID == 0 || medicalRecordID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_id and medical_record_id are required")
	}
	if input.PetID == 0 {
		return nil, apperrors.WrapInvalidInput("pet_id is required")
	}
	if input.Temperature == nil && input.HeartRate == nil && input.RespirationRate == nil && input.Weight == nil {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	if err := validateVitalWeight(input.Weight, input.WeightUnit); err != nil {
		return nil, err
	}
	if s.repo == nil || s.medicalRecordRepo == nil {
		return nil, apperrors.WrapInternalServerError("vital persistence dependencies are required")
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("vital transaction dependency is required")
	}
	if s.auditTx == nil {
		return nil, apperrors.WrapInternalServerError("vital audit dependency is required")
	}

	vital := &model.VitalRecord{
		ClinicID:        input.ClinicID,
		PetID:           input.PetID,
		MedicalRecordID: &medicalRecordID,
		RecordedAt:      input.RecordedAt,
		StaffID:         input.StaffID,
		Temperature:     input.Temperature,
		HeartRate:       input.HeartRate,
		RespirationRate: input.RespirationRate,
		Weight:          input.Weight,
		WeightUnit:      weightUnitOrDefault(input.WeightUnit),
		Notes:           input.Notes,
	}

	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		return s.createVitalInTx(txCtx, medicalRecordID, input, vital)
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "vital created",
		slog.Uint64("vital_id", vital.ID),
		slog.Uint64("medical_record_id", medicalRecordID))

	return vital, nil
}

func (s *vitalService) Update(ctx context.Context, clinicID, medicalRecordID, vitalID uint64, input *UpdateVitalInput) (*model.VitalRecord, error) {
	if input == nil {
		return nil, apperrors.WrapInvalidInput("input must not be nil")
	}
	fields := buildVitalUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if clinicID == 0 || medicalRecordID == 0 || vitalID == 0 {
		return nil, apperrors.WrapInvalidInput("clinic_id, medical_record_id, and vital_id are required")
	}
	if err := validateVitalWeight(input.Weight, input.WeightUnit); err != nil {
		return nil, err
	}
	if s.repo == nil || s.medicalRecordRepo == nil {
		return nil, apperrors.WrapInternalServerError("vital persistence dependencies are required")
	}
	if s.transactor == nil {
		return nil, apperrors.WrapInternalServerError("vital transaction dependency is required")
	}
	if s.auditTx == nil {
		return nil, apperrors.WrapInternalServerError("vital audit dependency is required")
	}
	// 所属確認: このvitalIDがclinicID・medicalRecordIDに属しているか検証
	existing, err := s.repo.FindByID(ctx, clinicID, vitalID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get vital record")
	}
	if existing == nil || existing.MedicalRecordID == nil || *existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("vital", "not found in medical record")
	}

	var result *model.VitalRecord
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		updated, err := s.updateVitalInTx(txCtx, clinicID, medicalRecordID, vitalID, input, fields, existing)
		if err != nil {
			return err
		}
		result = updated
		return nil
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "vital updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("vital_id", vitalID),
		slog.Uint64("medical_record_id", medicalRecordID))

	return result, nil
}

func (s *vitalService) lockDraftParent(
	ctx context.Context,
	clinicID, medicalRecordID uint64,
	findErrMsg, conflictMsg string,
) (*model.MedicalRecord, error) {
	if s.medicalRecordRepo == nil {
		return nil, apperrors.WrapInternalServerError("vital medical record validation dependency is required")
	}
	parent, err := s.medicalRecordRepo.LockByIDForUpdate(ctx, clinicID, medicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, findErrMsg)
	}
	if parent == nil ||
		parent.ID != medicalRecordID ||
		parent.ClinicID != clinicID {
		return nil, apperrors.WrapNotFound("medical_record", "relation")
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return nil, apperrors.WrapConflict(conflictMsg)
	}
	return parent, nil
}

func (s *vitalService) Delete(ctx context.Context, clinicID, medicalRecordID, vitalID uint64) error {
	if s.repo == nil || s.medicalRecordRepo == nil {
		return apperrors.WrapInternalServerError("vital persistence dependencies are required")
	}
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("vital transaction dependency is required")
	}
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("vital audit dependency is required")
	}
	// 所属確認: このvitalIDがclinicID・medicalRecordIDに属しているか検証
	existing, err := s.repo.FindByID(ctx, clinicID, vitalID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get vital record")
	}
	if existing == nil || existing.MedicalRecordID == nil || *existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("vital", "not found in medical record")
	}

	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// BE-refactor.md X-11: LockByIDForUpdate の行ロックで finalize と直列化し、確定と同時の
		// バイタル削除が確定済みカルテに混入する競合を防ぐ。
		if err := lockDraftMedicalRecord(txCtx, s.medicalRecordRepo, clinicID, medicalRecordID,
			"failed to find medical record", "確定済みカルテのバイタルは削除できません"); err != nil {
			return err
		}
		// pre-delete 値は delete 前に確定し、同一 ambient tx で fail-closed audit に渡す。
		oldValue := extractVitalImportantFields(existing)
		if err := s.repo.Delete(txCtx, clinicID, vitalID); err != nil {
			return apperrors.Wrap(err, "failed to delete vital record")
		}
		// BUG-015: vital delete audit は ambient tx 参加の LogEntryTx で fail-closed。
		if err := s.auditVitalTx(txCtx, clinicID, nil, "delete", vitalID, medicalRecordID, oldValue, nil); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	slog.InfoContext(ctx, "vital deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("vital_id", vitalID),
		slog.Uint64("medical_record_id", medicalRecordID))

	return nil
}
