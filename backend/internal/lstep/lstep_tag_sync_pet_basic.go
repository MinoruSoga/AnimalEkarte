package lstep

import (
	"context"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncPetBasicInfoTags は全生存ペットの基本情報タグを同期する（BE-005）。
func (s *lstepTagSyncService) SyncPetBasicInfoTags(ctx context.Context, clinicID, ownerID uint64) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "pet basic info")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	pets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find living pets for basic info tags", "error", err)
		return apperrors.Wrap(err, "failed to find living pets")
	}

	newTags := buildPetBasicInfoTags(pets)

	cachedTags, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag cache for pet basic info sync", "error", err)
		return apperrors.Wrap(err, "failed to find tag cache")
	}

	// C1 プレフィックスを DB から一括ロード（ループ外で 1 回のみ）
	var c1Prefixes []*model.LstepAutoManagedPrefix
	if s.tagConfigRepo != nil {
		loaded, loadErr := s.tagConfigRepo.FindAllAutoManagedPrefixes(ctx)
		if loadErr != nil {
			slog.ErrorContext(ctx, "failed to load auto managed prefixes for pet basic info sync", "error", loadErr)
			return apperrors.Wrap(loadErr, "failed to load auto managed prefixes")
		}
		c1Prefixes = loaded
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	oldSet := make(map[string]struct{})
	for _, c := range cachedTags {
		if isPetBasicInfoTagWithPrefixes(c.TagName, c1Prefixes) {
			oldSet[c.TagName] = struct{}{}
		}
	}

	newSet := make(map[string]struct{}, len(newTags))
	for _, t := range newTags {
		newSet[t] = struct{}{}
	}

	// 不要になったタグを削除
	apiFailed := false
	for old := range oldSet {
		if _, keep := newSet[old]; keep {
			continue
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, old, "pet basic info", "", false); err != nil {
			apiFailed = true
		}
	}

	// 新規タグを追加
	for _, tag := range newTags {
		if _, exists := oldSet[tag]; exists {
			continue
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, tag, "pet basic info", "", true); err != nil {
			return err
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// buildPetBasicInfoTags は生存ペット一覧から基本情報タグ一覧を生成する。
func buildPetBasicInfoTags(pets []model.Pet) []string {
	tagSet := make(map[string]struct{})
	var hasNeutered, hasIntact bool

	for i := range pets {
		p := &pets[i]
		fallback := "breed_mix_other"
		if p.AnimalSpecies != nil {
			if isDogSpeciesName(p.AnimalSpecies.Name) {
				fallback = "breed_mix_dog"
			} else if isCatSpeciesName(p.AnimalSpecies.Name) {
				fallback = "breed_mix_cat"
			}
		}
		tagSet[lstep.BreedTagName(p.Breed, fallback)] = struct{}{}

		switch p.Gender {
		case model.PetGenderMale:
			tagSet["sex_male"] = struct{}{}
		case model.PetGenderFemale:
			tagSet["sex_female"] = struct{}{}
		default:
			tagSet["sex_unknown"] = struct{}{}
		}

		if p.BirthDate != nil {
			tagSet["pet_birthday_"+p.BirthDate.Format("01-02")] = struct{}{}
			tagSet["birth_year_"+p.BirthDate.Format("2006")] = struct{}{}
		}

		if p.NeuteredDate != nil {
			hasNeutered = true
		} else {
			hasIntact = true
		}
	}

	if hasNeutered {
		tagSet["spay_neutered"] = struct{}{}
	}
	if hasIntact {
		tagSet["intact"] = struct{}{}
	}

	tags := make([]string, 0, len(tagSet))
	for t := range tagSet {
		tags = append(tags, t)
	}
	return tags
}

// isPetBasicInfoTagWithPrefixes は BE-005 ペット基本情報カテゴリ (C1) のタグかを判定する（純粋関数）。
// dbPrefixes は lstep_auto_managed_prefixes テーブルから取得した C1 カテゴリのレコード。
func isPetBasicInfoTagWithPrefixes(tag string, dbPrefixes []*model.LstepAutoManagedPrefix) bool {
	for _, p := range dbPrefixes {
		if p.Category != "C1" {
			continue
		}
		if tag == p.Prefix || strings.HasPrefix(tag, p.Prefix) {
			return true
		}
	}
	return false
}
