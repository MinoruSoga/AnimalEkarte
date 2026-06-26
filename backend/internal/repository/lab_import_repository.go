package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// LabImportJobRepository は lab_import_jobs の永続化インターフェース。
type LabImportJobRepository interface {
	// Create は新規インポートジョブを作成する。
	Create(ctx context.Context, job *model.LabImportJob) error
	// Update はジョブ状態を保存する（clinic_id スコープ）。
	Update(ctx context.Context, job *model.LabImportJob) error
	// FindByID はクリニックスコープで ID に一致するジョブを返す。
	FindByID(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LabImportJob, error)
	// FindByClinic はクリニックスコープで最新順にジョブ一覧を返す。
	FindByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LabImportJob, error)
}

// LabImportEventRepository は lab_import_events の永続化インターフェース。
type LabImportEventRepository interface {
	// Create は監査イベントを追記する。
	Create(ctx context.Context, event *model.LabImportEvent) error
	// FindByJob はジョブ ID に紐づくイベントを時系列昇順で返す。
	FindByJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error)
}

// ------------------------------------
// LabImportJobRepository 実装
// ------------------------------------

type labImportJobRepository struct{ db *gorm.DB }

// NewLabImportJobRepository は LabImportJobRepository を初期化して返す。
func NewLabImportJobRepository(db *gorm.DB) LabImportJobRepository {
	return &labImportJobRepository{db: db}
}

func (r *labImportJobRepository) Create(ctx context.Context, job *model.LabImportJob) error {
	if err := r.db.WithContext(ctx).Create(job).Error; err != nil {
		return apperrors.FromGORM(err, "lab_import_job", "")
	}
	return nil
}

func (r *labImportJobRepository) Update(ctx context.Context, job *model.LabImportJob) error {
	result := r.db.WithContext(ctx).
		Scopes(clinicScope(job.ClinicID)).
		Save(job)
	if result.Error != nil {
		return apperrors.FromGORM(result.Error, "lab_import_job", job.ID.String())
	}
	if result.RowsAffected == 0 {
		return apperrors.WrapNotFound("lab_import_job", job.ID.String())
	}
	return nil
}

func (r *labImportJobRepository) FindByID(ctx context.Context, clinicID uint64, id uuid.UUID) (*model.LabImportJob, error) {
	var job model.LabImportJob
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND id = ?", clinicID, id).
		First(&job).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_job", id.String())
	}
	return &job, nil
}

func (r *labImportJobRepository) FindByClinic(ctx context.Context, clinicID uint64, limit int) ([]*model.LabImportJob, error) {
	var jobs []*model.LabImportJob
	err := r.db.WithContext(ctx).
		Scopes(clinicScope(clinicID)).
		Order("created_at DESC").
		Limit(limit).
		Find(&jobs).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_job", "")
	}
	return jobs, nil
}

// ------------------------------------
// LabImportEventRepository 実装
// ------------------------------------

type labImportEventRepository struct{ db *gorm.DB }

// NewLabImportEventRepository は LabImportEventRepository を初期化して返す。
func NewLabImportEventRepository(db *gorm.DB) LabImportEventRepository {
	return &labImportEventRepository{db: db}
}

func (r *labImportEventRepository) Create(ctx context.Context, event *model.LabImportEvent) error {
	if err := r.db.WithContext(ctx).Create(event).Error; err != nil {
		return apperrors.FromGORM(err, "lab_import_event", "")
	}
	return nil
}

func (r *labImportEventRepository) FindByJob(ctx context.Context, clinicID uint64, jobID uuid.UUID) ([]*model.LabImportEvent, error) {
	var events []*model.LabImportEvent
	err := r.db.WithContext(ctx).
		Where("clinic_id = ? AND job_id = ?", clinicID, jobID).
		Order("created_at ASC").
		Find(&events).Error
	if err != nil {
		return nil, apperrors.FromGORM(err, "lab_import_event", jobID.String())
	}
	return events, nil
}
