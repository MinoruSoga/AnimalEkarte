package lstep

import (
	"context"
	"time"

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
