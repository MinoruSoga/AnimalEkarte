package medicalrecord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
	"github.com/animal-ekarte/backend/internal/persistence"
)

func TestExaminationPrintSnapshot_ConfirmedOfficialAndCrossClinicRejection(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicA = uint64(1)
	const clinicB = uint64(2)

	actorID := makeExaminationActor(t, db, clinicA, "print snapshot actor")
	examType := makeExamTypeMaster(t, db, clinicA, "print snapshot exam type")
	service := NewExaminationService(
		NewExaminationRepository(db),
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		persistence.NewTransactor(db),
	)

	items := []UpsertExamItemInput{{
		Name:            "WBC",
		InspectionValue: "12.5",
		Unit:            "10^3/uL",
		ReferenceValue:  "6-17",
		SortOrder:       1,
	}}
	confirmed, err := service.Create(ctx, clinicA, &CreateExaminationInput{
		ExamTypeID: examType.ID,
		Date:       time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC),
		Status:     model.ExaminationStatusConfirmed,
		ActorID:    &actorID,
		Items:      &items,
	})
	require.NoError(t, err)
	require.NotNil(t, confirmed.CurrentRevisionVersion)

	// Force stored is_assessed=true on the official revision item (print must not recompute).
	require.NoError(t, db.WithContext(ctx).
		Model(&model.ExaminationRevisionItem{}).
		Where("clinic_id = ? AND examination_id = ? AND version = ?", clinicA, confirmed.ID, *confirmed.CurrentRevisionVersion).
		Updates(map[string]any{
			"is_assessed": true,
			"is_abnormal": true,
			"status":      model.ExaminationResultStatusHigh,
		}).Error)

	snapshot, err := service.GetPrintSnapshot(ctx, clinicA, confirmed.ID, nil)
	require.NoError(t, err)
	require.NotNil(t, snapshot)
	assert.Equal(t, ExaminationPrintBoundaryOfficial, snapshot.PrintBoundary)
	assert.Empty(t, snapshot.Watermark)
	assert.Equal(t, model.ExaminationRevisionKindOfficial, snapshot.Kind)
	assert.Equal(t, model.ExaminationStatusConfirmed, snapshot.Status)
	assert.Equal(t, *confirmed.CurrentRevisionVersion, snapshot.Version)
	require.Len(t, snapshot.Items, 1)
	assert.Equal(t, "WBC", snapshot.Items[0].Name)
	assert.Equal(t, "12.5", snapshot.Items[0].InspectionValue)
	assert.True(t, snapshot.Items[0].IsAssessed, "print must use stored is_assessed, not recompute")
	assert.True(t, snapshot.Items[0].IsAbnormal)
	assert.Equal(t, model.ExaminationResultStatusHigh, snapshot.Items[0].Status)

	// Other-clinic must not leak existence.
	cross, err := service.GetPrintSnapshot(ctx, clinicB, confirmed.ID, nil)
	assert.True(t, apperrors.IsNotFound(err), "cross-clinic print must be NotFound")
	assert.Nil(t, cross)
}

func TestExaminationPrintSnapshot_DraftWatermarkAndUnconfirmedOldVersionIsolation(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	actorID := makeExaminationActor(t, db, clinicID, "print draft actor")
	examType := makeExamTypeMaster(t, db, clinicID, "print draft exam type")
	service := NewExaminationService(
		NewExaminationRepository(db),
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		persistence.NewTransactor(db),
	)

	officialItems := []UpsertExamItemInput{{
		Name: "ALT", InspectionValue: "40", SortOrder: 1,
	}}
	confirmed, err := service.Create(ctx, clinicID, &CreateExaminationInput{
		ExamTypeID:    examType.ID,
		Date:          time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC),
		ResultSummary: "official summary",
		Status:        model.ExaminationStatusConfirmed,
		ActorID:       &actorID,
		Items:         &officialItems,
	})
	require.NoError(t, err)
	officialVersion := *confirmed.CurrentRevisionVersion

	unconfirmed, err := service.Unconfirm(ctx, clinicID, confirmed.ID, UnconfirmExaminationInput{
		Reason: "print isolation test", ActorID: &actorID,
	})
	require.NoError(t, err)
	workingVersion := *unconfirmed.CurrentRevisionVersion
	assert.NotEqual(t, officialVersion, workingVersion)

	// Default (current) print is draft working revision.
	draft, err := service.GetPrintSnapshot(ctx, clinicID, confirmed.ID, nil)
	require.NoError(t, err)
	assert.Equal(t, ExaminationPrintBoundaryDraft, draft.PrintBoundary)
	assert.Equal(t, examinationPrintDraftWatermark, draft.Watermark)
	assert.Equal(t, workingVersion, draft.Version)
	assert.Equal(t, model.ExaminationRevisionKindWorking, draft.Kind)

	// Explicit old official version stays isolated from current draft content.
	official, err := service.GetPrintSnapshot(ctx, clinicID, confirmed.ID, &officialVersion)
	require.NoError(t, err)
	assert.Equal(t, ExaminationPrintBoundaryOfficial, official.PrintBoundary)
	assert.Empty(t, official.Watermark)
	assert.Equal(t, officialVersion, official.Version)
	assert.Equal(t, "official summary", official.ResultSummary)
	require.Len(t, official.Items, 1)
	assert.Equal(t, "ALT", official.Items[0].Name)
	assert.Equal(t, "40", official.Items[0].InspectionValue)

	// Mutate working parent summary/items — official print must not pick them up.
	newSummary := "dirty working only"
	_, err = service.Update(ctx, clinicID, confirmed.ID, UpdateExaminationInput{
		ResultSummary: &newSummary,
		Items: &[]UpsertExamItemInput{{
			Name: "ALT", InspectionValue: "999", SortOrder: 1,
		}},
		ActorID: &actorID,
	})
	require.NoError(t, err)

	officialAgain, err := service.GetPrintSnapshot(ctx, clinicID, confirmed.ID, &officialVersion)
	require.NoError(t, err)
	assert.Equal(t, "official summary", officialAgain.ResultSummary)
	require.Len(t, officialAgain.Items, 1)
	assert.Equal(t, "40", officialAgain.Items[0].InspectionValue, "old official must not include unsaved/current working values")
}

