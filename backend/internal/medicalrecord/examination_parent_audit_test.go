package medicalrecord

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

func TestExaminationAuditValueIncludesLabImportProvenance(t *testing.T) {
	jobID := uuid.MustParse("5dd60dc8-03d3-4db4-a430-a9d60b76eb2b")

	value := examinationAuditValue(&model.Examination{JobID: &jobID})

	assert.Equal(t, jobID.String(), value["job_id"])
	assert.Nil(t, examinationAuditValue(&model.Examination{})["job_id"])
}

func TestExaminationService_ConfirmWithItemsPersistsItemsBeforeStatusTransition(t *testing.T) {
	const (
		clinicID = uint64(1)
		examID   = uint64(10)
		actorID  = uint64(42)
	)
	status := model.ExaminationStatusPending
	order := make([]string, 0, 4)
	var auditEntry *AuditEntry
	repo := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, gotClinicID, gotExamID uint64) (*model.Examination, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examID, gotExamID)
			return &model.Examination{
				ID: examID, ClinicID: clinicID, ExamTypeID: 7, Status: status,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, gotClinicID, gotExamID uint64, fields map[string]any) (*model.Examination, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examID, gotExamID)
			assert.Equal(t, model.ExaminationStatusConfirmed, fields["status"])
			order = append(order, "status")
			status = model.ExaminationStatusConfirmed
			return &model.Examination{
				ID: examID, ClinicID: clinicID, ExamTypeID: 7, Status: status,
			}, nil
		},
		replaceItemsByExamIDFn: func(_ context.Context, gotClinicID, gotExamID uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examID, gotExamID)
			assert.NotEqual(t, model.ExaminationStatusConfirmed, status, "items must be persisted before the final confirmed transition")
			order = append(order, "items")
			return items, 0, nil
		},
		appendOfficialRevisionFn: func(_ context.Context, _, _, _ uint64, _ string) (uint64, error) {
			order = append(order, "revision")
			return initialExaminationRevisionVersion, nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, entry *AuditEntry) error {
		order = append(order, "audit")
		auditEntry = entry
		return nil
	}}
	svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), audit, &mockCheckupTransactor{})
	items := []UpsertExamItemInput{{Name: "WBC", InspectionValue: "5.0"}}

	got, err := svc.Update(context.Background(), clinicID, examID, UpdateExaminationInput{
		Status:  &[]model.ExaminationStatus{model.ExaminationStatusConfirmed}[0],
		Items:   &items,
		ActorID: &[]uint64{actorID}[0],
	})

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, model.ExaminationStatusConfirmed, got.Status)
	assert.Equal(t, []string{"items", "revision", "audit", "status"}, order)
	require.NotNil(t, auditEntry)
	assert.Equal(t, model.AuditActionExaminationConfirm, auditEntry.Action)
	assert.Equal(t, model.AuditResourceExamination, auditEntry.Resource)
	assert.Equal(t, actorID, *auditEntry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, auditEntry.ActorType)
	assert.Equal(t, model.ExaminationStatusPending, auditEntry.OldValue.(map[string]any)["status"])
	assert.Equal(t, model.ExaminationStatusConfirmed, auditEntry.NewValue.(map[string]any)["status"])
	assert.Equal(t, "confirm", auditEntry.Metadata.(map[string]any)["operation_type"])
	assert.Equal(t, examinationAuditReasonAuthenticatedRequest, auditEntry.Metadata.(map[string]any)["reason"])
}

func TestExaminationService_CreateConfirmedWithItemsPersistsItemsBeforeStatusTransition(t *testing.T) {
	const (
		clinicID = uint64(1)
		examID   = uint64(10)
		actorID  = uint64(42)
	)
	status := model.ExaminationStatusPending
	order := make([]string, 0, 4)
	repo := &mockExaminationRepository{
		findByIDFn: func(_ context.Context, gotClinicID, gotExamID uint64) (*model.Examination, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examID, gotExamID)
			return &model.Examination{ID: examID, ClinicID: clinicID, ExamTypeID: 7, Status: status}, nil
		},
		createFn: func(_ context.Context, exam *model.Examination) error {
			assert.Equal(t, model.ExaminationStatusPending, exam.Status)
			exam.ID = examID
			status = exam.Status
			return nil
		},
		updateFieldsFn: func(_ context.Context, gotClinicID, gotExamID uint64, fields map[string]any) (*model.Examination, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examID, gotExamID)
			assert.Equal(t, model.ExaminationStatusConfirmed, fields["status"])
			order = append(order, "status")
			status = model.ExaminationStatusConfirmed
			return &model.Examination{ID: examID, ClinicID: clinicID, ExamTypeID: 7, Status: status}, nil
		},
		replaceItemsByExamIDFn: func(_ context.Context, _, _ uint64, items []model.ExamResult) ([]model.ExamResult, int64, error) {
			assert.NotEqual(t, model.ExaminationStatusConfirmed, status)
			order = append(order, "items")
			return items, 0, nil
		},
		appendOfficialRevisionFn: func(_ context.Context, _, _, _ uint64, _ string) (uint64, error) {
			order = append(order, "revision")
			return initialExaminationRevisionVersion, nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, entry *AuditEntry) error {
		order = append(order, "audit")
		assert.Equal(t, model.AuditActionExaminationCreate, entry.Action)
		assert.Equal(t, model.ExaminationStatusConfirmed, entry.NewValue.(map[string]any)["status"])
		return nil
	}}
	svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), audit, &mockCheckupTransactor{})
	items := []UpsertExamItemInput{{Name: "WBC", InspectionValue: "5.0"}}

	got, err := svc.Create(context.Background(), clinicID, &CreateExaminationInput{
		ExamTypeID: 7,
		Status:     model.ExaminationStatusConfirmed,
		Items:      &items,
		ActorID:    &[]uint64{actorID}[0],
	})

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, model.ExaminationStatusConfirmed, got.Status)
	assert.Equal(t, []string{"items", "revision", "audit", "status"}, order)
}

