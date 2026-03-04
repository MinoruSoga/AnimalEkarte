package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func setupMedicalRecordTestRouter() (*gin.Engine, *MockService) {
	gin.SetMode(gin.TestMode)

	mockService := &MockService{}
	handler := &Handler{
		svc: mockService,
	}

	router := gin.New()
	router.GET("/medical-records", handler.GetAllMedicalRecords)
	router.GET("/medical-records/:id", handler.GetMedicalRecord)
	router.POST("/medical-records", handler.CreateMedicalRecord)
	router.PUT("/medical-records/:id", handler.UpdateMedicalRecord)
	router.DELETE("/medical-records/:id", handler.DeleteMedicalRecord)
	router.GET("/medical-records/paginated", handler.GetMedicalRecordsWithPagination)

	return router, mockService
}

func TestGetAllMedicalRecords(t *testing.T) {
	router, mockService := setupMedicalRecordTestRouter()

	testID1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	petID1 := uuid.MustParse("87a6e7a4-8b70-44b7-b7b5-e0b7e3b4e2bd")
	ownerID1 := uuid.MustParse("a26c8688-6699-4bb1-8217-2a5630d2925e")

	testID2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	petID2 := uuid.MustParse("87a6e7a4-8b70-44b7-b7b5-e0b7e3b4e2be")
	ownerID2 := uuid.MustParse("a26c8688-6699-4bb1-8217-2a5630d2925f")

	expectedRecords := []model.MedicalRecord{
		{
			ID:        testID1,
			RecordNo:  "MR123456",
			PetID:     petID1,
			OwnerID:   ownerID1,
			VisitDate: time.Now(),
			Status:    "確定済",
		},
		{
			ID:        testID2,
			RecordNo:  "MR123457",
			PetID:     petID2,
			OwnerID:   ownerID2,
			VisitDate: time.Now(),
			Status:    "作成中",
		},
	}

	mockService.On("GetAllMedicalRecords", mock.Anything).Return(expectedRecords, nil)

	req, _ := http.NewRequest("GET", "/medical-records", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []model.MedicalRecord
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, expectedRecords[0].ID, response[0].ID)

	mockService.AssertExpectations(t)
}

func TestGetMedicalRecordByID(t *testing.T) {
	router, mockService := setupMedicalRecordTestRouter()

	testID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	petID := uuid.MustParse("87a6e7a4-8b70-44b7-b7b5-e0b7e3b4e2bd")
	ownerID := uuid.MustParse("a26c8688-6699-4bb1-8217-2a5630d2925e")

	expectedRecord := &model.MedicalRecord{
		ID:        testID,
		RecordNo:  "MR123456",
		PetID:     petID,
		OwnerID:   ownerID,
		VisitDate: time.Now(),
		Status:    "確定済",
	}

	mockService.On("GetMedicalRecordByID", mock.Anything, "550e8400-e29b-41d4-a716-446655440000").Return(expectedRecord, nil)

	req, _ := http.NewRequest("GET", "/medical-records/550e8400-e29b-41d4-a716-446655440000", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.MedicalRecord
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedRecord.ID, response.ID)

	mockService.AssertExpectations(t)
}

func TestCreateMedicalRecord(t *testing.T) {
	router, mockService := setupMedicalRecordTestRouter()

	createReq := &model.CreateMedicalRecordRequest{
		PetID:          "87a6e7a4-8b70-44b7-b7b5-e0b7e3b4e2bd",
		OwnerID:        "a26c8688-6699-4bb1-8217-2a5630d2925e",
		VisitDate:      "2026-01-25T10:00:00Z",
		VisitType:      "初診",
		ChiefComplaint: "食欲不振",
		Subjective:     "飼い主によると昨日から食欲がない",
		Objective:      "体温正常、心拍正常",
		Assessment:     "軽度の消化器系問題",
		Plan:           "経過観察、食事指導",
		Diagnosis:      "軽度胃炎",
		Treatment:      "消化剤処方",
		Status:         "作成中",
	}

	newID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	petID := uuid.MustParse("87a6e7a4-8b70-44b7-b7b5-e0b7e3b4e2bd")
	ownerID := uuid.MustParse("a26c8688-6699-4bb1-8217-2a5630d2925e")

	expectedRecord := &model.MedicalRecord{
		ID:        newID,
		RecordNo:  "MR123456",
		PetID:     petID,
		OwnerID:   ownerID,
		VisitDate: time.Now(),
		Status:    createReq.Status,
	}

	mockService.On("CreateMedicalRecord", mock.Anything, createReq).Return(expectedRecord, nil)

	reqBody, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/medical-records", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.MedicalRecord
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedRecord.ID, response.ID)

	mockService.AssertExpectations(t)
}

