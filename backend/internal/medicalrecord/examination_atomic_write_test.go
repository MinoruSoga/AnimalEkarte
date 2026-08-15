package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestExamTypeRepository_FindByIDLocksTypeAndFieldsInAmbientTransaction(t *testing.T) {
	db := setupExaminationTestDB(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const clinicID = uint64(1)

	examType := makeExamTypeMaster(t, db, clinicID, "血液検査（ownership lock）")
	otherExamType := makeExamTypeMaster(t, db, clinicID, "尿検査（remap target）")
	field := &model.ExamTypeField{
		ExamTypeID: examType.ID,
		ClinicID:   clinicID,
		Name:       "WBC",
	}
	require.NoError(t, db.Create(field).Error)
	repo := NewExamTypeRepository(db)

	expectBlocked := func(query string, args ...any) error {
		done := make(chan error, 1)
		go func() {
			competingTx := db.WithContext(ctx).Begin()
			if competingTx.Error != nil {
				done <- competingTx.Error
				return
			}
			defer competingTx.Rollback()
			if err := competingTx.Exec("SET LOCAL lock_timeout = '200ms'").Error; err != nil {
				done <- err
				return
			}
			done <- competingTx.Exec(query, args...).Error
		}()

		select {
		case err := <-done:
			if err == nil {
				return errors.New("concurrent master mutation completed before examination transaction commit")
			}
			if !strings.Contains(err.Error(), "lock timeout") {
				return err
			}
			return nil
		case <-ctx.Done():
			return errors.New("timed out waiting for bounded concurrent master mutation")
		}
	}

	err := (testTransactor{db: db}).WithTx(ctx, func(txCtx context.Context) error {
		loaded, err := repo.FindByID(txCtx, clinicID, examType.ID)
		if err != nil {
			return err
		}
		if len(loaded.Items) != 1 || loaded.Items[0].ID != field.ID {
			return errors.New("same-clinic exam type field was not loaded for ownership validation")
		}
		if err := expectBlocked(
			"UPDATE exam_types SET name = ? WHERE id = ?",
			"concurrent rename", examType.ID,
		); err != nil {
			return err
		}
		return expectBlocked(
			"UPDATE exam_type_fields SET exam_type_id = ? WHERE id = ?",
			otherExamType.ID, field.ID,
		)
	})
	require.NoError(t, err)

	// Commit 後は fresh transaction から同じ更新が待たずに実行できる。
	afterCommitTx := db.WithContext(ctx).Begin()
	require.NoError(t, afterCommitTx.Error)
	defer afterCommitTx.Rollback()
	require.NoError(t, afterCommitTx.Exec("SET LOCAL lock_timeout = '200ms'").Error)
	require.NoError(t, afterCommitTx.Exec(
		"UPDATE exam_types SET name = ? WHERE id = ?",
		"post-commit rename", examType.ID,
	).Error)
	require.NoError(t, afterCommitTx.Exec(
		"UPDATE exam_type_fields SET exam_type_id = ? WHERE id = ?",
		otherExamType.ID, field.ID,
	).Error)
}

func TestCreateExamination_WithItemsRollsBackParentWhenItemPersistenceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupExaminationTestDB(t)
	const clinicID = uint64(1)

	examType := makeExamTypeMaster(t, db, clinicID, "血液検査（combined create rollback）")
	repo := NewExaminationRepository(db)
	service := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		testTransactor{db: db},
	)
	handler := NewExaminationHandler(service)

	body := map[string]any{
		"exam_type_id": examType.ID,
		"date":         time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		"machine":      "combined-create-overflow",
		"items": []map[string]any{{
			"name":       "WBC",
			"sort_order": int64(1 << 40), // integer の上限を超える決定的な DB エラー
		}},
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/examinations", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	setClinicID(c)
	setStaffID(c)
	handler.CreateExamination(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var parentCount int64
	require.NoError(t, db.Model(&model.Examination{}).
		Where("clinic_id = ? AND machine = ?", clinicID, "combined-create-overflow").
		Count(&parentCount).Error)
	assert.Zero(t, parentCount, "items の永続化失敗時は作成した parent も rollback されなければならない")
}

func TestUpdateExamination_WithItemsRollsBackParentAndItemsWhenItemPersistenceFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupExaminationTestDB(t)
	ctx := context.Background()
	const clinicID = uint64(1)

	examType := makeExamTypeMaster(t, db, clinicID, "血液検査（combined update rollback）")
	exam := makeExaminationRec(t, db, &model.Examination{
		ClinicID:   clinicID,
		ExamTypeID: examType.ID,
		Date:       time.Date(2026, 7, 26, 0, 0, 0, 0, time.UTC),
		Machine:    "before-update",
	})
	repo := NewExaminationRepository(db)
	_, _, err := repo.ReplaceItemsByExamID(ctx, clinicID, exam.ID, []model.ExamResult{{
		Name:            "WBC",
		InspectionValue: "5.0",
	}})
	require.NoError(t, err)

	service := NewExaminationService(
		repo,
		&mockMedicalRecordRepository{},
		NewExamTypeRepository(db),
		&mockAuditTxLogger{},
		testTransactor{db: db},
	)
	handler := NewExaminationHandler(service)

	body := map[string]any{
		"machine": "after-update",
		"items": []map[string]any{{
			"name":       "Glucose",
			"sort_order": int64(1 << 40), // integer の上限を超える決定的な DB エラー
		}},
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/examinations/1", bytes.NewReader(payload))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(exam.ID, 10)}}
	setClinicID(c)
	setStaffID(c)
	handler.UpdateExamination(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	persisted, err := repo.FindByID(ctx, clinicID, exam.ID)
	require.NoError(t, err)
	assert.Equal(t, "before-update", persisted.Machine, "items 失敗時は parent 更新も rollback されなければならない")
	items, err := repo.FindAllItemsByExamID(ctx, clinicID, exam.ID)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "WBC", items[0].Name, "items 失敗時は置換前の明細が残らなければならない")
}

func TestRegisterRoutes_CreateExaminationRequiresCreateAndEditPermissions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	createCalled := false
	service := &mockExaminationService{
		createFn: func(_ context.Context, _ uint64, _ *CreateExaminationInput) (*model.Examination, error) {
			createCalled = true
			return &model.Examination{ID: 42}, nil
		},
	}
	requirePermission := func(resource, action string) gin.HandlerFunc {
		return func(c *gin.Context) {
			if resource == string(model.ResourceExaminations) && action == "edit" {
				c.AbortWithStatus(http.StatusForbidden)
			}
		}
	}
	handler := &Handler{
		examination:       NewExaminationHandler(service),
		requirePermission: requirePermission,
	}

	router := gin.New()
	api := router.Group("/api/v1")
	api.Use(func(c *gin.Context) {
		c.Set("clinic_id", "1")
	})
	handler.RegisterRoutes(api)

	body := `{"exam_type_id":1,"date":"2026-07-27T00:00:00Z","items":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/examinations", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, createCalled, "edit 権限がない場合は combined create handler へ到達してはならない")
}
