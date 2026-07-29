package lstep

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncVaccineTag はワクチン接種記録からタグを同期する（BE-003）。
// 接種種別（dog/cat）とラビーズを date 付きタグとして付与する。
func (s *lstepTagSyncService) SyncVaccineTag(ctx context.Context, clinicID, ownerID, vaccinationID uint64) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "vaccine")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	vac, err := s.vacRepo.FindByID(ctx, clinicID, vaccinationID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find vaccination for tag sync", "error", err)
		return apperrors.Wrap(err, "failed to find vaccination")
	}

	tags := vaccineTagNames(vac)
	if len(tags) == 0 {
		return nil
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	// 同一カテゴリの古い日付タグを解除してから新タグを付与（ISSUE-006）
	newTagSet := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		newTagSet[t] = struct{}{}
	}
	apiFailed, err := s.removeStaleTagsByPrefixes(ctx, client, clinicID, ownerID, lineUserID,
		[]string{"vaccine_dog_", "vaccine_cat_", "vaccine_rabies_"}, newTagSet)
	if err != nil {
		// LSA-11: cache-read failure — zero category API calls, surface the error.
		return err
	}

	for _, tag := range tags {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, tag, "vaccine", "", true); err != nil {
			return err
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// vaccineTagNames はワクチン接種記録から付与すべきタグ名一覧を返す。
func vaccineTagNames(vac *model.Vaccination) []string {
	if vac.Vaccine == nil {
		return nil
	}
	date := vac.Date.Format(time.DateOnly)
	var tags []string

	species := vac.Vaccine.Species
	if species != nil {
		switch *species {
		case model.VaccineSpeciesDog:
			tags = append(tags, "vaccine_dog_"+date)
		case model.VaccineSpeciesCat:
			tags = append(tags, "vaccine_cat_"+date)
		case model.VaccineSpeciesBoth:
			tags = append(tags, "vaccine_dog_"+date, "vaccine_cat_"+date)
		}
	}

	if isRabiesVaccine(vac.Vaccine.Name) {
		tags = append(tags, "vaccine_rabies_"+date)
	}
	return tags
}

func isRabiesVaccine(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "rabies") || strings.Contains(name, "狂犬病")
}

// ResyncOwnerVaccineTags は飼い主の生存ワクチン記録から vaccine_* タグを再構築する（ISSUE-004）。
// 接種記録の更新・削除後に呼び出すこと。種別ごとに最新の接種日のみタグを保持する。
// 記録が0件の場合は全 vaccine_* タグを解除する。
func (s *lstepTagSyncService) ResyncOwnerVaccineTags(ctx context.Context, clinicID, ownerID uint64) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "vaccine resync")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for vaccine resync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	vaccinations, err := s.vacRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find vaccinations for resync", "error", err)
		return apperrors.Wrap(err, "failed to find vaccinations")
	}

	newTagSet := buildLatestVaccineTagSet(vaccinations)

	apiFailed, err := s.removeStaleTagsByPrefixes(ctx, client, clinicID, ownerID, lineUserID,
		[]string{"vaccine_dog_", "vaccine_cat_", "vaccine_rabies_"}, newTagSet)
	if err != nil {
		// LSA-11: cache-read failure — zero category API calls, surface the error.
		return err
	}

	for tag := range newTagSet {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, tag, "vaccine resync", "", true); err != nil {
			return err
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// buildLatestVaccineTagSet は接種記録一覧から「種別ごとの最新接種日」のタグ集合を返す。
// 同一種別に複数記録がある場合は最も新しい date のタグのみを採用する。
func buildLatestVaccineTagSet(vaccinations []model.Vaccination) map[string]struct{} {
	latestByPrefix := make(map[string]time.Time)
	updateLatest := func(prefix string, date time.Time) {
		if cur, ok := latestByPrefix[prefix]; !ok || date.After(cur) {
			latestByPrefix[prefix] = date
		}
	}
	for i := range vaccinations {
		v := &vaccinations[i]
		if v.Vaccine == nil {
			continue
		}
		species := v.Vaccine.Species
		if species != nil {
			switch *species {
			case model.VaccineSpeciesDog:
				updateLatest("vaccine_dog_", v.Date)
			case model.VaccineSpeciesCat:
				updateLatest("vaccine_cat_", v.Date)
			case model.VaccineSpeciesBoth:
				updateLatest("vaccine_dog_", v.Date)
				updateLatest("vaccine_cat_", v.Date)
			}
		}
		if isRabiesVaccine(v.Vaccine.Name) {
			updateLatest("vaccine_rabies_", v.Date)
		}
	}
	tagSet := make(map[string]struct{}, len(latestByPrefix))
	for prefix, date := range latestByPrefix {
		tagSet[prefix+date.Format(time.DateOnly)] = struct{}{}
	}
	return tagSet
}
