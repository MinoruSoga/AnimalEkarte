package medicalrecord

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/apperrors"
)

func TestRespondToServiceInputError_PassesAppErrorThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	respondToServiceInputError(c, apperrors.WrapInvalidInput("cage_id is required"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"cage_id is required"}`, w.Body.String())
}

func TestRespondToServiceInputError_RewritesPlainTimeParseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", nil)

	respondToServiceInputError(c, fmt.Errorf("invalid time format, expected HH:MM:SS"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"日時の形式が正しくありません"}`, w.Body.String())
}
