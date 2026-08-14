package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestDB_ExaminationService_CreateConfirmedWithItemsPersistsItemsBeforeStatusTransition(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	actorID := makeExaminationActor(t, db, clinicID, "TASK-027 confirm actor")
	examType := makeExamTypeMaster(t, db, clinicID, "TASK-026 initial confirm")
	repo := NewExaminationRepository(db)
	audit := &mockAuditTxLogger{logEntryTxFn: func(txCtx context.Context, _ *AuditEntry) error {
		require.NotNil(t, persistence.TxFromContext(txCtx), "parent audit must receive the examination ambient transaction")
		return nil
	}}
	svc := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		audit,
		testTransactor{db: db},
	)
	items := []UpsertExamItemInput{{Name: "WBC", InspectionValue: "5.0"}}

	exam, err := svc.Create(ctx, clinicID, &CreateExaminationInput{
		ExamTypeID: examType.ID,
		Date:       time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
		Status:     model.ExaminationStatusConfirmed,
		Items:      &items,
		ActorID:    ptrUint64(actorID),
	})

	require.NoError(t, err)
	require.NotNil(t, exam)
	persisted, err := repo.FindByID(ctx, clinicID, exam.ID)
	require.NoError(t, err)
	assert.Equal(t, model.ExaminationStatusConfirmed, persisted.Status)
	require.NotNil(t, persisted.CurrentRevisionVersion)
	assert.Equal(t, initialExaminationRevisionVersion, *persisted.CurrentRevisionVersion)
	saved, err := repo.FindAllItemsByExamID(ctx, clinicID, exam.ID)
	require.NoError(t, err)
	require.Len(t, saved, 1)
	assert.Equal(t, "WBC", saved[0].Name)
	var revisionCount int64
	require.NoError(t, db.Model(&model.ExaminationRevision{}).
		Where("clinic_id = ? AND examination_id = ?", clinicID, exam.ID).
		Count(&revisionCount).Error)
	assert.Equal(t, int64(1), revisionCount)
}

func TestDB_ExaminationService_CreateAuditUsesPersistedParentSnapshot(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const (
		clinicID = uint64(1)
		actorID  = uint64(42)
	)
	examType := makeExamTypeMaster(t, db, clinicID, "TASK-026 persisted audit snapshot")
	repo := NewExaminationRepository(db)
	var auditEntry *AuditEntry
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, entry *AuditEntry) error {
		auditEntry = entry
		return nil
	}}
	svc := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		audit,
		testTransactor{db: db},
	)
	inputDate := time.Date(2026, time.August, 1, 15, 45, 0, 0, time.FixedZone("JST", 9*60*60))

	created, err := svc.Create(ctx, clinicID, &CreateExaminationInput{
		ExamTypeID: examType.ID,
		Date:       inputDate,
		ActorID:    ptrUint64(actorID),
	})

	require.NoError(t, err)
	require.NotNil(t, created)
	persisted, err := repo.FindByID(ctx, clinicID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, auditEntry)
	auditDate, ok := auditEntry.NewValue.(map[string]any)["date"].(time.Time)
	require.True(t, ok)
	assert.Equal(t, persisted.Date, created.Date)
	assert.Equal(t, persisted.Date, auditDate)
}

