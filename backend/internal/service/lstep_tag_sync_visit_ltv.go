package service

import (
	"context"
	"log/slog"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// SyncLTVTopPercent は LTV 上位 20% の飼い主に LTV_上位20 タグを付与し、
// それ以外から解除する（FEAT-377）。
func (s *lstepTagSyncService) SyncLTVTopPercent(ctx context.Context, clinicID uint64) (int, []error) {
	if skip, err := s.shouldSkipSync(ctx, clinicID); err != nil {
		return 0, []error{err}
	} else if skip {
		return 0, nil
	}

	revenues, err := s.accountRepo.FindOwnersByAnnualRevenue(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to find revenues", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners by annual revenue")}
	}

	topN := 0
	if len(revenues) > 0 {
		topN = (len(revenues)*20 + 99) / 100
	}
	topOwnerIDs := make(map[uint64]struct{}, topN)
	for i := 0; i < topN; i++ {
		topOwnerIDs[revenues[i].OwnerID] = struct{}{}
	}

	owners, err := s.ownerRepo.FindAllWithLineUserID(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to find owners", "clinic_id", clinicID, "error", err)
		return 0, []error{apperrors.Wrap(err, "failed to find owners with line user id")}
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		return 0, []error{err}
	}
	if client == nil {
		return 0, nil
	}

	// N+1回避(PERF-FOLLOWUP-08): タグ解除候補(LINE連携済み・非top)の owner_id を先に収集し、
	// タグキャッシュを1クエリで一括取得する（checkup_sync_service_preview.go:49 と同型）。
	// FindAllWithLineUserID はカーソルページング未適用（全件ロード）のため、ページ単位ではなく
	// このバッチ全体で1回の FindByOwners で足りる。
	nonTopOwnerIDs := make([]uint64, 0, len(owners))
	for i := range owners {
		owner := &owners[i]
		if owner.LineUserID == nil || *owner.LineUserID == "" {
			continue
		}
		if _, isTop := topOwnerIDs[owner.ID]; isTop {
			continue
		}
		nonTopOwnerIDs = append(nonTopOwnerIDs, owner.ID)
	}
	// バッチ取得の失敗は、この関数の他の repo 取得（revenues/owners）と同じ致命度で早期 return する
	// （per-owner版と異なり、失敗した owner だけを errs に積んで続行することはできない — バッチが
	// 失敗した場合に「タグなし」として扱うと、実際は保持しているタグを誤って解除しないまま
	// success 扱いするサイレントな不整合になるため、fail-open にしない）。
	// blast radius 注意: 旧 per-owner 実装ではキャッシュ取得失敗は非top owner のみに閉じていたが
	// （top owner 分岐は per-owner cache 取得と独立していたため）、この early return は該当クリニックの
	// 処理全体（top owner への LTV_上位20 タグ付与も含む）を止める。revenues/owners 取得失敗と同じ
	// 「関数全体 fail-closed」パターンへ意図的に統一したトレードオフ — 次回 cron サイクルで自己修復される。
	// nonTopOwnerIDs が空なら呼ばない（無駄なクエリを避ける。tagCacheRepo は nil のまま
	// map lookup の zero value（空スライス）にフォールバックする）。
	var tagCacheByOwner map[uint64][]*model.LstepTagCache
	if len(nonTopOwnerIDs) > 0 {
		tagCacheByOwner, err = s.tagCacheRepo.FindByOwners(ctx, clinicID, nonTopOwnerIDs)
		if err != nil {
			slog.ErrorContext(ctx, "SyncLTVTopPercent: failed to batch load tag cache", "clinic_id", clinicID, "error", err)
			return 0, []error{apperrors.Wrap(err, "failed to batch load tag cache")}
		}
	}

	var errs []error
	count := 0
	for i := range owners {
		owner := &owners[i]
		if owner.LineUserID == nil || *owner.LineUserID == "" {
			continue
		}
		lineUserID := *owner.LineUserID
		_, isTop := topOwnerIDs[owner.ID]
		if isTop {
			if err := s.applyTagState(ctx, client, clinicID, owner.ID, lineUserID, ltvTop20Tag, "LTV top20", "", true); err != nil {
				errs = append(errs, err)
				continue
			}
		} else {
			cached := tagCacheByOwner[owner.ID]
			hasTag := false
			for _, c := range cached {
				if c.TagName == ltvTop20Tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				s.notifyAPISuccess(ctx, client, clinicID, owner.ID, lineUserID)
				count++
				continue
			}
			if err := s.applyTagState(ctx, client, clinicID, owner.ID, lineUserID, ltvTop20Tag, "LTV top20", "", false); err != nil {
				errs = append(errs, err)
				continue
			}
		}
		s.notifyAPISuccess(ctx, client, clinicID, owner.ID, lineUserID)
		count++
	}
	return count, errs
}
