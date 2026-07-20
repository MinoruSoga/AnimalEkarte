package medicalrecord

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/repository/repohelpers"
)

// CreateCheckupInput は健診記録作成の入力DTO
type CreateCheckupInput struct {
	ClinicID      uint64
	CheckupTypeID uint64
	PetID         *uint64
	Date          time.Time
	NextDate      *time.Time
	DoctorID      *uint64
	Result        string
}

// UpdateCheckupInput は健診記録更新の入力DTO
type UpdateCheckupInput struct {
	CheckupTypeID *uint64
	PetID         *uint64
	Date          *time.Time
	NextDate      *time.Time
	DoctorID      *uint64
	DoctorIDClear *bool // true のとき doctor_id を NULL にクリアする
	Result        *string
}

// ListCheckupsByClinicInput はクリニック横断一覧取得の入力DTO
type ListCheckupsByClinicInput struct {
	ClinicID      uint64
	StartDate     *string
	EndDate       *string
	NextStartDate *string
	NextEndDate   *string
	Page          int
	Limit         int
}

// CheckupService は健診記録のビジネスロジックを定義するインターフェース
func buildCheckupUpdate(input *UpdateCheckupInput) map[string]any {
	fields := map[string]any{}
	if input.CheckupTypeID != nil {
		fields["checkup_type_id"] = *input.CheckupTypeID
	}
	if input.PetID != nil {
		fields["pet_id"] = *input.PetID
	}
	if input.Date != nil {
		fields["date"] = *input.Date
	}
	if input.NextDate != nil {
		fields["next_date"] = *input.NextDate
	}
	if input.DoctorIDClear != nil && *input.DoctorIDClear {
		fields["doctor_id"] = nil
	} else if input.DoctorID != nil {
		fields["doctor_id"] = *input.DoctorID
	}
	if input.Result != nil {
		fields["result"] = *input.Result
	}
	return fields
}

type CheckupService interface {
	List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error)
	ListByClinic(ctx context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, int64, error)
	Create(ctx context.Context, medicalRecordID uint64, input *CreateCheckupInput) (*model.Checkup, error)
	Update(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, input *UpdateCheckupInput) (*model.Checkup, error)
	Delete(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) error
	// Wait は健診フォローアップトリガー goroutine（fire-and-forget）の完了を待つ。
	// graceful shutdown で server.Shutdown(ctx) の後に呼び出し、goroutine の孤児化を防ぐ（BE-refactor.md B-1）。
	Wait()
}

type checkupService struct {
	repo                 CheckupRepository
	medicalRecordRepo    medicalRecordFinder
	checkupTypeRepo      CheckupTypeRepository
	lstepDeliveryTrigger checkupFollowUpTrigger
	tagSyncSvc           checkupTagSyncer
	// wg は健診フォローアップトリガーの fire-and-forget goroutine を追跡する（BE-refactor.md B-1）。
	wg sync.WaitGroup
}

// Wait は送信中の健診フォローアップトリガー goroutine の完了を待つ（BE-refactor.md B-1）。
func (s *checkupService) Wait() {
	s.wg.Wait()
}

// NewCheckupService は CheckupService の実装を返す
func NewCheckupService(repo CheckupRepository, medicalRecordRepo medicalRecordFinder, checkupTypeRepo CheckupTypeRepository, lstepDeliveryTrigger checkupFollowUpTrigger, tagSyncSvc checkupTagSyncer) CheckupService {
	return &checkupService{repo: repo, medicalRecordRepo: medicalRecordRepo, checkupTypeRepo: checkupTypeRepo, lstepDeliveryTrigger: lstepDeliveryTrigger, tagSyncSvc: tagSyncSvc}
}

func (s *checkupService) List(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Checkup, error) {
	result, err := s.repo.FindByMedicalRecordID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list checkups", "error", err)
		return nil, apperrors.Wrap(err, "failed to list checkups")
	}
	return result, nil
}

