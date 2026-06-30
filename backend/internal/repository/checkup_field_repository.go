package repository

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// CheckupTypeFieldRepository は健診パッケージのフィールド定義マスタ（checkup_type_fields）アクセス。
type CheckupTypeFieldRepository interface {
	// FindByCheckupTypeID は指定 checkup_type の生存フィールド定義を sort_order 昇順で返す。
	FindByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error)
}

// CheckupFieldResultRepository は健診結果値（checkup_field_results）アクセス。
type CheckupFieldResultRepository interface {
	// FindByCheckupID は指定 checkup の結果値を sort_order 昇順で返す。
	FindByCheckupID(ctx context.Context, clinicID, checkupID uint64) ([]model.CheckupFieldResult, error)
	// FindByPetID は飼い主レポート用に pet 単位の健診結果を返す（生存 checkup 経由）。
	FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error)
	// ReplaceForCheckup は checkup_field_results を一括置換する（既存全削除→一括挿入をトランザクション内で実行）。
	ReplaceForCheckup(ctx context.Context, clinicID, checkupID uint64, results []model.CheckupFieldResult) ([]model.CheckupFieldResult, error)
}

type checkupTypeFieldRepository struct {
	db *gorm.DB
}

func NewCheckupTypeFieldRepository(db *gorm.DB) CheckupTypeFieldRepository {
	return &checkupTypeFieldRepository{db: db}
}

func (r *checkupTypeFieldRepository) FindByCheckupTypeID(ctx context.Context, clinicID, checkupTypeID uint64) ([]model.CheckupTypeField, error) {
	fields := make([]model.CheckupTypeField, 0)
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Where("checkup_type_id = ?", checkupTypeID).
		Order("sort_order ASC, id ASC").
		Find(&fields).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_type_field", fmt.Sprintf("checkup_type=%d", checkupTypeID))
	}
	return fields, nil
}

type checkupFieldResultRepository struct {
	db *gorm.DB
}

func NewCheckupFieldResultRepository(db *gorm.DB) CheckupFieldResultRepository {
	return &checkupFieldResultRepository{db: db}
}

func (r *checkupFieldResultRepository) FindByCheckupID(ctx context.Context, clinicID, checkupID uint64) ([]model.CheckupFieldResult, error) {
	results := make([]model.CheckupFieldResult, 0)
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		// P3.1: clinic-scoped マスタ Preload は clinic_id 述語必須。
		Preload("CheckupTypeField", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Where("checkup_id = ?", checkupID).
		Order("sort_order ASC, id ASC").
		Find(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_field_result", fmt.Sprintf("checkup=%d", checkupID))
	}
	return results, nil
}

func (r *checkupFieldResultRepository) FindByPetID(ctx context.Context, clinicID, petID uint64) ([]model.CheckupFieldResult, error) {
	results := make([]model.CheckupFieldResult, 0)
	// 結果は生存 checkup 経由で pet 単位に解決する。checkup_field_results.clinic_id と
	// checkups.clinic_id の双方を clinic でスコープし、クロステナント混入を防ぐ。
	err := r.db.WithContext(ctx).
		Joins("JOIN checkups ON checkups.id = checkup_field_results.checkup_id"+
			" AND checkups.clinic_id = ?"+
			" AND checkups.deleted_at IS NULL", clinicID).
		Where("checkup_field_results.clinic_id = ? AND checkups.pet_id = ?", clinicID, petID).
		// P3.1: clinic-scoped マスタ Preload は clinic_id 述語必須。
		Preload("CheckupTypeField", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		// 親 checkup と種別（飼い主レポートの日付・パッケージ名表示用）。
		Preload("Checkup", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Checkup.CheckupType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Order("checkups.date DESC, checkup_field_results.sort_order ASC").
		Find(&results).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "checkup_field_result", fmt.Sprintf("pet=%d", petID))
	}
	return results, nil
}

// ReplaceForCheckup は checkup_field_results を一括置換する（PUT セマンティクス）。
// 親 checkup の clinic 隔離はサービス層の FindByID で保証されている前提だが、
// トランザクション内でも clinic スコープ付きで再確認する（並行削除/clinic 越境防止）。
// results の CheckupID / ClinicID は本メソッド内で強制上書きする。
func (r *checkupFieldResultRepository) ReplaceForCheckup(ctx context.Context, clinicID, checkupID uint64, results []model.CheckupFieldResult) ([]model.CheckupFieldResult, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Checkup{}).
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", checkupID, clinicID).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "checkup", fmt.Sprintf("%d", checkupID))
		}
		if count == 0 {
			return apperrors.WrapNotFound("checkup", fmt.Sprintf("%d", checkupID))
		}

		// 既存結果を clinic スコープで全削除（CASCADE では消えないため明示削除）。
		if err := tx.Where("checkup_id = ? AND clinic_id = ?", checkupID, clinicID).
			Delete(&model.CheckupFieldResult{}).Error; err != nil {
			return apperrors.FromGORM(err, "checkup_field_result", fmt.Sprintf("checkup=%d", checkupID))
		}

		if len(results) == 0 {
			return nil
		}

		for i := range results {
			results[i].CheckupID = checkupID
			results[i].ClinicID = clinicID
			results[i].ID = 0 // 新規挿入のため auto increment に任せる
			// value_list は NOT NULL DEFAULT '{}'。nil の pq.StringArray は driver が
			// NULL を返すため、明示的に空配列へ正規化して GORM の default 省略挙動への依存を断つ。
			if results[i].ValueList == nil {
				results[i].ValueList = pq.StringArray{}
			}
		}
		if err := tx.Create(&results).Error; err != nil {
			return apperrors.FromGORM(err, "checkup_field_result", fmt.Sprintf("checkup=%d", checkupID))
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to replace checkup field results")
	}
	return r.FindByCheckupID(ctx, clinicID, checkupID)
}