func TestExaminationService_ConfirmedMutationsReturnConflict(t *testing.T) {
	const actorID = uint64(42)

	t.Run("update", func(t *testing.T) {
		updated := false
		audited := false
		repo := &mockExaminationRepository{
			lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.Examination, error) {
				return &model.Examination{ID: id, ClinicID: clinicID, Status: model.ExaminationStatusConfirmed}, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
				updated = true
				return nil, nil
			},
		}
		audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
			audited = true
			return nil
		}}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), audit, &mockCheckupTransactor{})
		summary := "must remain immutable"

		got, err := svc.Update(context.Background(), 1, 10, UpdateExaminationInput{
			ResultSummary: &summary,
			ActorID:       &[]uint64{actorID}[0],
		})

		assert.Nil(t, got)
		assert.True(t, apperrors.IsConflict(err))
		assert.False(t, updated)
		assert.False(t, audited)
	})

	t.Run("delete", func(t *testing.T) {
		counted := false
		deleted := false
		audited := false
		repo := &mockExaminationRepository{
			lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.Examination, error) {
				return &model.Examination{ID: id, ClinicID: clinicID, Status: model.ExaminationStatusConfirmed}, nil
			},
			countItemsByExamIDFn: func(_ context.Context, _, _ uint64) (int64, error) {
				counted = true
				return 0, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error {
				deleted = true
				return nil
			},
		}
		audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
			audited = true
			return nil
		}}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), audit, &mockCheckupTransactor{})

		err := svc.Delete(context.Background(), 1, 10, &[]uint64{actorID}[0])

		assert.True(t, apperrors.IsConflict(err))
		assert.False(t, counted)
		assert.False(t, deleted)
		assert.False(t, audited)
	})
}

func TestExaminationService_ParentMutationsWriteActorAndBeforeAfterAudit(t *testing.T) {
	const (
		clinicID = uint64(1)
		examID   = uint64(10)
		actorID  = uint64(42)
	)
	date := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	actor := actorID

	t.Run("create", func(t *testing.T) {
		var entry *AuditEntry
		repo := &mockExaminationRepository{
			findByIDFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return &model.Examination{
					ID: examID, ClinicID: clinicID, ExamTypeID: 7, Date: date,
					Status: model.ExaminationStatusPending,
				}, nil
			},
			createFn: func(_ context.Context, exam *model.Examination) error {
				exam.ID = examID
				return nil
			},
		}
		audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, got *AuditEntry) error {
			entry = got
			return nil
		}}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), audit, &mockCheckupTransactor{})

		got, err := svc.Create(context.Background(), clinicID, &CreateExaminationInput{
			ExamTypeID: 7,
			Date:       date,
			ActorID:    &actor,
		})

		require.NoError(t, err)
		require.NotNil(t, got)
		assertParentExaminationAudit(t, entry, model.AuditActionExaminationCreate, "create", clinicID, examID, actorID)
		assert.Nil(t, entry.OldValue)
		assert.Equal(t, model.ExaminationStatusPending, entry.NewValue.(map[string]any)["status"])
	})

	t.Run("update", func(t *testing.T) {
		var entry *AuditEntry
		before := &model.Examination{
			ID: examID, ClinicID: clinicID, ExamTypeID: 7, Date: date,
			ResultSummary: "before", Status: model.ExaminationStatusPending,
		}
		repo := &mockExaminationRepository{
			lockByIDForUpdateFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return before, nil
			},
			updateFieldsFn: func(_ context.Context, _, _ uint64, fields map[string]any) (*model.Examination, error) {
				after := *before
				after.ResultSummary = fields["result_summary"].(string)
				return &after, nil
			},
		}
		audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, got *AuditEntry) error {
			entry = got
			return nil
		}}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), audit, &mockCheckupTransactor{})
		summary := "after"

		got, err := svc.Update(context.Background(), clinicID, examID, UpdateExaminationInput{
			ResultSummary: &summary,
			ActorID:       &actor,
		})

		require.NoError(t, err)
		require.NotNil(t, got)
		assertParentExaminationAudit(t, entry, model.AuditActionExaminationUpdate, "update", clinicID, examID, actorID)
		assert.Equal(t, "before", entry.OldValue.(map[string]any)["result_summary"])
		assert.Equal(t, "after", entry.NewValue.(map[string]any)["result_summary"])
	})

	t.Run("delete", func(t *testing.T) {
		var entry *AuditEntry
		before := &model.Examination{
			ID: examID, ClinicID: clinicID, ExamTypeID: 7, Date: date,
			ResultSummary: "before delete", Status: model.ExaminationStatusPending,
		}
		repo := &mockExaminationRepository{
			lockByIDForUpdateFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
				return before, nil
			},
			deleteFn: func(_ context.Context, _, _ uint64) error { return nil },
		}
		audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, got *AuditEntry) error {
			entry = got
			return nil
		}}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), audit, &mockCheckupTransactor{})

		err := svc.Delete(context.Background(), clinicID, examID, &actor)

		require.NoError(t, err)
		assertParentExaminationAudit(t, entry, model.AuditActionExaminationDelete, "delete", clinicID, examID, actorID)
		assert.Equal(t, "before delete", entry.OldValue.(map[string]any)["result_summary"])
		assert.Nil(t, entry.NewValue)
	})
}

