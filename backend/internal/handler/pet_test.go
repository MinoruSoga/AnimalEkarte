package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

// MockPetService is a mock implementation of PetService
type MockPetService struct {
	mock.Mock
}

func (m *MockPetService) GetAllPets(ctx context.Context) ([]model.Pet, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Pet), args.Error(1)
}

func (m *MockPetService) GetPetByID(ctx context.Context, id string) (*model.Pet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Pet), args.Error(1)
}

func (m *MockPetService) CreatePet(ctx context.Context, req *model.CreatePetRequest) (*model.Pet, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Pet), args.Error(1)
}

func (m *MockPetService) UpdatePet(ctx context.Context, id string, req *model.UpdatePetRequest) (*model.Pet, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Pet), args.Error(1)
}

func (m *MockPetService) DeletePet(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h.RegisterRoutes(r)
	return r
}

func TestHealth(t *testing.T) {
	mockSvc := new(MockPetService)
	h := New(mockSvc)
	router := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestGetPets_Success(t *testing.T) {
	mockSvc := new(MockPetService)
	h := New(mockSvc)
	router := setupRouter(h)

	expectedPets := []model.Pet{
		{ID: uuid.New(), Name: "ポチ", Species: "犬"},
		{ID: uuid.New(), Name: "タマ", Species: "猫"},
	}

	mockSvc.On("GetAllPets", mock.Anything).Return(expectedPets, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/pets", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []model.Pet
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	mockSvc.AssertExpectations(t)
}

func TestGetPet_Success(t *testing.T) {
	mockSvc := new(MockPetService)
	h := New(mockSvc)
	router := setupRouter(h)

	petID := uuid.New()
	expectedPet := &model.Pet{ID: petID, Name: "ポチ", Species: "犬"}

	mockSvc.On("GetPetByID", mock.Anything, petID.String()).Return(expectedPet, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/pets/"+petID.String(), http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Pet
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ポチ", response.Name)
	mockSvc.AssertExpectations(t)
}

func TestGetPet_NotFound(t *testing.T) {
	mockSvc := new(MockPetService)
	h := New(mockSvc)
	router := setupRouter(h)

	petID := uuid.New()
	mockSvc.On("GetPetByID", mock.Anything, petID.String()).
		Return(nil, apperrors.WrapNotFound("pet", petID.String()))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/pets/"+petID.String(), http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestCreatePet_Success(t *testing.T) {
	mockSvc := new(MockPetService)
	h := New(mockSvc)
	router := setupRouter(h)

	reqBody := model.CreatePetRequest{
		Name:    "ポチ",
		Species: "犬",
		Breed:   "柴犬",
	}
	createdPet := &model.Pet{
		ID:      uuid.New(),
		Name:    "ポチ",
		Species: "犬",
		Breed:   "柴犬",
	}

	mockSvc.On("CreatePet", mock.Anything, mock.AnythingOfType("*model.CreatePetRequest")).
		Return(createdPet, nil)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/pets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response model.Pet
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ポチ", response.Name)
	mockSvc.AssertExpectations(t)
}

func TestCreatePet_BadRequest(t *testing.T) {
	mockSvc := new(MockPetService)
	h := New(mockSvc)
	router := setupRouter(h)

	// Invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/pets", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePet_Success(t *testing.T) {
	mockSvc := new(MockPetService)
	h := New(mockSvc)
	router := setupRouter(h)

	petID := uuid.New()
	reqBody := model.UpdatePetRequest{
		Name: "ポチ太郎",
	}
	updatedPet := &model.Pet{
		ID:      petID,
		Name:    "ポチ太郎",
		Species: "犬",
	}

	mockSvc.On("UpdatePet", mock.Anything, petID.String(), mock.AnythingOfType("*model.UpdatePetRequest")).
		Return(updatedPet, nil)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/v1/pets/"+petID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Pet
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ポチ太郎", response.Name)
	mockSvc.AssertExpectations(t)
}

func TestDeletePet_Success(t *testing.T) {
	mockSvc := new(MockPetService)
	h := New(mockSvc)
	router := setupRouter(h)

	petID := uuid.New()
	mockSvc.On("DeletePet", mock.Anything, petID.String()).Return(nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/pets/"+petID.String(), http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestDeletePet_NotFound(t *testing.T) {
	mockSvc := new(MockPetService)
	h := New(mockSvc)
	router := setupRouter(h)

	petID := uuid.New()
	mockSvc.On("DeletePet", mock.Anything, petID.String()).
		Return(apperrors.WrapNotFound("pet", petID.String()))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/pets/"+petID.String(), http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}
