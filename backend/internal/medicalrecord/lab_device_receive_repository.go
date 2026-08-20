package medicalrecord

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// LabDeviceReceiveRepository persists device jobs, waits, and station settings.
type LabDeviceReceiveRepository interface {
	FindJobByFingerprint(ctx context.Context, clinicID uint64, sourceType, fingerprint string) (*model.LabImportJob, error)
	FindJobByID(ctx context.Context, clinicID uint64, jobID uuid.UUID) (*model.LabImportJob, error)
	CreateJobWithItems(ctx context.Context, job *model.LabImportJob, items []model.LabImportJobItem) error
	UpdateJobPetID(ctx context.Context, clinicID uint64, jobID uuid.UUID, petID *uint64) (*model.LabImportJob, error)
	SaveJobItems(ctx context.Context, items []model.LabImportJobItem) error
	ListJobItems(ctx context.Context, clinicID uint64, jobIDs []uuid.UUID) (map[uuid.UUID][]model.LabImportJobItem, error)
	ListUnlinkedJobs(ctx context.Context, clinicID uint64, limit int) ([]model.LabImportJob, error)
	ListSavedJobs(ctx context.Context, clinicID uint64, limit int) ([]model.LabImportJob, error)
	ListReceivedJobs(ctx context.Context, clinicID uint64, since time.Time, limit int) ([]model.LabImportJob, error)
	FindActiveWait(ctx context.Context, clinicID uint64) (*model.LabDeviceWait, error)
	LockActiveWait(ctx context.Context, clinicID uint64) (*model.LabDeviceWait, error)
	InsertWait(ctx context.Context, wait *model.LabDeviceWait) error
	ClearWaitByID(ctx context.Context, clinicID, waitID uint64, at time.Time) error
	GetStation(ctx context.Context, clinicID uint64) (*model.LabDeviceStationSettings, error)
	UpsertStation(ctx context.Context, row *model.LabDeviceStationSettings) error
}

type labDeviceReceiveRepository struct{ db *gorm.DB }

// NewLabDeviceReceiveRepository initializes a LabDeviceReceiveRepository.
func NewLabDeviceReceiveRepository(db *gorm.DB) LabDeviceReceiveRepository {
	return &labDeviceReceiveRepository{db: db}
}

func (r *labDeviceReceiveRepository) q(ctx context.Context) *gorm.DB {
	return persistence.DBOrTx(ctx, r.db)
}

func (r *labDeviceReceiveRepository) FindJobByFingerprint(
	ctx context.Context,
	clinicID uint64,
	sourceType, fingerprint string,
) (*model.LabImportJob, error) {
	if fingerprint == "" {
		return nil, nil
	}
	var job model.LabImportJob
	err := r.q(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("source_type = ? AND source_fingerprint = ?", sourceType, fingerprint).
		Take(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "find lab device job by fingerprint")
	}
	return &job, nil
}

func (r *labDeviceReceiveRepository) FindJobByID(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
) (*model.LabImportJob, error) {
	var job model.LabImportJob
	err := r.q(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", jobID).
		Take(&job).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.WrapNotFound("lab_import_job", jobID.String())
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "find lab device job")
	}
	return &job, nil
}

func (r *labDeviceReceiveRepository) CreateJobWithItems(
	ctx context.Context,
	job *model.LabImportJob,
	items []model.LabImportJobItem,
) error {
	db := r.q(ctx)
	if err := db.Create(job).Error; err != nil {
		if apperrors.IsConflict(apperrors.FromGORM(err, "lab_import_job", "")) {
			return apperrors.WrapConflict("同じ検査フレームは既に取込済みです")
		}
		return apperrors.Wrap(err, "create lab device job")
	}
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		items[i].ClinicID = job.ClinicID
		items[i].JobID = job.ID
	}
	if err := db.Create(&items).Error; err != nil {
		return apperrors.Wrap(err, "create lab device job items")
	}
	return nil
}

