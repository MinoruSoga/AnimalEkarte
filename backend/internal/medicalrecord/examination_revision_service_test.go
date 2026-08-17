package medicalrecord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/animal-ekarte/backend/internal/reservation"
	"github.com/animal-ekarte/backend/internal/testdb"
)

const examinationRevisionMigrationTestSchema = "examination_revision_slice_a_test"

var errRollbackExaminationRevisionMigrationTest = errors.New("rollback examination revision migration test")

type examinationRevisionCapabilityRepository struct {
	ExaminationRepository
	appendOfficialRevisionFn func(
		ctx context.Context,
		clinicID, examinationID, actorID uint64,
		changeReason string,
	) (uint64, error)
	confirmWithRevisionCASFn func(
		ctx context.Context,
		clinicID, examinationID uint64,
		expectedStatus model.ExaminationStatus,
		version uint64,
	) (*model.Examination, error)
	findOfficialByIDFn  func(ctx context.Context, clinicID, examinationID uint64) (*ExaminationOfficialProjection, error)
	findPrintSnapshotFn func(ctx context.Context, clinicID, examinationID uint64, version *uint64) (*ExaminationPrintSnapshot, error)
}

func (r *examinationRevisionCapabilityRepository) AppendOfficialRevision(
	ctx context.Context,
	clinicID, examinationID, actorID uint64,
	changeReason string,
) (uint64, error) {
	return r.appendOfficialRevisionFn(ctx, clinicID, examinationID, actorID, changeReason)
}

func (r *examinationRevisionCapabilityRepository) ConfirmWithRevisionCAS(
	ctx context.Context,
	clinicID, examinationID uint64,
	expectedStatus model.ExaminationStatus,
	version uint64,
) (*model.Examination, error) {
	return r.confirmWithRevisionCASFn(ctx, clinicID, examinationID, expectedStatus, version)
}

func (r *examinationRevisionCapabilityRepository) FindOfficialByID(
	ctx context.Context,
	clinicID, examinationID uint64,
) (*ExaminationOfficialProjection, error) {
	return r.findOfficialByIDFn(ctx, clinicID, examinationID)
}

func (r *examinationRevisionCapabilityRepository) FindPrintSnapshot(
	ctx context.Context,
	clinicID, examinationID uint64,
	version *uint64,
) (*ExaminationPrintSnapshot, error) {
	if r.findPrintSnapshotFn != nil {
		return r.findPrintSnapshotFn(ctx, clinicID, examinationID, version)
	}
	return nil, apperrors.WrapNotFound("examination_print_snapshot", "mock")
}

func TestExaminationRevision_FirstConfirmAppendsBeforeAuditAndCAS(t *testing.T) {
	const (
		clinicID      = uint64(1)
		examinationID = uint64(10)
		actorID       = uint64(42)
	)

	events := make([]string, 0, 4)
	legacyStatusWrites := 0
	base := &mockExaminationRepository{
		lockByIDForUpdateFn: func(_ context.Context, gotClinicID, gotExaminationID uint64) (*model.Examination, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examinationID, gotExaminationID)
			return &model.Examination{
				ID: examinationID, ClinicID: clinicID, ExamTypeID: 7,
				Status: model.ExaminationStatusCompleted,
			}, nil
		},
		updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
			legacyStatusWrites++
			events = append(events, "legacy-status-update")
			return &model.Examination{
				ID: examinationID, ClinicID: clinicID, ExamTypeID: 7,
				Status: model.ExaminationStatusConfirmed,
			}, nil
		},
	}
	repo := &examinationRevisionCapabilityRepository{
		ExaminationRepository: base,
		appendOfficialRevisionFn: func(_ context.Context, gotClinicID, gotExaminationID, gotActorID uint64, reason string) (uint64, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examinationID, gotExaminationID)
			assert.Equal(t, actorID, gotActorID)
			assert.NotEmpty(t, reason)
			events = append(events, "revision")
			return 1, nil
		},
		confirmWithRevisionCASFn: func(_ context.Context, gotClinicID, gotExaminationID uint64, expectedStatus model.ExaminationStatus, version uint64) (*model.Examination, error) {
			assert.Equal(t, clinicID, gotClinicID)
			assert.Equal(t, examinationID, gotExaminationID)
			assert.Equal(t, model.ExaminationStatusCompleted, expectedStatus)
			assert.Equal(t, uint64(1), version)
			events = append(events, "cas")
			return &model.Examination{
				ID: examinationID, ClinicID: clinicID, ExamTypeID: 7,
				Status: model.ExaminationStatusConfirmed,
			}, nil
		},
	}
	audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
		events = append(events, "audit")
		return nil
	}}
	service := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		okExamTypeRepo(),
		audit,
		&mockCheckupTransactor{},
	)
	confirmed := model.ExaminationStatusConfirmed

	got, err := service.Update(context.Background(), clinicID, examinationID, UpdateExaminationInput{
		Status:  &confirmed,
		ActorID: ptrUint64(actorID),
	})

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, model.ExaminationStatusConfirmed, got.Status)
	assert.Equal(t, []string{"revision", "audit", "cas"}, events)
	assert.Zero(t, legacyStatusWrites, "confirmation must advance status and pointer through one CAS")
}

