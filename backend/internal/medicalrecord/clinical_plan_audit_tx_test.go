package medicalrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func clinicalPlanTestActorID() *uint64 {
	id := uint64(42)
	return &id
}

// BUG-010 residual: clinical plan Update は audit 依存 nil / tx 依存 nil を fail-closed で拒否する。
func TestClinicalPlanService_Update_NilAuditOrTransactorRejected(t *testing.T) {
	physicalExam := "所見"
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1, PhysicalExam: "旧", Version: 1}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			t.Fatal("update must not run without audit/tx deps")
			return nil
		},
	}

	t.Run("nil auditTx", func(t *testing.T) {
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, nil)
		plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{PhysicalExam: &physicalExam,
			ActorID: clinicalPlanTestActorID(),
		})
		assert.Error(t, err)
		assert.Nil(t, plan)
		assert.Contains(t, err.Error(), "audit dependency is required")
	})

	t.Run("nil transactor", func(t *testing.T) {
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), nil, &mockAuditTxLogger{})
		plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{PhysicalExam: &physicalExam,
			ActorID: clinicalPlanTestActorID(),
		})
		assert.Error(t, err)
		assert.Nil(t, plan)
	})
}

// BUG-010 residual: 成功時は同一 tx 経路で audit が1件書かれ、before/after が入る。
func TestClinicalPlanService_Update_WritesAuditOnSuccess(t *testing.T) {
	physicalExam := "新しい所見"
	findCalls := 0
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			findCalls++
			if findCalls == 1 {
				return &model.ClinicalPlan{
					ID: 1, MedicalRecordID: 1, PhysicalExam: "旧所見", Version: 2,
					DiagnosisDetails: "旧診断", TreatmentPolicy: "旧方針",
				}, nil
			}
			return &model.ClinicalPlan{
				ID: 1, MedicalRecordID: 1, PhysicalExam: physicalExam, Version: 3,
				DiagnosisDetails: "旧診断", TreatmentPolicy: "旧方針",
			}, nil
		},
		updateWithVersionFn: func(_ context.Context, _, _ uint64, fields map[string]any, _ *int) error {
			assert.Equal(t, physicalExam, fields["physical_exam"])
			return nil
		},
	}
	var gotEntry *AuditEntry
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, entry *AuditEntry) error {
		gotEntry = entry
		return nil
	}}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, audit)

	plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{PhysicalExam: &physicalExam,
		ActorID: clinicalPlanTestActorID(),
	})
	require.NoError(t, err)
	require.NotNil(t, plan)
	require.NotNil(t, gotEntry)
	assert.Equal(t, model.AuditActionClinicalPlanUpdate, gotEntry.Action)
	assert.Equal(t, model.AuditResourceClinicalPlan, gotEntry.Resource)
	assert.Equal(t, uint64(1), *gotEntry.ResourceID)
	require.NotNil(t, gotEntry.ActorID)
	assert.Equal(t, uint64(42), *gotEntry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, gotEntry.ActorType)
	assert.Equal(t, "旧所見", gotEntry.OldValue.(map[string]any)["physical_exam"])
	assert.Equal(t, physicalExam, gotEntry.NewValue.(map[string]any)["physical_exam"])
	assert.Equal(t, "update", gotEntry.Metadata.(map[string]any)["operation_type"])
}

// BUG-010 residual: audit 失敗時は service が error を返し、mock 上は更新後に audit が走る（tx 内順序）。
// DB 原子性は TestClinicalPlanService_Update_AuditFailureRollsBackDB で固定する。
func TestClinicalPlanService_Update_AuditFailureReturnsError(t *testing.T) {
	physicalExam := "所見"
	updateCalled := false
	auditCalled := false
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1, PhysicalExam: "旧", Version: 1}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			updateCalled = true
			return nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalled = true
		return errors.New("audit write failed")
	}}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, audit)

	plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{PhysicalExam: &physicalExam,
		ActorID: clinicalPlanTestActorID(),
	})
	assert.Error(t, err)
	assert.Nil(t, plan)
	assert.True(t, updateCalled, "update runs before audit in the same tx callback")
	assert.True(t, auditCalled)
}

// BUG-010 residual: 拒否経路（finalized）では audit も update も走らない。
func TestClinicalPlanService_Update_FinalizedDoesNotAudit(t *testing.T) {
	physicalExam := "後書き"
	auditCalled := false
	updateCalled := false
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1, PhysicalExam: "確定前"}, nil
		},
		updateFn: func(_ context.Context, _, _ uint64, _ map[string]any) error {
			updateCalled = true
			return nil
		},
	}
	medRec := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalled = true
		return nil
	}}
	svc := NewClinicalPlanService(repo, medRec, okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, audit)

	plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{PhysicalExam: &physicalExam,
		ActorID: clinicalPlanTestActorID(),
	})
	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, plan)
	assert.False(t, updateCalled)
	assert.False(t, auditCalled)
}

