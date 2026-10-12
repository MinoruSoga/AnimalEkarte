package lstep

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// deliveryBatchOwnerIDsLoader is satisfied by owner repositories that can bulk-load
// by ID (e.g. owner.Repository.FindByIDs). Consumer interface stays narrow;
// runBatch type-asserts at the page boundary (N+1 avoidance).
type deliveryBatchOwnerIDsLoader interface {
	FindByIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) ([]*model.Owner, error)
}

// deliveryBatchTagOwnersLoader is satisfied by tag-cache repositories that can
// bulk-load by owner set (LstepTagCacheRepository.FindByOwners).
type deliveryBatchTagOwnersLoader interface {
	FindByOwners(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error)
}

// deliveryBatchDayLogPageSize bounds FindByDateRangeWithFilters pages when
// preloading clinic-day trigger logs for already-fired / suppression checks.
const deliveryBatchDayLogPageSize = 500

// deliveryRevalidationHook は送信直前の candidate 再検証フック（EMR-308）。
// valid=true で送信する場合、hook が send コールバックを直列化境界（tentative
// mutation と同じ advisory lock）の内側で呼ぶ責務を持つ — 再検証〜送信の間の
// stale 化を防ぐため send は hook が制御する（Codex 指摘対応）。
// 戻り値 (valid, reason, err): valid=false のとき reason を excluded_reason として
// ログを excluded に更新し送信しない（send 未呼出）。valid=true は send 完了済みを
// 意味する。nil hook で再検証なし（呼び出し側が send を直接実行）。
type deliveryRevalidationHook func(ctx context.Context, ownerID uint64, send func(context.Context) error) (valid bool, reason string, err error)

func (s *lstepDeliveryTriggerService) runBatch(
	ctx context.Context,
	clinicID uint64,
	ownerIDs []uint64,
	triggerType string,
	tagName string,
	asOf time.Time,
	revalidate deliveryRevalidationHook,
) (int, []error) {
	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: failed to build lstep client", "clinic_id", clinicID, "trigger", triggerType, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to build lstep client")}
	}
	if client == nil {
		slog.InfoContext(ctx, "delivery trigger skipped: lstep client is not configured", "clinic_id", clinicID, "trigger", triggerType)
		return 0, nil
	}

	var errs []error
	count := 0
	if len(ownerIDs) == 0 {
		return 0, nil
	}

	// Clinic-day logs once per batch: already-fired + suppression share the same
	// source. Candidate-filtered so memory stays O(candidates × day triggers).
	candidateSet := make(map[uint64]struct{}, len(ownerIDs))
	for _, id := range ownerIDs {
		candidateSet[id] = struct{}{}
	}
	dayLogsByOwner, dayLogsLoaded := s.loadDeliveryDayLogsByOwner(ctx, clinicID, asOf, candidateSet)

	// Page candidates so bulk maps never grow unbounded with the full owner list.
	for start := 0; start < len(ownerIDs); start += lstepBatchPageSize {
		end := start + lstepBatchPageSize
		if end > len(ownerIDs) {
			end = len(ownerIDs)
		}
		page := ownerIDs[start:end]

		restore := s.installDeliveryBatchCaches(ctx, clinicID, page, dayLogsByOwner, dayLogsLoaded)
		for _, ownerID := range page {
			fired, loopErr := s.processSingleOwner(ctx, client, clinicID, ownerID, triggerType, tagName, asOf, revalidate)
			if loopErr != nil {
				errs = append(errs, loopErr)
				continue
			}
			if fired {
				count++
			}
		}
		restore()
	}
	return count, errs
}

// loadDeliveryDayLogsByOwner bulk-loads clinic-scoped day logs via the existing
// FindByDateRangeWithFilters API and indexes by owner_id for candidates only.
// On error, returns loaded=false so per-owner ExistsToday / FindByOwnerAndDate remain.
func (s *lstepDeliveryTriggerService) loadDeliveryDayLogsByOwner(
	ctx context.Context,
	clinicID uint64,
	asOf time.Time,
	candidateSet map[uint64]struct{},
) (map[uint64][]model.LstepDeliveryTriggerLog, bool) {
	if s.triggerLogRepo == nil || len(candidateSet) == 0 {
		return nil, false
	}
	dayStart, dayEnd, _ := jstHalfOpenDay(asOf)
	result := make(map[uint64][]model.LstepDeliveryTriggerLog)
	offset := 0
	for {
		rows, total, err := s.triggerLogRepo.FindByDateRangeWithFilters(
			ctx, clinicID, dayStart, dayEnd, "", "", deliveryBatchDayLogPageSize, offset,
		)
		if err != nil {
			slog.ErrorContext(ctx, "delivery trigger: bulk day-log load failed; falling back to per-owner reads",
				"clinic_id", clinicID, "error", err)
			return nil, false
		}
		for i := range rows {
			oid := rows[i].OwnerID
			if _, ok := candidateSet[oid]; !ok {
				continue
			}
			result[oid] = append(result[oid], rows[i].LstepDeliveryTriggerLog)
		}
		offset += len(rows)
		if len(rows) == 0 || offset >= int(total) || len(rows) < deliveryBatchDayLogPageSize {
			break
		}
	}
	return result, true
}

