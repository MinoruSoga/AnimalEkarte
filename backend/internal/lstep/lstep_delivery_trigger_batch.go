package lstep

import (
	"context"
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

func (s *lstepDeliveryTriggerService) runBatch(
	ctx context.Context,
	clinicID uint64,
	ownerIDs []uint64,
	triggerType string,
	tagName string,
	asOf time.Time,
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
			fired, loopErr := s.processSingleOwner(ctx, client, clinicID, ownerID, triggerType, tagName, asOf)
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
			if c.dayLogsByOwner[ownerID][i].TriggerType == triggerType {
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

// processSingleOwner は1飼い主分のトリガー処理を行い、配信実行されたか否かを返す。
func (s *lstepDeliveryTriggerService) processSingleOwner(
	ctx context.Context,
	client lstep.Client,
	clinicID, ownerID uint64,
	triggerType, tagName string,
	asOf time.Time,
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

	if err := s.applyTagAndLog(ctx, clinicID, client, *owner.LineUserID, tagName, logID); err != nil {
		return false, err
	}
	return true, nil
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
