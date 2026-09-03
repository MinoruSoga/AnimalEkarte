package labdeviceagent

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failingResponseWriter struct {
	header http.Header
}

const testConsumerToken = "test-consumer-token"

func (w *failingResponseWriter) Header() http.Header { return w.header }
func (w *failingResponseWriter) WriteHeader(int)     {}
func (w *failingResponseWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func claimConsumer(t *testing.T, handler http.Handler) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/claim", strings.NewReader(`{"clinic_id":"clinic-2"}`))
	req.Header.Set("Origin", "http://localhost:3003")
	req.Header.Set("X-Lab-Device-Consumer-Token", testConsumerToken)
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
	request.Header.Set("X-Lab-Device-Consumer-Token", testConsumerToken)
}

func TestHandlerRequiresConsumerTokenForProtectedOperations(t *testing.T) {
	queue := NewQueue(1)
	frame, err := queue.Enqueue([]byte{0x02, 0x03}, time.Now())
	require.NoError(t, err)

	tests := []struct {
		name           string
		configured     string
		provided       string
		claimStatus    int
		framesStatus   int
		decisionStatus int
	}{
		{name: "missing token", configured: testConsumerToken, claimStatus: http.StatusUnauthorized, framesStatus: http.StatusUnauthorized, decisionStatus: http.StatusUnauthorized},
		{name: "invalid token", configured: testConsumerToken, provided: "invalid-token", claimStatus: http.StatusUnauthorized, framesStatus: http.StatusUnauthorized, decisionStatus: http.StatusUnauthorized},
		{name: "empty configured token fails closed", configured: "", provided: testConsumerToken, claimStatus: http.StatusUnauthorized, framesStatus: http.StatusUnauthorized, decisionStatus: http.StatusUnauthorized},
		{name: "valid token", configured: testConsumerToken, provided: testConsumerToken, claimStatus: http.StatusOK, framesStatus: http.StatusOK, decisionStatus: http.StatusNoContent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(queue, &Status{}, "clinic-2", test.configured)
			claim := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/claim", strings.NewReader(`{"clinic_id":"clinic-2"}`))
			if test.provided != "" {
				claim.Header.Set("X-Lab-Device-Consumer-Token", test.provided)
			}
			claimResponse := httptest.NewRecorder()
			handler.ServeHTTP(claimResponse, claim)
			assert.Equal(t, test.claimStatus, claimResponse.Code)
			owner := "owner"
			if test.claimStatus == http.StatusOK {
				var claimPayload struct {
					Owner string `json:"owner"`
				}
				require.NoError(t, json.Unmarshal(claimResponse.Body.Bytes(), &claimPayload))
				owner = claimPayload.Owner
			}

			frames := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/frames", http.NoBody)
			frames.Header.Set("X-Clinic-ID", "clinic-2")
			frames.Header.Set("X-Lab-Device-Owner", owner)
			if test.provided != "" {
				frames.Header.Set("X-Lab-Device-Consumer-Token", test.provided)
			}
			framesResponse := httptest.NewRecorder()
			handler.ServeHTTP(framesResponse, frames)
			assert.Equal(t, test.framesStatus, framesResponse.Code)

			decision := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/frames/"+frame.ID+"/ack", http.NoBody)
			decision.Header.Set("X-Clinic-ID", "clinic-2")
			decision.Header.Set("X-Lab-Device-Owner", owner)
			if test.provided != "" {
				decision.Header.Set("X-Lab-Device-Consumer-Token", test.provided)
			}
			decisionResponse := httptest.NewRecorder()
			handler.ServeHTTP(decisionResponse, decision)
			assert.Equal(t, test.decisionStatus, decisionResponse.Code)
		})
	}
}

func TestHandlerHealthDoesNotRequireConsumerToken(t *testing.T) {
	handler := NewHandler(NewQueue(1), &Status{}, "clinic-2", "")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/health", http.NoBody))
	require.Equal(t, http.StatusOK, response.Code)
}