func TestUpdateMedicalRecord(t *testing.T) {
	router, mockService := setupMedicalRecordTestRouter()

	status := "確定済"
	treatment := "消化剤処方＋食事改善指導"
	notes := "飼い主に経過観察の重要性を説明"

	updateReq := &model.UpdateMedicalRecordRequest{
		Status:    &status,
		Treatment: &treatment,
		Notes:     &notes,
	}

	testID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	petID := uuid.MustParse("87a6e7a4-8b70-44b7-b7b5-e0b7e3b4e2bd")
	ownerID := uuid.MustParse("a26c8688-6699-4bb1-8217-2a5630d2925e")

	expectedRecord := &model.MedicalRecord{
		ID:        testID,
		RecordNo:  "MR123456",
		PetID:     petID,
		OwnerID:   ownerID,
		VisitDate: time.Now(),
		Status:    *updateReq.Status,
		Treatment: *updateReq.Treatment,
		Notes:     *updateReq.Notes,
	}

	mockService.On("UpdateMedicalRecord", mock.Anything, "550e8400-e29b-41d4-a716-446655440000", updateReq).Return(expectedRecord, nil)

	reqBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/medical-records/550e8400-e29b-41d4-a716-446655440000", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.MedicalRecord
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedRecord.Status, response.Status)

	mockService.AssertExpectations(t)
}

func TestDeleteMedicalRecord(t *testing.T) {
	router, mockService := setupMedicalRecordTestRouter()

	mockService.On("DeleteMedicalRecord", mock.Anything, "550e8400-e29b-41d4-a716-446655440000").Return(nil)

	req, _ := http.NewRequest("DELETE", "/medical-records/550e8400-e29b-41d4-a716-446655440000", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "medical record deleted", response["message"])

	mockService.AssertExpectations(t)
}

func TestGetMedicalRecordsWithPagination(t *testing.T) {
	router, mockService := setupMedicalRecordTestRouter()

	testID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	petID := uuid.MustParse("87a6e7a4-8b70-44b7-b7b5-e0b7e3b4e2bd")
	ownerID := uuid.MustParse("a26c8688-6699-4bb1-8217-2a5630d2925e")

	allRecords := []model.MedicalRecord{
		{
			ID:        testID,
			RecordNo:  "MR123456",
			PetID:     petID,
			OwnerID:   ownerID,
			VisitDate: time.Now(),
			Status:    "確定済",
		},
	}

	expectedResult := &model.PaginatedMedicalRecords{
		Records:     allRecords,
		CurrentPage: 1,
		PerPage:     10,
		Total:       1,
		TotalPages:  1,
		HasNext:     false,
		HasPrev:     false,
	}

	mockService.On("GetAllMedicalRecords", mock.Anything).Return(allRecords, nil)

	req, _ := http.NewRequest("GET", "/medical-records/paginated?page=1&limit=10", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.PaginatedMedicalRecords
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult.CurrentPage, response.CurrentPage)
	assert.Len(t, response.Records, 1)

	mockService.AssertExpectations(t)
}

func TestCreateMedicalRecord_InvalidRequest(t *testing.T) {
	router, _ := setupMedicalRecordTestRouter()

	invalidReq := map[string]interface{}{
		"pet_id": "", // 空のpet_idは無効
	}

	reqBody, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest("POST", "/medical-records", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetMedicalRecordByID_NotFound(t *testing.T) {
	router, mockService := setupMedicalRecordTestRouter()

	mockService.On("GetMedicalRecordByID", mock.Anything, "non-existent-id").Return(nil, apperrors.WrapNotFound("medical record with id %s not found", "non-existent-id"))

	req, _ := http.NewRequest("GET", "/medical-records/non-existent-id", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockService.AssertExpectations(t)
}
