package lstep

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

// healthPrevention page bulk loaders (G2F-02). Concrete medicalrecord repos implement these;
// consumer interfaces stay narrow and batch type-asserts at the page boundary.
type healthPreventionCheckupPageLoader interface {
	FindByOwnerIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]model.Checkup, error)
}

type healthPreventionVaccinationPageLoader interface {
	FindByOwnerIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]model.Vaccination, error)
}

type healthPreventionVisitSummaryPageLoader interface {
	FindOwnerVisitSummariesByOwnerIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64]*medicalrecord.OwnerVisitSummary, error)
}

// healthPreventionPageInputs holds clinic-scoped bulk inputs for one owner page.
type healthPreventionPageInputs struct {
	checkupsByOwner     map[uint64][]model.Checkup
	vaccinationsByOwner map[uint64][]model.Vaccination
	visitSummaryByOwner map[uint64]*medicalrecord.OwnerVisitSummary
	// loaded* is true when the corresponding bulk query ran (missing owner key means empty).
	// false keeps the per-owner fallback path for repos without bulk methods.
	loadedCheckups     bool
	loadedVaccinations bool
	loadedVisitSummary bool
}

// cachedTagMappings は tag code mappings を clinic 単位で 1 回取得しキャッシュする
// （BE-refactor.md E-7）。取得失敗時は warn ログを出し nil を返す（呼出元は per-owner
// フォールバックとして扱う）。label はログ文言の対象名（例: "filaria mappings"）。
func (s *lstepTagSyncService) cachedTagMappings(ctx context.Context, clinicID uint64, tagName, label string) []*model.LstepTagCodeMapping {
	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, tagName)
	if err != nil {
		slog.ErrorContext(ctx, "health-prevention batch: failed to cache "+label, "clinic_id", clinicID, "error", err)
		return nil // fallback to per-owner fetch
	}
	return mappings
}

func (s *lstepTagSyncService) SyncHealthPreventionTagsForClinic(ctx context.Context, clinicID uint64) (int, []error) {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return 0, []error{apperrors.Wrap(err, "failed to check lstep sync enabled")}
	} else if skip {
		return 0, nil
	}

	// PERF-FOLLOWUP-02: 無制限全件取得を避けるため、先頭ページのみ先に取得する。
	// 取得失敗時はマッピングキャッシュ（tagCodeRepo）に触れず即座に中断する（従来挙動を維持）。
	firstPage, err := s.ownerRepo.FindAllWithLineUserIDCursor(ctx, clinicID, 0, lstepBatchPageSize)
	if err != nil {
		slog.ErrorContext(ctx, "health-prevention batch: failed to find owners", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners with line user id")}
	}

	// Stage 1: tag code mappings を clinic 単位で 1 回キャッシュ。
	cachedMappings := s.cachedTagMappings(ctx, clinicID, HlthHealthcheckDoneTag, "mappings")

	// PERF-M2: Filaria / FleaTick / FoodPurchase の tag code mappings も clinic 単位で 1 回キャッシュ。
	cachedFilariaMappings := s.cachedTagMappings(ctx, clinicID, PrevFilariaTag, "filaria mappings")
	cachedFleaTickMappings := s.cachedTagMappings(ctx, clinicID, PrevFleaTickTag, "flea tick mappings")
	cachedFoodMappings := s.cachedTagMappings(ctx, clinicID, LtvFoodPurchaseTag, "food purchase mappings")

	// PERF-1: HealthPreventionThresholds を clinic 単位で 1 回だけ取得し、ループ内で再利用する。
	thresholds, tErr := s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
	if tErr != nil {
		slog.ErrorContext(ctx, "health-prevention batch: failed to get thresholds", "clinic_id", clinicID, "error", tErr)
		return 0, []error{apperrors.Wrap(tErr, "failed to get health prevention thresholds")}
	}

	var errs []error
	count := 0
	syncOwners := func(owners []model.Owner, pageInputs healthPreventionPageInputs) {
		for i := range owners {
			ownerID := owners[i].ID
			ownerCheckups := pageOwnerCheckups(pageInputs, ownerID)
			ownerVaccinations := pageOwnerVaccinations(pageInputs, ownerID)
			ownerVisitSummary := pageOwnerVisitSummary(pageInputs, ownerID)

			syncFns := []struct {
				name string
				fn   func() error
			}{
				{"SyncHealthcheckTags", func() error {
					return s.syncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, cachedMappings, &thresholds, ownerCheckups)
				}},
				{"SyncAnnual4CheckupTag", func() error {
					return s.syncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, cachedMappings, &thresholds, ownerCheckups, ownerVisitSummary)
				}},
				{"SyncVaccineDeadlineTag", func() error {
					return s.syncVaccineDeadlineTagWithInputs(ctx, clinicID, ownerID, thresholds, ownerVaccinations)
				}},
				{"SyncFilariaTag", func() error {
					return s.SyncFilariaTagWithMappings(ctx, clinicID, ownerID, cachedFilariaMappings, &thresholds)
				}},
				{"SyncFleaTickTag", func() error {
					return s.SyncFleaTickTagWithMappings(ctx, clinicID, ownerID, cachedFleaTickMappings, &thresholds)
				}},
				{"SyncFoodPurchaseTag", func() error {
					return s.SyncFoodPurchaseTagWithMappings(ctx, clinicID, ownerID, cachedFoodMappings, &thresholds)
				}},
			}
			ownerFailed := false
			for _, sf := range syncFns {
				if syncErr := sf.fn(); syncErr != nil {
					slog.ErrorContext(ctx, "health-prevention batch: sync failed",
						"clinic_id", clinicID, "owner_id", ownerID, "method", sf.name, "error", syncErr)
					errs = append(errs, apperrors.Wrap(syncErr, sf.name))
					ownerFailed = true
				}
			}
			if !ownerFailed {
				count++
			}
		}
	}

	// PERF-FOLLOWUP-02: カーソルページネーションでオーナーを取得しながら処理する。
	// 直前のページが pageSize ちょうどの場合のみ次ページを取得する（最後の空ページ取得を 1 回に抑える）。
	// G2F-02: 各ページで checkup / vaccination / visit-summary を clinic-scoped bulk 取得する。
	page := firstPage
	afterID := uint64(0)
	for len(page) > 0 {
		pageInputs, pageLoadErrs := s.loadHealthPreventionPageInputs(ctx, clinicID, page)
		errs = append(errs, pageLoadErrs...)
		syncOwners(page, pageInputs)
		afterID = page[len(page)-1].ID
		if len(page) < lstepBatchPageSize {
			break
		}
		var pageErr error
		page, pageErr = s.ownerRepo.FindAllWithLineUserIDCursor(ctx, clinicID, afterID, lstepBatchPageSize)
		if pageErr != nil {
			slog.ErrorContext(ctx, "health-prevention batch: failed to find owners (next page)",
				"clinic_id", clinicID, "after_id", afterID, "error", pageErr)
			errs = append(errs, apperrors.Wrap(pageErr, "failed to find owners with line user id"))
			break
		}
	}
	return count, errs
}