func TestExaminationRevision_ConfirmRejectsCrossClinicActorWithoutWrites(t *testing.T) {
	tests := []struct {
		name      string
		actorID   uint64
		appendErr error
	}{
		{
			name:      "actor belongs to another clinic",
			actorID:   9002,
			appendErr: apperrors.WrapNotFound("staff", "9002"),
		},
		{
			name:      "actor lookup fails",
			actorID:   42,
			appendErr: errors.New("actor lookup failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const (
				clinicID      = uint64(1)
				examinationID = uint64(10)
			)
			auditCalls := 0
			casCalls := 0
			legacyStatusWrites := 0
			base := &mockExaminationRepository{
				lockByIDForUpdateFn: func(_ context.Context, _, _ uint64) (*model.Examination, error) {
					return &model.Examination{
						ID: examinationID, ClinicID: clinicID, ExamTypeID: 7,
						Status: model.ExaminationStatusCompleted,
					}, nil
				},
				updateFieldsFn: func(_ context.Context, _, _ uint64, _ map[string]any) (*model.Examination, error) {
					legacyStatusWrites++
					return &model.Examination{ID: examinationID, ClinicID: clinicID, Status: model.ExaminationStatusConfirmed}, nil
				},
			}
			repo := &examinationRevisionCapabilityRepository{
				ExaminationRepository: base,
				appendOfficialRevisionFn: func(_ context.Context, _, _, gotActorID uint64, _ string) (uint64, error) {
					assert.Equal(t, tt.actorID, gotActorID)
					return 0, tt.appendErr
				},
				confirmWithRevisionCASFn: func(_ context.Context, _, _ uint64, _ model.ExaminationStatus, _ uint64) (*model.Examination, error) {
					casCalls++
					return nil, nil
				},
			}
			audit := &mockAuditTxLogger{logEntryTxFn: func(_ context.Context, _ *AuditEntry) error {
				auditCalls++
				return nil
			}}
			service := NewExaminationService(
				repo,
				&mockMedicalRecordRepository{},
				okExamTypeRepo(),
				audit,
				&mockCheckupTransactor{},
			)
			confirmed := model.ExaminationStatusConfirmed

			got, err := service.Update(context.Background(), clinicID, examinationID, UpdateExaminationInput{
				Status:  &confirmed,
				ActorID: ptrUint64(tt.actorID),
			})

			assert.Error(t, err)
			assert.Nil(t, got)
			assert.Zero(t, auditCalls)
			assert.Zero(t, casCalls)
			assert.Zero(t, legacyStatusWrites)
		})
	}
}

func TestExaminationRevision_OfficialReadCapabilityIsRevisionOnly(t *testing.T) {
	type officialReader interface {
		GetOfficialByID(ctx context.Context, clinicID, examinationID uint64) (*ExaminationOfficialProjection, error)
	}

	mutableLegacyReads := 0
	base := &mockExaminationRepository{
		findByIDFn: func(_ context.Context, clinicID, examinationID uint64) (*model.Examination, error) {
			mutableLegacyReads++
			return &model.Examination{
				ID: examinationID, ClinicID: clinicID,
				ResultSummary: "mutable legacy value",
				Status:        model.ExaminationStatusConfirmed,
			}, nil
		},
	}
	repo := &examinationRevisionCapabilityRepository{
		ExaminationRepository: base,
		findOfficialByIDFn: func(_ context.Context, clinicID, examinationID uint64) (*ExaminationOfficialProjection, error) {
			return &ExaminationOfficialProjection{
				Examination: model.Examination{
					ID: examinationID, ClinicID: clinicID,
					ResultSummary: "official revision value",
					Status:        model.ExaminationStatusConfirmed,
				},
				OfficialVersion: initialExaminationRevisionVersion,
			}, nil
		},
	}
	service := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		okExamTypeRepo(),
		&mockAuditTxLogger{},
		&mockCheckupTransactor{},
	)
	reader, ok := service.(officialReader)
	require.True(t, ok, "examination service must expose a revision-only official read capability")

	got, err := reader.GetOfficialByID(context.Background(), 1, 10)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "official revision value", got.ResultSummary)
	assert.Equal(t, initialExaminationRevisionVersion, got.OfficialVersion)
	assert.Zero(t, mutableLegacyReads, "official read must not fall back to mutable exams/exam_results")
}

