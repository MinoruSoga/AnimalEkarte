package medicalrecord

import (
	"context"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// vaccinationAuditActorKey はワクチン削除監査の actor（認証済み staff_id）を運ぶ
// domain-private context key。request/body 由来の値は決して載せない。
type vaccinationAuditActorKey struct{}

// withVaccinationAuditActor は認証済み staff actor を載せた context を返す。
// HTTP境界（DeleteVaccination）が httpapi.ExtractStaffID の結果だけを渡す。
func withVaccinationAuditActor(ctx context.Context, staffID uint64) context.Context {
	return context.WithValue(ctx, vaccinationAuditActorKey{}, staffID)
}

// vaccinationAuditActorFrom は withVaccinationAuditActor で載せた staff actor を取り出す。
// 不在なら監査 actor は system（内部直接呼出）扱い。
func vaccinationAuditActorFrom(ctx context.Context) (uint64, bool) {
	staffID, ok := ctx.Value(vaccinationAuditActorKey{}).(uint64)
	return staffID, ok
}

// vaccinationDeleteAuditSnapshot はワクチン削除監査の old_value に書く typed minimal
// snapshot。AuditEntry.OldValue（any）にそのまま載る。
type vaccinationDeleteAuditSnapshot struct {
	ID              uint64    `json:"id"`
	MedicalRecordID *uint64   `json:"medical_record_id,omitempty"`
	PetID           *uint64   `json:"pet_id,omitempty"`
	VaccineID       uint64    `json:"vaccine_id"`
	DoctorID        *uint64   `json:"doctor_id,omitempty"`
	Date            time.Time `json:"date"`
}

// vaccinationDeleteAuditSnapshotFrom はロック済みの削除対象行から typed snapshot を
// 組み立てる（削除後に参照できないため、削除前の locked state が正本）。
func vaccinationDeleteAuditSnapshotFrom(v *model.Vaccination) vaccinationDeleteAuditSnapshot {
	return vaccinationDeleteAuditSnapshot{
		ID:              v.ID,
		MedicalRecordID: v.MedicalRecordID,
		PetID:           v.PetID,
		VaccineID:       v.VaccineID,
		DoctorID:        v.DoctorID,
		Date:            v.Date,
	}
}

// auditVaccinationDeleteTx はワクチン削除の監査を ambient transaction へ参加させる
// （EMR-323 / P323-AUDIT — slog だけでは ticket を満たさない。vital_audit.go
// auditVitalTx の BUG-015 fail-closed パターンと同型）。
// actor は withVaccinationAuditActor 経由の認証済み staff のみ — request/body 由来の
// 値は載せない。actor context が無い内部直接呼出は system actor（ActorID nil）で記録する。
// nil audit 依存は fail-closed で拒否し、監査書込の失敗は呼び出し元の tx を rollback させる。
func (s *vaccinationService) auditVaccinationDeleteTx(ctx context.Context, clinicID, vaccinationID uint64, deleted *model.Vaccination) error {
	if s.auditTx == nil {
		return apperrors.WrapInternalServerError("vaccination audit dependency is required")
	}
	var actorID *uint64
	if staffID, ok := vaccinationAuditActorFrom(ctx); ok {
		actorID = &staffID
	}
	resourceID := vaccinationID
	entry := &AuditEntry{
		ClinicID:   &clinicID,
		ActorID:    actorID,
		ActorType:  sharedkernel.AuditActorTypeFor(actorID),
		Action:     "delete",
		Resource:   "vaccination",
		ResourceID: &resourceID,
		OldValue:   vaccinationDeleteAuditSnapshotFrom(deleted),
	}
	if err := s.auditTx.LogEntryTx(ctx, entry); err != nil {
		return apperrors.Wrap(err, "failed to write vaccination delete audit")
	}
	return nil
}
