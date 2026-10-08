package trimming

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/imageupload"
	"github.com/animal-ekarte/backend/internal/model"
)

// ---- test doubles ----

type uploadCall struct {
	key         string
	contentType string
	bodyBytes   int64
}

type signCall struct {
	key string
	ttl time.Duration
}

type fakeTrimmingImageUploader struct {
	uploads      []uploadCall
	deletes      []string
	signCalls    []signCall
	uploadErr    error
	deleteErr    error
	signErr      error
	signedPrefix string
}

func (f *fakeTrimmingImageUploader) Upload(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	if f.uploadErr != nil {
		return "", f.uploadErr
	}
	n, err := io.Copy(io.Discard, body)
	if err != nil {
		return "", err
	}
	f.uploads = append(f.uploads, uploadCall{key: key, contentType: contentType, bodyBytes: n})
	return key, nil
}

func (f *fakeTrimmingImageUploader) Delete(_ context.Context, key string) error {
	f.deletes = append(f.deletes, key)
	return f.deleteErr
}

func (f *fakeTrimmingImageUploader) GetSignedURL(_ context.Context, key string, ttl time.Duration) (string, error) {
	if f.signErr != nil {
		return "", f.signErr
	}
	f.signCalls = append(f.signCalls, signCall{key: key, ttl: ttl})
	prefix := f.signedPrefix
	if prefix == "" {
		prefix = "https://signed.example/"
	}
	return prefix + key, nil
}

type quotaAcquireRecord struct {
	clinicID      uint64
	staffID       uint64
	declaredBytes int64
}

type fakeImageUploadQuotaStore struct {
	acquireErr error
	acquires   []quotaAcquireRecord
	releases   int
}

func (f *fakeImageUploadQuotaStore) Acquire(
	_ context.Context,
	clinicID, staffID uint64,
	declaredBytes int64,
) (func(context.Context), error) {
	if f.acquireErr != nil {
		return nil, f.acquireErr
	}
	f.acquires = append(f.acquires, quotaAcquireRecord{
		clinicID:      clinicID,
		staffID:       staffID,
		declaredBytes: declaredBytes,
	})
	return func(context.Context) { f.releases++ }, nil
}

// ---- helpers ----

var (
	testPNGBytes  = append([]byte("\x89PNG\r\n\x1a\n"), bytes.Repeat([]byte{0x01}, 64)...)
	testJPEGBytes = append([]byte("\xff\xd8\xff\xe0"), bytes.Repeat([]byte{0x02}, 64)...)
	testGIFBytes  = append([]byte("GIF89a"), bytes.Repeat([]byte{0x03}, 64)...)
)

// zeroReader streams n zero bytes without materializing them (oversize tests).
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) { return len(p), nil }

func newMultipartUploadRequest(
	t *testing.T,
	fieldName, fileName, mimeType string,
	fileBody io.Reader,
) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {`form-data; name="` + fieldName + `"; filename="` + fileName + `"`},
		"Content-Type":        {mimeType},
	})
	require.NoError(t, err)
	_, err = io.Copy(part, fileBody)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trimmings/images/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func newTrimmingHandlerContext(w *httptest.ResponseRecorder, req *http.Request) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	setTrimmingWriteContext(c)
	return c
}

func newImageHandler(svc Service, uploader *fakeTrimmingImageUploader, quota *fakeImageUploadQuotaStore) *Handler {
	h := newHandlerWithTrimmingSvc(svc)
	h.WithImageSupport(ImageDependencies{
		Uploader: uploader,
		Quota:    quota,
	})
	return h
}

func trimmingDetailReservation(styleImage, completedImage string) *model.Reservation {
	return &model.Reservation{
		ID:       7,
		ClinicID: 1,
		TrimmingDetail: &model.AppointmentTrimmingDetail{
			StyleImage:     styleImage,
			CompletedImage: completedImage,
		},
	}
}

// ---- UploadTrimmingImage ----

