package service

import (
	"context"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
)

// SyncSeniorTag は飼い主の生存ペットに 7 歳以上の犬猫がいる場合 PET_シニア対象 タグを付与する（FEAT-377）。
func (s *lstepTagSyncService) SyncSeniorTag(ctx context.Context, clinicID, ownerID uint64) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "senior")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

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
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, seniorTag, "senior", "", true); err != nil {
			return err
		}
	} else {
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, seniorTag, "senior", "", false); err != nil {
			apiFailed = true
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