func TestHandlerExposesFramesWithoutLeakingPortIdentity(t *testing.T) {
	queue := NewQueue(4)
	frame, err := queue.Enqueue([]byte{0x02, 0xff, 0x03}, time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	status := &Status{}
	status.SetOpenPorts(2)
	handler := NewHandler(queue, status, "clinic-2", testConsumerToken)
	owner := claimConsumer(t, handler)

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/frames", http.NoBody)
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
	handler := NewHandler(queue, &Status{}, "clinic-2", testConsumerToken)
	owner := claimConsumer(t, handler)

	foreign := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/frames", http.NoBody)
	foreign.Header.Set("Origin", "https://example.invalid")
	foreignRes := httptest.NewRecorder()
	handler.ServeHTTP(foreignRes, foreign)
	require.Equal(t, http.StatusForbidden, foreignRes.Code)

	ack := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/frames/"+frame.ID+"/ack", http.NoBody)
	ack.Header.Set("Origin", "http://127.0.0.1:3003")
	authorizeRequest(ack, owner)
	ackRes := httptest.NewRecorder()
	handler.ServeHTTP(ackRes, ack)
	require.Equal(t, http.StatusNoContent, ackRes.Code)
	require.Empty(t, queue.Snapshot())
}

func TestHandlerRejectsUnexpectedHost(t *testing.T) {
	handler := NewHandler(NewQueue(1), &Status{}, "clinic-2", testConsumerToken)
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/health", http.NoBody)
	req.Host = "attacker.invalid"
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusForbidden, res.Code)
}

func TestHandlerRecordsResponseWriteFailure(t *testing.T) {
	status := &Status{}
	handler := NewHandler(NewQueue(1), status, "clinic-2", testConsumerToken)
	response := &failingResponseWriter{header: make(http.Header)}
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/health", http.NoBody))
	require.Equal(t, uint64(1), status.ResponseErrors())
	require.Equal(t, "response_write_failed", status.LastErrorCategory())
}

func TestHandlerSupportsPrivateNetworkPreflight(t *testing.T) {
	handler := NewHandler(NewQueue(1), &Status{}, "clinic-2", testConsumerToken)
	req := httptest.NewRequest(http.MethodOptions, "http://127.0.0.1:17654/frames", http.NoBody)
	req.Header.Set("Origin", "http://localhost:3003")
	req.Header.Set("Access-Control-Request-Private-Network", "true")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)
	require.Equal(t, "true", res.Header().Get("Access-Control-Allow-Private-Network"))
	require.Contains(t, res.Header().Get("Access-Control-Allow-Headers"), "X-Lab-Device-Consumer-Token")
}

