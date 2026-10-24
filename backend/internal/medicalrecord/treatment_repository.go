package medicalrecord

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Moved from internal/repository (BE9-2D sub-batch④b). The former package-private dbOrTx is
// swapped for persistence.DBOrTx (identical ambient-tx participation); paginate resolves to this
// package's pagination.go (same predicate). External callers only ever saw this via the
// internal/repository facade (TreatmentRepository alias), so no call site changes.

// TreatmentSortUpdate は並び順一括更新に使う軽量DTO
type TreatmentSortUpdate struct {
	ID              uint64
	ClinicID        uint64
	MedicalRecordID uint64 // service が施錠した親カルテに束縛する（MRD-01）
	SortOrder       int
}

// TreatmentRepository は治療項目の永続化インターフェース
type TreatmentRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error)
	// LockByIDForUpdate serializes treatment writes (discount recheck TOCTOU, dose snapshot).
	// Fail-closed without an ambient transaction — FOR UPDATE is meaningless under autocommit.
	LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Treatment, error)
	// FindUnbilledByPetID は医師確認済み・未会計の診療項目の候補。held for EMR-462/481:
	// 医師確認 set と請求元の紐付けの reader が未接続のため、固定エラー（422 finance_consumer_pending）を返す。
	FindUnbilledByPetID(ctx context.Context, clinicID, petID uint64) ([]model.Treatment, error)
	// FindHistoryByPetID は #158/#159 飼主レポート用: ペット単位で treatments を横断取得する。
	// medical_records JOIN で clinic_id 隔離し、medical_records.date 降順で返す。
	// filter.AnesthesiaOnly=true / filter.IsSurgery=true 指定時は procedures JOIN で絞り込む。
	FindHistoryByPetID(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error)
	// CountFinalizedUnconfirmedByPetAndDate は同日同ペットの「未会計対象化」診察カルテ件数を返す(#77)。
	// held for EMR-462/481: 旧実装が読んでいた billing_confirmations と billings.status は廃止された。
	// 医師確認 set と請求元の紐付けの reader が未接続のため、固定エラー（422 finance_consumer_pending）を返す。
	CountFinalizedUnconfirmedByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error)
	Create(ctx context.Context, treatment *model.Treatment) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateTreatmentInput) error
	Delete(ctx context.Context, clinicID, id uint64) error
	BulkUpdateSortOrder(ctx context.Context, updates []TreatmentSortUpdate) error
}

type treatmentRepository struct {
	db *gorm.DB
}

// NewTreatmentRepository はTreatmentRepositoryを初期化して返す
func NewTreatmentRepository(db *gorm.DB) TreatmentRepository {
	return &treatmentRepository{db: db}
}

func (r *treatmentRepository) FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error) {
	treatments := make([]model.Treatment, 0)
	if err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id AND medical_records.deleted_at IS NULL").
		Where("medical_records.clinic_id = ? AND treatments.medical_record_id = ? AND treatments.deleted_at IS NULL", clinicID, medicalRecordID).
		Scopes(treatmentReadPreloads(clinicID)).
		Order("treatments.sort_order ASC").
		Find(&treatments).Error; err != nil {
		return nil, apperrors.FromGORM(err, "treatment", "")
	}
	sanitizeTreatmentMasterRelations(treatments, clinicID)
	return treatments, nil
}

func (r *treatmentRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error) {
	var treatment model.Treatment
	err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id AND medical_records.deleted_at IS NULL").
		Where("medical_records.clinic_id = ? AND treatments.id = ? AND treatments.deleted_at IS NULL", clinicID, id).
		Scopes(treatmentReadPreloads(clinicID)).
		First(&treatment).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment", fmt.Sprintf("%d", id))
	}
	sanitizeTreatmentMasterRelation(&treatment, clinicID)
	return &treatment, nil
}

// LockByIDForUpdate は clinic-scoped FOR UPDATE で treatment 行を固定する（SEC-CS-F09）。
// Joins() は SELECT 専用のため clinic 隔離は Update と同じ subquery で表現する。
func (r *treatmentRepository) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Treatment, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("treatment lock requires an ambient transaction")
	}
	var treatment model.Treatment
	err := persistence.DBOrTx(ctx, r.db).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(
			"treatments.id = ? AND treatments.deleted_at IS NULL AND treatments.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)",
			id, clinicID,
		).
		First(&treatment).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment", fmt.Sprintf("%d", id))
	}
	return &treatment, nil
}

// FindUnbilledByPetID は「医師確認済み・未会計」の診療項目を返していた。held for EMR-462/481。
// 旧実装は billing_confirmations（pending/confirmed/returned）と billing_items.treatment_id・billings.status の
// NOT EXISTS で候補を絞っていたが、いずれも廃止された（001_init.sql:10889-10891,10921）。後継は医師確認 set
// （doctor_review_sets / records。462）と請求元の紐付け（billing_source_bindings。481）で、reader が未接続のため、
// 候補を 0 件で返さず固定エラー（422 finance_consumer_pending）で拒否する。
func (*treatmentRepository) FindUnbilledByPetID(_ context.Context, _, _ uint64) ([]model.Treatment, error) {
	return nil, apperrors.WrapFinanceConsumerPending("未会計の診療項目の候補は、医師確認と請求元の紐付けの読取が未接続のため取得できません")
}

