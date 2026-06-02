package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

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
			if delCacheErr := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, seniorTag); delCacheErr != nil {
				slog.WarnContext(ctx, "failed to delete tag from cache (best-effort)", "error", delCacheErr, "owner_id", ownerID, "tag", seniorTag)
			}
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