func TestUploadTrimmingImage_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	quota := &fakeImageUploadQuotaStore{}
	h := newImageHandler(&mockService{}, uploader, quota)

	req := newMultipartUploadRequest(t, "file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusCreated, w.Code)
	var body struct {
		Key       string `json:"key"`
		SignedURL string `json:"signed_url"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.True(t, strings.HasPrefix(body.Key, "trimming-images/1/"), "key must be clinic-scoped: %q", body.Key)
	assert.NotContains(t, body.Key, "style.png", "client filename must not leak into the object key")
	require.Len(t, uploader.uploads, 1)
	assert.Equal(t, body.Key, uploader.uploads[0].key)
	assert.Equal(t, "image/png", uploader.uploads[0].contentType)
	require.Len(t, uploader.signCalls, 1)
	assert.Equal(t, body.Key, uploader.signCalls[0].key)
	assert.Equal(t, 15*time.Minute, uploader.signCalls[0].ttl)
	assert.Equal(t, "https://signed.example/"+body.Key, body.SignedURL)
	require.Len(t, quota.acquires, 1)
	assert.Equal(t, uint64(1), quota.acquires[0].clinicID)
	assert.Equal(t, uint64(42), quota.acquires[0].staffID)
	assert.Equal(t, 1, quota.releases, "quota lease must be released")
	assert.Empty(t, uploader.deletes)
}

func TestUploadTrimmingImage_OversizedContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	quota := &fakeImageUploadQuotaStore{}
	h := newImageHandler(&mockService{}, uploader, quota)

	req := newMultipartUploadRequest(t, "file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	req.ContentLength = trimmingImageMaxRequestSize + 1
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Empty(t, quota.acquires, "quota must not be consumed for a pre-declared oversized request")
	assert.Empty(t, uploader.uploads)
}

func TestUploadTrimmingImage_OversizedChunkedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	quota := &fakeImageUploadQuotaStore{}
	h := newImageHandler(&mockService{}, uploader, quota)

	req := newMultipartUploadRequest(
		t, "file", "style.png", "image/png",
		io.LimitReader(zeroReader{}, int64(trimmingImageMaxRequestSize)),
	)
	// Force unknown length so the ContentLength pre-check cannot short-circuit;
	// MaxBytesReader must still bound the body.
	req.ContentLength = -1
	req.TransferEncoding = []string{"chunked"}
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	assert.Empty(t, uploader.uploads)
	assert.Equal(t, 1, quota.releases, "quota lease must still be released after reject")
}

func TestUploadTrimmingImage_FileSizeOverLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	quota := &fakeImageUploadQuotaStore{}
	h := newImageHandler(&mockService{}, uploader, quota)

	// Declared part size above the per-file cap but below the request cap is
	// rejected as invalid input (not 413) — mirrors the medical-record pipeline.
	req := newMultipartUploadRequest(
		t, "file", "style.png", "image/png",
		io.LimitReader(zeroReader{}, int64(trimmingImageMaxUploadSize)+1),
	)
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, uploader.uploads)
}

func TestUploadTrimmingImage_InvalidExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newImageHandler(&mockService{}, &fakeTrimmingImageUploader{}, &fakeImageUploadQuotaStore{})

	req := newMultipartUploadRequest(t, "file", "evil.exe", "image/png", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadTrimmingImage_ExtensionMIMEMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newImageHandler(&mockService{}, &fakeTrimmingImageUploader{}, &fakeImageUploadQuotaStore{})

	// .png 拡張子を image/gif と申告 — 拡張子と宣言 MIME の不一致は拒否。
	req := newMultipartUploadRequest(t, "file", "style.png", "image/gif", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadTrimmingImage_InvalidMIME(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newImageHandler(&mockService{}, &fakeTrimmingImageUploader{}, &fakeImageUploadQuotaStore{})

	req := newMultipartUploadRequest(t, "file", "doc.pdf", "application/pdf", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusBadRequest, w.Code, "pdf is not an allowed trimming image type")
}

func TestUploadTrimmingImage_SniffMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newImageHandler(&mockService{}, &fakeTrimmingImageUploader{}, &fakeImageUploadQuotaStore{})

	// 宣言 image/png だが実バイトはプレーンテキスト — content-sniff で拒否。
	req := newMultipartUploadRequest(t, "file", "style.png", "image/png", strings.NewReader("just some plain text"))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadTrimmingImage_MissingFileField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newImageHandler(&mockService{}, &fakeTrimmingImageUploader{}, &fakeImageUploadQuotaStore{})

	req := newMultipartUploadRequest(t, "not_file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUploadTrimmingImage_MissingClinicContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	quota := &fakeImageUploadQuotaStore{}
	h := newImageHandler(&mockService{}, uploader, quota)

	req := newMultipartUploadRequest(t, "file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	// clinic_id/user_id intentionally unset — clinic scope is fail-closed.
	h.UploadTrimmingImage(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Empty(t, uploader.uploads)
	assert.Empty(t, quota.acquires)
}

func TestUploadTrimmingImage_QuotaRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	quota := &fakeImageUploadQuotaStore{acquireErr: imageupload.ErrQuotaRateLimit}
	h := newImageHandler(&mockService{}, uploader, quota)

	req := newMultipartUploadRequest(t, "file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "rate limit exceeded")
	assert.Empty(t, uploader.uploads)
}

func TestUploadTrimmingImage_SharedStoreLimitMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	// Mirrors the error shape emitted by the shared postgres quota store:
	// a medical-record message wrapping the imageupload sentinel.
	quota := &fakeImageUploadQuotaStore{acquireErr: fmt.Errorf("medical record %w", imageupload.ErrQuotaRateLimit)}
	h := newImageHandler(&mockService{}, uploader, quota)

	req := newMultipartUploadRequest(t, "file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	// 429 body must be the trimming endpoint's fixed message — never the
	// wrapped store error's text (medical-record wording must not leak).
	assert.Contains(t, w.Body.String(), `"error":"image upload rate limit exceeded"`)
	assert.Empty(t, uploader.uploads)
}

func TestUploadTrimmingImage_QuotaUnavailableFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	quota := &fakeImageUploadQuotaStore{acquireErr: imageupload.ErrQuotaUnavailable}
	h := newImageHandler(&mockService{}, uploader, quota)

	req := newMultipartUploadRequest(t, "file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Empty(t, uploader.uploads)
}

func TestUploadTrimmingImage_NilDependenciesFailClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// No WithImageSupport — uploader/quota unset.
	h := newHandlerWithTrimmingSvc(&mockService{})

	req := newMultipartUploadRequest(t, "file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUploadTrimmingImage_SignFailureCleansUploadedObject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{signErr: errors.New("signing backend down")}
	quota := &fakeImageUploadQuotaStore{}
	h := newImageHandler(&mockService{}, uploader, quota)

	req := newMultipartUploadRequest(t, "file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	require.Len(t, uploader.uploads, 1)
	require.Equal(t, uploader.uploads[0].key, uploader.deletes[0],
		"the just-uploaded object must be deleted after downstream failure")
}

func TestUploadTrimmingImage_JPEGAndGIFAccepted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name     string
		filename string
		mime     string
		body     []byte
	}{
		{"jpeg", "style.jpg", "image/jpeg", testJPEGBytes},
		{"gif", "style.gif", "image/gif", testGIFBytes},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uploader := &fakeTrimmingImageUploader{}
			h := newImageHandler(&mockService{}, uploader, &fakeImageUploadQuotaStore{})

			req := newMultipartUploadRequest(t, "file", tc.filename, tc.mime, bytes.NewReader(tc.body))
			w := httptest.NewRecorder()
			h.UploadTrimmingImage(newTrimmingHandlerContext(w, req))

			assert.Equal(t, http.StatusCreated, w.Code)
		})
	}
}

// ---- requireAnyPermission route gating ----

func TestUploadTrimmingImageRoute_AnyOfCreateOrEditGates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var captured []PermissionRequirement
	allow := true
	stubAny := func(requirements ...PermissionRequirement) gin.HandlerFunc {
		captured = requirements
		return func(c *gin.Context) {
			if allow {
				c.Next()
				return
			}
			c.AbortWithStatus(http.StatusForbidden)
		}
	}

	uploader := &fakeTrimmingImageUploader{}
	h := NewHandlerWithPermission(&mockService{}, nil, nil, nil,
		func(_, _ string) gin.HandlerFunc { return func(c *gin.Context) { c.Next() } },
	).WithImageSupport(ImageDependencies{
		Uploader:             uploader,
		Quota:                &fakeImageUploadQuotaStore{},
		RequireAnyPermission: stubAny,
	})
	router := gin.New()
	h.RegisterRoutes(router.Group("/api/v1"))

	// The unkeyed upload endpoint must be gated by an any-of create|edit check.
	require.Equal(t, []PermissionRequirement{
		{Resource: "trimming", Action: "create"},
		{Resource: "trimming", Action: "edit"},
	}, captured)

	req := newMultipartUploadRequest(t, "file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	w := httptest.NewRecorder()
	// ServeHTTP path — middleware runs before the handler.
	router.ServeHTTP(w, req)
	// Without clinic context the handler fails closed, but the middleware ran.
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Denied any-of: handler never runs (no upload recorded).
	allow = false
	uploader.uploads = nil
	req2 := newMultipartUploadRequest(t, "file", "style.png", "image/png", bytes.NewReader(testPNGBytes))
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusForbidden, w2.Code)
	assert.Empty(t, uploader.uploads)
}

// ---- sign-on-read ----

func TestGetTrimming_SignsStoredImageKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		getByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return trimmingDetailReservation("trimming-images/1/style.jpg", "trimming-images/1/done.png"), nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trimmings/7", nil)
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.GetTrimming(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "https://signed.example/trimming-images/1/style.jpg", body["style_image"])
	assert.Equal(t, "https://signed.example/trimming-images/1/done.png", body["completed_image"])
	assert.Len(t, uploader.signCalls, 2)
}

func TestGetTrimming_PassesThroughNonObjectValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		getByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return trimmingDetailReservation("https://cdn.example/legacy.jpg", "data:image/png;base64,AAAA"), nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trimmings/7", nil)
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.GetTrimming(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "https://cdn.example/legacy.jpg", body["style_image"])
	assert.Equal(t, "data:image/png;base64,AAAA", body["completed_image"])
	assert.Empty(t, uploader.signCalls, "non-object values must not be signed")
}

func TestGetTrimming_EchoesForeignClinicKeyUnsigned(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		getByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			// Stored key belongs to clinic 2 while the record is clinic 1 —
			// never signed, but the read degrades to echo instead of 500.
			return trimmingDetailReservation("trimming-images/2/foreign.jpg", ""), nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trimmings/7", nil)
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.GetTrimming(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "trimming-images/2/foreign.jpg", body["style_image"])
	assert.Empty(t, uploader.signCalls, "foreign keys must never reach the signer")
}

func TestGetTrimming_EchoesMalformedKeyUnsigned(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		getByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return trimmingDetailReservation("trimming-images/", ""), nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trimmings/7", nil)
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.GetTrimming(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "trimming-images/", body["style_image"])
	assert.Empty(t, uploader.signCalls)
}

func TestGetTrimming_SignFailureEchoesStoredValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{signErr: errors.New("signing backend down")}
	svc := &mockService{
		getByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return trimmingDetailReservation("trimming-images/1/style.jpg", ""), nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trimmings/7", nil)
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.GetTrimming(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"style_image":"trimming-images/1/style.jpg"`,
		"sign failure degrades to unsigned echo of the stored key")
}

