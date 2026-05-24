package service

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository"
)

const maxReasonLength = 500

type CreateMedicalRecordAddendumInput struct {
	MedicalRecordID uint64
	AuthorUserID    uint64
	AfterText       string
	Reason          string
}

type MedicalRecordAddendumService interface {
	Create(ctx context.Context, clinicID uint64, input CreateMedicalRecordAddendumInput) (*model.MedicalRecordAddendum, error)
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error)
}

type medicalRecordAddendumService struct {
	repo          repository.MedicalRecordAddendumRepository
	medicalRecord repository.MedicalRecordRepository
	auditService  AuditService
}

func NewMedicalRecordAddendumService(
	repo repository.MedicalRecordAddendumRepository,
	medicalRecord repository.MedicalRecordRepository,
	auditService AuditService,
) MedicalRecordAddendumService {
	return &medicalRecordAddendumService{
		repo:          repo,
		medicalRecord: medicalRecord,
		auditService:  auditService,
	}
}

func (s *medicalRecordAddendumService) Create(ctx context.Context, clinicID uint64, input CreateMedicalRecordAddendumInput) (*model.MedicalRecordAddendum, error) {
	// バリデーション: AfterText 空文字禁止
	if input.AfterText == "" {
		return nil, apperrors.WrapInvalidInput("after_text is required")
	}

	// バリデーション: Reason 空文字禁止
	if input.Reason == "" {
		return nil, apperrors.WrapInvalidInput("reason is required")
	}

	// バリデーション: Reason 最大長
	if len([]rune(input.Reason)) > maxReasonLength {
		return nil, apperrors.WrapInvalidInput(fmt.Sprintf("reason must be %d characters or fewer", maxReasonLength))
	}

	// カルテ存在確認 + clinic_id 所属確認
	record, err := s.medicalRecord.FindByID(ctx, clinicID, input.MedicalRecordID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to find medical record")
	}

	// 確定済みカルテにのみ追記可能
	if record.Status != model.MedicalRecordStatusFinalized {
		return nil, apperrors.WrapInvalidInput("addendum can only be added to finalized medical records")
	}

	addendum := &model.MedicalRecordAddendum{
		MedicalRecordID: input.MedicalRecordID,
		ClinicID:        clinicID,
		AuthorUserID:    input.AuthorUserID,
		BeforeText:      "",
		AfterText:       input.AfterText,
		Reason:          input.Reason,
	}

	if err := s.repo.Create(ctx, addendum); err != nil {
		slog.ErrorContext(ctx, "failed to create medical record addendum", "error", err)
		return nil, apperrors.Wrap(err, "failed to create medical record addendum")
	}

	// 監査ログ: create（best-effort）
	if s.auditService != nil {
		authorID := input.AuthorUserID
		if err := s.auditService.LogAddendumCreate(ctx, clinicID, &authorID, addendum.ID, input.MedicalRecordID, addendum); err != nil {
			slog.ErrorContext(ctx, "audit log failed for addendum create", "error", err, "addendum_id", addendum.ID)
		}
	}

	return addendum, nil
}

func (s *medicalRecordAddendumService) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]*model.MedicalRecordAddendum, error) {
	// カルテ存在確認 + clinic_id 所属確認
	if _, err := s.medicalRecord.FindByID(ctx, clinicID, medicalRecordID); err != nil {
		return nil, apperrors.Wrap(err, "failed to find medical record")
	}

	addenda, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find medical record addenda", "error", err)
		return nil, apperrors.Wrap(err, "failed to find medical record addenda")
	}

	return addenda, nil
}
