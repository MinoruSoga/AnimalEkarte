package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
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
			if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, old); delCacheErr != nil {
				slog.WarnContext(ctx, "failed to delete tag from cache (best-effort)", "error", delCacheErr, "owner_id", ownerID, "tag", old)
			}
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
