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
}

// VitalService はバイタル記録のビジネスロジックインターフェース
type VitalService interface {
	List(ctx context.Context, medicalRecordID uint64) ([]model.VitalRecord, error)
	Create(ctx context.Context, medicalRecordID uint64, input *CreateVitalInput) (*model.VitalRecord, error)
	Update(ctx context.Context, clinicID, medicalRecordID, vitalID uint64, input *UpdateVitalInput) (*model.VitalRecord, error)
	Delete(ctx context.Context, clinicID, medicalRecordID, vitalID uint64) error
}

type vitalService struct {
	repo repository.VitalRepository
}

// NewVitalService はVitalServiceを初期化して返す
func NewVitalService(repo repository.VitalRepository) VitalService {
	return &vitalService{repo: repo}
}

func (s *vitalService) List(ctx context.Context, medicalRecordID uint64) ([]model.VitalRecord, error) {
	return s.repo.ListByMedicalRecordID(ctx, medicalRecordID)
}

func (s *vitalService) Create(ctx context.Context, medicalRecordID uint64, input *CreateVitalInput) (*model.VitalRecord, error) {
	if input.PetID == 0 {
		return nil, apperrors.WrapInvalidInput("pet_id is required")
	}
	if input.Temperature == nil && input.HeartRate == nil && input.RespirationRate == nil && input.Weight == nil {
		return nil, apperrors.WrapInvalidInput("少なくとも1つのバイタル値を入力してください")
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
		return nil, apperrors.Wrap(err, "failed to create vital record")
	}
	slog.InfoContext(ctx, "vital created",
		slog.Uint64("vital_id", vital.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return vital, nil
}

func (s *vitalService) Update(ctx context.Context, clinicID, medicalRecordID, vitalID uint64, input *UpdateVitalInput) (*model.VitalRecord, error) {
	// 所属確認: このvitalIDがclinicID・medicalRecordIDに属しているか検証
	existing, err := s.repo.FindByID(ctx, clinicID, vitalID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get vital record")
	}
	if existing.MedicalRecordID == nil || *existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("vital", "not found in medical record")
	}

	fields := buildVitalUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, vitalID, fields); err != nil {
		return nil, apperrors.Wrap(err, "failed to update vital record")
	}
	slog.InfoContext(ctx, "vital updated",
		slog.Uint64("vital_id", vitalID),
		slog.Uint64("medical_record_id", medicalRecordID))
	result, err := s.repo.FindByID(ctx, clinicID, vitalID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get vital record after update")
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
	if err := s.repo.Delete(ctx, vitalID); err != nil {
		return apperrors.Wrap(err, "failed to delete vital record")
	}
	slog.InfoContext(ctx, "vital deleted",
		slog.Uint64("vital_id", vitalID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return nil
}

// weightUnitOrDefault は nil の場合に BodyWeightUnitKg を返すヘルパー。
func weightUnitOrDefault(u *model.BodyWeightUnit) model.BodyWeightUnit {
	if u != nil {
		return *u
	}
	return model.BodyWeightUnitKg
}

// buildVitalUpdateFields はnilでないフィールドのみmap[string]anyに変換する
func buildVitalUpdateFields(input *UpdateVitalInput) map[string]any {
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
