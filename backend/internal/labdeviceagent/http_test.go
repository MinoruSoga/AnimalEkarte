package labdeviceagent

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type failingResponseWriter struct {
	header http.Header
}

func (w *failingResponseWriter) Header() http.Header { return w.header }
func (w *failingResponseWriter) WriteHeader(int)     {}
func (w *failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func claimConsumer(t *testing.T, handler http.Handler) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/claim", strings.NewReader(`{"clinic_id":"clinic-2"}`))
	req.Header.Set("Origin", "http://localhost:3003")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	var payload struct {
		Owner string `json:"owner"`
	}
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.NotEmpty(t, payload.Owner)
	return payload.Owner
}

func authorizeRequest(request *http.Request, owner string) {
	request.Header.Set("X-Clinic-ID", "clinic-2")
	request.Header.Set("X-Lab-Device-Owner", owner)
}

func TestHandlerExposesFramesWithoutLeakingPortIdentity(t *testing.T) {
	queue := NewQueue(4)
	frame, err := queue.Enqueue([]byte{0x02, 0xff, 0x03}, time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	status := &Status{}
	status.SetOpenPorts(2)
	handler := NewHandler(queue, status, "clinic-2")
	owner := claimConsumer(t, handler)

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/frames", nil)
	req.Header.Set("Origin", "http://localhost:3003")
	authorizeRequest(req, owner)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "http://localhost:3003", res.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "no-store", res.Header().Get("Cache-Control"))
	var payload struct {
		Frames []struct {
			ID            string `json:"id"`
			PayloadBase64 string `json:"payload_base64"`
		} `json:"frames"`
	}
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &payload))
	require.Equal(t, frame.ID, payload.Frames[0].ID)
	require.Equal(t, "Av8D", payload.Frames[0].PayloadBase64)
	require.NotContains(t, res.Body.String(), "usbserial")
}

func TestHandlerRejectsForeignOriginsAndAcknowledgesExactFrame(t *testing.T) {
	queue := NewQueue(2)
	frame, err := queue.Enqueue([]byte{0x02, 0x03}, time.Now())
	require.NoError(t, err)
	handler := NewHandler(queue, &Status{}, "clinic-2")
	owner := claimConsumer(t, handler)

	foreign := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/frames", nil)
	foreign.Header.Set("Origin", "https://example.invalid")
	foreignRes := httptest.NewRecorder()
	handler.ServeHTTP(foreignRes, foreign)
	require.Equal(t, http.StatusForbidden, foreignRes.Code)

	ack := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/frames/"+frame.ID+"/ack", nil)
	ack.Header.Set("Origin", "http://127.0.0.1:3003")
	authorizeRequest(ack, owner)
	ackRes := httptest.NewRecorder()
	handler.ServeHTTP(ackRes, ack)
	require.Equal(t, http.StatusNoContent, ackRes.Code)
	require.Empty(t, queue.Snapshot())
}

func TestHandlerRejectsUnexpectedHost(t *testing.T) {
	handler := NewHandler(NewQueue(1), &Status{}, "clinic-2")
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/health", nil)
	req.Host = "attacker.invalid"
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestHandlerRecordsResponseWriteFailure(t *testing.T) {
	status := &Status{}
	handler := NewHandler(NewQueue(1), status, "clinic-2")
	response := &failingResponseWriter{header: make(http.Header)}
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/health", nil))
	require.Equal(t, uint64(1), status.ResponseErrors())
	require.Equal(t, "response_write_failed", status.LastErrorCategory())
}

func TestHandlerSupportsPrivateNetworkPreflight(t *testing.T) {
	handler := NewHandler(NewQueue(1), &Status{}, "clinic-2")
	req := httptest.NewRequest(http.MethodOptions, "http://127.0.0.1:17654/frames", nil)
	req.Header.Set("Origin", "http://localhost:3003")
	req.Header.Set("Access-Control-Request-Private-Network", "true")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)
	require.Equal(t, "true", res.Header().Get("Access-Control-Allow-Private-Network"))
}

