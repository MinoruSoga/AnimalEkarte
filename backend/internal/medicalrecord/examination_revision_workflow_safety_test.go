package medicalrecord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/testdb"
)

type examinationRevisionWorkflowFailureRepository struct {
	ExaminationRepository
	ExaminationRevisionWorkflowRepository
	failureStage string
	failure      error
}

func (r *examinationRevisionWorkflowFailureRepository) AppendWorkingRevisionFromOfficial(
	ctx context.Context,
	clinicID, examinationID, officialVersion, actorID uint64,
	changeReason string,
) (uint64, error) {
	version, err := r.ExaminationRevisionWorkflowRepository.AppendWorkingRevisionFromOfficial(
		ctx, clinicID, examinationID, officialVersion, actorID, changeReason,
	)
	if err == nil && r.failureStage == "working_from_official" {
		return 0, r.failure
	}
	return version, err
}

func (r *examinationRevisionWorkflowFailureRepository) AppendWorkingRevisionFromCurrent(
	ctx context.Context,
	clinicID, examinationID, currentVersion, actorID uint64,
	changeReason string,
) (uint64, error) {
	version, err := r.ExaminationRevisionWorkflowRepository.AppendWorkingRevisionFromCurrent(
		ctx, clinicID, examinationID, currentVersion, actorID, changeReason,
	)
	if err == nil && r.failureStage == "working_from_current" {
		return 0, r.failure
	}
	return version, err
}

func (r *examinationRevisionWorkflowFailureRepository) AppendOfficialRevisionFromWorking(
	ctx context.Context,
	clinicID, examinationID, workingVersion, actorID uint64,
	changeReason string,
) (uint64, error) {
	version, err := r.ExaminationRevisionWorkflowRepository.AppendOfficialRevisionFromWorking(
		ctx, clinicID, examinationID, workingVersion, actorID, changeReason,
	)
	if err == nil && r.failureStage == "official_from_working" {
		return 0, r.failure
	}
	return version, err
}

func (r *examinationRevisionWorkflowFailureRepository) AdvanceRevisionCAS(
	ctx context.Context,
	clinicID, examinationID uint64,
	expectedStatus model.ExaminationStatus,
	expectedVersion uint64,
	nextStatus model.ExaminationStatus,
	nextVersion uint64,
) (*model.Examination, error) {
	if r.failureStage == "cas" {
		return nil, r.failure
	}
	return r.ExaminationRevisionWorkflowRepository.AdvanceRevisionCAS(
		ctx,
		clinicID,
		examinationID,
		expectedStatus,
		expectedVersion,
		nextStatus,
		nextVersion,
	)
}

