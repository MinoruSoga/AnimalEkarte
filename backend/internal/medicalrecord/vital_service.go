package medicalrecord

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
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

// buildVitalUpdate はnilでないフィールドのみmap[string]anyに変換する
func buildVitalUpdate(input *UpdateVitalInput) map[string]any {
	fields := map[string]any{}
	if input.RecordedAt != nil {
		fields["recorded_at"] = *input.RecordedAt
	}
	if input.StaffID != nil {
		fields["staff_id"] = *input.StaffID
	}
	if input.Temperature != nil {
		fields["temperature"] = *input.Temperature
	}
	if input.HeartRate != nil {
		fields["heart_rate"] = *input.HeartRate
	}
	if input.RespirationRate != nil {
		fields["respiration_rate"] = *input.RespirationRate
	}
	if input.Weight != nil {
		fields["weight"] = *input.Weight
	}
	if input.WeightUnit != nil {
		fields["weight_unit"] = *input.WeightUnit
	}
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	return fields
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
		slog.ErrorContext(ctx, "failed to list vital records", "error", err)
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
		// HC-006 + BE-refactor.md X-11: 親カルテが確定済みの場合は作成拒否。LockByIDForUpdate の
		// 行ロックで finalize と直列化し、確定と同時のバイタル追加が確定済みカルテに混入する競合を防ぐ。
		parent, err := s.lockDraftParent(
			txCtx,
			input.ClinicID,
			medicalRecordID,
			"failed to find medical record",
			"確定済みカルテにバイタルを追加できません",
		)
		if err != nil {
			return err
		}
		if err := validateVitalMedicalRecordRelation(parent, input.ClinicID, medicalRecordID, input.PetID); err != nil {
			return err
		}
		petID := input.PetID
		if err := validateClinicalRelations(
			txCtx,
			s.relations,
			input.ClinicID,
			parent,
			&petID,
			nil,
		); err != nil {
			return err
		}
		if err := validateClinicalStaffReference(
			txCtx,
			input.ClinicID,
			input.StaffID,
			s.staffs,
			s.staffAssignments,
		); err != nil {
			return err
		}
		if err := s.repo.Create(txCtx, vital); err != nil {
			slog.ErrorContext(txCtx, "failed to create vital record", "error", err)
			return apperrors.Wrap(err, "failed to create vital record")
		}
		// BUG-015: vital create audit は ambient tx 参加の LogEntryTx で fail-closed。
		var actorID *uint64
		if input.StaffID != nil {
			actorID = input.StaffID
		}
		if err := s.auditVitalTx(txCtx, input.ClinicID, actorID, "create", vital.ID, medicalRecordID, nil, extractVitalImportantFields(vital)); err != nil {
			return err
		}
		return nil
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
		slog.ErrorContext(ctx, "failed to get vital record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get vital record")
	}
	if existing == nil || existing.MedicalRecordID == nil || *existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("vital", "not found in medical record")
	}

	var result *model.VitalRecord
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		// BE-refactor.md X-11: LockByIDForUpdate の行ロックで finalize と直列化し、確定と同時の
		// バイタル編集が確定済みカルテに混入する競合を防ぐ。
		parent, err := s.lockDraftParent(
			txCtx,
			clinicID,
			medicalRecordID,
			"failed to find medical record",
			"確定済みカルテのバイタルは編集できません",
		)
		if err != nil {
			return err
		}
		if err := validateVitalMedicalRecordRelation(parent, clinicID, medicalRecordID, existing.PetID); err != nil {
			return err
		}
		petID := existing.PetID
		if err := validateClinicalRelations(
			txCtx,
			s.relations,
			clinicID,
			parent,
			&petID,
			nil,
		); err != nil {
			return err
		}
		if existing.ClinicID != clinicID {
			return apperrors.WrapNotFound("vital", "not found in medical record")
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
		if err := s.repo.Update(txCtx, clinicID, vitalID, fields); err != nil {
			slog.ErrorContext(txCtx, "failed to update vital record", "error", err)
			return apperrors.Wrap(err, "failed to update vital record")
		}
		result, err = s.repo.FindByID(txCtx, clinicID, vitalID)
		if err != nil {
			slog.ErrorContext(txCtx, "failed to get vital record after update", "error", err)
			return apperrors.Wrap(err, "failed to get vital record after update")
		}
		if err := validateUpdatedVitalRelation(
			result,
			clinicID,
			medicalRecordID,
			vitalID,
			existing.PetID,
		); err != nil {
			return err
		}
		if err := validateClinicalStaffReference(
			txCtx,
			clinicID,
			result.StaffID,
			s.staffs,
			s.staffAssignments,
		); err != nil {
			return err
		}
		if result.Staff != nil &&
			(result.StaffID == nil ||
				result.Staff.ID != *result.StaffID ||
				!result.Staff.IsActive) {
			return apperrors.WrapNotFound("staff", "nested relation")
		}
		// BUG-015: vital update audit は ambient tx 参加の LogEntryTx で fail-closed。
		oldDiff, newDiff := diffVitalImportantFields(existing, result)
		if oldDiff != nil {
			if err := s.auditVitalTx(txCtx, clinicID, input.ActorID, "update", vitalID, medicalRecordID, oldDiff, newDiff); err != nil {
				return err
			}
		}
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

func validateUpdatedVitalRelation(
	vital *model.VitalRecord,
	clinicID, medicalRecordID, vitalID, petID uint64,
) error {
	if vital == nil ||
		vital.ID != vitalID ||
		vital.ClinicID != clinicID ||
		vital.PetID != petID ||
		vital.MedicalRecordID == nil ||
		*vital.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("vital", "not found in medical record")
	}
	return nil
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
		slog.ErrorContext(ctx, findErrMsg, "error", err)
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

func validateVitalMedicalRecordRelation(
	parent *model.MedicalRecord,
	clinicID, medicalRecordID, petID uint64,
) error {
	if parent == nil ||
		parent.ID != medicalRecordID ||
		parent.ClinicID != clinicID ||
		parent.PetID == nil ||
		*parent.PetID != petID {
		return apperrors.WrapNotFound("medical_record", "relation")
	}
	return nil
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
			slog.ErrorContext(txCtx, "failed to delete vital record", "error", err, "clinic_id", clinicID, "vital_id", vitalID)
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

// auditVitalTx は vital create/update/delete の監査を ambient transaction へ参加させる（BUG-015）。
// LogVitalChange（base conn Create）は使わず、LogEntryTx → CreateTx と同型にする。
func (s *vitalService) auditVitalTx(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action string,
	vitalID, medicalRecordID uint64,
	oldValue, newValue map[string]any,
) error {
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("vital audit dependency is required")
	}
	resourceID := vitalID
	entry := &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  sharedkernel.AuditActorTypeFor(actorID),
		Action:     action,
		Resource:   "vital",
		ResourceID: &resourceID,
		OldValue:   oldValue,
		NewValue:   newValue,
		Metadata: map[string]any{
			"medical_record_id": medicalRecordID,
		},
	}
	if err := s.auditTx.LogEntryTx(ctx, entry); err != nil {
		slog.ErrorContext(ctx, "failed to write vital audit",
			"error", err,
			"action", action,
			"vital_id", vitalID,
			"medical_record_id", medicalRecordID,
		)
		return apperrors.Wrap(err, "failed to write vital "+action+" audit")
	}
	return nil
}

// validateVitalWeight は service 層での weight 構造検証（request 境界と二重）。
func validateVitalWeight(weight *float64, unit *model.BodyWeightUnit) error {
	if weight != nil {
		if math.IsNaN(*weight) || math.IsInf(*weight, 0) {
			return apperrors.WrapInvalidInput("weight must be a finite number")
		}
		if *weight <= 0 {
			return apperrors.WrapInvalidInput("weight must be greater than zero")
		}
	}
	if unit != nil {
		switch *unit {
		case model.BodyWeightUnitKg, model.BodyWeightUnitG:
			// ok
		default:
			return apperrors.WrapInvalidInput("weight_unit must be Kg or g")
		}
	}
	return nil
}

// weightUnitOrDefault は nil の場合に BodyWeightUnitKg を返すヘルパー。
func weightUnitOrDefault(u *model.BodyWeightUnit) model.BodyWeightUnit {
	if u != nil {
		return *u
	}
	return model.BodyWeightUnitKg
}
