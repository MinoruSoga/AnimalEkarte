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
	FindAllByLineUserID(ctx context.Context, lineUserID string) ([]model.Owner, error)
	UpdateLineUserID(ctx context.Context, clinicID, ownerID uint64, lineUserID *string) error
	UpdateLineFollowedAt(ctx context.Context, clinicID, ownerID uint64, t time.Time) error
	UpdateLineBlockedAt(ctx context.Context, clinicID, ownerID uint64, t time.Time) error
}

// lstepLineSettingReader は LINE 予約設定（reservation domain・topo で lstep は最下流のため
// import 不可）の LINE 認証情報参照の最小view。
type lstepLineSettingReader interface {
	FindByClinicID(ctx context.Context, clinicID uint64) (*model.LineReservationSetting, error)
	FindAll(ctx context.Context) ([]model.LineReservationSetting, error)
}

// lstepSharedFileService は共有ファイル（未移行 service）の署名付きURL取得の最小view。
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