func TestExaminationRevision_OfficialReadSeparatesOfficialVersionFromCurrentPointer(t *testing.T) {
	tests := []struct {
		name                  string
		officialVersion       uint64
		currentWorkingVersion uint64
		wantOfficialSummary   string
		currentWorkingSummary string
	}{
		{
			name:                  "official v1 remains identifiable while the parent selects working v2",
			officialVersion:       1,
			currentWorkingVersion: 2,
			wantOfficialSummary:   "immutable official v1",
			currentWorkingSummary: "editable working v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupExaminationTestDB(t)
			ctx := context.Background()
			const clinicID = uint64(1)

			actorID := makeExaminationActor(t, db, clinicID, "official projection actor")
			examType := makeExamTypeMaster(t, db, clinicID, "official projection exam type")
			exam := makeExaminationRec(t, db, &model.Examination{
				ClinicID:      clinicID,
				ExamTypeID:    examType.ID,
				Date:          time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
				ResultSummary: tt.currentWorkingSummary,
				Status:        model.ExaminationStatusCompleted,
			})
			displaySnapshot, err := json.Marshal(model.ExaminationDisplaySnapshot{
				ExamTypeName: examType.Name,
			})
			require.NoError(t, err)
			officialReason := examinationInitialConfirmReason
			workingReason := "unconfirmed_for_edit"
			require.NoError(t, db.WithContext(ctx).Create(&model.ExaminationRevision{
				ClinicID:        clinicID,
				ExaminationID:   exam.ID,
				Version:         tt.officialVersion,
				Kind:            model.ExaminationRevisionKindOfficial,
				Status:          model.ExaminationStatusConfirmed,
				ExamTypeID:      examType.ID,
				ActorID:         actorID,
				Date:            exam.Date,
				ResultSummary:   tt.wantOfficialSummary,
				DisplaySnapshot: displaySnapshot,
				SchemaVersion:   examinationRevisionSchemaVersion,
				ChangeReason:    &officialReason,
			}).Error)
			require.NoError(t, db.WithContext(ctx).Create(&model.ExaminationRevision{
				ClinicID:        clinicID,
				ExaminationID:   exam.ID,
				Version:         tt.currentWorkingVersion,
				Kind:            model.ExaminationRevisionKindWorking,
				Status:          model.ExaminationStatusCompleted,
				ExamTypeID:      examType.ID,
				ActorID:         actorID,
				Date:            exam.Date,
				ResultSummary:   tt.currentWorkingSummary,
				DisplaySnapshot: displaySnapshot,
				SchemaVersion:   examinationRevisionSchemaVersion,
				ChangeReason:    &workingReason,
			}).Error)
			require.NoError(t, db.WithContext(ctx).
				Model(&model.Examination{}).
				Where("clinic_id = ? AND id = ?", clinicID, exam.ID).
				Update("current_revision_version", tt.currentWorkingVersion).Error)

			service := NewExaminationService(
				NewExaminationRepository(db),
				&mockMedicalRecordRepository{},
				okExamTypeRepo(),
				&mockAuditTxLogger{},
				&mockCheckupTransactor{},
			)
			reader, ok := service.(ExaminationOfficialReader)
			require.True(t, ok, "official read must expose a dedicated projection with its own version identity")

			projection, err := reader.GetOfficialByID(ctx, clinicID, exam.ID)

			require.NoError(t, err)
			require.NotNil(t, projection)
			assert.Equal(t, tt.officialVersion, projection.OfficialVersion)
			assert.Equal(t, tt.wantOfficialSummary, projection.ResultSummary)
			assert.Nil(t, projection.CurrentRevisionVersion,
				"the immutable official snapshot must not repurpose the parent's current revision pointer")

			var parent model.Examination
			require.NoError(t, db.WithContext(ctx).
				Where("clinic_id = ? AND id = ?", clinicID, exam.ID).
				First(&parent).Error)
			require.NotNil(t, parent.CurrentRevisionVersion)
			assert.Equal(t, tt.currentWorkingVersion, *parent.CurrentRevisionVersion)
		})
	}
}

func TestExaminationRevision_UnconfirmWorkingEditItemsAndReconfirmAppendsVersions(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	actorID := makeExaminationActor(t, db, clinicID, "revision lifecycle actor")
	examType := makeExamTypeMaster(t, db, clinicID, "revision lifecycle exam type")
	auditEntries := make([]*AuditEntry, 0, 4)
	service := NewExaminationService(
		NewExaminationRepository(db),
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{logEntryTxFn: func(auditCtx context.Context, entry *AuditEntry) error {
			require.NotNil(t, persistence.TxFromContext(auditCtx))
			auditEntries = append(auditEntries, entry)
			return nil
		}},
		persistence.NewTransactor(db),
	)
	created, err := service.Create(ctx, clinicID, &CreateExaminationInput{
		ExamTypeID: examType.ID,
		Date:       time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
		Status:     model.ExaminationStatusConfirmed,
		ActorID:    &actorID,
	})
	require.NoError(t, err)
	require.NotNil(t, created.CurrentRevisionVersion)
	assert.Equal(t, uint64(1), *created.CurrentRevisionVersion)

	unconfirmed, err := service.Unconfirm(ctx, clinicID, created.ID, UnconfirmExaminationInput{
		Reason:  "  result correction requested  ",
		ActorID: &actorID,
	})
	require.NoError(t, err)
	require.NotNil(t, unconfirmed.CurrentRevisionVersion)
	assert.Equal(t, model.ExaminationStatusCompleted, unconfirmed.Status)
	assert.Equal(t, uint64(2), *unconfirmed.CurrentRevisionVersion)

	updatedSummary := "corrected working result"
	updatedItems := []UpsertExamItemInput{{
		Name: "manual result", InspectionValue: "positive", SortOrder: 1,
	}}
	updated, err := service.Update(ctx, clinicID, created.ID, UpdateExaminationInput{
		ResultSummary: &updatedSummary,
		Items:         &updatedItems,
		ActorID:       &actorID,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.CurrentRevisionVersion)
	assert.Equal(t, uint64(3), *updated.CurrentRevisionVersion)

	confirmedStatus := model.ExaminationStatusConfirmed
	reconfirmed, err := service.Update(ctx, clinicID, created.ID, UpdateExaminationInput{
		Status:  &confirmedStatus,
		ActorID: &actorID,
	})
	require.NoError(t, err)
	require.NotNil(t, reconfirmed.CurrentRevisionVersion)
	assert.Equal(t, model.ExaminationStatusConfirmed, reconfirmed.Status)
	assert.Equal(t, uint64(4), *reconfirmed.CurrentRevisionVersion)

	var revisions []model.ExaminationRevision
	require.NoError(t, db.Where("clinic_id = ? AND examination_id = ?", clinicID, created.ID).
		Order("version ASC").Find(&revisions).Error)
	require.Len(t, revisions, 4)
	assert.Equal(t, []model.ExaminationRevisionKind{
		model.ExaminationRevisionKindOfficial,
		model.ExaminationRevisionKindWorking,
		model.ExaminationRevisionKindWorking,
		model.ExaminationRevisionKindOfficial,
	}, []model.ExaminationRevisionKind{revisions[0].Kind, revisions[1].Kind, revisions[2].Kind, revisions[3].Kind})
	assert.Equal(t, "", revisions[0].ResultSummary, "the original official snapshot stays immutable")
	assert.Equal(t, "result correction requested", *revisions[1].ChangeReason)
	assert.Equal(t, updatedSummary, revisions[2].ResultSummary)
	assert.Equal(t, updatedSummary, revisions[3].ResultSummary)

	var latestItems []model.ExaminationRevisionItem
	require.NoError(t, db.Where(
		"clinic_id = ? AND examination_id = ? AND version = ?", clinicID, created.ID, uint64(4),
	).Find(&latestItems).Error)
	require.Len(t, latestItems, 1)
	assert.Equal(t, "manual result", latestItems[0].Name)

	require.Len(t, auditEntries, 4)
	assert.Equal(t, model.AuditActionExaminationUnconfirm, auditEntries[1].Action)
	unconfirmMetadata, ok := auditEntries[1].Metadata.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "result correction requested", unconfirmMetadata["reason"])
}