func TestDB_ExaminationService_ParentMutationAuditFailureRollsBack(t *testing.T) {
	errAudit := errors.New("TASK-026 parent audit failure")
	const (
		clinicID = uint64(1)
		actorID  = uint64(42)
	)

	t.Run("create", func(t *testing.T) {
		db := setupExaminationTestDB(t)
		ctx := context.Background()
		examType := makeExamTypeMaster(t, db, clinicID, "TASK-026 create rollback")
		repo := NewExaminationRepository(db)
		svc := newFailingAuditExaminationService(db, repo, errAudit)

		got, err := svc.Create(ctx, clinicID, &CreateExaminationInput{
			ExamTypeID: examType.ID,
			Date:       time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
			Machine:    "TASK-026-create-audit-rollback",
			ActorID:    ptrUint64(actorID),
		})

		assert.ErrorIs(t, err, errAudit)
		assert.Nil(t, got)
		var count int64
		require.NoError(t, db.Model(&model.Examination{}).
			Where("clinic_id = ? AND machine = ?", clinicID, "TASK-026-create-audit-rollback").
			Count(&count).Error)
		assert.Zero(t, count)
	})

	t.Run("update", func(t *testing.T) {
		db := setupExaminationTestDB(t)
		ctx := context.Background()
		examType := makeExamTypeMaster(t, db, clinicID, "TASK-026 update rollback")
		exam := makeExaminationRec(t, db, &model.Examination{
			ClinicID: clinicID, ExamTypeID: examType.ID,
			Date:          time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
			ResultSummary: "before", Status: model.ExaminationStatusPending,
		})
		repo := NewExaminationRepository(db)
		svc := newFailingAuditExaminationService(db, repo, errAudit)
		summary := "after"

		got, err := svc.Update(ctx, clinicID, exam.ID, UpdateExaminationInput{
			ResultSummary: &summary,
			ActorID:       ptrUint64(actorID),
		})

		assert.ErrorIs(t, err, errAudit)
		assert.Nil(t, got)
		persisted, findErr := repo.FindByID(ctx, clinicID, exam.ID)
		require.NoError(t, findErr)
		assert.Equal(t, "before", persisted.ResultSummary)
	})

	t.Run("confirm including new items", func(t *testing.T) {
		db := setupExaminationTestDB(t)
		ctx := context.Background()
		confirmedActorID := makeExaminationActor(t, db, clinicID, "TASK-027 rollback actor")
		examType := makeExamTypeMaster(t, db, clinicID, "TASK-026 confirm rollback")
		exam := makeExaminationRec(t, db, &model.Examination{
			ClinicID: clinicID, ExamTypeID: examType.ID,
			Date:   time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
			Status: model.ExaminationStatusPending,
		})
		repo := NewExaminationRepository(db)
		svc := newFailingAuditExaminationService(db, repo, errAudit)
		confirmed := model.ExaminationStatusConfirmed
		items := []UpsertExamItemInput{{Name: "Glucose", InspectionValue: "90"}}

		got, err := svc.Update(ctx, clinicID, exam.ID, UpdateExaminationInput{
			Status:  &confirmed,
			Items:   &items,
			ActorID: ptrUint64(confirmedActorID),
		})

		assert.ErrorIs(t, err, errAudit)
		assert.Nil(t, got)
		persisted, findErr := repo.FindByID(ctx, clinicID, exam.ID)
		require.NoError(t, findErr)
		assert.Equal(t, model.ExaminationStatusPending, persisted.Status)
		assert.Nil(t, persisted.CurrentRevisionVersion)
		saved, itemsErr := repo.FindAllItemsByExamID(ctx, clinicID, exam.ID)
		require.NoError(t, itemsErr)
		assert.Empty(t, saved)
		var revisionCount, revisionItemCount int64
		require.NoError(t, db.Model(&model.ExaminationRevision{}).
			Where("clinic_id = ? AND examination_id = ?", clinicID, exam.ID).
			Count(&revisionCount).Error)
		require.NoError(t, db.Model(&model.ExaminationRevisionItem{}).
			Where("clinic_id = ? AND examination_id = ?", clinicID, exam.ID).
			Count(&revisionItemCount).Error)
		assert.Zero(t, revisionCount)
		assert.Zero(t, revisionItemCount)
	})

	t.Run("delete", func(t *testing.T) {
		db := setupExaminationTestDB(t)
		ctx := context.Background()
		examType := makeExamTypeMaster(t, db, clinicID, "TASK-026 delete rollback")
		exam := makeExaminationRec(t, db, &model.Examination{
			ClinicID: clinicID, ExamTypeID: examType.ID,
			Date:   time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
			Status: model.ExaminationStatusPending,
		})
		repo := NewExaminationRepository(db)
		svc := newFailingAuditExaminationService(db, repo, errAudit)

		err := svc.Delete(ctx, clinicID, exam.ID, ptrUint64(actorID))

		assert.ErrorIs(t, err, errAudit)
		persisted, findErr := repo.FindByID(ctx, clinicID, exam.ID)
		require.NoError(t, findErr)
		assert.Equal(t, exam.ID, persisted.ID)
	})
}

func TestDB_ExaminationService_ClinicIsolationPreventsMutationAndAudit(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const (
		clinicA = uint64(1)
		clinicB = uint64(2)
		actorID = uint64(42)
	)
	examType := makeExamTypeMaster(t, db, clinicA, "TASK-026 clinic isolation")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicA, ExamTypeID: examType.ID,
		Date:          time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
		ResultSummary: "clinic A", Status: model.ExaminationStatusPending,
	})
	repo := NewExaminationRepository(db)
	auditCalls := 0
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalls++
		return nil
	}}
	svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, NewExamTypeRepository(db), audit, testTransactor{db: db})
	summary := "foreign update"
	confirmed := model.ExaminationStatusConfirmed

	_, updateErr := svc.Update(ctx, clinicB, exam.ID, UpdateExaminationInput{
		ResultSummary: &summary,
		ActorID:       ptrUint64(actorID),
	})
	_, confirmErr := svc.Update(ctx, clinicB, exam.ID, UpdateExaminationInput{
		Status:  &confirmed,
		ActorID: ptrUint64(actorID),
	})
	deleteErr := svc.Delete(ctx, clinicB, exam.ID, ptrUint64(actorID))

	assert.True(t, apperrors.IsNotFound(updateErr))
	assert.True(t, apperrors.IsNotFound(confirmErr))
	assert.True(t, apperrors.IsNotFound(deleteErr))
	assert.Zero(t, auditCalls)
	persisted, err := repo.FindByID(ctx, clinicA, exam.ID)
	require.NoError(t, err)
	assert.Equal(t, "clinic A", persisted.ResultSummary)
	assert.Equal(t, model.ExaminationStatusPending, persisted.Status)
}

func newFailingAuditExaminationService(
	db *gorm.DB,
	repo ExaminationRepository,
	errAudit error,
) ExaminationService {
	return NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
			return errAudit
		}},
		testTransactor{db: db},
	)
}