func TestExaminationRevision_UnconfirmAndWorkingWorkflowFailuresRollBack(t *testing.T) {
	tests := []struct {
		name         string
		prepareWork  bool
		failureStage string
		invoke       func(ExaminationService, context.Context, uint64, uint64, uint64) error
	}{
		{
			name:         "unconfirm revision append failure",
			failureStage: "working_from_official",
			invoke: func(service ExaminationService, ctx context.Context, clinicID, examinationID, actorID uint64) error {
				_, err := service.Unconfirm(ctx, clinicID, examinationID, UnconfirmExaminationInput{
					Reason: "correction required", ActorID: &actorID,
				})
				return err
			},
		},
		{
			name:         "unconfirm pointer CAS failure",
			failureStage: "cas",
			invoke: func(service ExaminationService, ctx context.Context, clinicID, examinationID, actorID uint64) error {
				_, err := service.Unconfirm(ctx, clinicID, examinationID, UnconfirmExaminationInput{
					Reason: "correction required", ActorID: &actorID,
				})
				return err
			},
		},
		{
			name:         "working parent update revision append failure",
			prepareWork:  true,
			failureStage: "working_from_current",
			invoke: func(service ExaminationService, ctx context.Context, clinicID, examinationID, actorID uint64) error {
				summary := "replacement summary"
				_, err := service.Update(ctx, clinicID, examinationID, UpdateExaminationInput{
					ResultSummary: &summary, ActorID: &actorID,
				})
				return err
			},
		},
		{
			name:         "working item replacement revision append failure",
			prepareWork:  true,
			failureStage: "working_from_current",
			invoke: func(service ExaminationService, ctx context.Context, clinicID, examinationID, actorID uint64) error {
				_, err := service.ReplaceItems(ctx, clinicID, examinationID, &actorID, []UpsertExamItemInput{{
					Name: "replacement item", InspectionValue: "7.0", SortOrder: 1,
				}})
				return err
			},
		},
		{
			name:         "reconfirm revision append failure",
			prepareWork:  true,
			failureStage: "official_from_working",
			invoke: func(service ExaminationService, ctx context.Context, clinicID, examinationID, actorID uint64) error {
				status := model.ExaminationStatusConfirmed
				_, err := service.Update(ctx, clinicID, examinationID, UpdateExaminationInput{
					Status: &status, ActorID: &actorID,
				})
				return err
			},
		},
		{
			name:         "reconfirm pointer CAS failure",
			prepareWork:  true,
			failureStage: "cas",
			invoke: func(service ExaminationService, ctx context.Context, clinicID, examinationID, actorID uint64) error {
				status := model.ExaminationStatusConfirmed
				_, err := service.Update(ctx, clinicID, examinationID, UpdateExaminationInput{
					Status: &status, ActorID: &actorID,
				})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupExaminationTestDB(t)
			require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AuditLog{}))
			ctx := context.Background()
			const clinicID = uint64(1)
			actorID := makeExaminationActor(t, db, clinicID, "workflow rollback actor")
			examType := makeExamTypeMaster(t, db, clinicID, "workflow rollback exam type")
			base := NewExaminationRepository(db)
			baseWorkflow, ok := base.(ExaminationRevisionWorkflowRepository)
			require.True(t, ok)
			creator := NewExaminationService(
				base,
				&mockMedicalRecordRepository{},
				NewExamTypeRepository(db),
				&mockAuditTxLogger{},
				persistence.NewTransactor(db),
			)
			confirmed, err := creator.Create(ctx, clinicID, &CreateExaminationInput{
				ExamTypeID: examType.ID,
				Date:       time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
				Status:     model.ExaminationStatusConfirmed,
				Items: &[]UpsertExamItemInput{{
					Name: "manual field", InspectionValue: "5.0", SortOrder: 1,
				}},
				ActorID: &actorID,
			})
			require.NoError(t, err)
			if tt.prepareWork {
				_, err = creator.Unconfirm(ctx, clinicID, confirmed.ID, UnconfirmExaminationInput{
					Reason: "prepare working revision", ActorID: &actorID,
				})
				require.NoError(t, err)
			}

			before, err := base.FindByID(ctx, clinicID, confirmed.ID)
			require.NoError(t, err)
			beforeItems, err := base.FindAllItemsByExamID(ctx, clinicID, confirmed.ID)
			require.NoError(t, err)
			var beforeRevisionCount int64
			require.NoError(t, db.Model(&model.ExaminationRevision{}).
				Where("clinic_id = ? AND examination_id = ?", clinicID, confirmed.ID).
				Count(&beforeRevisionCount).Error)

			failure := errors.New("injected " + tt.failureStage + " failure")
			repo := &examinationRevisionWorkflowFailureRepository{
				ExaminationRepository:                 base,
				ExaminationRevisionWorkflowRepository: baseWorkflow,
				failureStage:                          tt.failureStage,
				failure:                               failure,
			}
			auditMarker := fmt.Sprintf("revision-workflow-rollback-%s-%d", tt.failureStage, confirmed.ID)
			audit := AuditTxLoggerFunc(func(auditCtx context.Context, entry *AuditEntry) error {
				marshal := func(value any) (json.RawMessage, error) {
					if value == nil {
						return nil, nil
					}
					encoded, marshalErr := json.Marshal(value)
					return json.RawMessage(encoded), marshalErr
				}
				oldValue, marshalErr := marshal(entry.OldValue)
				if marshalErr != nil {
					return marshalErr
				}
				newValue, marshalErr := marshal(entry.NewValue)
				if marshalErr != nil {
					return marshalErr
				}
				metadata, marshalErr := marshal(entry.Metadata)
				if marshalErr != nil {
					return marshalErr
				}
				return persistence.DBOrTx(auditCtx, db).Create(&model.AuditLog{
					ClinicID: entry.ClinicID, ActorID: entry.ActorID, ActorType: entry.ActorType,
					Action: entry.Action, Resource: entry.Resource, ResourceID: entry.ResourceID,
					OldValue: oldValue, NewValue: newValue, Metadata: metadata, UserAgent: auditMarker,
				}).Error
			})
			service := NewExaminationService(
				repo,
				&mockMedicalRecordRepository{},
				NewExamTypeRepository(db),
				audit,
				persistence.NewTransactor(db),
			)

			err = tt.invoke(service, ctx, clinicID, confirmed.ID, actorID)

			assert.ErrorIs(t, err, failure)
			after, findErr := base.FindByID(ctx, clinicID, confirmed.ID)
			require.NoError(t, findErr)
			assert.Equal(t, before.Status, after.Status)
			assert.Equal(t, before.CurrentRevisionVersion, after.CurrentRevisionVersion)
			assert.Equal(t, before.ResultSummary, after.ResultSummary)
			afterItems, findItemsErr := base.FindAllItemsByExamID(ctx, clinicID, confirmed.ID)
			require.NoError(t, findItemsErr)
			require.Len(t, afterItems, len(beforeItems))
			for i := range beforeItems {
				assert.Equal(t, beforeItems[i].ID, afterItems[i].ID)
				assert.Equal(t, beforeItems[i].Name, afterItems[i].Name)
				assert.Equal(t, beforeItems[i].InspectionValue, afterItems[i].InspectionValue)
			}
			var afterRevisionCount, auditCount int64
			require.NoError(t, db.Model(&model.ExaminationRevision{}).
				Where("clinic_id = ? AND examination_id = ?", clinicID, confirmed.ID).
				Count(&afterRevisionCount).Error)
			require.NoError(t, db.Model(&model.AuditLog{}).
				Where("user_agent = ?", auditMarker).
				Count(&auditCount).Error)
			assert.Equal(t, beforeRevisionCount, afterRevisionCount)
			assert.Zero(t, auditCount)
		})
	}
}

