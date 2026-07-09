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
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "PET species")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

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
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, entry.tag, "PET species", "", entry.have); err != nil {
			apiFailed = true
			continue
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
