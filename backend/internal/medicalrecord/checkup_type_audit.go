package medicalrecord

import (
	"context"
	"log/slog"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// EMR-456: 健診種別の保険設定（is_non_insurance）の変更を、書込と同じ tx で fail-closed に監査する。
// action/resource は健診種別専有の値としてここに置く（共有 model の監査定数への集約は
// EMR-454 の共有契約で行う）。
const (
	checkupTypeAuditActionCreate = "checkup_type.create"
	checkupTypeAuditActionUpdate = "checkup_type.update"
	checkupTypeAuditResource     = "checkup_type"
)

func (s *checkupTypeService) withTx(ctx context.Context, fn func(context.Context) error) error {
	// nil transactor は composition/wiring 不備（サーバー内部）であり client 入力ではない → 500。
	if s.transactor == nil {
		return apperrors.WrapInternalServerError("checkup type transaction dependency is required")
	}
	return s.transactor.WithTx(ctx, fn)
}

// checkupTypeInsuranceAuditValue は監査に残す保険設定と版。nil の is_non_insurance は
// 「未保存（旧欠落）」として null のまま記録し、false/true に補完しない。
func checkupTypeInsuranceAuditValue(ct *model.CheckupType) map[string]any {
	if ct == nil {
		return nil
	}
	return map[string]any{
		"is_non_insurance": ct.IsNonInsurance,
		"version":          ct.Version,
	}
}

func (s *checkupTypeService) auditInsuranceTx(
	ctx context.Context,
	clinicID uint64,
	actorID *uint64,
	action string,
	before, after *model.CheckupType,
) error {
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("checkup type audit dependency is required")
	}
	if actorID == nil || *actorID == 0 {
		return apperrors.WrapInvalidInput("authenticated staff actor is required for checkup type insurance setting")
	}
	resourceID := after.ID
	entry := &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  model.AuditActorTypeStaff,
		Action:     action,
		Resource:   checkupTypeAuditResource,
		ResourceID: &resourceID,
		OldValue:   checkupTypeInsuranceAuditValue(before),
		NewValue:   checkupTypeInsuranceAuditValue(after),
		Metadata: map[string]any{
			"field": colCheckupTypeNonIns,
		},
	}
	if err := s.auditTx.LogEntryTx(ctx, entry); err != nil {
		slog.ErrorContext(ctx, "failed to audit checkup type insurance setting",
			"error", err,
			"clinic_id", clinicID,
			"checkup_type_id", resourceID,
		)
		return apperrors.Wrap(err, "failed to write checkup type insurance audit")
	}
	return nil
}

func sameOptionalBool(a, b *bool) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