func TestExaminationPrintSnapshot_MissingRevisionFailsClosed(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	examType := makeExamTypeMaster(t, db, clinicID, "legacy print exam type")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID:   clinicID,
		ExamTypeID: examType.ID,
		Date:       time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC),
		Status:     model.ExaminationStatusCompleted,
	})
	service := NewExaminationService(
		NewExaminationRepository(db),
		&mockMedicalRecordRepository{},
		okExamTypeRepo(),
		&mockAuditTxLogger{},
		&mockCheckupTransactor{},
	)

	got, err := service.GetPrintSnapshot(ctx, clinicID, exam.ID, nil)
	assert.True(t, apperrors.IsNotFound(err))
	assert.Nil(t, got)
}

func TestExaminationPrintSnapshot_ResponseKeepsStoredAssessment(t *testing.T) {
	snapshot := &ExaminationPrintSnapshot{
		ExaminationID: 10,
		ClinicID:      1,
		Version:       2,
		Kind:          model.ExaminationRevisionKindOfficial,
		Status:        model.ExaminationStatusConfirmed,
		PrintBoundary: ExaminationPrintBoundaryOfficial,
		Date:          time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC),
		ExamTypeID:    3,
		Display: model.ExaminationDisplaySnapshot{
			ExamTypeName: "CBC",
			PetName:      "Taro",
		},
		Items: []ExaminationPrintItem{{
			ID:              99,
			Name:            "WBC",
			InspectionValue: "1",
			// Stored true even if values would not reassess as assessed.
			IsAssessed: true,
			IsAbnormal: false,
			Status:     model.ExaminationResultStatusNormal,
			SortOrder:  1,
		}},
	}

	resp := toExaminationPrintSnapshotResponse(snapshot)
	require.Len(t, resp.Items, 1)
	assert.True(t, resp.Items[0].IsAssessed)
	assert.Equal(t, "official", resp.PrintBoundary)
	assert.Equal(t, "Taro", resp.Display.PetName)
	assert.NotContains(t, mustJSON(t, resp), "danger_reason")
}

func TestGetExaminationPrintSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		setupCtx   func(*gin.Context)
		svc        *mockExaminationService
		wantStatus int
		wantBody   string
	}{
		{
			name:  "returns official print snapshot",
			query: "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				getPrintSnapshotFn: func(_ context.Context, clinicID, examinationID uint64, version *uint64) (*ExaminationPrintSnapshot, error) {
					assert.Equal(t, uint64(1), clinicID)
					assert.Equal(t, uint64(10), examinationID)
					assert.Nil(t, version)
					return &ExaminationPrintSnapshot{
						ExaminationID: examinationID,
						ClinicID:      clinicID,
						Version:       1,
						Kind:          model.ExaminationRevisionKindOfficial,
						Status:        model.ExaminationStatusConfirmed,
						PrintBoundary: ExaminationPrintBoundaryOfficial,
						Date:          time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC),
						ExamTypeID:    2,
						Display:       model.ExaminationDisplaySnapshot{ExamTypeName: "CBC"},
						Items: []ExaminationPrintItem{{
							ID: 1, Name: "WBC", InspectionValue: "10", IsAssessed: true,
							Status: model.ExaminationResultStatusNormal,
						}},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"print_boundary":"official"`,
		},
		{
			name:  "passes version query",
			query: "?version=3",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				getPrintSnapshotFn: func(_ context.Context, _, _ uint64, version *uint64) (*ExaminationPrintSnapshot, error) {
					require.NotNil(t, version)
					assert.Equal(t, uint64(3), *version)
					return &ExaminationPrintSnapshot{
						ExaminationID: 10, ClinicID: 1, Version: 3,
						Kind: model.ExaminationRevisionKindWorking, Status: model.ExaminationStatusCompleted,
						PrintBoundary: ExaminationPrintBoundaryDraft, Watermark: examinationPrintDraftWatermark,
						Date: time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC), ExamTypeID: 2,
						Display: model.ExaminationDisplaySnapshot{},
						Items:   []ExaminationPrintItem{},
					}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   examinationPrintDraftWatermark,
		},
		{
			name:       "returns 401 when clinic_id missing",
			query:      "",
			setupCtx:   func(c *gin.Context) {},
			svc:        &mockExaminationService{},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:  "returns 404 when not found",
			query: "",
			setupCtx: func(c *gin.Context) { setClinicID(c) },
			svc: &mockExaminationService{
				getPrintSnapshotFn: func(_ context.Context, _, _ uint64, _ *uint64) (*ExaminationPrintSnapshot, error) {
					return nil, apperrors.WrapNotFound("examination_print_snapshot", "missing")
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandlerWithExaminationSvc(tt.svc)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/"+tt.query, nil)
			c.Params = gin.Params{{Key: "id", Value: "10"}}
			tt.setupCtx(c)

			h.GetExaminationPrintSnapshot(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}
		})
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return string(b)
}