func TestExaminationService_ParentMutationRequiresActorAndAuditDependencyBeforeWrite(t *testing.T) {
	actorID := uint64(42)
	dependencies := []struct {
		name    string
		actorID *uint64
		audit   AuditTxLogger
	}{
		{name: "missing actor", audit: &mockAuditTxLogger{}},
		{name: "zero actor", actorID: &[]uint64{0}[0], audit: &mockAuditTxLogger{}},
		{name: "missing audit dependency", actorID: &actorID},
	}
	operations := []struct {
		name string
		run  func(ExaminationService, *uint64) error
	}{
		{name: "create", run: func(svc ExaminationService, actorID *uint64) error {
			_, err := svc.Create(context.Background(), 1, &CreateExaminationInput{ExamTypeID: 7, ActorID: actorID})
			return err
		}},
		{name: "update", run: func(svc ExaminationService, actorID *uint64) error {
			summary := "after"
			_, err := svc.Update(context.Background(), 1, 10, UpdateExaminationInput{ResultSummary: &summary, ActorID: actorID})
			return err
		}},
		{name: "confirm", run: func(svc ExaminationService, actorID *uint64) error {
			confirmed := model.ExaminationStatusConfirmed
			_, err := svc.Update(context.Background(), 1, 10, UpdateExaminationInput{Status: &confirmed, ActorID: actorID})
			return err
		}},
		{name: "delete", run: func(svc ExaminationService, actorID *uint64) error {
			return svc.Delete(context.Background(), 1, 10, actorID)
		}},
	}

	for _, operation := range operations {
		for _, dependency := range dependencies {
			t.Run(operation.name+"/"+dependency.name, func(t *testing.T) {
				mutationCalls := 0
				repo := &mockExaminationRepository{
					createFn: func(_ context.Context, _ *model.Examination) error {
						mutationCalls++
						return nil
					},
					lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.Examination, error) {
						return &model.Examination{ID: id, ClinicID: clinicID, ExamTypeID: 7, Status: model.ExaminationStatusPending}, nil
					},
					updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
						mutationCalls++
						return &model.Examination{ID: 10}, nil
					},
					deleteFn: func(_ context.Context, _, _ uint64) error {
						mutationCalls++
						return nil
					},
				}
				svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), dependency.audit, &mockCheckupTransactor{})

				err := operation.run(svc, dependency.actorID)

				assert.Error(t, err)
				assert.Zero(t, mutationCalls)
			})
		}
	}
}

