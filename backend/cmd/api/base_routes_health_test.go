package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type pingStubConnector struct {
	err error
}

func (c pingStubConnector) Connect(context.Context) (driver.Conn, error) {
	if c.err != nil {
		return nil, c.err
	}
	return pingStubConn{}, nil
}

func (pingStubConnector) Driver() driver.Driver { return pingStubDriver{} }

type pingStubDriver struct{}

func (pingStubDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("pingStubDriver.Open must not be used")
}

type pingStubConn struct{}

func (pingStubConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unsupported")
}
func (pingStubConn) Close() error { return nil }
func (pingStubConn) Begin() (driver.Tx, error) {
	return nil, errors.New("unsupported")
}
func (pingStubConn) Ping(context.Context) error { return nil }

func gormWithPingablePool(t *testing.T, connErr error) *gorm.DB {
	t.Helper()
	sqlDB := sql.OpenDB(pingStubConnector{err: connErr})
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close test sql.DB: %v", err)
		}
	})
	return &gorm.DB{Config: &gorm.Config{ConnPool: sqlDB}}
}

func TestHealthDB_OK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/db", http.NoBody)

	healthDB(gormWithPingablePool(t, nil))(c)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), `"status":"ok"`)
}

func TestHealthDB_UnavailableWhenDBNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/db", http.NoBody)

	healthDB(nil)(c)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "error")
}

func TestHealthDB_UnavailableWhenPingFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodGet, "/health/db", http.NoBody)

	healthDB(gormWithPingablePool(t, errors.New("connect refused")))(c)

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.NotContains(t, recorder.Body.String(), "refused")
}
