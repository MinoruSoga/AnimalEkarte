package lstep

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// LstepLifecycleService はペット死亡・オーナーオプトアウトなど
// Lステップタグのライフサイクルイベントを処理する（BE-017）。
type LstepLifecycleService interface {
	// HandlePetDeath はペット死亡を記録し CPM タグを再同期する。
	HandlePetDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string, actorID *uint64) error
	// HandlePetRevival はペット死亡取り消しを記録し CPM タグを再同期する。
	HandlePetRevival(ctx context.Context, clinicID, petID uint64, actorID *uint64) error
	// HandleOwnerOptOut はオプトアウトを記録し Lステップの全タグを解除する。
	HandleOwnerOptOut(ctx context.Context, clinicID, ownerID uint64, reason string, actorID *uint64) error
	// HandleOwnerOptIn はオプトインを記録し CPM タグを再同期する。
	HandleOwnerOptIn(ctx context.Context, clinicID, ownerID uint64, actorID *uint64) error
	// HandleOwnerDeletion は飼主削除前に Lステップの全タグを解除しキャッシュを削除する。
	HandleOwnerDeletion(ctx context.Context, clinicID, ownerID uint64) error
}

type lstepLifecycleService struct {
	settingsSvc   LstepSettingsService
	ownerRepo     lifecycleOwnerDependency
	petRepo       lifecyclePetDependency
	tagCacheRepo  lifecycleTagCacheRepository
	syncSvc       lifecycleTagSyncer
	auditSvc      lifecycleOperationAuditor
	tagConfigRepo lifecycleTagConfigRepository
	// transactor / auditTx: BUG-407 fail-closed 化。HandlePetDeath/HandlePetRevival の
	// status/deceased_at 更新と一次監査ログ書込を同一 tx で束ね、監査失敗で status 更新も
	// ロールバックする（#211 返金の fail-closed 先例と同型）。
	transactor lifecycleTransactor
	auditTx    LifecycleAuditTxLogger
}

