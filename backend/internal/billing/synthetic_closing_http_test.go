package billing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterUATRoutes_StagingIsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &SyntheticClosingHandler{AppEnv: "staging", DBHost: "db", Password: "unused"}
	r := gin.New()
	RegisterUATRoutes(r.Group("/api/v1"), h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/uat/synthetic-closings", bytes.NewBufferString(`{"targetDate":"2026-09-07"}`))
	req.Host = "localhost"
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRegisterUATRoutes_ForeignHostIsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &SyntheticClosingHandler{AppEnv: "development", DBHost: "db", Password: "unused"}
	r := gin.New()
	RegisterUATRoutes(r.Group("/api/v1"), h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/uat/synthetic-closings", bytes.NewBufferString(`{"targetDate":"2026-09-07"}`))
	req.Host = "aws.connect.psdb.cloud"
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRegisterUATRoutes_CreateAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testdbSetupSyntheticClosing(t)
	h := &SyntheticClosingHandler{
		DB:       db,
		AppEnv:   "development",
		DBHost:   "db",
		Password: "s09-local-password",
	}
	r := gin.New()
	RegisterUATRoutes(r.Group("/api/v1"), h)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/uat/synthetic-closings", bytes.NewBufferString(`{"targetDate":"2026-09-07"}`))
	req.Host = "localhost:8080"
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	var body syntheticClosingHTTPResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.NotZero(t, body.ClinicID)
	require.Equal(t, SyntheticClosingLoginEmail(body.ClinicID), body.LoginEmail)
	require.Len(t, body.BillingIDs, 5)
	require.Len(t, body.CompletedAt, 5)
	require.Equal(t, SyntheticClosingCleanupToken(body.ClinicID), body.CleanupToken)

	del := httptest.NewRequest(http.MethodDelete, "/api/v1/uat/synthetic-closings/"+strconv.FormatUint(body.ClinicID, 10), nil)
	del.Host = "127.0.0.1"
	del.Header.Set(syntheticClosingCleanupHeader, body.CleanupToken)
	dw := httptest.NewRecorder()
	r.ServeHTTP(dw, del)
	assert.Equal(t, http.StatusNoContent, dw.Code)
}