func TestExaminationRevision_UnconfirmRejectsWrongStatusAndLegacyConfirmed(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)
	actorID := makeExaminationActor(t, db, clinicID, "unconfirm rejection actor")
	examType := makeExamTypeMaster(t, db, clinicID, "unconfirm rejection exam type")
	service := NewExaminationService(
		NewExaminationRepository(db),
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		persistence.NewTransactor(db),
	)

	for _, tt := range []struct {
		name   string
		status model.ExaminationStatus
	}{
		{name: "completed is not confirmed", status: model.ExaminationStatusCompleted},
		{name: "legacy confirmed has no official pointer", status: model.ExaminationStatusConfirmed},
	} {
		t.Run(tt.name, func(t *testing.T) {
			exam := makeExaminationRec(t, db, &model.Examination{
				ClinicID: clinicID, ExamTypeID: examType.ID,
				Date: time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC), Status: tt.status,
			})

			got, err := service.Unconfirm(ctx, clinicID, exam.ID, UnconfirmExaminationInput{
				Reason: "result correction requested", ActorID: &actorID,
			})

			assert.True(t, apperrors.IsConflict(err))
			assert.Nil(t, got)
			assertExaminationRevisionRows(t, db, clinicID, exam.ID, 0, 0)
		})
	}
}

func TestExaminationRevision_UnconfirmAuditFailureRollsBackWorkingRevisionAndPointer(t *testing.T) {
	db := setupExaminationTestDB(t)
	require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AuditLog{}))
	ctx := context.Background()
	const clinicID = uint64(1)
	actorID := makeExaminationActor(t, db, clinicID, "unconfirm rollback actor")
	examType := makeExamTypeMaster(t, db, clinicID, "unconfirm rollback exam type")
	repo := NewExaminationRepository(db)
	creator := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		persistence.NewTransactor(db),
	)
	confirmed, err := creator.Create(ctx, clinicID, &CreateExaminationInput{
		ExamTypeID: examType.ID,
		Date:       time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
		Status:     model.ExaminationStatusConfirmed,
		ActorID:    &actorID,
	})
	require.NoError(t, err)
	require.NotNil(t, confirmed.CurrentRevisionVersion)

	failure := errors.New("injected unconfirm audit failure")
	auditMarker := fmt.Sprintf("examination-unconfirm-rollback-%d", confirmed.ID)
	auditWrites := 0
	service := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{logEntryTxFn: func(auditCtx context.Context, entry *AuditEntry) error {
			marshalValue := func(value any) json.RawMessage {
				if value == nil {
					return nil
				}
				encoded, marshalErr := json.Marshal(value)
				require.NoError(t, marshalErr)
				return encoded
			}
			auditLog := &model.AuditLog{
				ClinicID: entry.ClinicID, ActorID: entry.ActorID, ActorType: entry.ActorType,
				Action: entry.Action, Resource: entry.Resource, ResourceID: entry.ResourceID,
				OldValue: marshalValue(entry.OldValue), NewValue: marshalValue(entry.NewValue),
				Metadata: marshalValue(entry.Metadata), UserAgent: auditMarker,
			}
			require.NoError(t, persistence.DBOrTx(auditCtx, db).Create(auditLog).Error)
			auditWrites++
			return failure
		}},
		persistence.NewTransactor(db),
	)

	got, err := service.Unconfirm(ctx, clinicID, confirmed.ID, UnconfirmExaminationInput{
		Reason: "result correction requested", ActorID: &actorID,
	})

	assert.ErrorIs(t, err, failure)
	assert.Nil(t, got)
	assert.Equal(t, 1, auditWrites, "the durable audit write must occur inside the transaction before rollback")
	var auditCount int64
	require.NoError(t, db.Model(&model.AuditLog{}).Where("user_agent = ?", auditMarker).Count(&auditCount).Error)
	assert.Zero(t, auditCount)
	persisted, err := repo.FindByID(ctx, clinicID, confirmed.ID)
	require.NoError(t, err)
	assert.Equal(t, model.ExaminationStatusConfirmed, persisted.Status)
	require.NotNil(t, persisted.CurrentRevisionVersion)
	assert.Equal(t, uint64(1), *persisted.CurrentRevisionVersion)
	assertExaminationRevisionRows(t, db, clinicID, confirmed.ID, 1, 0)
}

