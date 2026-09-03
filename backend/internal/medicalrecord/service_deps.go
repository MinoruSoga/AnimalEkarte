package medicalrecord

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/sharedkernel"
)

// This file declares medicalrecord's consumer-side views of dependencies that live outside
// this package (ADR-006 "aggregator 非経由"): in-package medical-record persistence,
// the shared transaction boundary (Transactor), the shared audit kernel, and LSTEP
// tag-sync / delivery-trigger services. Following Go's "define
// interfaces where they are consumed" guidance and the internal/manualarticle precedent
// (audit.go's AuditLogger), each is the minimal method set this package's services actually
// call — never the full source interface. Concrete implementations are passed in from the
// composition root (cmd/api/main.go) by structural typing; a nil dependency keeps the
// original nil-guard semantics of the moved services.

// medicalRecordFinder は checkupService が親カルテ存在/確定状態を読むための最小 view
// （repository.MedicalRecordRepository.FindByID 相当）。
type medicalRecordFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
}

// medicalRecordLocker は LockByIDForUpdate 1メソッドの narrow interface（BE-refactor.md E-5）。
// prescriptionService / checkupFieldResultService が確定と子書込の競合防止（X-11）に使う。
type medicalRecordLocker interface {
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.MedicalRecord, error)
}

// Transactor は repository.Transactor の consumer-side view（WithTx のみ）。
// prescriptionService / checkupFieldResultService が「削除+挿入+監査」を単一 tx に収めるために使う。
type Transactor interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// AuditEntry は共有監査入力のうち checkupFieldResultService の
// tx 内削除監査（#211 fail-closed）が実際に設定する部分集合を写したもの。Metadata は
// AuditLogInput.Metadata と field-for-field に一致させるため any を維持する。IPAddress/UserAgent は
// この経路では未使用のため持たない（manualarticle.AuditEntry が Metadata を持たないのと同方針）。
type AuditEntry struct {
	ClinicID   *uint64
	ActorID    *uint64
	ActorType  string
	Action     string
	Resource   string
	ResourceID *uint64
	OldValue   any
	NewValue   any
	Metadata   any
}

// AuditTxLogger は共有監査カーネル（tx 内 LogEntryTx）の
// consumer-side view。composition root が具象 service.AuditService.LogEntryTx(ctx, *service.AuditLogInput)
// を本シグネチャへ adapt する（manualarticle/main.go の adapter 先例）。
type AuditTxLogger interface {
	LogEntryTx(ctx context.Context, entry *AuditEntry) error
}

// checkupTagSyncer は checkupService が使う LstepTagSyncService の最小 view。nil 許容（未設定なら同期しない）。
type checkupTagSyncer interface {
	SyncCheckupTag(ctx context.Context, clinicID, ownerID, checkupTypeID uint64, checkupDate time.Time, nextDate *time.Time) error
	ResyncOwnerCheckupTags(ctx context.Context, clinicID, ownerID uint64) error
}

// vaccinationTagSyncer は vaccinationService が使う LstepTagSyncService の最小 view。nil 許容。
type vaccinationTagSyncer interface {
	SyncVaccineTag(ctx context.Context, clinicID, ownerID, vaccinationID uint64) error
	ResyncOwnerVaccineTags(ctx context.Context, clinicID, ownerID uint64) error
}

// vaccinationRelationVerifier は接種記録の Pet→Owner と担当医の現在所属を検証する最小 view。
// 具象は reservation domain の store だが、medicalrecord は consumer-side interface のみを持つ。
type vaccinationRelationVerifier = ClinicalRelationVerifier

// prescriptionTagSyncer は prescriptionService が使う LstepTagSyncService の最小 view。nil 許容。
type prescriptionTagSyncer interface {
	SyncPrescriptionTag(ctx context.Context, clinicID, ownerID uint64) error
}

// checkupFollowUpTrigger は checkupService の健診フォローアップ配信トリガーの最小 view
// （LstepDeliveryTriggerService.TriggerCheckupFollowUp 相当）。nil 許容（未設定ならトリガーしない）。
type checkupFollowUpTrigger interface {
	TriggerCheckupFollowUp(ctx context.Context, clinicID, ownerID uint64) error
}

