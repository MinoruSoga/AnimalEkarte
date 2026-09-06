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
	FindUnbilledByPetID(ctx context.Context, clinicID, petID uint64) ([]model.Treatment, error)
	// FindHistoryByPetID は #158/#159 飼主レポート用: ペット単位で treatments を横断取得する。
	// medical_records JOIN で clinic_id 隔離し、medical_records.date 降順で返す。
	// filter.AnesthesiaOnly=true / filter.IsSurgery=true 指定時は procedures JOIN で絞り込む。
	FindHistoryByPetID(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error)
	// CountFinalizedUnconfirmedByPetAndDate は同日同ペットの「未会計対象化」診察カルテ件数を返す(#77)。
	// finalized だが billing_confirmation 未confirmed かつ未会計のカルテ = 取り残し候補。
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

func (r *treatmentRepository) FindUnbilledByPetID(ctx context.Context, clinicID, petID uint64) ([]model.Treatment, error) {
	treatments := make([]model.Treatment, 0)
	err := r.db.WithContext(ctx).
		Joins("JOIN medical_records mr ON mr.id = treatments.medical_record_id AND mr.deleted_at IS NULL").
		Joins("JOIN billing_confirmations bc ON bc.medical_record_id = mr.id").
		Where("mr.pet_id = ? AND mr.clinic_id = ? AND bc.status = 'confirmed' AND treatments.deleted_at IS NULL", petID, clinicID).
		Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = mr.pet_id AND p.clinic_id = mr.clinic_id)").
		Where("NOT EXISTS (SELECT 1 FROM billing_items bi JOIN billings b ON b.id = bi.billing_id AND b.clinic_id = mr.clinic_id AND b.deleted_at IS NULL WHERE bi.treatment_id = treatments.id AND bi.deleted_at IS NULL AND b.status != 'cancelled')").
		Where("NOT EXISTS (SELECT 1 FROM billings b WHERE b.medical_record_id = mr.id AND b.clinic_id = mr.clinic_id AND b.status != 'cancelled' AND b.deleted_at IS NULL)").
		Scopes(treatmentReadPreloads(clinicID)).
		Order("treatments.sort_order ASC, treatments.id ASC").
		Find(&treatments).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment", "")
	}
	sanitizeTreatmentMasterRelations(treatments, clinicID)
	return treatments, nil
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

func (r *treatmentRepository) CountFinalizedUnconfirmedByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Joins("LEFT JOIN billing_confirmations bc ON bc.medical_record_id = medical_records.id").
		Where("medical_records.clinic_id = ? AND medical_records.pet_id = ? AND medical_records.deleted_at IS NULL", clinicID, petID).
		Where("EXISTS (SELECT 1 FROM pets p WHERE p.id = medical_records.pet_id AND p.clinic_id = medical_records.clinic_id)").
		Where("medical_records.status = ?", model.MedicalRecordStatusFinalized).
		Where("DATE(medical_records.date) = DATE(?)", date).
		Where("(bc.id IS NULL OR bc.status != 'confirmed')").
		Where("NOT EXISTS (SELECT 1 FROM billings b WHERE b.medical_record_id = medical_records.id AND b.clinic_id = medical_records.clinic_id AND b.status != 'cancelled' AND b.deleted_at IS NULL)").
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("clinic=%d pet=%d", clinicID, petID))
	}
	return count, nil
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

// Update は dbOrTx で ambient tx に参加する（Create と同じ理由、BE9-2D ④b）。
func (r *treatmentRepository) Update(ctx context.Context, clinicID, id uint64, cmd UpdateTreatmentInput) error {
	return r.update(ctx, clinicID, id, buildTreatmentUpdate(&cmd))
}

func (r *treatmentRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	// NOTE: GORM does not propagate Joins() into the generated UPDATE statement's SQL
	// (it is a SELECT-only clause), so a WHERE referencing the joined table fails with
	// "missing FROM-clause entry". clinic_id isolation must be expressed as a subquery instead.
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.Treatment{}).
		Where("treatments.id = ? AND treatments.deleted_at IS NULL AND treatments.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)", id, clinicID).
		Updates(fields)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "treatment", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("treatment", fmt.Sprintf("%d", id))
	}
	return nil
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
