package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncHealthcheckTags は健診履歴に基づき HLTH_健診あり / HLTH_健診未受診 を同期する（FEAT-379）。
func (s *lstepTagSyncService) SyncHealthcheckTags(ctx context.Context, clinicID, ownerID uint64) error {
	if s.tagCodeRepo == nil {
		return nil
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, HlthHealthcheckDoneTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for healthcheck", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	checkupCodes := extractTagCodes(mappings, model.CodeTypeCheckupType)
	if len(checkupCodes) == 0 {
		return nil
	}

	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for healthcheck tags", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	thresholds, err := s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get health prevention thresholds for healthcheck tags", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to get health prevention thresholds")
	}
	since := time.Now().AddDate(0, 0, -thresholds.LookbackDays)
	checkups, err := s.checkupRepo.FindByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find checkups for healthcheck tags", "error", err)
		return apperrors.Wrap(err, "failed to find checkups")
	}

	codeSet := strSet(checkupCodes)
	hasHealthcheck := false
	var lastCheckupDate time.Time
	for i := range checkups {
		if checkups[i].Date.Before(since) {
			continue
		}
		if checkups[i].CheckupType == nil {
			continue
		}
		if _, ok := codeSet[checkups[i].CheckupType.Name]; ok {
			hasHealthcheck = true
			if checkups[i].Date.After(lastCheckupDate) {
				lastCheckupDate = checkups[i].Date
			}
		}
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if hasHealthcheck {
		if addErr := client.AddTag(ctx, lineUserID, HlthHealthcheckDoneTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add healthcheck done tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add healthcheck done tag")
		}
		if err := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, HlthHealthcheckDoneTag, "auto", fmt.Sprintf("最終健診: %s", lastCheckupDate.Format("2006-01-02"))); err != nil {
			slog.WarnContext(ctx, "failed to upsert tag cache (non-fatal)", "tag", HlthHealthcheckDoneTag, "owner_id", ownerID, "error", err)
		}
		if delErr := client.RemoveTag(ctx, lineUserID, HlthHealthcheckNeverTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove healthcheck never tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			if err := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, HlthHealthcheckNeverTag); err != nil {
				slog.WarnContext(ctx, "failed to delete tag cache (non-fatal)", "tag", HlthHealthcheckNeverTag, "owner_id", ownerID, "error", err)
			}
		}
	} else {
		if addErr := client.AddTag(ctx, lineUserID, HlthHealthcheckNeverTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add healthcheck never tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add healthcheck never tag")
		}
		if err := s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, HlthHealthcheckNeverTag, "auto", ""); err != nil {
			slog.WarnContext(ctx, "failed to upsert tag cache (non-fatal)", "tag", HlthHealthcheckNeverTag, "owner_id", ownerID, "error", err)
		}
		if delErr := client.RemoveTag(ctx, lineUserID, HlthHealthcheckDoneTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove healthcheck done tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			if err := s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, HlthHealthcheckDoneTag); err != nil {
				slog.WarnContext(ctx, "failed to delete tag cache (non-fatal)", "tag", HlthHealthcheckDoneTag, "owner_id", ownerID, "error", err)
			}
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncAnnual4CheckupTag は年2回以上来院かつ健診履歴がある飼い主に HLTH_年4回候補 を付与する（FEAT-379）。
func (s *lstepTagSyncService) SyncAnnual4CheckupTag(ctx context.Context, clinicID, ownerID uint64) error {
	if s.tagCodeRepo == nil {
		return nil
	}
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return err
	} else if skip {
		return nil
	}

	mappings, err := s.tagCodeRepo.FindByClinicIDAndTagName(ctx, clinicID, HlthHealthcheckDoneTag)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find tag code mappings for annual4checkup", "error", err)
		return apperrors.Wrap(err, "failed to find tag code mappings")
	}
	checkupCodes := extractTagCodes(mappings, model.CodeTypeCheckupType)
	if len(checkupCodes) == 0 {
		return nil
	}

	optOut, owner, err := s.checkOptOut(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to check opt-out for annual4checkup tag", "error", err)
		return apperrors.Wrap(err, "failed to check opt-out")
	}
	if optOut {
		return nil
	}
	if owner.LineUserID == nil || *owner.LineUserID == "" {
		return nil
	}
	lineUserID := *owner.LineUserID

	thresholds, err := s.settingsSvc.GetHealthPreventionThresholds(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to get health prevention thresholds for annual4checkup tag", "error", err, "clinic_id", clinicID)
		return apperrors.Wrap(err, "failed to get health prevention thresholds")
	}
	since := time.Now().AddDate(0, 0, -thresholds.LookbackDays)
	checkups, err := s.checkupRepo.FindByOwnerID(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find checkups for annual4checkup tag", "error", err)
		return apperrors.Wrap(err, "failed to find checkups")
	}

	codeSet := strSet(checkupCodes)
	hasHealthcheck := false
	for i := range checkups {
		if checkups[i].Date.Before(since) {
			continue
		}
		if checkups[i].CheckupType == nil {
			continue
		}
		if _, ok := codeSet[checkups[i].CheckupType.Name]; ok {
			hasHealthcheck = true
			break
		}
	}

	visitSummary, err := s.medRecordRepo.FindOwnerVisitSummary(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find visit summary for annual4checkup tag", "error", err)
		return apperrors.Wrap(err, "failed to find visit summary")
	}

	qualified := hasHealthcheck && visitSummary.AnnualCount >= 2

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return err
	}
	if client == nil {
		return nil
	}

	apiFailed := false
	if qualified {
		if addErr := client.AddTag(ctx, lineUserID, HlthAnnual4CheckupTag); addErr != nil {
			slog.ErrorContext(ctx, "failed to add annual4checkup tag", "error", addErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			return apperrors.Wrap(addErr, "failed to add annual4checkup tag")
		}
		_ = s.tagCacheRepo.UpsertTag(ctx, clinicID, ownerID, HlthAnnual4CheckupTag, "auto", "")
	} else {
		if delErr := client.RemoveTag(ctx, lineUserID, HlthAnnual4CheckupTag); delErr != nil {
			slog.ErrorContext(ctx, "failed to remove annual4checkup tag", "error", delErr)
			s.notifyAPIFailure(ctx, client, clinicID, ownerID, lineUserID)
			apiFailed = true
		} else {
			_ = s.tagCacheRepo.DeleteTag(ctx, clinicID, ownerID, HlthAnnual4CheckupTag)
		}
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}

// SyncVaccineDeadlineTag はワクチン次回予定日が VaccineDeadlineDays 以内なら
// PREV_ワクチン期限 を付与し、それ以外は解除する（FEAT-379）。