// ── lab import/report saga consumer-side views (BE9-2D sub-batch③) ──
// The lab services (labImportExaminationService / labReportQueryService / labAuditLogger)
// live in this medicalrecord package. Following the same "define interfaces where they
// are consumed" rule as above, each below is the minimal method set the lab services actually
// call. The composition root (cmd/api/main.go) passes concrete implementations by structural typing.

// examinationImportRepo is labImportExaminationService's write-side view of the examination
// repository: Create + ReplaceItemsByExamID persist the exam and its results, and Delete performs
// the P2-7 orphan-exam compensation (dropping it silently would make a failed row un-retriable).
type examinationImportRepo interface {
	Create(ctx context.Context, exam *model.Examination) error
	ReplaceItemsByExamID(ctx context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, int64, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

// examinationReportRepo is labReportQueryService's read-side view of the examination repository
// (clinic-scoped report queries only).
type examinationReportRepo interface {
	FindByJobID(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.Examination, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
}

// petFinder is labImportExaminationService's minimal view of the pet repository, used to verify
// a request-derived pet_id belongs to the import clinic before persisting the exam (P1-2, #186).
type petFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Pet, error)
}

// AuditLogger is the non-tx consumer-side view of the shared audit kernel
// that labAuditLogger.logBestEffort writes through.
// Distinct from AuditTxLogger above (tx-internal LogEntryTx): lab audit is best-effort and does
// not join the import flow's transaction. The composition root adapts the concrete
// audit LogEntry to this signature.
type AuditLogger interface {
	LogEntry(ctx context.Context, entry *AuditEntry) error
}

// BUG-015: vitalService は best-effort の LogVitalChange を廃止し、AuditTxLogger.LogEntryTx
// （ambient tx 参加・fail-closed）を使う。nil audit 依存は Create/Update/Delete 入口で拒否する。

// ── treatment consumer-side views (BE9-2D sub-batch④b) ──
// treatmentService lives in this medicalrecord package (Transactor.WithTx + per-repo DBOrTx).
// The treatment/vital repositories are in-package concrete types; the views below cover
// remaining cross-domain dependencies. Concrete implementations are passed in from cmd/api/main.go.

// treatmentMedicalRecordRepo is treatmentService's view of the medical-record repository:
// FindByID for the pre-tx fast-fail finalized check, LockByIDForUpdate for the X-11 row lock
// (lockDraftMedicalRecord).
type treatmentMedicalRecordRepo interface {
	medicalRecordFinder
	medicalRecordLocker
}

// medicineFinder is treatmentService's master-FK ownership-validation + #201 dose-revalidation
// read view over the medicine master (repository.MedicineRepository.FindByID 相当).
type medicineFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Medicine, error)
}

// procedureFinder / consultationFinder are treatmentService's master-FK ownership-validation
// read views (#124/#125 同型 mislink 防止の write 前検証)。
type procedureFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Procedure, error)
}

type consultationFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
}

// treatmentInventoryRepo is treatmentService's view of the inventory repository. FindByID rejects
// cross-clinic treatment links, while DecreaseStock repeats clinic scope in the atomic stock update.
type treatmentInventoryRepo interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	DecreaseStock(ctx context.Context, clinicID, id uint64, quantity float64) error
}

// doseParamFinder is treatmentService's per-species dose-param read view（#201 B-2 保存時再検証、
// species で引くことで犬↔猫越境を構造的に排除する HIGH-4 経路）.
type doseParamFinder interface {
	FindByMedicineAndSpecies(ctx context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies) (*model.MedicineDoseParam, error)
}

// ── hospitalization/discharge-with-billing consumer-side views (BE9-2D ⑤) ──
// hospitalizationService lives in this medicalrecord package (WithTx + 個別注入).
// hospitalization/hospitalization_plan/daily_record/care_plan_item の各 repository は in-package
// 具象。以下は他 domain 依存の最小 view。owner/pet link
// 検証は sharedkernel.OwnerPetLinkVerifier。

