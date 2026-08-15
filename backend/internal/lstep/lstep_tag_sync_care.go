package lstep

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

// SyncPrescriptionTag は飼い主の全アクティブ処方を取得し、補充推奨日が最も遅い処方に基づいて
// refill_due_* タグを更新する（BE-009）。duration_days < 7 の場合は prescribed_at + 1 日を使用する。
// 最新の refill_due が現在日時を過ぎている場合は refill_due_* タグをすべて削除して終了する。
func (s *lstepTagSyncService) SyncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "prescription")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for prescription tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	prescriptions, err := s.prescriptionRepo.FindActiveByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load prescriptions for tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load prescriptions")
	}

	// 補充推奨日の計算: prescribed_at + duration_days - 7。duration_days < 7 なら prescribed_at + 1
	var latestRefillDue *time.Time
	for i := range prescriptions {
		p := &prescriptions[i]
		var refill time.Time
		if p.DurationDays < 7 {
			refill = p.PrescribedAt.AddDate(0, 0, 1)
		} else {
			refill = p.PrescribedAt.AddDate(0, 0, p.DurationDays-7)
		}
		if latestRefillDue == nil || refill.After(*latestRefillDue) {
			t := refill
			latestRefillDue = &t
		}
	}

	// 旧 refill_due_* タグをキャッシュ経由で削除
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load tag cache for prescription tag sync", "error", err)
		return apperrors.Wrap(err, "failed to load tag cache")
	}
	apiFailed := false
	for _, c := range cached {
		if strings.HasPrefix(c.TagName, tagPrefixRefillDue) {
			if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, c.TagName, "refill_due", "", false); err != nil {
				apiFailed = true
				continue
			}
		}
	}

	// 補充推奨日が未来にある場合のみ新タグを付与
	if latestRefillDue == nil || !latestRefillDue.After(time.Now()) {
		if !apiFailed {
			s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
		}
		return nil
	}
	newTag := tagPrefixRefillDue + latestRefillDue.Format(time.DateOnly)
	if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, newTag, "refill_due", "", true); err != nil {
		return err
	}
	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
