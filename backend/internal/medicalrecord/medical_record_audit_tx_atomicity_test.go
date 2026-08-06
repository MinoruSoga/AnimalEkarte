package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

var errMedicalRecordAuditWrite = errors.New("medical record audit write failed")

type failingMedicalRecordAuditLogger struct{}

func (*failingMedicalRecordAuditLogger) LogMedicalRecordChange(
	context.Context,
	uint64,
	*uint64,
	string,
	uint64,
	map[string]any,
	map[string]any,
) error {
	return errMedicalRecordAuditWrite
}

func (*failingMedicalRecordAuditLogger) LogAddendumCreate(
	context.Context,
	uint64,
	*uint64,
	uint64,
	uint64,
	*model.MedicalRecordAddendum,
) error {
	return errMedicalRecordAuditWrite
}

func (*failingMedicalRecordAuditLogger) LogEntryTx(context.Context, *AuditEntry) error {
	return errMedicalRecordAuditWrite
}

func TestMedicalRecordService_Update_FinalizeAuditFailureRollsBack(t *testing.T) {
	const clinicID = uint64(71001)
	db := setupMedicalRecordAddendumTestDB(t)
	record := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: "AUDIT-TX-FINALIZE",
		Date:     time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC),
		Status:   model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(record).Error)
	originalVersion := record.Version

	repo := NewMedicalRecordRepository(db)
	auditFailure := &failingMedicalRecordAuditLogger{}
	service := NewMedicalRecordServiceWithTxAudit(
		repo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		auditFailure,
		auditFailure,
		persistence.NewTransactor(db),
	)
	finalized := model.MedicalRecordStatusFinalized
	actorID := uint64(101)

	updated, err := service.Update(context.Background(), clinicID, record.ID, UpdateMedicalRecordInput{
		Status:  &finalized,
		ActorID: &actorID,
	})

	require.ErrorIs(t, err, errMedicalRecordAuditWrite)
	assert.Nil(t, updated)

	var persisted model.MedicalRecord
	require.NoError(t, db.WithContext(context.Background()).First(&persisted, record.ID).Error)
	assert.Equal(t, model.MedicalRecordStatusDraft, persisted.Status)
	assert.Equal(t, originalVersion, persisted.Version)
}

func TestMedicalRecordAddendumService_Create_AuditFailureRollsBack(t *testing.T) {
	const clinicID = uint64(71002)
	db := setupMedicalRecordAddendumTestDB(t)
	staff := makeDoctor(t, db, clinicID, "監査失敗テスト医師")
	record := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: "AUDIT-TX-ADDENDUM",
		Date:     time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC),
		Status:   model.MedicalRecordStatusFinalized,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(record).Error)

	service := NewMedicalRecordAddendumService(
		NewMedicalRecordAddendumRepository(db),
		NewMedicalRecordRepository(db),
		&failingMedicalRecordAuditLogger{},
		persistence.NewTransactor(db),
	)

	addendum, err := service.Create(context.Background(), clinicID, CreateMedicalRecordAddendumInput{
		MedicalRecordID: record.ID,
		AuthorUserID:    staff.ID,
		AfterText:       "訂正後の本文",
		Reason:          "監査失敗時のロールバック確認",
	})

	require.ErrorIs(t, err, errMedicalRecordAuditWrite)
	assert.Nil(t, addendum)

	var count int64
	require.NoError(t, db.WithContext(context.Background()).
		Model(&model.MedicalRecordAddendum{}).
		Where("medical_record_id = ? AND clinic_id = ?", record.ID, clinicID).
		Count(&count).Error)
	assert.Zero(t, count)
}

func TestMedicalRecordService_Create_AuditFailureRemainsBestEffort(t *testing.T) {
	const clinicID = uint64(71003)
	db := setupMedicalRecordAddendumTestDB(t)
	repo := NewMedicalRecordRepository(db)
	service := NewMedicalRecordServiceWithTxAudit(
		repo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&failingMedicalRecordAuditLogger{},
		nil,
		persistence.NewTransactor(db))
	draft := model.MedicalRecordStatusDraft

	created, err := service.Create(context.Background(), clinicID, &CreateMedicalRecordInput{
		RecordNo:  "AUDIT-BEST-EFFORT-CREATE",
		Date:      time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC),
		Status:    &draft,
		VisitType: model.VisitTypeRevisit,
	})

	require.NoError(t, err)
	require.NotNil(t, created)

	var persisted model.MedicalRecord
	require.NoError(t, db.WithContext(context.Background()).First(&persisted, created.ID).Error)
	assert.Equal(t, "AUDIT-BEST-EFFORT-CREATE", persisted.RecordNo)
}

func TestMedicalRecordService_Update_AuditFailureRemainsBestEffort(t *testing.T) {
	const clinicID = uint64(71004)
	db := setupMedicalRecordAddendumTestDB(t)
	record := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: "AUDIT-BEST-EFFORT-UPDATE",
		Date:     time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC),
		Status:   model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(record).Error)

	service := NewMedicalRecordServiceWithTxAudit(
		NewMedicalRecordRepository(db),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&failingMedicalRecordAuditLogger{},
		nil,
		persistence.NewTransactor(db))
	updatedDate := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	updated, err := service.Update(context.Background(), clinicID, record.ID, UpdateMedicalRecordInput{
		Date: &updatedDate,
	})

	require.NoError(t, err)
	require.NotNil(t, updated)

	var persisted model.MedicalRecord
	require.NoError(t, db.WithContext(context.Background()).First(&persisted, record.ID).Error)
	assert.True(t, persisted.Date.Equal(updatedDate))
}