func TestHandlerHealthReportsPartialPortFailureWithoutSensitiveDetails(t *testing.T) {
	status := &Status{}
	status.SetConfiguredPorts(2)
	status.SetOpenPorts(1)
	status.AddOpenError()
	handler := NewHandler(NewQueue(1), status, "clinic-2", testConsumerToken)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/health", http.NoBody))

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
	handler := NewHandler(queue, &Status{}, "clinic-2", testConsumerToken)
	owner := claimConsumer(t, handler)

	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/health", http.NoBody))
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
	rejectRequest := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/frames/"+frame.ID+"/reject", http.NoBody)
	authorizeRequest(rejectRequest, owner)
	handler.ServeHTTP(reject, rejectRequest)
	require.Equal(t, http.StatusNoContent, reject.Code)

	missing := httptest.NewRecorder()
	missingRequest := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/frames/missing/ack", http.NoBody)
	authorizeRequest(missingRequest, owner)
	handler.ServeHTTP(missing, missingRequest)
	require.Equal(t, http.StatusNotFound, missing.Code)

	unknown := httptest.NewRecorder()
	handler.ServeHTTP(unknown, httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/unknown", http.NoBody))
	require.Equal(t, http.StatusNotFound, unknown.Code)
}

func TestHandlerBindsClinicAndAllowsOnlyOneConsumer(t *testing.T) {
	handler := NewHandler(NewQueue(1), &Status{}, "clinic-2", testConsumerToken)
	owner := claimConsumer(t, handler)
	require.NotEmpty(t, owner)

	second := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/claim", strings.NewReader(`{"clinic_id":"clinic-2"}`))
	second.Header.Set("X-Lab-Device-Consumer-Token", testConsumerToken)
	secondRes := httptest.NewRecorder()
	handler.ServeHTTP(secondRes, second)
	require.Equal(t, http.StatusConflict, secondRes.Code)

	renew := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/claim", strings.NewReader(`{"clinic_id":"clinic-2"}`))
	renew.Header.Set("X-Lab-Device-Consumer-Token", testConsumerToken)
	renew.Header.Set("X-Lab-Device-Owner", owner)
	renewRes := httptest.NewRecorder()
	handler.ServeHTTP(renewRes, renew)
	require.Equal(t, http.StatusOK, renewRes.Code)

	wrongClinic := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:17654/claim", strings.NewReader(`{"clinic_id":"clinic-3"}`))
	wrongClinic.Header.Set("X-Lab-Device-Consumer-Token", testConsumerToken)
	wrongClinicRes := httptest.NewRecorder()
	handler.ServeHTTP(wrongClinicRes, wrongClinic)
	require.Equal(t, http.StatusForbidden, wrongClinicRes.Code)

	unauthorized := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:17654/frames", http.NoBody)
	unauthorized.Header.Set("X-Lab-Device-Consumer-Token", testConsumerToken)
	unauthorizedRes := httptest.NewRecorder()
	handler.ServeHTTP(unauthorizedRes, unauthorized)
	require.Equal(t, http.StatusConflict, unauthorizedRes.Code)
}

func TestHandlerAllowsConfiguredDeployedOriginAndPNA(t *testing.T) {
	handler := NewHandler(NewQueue(1), &Status{}, "clinic-2", testConsumerToken, "https://staging.example.test")
	req := httptest.NewRequest(http.MethodOptions, "http://127.0.0.1:17654/frames", http.NoBody)
	req.Header.Set("Origin", "https://staging.example.test")
	req.Header.Set("Access-Control-Request-Private-Network", "true")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)
	require.Equal(t, "https://staging.example.test", res.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "true", res.Header().Get("Access-Control-Allow-Private-Network"))
}

func TestNormalizeAllowedOriginRejectsUnsafeValues(t *testing.T) {
	for _, raw := range []string{"*", "https://*.example.test", "https://example.test/path", "https://user@example.test", "https://example.test:bad", "https://example.test:", "https://example.test:70000", "file:///tmp/x", "http://example.test", `https://example.test\evil`, "https://127.1", "https://0x7f000001", "https://0x7f.0.0.1", "https://0x7f.1", "https://2130706433", "https://0177.0.0.1", "https://127.0.0.1.", "https://[::ffff:192.0.2.128]", "https://[::ffff:c000:280]"} {
		_, ok := NormalizeAllowedOrigin(raw)
		require.False(t, ok, raw)
	}
}

func TestNormalizeAllowedOriginReturnsCanonicalBrowserOrigin(t *testing.T) {
	tests := map[string]string{
		"https://EXAMPLE.test":           "https://example.test",
		"https://Example.test:443":       "https://example.test",
		"https://Example.test:0443":      "https://example.test",
		"https://Example.test:8443":      "https://example.test:8443",
		"https://[2001:0DB8:0:0::1]:443": "https://[2001:db8::1]",
		"https://[2001:DB8::1]:8443":     "https://[2001:db8::1]:8443",
		"http://LOCALHOST:80":            "http://localhost",
		"http://LOCALHOST:3003":          "http://localhost:3003",
		"http://127.0.0.1:80":            "http://127.0.0.1",
		"https://127.0.0.1":              "https://127.0.0.1",
		"https://service.123.example":    "https://service.123.example",
	}
	for raw, expected := range tests {
		t.Run(raw, func(t *testing.T) {
			actual, ok := NormalizeAllowedOrigin(raw)
			require.True(t, ok)
			require.Equal(t, expected, actual)
		})
	}
}

