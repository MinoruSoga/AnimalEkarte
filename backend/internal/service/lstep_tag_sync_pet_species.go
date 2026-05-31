package service

import (
	"context"
	"log/slog"
	"strings"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

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