func (s *checkupService) ListByClinic(ctx context.Context, input ListCheckupsByClinicInput) ([]model.Checkup, int64, error) {
	slog.InfoContext(ctx, "listing checkups by clinic", slog.Uint64("clinic_id", input.ClinicID))
	result, total, err := s.repo.FindByClinicID(ctx, input.ClinicID, CheckupFilters{
		StartDate:     input.StartDate,
		EndDate:       input.EndDate,
		NextStartDate: input.NextStartDate,
		NextEndDate:   input.NextEndDate,
	}, input.Page, input.Limit)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list checkups by clinic", "error", err)
		return nil, 0, apperrors.Wrap(err, "failed to list checkups by clinic")
	}
	return result, total, nil
}

func (s *checkupService) Create(ctx context.Context, medicalRecordID uint64, input *CreateCheckupInput) (*model.Checkup, error) {
	parent, err := s.medicalRecordRepo.FindByID(ctx, input.ClinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to find medical record")
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return nil, apperrors.WrapConflict("確定済みカルテのため健診記録は追加できません")
	}
	// クロステナント write 防止: checkup_type が caller の clinic に属することを検証する。
	if input.CheckupTypeID != 0 {
		if _, err := s.checkupTypeRepo.FindByID(ctx, input.ClinicID, input.CheckupTypeID); err != nil {
			return nil, apperrors.Wrap(err, "failed to verify checkup type ownership")
		}
	}
	checkup := &model.Checkup{
		ClinicID:        input.ClinicID,
		MedicalRecordID: medicalRecordID,
		CheckupTypeID:   input.CheckupTypeID,
		PetID:           input.PetID,
		Date:            input.Date,
		NextDate:        input.NextDate,
		DoctorID:        input.DoctorID,
		Result:          input.Result,
	}
	if err := s.repo.Create(ctx, checkup); err != nil {
		slog.ErrorContext(ctx, "failed to create checkup", "error", err)
		return nil, apperrors.Wrap(err, "failed to create checkup")
	}
	slog.InfoContext(ctx, "checkup created",
		slog.Uint64("clinic_id", input.ClinicID),
		slog.Uint64("checkup_id", checkup.ID),
		slog.Uint64("medical_record_id", medicalRecordID))
	created, err := s.repo.FindByID(ctx, input.ClinicID, checkup.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get checkup after create", "error", err)
		return nil, apperrors.Wrap(err, "failed to get checkup after create")
	}

	// 健診フォローアップトリガー（非同期・非致命的）
	if s.lstepDeliveryTrigger != nil && created.MedicalRecord != nil && created.MedicalRecord.OwnerID != nil {
		clinicID := input.ClinicID
		ownerID := *created.MedicalRecord.OwnerID
		svc := s.lstepDeliveryTrigger
		// repohelpers.DetachTx: HTTP request の cancel から切離しつつ tracing context を保持。
		// M-4: fire-and-forget goroutine のため、context.WithTimeout で明示的に制限を設定。
		// goroutine 境界を跨ぐ WithoutCancel は ambient tx を剥がすこと（BE-refactor.md B-2）。
		bgCtx := repohelpers.DetachTx(ctx)
		s.wg.Add(1)
		goSafe("checkup followup trigger", func() {
			defer s.wg.Done()
			trigCtx, cancel := context.WithTimeout(bgCtx, 35*time.Second)
			defer cancel()
			if err := svc.TriggerCheckupFollowUp(trigCtx, clinicID, ownerID); err != nil {
				slog.WarnContext(trigCtx, "checkup followup trigger failed (non-fatal)", "error", err, "owner_id", ownerID)
			}
		})
	}
	s.syncCheckupTag(ctx, input.ClinicID, created)

	return created, nil
}