// NewLstepLifecycleService は LstepLifecycleService を初期化して返す。
// tagConfigRepo が nil の場合はペット由来タグ削除でフォールバック値を使用する。
func NewLstepLifecycleService(
	settingsSvc LstepSettingsService,
	ownerRepo lifecycleOwnerDependency,
	petRepo lifecyclePetDependency,
	tagCacheRepo lifecycleTagCacheRepository,
	syncSvc lifecycleTagSyncer,
	auditSvc lifecycleOperationAuditor,
	tagConfigRepo lifecycleTagConfigRepository,
	transactor lifecycleTransactor,
	auditTx LifecycleAuditTxLogger,
) LstepLifecycleService {
	return &lstepLifecycleService{
		settingsSvc:   settingsSvc,
		ownerRepo:     ownerRepo,
		petRepo:       petRepo,
		tagCacheRepo:  tagCacheRepo,
		syncSvc:       syncSvc,
		auditSvc:      auditSvc,
		tagConfigRepo: tagConfigRepo,
		transactor:    transactor,
		auditTx:       auditTx,
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
func (s *lstepLifecycleService) HandlePetDeath(ctx context.Context, clinicID, petID uint64, deceasedAt time.Time, reason string, actorID *uint64) error {
	// P1: FindByID before Update
	pet, err := s.petRepo.FindByID(ctx, clinicID, petID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find pet")
	}
	if pet.Status == model.PetStatusDeceased {
		return apperrors.WrapConflict("死亡記録は既に登録されています")
	}
	// BUG-407: status は deceased_at と独立した二重管理フィールドのため、
	// 同一 Update で "deceased" へ揃える。分離したままだと、外側フォームの
	// 生死ラジオが未追従のまま次回の外側「更新」で status="alive" に
	// 上書きされ、deceased_at だけ残る不整合状態になる。
	//
	// BUG-407 (audit fail-closed): status/deceased_at 更新と一次監査ログ（pet_death）書込を
	// 同一 tx で原子化する。監査書込が失敗したら status 更新もロールバックする — #211 返金の
	// fail-closed 先例と同型（監査書込自体の失敗が死亡記録という臨床アクションをブロックする、
	// という設計判断そのものが本 BUG-407 修正の核心。旧実装の best-effort/WarnContext は廃止）。
	actorType := sharedkernel.AuditActorTypeFor(actorID)
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.petRepo.RecordDeath(txCtx, clinicID, petID, deceasedAt, reason); err != nil {
			return apperrors.Wrap(err, "failed to record pet death")
		}
		if auditErr := s.auditTx.LogEntryTx(txCtx, &LifecycleAuditEntry{
			ClinicID:   &clinicID,
			ActorID:    actorID,
			ActorType:  actorType,
			Action:     "pet_death",
			Resource:   "pet",
			ResourceID: &petID,
		}); auditErr != nil {
			return apperrors.Wrap(auditErr, "failed to write pet death audit log")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to record pet death in transaction")
	}

	ownerID := pet.OwnerID

	// 生存ペットを確認 — 全滅時は配信停止としてタグを全解除
	// LSB-03 / X-01: death は既に commit 済み。後段 read 失敗で応答を失敗へ反転しない。
	livingPets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find living pets after pet death", "error", err, "pet_id", petID)
		return nil
	}

	if len(livingPets) == 0 {
		s.clearAllTagsIfLinked(ctx, clinicID, ownerID)
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
	if err := s.auditSvc.LogLstepOperation(ctx, clinicID, nil, "pet_death_tag_sync", "pet", &petID); err != nil {
		slog.WarnContext(ctx, "audit log failed for pet death tag sync", "error", err, "pet_id", petID)
	}

	return nil
}

func (s *lstepLifecycleService) clearAllTagsIfLinked(ctx context.Context, clinicID, ownerID uint64) {
	owner, findErr := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if findErr != nil {
		slog.ErrorContext(ctx, "failed to load owner for lstep tag clear", "error", findErr, "clinic_id", clinicID, "owner_id", ownerID)
		return
	}
	if owner == nil || owner.LineUserID == nil || *owner.LineUserID == "" {
		slog.InfoContext(ctx, "lstep tag clear skipped: owner is not LINE-linked", "clinic_id", clinicID, "owner_id", ownerID)
		return
	}
	if removeErr := s.removeAllTagsFromLstep(ctx, clinicID, ownerID, *owner.LineUserID); removeErr != nil {
		slog.ErrorContext(ctx, "failed to remove lstep tags on all-pets-dead", "error", removeErr)
	}
}

// HandlePetRevival はペット死亡取り消しを記録し CPM タグを再同期する。
func (s *lstepLifecycleService) HandlePetRevival(ctx context.Context, clinicID, petID uint64, actorID *uint64) error {
	// P1: FindByID before Update
	pet, err := s.petRepo.FindByID(ctx, clinicID, petID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find pet")
	}
	if pet.Status != model.PetStatusDeceased {
		return apperrors.WrapConflict("死亡記録が登録されていないため解除できません")
	}
	// BUG-407: 死亡取り消し時も status を "alive" に戻し、deceased_at/status の
	// 二重管理不整合を防ぐ（HandlePetDeath と対称）。
	//
	// BUG-407 (audit fail-closed): status 更新と一次監査ログ（pet_revival）書込を同一 tx で
	// 原子化する。監査書込が失敗したら status 更新もロールバックする（HandlePetDeath と対称）。
	actorType := sharedkernel.AuditActorTypeFor(actorID)
	if err := s.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.petRepo.ClearDeath(txCtx, clinicID, petID); err != nil {
			return apperrors.Wrap(err, "failed to record pet revival")
		}
		if auditErr := s.auditTx.LogEntryTx(txCtx, &LifecycleAuditEntry{
			ClinicID:   &clinicID,
			ActorID:    actorID,
			ActorType:  actorType,
			Action:     "pet_revival",
			Resource:   "pet",
			ResourceID: &petID,
		}); auditErr != nil {
			return apperrors.Wrap(auditErr, "failed to write pet revival audit log")
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "failed to record pet revival in transaction")
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
	if err := s.auditSvc.LogLstepOperation(ctx, clinicID, nil, "pet_revival_tag_sync", "pet", &petID); err != nil {
		slog.WarnContext(ctx, "audit log failed for pet revival tag sync", "error", err, "pet_id", petID)
	}

	return nil
}

// HandleOwnerOptOut はオプトアウトを記録し Lステップの全タグを解除する。
func (s *lstepLifecycleService) HandleOwnerOptOut(ctx context.Context, clinicID, ownerID uint64, reason string, actorID *uint64) error {
	// P1: FindByID before Update
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find owner")
	}

	now := time.Now()
	if err := s.ownerRepo.RecordLstepOptOut(ctx, clinicID, ownerID, now, reason); err != nil {
		return apperrors.Wrap(err, "failed to record owner opt-out")
	}

	// LSA-05: 配信同意変更は destructive — best-effort 監査を残す（actor 付き）
	if auditErr := s.auditSvc.LogLstepOperation(ctx, clinicID, actorID, "owner_lstep_opt_out", "owner", &ownerID); auditErr != nil {
		slog.WarnContext(ctx, "audit log failed for owner lstep opt-out", "error", auditErr, "owner_id", ownerID)
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
func (s *lstepLifecycleService) HandleOwnerOptIn(ctx context.Context, clinicID, ownerID uint64, actorID *uint64) error {
	// P1: FindByID before Update
	if _, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID); err != nil {
		return apperrors.Wrap(err, "failed to find owner")
	}

	if err := s.ownerRepo.ClearLstepOptOut(ctx, clinicID, ownerID); err != nil {
		return apperrors.Wrap(err, "failed to record owner opt-in")
	}

	// LSA-05: 配信同意変更の監査
	if auditErr := s.auditSvc.LogLstepOperation(ctx, clinicID, actorID, "owner_lstep_opt_in", "owner", &ownerID); auditErr != nil {
		slog.WarnContext(ctx, "audit log failed for owner lstep opt-in", "error", auditErr, "owner_id", ownerID)
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
		// owner.DeleteOwner は戻り値を捨てる（best-effort）。RespondError しないので診断ログを残す。
		slog.ErrorContext(ctx, "failed to find owner for deletion cleanup", "error", err)
		return apperrors.Wrap(err, "failed to find owner")
	}

	// LSA-05: 削除前クリーンアップは irreversible — actor は owner HTTP 側に無いため nil で監査
	if auditErr := s.auditSvc.LogLstepOperation(ctx, clinicID, nil, "owner_lstep_deletion_cleanup", "owner", &ownerID); auditErr != nil {
		slog.WarnContext(ctx, "audit log failed for owner deletion lstep cleanup", "error", auditErr, "owner_id", ownerID)
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

// removeAllTagsFromLstep はキャッシュを参照して Lステップからタグを解除する（LSB-02 / DEC-35）。
// リモート解除に成功したタグだけローカル cache から DeleteTag する。
// client==nil または RemoveTag 失敗時は cache を残し、再試行可能な根拠を保持する。
func (s *lstepLifecycleService) removeAllTagsFromLstep(ctx context.Context, clinicID, ownerID uint64, lineUserID string) error {
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}

	tags, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		return apperrors.Wrap(err, "failed to find tag cache")
	}

	if client == nil {
		// 同期無効・APIキー未設定: リモート解除不能なのに cache を消さない（LSB-02）
		slog.WarnContext(ctx, "lstep client unavailable; keeping tag cache for retry", "owner_id", ownerID, "tag_count", len(tags))
		return nil
	}

	var failed int
	for _, t := range tags {
		if removeErr := client.RemoveTag(ctx, lineUserID, t.TagName); removeErr != nil {
			failed++
			slog.ErrorContext(ctx, "failed to remove tag on cleanup", "error", removeErr, "tag", t.TagName)
			continue
		}
		if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, t.TagName); delErr != nil {
			failed++
			slog.ErrorContext(ctx, "failed to delete tag cache after remote remove", "error", delErr, "tag", t.TagName)
		}
	}
	if failed > 0 {
		slog.WarnContext(ctx, "partial lstep tag cleanup; residual cache retained for retry",
			"owner_id", ownerID, "failed", failed, "total", len(tags))
	}

	return nil
}

// removePetDerivedTagsFromLstep は死亡ペット由来のタグ（ワクチン・健診カテゴリ）を
// Lステップおよびキャッシュから解除する（best-effort）。
// DB に登録されたプレフィックス（C2 カテゴリ: vaccine_/checkup_ 等）を使用し、
// DB 未設定時はフォールバック値 ["vaccine_", "checkup_done_"] を使用する。
func (s *lstepLifecycleService) removePetDerivedTagsFromLstep(ctx context.Context, client lstep.Client, clinicID, ownerID uint64, lineUserID string) {
	petDerivedPrefixes := s.loadPetDerivedPrefixes(ctx)
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for pet-derived tag removal", "error", err)
		return
	}
	for _, t := range cached {
		for _, prefix := range petDerivedPrefixes {
			if !strings.HasPrefix(t.TagName, prefix) {
				continue
			}
			if removeErr := client.RemoveTag(ctx, lineUserID, t.TagName); removeErr != nil {
				slog.ErrorContext(ctx, "failed to remove pet-derived tag", "error", removeErr, "tag", t.TagName)
				break
			}
			if delErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, t.TagName); delErr != nil {
				slog.ErrorContext(ctx, "failed to delete pet-derived tag cache", "error", delErr, "tag", t.TagName)
			}
			break
		}
	}
}

// loadPetDerivedPrefixes は DB から C2 カテゴリのプレフィックスを読み込む。
// DB 未設定またはエラー時は ["vaccine_", "checkup_done_"] にフォールバックする。
func (s *lstepLifecycleService) loadPetDerivedPrefixes(ctx context.Context) []string {
	if s.tagConfigRepo == nil {
		return petDerivedPrefixFallback()
	}
	all, err := s.tagConfigRepo.FindAllAutoManagedPrefixes(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load auto managed prefixes for pet-derived tags, using fallback", "error", err)
		return petDerivedPrefixFallback()
	}
	var prefixes []string
	for _, p := range all {
		if p.Category == "C2" {
			prefixes = append(prefixes, p.Prefix)
		}
	}
	if len(prefixes) == 0 {
		return petDerivedPrefixFallback()
	}
	return prefixes
}

// petDerivedPrefixFallback は DB 未設定時に使用する静的フォールバック値を返す。
// migration 006 の C2 シードが存在すれば通常はこちらは使われない。
func petDerivedPrefixFallback() []string {
	return []string{"vaccine_", TagPrefixCheckupDone}
}
