package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/repository"
)

// LstepLifecycleService はペット死亡・オーナーオプトアウトなど
// Lステップタグのライフサイクルイベントを処理する（BE-017）。
type LstepLifecycleService interface {
	// HandlePetDeath はペット死亡を記録し CPM タグを再同期する。
	HandlePetDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error
	// HandlePetRevival はペット死亡取り消しを記録し CPM タグを再同期する。
	HandlePetRevival(ctx context.Context, clinicID, petID uint64) error
	// HandleOwnerOptOut はオプトアウトを記録し Lステップの全タグを解除する。
	HandleOwnerOptOut(ctx context.Context, clinicID, ownerID uint64, reason string) error
	// HandleOwnerOptIn はオプトインを記録し CPM タグを再同期する。
	HandleOwnerOptIn(ctx context.Context, clinicID, ownerID uint64) error
	// HandleOwnerDeletion は飼主削除前に Lステップの全タグを解除しキャッシュを削除する。
	HandleOwnerDeletion(ctx context.Context, clinicID, ownerID uint64) error
}

type lstepLifecycleService struct {
	settingsSvc  LstepSettingsService
	ownerRepo    repository.OwnerRepository
	petRepo      repository.PetRepository
	tagCacheRepo repository.LstepTagCacheRepository
	syncSvc      LstepTagSyncService
}

// NewLstepLifecycleService は LstepLifecycleService を初期化して返す。
func NewLstepLifecycleService(
	settingsSvc LstepSettingsService,
	ownerRepo repository.OwnerRepository,
	petRepo repository.PetRepository,
	tagCacheRepo repository.LstepTagCacheRepository,
	syncSvc LstepTagSyncService,
) LstepLifecycleService {
	return &lstepLifecycleService{
		settingsSvc:  settingsSvc,
		ownerRepo:    ownerRepo,
		petRepo:      petRepo,
		tagCacheRepo: tagCacheRepo,
		syncSvc:      syncSvc,
	}
}

// buildClient はクリニック設定から lstep.Client を構築する。
// API キーが未設定の場合は nil, nil を返す。
func (s *lstepLifecycleService) buildClient(ctx context.Context, clinicID uint64) (lstep.Client, error) {
	apiKey, baseURL, _, err := s.settingsSvc.GetRawCredentials(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get lstep credentials")
	}
	if apiKey == "" {
		return nil, nil
	}
	return lstep.NewClient(apiKey, baseURL), nil
}

// HandlePetDeath はペット死亡を記録し、オーナーの CPM タグを再同期する。
func (s *lstepLifecycleService) HandlePetDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string) error {
	// P1: FindByID before Update
	pet, err := s.petRepo.FindByID(ctx, clinicID, petID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find pet for death recording", "error", err)
		return apperrors.Wrap(err, "failed to find pet")
	}

	fields := map[string]any{
		"deceased_at":     deceasedAt,
		"deceased_reason": reason,
	}
	if err := s.petRepo.Update(ctx, clinicID, petID, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update pet deceased fields", "error", err)
		return apperrors.Wrap(err, "failed to record pet death")
	}

	// CPM 再同期は best-effort — 死亡記録の成否に影響させない
	if syncErr := s.syncSvc.SyncCPMStageTag(ctx, clinicID, pet.OwnerID); syncErr != nil {
		slog.ErrorContext(ctx, "failed to sync CPM tag after pet death", "error", syncErr)
	}

	return nil
}

// HandlePetRevival はペット死亡取り消しを記録し CPM タグを再同期する。
func (s *lstepLifecycleService) HandlePetRevival(ctx context.Context, clinicID, petID uint64) error {
	// P1: FindByID before Update
	pet, err := s.petRepo.FindByID(ctx, clinicID, petID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find pet for revival", "error", err)
		return apperrors.Wrap(err, "failed to find pet")
	}

	fields := map[string]any{
		"deceased_at":     nil,
		"deceased_reason": nil,
	}
	if err := s.petRepo.Update(ctx, clinicID, petID, fields); err != nil {
		slog.ErrorContext(ctx, "failed to clear pet deceased fields", "error", err)
		return apperrors.Wrap(err, "failed to record pet revival")
	}

	if syncErr := s.syncSvc.SyncCPMStageTag(ctx, clinicID, pet.OwnerID); syncErr != nil {
		slog.ErrorContext(ctx, "failed to sync CPM tag after pet revival", "error", syncErr)
	}

	return nil
}

