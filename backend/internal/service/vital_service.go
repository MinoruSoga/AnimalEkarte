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
	RecordedAt      time.Time
	StaffID         *uint64
	Temperature     *float64
	HeartRate       *int
	RespirationRate *int
	Weight          *float64
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
	Notes           *string
}

// VitalService はバイタル記録のビジネスロジックインターフェース
type VitalService interface {
	List(ctx context.Context, medicalRecordID uint64) ([]model.VitalRecord, error)
	Create(ctx context.Context, medicalRecordID uint64, input *CreateVitalInput) (*model.VitalRecord, error)
	Update(ctx context.Context, medicalRecordID, vitalID uint64, input *UpdateVitalInput) (*model.VitalRecord, error)
	Delete(ctx context.Context, medicalRecordID, vitalID uint64) error
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
	vital := &model.VitalRecord{
		MedicalRecordID: &medicalRecordID,
		RecordedAt:      input.RecordedAt,
		StaffID:         input.StaffID,
		Temperature:     input.Temperature,
		HeartRate:       input.HeartRate,
		RespirationRate: input.RespirationRate,
		Weight:          input.Weight,
		Notes:           input.Notes,
	}
	if err := s.repo.Create(ctx, vital); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "vital created",
		slog.Uint64("vital_id", vital.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return vital, nil
}

func (s *vitalService) Update(ctx context.Context, medicalRecordID, vitalID uint64, input *UpdateVitalInput) (*model.VitalRecord, error) {
	// 所属確認: このvitalIDがmedicalRecordIDに属しているか検証
	existing, err := s.repo.FindByID(ctx, vitalID)
	if err != nil {
		return nil, err
	}
	if existing.MedicalRecordID == nil || *existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("vital", "not found in medical record")
	}

	fields := buildVitalUpdateFields(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, vitalID, fields); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "vital updated",
		slog.Uint64("vital_id", vitalID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return s.repo.FindByID(ctx, vitalID)
}

func (s *vitalService) Delete(ctx context.Context, medicalRecordID, vitalID uint64) error {
	// 所属確認: このvitalIDがmedicalRecordIDに属しているか検証
	existing, err := s.repo.FindByID(ctx, vitalID)
	if err != nil {
		return err
	}
	if existing.MedicalRecordID == nil || *existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("vital", "not found in medical record")
	}
	if err := s.repo.Delete(ctx, vitalID); err != nil {
		return err
	}
	slog.InfoContext(ctx, "vital deleted",
		slog.Uint64("vital_id", vitalID),
		slog.Uint64("medical_record_id", medicalRecordID))
	return nil
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
	if input.Notes != nil {
		fields["notes"] = *input.Notes
	}
	return fields
}