func TestExaminationRevisionMigration_DeclaresTenantSafeAppendOnlyContract(t *testing.T) {
	ddl := readExaminationRevisionMigration(t)
	normalizedDDL := strings.Join(strings.Fields(ddl), " ")

	requiredFragments := []string{
		"CREATE TABLE examination_revisions",
		"CREATE TABLE examination_revision_items",
		"ADD COLUMN current_revision_version BIGINT NULL",
		"CHECK (current_revision_version >= 1)",
		"UNIQUE (clinic_id, examination_id, version)",
		"FOREIGN KEY (clinic_id, examination_id) REFERENCES exams (clinic_id, id) ON DELETE RESTRICT NOT DEFERRABLE",
		"FOREIGN KEY (clinic_id, examination_id, version) REFERENCES examination_revisions (clinic_id, examination_id, version) ON DELETE RESTRICT NOT DEFERRABLE",
		"FOREIGN KEY (clinic_id, id, current_revision_version) REFERENCES examination_revisions (clinic_id, examination_id, version) ON DELETE RESTRICT NOT DEFERRABLE",
		"ON examination_revisions (clinic_id, examination_id, kind, version DESC)",
		"ON examination_revision_items (clinic_id, examination_id, version, sort_order, id)",
		"ON exams (clinic_id, id, current_revision_version)",
		"BEFORE UPDATE OR DELETE ON examination_revisions",
		"BEFORE UPDATE OR DELETE ON examination_revision_items",
		"BEFORE INSERT ON examination_revision_items",
		"schema_version SMALLINT NOT NULL DEFAULT 1",
		"CONSTRAINT ck_examination_revisions_schema_version CHECK (schema_version = 1)",
		"change_reason TEXT NULL",
		"CONSTRAINT ck_examination_revisions_change_reason CHECK (change_reason IS NULL OR btrim(change_reason) <> '')",
		"CONSTRAINT ck_examination_revision_items_reference_range CHECK (ref_min IS NULL OR ref_max IS NULL OR ref_min <= ref_max)",
		"(kind = 'official' AND status = 'confirmed')",
		"(kind = 'working' AND status IN ('pending', 'in_progress', 'result_entered', 'completed'))",
	}
	for _, fragment := range requiredFragments {
		t.Run(fragment, func(t *testing.T) {
			assert.Contains(t, normalizedDDL, fragment)
		})
	}
	for _, key := range []string{
		"medical_record_no",
		"pet_name",
		"medical_record_owner_name",
		"pet_owner_name",
		"species_name",
		"exam_type_name",
		"doctor_name",
	} {
		t.Run("display snapshot "+key, func(t *testing.T) {
			assert.Contains(t, ddl, "display_snapshot -> '"+key+"'")
			assert.Contains(t, ddl, "jsonb_typeof(display_snapshot -> '"+key+"')")
		})
	}
	assert.Equal(t, 3, strings.Count(ddl, "FOREIGN KEY ("), "004 must add exactly the three reviewed composite foreign keys")
	assert.Equal(t, 2, strings.Count(ddl, "SELECT app_private.apply_rls_policy("))
	assert.NotContains(t, normalizedDDL, "UNIQUE (clinic_id, examination_id, version, sort_order, id)")
	assert.NotContains(t, ddl, "FORCE ROW LEVEL SECURITY")
}

func TestExaminationRevisionMigration_RejectsUpdateAndDelete(t *testing.T) {
	withExaminationRevisionMigrationSchema(t, func(db *gorm.DB) {
		tests := []struct {
			name string
			run  func(*gorm.DB) error
		}{
			{
				name: "revision update",
				run: func(db *gorm.DB) error {
					return db.Exec("UPDATE examination_revisions SET result_summary = 'mutated' WHERE id = 1").Error
				},
			},
			{
				name: "revision delete",
				run: func(db *gorm.DB) error {
					return db.Exec("DELETE FROM examination_revisions WHERE id = 1").Error
				},
			},
			{
				name: "revision item update",
				run: func(db *gorm.DB) error {
					return db.Exec("UPDATE examination_revision_items SET inspection_value = 'mutated' WHERE id = 1").Error
				},
			},
			{
				name: "revision item delete",
				run: func(db *gorm.DB) error {
					return db.Exec("DELETE FROM examination_revision_items WHERE id = 1").Error
				},
			},
		}

		for i, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				savepoint := fmt.Sprintf("examination_revision_case_%d", i)
				require.NoError(t, db.Exec("SAVEPOINT "+savepoint).Error)

				err := tt.run(db)
				require.Error(t, err)
				assertPostgresCode(t, err, "23514")

				require.NoError(t, db.Exec("ROLLBACK TO SAVEPOINT "+savepoint).Error)
			})
		}
	})
}

