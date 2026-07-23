package lstep

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
