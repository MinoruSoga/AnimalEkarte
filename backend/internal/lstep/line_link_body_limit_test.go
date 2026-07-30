package lstep

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/animal-ekarte/backend/internal/model"
)

type webhookReadTracker struct {
	reads int
}

func (r *webhookReadTracker) Read([]byte) (int, error) {
	r.reads++
	return 0, errors.New("request body must not be read")
}

func (*webhookReadTracker) Close() error {
	return nil
}

var _ io.ReadCloser = (*webhookReadTracker)(nil)

func TestReceiveLineWebhook_MissingSignatureDoesNotReadBody(t *testing.T) {
	svc := &mockLineLinkService{
		handleWebhookFn: func(context.Context, []byte, string) error {
			t.Fatal("service must not run without a signature")
			return nil
		},
	}
	router := newPostReceiveLineWebhookRouter(svc)
	tracker := &webhookReadTracker{}
	request := httptest.NewRequest(http.MethodPost, "/line/webhook", http.NoBody)
	request.Body = tracker
	request.ContentLength = -1
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	assert.Zero(t, tracker.reads)
}

func TestReceiveLineWebhook_RejectsChunkedOversizedBodyBeforeService(t *testing.T) {
	serviceCalled := false
	svc := &mockLineLinkService{
		handleWebhookFn: func(context.Context, []byte, string) error {
			serviceCalled = true
			return nil
		},
	}
	router := newPostReceiveLineWebhookRouter(svc)
	body := bytes.NewReader(make([]byte, maxLineWebhookRequestBytes+1))
	request := httptest.NewRequest(http.MethodPost, "/line/webhook", body)
	request.ContentLength = -1
	request.Header.Set("X-Line-Signature", "signed")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	assert.False(t, serviceCalled)
}

// TestMaxLineWebhookRequestBytes_IsBoundedForAmplificationControl documents the
// SEC-CS-F05 body cap. follow/unfollow payloads are small; 2MiB would let an
// invalid webhook amplify per-clinic HMAC work.
func TestMaxLineWebhookRequestBytes_IsBoundedForAmplificationControl(t *testing.T) {
	assert.Equal(t, int64(256*1024), maxLineWebhookRequestBytes)
}

func TestReceiveLineWebhook_RejectsContentLengthOverLimitBeforeService(t *testing.T) {
	serviceCalled := false
	svc := &mockLineLinkService{
		handleWebhookFn: func(context.Context, []byte, string) error {
			serviceCalled = true
			return nil
		},
	}
	router := newPostReceiveLineWebhookRouter(svc)
	// ContentLength over the cap must fail closed before body read / service.
	request := httptest.NewRequest(http.MethodPost, "/line/webhook", strings.NewReader("{}"))
	request.ContentLength = maxLineWebhookRequestBytes + 1
	request.Header.Set("X-Line-Signature", "signed")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	assert.False(t, serviceCalled)
}

func TestLinkLiffAccount_RejectsChunkedOversizedBodyBeforeService(t *testing.T) {
	serviceCalled := false
	svc := &mockLineLinkService{
		linkAccountFn: func(context.Context, uint64, LinkAccountInput) (*model.Owner, error) {
			serviceCalled = true
			return &model.Owner{ID: 1}, nil
		},
	}
	router := newPostLiffLinkAccountRouter(svc)
	body := strings.NewReader(
		`{"link_token":"` +
			strings.Repeat("x", int(maxLineLinkRequestBytes)) +
			`","line_id_token":"token"}`,
	)
	request := httptest.NewRequest(http.MethodPost, "/liff/1/link", body)
	request.ContentLength = -1
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	assert.False(t, serviceCalled)
}
