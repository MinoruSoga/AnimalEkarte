package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncOwnerAnimalClassificationTags は飼い主の動物分類タグを同期する（BE-005）。
func (s *lstepTagSyncService) SyncOwnerAnimalClassificationTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for animal classification tags", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}

	pets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find living pets for classification tags", "error", err)
		return apperrors.Wrap(err, "failed to find living pets")
	}

	var hasDog, hasCat bool
	for i := range pets {
		p := &pets[i]
		if p.AnimalSpecies == nil {
			continue
		}
		if strings.Contains(p.AnimalSpecies.Name, "犬") {
			hasDog = true
		}
		if strings.Contains(p.AnimalSpecies.Name, "猫") {
			hasCat = true
		}
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	lineUserID := *owner.LineUserID

	var newTag string
	switch {
	case hasDog && hasCat:
		newTag = "has_both"
	case hasDog:
		newTag = "has_dog"
	case hasCat:
		newTag = "has_cat"
	}

	// 旧分類タグを全削除してから新タグを付与
	apiFailed := false
	for _, old := range []string{"has_dog", "has_cat", "has_both"} {
		if old == newTag {
			continue
		}
		if delErr := client.RemoveTag(ctx, lineUserID, old); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove old classification tag", "error", delErr, "tag", old)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, old)
		}
	}

	if newTag == "" {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}

	if addErr := client.AddTag(ctx, lineUserID, newTag); addErr != nil {
		slog.ErrorContext(ctx, "failed to add classification tag", "error", addErr, "tag", newTag)
		s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
		return apperrors.Wrap(addErr, fmt.Sprintf("failed to add classification tag %s", newTag))
	}
	if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, newTag, "auto", ""); cacheErr != nil {
		slog.ErrorContext(ctx, "failed to upsert classification tag cache", "error", cacheErr, "tag", newTag)
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncPetBasicInfoTags は全生存ペットの基本情報タグを同期する（BE-005）。
func (s *lstepTagSyncService) SyncPetBasicInfoTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for pet basic info tags", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
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

	lineUserID := *owner.LineUserID

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
		if delErr := client.RemoveTag(ctx, lineUserID, old); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove stale pet basic info tag", "error", delErr, "tag", old)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, old)
		}
	}

	// 新規タグを追加
	for _, tag := range newTags {
		if _, exists := oldSet[tag]; exists {
			continue
		}
		if addErr := client.AddTag(ctx, lineUserID, tag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add pet basic info tag", "error", addErr, "tag", tag)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, fmt.Sprintf("failed to add pet basic info tag %s", tag))
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, tag, "auto", ""); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert pet basic info tag cache", "error", cacheErr, "tag", tag)
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
			if strings.Contains(p.AnimalSpecies.Name, "犬") {
				fallback = "breed_mix_dog"
			} else if strings.Contains(p.AnimalSpecies.Name, "猫") {
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

// petSpeciesFlags は生存ペットのスライスから犬/猫の有無を返す純粋関数（FEAT-377）。
func petSpeciesFlags(pets []model.Pet) (hasDog, hasCat bool) {
	for i := range pets {
		p := &pets[i]
		if p.AnimalSpecies == nil {
			continue
		}
		name := p.AnimalSpecies.Name
		if strings.Contains(name, "犬") {
			hasDog = true
		}
		if strings.Contains(name, "猫") {
			hasCat = true
		}
	}
	return hasDog, hasCat
}

// hasSeniorPet は生存ペットのスライスに now 時点で 7 歳以上の犬猫が存在するかを返す純粋関数（FEAT-377）。
func hasSeniorPet(pets []model.Pet, now time.Time) bool {
	const seniorAgeYears = 7
	for i := range pets {
		p := &pets[i]
		if p.BirthDate == nil || p.AnimalSpecies == nil {
			continue
		}
		name := p.AnimalSpecies.Name
		if !strings.Contains(name, "犬") && !strings.Contains(name, "猫") {
			continue
		}
		ageYears := now.Year() - p.BirthDate.Year()
		if now.Before(p.BirthDate.AddDate(ageYears, 0, 0)) {
			ageYears--
		}
		if ageYears >= seniorAgeYears {
			return true
		}
	}
	return false
}

// SyncPetSpeciesTags は飼い主の生存ペット種別タグ（PET_犬あり / PET_猫あり）を同期する（FEAT-377）。
func (s *lstepTagSyncService) SyncPetSpeciesTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for PET species tags sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	pets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find living pets for PET species tags", "error", err)
		return apperrors.Wrap(err, "failed to find living pets")
	}

	hasDog, hasCat := petSpeciesFlags(pets)

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	const (
		tagDog = "PET_犬あり"
		tagCat = "PET_猫あり"
	)
	apiFailed := false
	for _, entry := range []struct {
		tag  string
		have bool
	}{
		{tagDog, hasDog},
		{tagCat, hasCat},
	} {
		if entry.have {
			if addErr := client.AddTag(ctx, lineUserID, entry.tag); addErr != nil {
				slog.ErrorContext(ctx, "failed to add PET species tag", "error", addErr, "tag", entry.tag)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, entry.tag, "auto", ""); cacheErr != nil {
				slog.ErrorContext(ctx, "failed to upsert PET species tag cache", "error", cacheErr)
			}
		} else {
			if delErr := client.RemoveTag(ctx, lineUserID, entry.tag); delErr != nil {
				slog.ErrorContext(ctx, "failed to remove PET species tag", "error", delErr, "tag", entry.tag)
				s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
				apiFailed = true
				continue
			}
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, entry.tag)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncSeniorTag は飼い主の生存ペットに 7 歳以上の犬猫がいる場合 PET_シニア対象 タグを付与する（FEAT-377）。
func (s *lstepTagSyncService) SyncSeniorTag(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}
	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for senior tag sync", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	pets, err := s.petRepo.FindLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find living pets for senior tag", "error", err)
		return apperrors.Wrap(err, "failed to find living pets")
	}

	isSenior := hasSeniorPet(pets, time.Now())

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	const seniorTag = "PET_シニア対象"
	apiFailed := false
	if isSenior {
		if addErr := client.AddTag(ctx, lineUserID, seniorTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add senior tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add senior tag")
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, seniorTag, "auto", ""); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert senior tag cache", "error", cacheErr)
		}
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, seniorTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove senior tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, seniorTag)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncExclusionTags は配信停止条件に基づき EXCL_配信停止 タグを同期する（FEAT-377/FEAT-381）。
// 配信停止条件: LstepOptOut=true / DeliveryExcluded=true / IsTransferred=true /
//
//	会員ステータスが deceased / transferred / 全ペット死亡。
//
// 注: checkOptOut は呼ばない（このメソッド自体が opt-out 判定の実装）。
func (s *lstepTagSyncService) SyncExclusionTags(ctx context.Context, clinicID, ownerID uint64) error {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	owner, err := s.ownerRepo.FindByID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find owner for exclusion tags", "error", err)
		return apperrors.Wrap(err, "failed to find owner")
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	// 全ペット死亡判定
	totalPets, err := s.petRepo.CountByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count pets for exclusion tags", "error", err)
		return apperrors.Wrap(err, "failed to count pets")
	}
	livingPets, err := s.petRepo.CountLivingByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to count living pets for exclusion tags", "error", err)
		return apperrors.Wrap(err, "failed to count living pets")
	}
	allPetsDead := totalPets > 0 && livingPets == 0

	shouldExclude := owner.LstepOptOut ||
		owner.DeliveryExcluded ||
		owner.IsTransferred ||
		owner.MembershipType == model.MembershipTypeDeceased ||
		owner.MembershipType == model.MembershipTypeTransferred ||
		allPetsDead

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if shouldExclude {
		if addErr := client.AddTag(ctx, lineUserID, exclTagDeliveryStop); addErr != nil {
			slog.ErrorContext(ctx, "failed to add EXCL exclusion tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add EXCL exclusion tag")
		}
		if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, exclTagDeliveryStop, "auto", ""); cacheErr != nil {
			slog.ErrorContext(ctx, "failed to upsert EXCL exclusion tag cache", "error", cacheErr)
		}
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, exclTagDeliveryStop); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove EXCL exclusion tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, exclTagDeliveryStop)
		}
	}

	// FEAT-381-2: EXCL_配信注意 タグは delivery_caution フラグのみで独立して制御する。
	if owner.DeliveryCaution {
		if addErr := client.AddTag(ctx, lineUserID, exclTagDeliveryCaution); addErr != nil {
			slog.ErrorContext(ctx, "failed to add EXCL caution tag", "error", addErr)
			apiFailed = true
		} else {
			if cacheErr := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, exclTagDeliveryCaution, "auto", ""); cacheErr != nil {
				slog.ErrorContext(ctx, "failed to upsert EXCL caution tag cache", "error", cacheErr)
			}
		}
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, exclTagDeliveryCaution); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove EXCL caution tag", "error", delErr)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, exclTagDeliveryCaution)
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
