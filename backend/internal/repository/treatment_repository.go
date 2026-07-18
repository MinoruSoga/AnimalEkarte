package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// TreatmentSortUpdate は並び順一括更新に使う軽量DTO
type TreatmentSortUpdate struct {
	ID        uint64
	ClinicID  uint64
	SortOrder int
}

// TreatmentRepository は治療項目の永続化インターフェース
type TreatmentRepository interface {
	FindByMedicalRecordID(ctx context.Context, clinicID, medicalRecordID uint64) ([]model.Treatment, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error)
	FindUnbilledByPetID(ctx context.Context, clinicID, petID uint64) ([]model.Treatment, error)
	// FindHistoryByPetID は #158/#159 飼主レポート用: ペット単位で treatments を横断取得する。
	// medical_records JOIN で clinic_id 隔離し、medical_records.date 降順で返す。
	// filter.AnesthesiaOnly=true / filter.IsSurgery=true 指定時は procedures JOIN で絞り込む。
	FindHistoryByPetID(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error)
	// CountFinalizedUnconfirmedByPetAndDate は同日同ペットの「未会計対象化」診察カルテ件数を返す(#77)。
	// finalized だが billing_confirmation 未confirmed かつ未会計のカルテ = 取り残し候補。
	CountFinalizedUnconfirmedByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error)
	Create(ctx context.Context, treatment *model.Treatment) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error
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
		Preload("Consultation", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Procedure", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Medicine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Order("treatments.sort_order ASC").
		Find(&treatments).Error; err != nil {
		return nil, apperrors.FromGORM(err, "treatment", "")
	}
	return treatments, nil
}

func (r *treatmentRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Treatment, error) {
	var treatment model.Treatment
	err := r.db.WithContext(ctx).
		Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id AND medical_records.deleted_at IS NULL").
		Where("medical_records.clinic_id = ? AND treatments.id = ? AND treatments.deleted_at IS NULL", clinicID, id).
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
		Where("NOT EXISTS (SELECT 1 FROM billing_items bi JOIN billings b ON b.id = bi.billing_id AND b.deleted_at IS NULL WHERE bi.treatment_id = treatments.id AND bi.deleted_at IS NULL AND b.status != 'cancelled')").
		Where("NOT EXISTS (SELECT 1 FROM billings b WHERE b.medical_record_id = mr.id AND b.status != 'cancelled' AND b.deleted_at IS NULL)").
		Order("treatments.sort_order ASC, treatments.id ASC").
		Find(&treatments).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "treatment", "")
	}
	return treatments, nil
}

func (r *treatmentRepository) FindHistoryByPetID(ctx context.Context, clinicID, petID uint64, filter model.PetTreatmentHistoryFilter, page, limit int) ([]model.Treatment, int64, error) {
	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).
			Model(&model.Treatment{}).
			Joins("JOIN medical_records ON medical_records.id = treatments.medical_record_id AND medical_records.deleted_at IS NULL").
			Where("medical_records.clinic_id = ? AND medical_records.pet_id = ? AND treatments.deleted_at IS NULL", clinicID, petID)
		if filter.ItemType != nil {
			q = q.Where("treatments.item_type = ?", *filter.ItemType)
		}
		// #159: 処置 JOIN フィルタ（procedure_id が NULL 以外の行のみ通る INNER JOIN で暗黙 item_type 絞り込み）
		if filter.AnesthesiaOnly || filter.IsSurgery {
			q = q.Joins("JOIN procedures ON procedures.id = treatments.procedure_id AND procedures.deleted_at IS NULL")
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
		Preload("MedicalRecord", "deleted_at IS NULL").
		Preload("Procedure", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Medicine", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Order("medical_records.date DESC, treatments.sort_order ASC, treatments.id DESC").
		Scopes(paginate(page, limit)).
		Find(&treatments).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "treatment", fmt.Sprintf("pet=%d", petID))
	}
	return treatments, total, nil
}

func (r *treatmentRepository) CountFinalizedUnconfirmedByPetAndDate(ctx context.Context, clinicID, petID uint64, date time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.MedicalRecord{}).
		Joins("LEFT JOIN billing_confirmations bc ON bc.medical_record_id = medical_records.id").
		Where("medical_records.clinic_id = ? AND medical_records.pet_id = ? AND medical_records.deleted_at IS NULL", clinicID, petID).
		Where("medical_records.status = ?", model.MedicalRecordStatusFinalized).
		Where("DATE(medical_records.date) = DATE(?)", date).
		Where("(bc.id IS NULL OR bc.status != 'confirmed')").
		Where("NOT EXISTS (SELECT 1 FROM billings b WHERE b.medical_record_id = medical_records.id AND b.status != 'cancelled' AND b.deleted_at IS NULL)").
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "medical_record", fmt.Sprintf("clinic=%d pet=%d", clinicID, petID))
	}
	return count, nil
}

func (r *treatmentRepository) Create(ctx context.Context, treatment *model.Treatment) error {
	if err := r.db.WithContext(ctx).Create(treatment).Error; err != nil {
		return apperrors.FromGORM(err, "treatment", "")
	}
	return nil
}

func (r *treatmentRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) error {
	// NOTE: GORM does not propagate Joins() into the generated UPDATE statement's SQL
	// (it is a SELECT-only clause), so a WHERE referencing the joined table fails with
	// "missing FROM-clause entry". clinic_id isolation must be expressed as a subquery instead.
	result := r.db.WithContext(ctx).
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

func (r *treatmentRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	// NOTE: see Update — Joins() does not propagate into DELETE's SQL either.
	result := r.db.WithContext(ctx).
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
	if err := dbOrTx(ctx, r.db).Transaction(func(tx *gorm.DB) error {
		for _, u := range updates {
			result := tx.Model(&model.Treatment{}).
				Where("treatments.id = ? AND treatments.deleted_at IS NULL AND treatments.medical_record_id IN (SELECT id FROM medical_records WHERE clinic_id = ?)", u.ID, u.ClinicID).
				Update("sort_order", u.SortOrder)
			if result.Error != nil {
				return apperrors.FromGORM(result.Error, "treatment", fmt.Sprintf("%d", u.ID))
			}
		}
		return nil
	}); err != nil {
		return apperrors.Wrap(err, "bulk update sort order")
	}
	return nil
}