// BUG-010 residual: stale version 拒否時も audit ゼロ。
func TestClinicalPlanService_Update_StaleVersionDoesNotAudit(t *testing.T) {
	physicalExam := "競合更新"
	version := 1
	auditCalled := false
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1, PhysicalExam: "現行", Version: 2}, nil
		},
		updateWithVersionFn: func(_ context.Context, _, _ uint64, _ map[string]any, expectedVersion *int) error {
			require.NotNil(t, expectedVersion)
			assert.Equal(t, 1, *expectedVersion)
			return apperrors.WrapConflict("他のユーザーがこの所見・診断を変更しました。再読み込みしてください")
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalled = true
		return nil
	}}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, audit)

	plan, err := svc.Update(context.Background(), 1, 1, &UpdateClinicalPlanInput{
		PhysicalExam: &physicalExam,
		Version:      &version,
		ActorID:      clinicalPlanTestActorID(),
	})
	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.Nil(t, plan)
	assert.False(t, auditCalled)
}

// BUG-010 residual DB: audit 失敗で clinical_plans 更新が rollback される（同一 DBOrTx）。
func TestClinicalPlanService_Update_AuditFailureRollsBackDB(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	const clinicA = uint64(1)
	mr := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-AUDIT-RB")
	repo := NewClinicalPlanRepository(db)
	plan := &model.ClinicalPlan{
		MedicalRecordID:  mr.ID,
		PhysicalExam:     "rollback-should-keep",
		DiagnosisDetails: "診断詳細",
		TreatmentPolicy:  "治療方針",
		Version:          1,
	}
	require.NoError(t, repo.Create(context.Background(), plan))

	failingAudit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		return errors.New("simulated clinical plan audit failure")
	}}
	svc := NewClinicalPlanService(
		repo,
		okMedRecForPlan(),
		okDiagnosisTypeRepo(),
		okDiagnosisNameRepo(),
		persistence.NewTransactor(db),
		failingAudit,
	)

	newExam := "must-not-persist"
	version := 1
	updated, err := svc.Update(context.Background(), clinicA, mr.ID, &UpdateClinicalPlanInput{
		PhysicalExam: &newExam,
		Version:      &version,
		ActorID:      clinicalPlanTestActorID(),
	})
	require.Error(t, err)
	assert.Nil(t, updated)

	persisted, err := repo.FindByMedicalRecordID(context.Background(), clinicA, mr.ID)
	require.NoError(t, err)
	assert.Equal(t, "rollback-should-keep", persisted.PhysicalExam)
	assert.Equal(t, 1, persisted.Version)
}

// BUG-010 residual DB: 成功時は更新と audit が commit され version が進む。
func TestClinicalPlanService_Update_AuditSuccessCommitsDB(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	const clinicA = uint64(1)
	mr := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-AUDIT-OK")
	repo := NewClinicalPlanRepository(db)
	plan := &model.ClinicalPlan{
		MedicalRecordID: mr.ID,
		PhysicalExam:    "before",
		Version:         1,
	}
	require.NoError(t, repo.Create(context.Background(), plan))

	var auditEntries int
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, entry *AuditEntry) error {
		auditEntries++
		assert.Equal(t, model.AuditActionClinicalPlanUpdate, entry.Action)
		return nil
	}}
	svc := NewClinicalPlanService(
		repo,
		okMedRecForPlan(),
		okDiagnosisTypeRepo(),
		okDiagnosisNameRepo(),
		persistence.NewTransactor(db),
		audit,
	)

	newExam := "after"
	version := 1
	updated, err := svc.Update(context.Background(), clinicA, mr.ID, &UpdateClinicalPlanInput{
		PhysicalExam: &newExam,
		Version:      &version,
		ActorID:      clinicalPlanTestActorID(),
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "after", updated.PhysicalExam)
	assert.Equal(t, 2, updated.Version)
	assert.Equal(t, 1, auditEntries)

	persisted, err := repo.FindByMedicalRecordID(context.Background(), clinicA, mr.ID)
	require.NoError(t, err)
	assert.Equal(t, "after", persisted.PhysicalExam)
	assert.Equal(t, 2, persisted.Version)
}

// ---- Delete audit (BUG-010 residual) ----

