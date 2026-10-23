package lstep

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/billing"
	"github.com/animal-ekarte/backend/internal/medicalrecord"
	"github.com/animal-ekarte/backend/internal/model"
)

// consumer-side narrow views（ADR-006/BE9-2C L①）: lstep が必要とする未移行 domain repo /
// 横断サービスの最小メソッド集合。具象は service 集約または cmd/api/main.go が注入する。

// lstepClinicSettingsRepo は病院設定（clinic domain 未移行）の CPM/休眠/予防しきい値更新の最小view。
type lstepClinicSettingsRepo interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (*model.ClinicSettings, error)
	UpdateCPMVersion(ctx context.Context, clinicID uint64, version string) error
	UpdateCPMV1Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV1Thresholds) error
	UpdateCPMV2Thresholds(ctx context.Context, clinicID uint64, thresholds model.CPMV2Thresholds) error
	UpdateDormantThresholds(ctx context.Context, clinicID uint64, thresholds model.DormantThresholds) error
	UpdateHealthPreventionThresholds(ctx context.Context, clinicID uint64, thresholds model.HealthPreventionThresholds) error
}

// lstepAuditLogger は Lステップ操作監査（best-effort）の最小view。
type lstepAuditLogger interface {
	LogLstepOperation(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64) error
}

// lstepOwnerRepo は飼主（owner domain 未移行）の LINE 連携フィールド更新の最小view。
type lstepOwnerRepo interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
}

// lineLinkOwnerRepo is the public LINE account-link/webhook owner capability.
// Webhook lookups are always scoped by the clinic whose channel secret uniquely
// matched the request signature.
type lineLinkOwnerRepo interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Owner, error)
	FindByLineUserID(ctx context.Context, clinicID uint64, lineUserID string) (*model.Owner, error)
	LockLineLinkOwner(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
	UpdateLineUserID(ctx context.Context, clinicID, ownerID uint64, lineUserID *string) error
	UpdateLineFollowedAt(
		ctx context.Context,
		clinicID, ownerID uint64,
		expectedLineUserID string,
		eventAt time.Time,
	) (bool, error)
	UpdateLineBlockedAt(
		ctx context.Context,
		clinicID, ownerID uint64,
		expectedLineUserID string,
		eventAt time.Time,
	) (bool, error)
}

// LineLinkAuditTxLogger is the consumer-side audit contract for a successful
// public LINE account link. Implementations must write through the ambient tx.
type LineLinkAuditTxLogger interface {
	LogOwnerLineLinkTx(ctx context.Context, clinicID, ownerID uint64) error
}

// lineWebhookRouteReader は webhook の destination（LINE bot user ID）から clinic を引く最小view
// （reservation domain 実装・topo で lstep は最下流のため import 不可）。credential payload は返さない。
// 旧 line_reservation_settings.line_bot_user_id は DROP 済みで置換先の routing 保存は未決定（EMR-504 Q7）。
// 実装は未決定の間 apperrors.CodeLineWebhookRoutingPending（422）を返して fail-closed する。
type lineWebhookRouteReader interface {
	FindWebhookRouteByLineBotUserID(ctx context.Context, lineBotUserID string) (clinicID uint64, err error)
}

// lineChannelCredentialReader は canonical clinic_integrations credential を1件だけ読む。
type lineChannelCredentialReader interface {
	FindCredentialByClinicServiceKey(ctx context.Context, clinicID uint64, service, keyName string) (*model.ClinicIntegration, error)
}

// lstepSharedFileService は共有ファイル署名付きURL取得のconsumer-side最小view。
// LINE 配送は外部プラットフォームのフェッチ用途のため delivery TTL を使う。
type lstepSharedFileService interface {
	GetSignedURLForDelivery(ctx context.Context, clinicID, id uint64) (string, error)
}

// tagOwnerFinder は手動タグ操作が必要とする飼主参照の最小view。
type tagOwnerFinder interface {
	FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
}