// installDeliveryBatchCaches bulk-loads owners/tags for one page and swaps service
// repos with cache wrappers so processSingleOwner / alreadyFiredToday /
// applySuppression / checkExclusion hit memory maps instead of per-owner SQL.
// restore must be called after the page loop (defers field restoration).
func (s *lstepDeliveryTriggerService) installDeliveryBatchCaches(
	ctx context.Context,
	clinicID uint64,
	page []uint64,
	dayLogsByOwner map[uint64][]model.LstepDeliveryTriggerLog,
	dayLogsLoaded bool,
) (restore func()) {
	origOwner := s.ownerRepo
	origTag := s.tagCacheRepo
	origLog := s.triggerLogRepo

	if loader, ok := any(s.ownerRepo).(deliveryBatchOwnerIDsLoader); ok && s.ownerRepo != nil && len(page) > 0 {
		owners, err := loader.FindByIDs(ctx, clinicID, page)
		if err != nil {
			slog.ErrorContext(ctx, "delivery trigger: bulk owner load failed; falling back to per-owner FindByID",
				"clinic_id", clinicID, "owner_count", len(page), "error", err)
		} else {
			s.ownerRepo = &deliveryBatchOwnerCache{inner: origOwner, byID: ownersByID(owners)}
		}
	}

	if loader, ok := any(s.tagCacheRepo).(deliveryBatchTagOwnersLoader); ok && s.tagCacheRepo != nil && len(page) > 0 {
		byOwner, err := loader.FindByOwners(ctx, clinicID, page)
		if err != nil {
			slog.ErrorContext(ctx, "delivery trigger: bulk tag-cache load failed; falling back to per-owner FindByOwner",
				"clinic_id", clinicID, "owner_count", len(page), "error", err)
		} else {
			s.tagCacheRepo = &deliveryBatchTagCache{inner: origTag, byOwner: coalesceOwnerMap(byOwner), loaded: true}
		}
	}

	if dayLogsLoaded && s.triggerLogRepo != nil {
		s.triggerLogRepo = &deliveryBatchTriggerLogCache{
			inner:          origLog,
			dayLogsByOwner: coalesceOwnerMap(dayLogsByOwner),
			loaded:         true,
		}
	}

	return func() {
		s.ownerRepo = origOwner
		s.tagCacheRepo = origTag
		s.triggerLogRepo = origLog
	}
}

// deliveryBatchOwnerCache serves FindByID from a page bulk map; cache miss falls
// back to the inner repo so mocks without real FindByIDs data keep working and
// production missing IDs still resolve via FindByID NotFound semantics.
type deliveryBatchOwnerCache struct {
	inner deliveryOwnerRepository
	byID  map[uint64]*model.Owner
}

func (c *deliveryBatchOwnerCache) FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error) {
	if o, ok := c.byID[ownerID]; ok {
		return o, nil
	}
	if c.inner == nil {
		return nil, apperrors.WrapNotFound("owner", fmt.Sprintf("%d", ownerID))
	}
	return c.inner.FindByID(ctx, clinicID, ownerID)
}

// deliveryBatchTagCache serves FindByOwner from FindByOwners bulk results.
// Missing keys are authoritative empty tag lists (same contract as FindByOwners).
type deliveryBatchTagCache struct {
	inner   deliveryTagCacheRepository
	byOwner map[uint64][]*model.LstepTagCache
	loaded  bool
}

func (c *deliveryBatchTagCache) FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]*model.LstepTagCache, error) {
	if c.loaded {
		tags := c.byOwner[ownerID]
		if tags == nil {
			return []*model.LstepTagCache{}, nil
		}
		return tags, nil
	}
	if c.inner == nil {
		return nil, nil
	}
	return c.inner.FindByOwner(ctx, clinicID, ownerID)
}

