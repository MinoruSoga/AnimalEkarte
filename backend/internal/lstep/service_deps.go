package lstep

import (
	"context"

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