func TestExaminationRevision_PetIdentityCannotChangeThroughMedicalRecordAfterHistory(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	actorID := makeExaminationActor(t, db, clinicID, "patient identity actor")
	examType := makeExamTypeMaster(t, db, clinicID, "patient identity exam type")
	owner := testdb.MakeTestOwner(t, db, clinicID, "patient identity owner")
	pet := makeSpeciesAndPet(t, db, clinicID, owner.ID, "patient identity pet")
	initialRecord := &model.MedicalRecord{
		ClinicID: clinicID, RecordNo: "MR-IDENTITY-EMPTY", Date: time.Now(), Status: model.MedicalRecordStatusDraft,
	}
	replacementRecord := &model.MedicalRecord{
		ClinicID: clinicID, RecordNo: "MR-IDENTITY-PET", Date: time.Now(),
		OwnerID: &owner.ID, PetID: &pet.ID, Status: model.MedicalRecordStatusDraft,
	}
	require.NoError(t, db.Create(initialRecord).Error)
	require.NoError(t, db.Create(replacementRecord).Error)
	repo := NewExaminationRepository(db)
	service := NewExaminationService(
		repo,
		NewMedicalRecordRepository(db),
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		persistence.NewTransactor(db),
		reservation.NewReservationRepository(db),
	)
	confirmed, err := service.Create(ctx, clinicID, &CreateExaminationInput{
		MedicalRecordID: &initialRecord.ID,
		ExamTypeID:      examType.ID,
		Date:            time.Now(),
		Status:          model.ExaminationStatusConfirmed,
		ActorID:         &actorID,
	})
	require.NoError(t, err)
	working, err := service.Unconfirm(ctx, clinicID, confirmed.ID, UnconfirmExaminationInput{
		Reason: "patient identity guard", ActorID: &actorID,
	})
	require.NoError(t, err)
	require.NotNil(t, working.CurrentRevisionVersion)

	got, err := service.Update(ctx, clinicID, confirmed.ID, UpdateExaminationInput{
		MedicalRecordID: &replacementRecord.ID,
		ActorID:         &actorID,
	})

	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, got)
	persisted, findErr := repo.FindByID(ctx, clinicID, confirmed.ID)
	require.NoError(t, findErr)
	require.NotNil(t, persisted.MedicalRecordID)
	assert.Equal(t, initialRecord.ID, *persisted.MedicalRecordID)
	assert.Equal(t, uint64(2), *persisted.CurrentRevisionVersion)
}