func TestHandlerHealthReportsPartialPortFailureWithoutSensitiveDetails(t *testing.T) {
	status := &Status{}
	status.SetConfiguredPorts(2)
	status.SetOpenPorts(1)
	status.AddOpenError()
	handler := NewHandler(NewQueue(1), status, "clinic-2")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/health", nil))

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), `"status":"degraded"`)
	require.Contains(t, res.Body.String(), `"configured_ports":2`)
	require.Contains(t, res.Body.String(), `"open_ports":1`)
	require.Contains(t, res.Body.String(), `"last_error_category":"port_open_failed"`)
	require.NotContains(t, res.Body.String(), "usbserial")
	require.NotContains(t, res.Body.String(), "permission denied")
}

func TestHandlerHealthRejectAndUnknownRoutes(t *testing.T) {
	queue := NewQueue(1)
	frame, err := queue.Enqueue([]byte{0x02, 0x03}, time.Now())
	require.NoError(t, err)
	handler := NewHandler(queue, &Status{}, "clinic-2")
	owner := claimConsumer(t, handler)

	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/health", nil))
	require.Equal(t, http.StatusOK, health.Code)
	require.Contains(t, health.Body.String(), `"pending":1`)
	require.Contains(t, health.Body.String(), `"configured_ports":0`)
	require.Contains(t, health.Body.String(), `"port_discovery_failures_total":0`)
	require.Contains(t, health.Body.String(), `"port_open_failures_total":0`)
	require.Contains(t, health.Body.String(), `"queue_failures_total":0`)
	require.Contains(t, health.Body.String(), `"port_close_failures_total":0`)
	require.Contains(t, health.Body.String(), `"response_failures_total":0`)
	require.Contains(t, health.Body.String(), `"last_error_category":"none"`)

	reject := httptest.NewRecorder()
	rejectRequest := httptest.NewRequest(
		http.MethodPost, "http://127.0.0.1:17654/frames/"+frame.ID+"/reject", nil,
	)
	authorizeRequest(rejectRequest, owner)
	handler.ServeHTTP(reject, rejectRequest)
	require.Equal(t, http.StatusNoContent, reject.Code)

	missing := httptest.NewRecorder()
	missingRequest := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/frames/missing/ack", nil)
	authorizeRequest(missingRequest, owner)
	handler.ServeHTTP(missing, missingRequest)
	require.Equal(t, http.StatusNotFound, missing.Code)

	unknown := httptest.NewRecorder()
	handler.ServeHTTP(unknown, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/unknown", nil))
	require.Equal(t, http.StatusNotFound, unknown.Code)
}

func TestHandlerBindsClinicAndAllowsOnlyOneConsumer(t *testing.T) {
	handler := NewHandler(NewQueue(1), &Status{}, "clinic-2")
	owner := claimConsumer(t, handler)
	require.NotEmpty(t, owner)

	second := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/claim", strings.NewReader(`{"clinic_id":"clinic-2"}`))
	secondRes := httptest.NewRecorder()
	handler.ServeHTTP(secondRes, second)
	require.Equal(t, http.StatusConflict, secondRes.Code)

	renew := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/claim", strings.NewReader(`{"clinic_id":"clinic-2"}`))
	renew.Header.Set("X-Lab-Device-Owner", owner)
	renewRes := httptest.NewRecorder()
	handler.ServeHTTP(renewRes, renew)
	require.Equal(t, http.StatusOK, renewRes.Code)

	wrongClinic := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/claim", strings.NewReader(`{"clinic_id":"clinic-3"}`))
	wrongClinicRes := httptest.NewRecorder()
	handler.ServeHTTP(wrongClinicRes, wrongClinic)
	require.Equal(t, http.StatusForbidden, wrongClinicRes.Code)

	unauthorized := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/frames", nil)
	unauthorizedRes := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedRes, unauthorized)
	require.Equal(t, http.StatusConflict, unauthorizedRes.Code)
}