func (r *treatmentRepository) FindHistoryByPetID(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error) {
	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).
			Model(&model.Treatment{}).
			Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id AND medical_records.deleted_at IS NULL").
			Where("medical_records.clinic_id = ? AND medical_records.pet_id = ? AND treatments.deleted_at IS NULL", clinicID, petID)
		q = q.Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = medical_records.pet_id AND p.clinic_id = medical_records.clinic_id)")
		if filter.ItemType != nil {
			q = q.Where("treatments.item_type = ?", *filter.ItemType)
		}
		// #159: 処置 JOIN フィルタ（procedure_id が NULL 以外の行のみ通る INNER JOIN で暗黙 item_type 絞り込み）
		if filter.AnesthesiaOnly || filter.IsSurgery {
			q = q.Joins("JOIN procedures ON procedures.id = treatments.procedure_id AND procedures.clinic_id = medical_records.clinic_id AND procedures.deleted_at IS NULL")
			if filter.AnesthesiaOnly {
				q = q.Where("procedures.anesthesia != ?", string(model.AnesthesiaTypeNone))
			}
			if filter.IsSurgery {
				q = q.Where("procedures.is_surgery = ?", true)
			}
		}
		return q
	}

	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "treatment", fmt.Sprintf("pet=%d", petID))
	}

	treatments := make([]model.Treatment, 0)
	if err := buildBase().
		Scopes(treatmentReadPreloads(clinicID)).
		Order("medical_records.date DESC, treatments.sort_order ASC, treatments.id DESC").
		Scopes(paginate(page, limit)).
		Find(&treatments).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "treatment", fmt.Sprintf("pet=%d", petID))
	}
	sanitizeTreatmentMasterRelations(treatments, clinicID)
	return treatments, total, nil
}

func treatmentReadPreloads(clinicID uint64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Preload("MedicalRecord", "clinic_id = ? AND deleted_at IS NULL", clinicID).
			Preload("Consultation", "clinic_id = ? AND deleted_at IS NULL", clinicID).
			Preload("Procedure", "clinic_id = ? AND deleted_at IS NULL", clinicID).
			Preload("Medicine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
			Preload("Inventory", "clinic_id = ? AND deleted_at IS NULL", clinicID)
	}
}

// sanitizeTreatmentMasterRelations preserves the clinic-owned treatment row and
// its clinical text while clearing polluted nullable master references. Scoped
// preloads are the ownership proof: if a raw FK does not resolve to the exact
// parent clinic, neither the association nor the raw ID may reach consumers.
func sanitizeTreatmentMasterRelations(treatments []model.Treatment, clinicID uint64) {
	for i := range treatments {
		sanitizeTreatmentMasterRelation(&treatments[i], clinicID)
	}
}

func sanitizeTreatmentMasterRelation(treatment *model.Treatment, clinicID uint64) {
	if treatment == nil {
		return
	}
	if treatment.ConsultationID != nil &&
		(treatment.Consultation == nil ||
			treatment.Consultation.ID != *treatment.ConsultationID ||
			treatment.Consultation.ClinicID != clinicID) {
		treatment.ConsultationID = nil
		treatment.Consultation = nil
	}
	if treatment.ProcedureID != nil &&
		(treatment.Procedure == nil ||
			treatment.Procedure.ID != *treatment.ProcedureID ||
			treatment.Procedure.ClinicID != clinicID) {
		treatment.ProcedureID = nil
		treatment.Procedure = nil
	}
	if treatment.MedicineID != nil &&
		(treatment.Medicine == nil ||
			treatment.Medicine.ID != *treatment.MedicineID ||
			treatment.Medicine.ClinicID != clinicID) {
		treatment.MedicineID = nil
		treatment.Medicine = nil
	}
	if treatment.InventoryID != nil &&
		(treatment.Inventory == nil ||
			treatment.Inventory.ID != *treatment.InventoryID ||
			treatment.Inventory.ClinicID != clinicID) {
		treatment.InventoryID = nil
		treatment.Inventory = nil
	}
}

// CountFinalizedUnconfirmedByPetAndDate は同日同ペットの「未会計対象化」診察カルテ件数を返していた(#77)。held for EMR-462/481。
// 旧実装が読んでいた billing_confirmations と billings.status は廃止された。後継の reader（医師確認 set・請求元の紐付け）が
// 未接続のため、0 件を返さず固定エラー（422 finance_consumer_pending）で拒否する。
// reader の実装時は、診療終了済みの判定に EMR-495 Q5（medical_record_consultation_events の最新 revision が ended）を使うこと。
func (*treatmentRepository) CountFinalizedUnconfirmedByPetAndDate(_ context.Context, _, _ uint64, _ time.Time) (int64, error) {
	return 0, apperrors.WrapFinanceConsumerPending("同日の未会計対象化の診察件数は、医師確認と請求元の紐付けの読取が未接続のため取得できません")
}

