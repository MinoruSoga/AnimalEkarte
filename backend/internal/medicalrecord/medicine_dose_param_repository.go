// Package medicalrecord provides medicine dose param persistence.
package medicalrecord

// Moved from internal/repository (BE9-2D ⑥ Batch A). 旧 package-private helper は repohelpers
// 同等物へ置換（同一述語/ambient-tx参加）。外部は internal/repository の facade alias 経由で不変。

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// MedicineDoseParamRepository は薬剤 × 種の投与量パラメータの永続化（#201）。
// 子テーブルに clinic_id を非正規化保持し clinicScope を直適用する（JOIN スコープは base.go で不可）。
type MedicineDoseParamRepository interface {
	FindByMedicineID(ctx context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error)
	FindByMedicineAndSpecies(ctx context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies) (*model.MedicineDoseParam, error)
	Create(ctx context.Context, clinicID uint64, param *model.MedicineDoseParam) error
	Update(ctx context.Context, clinicID, id uint64, cmd MedicineDoseParamInput) (*model.MedicineDoseParam, error)
	Delete(ctx context.Context, clinicID, id uint64) error
}

type medicineDoseParamRepository struct{ db *gorm.DB }

func NewMedicineDoseParamRepository(db *gorm.DB) MedicineDoseParamRepository {
	return &medicineDoseParamRepository{db: db}
}

func (r *medicineDoseParamRepository) FindByMedicineID(ctx context.Context, clinicID, medicineID uint64) ([]model.MedicineDoseParam, error) {
	params := make([]model.MedicineDoseParam, 0)
	if err := persistence.DBOrTx(ctx, r.db).
		Model(&model.MedicineDoseParam{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("medicine_id = ?", medicineID).
		Order("species ASC").
		Find(&params).Error; err != nil {
		return nil, apperrors.FromGORM(err, "medicine_dose_param", "")
	}
	return params, nil
}

// FindByMedicineAndSpecies は計算時の (薬剤, 種) ルックアップ。
// 該当パラメータがなければ NotFound を返す（呼び出し側は fail-closed で手動入力へ）。
func (r *medicineDoseParamRepository) FindByMedicineAndSpecies(ctx context.Context, clinicID, medicineID uint64, species model.MedicineDoseSpecies) (*model.MedicineDoseParam, error) {
	var param model.MedicineDoseParam
	if err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("medicine_id = ? AND species = ?", medicineID, species).
		First(&param).Error; err != nil {
		return nil, apperrors.FromGORM(err, "medicine_dose_param", fmt.Sprintf("%d/%s", medicineID, species))
	}
	return &param, nil
}

// Create は信頼できるスコープ由来の clinicID を param に強制設定してから INSERT する。
// 呼び出し側がセットした param.ClinicID は信頼せず上書きする（クロステナント書込防止）。
func (r *medicineDoseParamRepository) Create(ctx context.Context, clinicID uint64, param *model.MedicineDoseParam) error {
	param.ClinicID = clinicID
	if err := persistence.DBOrTx(ctx, r.db).Create(param).Error; err != nil {
		return apperrors.FromGORM(err, "medicine_dose_param", "")
	}
	return nil
}

func (r *medicineDoseParamRepository) Update(ctx context.Context, clinicID, id uint64, cmd MedicineDoseParamInput) (*model.MedicineDoseParam, error) {
	return r.update(ctx, clinicID, id, buildDoseParamReplaceFields(&cmd))
}

func (r *medicineDoseParamRepository) update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.MedicineDoseParam, error) {
	result := persistence.DBOrTx(ctx, r.db).
		Model(&model.MedicineDoseParam{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "medicine_dose_param", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("medicine_dose_param", fmt.Sprintf("%d", id))
	}
	var param model.MedicineDoseParam
	if err := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).Where("id = ?", id).First(&param).Error; err != nil {
		return nil, apperrors.FromGORM(err, "medicine_dose_param", fmt.Sprintf("%d", id))
	}
	return &param, nil
}

func (r *medicineDoseParamRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := persistence.DBOrTx(ctx, r.db).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", id).
		Delete(&model.MedicineDoseParam{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "medicine_dose_param", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("medicine_dose_param", fmt.Sprintf("%d", id))
	}
	return nil
}
