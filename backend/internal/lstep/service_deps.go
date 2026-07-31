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

// lstepLineSettingReader は LINE 予約設定（reservation domain・topo で lstep は最下流のため
// import 不可）の予約設定と webhook routing metadata 参照の最小view。
type lstepLineSettingReader interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	// FindWebhookRouteByLineBotUserID は credential payload を返さず、destination に対応する
	// clinic identity と旧 credential の存在有無だけを返す。
	FindWebhookRouteByLineBotUserID(ctx context.Context, lineBotUserID string) (clinicID uint64, legacyCredentialPresent bool, err error)
	FindAll(ctx context.Context) ([]model.LineReservationSetting, error)
}

// lineChannelCredentialReader は canonical clinic_integrations credential を1件だけ読む。
type lineChannelCredentialReader interface {
	FindCredentialByClinicServiceKey(ctx context.Context, clinicID uint64, service, keyName string) (*model.ClinicIntegration, error)
}

// lstepSharedFileService は共有ファイル署名付きURL取得のconsumer-side最小view。
type lstepSharedFileService interface {
	GetSignedURL(ctx context.Context, clinicID, id uint64) (string, error)
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
	FindLatestByOwner(ctx context.Context, clinicID, ownerID uint64) (*model.MedicalRecord, error)
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

type tagSyncPrescriptionRepo interface {
	FindActiveByOwner(ctx context.Context, clinicID, ownerID uint64) ([]model.Prescription, error)
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
