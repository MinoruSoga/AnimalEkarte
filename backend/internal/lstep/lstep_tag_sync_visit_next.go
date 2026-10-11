package lstep

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncNextVisitTag は次回来院タグを同期する（BE-006 / EMR-308）。
// 飼い主の次回 active appointment を参照し、古い next_visit_* タグを差し替える。
// tentative は intended_date、それ以外は start_time の JST 暦日で
// next_visit_YYYY-MM-DD を生成する。次回来院がない場合はすべての next_visit_* タグを削除する。
// appointment finder 未注入は構成不全 —— タグ除去へフォールバックすると
// 誤配線の呼び出し元（cmd/lstep-migrate 等）が既存タグを黙って剥奪するため、
// タグに触れずエラーを返す（fail-closed）。
func (s *lstepTagSyncService) SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "next visit")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for next visit tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	// EMR-308: source of truth はカルテの旧推奨日カラムではなく
	// 次回 appointment（tentative は intended_date、実予約は start_time）に移った。
	if s.nextVisitAppointmentRepo == nil {
		slog.ErrorContext(ctx, "next visit appointment repository is not configured", "clinic_id", clinicID, "owner_id", ownerID)
		return apperrors.WrapInternalServerError("next visit appointment repository is not configured")
	}
	next, err := s.nextVisitAppointmentRepo.FindNextAppointmentForOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find next appointment for next visit tag", "error", err)
		return apperrors.Wrap(err, "failed to find next appointment")
	}
	nextVisitDate := nextVisitDateFromAppointment(next)

	// 既存の next_visit_* タグを取得し、現在のタグ追加後に古いものを削除
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for next visit tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	// Add the current tag before retiring old ones. If this fails, the stale
	// cache remains a candidate for the expiry batch's next retry.
	newTag := ""
	if nextVisitDate != nil {
		newTag = tagPrefixNextVisit + nextVisitDate.In(config.JST).Format(time.DateOnly)
		alreadyCached := false
		for _, c := range cached {
			if c.TagName == newTag {
				alreadyCached = true
				break
			}
		}
		if !alreadyCached {
			if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, newTag, "next visit", "", true); err != nil {
				return err
			}
		}
	}
	var removalErrs []error
	for _, c := range cached {
		if !strings.HasPrefix(c.TagName, tagPrefixNextVisit) || c.TagName == newTag {
			continue
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, c.TagName, "next visit", "", false); err != nil {
			removalErrs = append(removalErrs, err)
		}
	}
	if len(removalErrs) == 0 {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return errors.Join(removalErrs...)

}

// nextVisitDateFromAppointment は appointment からタグ用の意味上の予定日を返す。
// tentative（予定・未確定）は intended_date、それ以外の active 予約は start_time。
// 該当なし・tentative で intended_date 未設定の場合は nil（＝ next_visit_* 除去）。
func nextVisitDateFromAppointment(appointment *model.Reservation) *time.Time {
	if appointment == nil {
		return nil
	}
	if appointment.Status == model.ReservationStatusTentative {
		return appointment.IntendedDate
	}
	return &appointment.StartTime
}
