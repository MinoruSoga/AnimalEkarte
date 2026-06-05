package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
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
	repo              repository.VitalRepository
	medicalRecordRepo repository.MedicalRecordRepository
	auditService      AuditService
}

// NewVitalService はVitalServiceを初期化して返す
func NewVitalService(repo repository.VitalRepository, medicalRecordRepo repository.MedicalRecordRepository, auditService AuditService) VitalService {
	return &vitalService{repo: repo, medicalRecordRepo: medicalRecordRepo, auditService: auditService}
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
	if input.PetID == 0 {
		return nil, apperrors.WrapInvalidInput("pet_id is required")
	}
	if input.Temperature == nil && input.HeartRate == nil && input.RespirationRate == nil && input.Weight == nil {
		return nil, apperrors.WrapInvalidInput(ErrMsgAtLeastOneField)
	}

	// HC-006: 親カルテが確定済みの場合は作成拒否
	parent, err := s.medicalRecordRepo.FindByID(ctx, input.ClinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to find medical record")
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return nil, apperrors.WrapConflict("確定済みカルテにバイタルを追加できません")
	}

	vital := &model.VitalRecord{
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
	if err := s.repo.Create(ctx, vital); err != nil {
		slog.ErrorContext(ctx, "failed to create vital record", "error", err)
		return nil, apperrors.Wrap(err, "failed to create vital record")
	}
	slog.InfoContext(ctx, "vital created",
		slog.Uint64("vital_id", vital.ID),
		slog.Uint64("medical_record_id", medicalRecordID))

	// 監査ログ: create（best-effort）
	if s.auditService != nil && input.ClinicID != 0 {
		var actorID *uint64
		if input.StaffID != nil {
			actorID = input.StaffID
		}
		newValue := extractVitalImportantFields(vital)
		if err := s.auditService.LogVitalChange(ctx, input.ClinicID, actorID, "create", vital.ID, medicalRecordID, nil, newValue); err != nil {
			slog.ErrorContext(ctx, "audit log failed for vital create", "error", err, "vital_id", vital.ID)
		}
	}

	return vital, nil
}

func (s *vitalService) Update(ctx context.Context, clinicID, medicalRecordID, vitalID uint64, input *UpdateVitalInput) (*model.VitalRecord, error) {
	// 所属確認: このvitalIDがclinicID・medicalRecordIDに属しているか検証
	existing, err := s.repo.FindByID(ctx, clinicID, vitalID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get vital record", "error", err)
		return nil, apperrors.Wrap(err, "failed to get vital record")
	}
	if existing.MedicalRecordID == nil || *existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("vital", "not found in medical record")
	}
	parent, err := s.medicalRecordRepo.FindByID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to find medical record")
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return nil, apperrors.WrapConflict("確定済みカルテのバイタルは編集できません")
	}

	fields := buildVitalUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, vitalID, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update vital record", "error", err)
		return nil, apperrors.Wrap(err, "failed to update vital record")
	}
	slog.InfoContext(ctx, "vital updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("vital_id", vitalID),
		slog.Uint64("medical_record_id", medicalRecordID))
	result, err := s.repo.FindByID(ctx, clinicID, vitalID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get vital record after update", "error", err)
		return nil, apperrors.Wrap(err, "failed to get vital record after update")
	}

	// 監査ログ: update（best-effort）
	if s.auditService != nil {
		oldDiff, newDiff := diffVitalImportantFields(existing, result)
		if oldDiff != nil {
			if err := s.auditService.LogVitalChange(ctx, clinicID, input.ActorID, "update", vitalID, medicalRecordID, oldDiff, newDiff); err != nil {
				slog.ErrorContext(ctx, "audit log failed for vital update", "error", err, "vital_id", vitalID)
			}
		}
	}

	return result, nil
}

func (s *vitalService) Delete(ctx context.Context, clinicID, medicalRecordID, vitalID uint64) error {
	// 所属確認: このvitalIDがclinicID・medicalRecordIDに属しているか検証
	existing, err := s.repo.FindByID(ctx, clinicID, vitalID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get vital record")
	}
	if existing.MedicalRecordID == nil || *existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("vital", "not found in medical record")
	}
	parent, err := s.medicalRecordRepo.FindByID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find medical record", "error", err)
		return apperrors.Wrap(err, "failed to find medical record")
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return apperrors.WrapConflict("確定済みカルテのバイタルは削除できません")
	}
	oldValue := extractVitalImportantFields(existing)
	if err := s.repo.Delete(ctx, clinicID, vitalID); err != nil {
		return apperrors.Wrap(err, "failed to delete vital record")
	}
	slog.InfoContext(ctx, "vital deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("vital_id", vitalID),
		slog.Uint64("medical_record_id", medicalRecordID))

	// 監査ログ: delete（best-effort, actorID=nil）
	if s.auditService != nil {
		if err := s.auditService.LogVitalChange(ctx, clinicID, nil, "delete", vitalID, medicalRecordID, oldValue, nil); err != nil {
			slog.ErrorContext(ctx, "audit log failed for vital delete", "error", err, "vital_id", vitalID)
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
