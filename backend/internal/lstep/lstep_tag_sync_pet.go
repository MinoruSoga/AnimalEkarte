package lstep

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// SyncOwnerAnimalClassificationTags は飼い主の動物分類タグを同期する（BE-005）。
func (s *lstepTagSyncService) SyncOwnerAnimalClassificationTags(ctx context.Context, clinicID, ownerID uint64) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "animal classification")
	if err != nil {
		return err
	}
	if !ok {
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
		if isDogSpeciesName(p.AnimalSpecies.Name) {
			hasDog = true
		}
		if isCatSpeciesName(p.AnimalSpecies.Name) {
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
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, old, "animal classification", "", false); err != nil {
			apiFailed = true
		}
	}

	if newTag == "" {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}

	if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, newTag, "animal classification", "", true); err != nil {
		return err
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
