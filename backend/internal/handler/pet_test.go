package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/animal-ekarte/backend/internal/model"
)

func TestGetPets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockService) // Use MockService instead of MockPetService
	h := New(mockSvc)

	r := gin.New()
	r.GET("/pets", h.GetPets)

	expectedPets := []model.Pet{
		{
			ID:   uuid.New(),
			Name: "Pochi",
		},
	}

	mockSvc.On("GetAllPets", mock.Anything).Return(expectedPets, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/pets", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockService)
	h := New(mockSvc)

	r := gin.New()
	r.GET("/pets/:id", h.GetPet)

	id := uuid.New()
	expectedPet := &model.Pet{
		ID:   id,
		Name: "Pochi",
	}

	mockSvc.On("GetPetByID", mock.Anything, id.String()).Return(expectedPet, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/pets/"+id.String(), http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreatePet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockService)
	h := New(mockSvc)

	r := gin.New()
	r.POST("/pets", h.CreatePet)

	reqBody := model.CreatePetRequest{
		Name:    "Pochi",
		Species: "Dog",
		OwnerID: uuid.New().String(),
	}
	expectedPet := &model.Pet{
		ID:      uuid.New(),
		Name:    "Pochi",
		Species: "Dog",
	}

	// Mock expects a pointer to request
	mockSvc.On("CreatePet", mock.Anything, &reqBody).Return(expectedPet, nil)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/pets", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
