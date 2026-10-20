// Package consultation owns consultations master data access.
package medicalrecord

// Moved from internal/repository/consultation (BE9-2D ⑥ Batch A roll-up・BE8-4 subpackage)。
// generic Repository/New は entity-specific 名へ改名（①⑤先例）— 外部は facade alias 経由で不変。

import (
	"context"
	"fmt"
	"sort"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// Repository is the data access interface for consultations.
type ConsultationRepository interface {
	FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error)
	Create(ctx context.Context, consultation *model.Consultation) error
	Update(ctx context.Context, clinicID, id uint64, cmd UpdateConsultationInput) (*model.Consultation, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	Reorder(ctx context.Context, clinicID uint64, ids []uint64) error
	CountUsageByConsultationID(ctx context.Context, clinicID, consultationID uint64) (int64, error)
	CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error)
}

type consultationRepositoryImpl struct{ db *gorm.DB }

// New constructs a Repository.
func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
	return &consultationRepositoryImpl{db: db}
}

func (r *consultationRepositoryImpl) FindAll(ctx context.Context, clinicID uint64) ([]model.Consultation, error) {
	consultations := make([]model.Consultation, 0)
	// G2F-11: vaccine/procedure と同型の master list safety Limit（unbounded Find 防止）。
	err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Order("sort_order ASC, name ASC").
		Limit(persistence.MaxMasterListRows).
		Find(&consultations).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "consultation", "")
	}
	return consultations, nil
}

func (r *consultationRepositoryImpl) FindByID(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	db := persistence.DBOrTx(ctx, r.db)
	return persistence.FindByIDScoped[model.Consultation](ctx, db, "consultation", clinicID, id)
}

// LockWriteRows locks target and requested parent before validation/mutation.
func (r *consultationRepositoryImpl) LockWriteRows(ctx context.Context, clinicID uint64, ids []uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return apperrors.WrapInternalServerError("master parent lock requires an ambient transaction")
	}
	// Sort again at the repository boundary; callers cannot invert the lock order.
	ordered := append([]uint64(nil), ids...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	for i, id := range ordered {
		if id == 0 {
			return apperrors.WrapInvalidInput("master write row id must be positive")
		}
		if i > 0 && id == ordered[i-1] {
			continue
		}
		db := persistence.DBOrTx(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"})
		if _, err := persistence.FindByIDScoped[model.Consultation](ctx, db, "consultation", clinicID, id); err != nil {
			return err
		}
	}
	return nil
}

// LockByIDForUpdate locks only the mutation target for an exact audit before-image.
func (r *consultationRepositoryImpl) LockByIDForUpdate(ctx context.Context, clinicID, id uint64) (*model.Consultation, error) {
	if persistence.TxFromContext(ctx) == nil {
		return nil, apperrors.WrapInternalServerError("master insurance lock requires an ambient transaction")
	}
	db := persistence.DBOrTx(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"})
	return persistence.FindByIDScoped[model.Consultation](ctx, db, "consultation", clinicID, id)
}

func (r *consultationRepositoryImpl) Create(ctx context.Context, consultation *model.Consultation) error {
	db := persistence.DBOrTx(ctx, r.db)
	wantActive := consultation.IsActive
	if err := db.Create(consultation).Error; err != nil {
		return apperrors.FromGORM(err, "consultation", "")
	}
	if !wantActive {
		if err := db.Model(consultation).Update("is_active", false).Error; err != nil {
			return apperrors.FromGORM(err, "consultation", fmt.Sprintf("%d", consultation.ID))
		}
		consultation.IsActive = false
	}
	return nil
}

// Update compares the editing-start version; the storage trigger owns increments.
func (r *consultationRepositoryImpl) Update(ctx context.Context, clinicID, id uint64, cmd UpdateConsultationInput) (*model.Consultation, error) {
	if cmd.Version < 1 {
		return nil, apperrors.WrapInvalidInput("positive editing-start version is required")
	}
	fields := buildConsultationUpdate(&cmd)
	if len(fields) == 0 {
		return nil, apperrors.WrapInvalidInput(errMsgAtLeastOneField)
	}
	result := persistence.DBOrTx(ctx, r.db).Model(&model.Consultation{}).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ? AND version = ?", id, cmd.Version).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "consultation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		if _, err := r.FindByID(ctx, clinicID, id); err != nil {
			return nil, err
		}
		return nil, apperrors.WrapConflict("編集後に変更されています。再読み込みしてから保存してください")
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *consultationRepositoryImpl) Delete(ctx context.Context, clinicID, id uint64) error {
	if persistence.TxFromContext(ctx) == nil {
		return persistence.NewTransactor(r.db).WithTx(ctx, func(txCtx context.Context) error { return r.Delete(txCtx, clinicID, id) })
	}
	// Acquire the target before starting the guarded mutation statement. A wait
	// must finish before its READ COMMITTED snapshot checks newly committed children.
	if err := r.LockWriteRows(ctx, clinicID, []uint64{id}); err != nil {
		return err
	}
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Where(`NOT EXISTS (
			SELECT 1 FROM consultations children
			WHERE children.parent_id = consultations.id
			  AND children.clinic_id = ?
			  AND children.deleted_at IS NULL
		)`, clinicID).
		Where(`NOT EXISTS (
			SELECT 1 FROM treatments
			JOIN medical_records ON medical_records.id = treatments.medical_record_id
			  AND medical_records.clinic_id = ?
			  AND medical_records.deleted_at IS NULL
			WHERE treatments.consultation_id = consultations.id
			  AND treatments.deleted_at IS NULL
		)`, clinicID).
		Delete(&model.Consultation{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "consultation", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return r.normalizeConsultationDeleteMiss(ctx, clinicID, id)
	}
	return nil
}

func (r *consultationRepositoryImpl) normalizeConsultationDeleteMiss(ctx context.Context, clinicID, id uint64) error {
	if _, err := r.FindByID(ctx, clinicID, id); err != nil {
		return err
	}
	childCount, err := r.CountChildrenByParentID(ctx, clinicID, id)
	if err != nil {
		return err
	}
	if childCount > 0 {
		return apperrors.WrapConflict("この診察項目にはサブ項目が登録されているため削除できません")
	}
	return apperrors.WrapConflict("この診察項目はカルテで使用中のため削除できません")
}

func (r *consultationRepositoryImpl) Reorder(ctx context.Context, clinicID uint64, ids []uint64) error {
	return persistence.ReorderByClinicID(ctx, r.db, &model.Consultation{}, "consultation", clinicID, ids, "sort_order")
}

// CountChildrenByParentID は指定した親 ID を持つ子診察項目の件数を返す。
// 親を削除する前に孤立子が残らないことを確認するために使用する。
func (r *consultationRepositoryImpl) CountChildrenByParentID(ctx context.Context, clinicID, parentID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Consultation{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "consultation", fmt.Sprintf("%d", parentID))
	}
	return count, nil
}

// CountUsageByConsultationID は診察マスタを参照している treatments の件数を返す（BUG-107）
// treatments テーブルに直接 clinic_id がないため medical_records を JOIN してテナント分離する
func (r *consultationRepositoryImpl) CountUsageByConsultationID(ctx context.Context, clinicID, consultationID uint64) (int64, error) {
	var count int64
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.Treatment{}).
		Scopes(persistence.MedicalRecordTenantScope("treatments", clinicID)).
		Where("treatments.consultation_id = ? AND treatments.deleted_at IS NULL", consultationID).
		Count(&count).Error; err != nil {
		return 0, apperrors.FromGORM(err, "treatment", "")
	}
	return count, nil
}
