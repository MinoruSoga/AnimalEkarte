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

func TestGetAllOwners(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockService)
	h := New(mockSvc)

	r := gin.New()
	r.GET("/owners", h.GetAllOwners)

	expectedOwners := []model.Owner{
		{
			ID:   uuid.New(),
			Name: "Test Owner",
		},
	}

	mockSvc.On("GetAllOwners", mock.Anything).Return(expectedOwners, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/owners", http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []model.Owner
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedOwners), len(response))
	assert.Equal(t, expectedOwners[0].ID, response[0].ID)
}

func TestGetOwnerByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockService)
	h := New(mockSvc)

	r := gin.New()
	r.GET("/owners/:id", h.GetOwnerByID)

	id := uuid.New()
	expectedOwner := &model.Owner{
		ID:   id,
		Name: "Test Owner",
	}

	mockSvc.On("GetOwnerByID", mock.Anything, id.String()).Return(expectedOwner, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/owners/"+id.String(), http.NoBody)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response model.Owner
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedOwner.ID, response.ID)
}

func TestCreateOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockService)
	h := New(mockSvc)

	r := gin.New()
	r.POST("/owners", h.CreateOwner)

	reqBody := model.CreateOwnerRequest{
		Name: "New Owner",
	}
	expectedOwner := &model.Owner{
		ID:   uuid.New(),
		Name: "New Owner",
	}

	mockSvc.On("CreateOwner", mock.Anything, &reqBody).Return(expectedOwner, nil)

	body, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/owners", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
