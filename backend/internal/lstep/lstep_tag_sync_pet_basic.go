package lstep

import (
	"context"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/commonbasic"
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

	newTags, err := buildPetBasicInfoTags(pets)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build pet basic info tags", "error", err)
		return apperrors.Wrap(err, "failed to build pet basic info tags")
	}

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

// buildPetBasicInfoTags は生存ペット一覧から基本情報タグ一覧を生成する。品種・性別・誕生日・去勢は hydrate 済みの
// 現行 basic revision から読む。未 hydrate は unresolved の error で返し（空値や旧 pets 列への代替はしない）、同期しない。
func buildPetBasicInfoTags(pets []model.Pet) ([]string, error) {
	tagSet := make(map[string]struct{})
	var hasNeutered, hasIntact bool

	for i := range pets {
		p := &pets[i]
		basic, err := commonbasic.PetOf(p)
		if err != nil {
			return nil, err
		}
		fallback := "breed_mix_other"
		if p.AnimalSpecies != nil {
			if isDogSpeciesName(p.AnimalSpecies.Name) {
				fallback = "breed_mix_dog"
			} else if isCatSpeciesName(p.AnimalSpecies.Name) {
				fallback = "breed_mix_cat"
			}
		}
		// 品種が未設定（NULL）のペットは空の品種名の入力（= BreedTagName の fallback タグ）として扱う。
		breed := ""
		if basic.Breed != nil {
			breed = *basic.Breed
		}
		tagSet[lstep.BreedTagName(breed, fallback)] = struct{}{}

		// 性別が未設定（NULL）のペットは default 分岐（sex_unknown）になる。
		switch {
		case basic.Gender != nil && *basic.Gender == model.PetGenderMale:
			tagSet["sex_male"] = struct{}{}
		case basic.Gender != nil && *basic.Gender == model.PetGenderFemale:
			tagSet["sex_female"] = struct{}{}
		default:
			tagSet["sex_unknown"] = struct{}{}
		}

		if basic.BirthDate != nil {
			tagSet["pet_birthday_"+basic.BirthDate.Format("01-02")] = struct{}{}
			tagSet["birth_year_"+basic.BirthDate.Format("2006")] = struct{}{}
		}

		switch {
		case basic.NeuteredStatus == model.PetNeuteredStatusDone:
			hasNeutered = true
		case basic.NeuteredStatus == model.PetNeuteredStatusNotDone:
			hasIntact = true
		case basic.NeuteredDate != nil:
			hasNeutered = true
		default:
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
	return tags, nil
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