func TestListTrimmings_OneBadImageDoesNotFailList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		listFn: func(_ context.Context, _ uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
			return []model.Reservation{
				*trimmingDetailReservation("trimming-images/1/style.jpg", ""),
				*trimmingDetailReservation("trimming-images/2/foreign.jpg", "trimming-images/1/done.png"),
			}, 2, nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trimmings", nil)
	w := httptest.NewRecorder()
	h.ListTrimmings(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "https://signed.example/trimming-images/1/style.jpg")
	assert.Contains(t, w.Body.String(), `"style_image":"trimming-images/2/foreign.jpg"`)
	assert.NotContains(t, w.Body.String(), "signed.example/trimming-images/2/")
}

func TestListTrimmings_SignsImageKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		listFn: func(_ context.Context, _ uint64, _, _ *uint64, _, _ *string, _, _ int) ([]model.Reservation, int64, error) {
			return []model.Reservation{*trimmingDetailReservation("trimming-images/1/style.jpg", "")}, 1, nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trimmings", nil)
	w := httptest.NewRecorder()
	h.ListTrimmings(newTrimmingHandlerContext(w, req))

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "https://signed.example/trimming-images/1/style.jpg")
}

// ---- superseded-key cleanup ----

func TestUpdateTrimming_DeletesReplacedImageKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		getByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return trimmingDetailReservation("trimming-images/1/old-style.jpg", "trimming-images/1/done.png"), nil
		},
		updateFn: func(_ context.Context, _, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error) {
			require.NotNil(t, input.StyleImage)
			return trimmingDetailReservation(*input.StyleImage, "trimming-images/1/done.png"), nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/trimmings/7",
		strings.NewReader(`{"style_image":"trimming-images/1/new-style.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.UpdateTrimming(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{"trimming-images/1/old-style.jpg"}, uploader.deletes,
		"only the superseded key must be deleted")
}

func TestUpdateTrimming_DeletesClearedImageKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		getByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return trimmingDetailReservation("trimming-images/1/old-style.jpg", "trimming-images/1/done.png"), nil
		},
		updateFn: func(_ context.Context, _, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error) {
			require.NotNil(t, input.StyleImage)
			assert.Empty(t, *input.StyleImage)
			return trimmingDetailReservation("", "trimming-images/1/done.png"), nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/trimmings/7",
		strings.NewReader(`{"style_image":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.UpdateTrimming(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{"trimming-images/1/old-style.jpg"}, uploader.deletes)
}

func TestUpdateTrimming_UntouchedImagesNotDeleted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		getByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return trimmingDetailReservation("trimming-images/1/style.jpg", "trimming-images/1/done.png"), nil
		},
		updateFn: func(_ context.Context, _, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error) {
			assert.Nil(t, input.StyleImage)
			assert.Nil(t, input.CompletedImage)
			return trimmingDetailReservation("trimming-images/1/style.jpg", "trimming-images/1/done.png"), nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/trimmings/7",
		strings.NewReader(`{"remarks":"updated"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.UpdateTrimming(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, uploader.deletes, "untouched image keys must not be deleted")
}

func TestUpdateTrimming_NeverDeletesForeignKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		getByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return trimmingDetailReservation("trimming-images/2/foreign.jpg", "trimming-images/1/style.jpg"), nil
		},
		updateFn: func(_ context.Context, _, id uint64, input *UpdateTrimmingInput) (*model.Reservation, error) {
			return trimmingDetailReservation("trimming-images/1/new.jpg", "trimming-images/1/style.jpg"), nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/trimmings/7",
		strings.NewReader(`{"style_image":"trimming-images/1/new.jpg"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.UpdateTrimming(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, uploader.deletes, "trimming-images/2/foreign.jpg",
		"foreign-clinic keys must never be deleted")
}

func TestDeleteTrimming_DeletesOwnedImageKeys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		getByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return trimmingDetailReservation("trimming-images/1/style.jpg", "trimming-images/1/done.png"), nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/trimmings/7", nil)
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.DeleteTrimming(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.ElementsMatch(t,
		[]string{"trimming-images/1/style.jpg", "trimming-images/1/done.png"},
		uploader.deletes)
}

func TestDeleteTrimming_SkipsForeignAndLegacyValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		getByIDFn: func(_ context.Context, _, id uint64) (*model.Reservation, error) {
			return trimmingDetailReservation("trimming-images/2/foreign.jpg", "https://cdn.example/legacy.jpg"), nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/trimmings/7", nil)
	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, req)
	c.Params = gin.Params{{Key: "id", Value: "7"}}
	h.DeleteTrimming(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.Empty(t, uploader.deletes,
		"foreign keys and non-object values are never deleted by cleanup")
}

// ---- DeleteTrimmingImage ----

func newDeleteImageRequest(t *testing.T, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/trimmings/images", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestDeleteTrimmingImage_OwnedKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	h := newImageHandler(&mockService{}, uploader, &fakeImageUploadQuotaStore{})

	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w, newDeleteImageRequest(t, `{"key":"trimming-images/1/style.jpg"}`))
	h.DeleteTrimmingImage(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.Equal(t, []string{"trimming-images/1/style.jpg"}, uploader.deletes)
}

func TestDeleteTrimmingImage_QueryWrappedUploadsValueResolves(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	h := newImageHandler(&mockService{}, uploader, &fakeImageUploadQuotaStore{})

	w := httptest.NewRecorder()
	c := newTrimmingHandlerContext(w,
		newDeleteImageRequest(t, `{"key":"/uploads/trimming-images/1/style.jpg?v=123#frag"}`))
	h.DeleteTrimmingImage(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	assert.Equal(t, []string{"trimming-images/1/style.jpg"}, uploader.deletes,
		"query/fragment must be stripped before ownership check and delete")
}

func TestDeleteTrimmingImage_ForeignKeyRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	h := newImageHandler(&mockService{}, uploader, &fakeImageUploadQuotaStore{})

	w := httptest.NewRecorder()
	h.DeleteTrimmingImage(newTrimmingHandlerContext(w,
		newDeleteImageRequest(t, `{"key":"trimming-images/2/foreign.jpg"}`)))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Empty(t, uploader.deletes)
}

func TestDeleteTrimmingImage_MissingOrMalformedKeyRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	h := newImageHandler(&mockService{}, uploader, &fakeImageUploadQuotaStore{})

	for _, body := range []string{`{"key":""}`, `{"key":"trimming-images/"}`, `{}`, `not-json`} {
		w := httptest.NewRecorder()
		h.DeleteTrimmingImage(newTrimmingHandlerContext(w, newDeleteImageRequest(t, body)))
		assert.Equal(t, http.StatusBadRequest, w.Code, "body=%q", body)
	}
	assert.Empty(t, uploader.deletes)
}

func TestDeleteTrimmingImage_UploaderNotConfiguredFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWithTrimmingSvc(&mockService{})

	w := httptest.NewRecorder()
	h.DeleteTrimmingImage(newTrimmingHandlerContext(w,
		newDeleteImageRequest(t, `{"key":"trimming-images/1/style.jpg"}`)))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestDeleteTrimmingImage_ReferencedKeyRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		imageKeyInUseFn: func(_ context.Context, clinicID uint64, key string) (bool, error) {
			assert.Equal(t, uint64(1), clinicID)
			assert.Equal(t, "trimming-images/1/style.jpg", key)
			return true, nil
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	w := httptest.NewRecorder()
	h.DeleteTrimmingImage(newTrimmingHandlerContext(w,
		newDeleteImageRequest(t, `{"key":"trimming-images/1/style.jpg"}`)))

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Empty(t, uploader.deletes, "a record-referenced object must never be deleted")
}

func TestDeleteTrimmingImage_ReferenceCheckErrorFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	svc := &mockService{
		imageKeyInUseFn: func(context.Context, uint64, string) (bool, error) {
			return false, errors.New("db down")
		},
	}
	h := newImageHandler(svc, uploader, &fakeImageUploadQuotaStore{})

	w := httptest.NewRecorder()
	h.DeleteTrimmingImage(newTrimmingHandlerContext(w,
		newDeleteImageRequest(t, `{"key":"trimming-images/1/style.jpg"}`)))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Empty(t, uploader.deletes, "unchecked keys must not be deleted")
}

func TestDeleteTrimmingImage_WithoutServiceFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uploader := &fakeTrimmingImageUploader{}
	h := (&Handler{}).WithImageSupport(ImageDependencies{Uploader: uploader})

	w := httptest.NewRecorder()
	h.DeleteTrimmingImage(newTrimmingHandlerContext(w,
		newDeleteImageRequest(t, `{"key":"trimming-images/1/style.jpg"}`)))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Empty(t, uploader.deletes)
}