func TestExaminationRevision_UnconfirmAndReconfirmRejectFinalizedMedicalRecord(t *testing.T) {
	for _, operation := range []string{"unconfirm", "reconfirm"} {
		t.Run(operation, func(t *testing.T) {
			db := setupExaminationTestDB(t)
			ctx := context.Background()
			const clinicID = uint64(1)
			actorID := makeExaminationActor(t, db, clinicID, "finalized record actor")
			examType := makeExamTypeMaster(t, db, clinicID, "finalized record exam type")
			record := &model.MedicalRecord{
				ClinicID: clinicID, RecordNo: "MR-FINALIZED-REVISION", Date: time.Now(), Status: model.MedicalRecordStatusDraft,
			}
			require.NoError(t, db.Create(record).Error)
			repo := NewExaminationRepository(db)
			service := NewExaminationService(
				repo,
				NewMedicalRecordRepository(db),
				NewExamTypeRepository(db),
				&mockAuditTxLogger{},
				persistence.NewTransactor(db),
				reservation.NewReservationRepository(db),
			)
			confirmed, err := service.Create(ctx, clinicID, &CreateExaminationInput{
				MedicalRecordID: &record.ID, ExamTypeID: examType.ID, Date: time.Now(),
				Status: model.ExaminationStatusConfirmed, ActorID: &actorID,
			})
			require.NoError(t, err)
			if operation == "reconfirm" {
				_, err = service.Unconfirm(ctx, clinicID, confirmed.ID, UnconfirmExaminationInput{
					Reason: "prepare reconfirm", ActorID: &actorID,
				})
				require.NoError(t, err)
			}
			require.NoError(t, db.Model(&model.MedicalRecord{}).
				Where("clinic_id = ? AND id = ?", clinicID, record.ID).
				Update("status", model.MedicalRecordStatusFinalized).Error)

			if operation == "unconfirm" {
				_, err = service.Unconfirm(ctx, clinicID, confirmed.ID, UnconfirmExaminationInput{
					Reason: "must reject finalized record", ActorID: &actorID,
				})
			} else {
				status := model.ExaminationStatusConfirmed
				_, err = service.Update(ctx, clinicID, confirmed.ID, UpdateExaminationInput{
					Status: &status, ActorID: &actorID,
				})
			}

			assert.True(t, apperrors.IsConflict(err))
			persisted, findErr := repo.FindByID(ctx, clinicID, confirmed.ID)
			require.NoError(t, findErr)
			if operation == "unconfirm" {
				assert.Equal(t, model.ExaminationStatusConfirmed, persisted.Status)
				assert.Equal(t, uint64(1), *persisted.CurrentRevisionVersion)
			} else {
				assert.Equal(t, model.ExaminationStatusCompleted, persisted.Status)
				assert.Equal(t, uint64(2), *persisted.CurrentRevisionVersion)
			}
		})
	}
}