func (r *labDeviceReceiveRepository) UpdateJobPetID(
	ctx context.Context,
	clinicID uint64,
	jobID uuid.UUID,
	petID *uint64,
) (*model.LabImportJob, error) {
	res := r.q(ctx).
		Model(&model.LabImportJob{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ?", jobID).
		Update("pet_id", petID)
	if res.Error != nil {
		return nil, apperrors.Wrap(res.Error, "update lab device job pet")
	}
	if res.RowsAffected == 0 {
		return nil, apperrors.WrapNotFound("lab_import_job", jobID.String())
	}
	return r.FindJobByID(ctx, clinicID, jobID)
}

func (r *labDeviceReceiveRepository) SaveJobItems(ctx context.Context, items []model.LabImportJobItem) error {
	for i := range items {
		if err := r.q(ctx).Save(&items[i]).Error; err != nil {
			return apperrors.Wrap(err, "save lab device job item")
		}
	}
	return nil
}

func (r *labDeviceReceiveRepository) ListJobItems(
	ctx context.Context,
	clinicID uint64,
	jobIDs []uuid.UUID,
) (map[uuid.UUID][]model.LabImportJobItem, error) {
	out := make(map[uuid.UUID][]model.LabImportJobItem, len(jobIDs))
	if len(jobIDs) == 0 {
		return out, nil
	}
	rows := make([]model.LabImportJobItem, 0)
	if err := r.q(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("job_id IN ?", jobIDs).
		Order("job_id ASC, sort_order ASC, id ASC").
		Find(&rows).Error; err != nil {
		return nil, apperrors.Wrap(err, "list lab device job items")
	}
	for i := range rows {
		out[rows[i].JobID] = append(out[rows[i].JobID], rows[i])
	}
	return out, nil
}

func (r *labDeviceReceiveRepository) ListUnlinkedJobs(
	ctx context.Context,
	clinicID uint64,
	limit int,
) ([]model.LabImportJob, error) {
	return r.listDeviceJobs(ctx, clinicID, limit, true)
}

func (r *labDeviceReceiveRepository) ListSavedJobs(
	ctx context.Context,
	clinicID uint64,
	limit int,
) ([]model.LabImportJob, error) {
	return r.listDeviceJobs(ctx, clinicID, limit, false)
}

func (r *labDeviceReceiveRepository) ListReceivedJobs(
	ctx context.Context,
	clinicID uint64,
	since time.Time,
	limit int,
) ([]model.LabImportJob, error) {
	rows := make([]model.LabImportJob, 0)
	err := r.q(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("source_type IN ?", labDeviceSourceTypes()).
		Where("COALESCE(received_at, created_at) >= ?", since).
		Order("COALESCE(received_at, created_at) DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, apperrors.Wrap(err, "list received lab device jobs")
	}
	return rows, nil
}

func (r *labDeviceReceiveRepository) listDeviceJobs(
	ctx context.Context,
	clinicID uint64,
	limit int,
	unlinked bool,
) ([]model.LabImportJob, error) {
	rows := make([]model.LabImportJob, 0)
	q := r.q(ctx).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("source_type IN ?", labDeviceSourceTypes())
	if unlinked {
		q = q.Where("pet_id IS NULL")
	} else {
		q = q.Where("pet_id IS NOT NULL")
	}
	if err := q.Order("COALESCE(received_at, created_at) DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, apperrors.Wrap(err, "list lab device jobs")
	}
	return rows, nil
}

func (r *labDeviceReceiveRepository) findActiveWait(
	ctx context.Context,
	clinicID uint64,
	forUpdate bool,
) (*model.LabDeviceWait, error) {
	var wait model.LabDeviceWait
	q := r.q(ctx).Scopes(persistence.ClinicScope(clinicID)).Where("cleared_at IS NULL")
	if forUpdate {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	err := q.Take(&wait).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "find lab device wait")
	}
	return &wait, nil
}

func (r *labDeviceReceiveRepository) FindActiveWait(
	ctx context.Context,
	clinicID uint64,
) (*model.LabDeviceWait, error) {
	return r.findActiveWait(ctx, clinicID, false)
}

func (r *labDeviceReceiveRepository) LockActiveWait(
	ctx context.Context,
	clinicID uint64,
) (*model.LabDeviceWait, error) {
	return r.findActiveWait(ctx, clinicID, true)
}

func (r *labDeviceReceiveRepository) InsertWait(ctx context.Context, wait *model.LabDeviceWait) error {
	if err := r.q(ctx).Create(wait).Error; err != nil {
		return apperrors.Wrap(err, "insert lab device wait")
	}
	return nil
}

func (r *labDeviceReceiveRepository) ClearWaitByID(
	ctx context.Context,
	clinicID, waitID uint64,
	at time.Time,
) error {
	res := r.q(ctx).
		Model(&model.LabDeviceWait{}).
		Scopes(persistence.ClinicScope(clinicID)).
		Where("id = ? AND cleared_at IS NULL", waitID).
		Update("cleared_at", at)
	if res.Error != nil {
		return apperrors.Wrap(res.Error, "clear lab device wait")
	}
	return nil
}

func (r *labDeviceReceiveRepository) GetStation(
	ctx context.Context,
	clinicID uint64,
) (*model.LabDeviceStationSettings, error) {
	var row model.LabDeviceStationSettings
	err := r.q(ctx).Where("clinic_id = ?", clinicID).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.Wrap(err, "get lab device station")
	}
	return &row, nil
}

func (r *labDeviceReceiveRepository) UpsertStation(
	ctx context.Context,
	row *model.LabDeviceStationSettings,
) error {
	if err := r.q(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "clinic_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"wait_ttl_seconds", "slots_json", "updated_at"}),
	}).Create(row).Error; err != nil {
		return apperrors.Wrap(err, "upsert lab device station")
	}
	return nil
}

func labDeviceSourceTypes() []string {
	return []string{
		string(model.LabImportSourceTypeFujiNX600),
		string(model.LabImportSourceTypeFujiAU10V),
		string(model.LabImportSourceTypeArkrayPU4010),
		string(model.LabImportSourceTypeIDEXXVetLab),
	}
}

func validLabDeviceSlotsJSON(raw string) error {
	if raw == "" {
		return apperrors.WrapInvalidInput("slots_json is required")
	}
	var slots []map[string]any
	if err := json.Unmarshal([]byte(raw), &slots); err != nil {
		return apperrors.WrapInvalidInput("slots_json is invalid")
	}
	if len(slots) == 0 || len(slots) > 8 {
		return apperrors.WrapInvalidInput("slots_json size is invalid")
	}
	for _, slot := range slots {
		source, _ := slot["source_type"].(string)
		if !isLabDeviceSourceType(source) {
			return apperrors.WrapInvalidInput("slots_json source_type is invalid")
		}
		if parity, ok := slot["parity"].(string); ok {
			switch parity {
			case "", "none", "even", "odd":
			default:
				return apperrors.WrapInvalidInput("slots_json parity is invalid")
			}
		}
	}
	return nil
}