// BUG-010 residual: Delete は audit/tx 依存 nil と actor 欠落を fail-closed で拒否し削除しない。
func TestClinicalPlanService_Delete_NilAuditOrTransactorOrActorRejected(t *testing.T) {
	deleteCalled := false
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{
				ID: 1, MedicalRecordID: 1, PhysicalExam: "残す所見",
				DiagnosisDetails: "残す診断", TreatmentPolicy: "残す方針", Version: 1,
			}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
	}

	t.Run("nil auditTx", func(t *testing.T) {
		deleteCalled = false
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, nil)
		err := svc.Delete(context.Background(), 1, 1, clinicalPlanTestActorID())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "audit dependency is required")
		assert.False(t, deleteCalled)
	})

	t.Run("nil transactor", func(t *testing.T) {
		deleteCalled = false
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), nil, &mockAuditTxLogger{})
		err := svc.Delete(context.Background(), 1, 1, clinicalPlanTestActorID())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transaction dependency is required")
		assert.False(t, deleteCalled)
	})

	t.Run("nil actor", func(t *testing.T) {
		deleteCalled = false
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})
		err := svc.Delete(context.Background(), 1, 1, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authenticated staff actor is required")
		assert.False(t, deleteCalled)
	})

	t.Run("zero actor", func(t *testing.T) {
		deleteCalled = false
		zero := uint64(0)
		svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, &mockAuditTxLogger{})
		err := svc.Delete(context.Background(), 1, 1, &zero)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authenticated staff actor is required")
		assert.False(t, deleteCalled)
	})
}

// BUG-010 residual: 削除成功時は削除前フィールド値と staff actor を audit に書く。
func TestClinicalPlanService_Delete_WritesAuditOnSuccess(t *testing.T) {
	deleteCalled := false
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{
				ID: 7, MedicalRecordID: 1,
				PhysicalExam: "削除前所見", DiagnosisDetails: "削除前診断", TreatmentPolicy: "削除前方針",
				Version: 3,
			}, nil
		},
		deleteFn: func(_ context.Context, clinicID, planID uint64) error {
			deleteCalled = true
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, uint64(7), planID)
			return nil
		},
	}
	var gotEntry *AuditEntry
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, entry *AuditEntry) error {
		gotEntry = entry
		return nil
	}}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, audit)

	err := svc.Delete(context.Background(), 1, 1, clinicalPlanTestActorID())
	require.NoError(t, err)
	assert.True(t, deleteCalled)
	require.NotNil(t, gotEntry)
	assert.Equal(t, model.AuditActionClinicalPlanDelete, gotEntry.Action)
	assert.Equal(t, model.AuditResourceClinicalPlan, gotEntry.Resource)
	assert.Equal(t, uint64(7), *gotEntry.ResourceID)
	require.NotNil(t, gotEntry.ActorID)
	assert.Equal(t, uint64(42), *gotEntry.ActorID)
	assert.Equal(t, model.AuditActorTypeStaff, gotEntry.ActorType)
	oldVal, ok := gotEntry.OldValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "削除前所見", oldVal["physical_exam"])
	assert.Equal(t, "削除前診断", oldVal["diagnosis_details"])
	assert.Equal(t, "削除前方針", oldVal["treatment_policy"])
	assert.Equal(t, 3, oldVal["version"])
	assert.Nil(t, gotEntry.NewValue)
	assert.Equal(t, "delete", gotEntry.Metadata.(map[string]any)["operation_type"])
}

// BUG-010 residual: audit 失敗時は error を返し、delete は同一 tx callback 内で先に走る。
func TestClinicalPlanService_Delete_AuditFailureReturnsError(t *testing.T) {
	deleteCalled := false
	auditCalled := false
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1, PhysicalExam: "旧"}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalled = true
		return errors.New("audit write failed")
	}}
	svc := NewClinicalPlanService(repo, okMedRecForPlan(), okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, audit)

	err := svc.Delete(context.Background(), 1, 1, clinicalPlanTestActorID())
	assert.Error(t, err)
	assert.True(t, deleteCalled, "delete runs before audit in the same tx callback")
	assert.True(t, auditCalled)
}

// BUG-010 residual: 拒否経路（finalized）では audit も delete も走らない。
func TestClinicalPlanService_Delete_FinalizedDoesNotAudit(t *testing.T) {
	auditCalled := false
	deleteCalled := false
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return &model.ClinicalPlan{ID: 1, MedicalRecordID: 1, PhysicalExam: "確定前"}, nil
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
	}
	medRec := &mockMedicalRecordRepository{
		findByIDFn: func(_ context.Context, _, _ uint64) (*model.MedicalRecord, error) {
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalled = true
		return nil
	}}
	svc := NewClinicalPlanService(repo, medRec, okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, audit)

	err := svc.Delete(context.Background(), 1, 1, clinicalPlanTestActorID())
	assert.Error(t, err)
	assert.True(t, apperrors.IsConflict(err))
	assert.False(t, deleteCalled)
	assert.False(t, auditCalled)
}

