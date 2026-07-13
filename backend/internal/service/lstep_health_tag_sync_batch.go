package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// healthPreventionBatchPageSize は PERF-FOLLOWUP-02 のカーソルページネーション 1 ページあたりの件数。
const healthPreventionBatchPageSize = 500

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
	firstPage, err := s.ownerRepo.FindAllWithLineUserIDCursor(ctx, clinicID, 0, healthPreventionBatchPageSize)
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
	syncOwners := func(owners []model.Owner) {
		for i := range owners {
			ownerID := owners[i].ID
			syncFns := []struct {
				name string
				fn   func() error
			}{
				{"SyncHealthcheckTags", func() error {
					return s.SyncHealthcheckTagsWithMappings(ctx, clinicID, ownerID, cachedMappings, &thresholds)
				}},
				{"SyncAnnual4CheckupTag", func() error {
					return s.SyncAnnual4CheckupTagWithMappings(ctx, clinicID, ownerID, cachedMappings, &thresholds)
				}},
				{"SyncVaccineDeadlineTag", func() error {
					return s.syncVaccineDeadlineTagImpl(ctx, clinicID, ownerID, thresholds)
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
	page := firstPage
	afterID := uint64(0)
	for {
		if len(page) == 0 {
			break
		}
		syncOwners(page)
		afterID = page[len(page)-1].ID
		if len(page) < healthPreventionBatchPageSize {
			break
		}
		var pageErr error
		page, pageErr = s.ownerRepo.FindAllWithLineUserIDCursor(ctx, clinicID, afterID, healthPreventionBatchPageSize)
		if pageErr != nil {
			slog.ErrorContext(ctx, "health-prevention batch: failed to find owners (next page)",
				"clinic_id", clinicID, "after_id", afterID, "error", pageErr)
			errs = append(errs, apperrors.Wrap(pageErr, "failed to find owners with line user id"))
			break
		}
	}
	return count, errs
}
