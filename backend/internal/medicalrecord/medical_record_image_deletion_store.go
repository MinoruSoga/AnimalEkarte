package medicalrecord

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/persistence"
)

type sqlMedicalRecordImageDeletionStore struct {
	db *gorm.DB
}

func newSQLMedicalRecordImageDeletionStore(db *gorm.DB) medicalRecordImageDeletionStore {
	return &sqlMedicalRecordImageDeletionStore{db: db}
}

type imageDeletionScan struct {
	ClinicID           uint64 `gorm:"column:clinic_id"`
	MedicalRecordID    uint64 `gorm:"column:medical_record_id"`
	ImageID            uint64 `gorm:"column:image_id"`
	ImageObjectKey     string `gorm:"column:image_object_key"`
	ThumbnailObjectKey string `gorm:"column:thumbnail_object_key"`
	StorageStatus      string `gorm:"column:storage_status"`
	AttemptCount       int    `gorm:"column:attempt_count"`
	LastError          string `gorm:"column:last_error"`
	ActorStaffID       uint64 `gorm:"column:actor_staff_id"`
}

func (s *sqlMedicalRecordImageDeletionStore) Insert(ctx context.Context, row medicalRecordImageDeletion) error {
	if s == nil || s.db == nil {
		return errors.New("medical record image deletion store is not configured")
	}
	err := persistence.DBOrTx(ctx, s.db).Exec(`
		INSERT INTO medical_record_image_deletions (
			clinic_id, medical_record_id, image_id, image_object_key, thumbnail_object_key,
			storage_status, attempt_count, last_error, actor_staff_id
		) VALUES (?, ?, ?, ?, ?, ?, 0, '', ?)
	`, row.ClinicID, row.MedicalRecordID, row.ImageID, row.ImageObjectKey, row.ThumbnailObjectKey,
		row.StorageStatus, row.ActorStaffID).Error
	if isMedicalRecordImageDeletionConflict(err) {
		return errMedicalRecordImageDeletionExists
	}
	return err
}

func (s *sqlMedicalRecordImageDeletionStore) FindByImageID(
	ctx context.Context,
	clinicID, imageID uint64,
) (*medicalRecordImageDeletion, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("medical record image deletion store is not configured")
	}
	rows := make([]imageDeletionScan, 0, 1)
	err := persistence.DBOrTx(ctx, s.db).Raw(`
		SELECT clinic_id, medical_record_id, image_id, image_object_key, thumbnail_object_key,
		       storage_status, attempt_count, last_error, actor_staff_id
		FROM medical_record_image_deletions
		WHERE clinic_id = ? AND image_id = ? AND deleted_at IS NULL
	`, clinicID, imageID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	row := rows[0]
	return &medicalRecordImageDeletion{
		ClinicID:           row.ClinicID,
		MedicalRecordID:    row.MedicalRecordID,
		ImageID:            row.ImageID,
		ImageObjectKey:     row.ImageObjectKey,
		ThumbnailObjectKey: row.ThumbnailObjectKey,
		StorageStatus:      row.StorageStatus,
		AttemptCount:       row.AttemptCount,
		LastError:          row.LastError,
		ActorStaffID:       row.ActorStaffID,
	}, nil
}

func (s *sqlMedicalRecordImageDeletionStore) UpdateStorageResult(
	ctx context.Context,
	clinicID, imageID uint64,
	status, lastError string,
) error {
	if s == nil || s.db == nil {
		return errors.New("medical record image deletion store is not configured")
	}
	return persistence.DBOrTx(ctx, s.db).Exec(`
		UPDATE medical_record_image_deletions
		SET storage_status = ?, last_error = ?, attempt_count = attempt_count + 1, updated_at = NOW()
		WHERE clinic_id = ? AND image_id = ? AND deleted_at IS NULL
	`, status, lastError, clinicID, imageID).Error
}

func isMedicalRecordImageDeletionConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