func TestHandlerRejectsBrowserCanonicalOriginWhenRewrittenConfigurationIsInvalid(t *testing.T) {
	tests := map[string]string{
		"https://0x7f000001":           "https://127.0.0.1",
		"https://0x7f.0.0.1":           "https://127.0.0.1",
		"https://0x7f.1":               "https://127.0.0.1",
		"https://2130706433":           "https://127.0.0.1",
		"https://127.1":                "https://127.0.0.1",
		"https://0177.0.0.1":           "https://127.0.0.1",
		"https://127.0.0.1.":           "https://127.0.0.1",
		"https://[::ffff:192.0.2.128]": "https://[::ffff:c000:280]",
		"https://[::ffff:c000:280]":    "https://[::ffff:c000:280]",
	}
	for configured, browserOrigin := range tests {
		t.Run(configured, func(t *testing.T) {
			handler := NewHandler(NewQueue(1), &Status{}, "clinic-2", testConsumerToken, configured)
			request := httptest.NewRequest(http.MethodOptions, "http://127.0.0.1:17654/frames", http.NoBody)
			request.Header.Set("Origin", browserOrigin)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			require.Equal(t, http.StatusForbidden, response.Code)
			require.Empty(t, response.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestHandlerAcceptsOnlyCanonicalRequestOriginForCanonicalizedConfiguration(t *testing.T) {
	handler := NewHandler(NewQueue(1), &Status{}, "clinic-2", testConsumerToken, "https://EXAMPLE.test:0443")

	canonical := httptest.NewRequest(http.MethodOptions, "http://127.0.0.1:17654/frames", http.NoBody)
	canonical.Header.Set("Origin", "https://example.test")
	canonicalResponse := httptest.NewRecorder()
	handler.ServeHTTP(canonicalResponse, canonical)
	require.Equal(t, http.StatusNoContent, canonicalResponse.Code)
	require.Equal(t, "https://example.test", canonicalResponse.Header().Get("Access-Control-Allow-Origin"))

	for _, origin := range []string{"https://EXAMPLE.test", "https://example.test:443", "https://attacker.example.test"} {
		request := httptest.NewRequest(http.MethodOptions, "http://127.0.0.1:17654/frames", http.NoBody)
		request.Header.Set("Origin", origin)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		require.Equal(t, http.StatusForbidden, response.Code, origin)
		require.Empty(t, response.Header().Get("Access-Control-Allow-Origin"), origin)
	}
}

type originParityCorpus struct {
	Cases []struct {
		Raw       string `json:"raw"`
		Canonical string `json:"canonical"`
	} `json:"cases"`
}

func loadOriginParityCorpus(t *testing.T) originParityCorpus {
	t.Helper()
	content, err := os.ReadFile("testdata/origin_parity.json")
	require.NoError(t, err)
	var corpus originParityCorpus
	require.NoError(t, json.Unmarshal(content, &corpus))
	require.NotEmpty(t, corpus.Cases)
	return corpus
}

func TestNormalizeAllowedOriginMatchesSharedBrowserParityCorpus(t *testing.T) {
	for _, test := range loadOriginParityCorpus(t).Cases {
		t.Run(test.Raw, func(t *testing.T) {
			actual, ok := NormalizeAllowedOrigin(test.Raw)
			if test.Canonical == "" {
				require.False(t, ok)
				require.Empty(t, actual)
				return
			}
			require.True(t, ok)
			require.Equal(t, test.Canonical, actual)
		})
	}
}

func TestHandlerExactCORSMatchesSharedOriginContract(t *testing.T) {
	for _, test := range loadOriginParityCorpus(t).Cases {
		t.Run(test.Raw, func(t *testing.T) {
			handler := NewHandler(NewQueue(1), &Status{}, "clinic-2", testConsumerToken, test.Raw)
			request := httptest.NewRequest(http.MethodOptions, "http://127.0.0.1:17654/frames", http.NoBody)
			requestOrigin := test.Canonical
			expectedStatus := http.StatusNoContent
			if requestOrigin == "" {
				requestOrigin = test.Raw
				expectedStatus = http.StatusForbidden
			}
			request.Header.Set("Origin", requestOrigin)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			require.Equal(t, expectedStatus, response.Code)
			require.Equal(t, test.Canonical, response.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}
