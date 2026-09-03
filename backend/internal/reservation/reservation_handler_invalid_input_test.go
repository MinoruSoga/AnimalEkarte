package reservation

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestListReservations_InvalidDateUsesFixedJapanese(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithReservationSvc(&mockReservationService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/?page=1&limit=10&date=2026/03/24", http.NoBody)
	setClinicID(c)

	h.ListReservations(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), errMsgInvalidDateTime)
	assert.NotContains(t, w.Body.String(), "invalid date format")
	assert.NotContains(t, w.Body.String(), ": invalid input")
}
