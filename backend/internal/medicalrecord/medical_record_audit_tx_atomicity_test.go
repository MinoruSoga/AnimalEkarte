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

func TestMedicalRecordAddendumService_Create_AuditFailureRollsBack(t *testing.T) {
	const clinicID = uint64(71002)
	db := setupMedicalRecordAddendumTestDB(t)
	staff := makeDoctor(t, db, clinicID, "監査失敗テスト医師")
	record := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: "AUDIT-TX-ADDENDUM",
		Date:     time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC),
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
	created, err := service.Create(context.Background(), clinicID, &CreateMedicalRecordInput{
		RecordNo:  "AUDIT-BEST-EFFORT-CREATE",
		Date:      time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC),
		VisitType: model.VisitTypeRevisit,
	})

	require.NoError(t, err)
	require.NotNil(t, created)

	var persisted model.MedicalRecord
	require.NoError(t, db.WithContext(context.Background()).First(&persisted, created.ID).Error)
	assert.Equal(t, "AUDIT-BEST-EFFORT-CREATE", persisted.RecordNo)
}

// EMR-495（監査 fail-closed）: 通常更新の監査は更新と同じ tx で書く。監査の書込に失敗したら更新ごと rollback し、
// エラーを返す（best-effort で成功扱いにしない）。create の監査は best-effort のまま。
func TestMedicalRecordService_Update_AuditFailureRollsBack(t *testing.T) {
	const clinicID = uint64(71004)
	db := setupMedicalRecordAddendumTestDB(t)
	originalDate := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
	record := &model.MedicalRecord{
		ClinicID: clinicID,
		RecordNo: "AUDIT-FAIL-CLOSED-UPDATE",
		Date:     originalDate,
	}
	require.NoError(t, db.WithContext(context.Background()).Create(record).Error)
	var before model.MedicalRecord
	require.NoError(t, db.WithContext(context.Background()).First(&before, record.ID).Error)

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
		nil,
		&failingMedicalRecordAuditLogger{},
		persistence.NewTransactor(db))
	updatedDate := time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC)

	updated, err := service.Update(context.Background(), clinicID, record.ID, UpdateMedicalRecordInput{
		Date: &updatedDate,
	})

	require.ErrorIs(t, err, errMedicalRecordAuditWrite)
	assert.Nil(t, updated)

	var persisted model.MedicalRecord
	require.NoError(t, db.WithContext(context.Background()).First(&persisted, record.ID).Error)
	assert.True(t, persisted.Date.Equal(originalDate), "audit failure must roll back the update")
	assert.Equal(t, before.Version, persisted.Version, "version must not advance when the update is rolled back")
}