func (c *deliveryBatchTagCache) FindOwnerIDsByTag(ctx context.Context, clinicID uint64, tagName string) ([]uint64, error) {
	if c.inner == nil {
		return nil, nil
	}
	return c.inner.FindOwnerIDsByTag(ctx, clinicID, tagName)
}

// deliveryBatchTriggerLogCache serves ExistsTodayByOwnerAndType and
// FindByOwnerAndDate from clinic-day bulk preload. Writes and other methods
// pass through to the inner repository (LSA-15 claim lock stays on the real repo).
type deliveryBatchTriggerLogCache struct {
	inner          LstepDeliveryTriggerLogRepository
	dayLogsByOwner map[uint64][]model.LstepDeliveryTriggerLog
	loaded         bool
}

func (c *deliveryBatchTriggerLogCache) Create(ctx context.Context, log *model.LstepDeliveryTriggerLog) error {
	return c.inner.Create(ctx, log)
}

func (c *deliveryBatchTriggerLogCache) CreateIfAbsentToday(ctx context.Context, log *model.LstepDeliveryTriggerLog) (bool, error) {
	return c.inner.CreateIfAbsentToday(ctx, log)
}

func (c *deliveryBatchTriggerLogCache) ExistsTodayByOwnerAndType(
	ctx context.Context,
	clinicID, ownerID uint64,
	triggerType string,
	date time.Time,
) (bool, error) {
	if c.loaded {
		for i := range c.dayLogsByOwner[ownerID] {
			if c.dayLogsByOwner[ownerID][i].TriggerType == triggerType && !retryablePreSendClaim(&c.dayLogsByOwner[ownerID][i]) {
				return true, nil
			}
		}
		return false, nil
	}
	return c.inner.ExistsTodayByOwnerAndType(ctx, clinicID, ownerID, triggerType, date)
}

func (c *deliveryBatchTriggerLogCache) UpdateStatus(
	ctx context.Context,
	clinicID, id uint64,
	status string,
	firedAt *time.Time,
	excludedReason *string,
) error {
	return c.inner.UpdateStatus(ctx, clinicID, id, status, firedAt, excludedReason)
}

func (c *deliveryBatchTriggerLogCache) CountByStatusAndDateRange(
	ctx context.Context,
	clinicID uint64,
	from, to time.Time,
	triggerType string,
) (map[string]int64, error) {
	return c.inner.CountByStatusAndDateRange(ctx, clinicID, from, to, triggerType)
}

func (c *deliveryBatchTriggerLogCache) CountExcludedReasonByDateRange(
	ctx context.Context,
	clinicID uint64,
	from, to time.Time,
	triggerType string,
) (map[string]int64, error) {
	return c.inner.CountExcludedReasonByDateRange(ctx, clinicID, from, to, triggerType)
}

func (c *deliveryBatchTriggerLogCache) CountSuppressedByPriorityDateRange(
	ctx context.Context,
	clinicID uint64,
	from, to time.Time,
	triggerType string,
) (int64, error) {
	return c.inner.CountSuppressedByPriorityDateRange(ctx, clinicID, from, to, triggerType)
}

func (c *deliveryBatchTriggerLogCache) FindByDateRangeWithFilters(
	ctx context.Context,
	clinicID uint64,
	from, to time.Time,
	triggerType, status string,
	limit, offset int,
) ([]DeliveryTriggerLogRow, int64, error) {
	return c.inner.FindByDateRangeWithFilters(ctx, clinicID, from, to, triggerType, status, limit, offset)
}

func (c *deliveryBatchTriggerLogCache) CountByTypeAndStatus(
	ctx context.Context,
	clinicID uint64,
	from, to time.Time,
) ([]DeliveryStatsRow, error) {
	return c.inner.CountByTypeAndStatus(ctx, clinicID, from, to)
}

func (c *deliveryBatchTriggerLogCache) CountVisitConversionsByType(
	ctx context.Context,
	clinicID uint64,
	from, to time.Time,
	days int,
) ([]VisitConversionRow, error) {
	return c.inner.CountVisitConversionsByType(ctx, clinicID, from, to, days)
}

func (c *deliveryBatchTriggerLogCache) FindByOwnerAndDate(
	ctx context.Context,
	clinicID, ownerID uint64,
	date time.Time,
) ([]model.LstepDeliveryTriggerLog, error) {
	if c.loaded {
		logs := c.dayLogsByOwner[ownerID]
		if logs == nil {
			return []model.LstepDeliveryTriggerLog{}, nil
		}
		// Return a copy so demote/suppression callers cannot mutate the shared preload.
		out := make([]model.LstepDeliveryTriggerLog, len(logs))
		copy(out, logs)
		return out, nil
	}
	return c.inner.FindByOwnerAndDate(ctx, clinicID, ownerID, date)
}