func TestExaminationRevision_FirstConfirmPersistsImmutableOfficialSnapshot(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	recordOwner := makeTestOwner(t, db, clinicID, "record owner before")
	petOwner := makeTestOwner(t, db, clinicID, "pet owner before")
	pet := makeSpeciesAndPet(t, db, clinicID, petOwner.ID, "pet before")
	record := makeClinicalRelationRecord(t, db, clinicID, recordOwner.ID, pet.ID, "MR-BEFORE")
	doctor := makeDoctor(t, db, clinicID, "doctor before")
	require.NoError(t, db.Create(&model.StaffClinicAssignment{StaffID: doctor.ID, ClinicID: clinicID}).Error)
	actorID := makeExaminationActor(t, db, clinicID, "revision actor")
	examType := makeExamTypeMaster(t, db, clinicID, "exam type before")
	field := &model.ExamTypeField{
		ClinicID: clinicID, ExamTypeID: examType.ID, Name: "WBC", Unit: "10^3/uL", SortOrder: 1,
	}
	require.NoError(t, db.Create(field).Error)
	referenceMin, referenceMax := 4.0, 10.0
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID: clinicID, MedicalRecordID: &record.ID, PetID: &pet.ID,
		ExamTypeID: examType.ID, DoctorID: &doctor.ID,
		Date:          time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
		ResultSummary: "summary before", Machine: "machine before",
		Status: model.ExaminationStatusCompleted,
	})
	require.NoError(t, db.Create(&model.ExamResult{
		ExamID: exam.ID, ExamTypeItemID: &field.ID, Name: "WBC",
		InspectionValue: "11.0", Unit: "10^3/uL", RefMin: &referenceMin, RefMax: &referenceMax,
		// Deliberately stale derived values prove first-confirm recomputes the immutable snapshot.
		IsAbnormal: false, Status: model.ExaminationResultStatusNormal, SortOrder: 1,
	}).Error)

	var auditEntry *AuditEntry
	repo := NewExaminationRepository(db)
	service := NewExaminationService(
		repo,
		NewMedicalRecordRepository(db),
		NewExamTypeRepository(db),
		&mockAuditTxLogger{logEntryTxFn: func(auditCtx context.Context, entry *AuditEntry) error {
			require.NotNil(t, persistence.TxFromContext(auditCtx))
			auditEntry = entry
			return nil
		}},
		persistence.NewTransactor(db),
		reservation.NewReservationRepository(db),
	)
	confirmedStatus := model.ExaminationStatusConfirmed

	confirmed, err := service.Update(ctx, clinicID, exam.ID, UpdateExaminationInput{
		Status: &confirmedStatus, ActorID: &actorID,
	})

	require.NoError(t, err)
	require.NotNil(t, confirmed)
	require.NotNil(t, confirmed.CurrentRevisionVersion)
	assert.Equal(t, initialExaminationRevisionVersion, *confirmed.CurrentRevisionVersion)
	assert.Equal(t, model.ExaminationStatusConfirmed, confirmed.Status)
	require.NotNil(t, auditEntry)
	assert.Equal(t, model.AuditActionExaminationConfirm, auditEntry.Action)
	auditAfter, ok := auditEntry.NewValue.(map[string]any)
	require.True(t, ok)
	assert.Equal(t, initialExaminationRevisionVersion, auditAfter["current_revision_version"])

	var revision model.ExaminationRevision
	require.NoError(t, db.Where(
		"clinic_id = ? AND examination_id = ? AND version = ?",
		clinicID,
		exam.ID,
		initialExaminationRevisionVersion,
	).First(&revision).Error)
	assert.Equal(t, model.ExaminationRevisionKindOfficial, revision.Kind)
	assert.Equal(t, model.ExaminationStatusConfirmed, revision.Status)
	assert.Equal(t, recordOwner.ID, *revision.MedicalRecordOwnerID)
	assert.Equal(t, petOwner.ID, *revision.PetOwnerID)
	assert.Equal(t, pet.AnimalSpeciesID, *revision.AnimalSpeciesID)
	assert.Equal(t, actorID, revision.ActorID)
	var persistedChangeReason string
	require.NoError(t, db.Model(&model.ExaminationRevision{}).
		Select("change_reason").
		Where("id = ?", revision.ID).
		Scan(&persistedChangeReason).Error)
	assert.Equal(t, examinationInitialConfirmReason, persistedChangeReason)

	var revisionItems []model.ExaminationRevisionItem
	require.NoError(t, db.Where(
		"clinic_id = ? AND examination_id = ? AND version = ?",
		clinicID,
		exam.ID,
		initialExaminationRevisionVersion,
	).Find(&revisionItems).Error)
	require.Len(t, revisionItems, 1)
	assert.Equal(t, "11.0", revisionItems[0].InspectionValue)
	assert.True(t, revisionItems[0].IsAssessed)
	assert.True(t, revisionItems[0].IsAbnormal)
	assert.Equal(t, model.ExaminationResultStatusHigh, revisionItems[0].Status)

	mutateExaminationRevisionSources(t, db, recordOwner, petOwner, pet, record, doctor, examType, exam)
	reader, ok := service.(ExaminationOfficialReader)
	require.True(t, ok)

	official, err := reader.GetOfficialByID(ctx, clinicID, exam.ID)

	require.NoError(t, err)
	require.NotNil(t, official)
	assert.Equal(t, initialExaminationRevisionVersion, official.OfficialVersion)
	assert.Nil(t, official.CurrentRevisionVersion)
	assert.Equal(t, "summary before", official.ResultSummary)
	assert.Equal(t, "machine before", official.Machine)
	require.NotNil(t, official.MedicalRecord)
	assert.Equal(t, "MR-BEFORE", official.MedicalRecord.RecordNo)
	require.NotNil(t, official.MedicalRecord.Owner)
	assert.Equal(t, "record owner before", official.MedicalRecord.Owner.Name)
	require.NotNil(t, official.Pet)
	assert.Equal(t, "pet before", official.Pet.Name)
	require.NotNil(t, official.Pet.Owner)
	assert.Equal(t, "pet owner before", official.Pet.Owner.Name)
	require.NotNil(t, official.Pet.AnimalSpecies)
	assert.Equal(t, "犬", official.Pet.AnimalSpecies.Name)
	require.NotNil(t, official.Doctor)
	assert.Equal(t, "doctor before", official.Doctor.Name)
	require.NotNil(t, official.ExaminationType)
	assert.Equal(t, "exam type before", official.ExaminationType.Name)
	require.Len(t, official.Items, 1)
	assert.Equal(t, "11.0", official.Items[0].InspectionValue)

	t.Run("CrossClinic official read is not found", func(t *testing.T) {
		foreign, readErr := reader.GetOfficialByID(ctx, clinicID+1, exam.ID)
		assert.True(t, apperrors.IsNotFound(readErr))
		assert.Nil(t, foreign)
	})

	t.Run("legacy confirmed row without revision fails closed", func(t *testing.T) {
		legacy := makeExaminationRec(t, db, &model.Examination{
			ClinicID: clinicID, ExamTypeID: examType.ID,
			Date:   time.Date(2026, time.August, 2, 0, 0, 0, 0, time.UTC),
			Status: model.ExaminationStatusConfirmed,
		})
		got, readErr := reader.GetOfficialByID(ctx, clinicID, legacy.ID)
		assert.True(t, apperrors.IsNotFound(readErr))
		assert.Nil(t, got)
	})
}