// BUG-010 residual: 対象不在では audit ゼロ・delete ゼロ。
func TestClinicalPlanService_Delete_NotFoundDoesNotAudit(t *testing.T) {
	auditCalled := false
	deleteCalled := false
	parentLockCalled := false
	repo := &mockClinicalPlanRepository{
		findByMedicalRecordIDFn: func(_ context.Context, _, _ uint64) (*model.ClinicalPlan, error) {
			return nil, apperrors.WrapNotFound("clinical_plan", "1")
		},
		deleteFn: func(_ context.Context, _, _ uint64) error {
			deleteCalled = true
			return nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		auditCalled = true
		return nil
	}}
	svc := NewClinicalPlanService(repo, &mockMedicalRecordRepository{
		lockByIDForUpdateFn: func(context.Context, uint64, uint64) (*model.MedicalRecord, error) {
			parentLockCalled = true
			return &model.MedicalRecord{Status: model.MedicalRecordStatusFinalized}, nil
		},
	}, okDiagnosisTypeRepo(), okDiagnosisNameRepo(), &mockCheckupTransactor{}, audit)

	err := svc.Delete(context.Background(), 1, 1, clinicalPlanTestActorID())
	assert.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
	assert.False(t, parentLockCalled)
	assert.False(t, deleteCalled)
	assert.False(t, auditCalled)
}

// BUG-010 residual DB: audit 失敗で clinical_plans 削除が rollback される（同一 DBOrTx）。
func TestClinicalPlanService_Delete_AuditFailureRollsBackDB(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	const clinicA = uint64(1)
	mr := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-DEL-AUDIT-RB")
	repo := NewClinicalPlanRepository(db)
	plan := &model.ClinicalPlan{
		MedicalRecordID:  mr.ID,
		PhysicalExam:     "delete-rollback-should-keep",
		DiagnosisDetails: "診断詳細",
		TreatmentPolicy:  "治療方針",
		Version:          1,
	}
	require.NoError(t, repo.Create(context.Background(), plan))

	failingAudit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		return errors.New("simulated clinical plan delete audit failure")
	}}
	svc := NewClinicalPlanService(
		repo,
		okMedRecForPlan(),
		okDiagnosisTypeRepo(),
		okDiagnosisNameRepo(),
		persistence.NewTransactor(db),
		failingAudit,
	)

	err := svc.Delete(context.Background(), clinicA, mr.ID, clinicalPlanTestActorID())
	require.Error(t, err)

	persisted, err := repo.FindByMedicalRecordID(context.Background(), clinicA, mr.ID)
	require.NoError(t, err)
	assert.Equal(t, "delete-rollback-should-keep", persisted.PhysicalExam)
	assert.Equal(t, 1, persisted.Version)
}

// BUG-010 residual DB: 成功時は削除と audit が commit される。
func TestClinicalPlanService_Delete_AuditSuccessCommitsDB(t *testing.T) {
	db := setupClinicalPlanTestDB(t)
	const clinicA = uint64(1)
	mr := makeClinicalPlanMedicalRecord(t, db, clinicA, "MR-CP-DEL-AUDIT-OK")
	repo := NewClinicalPlanRepository(db)
	plan := &model.ClinicalPlan{
		MedicalRecordID: mr.ID,
		PhysicalExam:    "to-delete",
		Version:         1,
	}
	require.NoError(t, repo.Create(context.Background(), plan))

	var auditEntries int
	var gotAction string
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, entry *AuditEntry) error {
		auditEntries++
		gotAction = entry.Action
		assert.Equal(t, model.AuditActionClinicalPlanDelete, entry.Action)
		require.NotNil(t, entry.OldValue)
		assert.Equal(t, "to-delete", entry.OldValue.(map[string]any)["physical_exam"])
		return nil
	}}
	svc := NewClinicalPlanService(
		repo,
		okMedRecForPlan(),
		okDiagnosisTypeRepo(),
		okDiagnosisNameRepo(),
		persistence.NewTransactor(db),
		audit,
	)

	err := svc.Delete(context.Background(), clinicA, mr.ID, clinicalPlanTestActorID())
	require.NoError(t, err)
	assert.Equal(t, 1, auditEntries)
	assert.Equal(t, model.AuditActionClinicalPlanDelete, gotAction)

	_, err = repo.FindByMedicalRecordID(context.Background(), clinicA, mr.ID)
	require.Error(t, err)
	assert.True(t, apperrors.IsNotFound(err))
}