// tagSyncOwnerRepo はタグ同期が必要とする飼主参照の最小view。
type tagSyncOwnerRepo interface {
	FindByID(ctx context.Context, clinicID, ownerID uint64) (*model.Owner, error)
	FindAllWithLineUserID(ctx context.Context, clinicID uint64) ([]model.Owner, error)
	FindAllWithLineUserIDCursor(ctx context.Context, clinicID, afterID uint64, limit int) ([]model.Owner, error)
}

type tagSyncVaccinationRepo interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Vaccination, error)
	FindByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Vaccination, error)
}

type tagSyncMedicalRecordRepo interface {
	FindOwnerVisitSummary(ctx context.Context, clinicID, ownerID uint64) (*medicalrecord.OwnerVisitSummary, error)
}

// tagSyncNextVisitAppointmentRepo は next_visit_* タグ同期が必要とする
// 次回 appointment 参照の最小view（EMR-308）。source of truth はカルテの
// 旧推奨日カラムではなく予約（tentative は intended_date、
// それ以外は start_time）に移る。具象は reservation domain の store で、
// composition root が注入する（lstep は最下流で reservation を import できない
// ため consumer-side interface で受け取る）。nil 許容。
type tagSyncNextVisitAppointmentRepo interface {
	// FindNextAppointmentForOwner は飼い主の次回来院に相当する active appointment
	// を返す。該当なしは (nil, nil)。tentative の意味上の日付は IntendedDate、
	// それ以外は StartTime。
	FindNextAppointmentForOwner(ctx context.Context, clinicID, ownerID uint64) (*model.Reservation, error)
}

type tagSyncAccountingRepo interface {
	SumPaidByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	MaxSingleVisitAmountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	FindOwnersByAnnualRevenue(ctx context.Context, clinicID uint64) ([]billing.OwnerAnnualRevenue, error)
}

type tagSyncPetRepo interface {
	FindLivingByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Pet, error)
	CountByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
	CountLivingByOwner(ctx context.Context, clinicID, ownerID uint64) (int64, error)
}

type tagSyncCheckupRepo interface {
	FindByOwnerID(ctx context.Context, clinicID, ownerID uint64) ([]model.Checkup, error)
}

type tagSyncBillingItemRepo interface {
	HasItemByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
	HasFoodPurchaseByOwnerSince(ctx context.Context, clinicID, ownerID uint64, since time.Time, names []string) (bool, error)
}

type tagSummaryRepo interface {
	TagSummary(ctx context.Context, clinicID uint64) ([]TagSummaryRow, int64, error)
	FindOwnersByTag(ctx context.Context, clinicID uint64, tagName, nameQuery string, offset, limit int) ([]TagOwnerRow, int64, error)
}

// checkup-sync consumes only these batch-scoped views from owner/pet/tag cache.
// Keeping the views local prevents the lstep package from depending on the
// repository or service aggregators.
type checkupSyncOwnerRepo interface {
	FindByIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) ([]*model.Owner, error)
}

type checkupSyncPetRepo interface {
	CountLivingByOwnerIDs(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64]int64, error)
}

type checkupSyncTagCacheRepo interface {
	FindByOwners(ctx context.Context, clinicID uint64, ownerIDs []uint64) (map[uint64][]*model.LstepTagCache, error)
	UpsertTag(ctx context.Context, clinicID, ownerID uint64, tagName, category, reason string) error
}

type checkupSyncSettingsService interface {
	IsSyncEnabled(ctx context.Context, clinicID uint64) (bool, error)
	GetRawCredentials(ctx context.Context, clinicID uint64) (apiKey, baseURL, lineToken string, err error)
	GetCPMV1Thresholds(ctx context.Context, clinicID uint64) (model.CPMV1Thresholds, error)
}

type checkupSyncAuditLogger interface {
	LogLstepOperationWithMetadata(ctx context.Context, clinicID uint64, actorID *uint64, action, resource string, resourceID *uint64, metadata any) error
}