func (s *checkupService) Update(ctx context.Context, clinicID, medicalRecordID, checkupID uint64, input *UpdateCheckupInput) (*model.Checkup, error) {
	// 親カルテ所属確認（clinic_id スコープ済み）
	existing, err := s.repo.FindByID(ctx, clinicID, checkupID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get checkup", "error", err)
		return nil, apperrors.Wrap(err, "failed to get checkup")
	}
	if existing.MedicalRecordID != medicalRecordID {
		return nil, apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", checkupID))
	}
	parent, err := s.medicalRecordRepo.FindByID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find medical record", "error", err)
		return nil, apperrors.Wrap(err, "failed to find medical record")
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return nil, apperrors.WrapConflict("確定済みカルテのため健診記録は編集できません")
	}
	// クロステナント write 防止: 貼り替え先 checkup_type の所有権を検証する。
	if err := validateOwnedMasterFK(ctx, "checkup type", clinicID, input.CheckupTypeID,
		func(actx context.Context, cid, mid uint64) error {
			_, err := s.checkupTypeRepo.FindByID(actx, cid, mid)
			return err
		}); err != nil {
		return nil, err
	}
	fields := buildCheckupUpdate(input)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput("at least one field must be provided")
	}
	if err := s.repo.Update(ctx, clinicID, checkupID, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update checkup", "error", err)
		return nil, apperrors.Wrap(err, "failed to update checkup")
	}
	slog.InfoContext(ctx, "checkup updated",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("checkup_id", checkupID),
		slog.Uint64("medical_record_id", medicalRecordID))
	updated, err := s.repo.FindByID(ctx, clinicID, checkupID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get checkup after update", "error", err)
		return nil, apperrors.Wrap(err, "failed to get checkup after update")
	}
	s.resyncOwnerCheckupTags(ctx, clinicID, updated)
	return updated, nil
}

func (s *checkupService) Delete(ctx context.Context, clinicID, medicalRecordID, checkupID uint64) error {
	existing, err := s.repo.FindByID(ctx, clinicID, checkupID)
	if err != nil {
		return apperrors.Wrap(err, "failed to get checkup")
	}
	if existing.MedicalRecordID != medicalRecordID {
		return apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", checkupID))
	}
	parent, err := s.medicalRecordRepo.FindByID(ctx, clinicID, medicalRecordID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find medical record", "error", err)
		return apperrors.Wrap(err, "failed to find medical record")
	}
	if parent.Status == model.MedicalRecordStatusFinalized {
		return apperrors.WrapConflict("確定済みカルテのため健診記録は削除できません")
	}
	if err := s.repo.Delete(ctx, clinicID, checkupID); err != nil {
		slog.ErrorContext(ctx, "failed to delete checkup", "error", err, "clinic_id", clinicID, "checkup_id", checkupID)
		return apperrors.Wrap(err, "failed to delete checkup")
	}
	slog.InfoContext(ctx, "checkup deleted",
		slog.Uint64("clinic_id", clinicID),
		slog.Uint64("checkup_id", checkupID),
		slog.Uint64("medical_record_id", medicalRecordID))
	s.resyncOwnerCheckupTags(ctx, clinicID, existing)
	return nil
}

func (s *checkupService) syncCheckupTag(ctx context.Context, clinicID uint64, checkup *model.Checkup) {
	if s.tagSyncSvc == nil || checkup == nil || checkup.MedicalRecord == nil || checkup.MedicalRecord.OwnerID == nil {
		return
	}
	ownerID := *checkup.MedicalRecord.OwnerID
	if err := s.tagSyncSvc.SyncCheckupTag(ctx, clinicID, ownerID, checkup.CheckupTypeID, checkup.Date, checkup.NextDate); err != nil {
		slog.ErrorContext(ctx, "failed to sync checkup tag", "error", err, "clinic_id", clinicID, "owner_id", ownerID, "checkup_id", checkup.ID)
	}
}

func (s *checkupService) resyncOwnerCheckupTags(ctx context.Context, clinicID uint64, checkup *model.Checkup) {
	if s.tagSyncSvc == nil || checkup == nil || checkup.MedicalRecord == nil || checkup.MedicalRecord.OwnerID == nil {
		return
	}
	ownerID := *checkup.MedicalRecord.OwnerID
	if err := s.tagSyncSvc.ResyncOwnerCheckupTags(ctx, clinicID, ownerID); err != nil {
		slog.ErrorContext(ctx, "failed to resync owner checkup tags", "error", err, "clinic_id", clinicID, "owner_id", ownerID, "checkup_id", checkup.ID)
	}
}