func TestExaminationRevision_ConfirmFailureRollsBackRevisionAuditAndCAS(t *testing.T) {
	tests := []struct {
		name            string
		failureStage    string
		wantAuditCalls  int
		wantAuditWrites int
		wantCASCalls    int
	}{
		{
			name:            "revision append reports failure after writes",
			failureStage:    "revision",
			wantAuditCalls:  0,
			wantAuditWrites: 0,
			wantCASCalls:    0,
		},
		{
			name:            "audit failure after revision append",
			failureStage:    "audit",
			wantAuditCalls:  1,
			wantAuditWrites: 0,
			wantCASCalls:    0,
		},
		{
			name:            "CAS failure after revision and audit",
			failureStage:    "cas",
			wantAuditCalls:  1,
			wantAuditWrites: 1,
			wantCASCalls:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupExaminationTestDB(t)
			require.NoError(t, testdb.EnsureAutoMigrated(db, &model.AuditLog{}))
			ctx := context.Background()
			const clinicID = uint64(1)
			failure := errors.New("injected " + tt.failureStage + " failure")
			actorID := makeExaminationActor(t, db, clinicID, "rollback actor")
			examType := makeExamTypeMaster(t, db, clinicID, "rollback exam type")
			field := &model.ExamTypeField{
				ClinicID: clinicID, ExamTypeID: examType.ID, Name: "WBC", SortOrder: 1,
			}
			require.NoError(t, db.Create(field).Error)
			exam := makeExaminationRec(t, db, &model.Examination{
				ClinicID: clinicID, ExamTypeID: examType.ID,
				Date:   time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
				Status: model.ExaminationStatusCompleted,
			})
			auditMarker := fmt.Sprintf("examination-confirm-rollback-%s-%d", tt.failureStage, exam.ID)
			require.NoError(t, db.Create(&model.ExamResult{
				ExamID: exam.ID, ExamTypeItemID: &field.ID, Name: "WBC",
				InspectionValue: "5.0", Status: model.ExaminationResultStatusNormal,
			}).Error)

			base := NewExaminationRepository(db)
			concreteRevisions, ok := base.(ExaminationRevisionRepository)
			require.True(t, ok)
			auditCalls, auditWrites, casCalls := 0, 0, 0
			repo := &examinationRevisionCapabilityRepository{
				ExaminationRepository: base,
				appendOfficialRevisionFn: func(
					appendCtx context.Context,
					gotClinicID, examinationID, gotActorID uint64,
					changeReason string,
				) (uint64, error) {
					version, appendErr := concreteRevisions.AppendOfficialRevision(
						appendCtx,
						gotClinicID,
						examinationID,
						gotActorID,
						changeReason,
					)
					if appendErr != nil {
						return 0, appendErr
					}
					if tt.failureStage == "revision" {
						return 0, failure
					}
					return version, nil
				},
				confirmWithRevisionCASFn: func(
					casCtx context.Context,
					gotClinicID, examinationID uint64,
					expectedStatus model.ExaminationStatus,
					version uint64,
				) (*model.Examination, error) {
					casCalls++
					if tt.failureStage == "cas" {
						return nil, failure
					}
					return concreteRevisions.ConfirmWithRevisionCAS(
						casCtx,
						gotClinicID,
						examinationID,
						expectedStatus,
						version,
					)
				},
				findOfficialByIDFn: concreteRevisions.FindOfficialByID,
			}
			service := NewExaminationService(
				repo,
				&mockMedicalRecordRepository{},
				NewExamTypeRepository(db),
				&mockAuditTxLogger{logEntryTxFn: func(auditCtx context.Context, entry *AuditEntry) error {
					auditCalls++
					if tt.failureStage == "audit" {
						return failure
					}
					marshalAuditValue := func(value any) (json.RawMessage, error) {
						if value == nil {
							return nil, nil
						}
						encoded, marshalErr := json.Marshal(value)
						if marshalErr != nil {
							return nil, marshalErr
						}
						return json.RawMessage(encoded), nil
					}
					oldValue, marshalErr := marshalAuditValue(entry.OldValue)
					if marshalErr != nil {
						return fmt.Errorf("marshal rollback audit old value: %w", marshalErr)
					}
					newValue, marshalErr := marshalAuditValue(entry.NewValue)
					if marshalErr != nil {
						return fmt.Errorf("marshal rollback audit new value: %w", marshalErr)
					}
					metadata, marshalErr := marshalAuditValue(entry.Metadata)
					if marshalErr != nil {
						return fmt.Errorf("marshal rollback audit metadata: %w", marshalErr)
					}
					auditLog := &model.AuditLog{
						ClinicID:   entry.ClinicID,
						ActorID:    entry.ActorID,
						ActorType:  entry.ActorType,
						Action:     entry.Action,
						Resource:   entry.Resource,
						ResourceID: entry.ResourceID,
						OldValue:   oldValue,
						NewValue:   newValue,
						Metadata:   metadata,
						UserAgent:  auditMarker,
					}
					if createErr := persistence.DBOrTx(auditCtx, db).Create(auditLog).Error; createErr != nil {
						return createErr
					}
					auditWrites++
					assert.NotZero(t, auditLog.ID)
					return nil
				}},
				persistence.NewTransactor(db),
			)
			confirmedStatus := model.ExaminationStatusConfirmed

			got, err := service.Update(ctx, clinicID, exam.ID, UpdateExaminationInput{
				Status: &confirmedStatus, ActorID: &actorID,
			})

			assert.ErrorIs(t, err, failure)
			assert.Nil(t, got)
			assert.Equal(t, tt.wantAuditCalls, auditCalls)
			assert.Equal(t, tt.wantAuditWrites, auditWrites)
			assert.Equal(t, tt.wantCASCalls, casCalls)
			var persistedAuditCount int64
			require.NoError(t, db.Model(&model.AuditLog{}).
				Where("user_agent = ?", auditMarker).
				Count(&persistedAuditCount).Error)
			assert.Zero(t, persistedAuditCount)
			persisted, findErr := base.FindByID(ctx, clinicID, exam.ID)
			require.NoError(t, findErr)
			assert.Equal(t, model.ExaminationStatusCompleted, persisted.Status)
			assert.Nil(t, persisted.CurrentRevisionVersion)
			assertExaminationRevisionRows(t, db, clinicID, exam.ID, 0, 0)
			var mutableItemCount int64
			require.NoError(t, db.Model(&model.ExamResult{}).
				Where("exam_id = ?", exam.ID).
				Count(&mutableItemCount).Error)
			assert.Equal(t, int64(1), mutableItemCount)
		})
	}
}

