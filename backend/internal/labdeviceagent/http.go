package labdeviceagent

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const ListenAddress = "127.0.0.1:17654"

type Status struct {
	openPorts       atomic.Int64
	configuredPorts atomic.Int64
	inputOverflow   atomic.Uint64
	discoveryErrors atomic.Uint64
	openErrors      atomic.Uint64
	queueErrors     atomic.Uint64
	closeErrors     atomic.Uint64
	responseErrors  atomic.Uint64
	lastError       atomic.Value
}

func (s *Status) SetConfiguredPorts(count int) {
	s.configuredPorts.Store(int64(count))
}

func (s *Status) ConfiguredPorts() int {
	return int(s.configuredPorts.Load())
}

func (s *Status) SetOpenPorts(count int) {
	s.openPorts.Store(int64(count))
}

func (s *Status) OpenPorts() int {
	return int(s.openPorts.Load())
}

func (s *Status) AddOpenPorts(delta int) {
	s.openPorts.Add(int64(delta))
}

func (s *Status) AddInputOverflow() {
	s.inputOverflow.Add(1)
}

func (s *Status) InputOverflow() uint64 {
	return s.inputOverflow.Load()
}

func (s *Status) AddDiscoveryError() {
	s.discoveryErrors.Add(1)
	s.lastError.Store("discovery_failed")
}

func (s *Status) DiscoveryErrors() uint64 {
	return s.discoveryErrors.Load()
}

func (s *Status) AddOpenError() {
	s.openErrors.Add(1)
	s.lastError.Store("port_open_failed")
}

func (s *Status) OpenErrors() uint64 {
	return s.openErrors.Load()
}

func (s *Status) AddQueueError() {
	s.queueErrors.Add(1)
	s.lastError.Store("queue_write_failed")
}

func (s *Status) QueueErrors() uint64 {
	return s.queueErrors.Load()
}

func (s *Status) AddCloseError() {
	s.closeErrors.Add(1)
	s.lastError.Store("port_close_failed")
}

func (s *Status) CloseErrors() uint64 {
	return s.closeErrors.Load()
}

func (s *Status) AddResponseError() {
	s.responseErrors.Add(1)
	s.lastError.Store("response_write_failed")
}

func (s *Status) ResponseErrors() uint64 {
	return s.responseErrors.Load()
}

func (s *Status) LastErrorCategory() string {
	value := s.lastError.Load()
	if category, ok := value.(string); ok {
		return category
	}
	return "none"
}

type consumerLease struct {
	mu             sync.Mutex
	expectedClinic string
	owner          string
	expiresAt      time.Time
}

const consumerLeaseDuration = 15 * time.Second

type handler struct {
	queue  *Queue
	status *Status
	lease  *consumerLease
	allowedOrigins map[string]struct{}
}

func NewHandler(queue *Queue, status *Status, expectedClinic string, configuredOrigins ...string) http.Handler {
	origins := map[string]struct{}{
		"http://localhost:3003": {},
		"http://127.0.0.1:3003": {},
	}
	for _, origin := range configuredOrigins {
		if normalized, ok := NormalizeAllowedOrigin(origin); ok { origins[normalized] = struct{}{} }
	}
	return &handler{queue: queue, status: status, lease: &consumerLease{expectedClinic: expectedClinic}, allowedOrigins: origins}
}

// NormalizeAllowedOrigin accepts exact HTTP(S) origins only. Paths, credentials,
// wildcards, and query/fragment components are rejected.
func NormalizeAllowedOrigin(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || strings.Contains(parsed.Host, "*") {
		return "", false
	}
	if parsed.Scheme == "http" && parsed.Hostname() != "localhost" && parsed.Hostname() != "127.0.0.1" {
		return "", false
	}
	return parsed.Scheme + "://" + parsed.Host, true
}

func (h *handler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	if request.Host != "127.0.0.1:17654" && request.Host != "localhost:17654" {
		http.Error(response, "forbidden", http.StatusForbidden)
		return
	}
	if !h.allowBrowserRequest(response, request) {
		http.Error(response, "forbidden", http.StatusForbidden)
		return
	}
	if request.Method == http.MethodOptions {
		response.WriteHeader(http.StatusNoContent)
		return
	}
	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/health":
		h.writeHealth(response)
	case request.Method == http.MethodPost && request.URL.Path == "/claim":
		h.claim(response, request)
	case request.Method == http.MethodGet && request.URL.Path == "/frames":
		if !h.authorizeConsumer(response, request) {
			return
		}
		h.writeFrames(response)
	case request.Method == http.MethodPost && strings.HasPrefix(request.URL.Path, "/frames/"):
		if !h.authorizeConsumer(response, request) {
			return
		}
		h.decideFrame(response, request.URL.Path)
	default:
		http.NotFound(response, request)
	}
}

func (h *handler) allowBrowserRequest(response http.ResponseWriter, request *http.Request) bool {
	origin := request.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if _, allowed := h.allowedOrigins[origin]; !allowed {
		return false
	}
	response.Header().Set("Access-Control-Allow-Origin", origin)
	response.Header().Set("Vary", "Origin")
	response.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	response.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Clinic-ID, X-Lab-Device-Owner")
	if request.Header.Get("Access-Control-Request-Private-Network") == "true" {
		response.Header().Set("Access-Control-Allow-Private-Network", "true")
	}
	return true
}