func TestExaminationRevision_UnconfirmThenDeletePreservesConfirmedHistory(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	actorID := makeExaminationActor(t, db, clinicID, "delete protection actor")
	examType := makeExamTypeMaster(t, db, clinicID, "delete protection exam type")
	repo := NewExaminationRepository(db)
	service := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		persistence.NewTransactor(db),
	)
	confirmed, err := service.Create(ctx, clinicID, &CreateExaminationInput{
		ExamTypeID: examType.ID, Date: time.Now(), Status: model.ExaminationStatusConfirmed, ActorID: &actorID,
	})
	require.NoError(t, err)
	_, err = service.Unconfirm(ctx, clinicID, confirmed.ID, UnconfirmExaminationInput{
		Reason: "delete protection", ActorID: &actorID,
	})
	require.NoError(t, err)

	err = service.Delete(ctx, clinicID, confirmed.ID, &actorID)

	assert.True(t, apperrors.IsConflict(err))
	persisted, findErr := repo.FindByID(ctx, clinicID, confirmed.ID)
	require.NoError(t, findErr)
	assert.False(t, persisted.DeletedAt.Valid)
}

func TestExaminationRevision_UnconfirmAndReconfirmRejectCrossClinicRevisionRelations(t *testing.T) {
	for _, operation := range []string{"unconfirm", "reconfirm"} {
		for _, pollutedRelation := range []string{"job", "item_field"} {
			t.Run(operation+"_"+pollutedRelation, func(t *testing.T) {
				db := setupExaminationTestDB(t)
				require.NoError(t, testdb.EnsureAutoMigrated(db, &model.LabImportJob{}))
				ctx := context.Background()
				const clinicA, clinicB = uint64(1), uint64(2)
				actorID := makeExaminationActor(t, db, clinicA, "relation validation actor")
				examTypeA := makeExamTypeMaster(t, db, clinicA, "relation validation exam type")
				examTypeB := makeExamTypeMaster(t, db, clinicB, "foreign relation exam type")
				ownerA := testdb.MakeTestOwner(t, db, clinicA, "relation validation owner")
				petA := makeSpeciesAndPet(t, db, clinicA, ownerA.ID, "relation validation pet")
				fieldA := &model.ExamTypeField{ClinicID: clinicA, ExamTypeID: examTypeA.ID, Name: "field A", SortOrder: 1}
				fieldB := &model.ExamTypeField{ClinicID: clinicB, ExamTypeID: examTypeB.ID, Name: "field B", SortOrder: 1}
				require.NoError(t, db.Create(fieldA).Error)
				require.NoError(t, db.Create(fieldB).Error)
				foreignJob := &model.LabImportJob{
					ID: uuid.New(), ClinicID: clinicB, SourceType: model.LabImportSourceTypeFixture,
					Status: model.LabImportJobStatusPersisted,
				}
				require.NoError(t, db.Create(foreignJob).Error)
				repo := NewExaminationRepository(db)
				service := NewExaminationService(
					repo,
					&mockMedicalRecordRepository{},
					NewExamTypeRepository(db),
					&mockAuditTxLogger{},
					persistence.NewTransactor(db),
				)
				confirmed, err := service.Create(ctx, clinicA, &CreateExaminationInput{
					PetID: &petA.ID, ExamTypeID: examTypeA.ID, Date: time.Now(), Status: model.ExaminationStatusConfirmed,
					Items: &[]UpsertExamItemInput{{
						ExamTypeFieldID: &fieldA.ID, Name: "field A", InspectionValue: "5.0", SortOrder: 1,
					}},
					ActorID: &actorID,
				})
				require.NoError(t, err)
				version := uint64(1)
				if operation == "reconfirm" {
					working, unconfirmErr := service.Unconfirm(ctx, clinicA, confirmed.ID, UnconfirmExaminationInput{
						Reason: "prepare relation validation", ActorID: &actorID,
					})
					require.NoError(t, unconfirmErr)
					version = *working.CurrentRevisionVersion
				}
				if pollutedRelation == "job" {
					require.NoError(t, db.Model(&model.ExaminationRevision{}).
						Where("clinic_id = ? AND examination_id = ? AND version = ?", clinicA, confirmed.ID, version).
						Update("job_id", foreignJob.ID).Error)
				} else {
					require.NoError(t, db.Model(&model.ExaminationRevisionItem{}).
						Where("clinic_id = ? AND examination_id = ? AND version = ?", clinicA, confirmed.ID, version).
						Update("exam_type_field_id", fieldB.ID).Error)
				}

				if operation == "unconfirm" {
					_, err = service.Unconfirm(ctx, clinicA, confirmed.ID, UnconfirmExaminationInput{
						Reason: "must reject polluted relations", ActorID: &actorID,
					})
				} else {
					status := model.ExaminationStatusConfirmed
					_, err = service.Update(ctx, clinicA, confirmed.ID, UpdateExaminationInput{
						Status: &status, ActorID: &actorID,
					})
				}

				assert.Error(t, err)
				persisted, findErr := repo.FindByID(ctx, clinicA, confirmed.ID)
				require.NoError(t, findErr)
				assert.Equal(t, version, *persisted.CurrentRevisionVersion)
				if operation == "unconfirm" {
					assert.Equal(t, model.ExaminationStatusConfirmed, persisted.Status)
				} else {
					assert.Equal(t, model.ExaminationStatusCompleted, persisted.Status)
				}
				var revisionCount int64
				require.NoError(t, db.Model(&model.ExaminationRevision{}).
					Where("clinic_id = ? AND examination_id = ?", clinicA, confirmed.ID).
					Count(&revisionCount).Error)
				assert.Equal(t, int64(version), revisionCount)
				items, itemsErr := repo.FindAllItemsByExamID(ctx, clinicA, confirmed.ID)
				require.NoError(t, itemsErr)
				require.Len(t, items, 1)
				assert.Equal(t, fieldA.ID, *items[0].ExamTypeItemID)
			})
		}
	}
}

