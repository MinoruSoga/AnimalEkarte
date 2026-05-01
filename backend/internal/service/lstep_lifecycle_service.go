package service

import (
	"context"
	"log/slog"
	"strings"
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
	auditSvc     AuditService
}

// NewLstepLifecycleService は LstepLifecycleService を初期化して返す。
func NewLstepLifecycleService(
	settingsSvc LstepSettingsService,
	ownerRepo repository.OwnerRepository,
	petRepo repository.PetRepository,
	tagCacheRepo repository.LstepTagCacheRepository,
	syncSvc LstepTagSyncService,
	auditSvc AuditService,
) LstepLifecycleService {
	return &lstepLifecycleService{
		settingsSvc:  settingsSvc,
		ownerRepo:    ownerRepo,
		petRepo:      petRepo,
		tagCacheRepo: tagCacheRepo,
		syncSvc:      syncSvc,
		auditSvc:     auditSvc,
	}
}

// buildClient はクリニック設定から lstep.Client を構築する。
// 同期無効（is_sync_enabled=false）または API キー未設定の場合は nil, nil を返す。
func (s *lstepLifecycleService) buildClient(ctx context.Context, clinicID uint64) (lstep.Client, error) {
	enabled, err := s.settingsSvc.IsSyncEnabled(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check lstep sync enabled")
	}
	if !enabled {
		return nil, nil
	}
	apiKey, baseURL, _, err := s.settingsSvc.GetRawCredentials(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get lstep credentials")
	}
	if apiKey == "" {
		return nil, nil
	}
	return lstep.NewClient(apiKey, baseURL), nil
}

// HandlePetDeath はペット死亡を記録し、オーナーのタグを再同期する。
// 全ペットが死亡した場合は Lステップ全タグを解除する（配信停止）。
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

	ownerID := pet.OwnerID

	// 生存ペットを確認 — 全滅時は配信停止としてタグを全解除
	livingPets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find living pets after pet death", "error", err)
		// タグ同期失敗は死亡記録を巻き戻さない
		return nil
	}

	if len(livingPets) == 0 {
		// 全ペット死亡 → Lステップタグを全解除
		owner, findErr := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
		if findErr == nil && owner != nil && owner.LineUserID != nil && *owner.LineUserID != "" {
			if removeErr := s.removeAllTagsFromLstep(ctx, clinicID, ownerID, *owner.LineUserID); removeErr != nil {
				slog.ErrorContext(ctx, "failed to remove lstep tags on all-pets-dead", "error", removeErr)
			}
		}
		return nil
	}

	// 生存ペットあり — ペット由来タグを再同期（best-effort）
	if syncErr := s.syncSvc.SyncOwnerAnimalClassificationTags(ctx, clinicID, ownerID); syncErr != nil {
		slog.ErrorContext(ctx, "failed to sync animal classification tags after pet death", "error", syncErr)
	}
	if syncErr := s.syncSvc.SyncPetBasicInfoTags(ctx, clinicID, ownerID); syncErr != nil {
		slog.ErrorContext(ctx, "failed to sync pet basic info tags after pet death", "error", syncErr)
	}
	if syncErr := s.syncSvc.SyncCPMStageTag(ctx, clinicID, ownerID); syncErr != nil {
		slog.ErrorContext(ctx, "failed to sync CPM tag after pet death", "error", syncErr)
	}

	// 死亡ペット由来のワクチン・健診タグを解除する（best-effort）
	if cleanupOwner, findErr := s.ownerRepo.FindByID(ctx, clinicID, ownerID); findErr == nil &&
		cleanupOwner != nil && cleanupOwner.LineUserID != nil && *cleanupOwner.LineUserID != "" {
		if client, clientErr := s.buildClient(ctx, clinicID); clientErr == nil && client != nil {
			s.removePetDerivedTagsFromLstep(ctx, client, clinicID, ownerID, *cleanupOwner.LineUserID)
		}
	}

	// 監査ログ（best-effort）
	_ = s.auditSvc.LogLstepOperation(ctx, clinicID, nil, "pet_death_tag_sync", "pet", &petID)

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

	if syncErr := s.syncSvc.SyncOwnerAnimalClassificationTags(ctx, clinicID, pet.OwnerID); syncErr != nil {
		slog.ErrorContext(ctx, "failed to sync animal classification tags after pet revival", "error", syncErr)
	}
	if syncErr := s.syncSvc.SyncPetBasicInfoTags(ctx, clinicID, pet.OwnerID); syncErr != nil {
		slog.ErrorContext(ctx, "failed to sync pet basic info tags after pet revival", "error", syncErr)
	}
	if syncErr := s.syncSvc.SyncCPMStageTag(ctx, clinicID, pet.OwnerID); syncErr != nil {
		slog.ErrorContext(ctx, "failed to sync CPM tag after pet revival", "error", syncErr)
	}

	// 監査ログ（best-effort）
	_ = s.auditSvc.LogLstepOperation(ctx, clinicID, nil, "pet_revival_tag_sync", "pet", &petID)

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

// removePetDerivedTagsFromLstep は死亡ペット由来のタグ（ワクチン・健診カテゴリ）を
// Lステップおよびキャッシュから解除する（best-effort）。
func (s *lstepLifecycleService) removePetDerivedTagsFromLstep(ctx context.Context, client lstep.Client, clinicID, ownerID uint64, lineUserID string) {
	petDerivedPrefixes := []string{"vaccine_dog_", "vaccine_cat_", "vaccine_rabies_", "checkup_done_"}
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for pet-derived tag removal", "error", err)
		return
	}
	for _, t := range cached {
		for _, prefix := range petDerivedPrefixes {
			if strings.HasPrefix(t.TagName, prefix) {
				if removeErr := client.RemoveTag(ctx, lineUserID, t.TagName); removeErr != nil {
					slog.ErrorContext(ctx, "failed to remove pet-derived tag", "error", removeErr, "tag", t.TagName)
				} else {
					_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, t.TagName)
				}
				break
			}
		}
	}
}