func (h *handler) writeHealth(response http.ResponseWriter) {
	stats := h.queue.Stats()
	state := "running"
	if stats.Overflow > 0 || stats.Rejected > 0 || h.status.InputOverflow() > 0 || h.status.QueueErrors() > 0 ||
		h.status.CloseErrors() > 0 || h.status.ResponseErrors() > 0 ||
		h.status.OpenPorts() < h.status.ConfiguredPorts() ||
		(h.status.OpenPorts() == 0 && (h.status.DiscoveryErrors() > 0 || h.status.OpenErrors() > 0)) {
		state = "degraded"
	}
	h.writeJSON(response, http.StatusOK, map[string]any{
		"status":                        state,
		"open_ports":                    h.status.OpenPorts(),
		"configured_ports":              h.status.ConfiguredPorts(),
		"input_overflow":                h.status.InputOverflow(),
		"port_discovery_failures_total": h.status.DiscoveryErrors(),
		"port_open_failures_total":      h.status.OpenErrors(),
		"queue_failures_total":          h.status.QueueErrors(),
		"port_close_failures_total":     h.status.CloseErrors(),
		"response_failures_total":       h.status.ResponseErrors(),
		"last_error_category":           h.status.LastErrorCategory(),
		"queue":                         stats,
	})
}

func (h *handler) claim(response http.ResponseWriter, request *http.Request) {
	var input struct {
		ClinicID string `json:"clinic_id"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(response, request.Body, 1024))
	if err := decoder.Decode(&input); err != nil || input.ClinicID == "" || input.ClinicID != h.lease.expectedClinic {
		http.Error(response, "clinic mismatch", http.StatusForbidden)
		return
	}
	h.lease.mu.Lock()
	defer h.lease.mu.Unlock()
	now := time.Now()
	if h.lease.owner != "" && now.Before(h.lease.expiresAt) {
		if request.Header.Get("X-Lab-Device-Owner") != h.lease.owner {
			http.Error(response, "consumer already active", http.StatusConflict)
			return
		}
		h.lease.expiresAt = now.Add(consumerLeaseDuration)
		h.writeJSON(response, http.StatusOK, map[string]any{
			"owner":         h.lease.owner,
			"lease_seconds": int(consumerLeaseDuration.Seconds()),
		})
		return
	}
	random := make([]byte, 24)
	if _, err := rand.Read(random); err != nil {
		http.Error(response, "claim unavailable", http.StatusInternalServerError)
		return
	}
	h.lease.owner = base64.RawURLEncoding.EncodeToString(random)
	h.lease.expiresAt = now.Add(consumerLeaseDuration)
	h.writeJSON(response, http.StatusOK, map[string]any{
		"owner":         h.lease.owner,
		"lease_seconds": int(consumerLeaseDuration.Seconds()),
	})
}

func (h *handler) authorizeConsumer(response http.ResponseWriter, request *http.Request) bool {
	clinicID := request.Header.Get("X-Clinic-ID")
	owner := request.Header.Get("X-Lab-Device-Owner")
	h.lease.mu.Lock()
	defer h.lease.mu.Unlock()
	if clinicID != h.lease.expectedClinic || owner == "" || owner != h.lease.owner || time.Now().After(h.lease.expiresAt) {
		http.Error(response, "consumer lease required", http.StatusConflict)
		return false
	}
	h.lease.expiresAt = time.Now().Add(consumerLeaseDuration)
	return true
}

func (h *handler) writeFrames(response http.ResponseWriter) {
	type frameResponse struct {
		ID            string `json:"id"`
		PayloadBase64 string `json:"payload_base64"`
		ReceivedAt    string `json:"received_at"`
	}
	frames := h.queue.Snapshot()
	items := make([]frameResponse, 0, len(frames))
	for _, frame := range frames {
		items = append(items, frameResponse{
			ID:            frame.ID,
			PayloadBase64: base64.StdEncoding.EncodeToString(frame.Raw),
			ReceivedAt:    frame.ReceivedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
		})
	}
	h.writeJSON(response, http.StatusOK, map[string]any{"frames": items})
}

func (h *handler) decideFrame(response http.ResponseWriter, path string) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "frames" || parts[1] == "" {
		http.NotFound(response, nil)
		return
	}
	var err error
	switch parts[2] {
	case "ack":
		err = h.queue.Ack(parts[1])
	case "reject":
		err = h.queue.Reject(parts[1])
	default:
		http.NotFound(response, nil)
		return
	}
	if errors.Is(err, ErrFrameNotFound) {
		http.Error(response, "frame not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(response, "frame decision failed", http.StatusInternalServerError)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (h *handler) writeJSON(response http.ResponseWriter, status int, value any) {
	if err := writeJSON(response, status, value); err != nil {
		h.status.AddResponseError()
	}
}

func writeJSON(response http.ResponseWriter, status int, value any) error {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(value); err != nil {
		http.Error(response, "response encoding failed", http.StatusInternalServerError)
		return err
	}
	response.Header().Set("Content-Type", "application/json")
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(status)
	_, err := response.Write(body.Bytes())
	return err
}