// cageFinder は入院の CageID master-FK 所有権検証 read view（X-14 同型）。
type cageFinder interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Cage, error)
}

// accountingCreator は DischargeWithBilling の会計レコード作成 write view
// （repository.AccountingRepository.Create 相当・R1-1 dbOrTx 参加済み）。
type accountingCreator interface {
	Create(ctx context.Context, clinicID uint64, billing *model.Billing) error
}

// billingItemWriter は DischargeWithBilling の明細作成+合計更新 write view
// （repository.BillingItemRepository の部分集合・R1-1 dbOrTx 参加済み。billing 側の型は
// import しない — *model.Billing/BillingItem は model 帰属のため ADR-006 の billing→medicalrecord
// 逆依存禁止に抵触しない）。
type billingItemWriter interface {
	Create(ctx context.Context, item *model.BillingItem) error
	UpdateBillingTotals(ctx context.Context, clinicID, billingID uint64, subtotal, taxTotal, totalAmount int64) error
}

// medicineInventoryRepo は medicineService の BUG-320 連携在庫同期 write view + ParentID/
// InventoryID master-FK 検証 read view（inventory.Repository の部分集合・BE9-2D ⑥）。
type medicineInventoryRepo interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.InventoryItem, error)
	Create(ctx context.Context, clinicID uint64, item *model.InventoryItem) error
	UpdateNameByMedicineCategory(ctx context.Context, clinicID uint64, oldName, newName string) error
	DeleteByNameAndMedicineCategory(ctx context.Context, clinicID uint64, name string) error
	// Delete removes inventory by clinic-scoped id (MRC-02: link medicines.inventory_id).
	Delete(ctx context.Context, clinicID, id uint64) error
}

// ── medical_record 本体/addendum consumer-side views (BE9-2D ⑦) ──

// mrLineCustomerRepo は LINE 顧客の read view（medical_record の LINE 連携 hydrate 用）。
type mrLineCustomerRepo interface {
	FindByID(ctx context.Context, clinicID, id uint64) (*model.LineCustomer, error)
}

// mrReservationRepo は予約読取+カルテ作成時の欠損context補完 view。
type mrReservationRepo interface {
	sharedkernel.OwnerPetLinkVerifier
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Reservation, error)
	// FindPetByIDInClinic は死亡ペットへの新規カルテ作成拒否（BUG-002 / SD-10）に使う。
	FindPetByIDInClinic(ctx context.Context, clinicID, petID uint64) (*model.Pet, error)
	AssertMedicalRecordDoctorInClinic(ctx context.Context, clinicID, doctorID uint64) error
	BackfillForMedicalRecord(ctx context.Context, clinicID, id uint64, ownerID, petID, doctorID *uint64) error
	PrepareForMedicalRecordFinalization(ctx context.Context, clinicID, id uint64) error
}

// mrDeliveryTrigger は初診ウェルカム配信トリガーの view（nil 許容）。
type mrDeliveryTrigger interface {
	TriggerFirstVisitWelcome(ctx context.Context, clinicID, ownerID uint64) error
}

// mrTagSyncer は来院系 Lstep タグ同期 view（nil 許容）。
type mrTagSyncer interface {
	SyncNextVisitTag(ctx context.Context, clinicID, ownerID uint64) error
	SyncVisitCompletionTags(ctx context.Context, clinicID, ownerID uint64) error
}

// mrAuditLogger は共有監査カーネルのカルテ変更/追記記録 view（vitalAuditLogger 同様、signature が
// primitives+map/model 型のみのため composition root は具象 service.AuditService を無 adapter で直渡し）。
type mrAuditLogger interface {
	LogMedicalRecordChange(ctx context.Context, clinicID uint64, actorID *uint64, action string, recordID uint64, oldValue, newValue map[string]any) error
	LogAddendumCreate(ctx context.Context, clinicID uint64, actorID *uint64, addendumID, medicalRecordID uint64, addendum *model.MedicalRecordAddendum) error
}
