package medicalrecord

import (
	"context"
	"fmt"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

const maxReasonLength = 500

// consultationEventKindEnded は medical_record_consultation_events.event_kind の終了値（001 CHECK: ended / end_undone）。
const consultationEventKindEnded = "ended"

// addendumRequiresEndedConsultationMessage は診療終了の記録がないカルテへの訂正追記を拒否する固定文言（EMR-495 Q2）。
const addendumRequiresEndedConsultationMessage = "訂正追記は診療終了が記録されたカルテにのみ追加できます。終了の記録は予約の終了操作（POST /api/v1/reservations/{id}/end-consultation）で行います。"

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

type medicalRecordAddendumParent interface {
	medicalRecordFinder
	medicalRecordLocker
}

type medicalRecordAddendumService struct {
	repo          MedicalRecordAddendumRepository
	medicalRecord medicalRecordAddendumParent
	auditTx       AuditTxLogger
	tx            Transactor
}

func NewMedicalRecordAddendumService(
	repo MedicalRecordAddendumRepository,
	medicalRecord medicalRecordAddendumParent,
	auditTx AuditTxLogger,
	tx Transactor,
) MedicalRecordAddendumService {
	return &medicalRecordAddendumService{
		repo:          repo,
		medicalRecord: medicalRecord,
		auditTx:       auditTx,
		tx:            tx,
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

	if s.tx == nil {
		return nil, apperrors.WrapInternalServerError("medical record addendum write requires a transaction dependency")
	}
	if s.auditTx == nil {
		return nil, apperrors.WrapInternalServerError("medical record addendum audit dependency is required")
	}

	addendum := &model.MedicalRecordAddendum{
		MedicalRecordID: input.MedicalRecordID,
		ClinicID:        clinicID,
		AuthorUserID:    input.AuthorUserID,
		BeforeText:      "",
		AfterText:       input.AfterText,
		Reason:          input.Reason,
	}

	if err := s.tx.WithTx(ctx, func(txCtx context.Context) error {
		// clinic-scoped FOR UPDATE lock serializes parent validation with the addendum insert.
		record, err := s.medicalRecord.LockByIDForUpdate(txCtx, clinicID, input.MedicalRecordID)
		if err != nil {
			return apperrors.Wrap(err, "failed to lock medical record for addendum")
		}
		if record == nil {
			return apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", input.MedicalRecordID))
		}
		// EMR-495 Q2: 旧 status=finalized の前提を、最新の診療終了イベントが ended であることに置き換える。
		// 同じ tx・親カルテ行ロックの下で読み、イベントを捏造しない（履歴なし・end_undone は不許可）。
		kind, found, err := s.repo.LatestConsultationEventKind(txCtx, clinicID, input.MedicalRecordID)
		if err != nil {
			return apperrors.Wrap(err, "failed to read consultation end history")
		}
		if !found || kind != consultationEventKindEnded {
			return apperrors.WrapMedicalRecordStatusRetired(addendumRequiresEndedConsultationMessage)
		}
		if err := s.repo.Create(txCtx, addendum); err != nil {
			return apperrors.Wrap(err, "failed to create medical record addendum")
		}
		authorID := input.AuthorUserID
		resourceID := addendum.ID
		if err := s.auditTx.LogEntryTx(txCtx, &AuditEntry{
			ClinicID:   &clinicID,
			ActorID:    &authorID,
			ActorType:  model.AuditActorTypeStaff,
			Action:     "create",
			Resource:   "medical_record_addendum",
			ResourceID: &resourceID,
			NewValue: map[string]any{
				"before_text":    addendum.BeforeText,
				"after_text":     addendum.AfterText,
				"reason":         addendum.Reason,
				"author_user_id": addendum.AuthorUserID,
			},
			Metadata: map[string]any{"medical_record_id": input.MedicalRecordID},
		}); err != nil {
			return apperrors.Wrap(err, "failed to audit medical record addendum create")
		}
		return nil
	}); err != nil {
		return nil, err
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
		return nil, apperrors.Wrap(err, "failed to find medical record addenda")
	}

	return addenda, nil
}