func (c *deliveryBatchTriggerLogCache) UpdateSuppressed(ctx context.Context, clinicID, logID uint64, reason string) error {
	return c.inner.UpdateSuppressed(ctx, clinicID, logID, reason)
}

func (c *deliveryBatchTriggerLogCache) MarkKnownUnsent(ctx context.Context, clinicID, id uint64, reason string) error {
	return c.inner.MarkKnownUnsent(ctx, clinicID, id, reason)
}

// processSingleOwner は1飼い主分のトリガー処理を行い、配信実行されたか否かを返す。
// revalidate != nil の場合、日次スロット claim 後・LSTEP 送信直前に candidate の
// 有効性を再検証し、stale な candidate は excluded として記録して送信しない（EMR-308）。
func (s *lstepDeliveryTriggerService) processSingleOwner(
	ctx context.Context,
	client lstep.Client,
	clinicID, ownerID uint64,
	triggerType, tagName string,
	asOf time.Time,
	revalidate deliveryRevalidationHook,
) (bool, error) {
	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: failed to validate owner scope", "owner_id", ownerID, "error", err)
		return false, apperrors.Wrap(err, "failed to find owner")
	}
	// Fail-closed: bulk cache (ce79e0c23) may fall back to FindByID after a map miss.
	// A (nil, nil) return is a contract violation — never treat as delivery-eligible.
	if owner == nil {
		slog.ErrorContext(ctx, "delivery trigger: owner repo returned nil without error", "owner_id", ownerID, "clinic_id", clinicID)
		return false, apperrors.WrapNotFound("owner", fmt.Sprintf("%d", ownerID))
	}

	already, err := s.alreadyFiredToday(ctx, clinicID, ownerID, triggerType, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: alreadyFiredToday error", "owner_id", ownerID, "trigger", triggerType, "error", err)
		return false, err
	}
	if already {
		return false, nil
	}

	// Q23: 優先順位による配信抑制チェック
	suppressed, err := s.applySuppression(ctx, clinicID, ownerID, triggerType, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: applySuppression error", "owner_id", ownerID, "trigger", triggerType, "error", err)
		return false, err
	}
	if suppressed {
		suppressionReason := fmt.Sprintf("owner_id=%d already has higher-priority trigger on %s", ownerID, asOf.Format(time.DateOnly))
		suppressedLog := &model.LstepDeliveryTriggerLog{
			OwnerID:              ownerID,
			ClinicID:             clinicID,
			TriggerType:          triggerType,
			ScheduledAt:          asOf,
			Status:               model.TriggerStatusScheduled,
			SuppressedByPriority: true,
			SuppressionReason:    &suppressionReason,
		}
		// LSA-15: claim day slot under lock so concurrent workers cannot double-insert.
		created, createErr := s.triggerLogRepo.CreateIfAbsentToday(ctx, suppressedLog)
		if createErr != nil {
			slog.ErrorContext(ctx, "delivery trigger: failed to create suppressed log", "owner_id", ownerID, "trigger", triggerType, "error", createErr)
			return false, apperrors.Wrap(createErr, "failed to create suppressed trigger log")
		}
		_ = created // already-exists is an idempotent no-op for suppressed path
		return false, nil
	}

	excluded, reason, err := s.checkExclusion(ctx, clinicID, ownerID, owner)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: checkExclusion error", "owner_id", ownerID, "error", err)
		return false, err
	}

	logID, claimed, err := s.recordTrigger(ctx, clinicID, ownerID, triggerType, asOf)
	if err != nil {
		slog.ErrorContext(ctx, "delivery trigger: recordTrigger error", "owner_id", ownerID, "error", err)
		return false, err
	}
	if !claimed {
		// Concurrent worker already created today's log for this owner/type.
		return false, nil
	}

	if excluded {
		// LSA-12 / DEC-35: excluded status 更新失敗は silent success にしない
		if updateErr := s.triggerLogRepo.UpdateStatus(ctx, clinicID, logID, model.TriggerStatusExcluded, nil, &reason); updateErr != nil {
			slog.ErrorContext(ctx, "failed to record trigger log excluded status", "log_id", logID, "error", updateErr)
			return false, apperrors.Wrap(updateErr, "failed to update trigger log status to excluded")
		}
		return false, nil
	}

	// EMR-308: 送信直前の candidate 再検証。tentative リマインダーは candidate 抽出と
	// 送信の間に appointment が取消・完了・削除・昇格・日付変更され得るため、
	// claim 済みスロットを excluded に記録して外部送信を抑止する。
	// Codex: 再検証〜送信は tentative mutation と同じ advisory lock 内で行う —
	// send コールバックを hook へ渡し、hook の直列化境界内で送信させる。
	locker := deliveryOwnerLockerOf(s.tagCacheRepo)
	if locker == nil {
		markErr := markDeliveryKnownUnsent(ctx, s.triggerLogRepo, clinicID, logID)
		return false, errors.Join(errDeliveryOwnerLockUnavailable, markErr)
	}
	freshOwnerRepo := unwrapDeliveryOwnerRepo(s.ownerRepo)
	freshTagRepo := unwrapDeliveryTagRepo(s.tagCacheRepo)

	sendStarted := false
	var sendErr error
	lockExcluded := false
	lockExcludedReason := ""
	send := func(callCtx context.Context, lineUserID string) error {
		sendStarted = true
		sendErr = s.addDeliveryTag(callCtx, client, lineUserID, tagName)
		return sendErr
	}
	lockErr := locker.WithDeliveryOwnerLock(ctx, clinicID, ownerID, func(syncCtx context.Context) error {
		freshOwner, err := freshOwnerRepo.FindByID(syncCtx, clinicID, ownerID)
		if err != nil {
			return apperrors.Wrap(err, "failed to re-read owner under delivery lock")
		}
		if freshOwner == nil {
			return apperrors.WrapNotFound("owner", fmt.Sprintf("%d", ownerID))
		}
		if s.settingsSvc != nil {
			enabled, err := s.settingsSvc.IsSyncEnabled(syncCtx, clinicID)
			if err != nil {
				return apperrors.Wrap(err, "failed to re-check lstep sync under delivery lock")
			}
			if !enabled {
				lockExcluded = true
				lockExcludedReason = "lstep_sync_disabled"
				return nil
			}
		}
		lockExcludedNow, reason, err := s.checkExclusionWithTags(syncCtx, freshTagRepo, clinicID, ownerID, freshOwner)
		if err != nil {
			return err
		}
		if lockExcludedNow {
			lockExcluded = true
			lockExcludedReason = reason
			return nil
		}
		if revalidate != nil {
			stillValid, staleReason, revalidateErr := revalidate(syncCtx, ownerID, func(callCtx context.Context) error {
				return send(callCtx, *freshOwner.LineUserID)
			})
			if revalidateErr != nil {
				return revalidateErr
			}
			if !stillValid {
				lockExcluded = true
				lockExcludedReason = staleReason
				return nil
			}
			if !sendStarted {
				return apperrors.WrapInternalServerError("revalidation hook returned valid without invoking send")
			}
			return sendErr
		}
		return send(syncCtx, *freshOwner.LineUserID)
	})

	if sendStarted {
		outCtx, cancel := detachedDeliveryOutcomeContext(ctx)
		outcomeErr := s.recordDeliveryOutcome(outCtx, clinicID, logID, tagName, sendErr)
		cancel()
		if sendErr != nil || lockErr != nil || outcomeErr != nil {
			return false, errors.Join(lockErr, outcomeErr)
		}
		return true, nil
	}
	if lockExcluded {
		outCtx, cancel := detachedDeliveryOutcomeContext(ctx)
		updateErr := s.triggerLogRepo.UpdateStatus(outCtx, clinicID, logID, model.TriggerStatusExcluded, nil, &lockExcludedReason)
		cancel()
		if updateErr != nil {
			slog.ErrorContext(ctx, "failed to record trigger log excluded status", "log_id", logID, "error", updateErr)
		}
		return false, errors.Join(lockErr, updateErr)
	}
	markErr := markDeliveryKnownUnsent(ctx, s.triggerLogRepo, clinicID, logID)
	if markErr != nil {
		slog.ErrorContext(ctx, "failed to record trigger log known-unsent status", "log_id", logID, "error", markErr)
	}
	return false, errors.Join(lockErr, markErr)
}

func ownersByID(owners []*model.Owner) map[uint64]*model.Owner {
	byID := make(map[uint64]*model.Owner, len(owners))
	for _, o := range owners {
		if o == nil {
			continue
		}
		byID[o.ID] = o
	}
	return byID
}

// ---- public trigger methods ----
