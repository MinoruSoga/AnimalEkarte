package reservation

import (
	"context"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/persistence"
)

// ---------------------------------------------------------------------------
// EMR-308: next-visit tag sync invocation（post-commit best-effort）
//
// 予約変更容易点（intent 設定/解除、tentative 確定・昇格、作成・更新・削除、
// 管理/LIFF 経路）は commit 成功後に飼い主の next_visit_* タグ再同期を要求する。
// LSTEP への実処理は composition root が注入する NextVisitTagSyncer に委譲し、
// 未注入・同期失敗は予約変更容易点の成否へ影響させない（best-effort）。
// ---------------------------------------------------------------------------

// nextVisitTagSyncTimeout bounds the entire post-commit owner batch, including
// DB reads and external calls. A later owner does not receive a fresh budget.
const nextVisitTagSyncTimeout = 30 * time.Second

// syncNextVisitTagBestEffort は飼い主1件の next-visit タグ同期を呼び出す。
// syncer 未注入・owner 未確定のときは何もしない。失敗は warn ログのみ。
func syncNextVisitTagBestEffort(ctx context.Context, syncer NextVisitTagSyncer, clinicID uint64, ownerID *uint64) {
	syncNextVisitTagsBestEffort(ctx, syncer, clinicID, ownerID)
}

// syncNextVisitTagsBestEffort は複数飼い主候補（例: 更新前後の新旧 owner、
// バッチ作成の各 owner）を重複排除して順に同期する。nil/0 はスキップする。
func syncNextVisitTagsBestEffort(ctx context.Context, syncer NextVisitTagSyncer, clinicID uint64, ownerIDs ...*uint64) {
	if syncer == nil {
		return
	}
	// Commit済みの同期はrequest取消に依存させない。trace等の値は保持し、
	// 終了済みtransactionは外す。同期処理と警告ログには同じ期限付きcontextを使う。
	syncCtx, cancel := context.WithTimeout(persistence.DetachTx(ctx), nextVisitTagSyncTimeout)
	defer cancel()

	seen := make(map[uint64]struct{}, len(ownerIDs))
	for _, ownerID := range ownerIDs {
		if ownerID == nil || *ownerID == 0 {
			continue
		}
		if _, dup := seen[*ownerID]; dup {
			continue
		}
		seen[*ownerID] = struct{}{}
		if err := syncer.SyncNextVisitTag(syncCtx, clinicID, *ownerID); err != nil {
			slog.WarnContext(syncCtx, "next visit tag sync failed (best-effort)",
				slog.Uint64("clinic_id", clinicID),
				slog.Uint64("owner_id", *ownerID),
				slog.Any("error", err))
		}
	}
}