// HandleOwnerOptOut はオプトアウトを記録し Lステップの全タグを解除する。
func (s *lstepLifecycleService) HandleOwnerOptOut(ctx context.Context, clinicID, ownerID uint64, reason string) error {
	// P1: FindByID before Update
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner for opt-out", "error", err)
		return apperrors.Wrap(err, "failed to find owner")
	}

	now := time.Now()
	fields := map[string]any{
		"lstep_opt_out":        true,
		"lstep_opt_out_at":     now,
		"lstep_opt_out_reason": reason,
	}
	if err := s.ownerRepo.Update(ctx, clinicID, ownerID, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update owner opt-out fields", "error", err)
		return apperrors.Wrap(err, "failed to record owner opt-out")
	}

	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	// タグ解除は best-effort — opt-out 記録の成否に影響させない
	if removeErr := s.removeAllTagsFromLstep(ctx, clinicID, ownerID, *owner.LineUserID); removeErr != nil {
		slog.ErrorContext(ctx, "failed to remove lstep tags on opt-out", "error", removeErr)
	}

	return nil
}

// HandleOwnerOptIn はオプトインを記録し CPM タグを再同期する。
func (s *lstepLifecycleService) HandleOwnerOptIn(ctx context.Context, clinicID, ownerID uint64) error {
	// P1: FindByID before Update
	if _, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID); err != nil {
		slog.ErrorContext(ctx, "failed to find owner for opt-in", "error", err)
		return apperrors.Wrap(err, "failed to find owner")
	}

	fields := map[string]any{
		"lstep_opt_out":        false,
		"lstep_opt_out_at":     nil,
		"lstep_opt_out_reason": nil,
	}
	if err := s.ownerRepo.Update(ctx, clinicID, ownerID, fields); err != nil {
		slog.ErrorContext(ctx, "failed to update owner opt-in fields", "error", err)
		return apperrors.Wrap(err, "failed to record owner opt-in")
	}

	if syncErr := s.syncSvc.SyncCPMStageTag(ctx, clinicID, ownerID); syncErr != nil {
		slog.ErrorContext(ctx, "failed to sync CPM tag after opt-in", "error", syncErr)
	}

	return nil
}

// HandleOwnerDeletion は飼主削除前に Lステップの全タグを解除しキャッシュを削除する。
func (s *lstepLifecycleService) HandleOwnerDeletion(ctx context.Context, clinicID, ownerID uint64) error {
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner for deletion cleanup", "error", err)
		return apperrors.Wrap(err, "failed to find owner")
	}

	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	if removeErr := s.removeAllTagsFromLstep(ctx, clinicID, ownerID, *owner.LineUserID); removeErr != nil {
		slog.ErrorContext(ctx, "failed to remove lstep tags on owner deletion", "error", removeErr)
		// タグ解除失敗は削除フローを止めない
	}

	return nil
}

// removeAllTagsFromLstep はキャッシュを参照して Lステップから全タグを解除しキャッシュを削除する。
func (s *lstepLifecycleService) removeAllTagsFromLstep(ctx context.Context, clinicID, ownerID uint64, lineUserID string) error {
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}

	tags, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find tag cache")
	}

	if client != nil {
		for _, t := range tags {
			if removeErr := client.RemoveTag(ctx, lineUserID, t.TagName); removeErr != nil {
				slog.ErrorContext(ctx, "failed to remove tag on cleanup", "error", removeErr, "tag", t.TagName)
			}
		}
	}

	if err := s.tagCacheRepo.DeleteAllByOwner(ctx, clinicID, ownerID); err != nil {
		return apperrors.Wrap(err, "failed to delete tag cache")
	}

	return nil
}