func TestExaminationRevision_UnconfirmRejectsCrossClinicExamAndActorWithoutWrites(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	actorA := makeExaminationActor(t, db, clinicA, "clinic A actor")
	actorB := makeExaminationActor(t, db, clinicB, "clinic B actor")
	examType := makeExamTypeMaster(t, db, clinicA, "cross clinic guard exam type")
	repo := NewExaminationRepository(db)
	auditCalls := 0
	service := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{logEntryTxFn: func(context.Context, *AuditEntry) error {
			auditCalls++
			return nil
		}},
		persistence.NewTransactor(db),
	)
	confirmed, err := service.Create(ctx, clinicA, &CreateExaminationInput{
		ExamTypeID: examType.ID, Date: time.Now(), Status: model.ExaminationStatusConfirmed, ActorID: &actorA,
	})
	require.NoError(t, err)
	auditCalls = 0

	for _, attempt := range []struct {
		name     string
		clinicID uint64
		actorID  uint64
	}{
		{name: "foreign clinic scope", clinicID: clinicB, actorID: actorB},
		{name: "foreign clinic actor", clinicID: clinicA, actorID: actorB},
	} {
		t.Run(attempt.name, func(t *testing.T) {
			got, unconfirmErr := service.Unconfirm(ctx, attempt.clinicID, confirmed.ID, UnconfirmExaminationInput{
				Reason: "cross clinic attempt", ActorID: &attempt.actorID,
			})
			assert.Error(t, unconfirmErr)
			assert.Nil(t, got)
		})
	}

	persisted, findErr := repo.FindByID(ctx, clinicA, confirmed.ID)
	require.NoError(t, findErr)
	assert.Equal(t, model.ExaminationStatusConfirmed, persisted.Status)
	assert.Equal(t, uint64(1), *persisted.CurrentRevisionVersion)
	assert.Zero(t, auditCalls)
	assertExaminationRevisionRows(t, db, clinicA, confirmed.ID, 1, 0)
}

