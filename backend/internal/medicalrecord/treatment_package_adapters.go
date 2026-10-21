package medicalrecord

// Treatment package adapters (EMR-521 stage 3): connect the service's
// consumer-side view authorizer and per-kind finders to the existing
// permission check and existing master repositories. No existing single-master
// reader is reimplemented — the repository interfaces already satisfy the
// finder signatures, and the composition root passes them through
// NewTreatmentPackageDependencies.

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/httpapi"
	"github.com/animal-ekarte/backend/internal/model"
)

// packageViewCheckFunc はリクエスト単位の per-clinic RBAC 評価器。HTTP 境界が
// permission middleware が載せた httpapi.ClinicPermissionChecker を gin.Context に
// 束縛して ctx に載せる。ctx 以外の経路（body/query）から clinic/actor を受けない。
type packageViewCheckFunc func(clinicID uint64, resource model.Resource, action string) bool

// packageViewCheckContextKey はリクエスト単位 authorizer を運ぶ domain-private
// context key（vaccination_audit.go の withVaccinationAuditActor と同型）。
type packageViewCheckContextKey struct{}

func withPackageViewCheck(ctx context.Context, check packageViewCheckFunc) context.Context {
	return context.WithValue(ctx, packageViewCheckContextKey{}, check)
}

func packageViewCheckFrom(ctx context.Context) (packageViewCheckFunc, bool) {
	check, ok := ctx.Value(packageViewCheckContextKey{}).(packageViewCheckFunc)
	if !ok || check == nil {
		return nil, false
	}
	return check, true
}

// contextViewAuthorizer は Granted の ctx に載ったリクエスト単位 checker へ
// 委譲する singleton adapter。checker 欠落は fail-closed（deny）。
type contextViewAuthorizer struct{}

func (contextViewAuthorizer) Granted(ctx context.Context, clinicID uint64, resource model.Resource, action string) bool {
	check, ok := packageViewCheckFrom(ctx)
	return ok && check(clinicID, resource, action)
}

// treatmentPackageViewCheckContext はこのリクエストの per-clinic checker を
// service が読む request ctx へ束縛する。middleware が checker を載せていない
// 場合、system_admin は auth.checkPermissionInClinic と同じ explicit bypass で
// 許可し、それ以外は fail-closed（呼び出し元が 403 を返す）。
func treatmentPackageViewCheckContext(c *gin.Context) (context.Context, bool) {
	if check, ok := httpapi.PeekClinicPermissionChecker(c); ok {
		bound := packageViewCheckFunc(func(clinicID uint64, resource model.Resource, action string) bool {
			return check(c, clinicID, string(resource), action)
		})
		return withPackageViewCheck(c.Request.Context(), bound), true
	}
	if admin, ok := httpapi.PeekIsSystemAdmin(c); ok && admin {
		return withPackageViewCheck(c.Request.Context(), func(uint64, model.Resource, string) bool {
			return true
		}), true
	}
	return nil, false
}

// NewTreatmentPackageDependencies は composition root 専用の結線 adapter。
// 既存 master repository は各 finder interface を構造的に満たすため再実装せず、
// ここで束ねて service へ渡す。authorizer は ctx 経由のリクエスト単位評価を
// 読む singleton（contextViewAuthorizer）を組み込む。
func NewTreatmentPackageDependencies(
	tx Transactor,
	audit AuditTxLogger,
	consultations packageConsultationFinder,
	procedures packageProcedureFinder,
	medicines packageMedicineFinder,
	examTypes packageExamTypeFinder,
	vaccines packageVaccineFinder,
	checkupTypes packageCheckupTypeFinder,
	merchandiseItems packageMerchandiseItemFinder,
) TreatmentPackageDependencies {
	return TreatmentPackageDependencies{
		Transactor:       tx,
		Audit:            audit,
		Authorizer:       contextViewAuthorizer{},
		Consultations:    consultations,
		Procedures:       procedures,
		Medicines:        medicines,
		ExamTypes:        examTypes,
		Vaccines:         vaccines,
		CheckupTypes:     checkupTypes,
		MerchandiseItems: merchandiseItems,
	}
}