func readExaminationRevisionMigration(t *testing.T) string {
	t.Helper()
	// 004_examination_revisions.sql was folded into 001_init.sql; extract that section.
	initPath := filepath.Join("..", "..", "migrations", "001_init.sql")
	raw, err := os.ReadFile(initPath)
	require.NoError(t, err, "001_init.sql must contain examination revision DDL")
	text := string(raw)
	const startMarker = "-- Source file: 004_examination_revisions.sql"
	start := strings.Index(text, startMarker)
	require.GreaterOrEqual(t, start, 0, "004 examination revision section missing from 001_init.sql")
	rest := text[start:]
	// Next consolidated source marker after this block (if any).
	next := strings.Index(rest[len(startMarker):], "\n-- Source file: ")
	if next >= 0 {
		rest = rest[:len(startMarker)+next]
	}
	return rest
}

func withExaminationRevisionMigrationSchema(t *testing.T, fn func(*gorm.DB)) {
	t.Helper()
	db := testdb.SetupIsolatedTestDB(t)

	err := db.Transaction(func(tx *gorm.DB) error {
		require.NoError(t, tx.Exec("CREATE SCHEMA "+examinationRevisionMigrationTestSchema).Error)
		require.NoError(t, tx.Exec(
			"SET LOCAL search_path TO "+examinationRevisionMigrationTestSchema+", public",
		).Error)
		installExaminationRevisionRLSFunctions(t, tx)
		require.NoError(t, tx.Exec(`
			CREATE TABLE exams (
				id BIGINT PRIMARY KEY,
				clinic_id BIGINT NOT NULL,
				status exam_status NOT NULL
			)
		`).Error)

		require.NoError(t, tx.Exec(readExaminationRevisionMigration(t)).Error)
		require.NoError(t, tx.Exec(`
			INSERT INTO exams (id, clinic_id, status)
			VALUES (10, 1, 'completed')
		`).Error)
		require.NoError(t, tx.Exec(`
			INSERT INTO examination_revisions (
				id, clinic_id, examination_id, version, kind, status,
				exam_type_id, actor_id, date, display_snapshot, change_reason
			) VALUES (
				1, 1, 10, 1, 'official', 'confirmed',
				100, 42, DATE '2026-08-03',
				'{
					"medical_record_no": "MR-001",
					"pet_name": "patient",
					"medical_record_owner_name": "record owner",
					"pet_owner_name": "pet owner",
					"species_name": "dog",
					"exam_type_name": "blood",
					"doctor_name": "doctor"
				}'::jsonb,
				'initial_confirmation'
			)
		`).Error)
		require.NoError(t, tx.Exec(`
			INSERT INTO examination_revision_items (
				id, clinic_id, examination_id, version, name,
				is_assessed, is_abnormal, status
			) VALUES (1, 1, 10, 1, 'WBC', TRUE, FALSE, 'normal')
		`).Error)
		require.NoError(t, tx.Exec(`
			UPDATE exams
			SET status = 'confirmed', current_revision_version = 1
			WHERE id = 10 AND clinic_id = 1
		`).Error)

		fn(tx)
		return errRollbackExaminationRevisionMigrationTest
	})
	require.ErrorIs(t, err, errRollbackExaminationRevisionMigrationTest)
}

func mutateExaminationRevisionSources(
	t *testing.T,
	db *gorm.DB,
	recordOwner, petOwner *model.Owner,
	pet *model.Pet,
	record *model.MedicalRecord,
	doctor *model.Staff,
	examType *model.ExaminationType,
	exam *model.Examination,
) {
	t.Helper()
	require.NoError(t, db.Model(recordOwner).Update("name", "record owner after").Error)
	require.NoError(t, db.Model(petOwner).Update("name", "pet owner after").Error)
	require.NoError(t, db.Model(pet).Update("name", "pet after").Error)
	require.NoError(t, db.Model(&model.AnimalSpecies{}).
		Where("id = ?", pet.AnimalSpeciesID).
		Update("name", "species after").Error)
	require.NoError(t, db.Model(record).Update("record_no", "MR-AFTER").Error)
	require.NoError(t, db.Model(doctor).Update("name", "doctor after").Error)
	require.NoError(t, db.Model(examType).Update("name", "exam type after").Error)
	require.NoError(t, db.Model(exam).Updates(map[string]any{
		"result_summary": "summary after",
		"machine":        "machine after",
	}).Error)
	require.NoError(t, db.Model(&model.ExamResult{}).
		Where("exam_id = ?", exam.ID).
		Update("inspection_value", "2.0").Error)
}

func assertExaminationRevisionRows(
	t *testing.T,
	db *gorm.DB,
	clinicID, examinationID uint64,
	wantRevisions, wantItems int64,
) {
	t.Helper()
	var revisionCount, itemCount int64
	require.NoError(t, db.Model(&model.ExaminationRevision{}).
		Where("clinic_id = ? AND examination_id = ?", clinicID, examinationID).
		Count(&revisionCount).Error)
	require.NoError(t, db.Model(&model.ExaminationRevisionItem{}).
		Where("clinic_id = ? AND examination_id = ?", clinicID, examinationID).
		Count(&itemCount).Error)
	assert.Equal(t, wantRevisions, revisionCount)
	assert.Equal(t, wantItems, itemCount)
}