func TestExaminationRevision_ReconfirmRejectsCrossClinicExamAndActorWithoutWrites(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicA, clinicB = uint64(1), uint64(2)
	actorA := makeExaminationActor(t, db, clinicA, "reconfirm clinic A actor")
	actorB := makeExaminationActor(t, db, clinicB, "reconfirm clinic B actor")
	examType := makeExamTypeMaster(t, db, clinicA, "reconfirm cross clinic guard exam type")
	repo := NewExaminationRepository(db)
	auditCalls := 0
	service := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{logEntryTxFn: func(context.Context, *AuditEntry) error {
			auditCalls++
			return nil
		}},
		persistence.NewTransactor(db),
	)
	confirmed, err := service.Create(ctx, clinicA, &CreateExaminationInput{
		ExamTypeID: examType.ID, Date: time.Now(), Status: model.ExaminationStatusConfirmed, ActorID: &actorA,
	})
	require.NoError(t, err)
	working, err := service.Unconfirm(ctx, clinicA, confirmed.ID, UnconfirmExaminationInput{
		Reason: "prepare cross clinic reconfirm", ActorID: &actorA,
	})
	require.NoError(t, err)
	require.NotNil(t, working.CurrentRevisionVersion)
	auditCalls = 0
	confirmedStatus := model.ExaminationStatusConfirmed

	for _, attempt := range []struct {
		name     string
		clinicID uint64
		actorID  uint64
	}{
		{name: "foreign clinic scope", clinicID: clinicB, actorID: actorB},
		{name: "foreign clinic actor", clinicID: clinicA, actorID: actorB},
	} {
		t.Run(attempt.name, func(t *testing.T) {
			got, reconfirmErr := service.Update(ctx, attempt.clinicID, confirmed.ID, UpdateExaminationInput{
				Status: &confirmedStatus, ActorID: &attempt.actorID,
			})
			assert.Error(t, reconfirmErr)
			assert.Nil(t, got)
		})
	}

	persisted, findErr := repo.FindByID(ctx, clinicA, confirmed.ID)
	require.NoError(t, findErr)
	assert.Equal(t, model.ExaminationStatusCompleted, persisted.Status)
	assert.Equal(t, uint64(2), *persisted.CurrentRevisionVersion)
	assert.Zero(t, auditCalls)
	assertExaminationRevisionRows(t, db, clinicA, confirmed.ID, 2, 0)
}

func TestExaminationRevision_UnconfirmCASRejectsStaleStatusAndVersion(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	actorID := makeExaminationActor(t, db, clinicID, "stale CAS actor")
	examType := makeExamTypeMaster(t, db, clinicID, "stale CAS exam type")
	repo := NewExaminationRepository(db)
	workflow, ok := repo.(ExaminationRevisionWorkflowRepository)
	require.True(t, ok)
	service := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		persistence.NewTransactor(db),
	)
	confirmed, err := service.Create(ctx, clinicID, &CreateExaminationInput{
		ExamTypeID: examType.ID, Date: time.Now(), Status: model.ExaminationStatusConfirmed, ActorID: &actorID,
	})
	require.NoError(t, err)

	for _, attempt := range []struct {
		name            string
		expectedStatus  model.ExaminationStatus
		expectedVersion uint64
	}{
		{name: "stale status", expectedStatus: model.ExaminationStatusCompleted, expectedVersion: 1},
		{name: "stale version", expectedStatus: model.ExaminationStatusConfirmed, expectedVersion: 2},
	} {
		t.Run(attempt.name, func(t *testing.T) {
			err := persistence.NewTransactor(db).WithTx(ctx, func(txCtx context.Context) error {
				got, casErr := workflow.AdvanceRevisionCAS(
					txCtx,
					clinicID,
					confirmed.ID,
					attempt.expectedStatus,
					attempt.expectedVersion,
					model.ExaminationStatusCompleted,
					attempt.expectedVersion+1,
				)
				assert.Nil(t, got)
				return casErr
			})
			assert.True(t, apperrors.IsConflict(err))
		})
	}

	persisted, findErr := repo.FindByID(ctx, clinicID, confirmed.ID)
	require.NoError(t, findErr)
	assert.Equal(t, model.ExaminationStatusConfirmed, persisted.Status)
	assert.Equal(t, uint64(1), *persisted.CurrentRevisionVersion)
	assertExaminationRevisionRows(t, db, clinicID, confirmed.ID, 1, 0)
}
