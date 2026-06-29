package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

type ExaminationRepository interface {
	FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error)
	FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error)
	// FindByJobID は clinic_id + job_id で絞り込んだ exams を日付降順で返す（Phase 4B.2）。
	FindByJobID(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.Examination, error)
	Create(ctx context.Context, exam *model.Examination) error
	Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error)
	Delete(ctx context.Context, clinicID, id uint64) error
	CountItemsByExamID(ctx context.Context, clinicID, examID uint64) (int64, error)
	FindAllItemsByExamID(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error)
	ReplaceItemsByExamID(ctx context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, error)
}

type examinationRepository struct {
	db *gorm.DB
}

func NewExaminationRepository(db *gorm.DB) ExaminationRepository {
	return &examinationRepository{db: db}
}

func (r *examinationRepository) FindAll(ctx context.Context, clinicID uint64, petID, ownerID *uint64, status, startDate, endDate *string, page, limit int) ([]model.Examination, int64, error) {
	buildBase := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&model.Examination{}).
			Where("exams.clinic_id = ?", clinicID)
		if petID != nil {
			q = q.Where("exams.pet_id = ?", *petID)
		}
		if ownerID != nil {
			q = q.Joins("JOIN pets ON pets.id = exams.pet_id AND pets.deleted_at IS NULL").Where("pets.owner_id = ?", *ownerID)
		}
		if status != nil {
			q = q.Where("exams.status = ?", *status)
		}
		if startDate != nil {
			q = q.Where("exams.date >= ?", *startDate)
		}
		if endDate != nil {
			q = q.Where("exams.date <= ?", *endDate)
		}
		return q
	}

	var total int64
	if err := buildBase().Count(&total).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "exam", "")
	}

	exams := make([]model.Examination, 0)
	if err := buildBase().Preload("ExaminationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).Preload("Pet", "deleted_at IS NULL").Preload("Pet.Owner", "deleted_at IS NULL").Preload("Doctor", "deleted_at IS NULL").Preload("Items").
		Offset((page - 1) * limit).Limit(limit).Order("exams.date DESC, exams.created_at DESC").
		Find(&exams).Error; err != nil {
		return nil, 0, apperrors.FromGORM(err, "exam", "")
	}
	return exams, total, nil
}

func (r *examinationRepository) FindByID(ctx context.Context, clinicID, id uint64) (*model.Examination, error) {
	var exam model.Examination
	err := r.db.WithContext(ctx).
		Where("exams.id = ? AND exams.clinic_id = ?", id, clinicID).
		Preload("ExaminationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).Preload("Pet", "deleted_at IS NULL").Preload("Pet.Owner", "deleted_at IS NULL").Preload("Doctor", "deleted_at IS NULL").Preload("Items").
		First(&exam).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", id))
	}
	return &exam, nil
}

// FindByJobID は clinic_id + job_id scope で exams を日付降順・作成日時降順で返す。
// exam_results と exam_type を Preload する（report DTO 組み立てに必要）。
// clinic scope は WHERE exams.clinic_id = ? で保証する（exam_results は clinic_id を持たないため JOIN で隔離）。
func (r *examinationRepository) FindByJobID(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.Examination, error) {
	var exams []*model.Examination
	err := r.db.WithContext(ctx).
		Where("exams.clinic_id = ? AND exams.job_id = ?", clinicID, jobID).
		Preload("ExaminationType", "clinic_id = ? AND deleted_at IS NULL", clinicID).
		Preload("Items").
		Order("exams.date DESC, exams.created_at DESC").
		Find(&exams).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam", fmt.Sprintf("job_id=%s", jobID))
	}
	return exams, nil
}

func (r *examinationRepository) Create(ctx context.Context, exam *model.Examination) error {
	err := r.db.WithContext(ctx).Create(exam).Error
	if err != nil {
		return apperrors.FromGORM(err, "exam", "")
	}
	return nil
}

func (r *examinationRepository) Update(ctx context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error) {
	result := r.db.WithContext(ctx).
		Model(&model.Examination{}).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return nil, apperrors.FromGORM(result.Error, "exam", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("exam", fmt.Sprintf("%d", id))
	}
	return r.FindByID(ctx, clinicID, id)
}

func (r *examinationRepository) Delete(ctx context.Context, clinicID, id uint64) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).Where("id = ?", id).
		Delete(&model.Examination{})
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "exam", fmt.Sprintf("%d", id))
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("exam", fmt.Sprintf("%d", id))
	}
	return nil
}

func (r *examinationRepository) CountItemsByExamID(ctx context.Context, clinicID, examID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ExamResult{}).
		Joins("JOIN exams ON exam_results.exam_id = exams.id AND exams.deleted_at IS NULL").
		Where("exams.clinic_id = ? AND exam_results.exam_id = ? ", clinicID, examID).
		Count(&count).Error
	if err != nil {
		return 0, apperrors.FromGORM(err, "exam_item", fmt.Sprintf("exam_id=%d", examID))
	}
	return count, nil
}

// FindAllItemsByExamID は exam_results を sort_order ASC で取得する。
// clinic_id 隔離は exams を JOIN して保証する（exam_results 自体は clinic_id を持たない）。
func (r *examinationRepository) FindAllItemsByExamID(ctx context.Context, clinicID, examID uint64) ([]model.ExamResult, error) {
	items := make([]model.ExamResult, 0)
	err := r.db.WithContext(ctx).
		Model(&model.ExamResult{}).
		Joins("JOIN exams ON exam_results.exam_id = exams.id AND exams.deleted_at IS NULL").
		Where("exams.clinic_id = ? AND exam_results.exam_id = ?", clinicID, examID).
		Order("exam_results.sort_order ASC, exam_results.id ASC").
		Find(&items).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "exam_item", fmt.Sprintf("exam_id=%d", examID))
	}
	return items, nil
}

// ReplaceItemsByExamID は exam_results を一括置換する（既存全削除→一括挿入をトランザクション内で実行）。
// 親 exam の clinic_id 隔離はサービス層の FindByID で保証されている前提。
// items の ExamID は本メソッド内で examID に強制上書きする（呼び出し側の指定ミスを防ぐ）。
func (r *examinationRepository) ReplaceItemsByExamID(ctx context.Context, clinicID, examID uint64, items []model.ExamResult) ([]model.ExamResult, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 親 exam の存在 + clinic 隔離をトランザクション内で再確認（並行削除/clinic 越境を防ぐ）
		var count int64
		if err := tx.Model(&model.Examination{}).
			Where("id = ? AND clinic_id = ? AND deleted_at IS NULL", examID, clinicID).
			Count(&count).Error; err != nil {
			return apperrors.FromGORM(err, "exam", fmt.Sprintf("%d", examID))
		}
		if count == 0 {
			return apperrors.WrapNotFound("exam", fmt.Sprintf("%d", examID))
		}

		// 既存 items を全削除（exam_results は CASCADE では消えないため明示削除）
		if err := tx.Where("exam_id = ?", examID).Delete(&model.ExamResult{}).Error; err != nil {
			return apperrors.FromGORM(err, "exam_item", fmt.Sprintf("exam_id=%d", examID))
		}

		if len(items) == 0 {
			return nil
		}

		// ExamID を強制上書きして一括挿入
		for i := range items {
			items[i].ExamID = examID
			items[i].ID = 0 // 新規挿入のため auto increment に任せる
		}
		if err := tx.Create(&items).Error; err != nil {
			return apperrors.FromGORM(err, "exam_item", fmt.Sprintf("exam_id=%d", examID))
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to replace exam items")
	}
	return r.FindAllItemsByExamID(ctx, clinicID, examID)
}