// loadHealthPreventionPageInputs bulk-loads child history for one owner page.
// Missing bulk loaders fall back to per-owner queries inside the sync methods.
// A bulk-load error is recorded and that category falls back (partial failure accounting).
func (s *lstepTagSyncService) loadHealthPreventionPageInputs(
	ctx context.Context,
	clinicID uint64,
	owners []model.Owner,
) (healthPreventionPageInputs, []error) {
	var inputs healthPreventionPageInputs
	var errs []error
	if len(owners) == 0 {
		return inputs, nil
	}
	ownerIDs := make([]uint64, len(owners))
	for i := range owners {
		ownerIDs[i] = owners[i].ID
	}

	if loader, ok := any(s.checkupRepo).(healthPreventionCheckupPageLoader); ok {
		byOwner, err := loader.FindByOwnerIDs(ctx, clinicID, ownerIDs)
		if err != nil {
			slog.ErrorContext(ctx, "health-prevention batch: bulk checkup load failed",
				"clinic_id", clinicID, "owner_count", len(ownerIDs), "error", err)
			errs = append(errs, apperrors.Wrap(err, "failed to bulk-load checkups"))
		} else {
			inputs.checkupsByOwner = coalesceOwnerMap(byOwner)
			inputs.loadedCheckups = true
		}
	}

	if loader, ok := any(s.vacRepo).(healthPreventionVaccinationPageLoader); ok {
		byOwner, err := loader.FindByOwnerIDs(ctx, clinicID, ownerIDs)
		if err != nil {
			slog.ErrorContext(ctx, "health-prevention batch: bulk vaccination load failed",
				"clinic_id", clinicID, "owner_count", len(ownerIDs), "error", err)
			errs = append(errs, apperrors.Wrap(err, "failed to bulk-load vaccinations"))
		} else {
			inputs.vaccinationsByOwner = coalesceOwnerMap(byOwner)
			inputs.loadedVaccinations = true
		}
	}

	if loader, ok := any(s.medRecordRepo).(healthPreventionVisitSummaryPageLoader); ok {
		byOwner, err := loader.FindOwnerVisitSummariesByOwnerIDs(ctx, clinicID, ownerIDs)
		if err != nil {
			slog.ErrorContext(ctx, "health-prevention batch: bulk visit-summary load failed",
				"clinic_id", clinicID, "owner_count", len(ownerIDs), "error", err)
			errs = append(errs, apperrors.Wrap(err, "failed to bulk-load visit summaries"))
		} else {
			inputs.visitSummaryByOwner = coalesceOwnerMap(byOwner)
			inputs.loadedVisitSummary = true
		}
	}

	return inputs, errs
}

func coalesceOwnerMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return map[K]V{}
	}
	return m
}

func pageOwnerCheckups(inputs healthPreventionPageInputs, ownerID uint64) *[]model.Checkup {
	if !inputs.loadedCheckups {
		return nil
	}
	items := inputs.checkupsByOwner[ownerID]
	if items == nil {
		items = []model.Checkup{}
	}
	return &items
}

func pageOwnerVaccinations(inputs healthPreventionPageInputs, ownerID uint64) *[]model.Vaccination {
	if !inputs.loadedVaccinations {
		return nil
	}
	items := inputs.vaccinationsByOwner[ownerID]
	if items == nil {
		items = []model.Vaccination{}
	}
	return &items
}

func pageOwnerVisitSummary(inputs healthPreventionPageInputs, ownerID uint64) **medicalrecord.OwnerVisitSummary {
	if !inputs.loadedVisitSummary {
		return nil
	}
	summary := inputs.visitSummaryByOwner[ownerID]
	return &summary
}