// Create は persistence.DBOrTx(ctx, r.db) で ambient tx（Transactor.WithTx）に参加する（BE9-2D ④b）。
// treatmentService.Create が lockDraftMedicalRecord の行ロック・在庫減算・逸脱監査と同一 tx で
// 呼ぶため、tx 非参加だと X-11 finalize 直列化と atomicity（CLAUDE.md 不変条件）が壊れる。
func (r *treatmentRepository) Create(ctx context.Context, treatment *model.Treatment) error {
	if err := persistence.DBOrTx(ctx, r.db).Create(treatment).Error; err != nil {
		return apperrors.FromGORM(err, "treatment", "")
	}
	return nil
}

// Update は dbOrTx で ambient tx に参加する clinic-scoped な部分更新（Create と同じ理由、BE9-2D ④b）。
// cmd.Version が非 nil の場合は expectedVersion として
// WHERE version=? 述語を追加する（UAT-R2-EXCLUSIVE-LOCK: 楽観ロックの原子化、
// clinical_plan_repository.go の Update と同型）。成功時は version+1 を書き戻す。
// cmd.Version が nil の場合は従来どおり version 述語なし（照合スキップ、後方互換）。
// RowsAffected==0 は「存在しない/他クリニック」と「バージョン不一致」を区別するため
// existsInClinic で再照会し正しいエラー種別に正規化する
// （clinical_plan_repository.go の conflictAfterZeroClinicalPlanRows と同型。
// 親カルテの draft ガードは service 側の lockDraftMedicalRecord FOR UPDATE が担うため
// ここでは行の存否のみを見る）。
func (r *treatmentRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateTreatmentInput) error {
	return r.update(ctx, clinicID, id, buildTreatmentUpdate(&cmd), cmd.Version)
}

func (r *treatmentRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any, expectedVersion *int) error {
	// NOTE: GORM does not propagate Joins() into the generated UPDATE statement's SQL
	// (it is a SELECT-only clause), so a WHERE referencing the joined table fails with
	// "missing FROM-clause entry". clinic_id isolation must be expressed as a subquery instead.
	q := persistence.DBOrTx(ctx, r.db).
		Model(&model.Treatment{}).
		Where("treatments.id = ? AND treatments.deleted_at IS NULL AND treatments.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID)
	if expectedVersion != nil {
		q = q.Where("treatments.version = ?", *expectedVersion)
	}
	fields["version"] = gorm.Expr("version + 1")
	result := q.Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.conflictAfterZeroTreatmentRows(ctx, clinicID, id, expectedVersion)
	}
	return nil
}

// existsInClinic は id の treatment が clinicID 配下（親カルテの draft/finalized を問わず）に
// 存在するかを返す。Update の RowsAffected==0 を「存在しない」と「バージョン不一致」で
// 区別するために使う（clinical_plan_repository.go の existsInClinic と同型）。
func (r *treatmentRepository) existsInClinic(ctx context.Context, clinicID, id uint64) (bool, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Treatment{}).
		Where("treatments.id = ? AND treatments.deleted_at IS NULL AND treatments.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
		Count(&count).Error; err != nil {
		return false, apperrors.FromGORM(err, "treatment", fmt.Sprintf("%d", id))
	}
	return count > 0, nil
}

func (r *treatmentRepository) conflictAfterZeroTreatmentRows(ctx context.Context, clinicID, id uint64, expectedVersion *int) error {
	// version 述語なし（照合スキップ）で 0 行 = 行自体が存在しない → 従来どおり NotFound。
	if expectedVersion == nil {
		return apperrors.WrapNotFound("treatment", fmt.Sprintf("%d", id))
	}
	exists, err := r.existsInClinic(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.WrapNotFound("treatment", fmt.Sprintf("%d", id))
	}
	return apperrors.WrapConflict("他のユーザーがこの治療を変更しました。再読み込みしてください")
}

// Delete は dbOrTx で ambient tx に参加する（Create と同じ理由、BE9-2D ④b）。
func (r *treatmentRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	// NOTE: see Update — Joins() does not propagate into DELETE's SQL either.
	result := persistence.DBOrTx(ctx, r.db).
		Where("treatments.id = ? AND treatments.deleted_at IS NULL AND treatments.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
		Delete(&model.Treatment{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *treatmentRepository) BulkUpdateSortOrder(ctx context.Context, updates []TreatmentSortUpdate) error {
	if err := persistence.DBOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		for _, u := range updates {
			// MRD-01: RowsAffected 確認 + medical_record_id を service 施錠済み親に束縛。
			result := tx.Model(&model.Treatment{}).
				Where(
					"treatments.id = ? AND treatments.deleted_at IS NULL AND treatments.medical_record_id = ? AND treatments.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)",
					u.ID, u.MedicalRecordID, u.ClinicID,
				).
				Update("sort_order", u.SortOrder)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "treatment", fmt.Sprintf("%d", u.ID))
			}
			if result.RowsAffected == 0 {
				return apperrors.WrapNotFound("treatment", fmt.Sprintf("%d", u.ID))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "bulk update sort order")
	}
	return nil
}