func TestExaminationService_ParentMutationAuditFailureRollsBack(t *testing.T) {
	errAudit := errors.New("examination audit failed")
	actorID := uint64(42)

	t.Run("create", func(t *testing.T) {
		state := examinationMutationState{}
		tx := &examinationStateRollbackTransactor{state: &state}
		repo := &mockExaminationRepository{createFn: func(_ context.Context, exam *model.Examination) error {
			state.exists = true
			exam.ID = 10
			return nil
		}}
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), &mockAuditTxLogger{
			logEntryTxFn: func(_ context.Context, _ *AuditEntry) error { return errAudit },
		}, tx)

		got, err := svc.Create(context.Background(), 1, &CreateExaminationInput{ExamTypeID: 7, ActorID: &actorID})

		assert.ErrorIs(t, err, errAudit)
		assert.Nil(t, got)
		assert.False(t, state.exists)
		assert.True(t, tx.rolledBack)
	})

	t.Run("update", func(t *testing.T) {
		state := examinationMutationState{exists: true, status: model.ExaminationStatusPending, resultSummary: "before"}
		tx := &examinationStateRollbackTransactor{state: &state}
		repo := statefulExaminationRepository(&state)
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), &mockAuditTxLogger{
			logEntryTxFn: func(_ context.Context, _ *AuditEntry) error { return errAudit },
		}, tx)
		summary := "after"

		got, err := svc.Update(context.Background(), 1, 10, UpdateExaminationInput{ResultSummary: &summary, ActorID: &actorID})

		assert.ErrorIs(t, err, errAudit)
		assert.Nil(t, got)
		assert.Equal(t, "before", state.resultSummary)
		assert.True(t, tx.rolledBack)
	})

	t.Run("confirm", func(t *testing.T) {
		state := examinationMutationState{exists: true, status: model.ExaminationStatusPending}
		tx := &examinationStateRollbackTransactor{state: &state}
		repo := statefulExaminationRepository(&state)
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), &mockAuditTxLogger{
			logEntryTxFn: func(_ context.Context, _ *AuditEntry) error { return errAudit },
		}, tx)
		confirmed := model.ExaminationStatusConfirmed

		got, err := svc.Update(context.Background(), 1, 10, UpdateExaminationInput{Status: &confirmed, ActorID: &actorID})

		assert.ErrorIs(t, err, errAudit)
		assert.Nil(t, got)
		assert.Equal(t, model.ExaminationStatusPending, state.status)
		assert.True(t, tx.rolledBack)
	})

	t.Run("delete", func(t *testing.T) {
		state := examinationMutationState{exists: true, status: model.ExaminationStatusPending}
		tx := &examinationStateRollbackTransactor{state: &state}
		repo := statefulExaminationRepository(&state)
		svc := NewExaminationService(repo, &mockMedicalRecordRepository{}, okExamTypeRepo(), &mockAuditTxLogger{
			logEntryTxFn: func(_ context.Context, _ *AuditEntry) error { return errAudit },
		}, tx)

		err := svc.Delete(context.Background(), 1, 10, &actorID)

		assert.ErrorIs(t, err, errAudit)
		assert.True(t, state.exists)
		assert.True(t, tx.rolledBack)
	})
}

func assertParentExaminationAudit(
	t *testing.T,
	entry *AuditEntry,
	action, operation string,
	clinicID, examID, actorID uint64,
) {
	t.Helper()
	require.NotNil(t, entry)
	require.NotNil(t, entry.ClinicID)
	require.NotNil(t, entry.ActorID)
	require.NotNil(t, entry.ResourceID)
	assert.Equal(t, clinicID, *entry.ClinicID)
	assert.Equal(t, actorID, *entry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, entry.ActorType)
	assert.Equal(t, action, entry.Action)
	assert.Equal(t, model.AuditResourceExamination, entry.Resource)
	assert.Equal(t, examID, *entry.ResourceID)
	metadata, ok := entry.Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, operation, metadata["operation_type"])
	assert.Equal(t, examinationAuditReasonAuthenticatedRequest, metadata["reason"])
}

type examinationMutationState struct {
	exists        bool
	status        model.ExaminationStatus
	resultSummary string
}

type examinationStateRollbackTransactor struct {
	state      *examinationMutationState
	rolledBack bool
}

func (t *examinationStateRollbackTransactor) WithTx(ctx context.Context, fn func(context.Context) error) error {
	snapshot := *t.state
	err := fn(ctx)
	if err != nil {
		*t.state = snapshot
		t.rolledBack = true
	}
	return err
}

func statefulExaminationRepository(state *examinationMutationState) *mockExaminationRepository {
	return &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, clinicID, id uint64) (*model.Examination, error) {
			if !state.exists {
				return nil, apperrors.WrapNotFound("examination", "not found")
			}
			return &model.Examination{
				ID: id, ClinicID: clinicID, ExamTypeID: 7,
				Status: state.status, ResultSummary: state.resultSummary,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, clinicID, id uint64, fields map[string]any) (*model.Examination, error) {
			if status, ok := fields["status"].(model.ExaminationStatus); ok {
				state.status = status
			}
			if summary, ok := fields["result_summary"].(string); ok {
				state.resultSummary = summary
			}
			return &model.Examination{
				ID: id, ClinicID: clinicID, ExamTypeID: 7,
				Status: state.status, ResultSummary: state.resultSummary,
			}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			state.exists = false
			return nil
		},
	}
}
