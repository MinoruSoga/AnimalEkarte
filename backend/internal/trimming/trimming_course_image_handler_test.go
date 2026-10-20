package trimming

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type courseHandlerQuota struct{ acquired bool }

func (q *courseHandlerQuota) Acquire(context.Context, uint64, uint64, int64) (func(context.Context), error) {
	q.acquired = true
	return func(context.Context) {}, nil
}

type courseHandlerStorage struct {
	t                             *testing.T
	repo                          *draftTestRepo
	life                          *draftTestLife
	tx                            *draftTestTx
	quota                         *courseHandlerQuota
	key                           string
	deletes                       int
	uploadErr, signErr, deleteErr error
}

func (s *courseHandlerStorage) Upload(_ context.Context, key string, r io.Reader, _ string) (string, error) {
	if !s.quota.acquired || s.repo.row == nil || s.repo.row.ObjectKey != key {
		s.t.Fatal("storage upload before quota/provenance reservation")
	}
	s.key = key
	if s.uploadErr != nil {
		return "", s.uploadErr
	}
	_, err := io.Copy(io.Discard, r)
	return key, err
}
func (s *courseHandlerStorage) GetSignedURL(context.Context, string, time.Duration) (string, error) {
	if s.repo.row.UploadedAt == nil {
		s.t.Fatal("signed before upload completion")
	}
	return "https://signed.example/image", s.signErr
}
func (s *courseHandlerStorage) Delete(_ context.Context, key string) error {
	if !s.life.gone || s.tx.committed < 3 {
		s.t.Fatal("storage delete before committed tombstone")
	}
	s.deletes++
	return s.deleteErr
}
func courseHandlerContext(w *httptest.ResponseRecorder, req *http.Request, actor string) *gin.Context {
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("clinic_id", "7")
	c.Set("user_id", actor)
	return c
}
func courseHandlerUploadRequest(t *testing.T) *http.Request {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	part, err := writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {`form-data; name="file"; filename="photo.png"`}, "Content-Type": {"image/png"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte{0x89, 'P', 'N', 'G', 13, 10, 26, 10, 0, 0, 0, 0}); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/masters/trimming-courses/images/upload", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}
func TestCourseImageHandlers_RealUploadProtocolAndDiscardRetry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc, repo, life, tx := newDraftTest(t)
	quota := &courseHandlerQuota{}
	storage := &courseHandlerStorage{t: t, repo: repo, life: life, tx: tx, quota: quota}
	h := (&Handler{}).WithImageSupport(ImageDependencies{Uploader: storage, Quota: quota, CourseImages: svc})
	w := httptest.NewRecorder()
	h.UploadCourseImage(courseHandlerContext(w, courseHandlerUploadRequest(t), "12"))
	if w.Code != http.StatusCreated {
		t.Fatal(w.Code, w.Body.String())
	}
	var response struct {
		Key string `json:"key"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !generatedCourseImageKey.MatchString(response.Key) || repo.row.UploadedAt == nil || repo.row.SubmittedAt != nil {
		t.Fatal("invalid upload proof")
	}
	discard := func(actor string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(`{"key":"`+response.Key+`"}`))
		c := courseHandlerContext(w, req, actor)
		h.DiscardCourseImage(c)
		c.Writer.WriteHeaderNow()
		return w.Code
	}
	if status := discard("99"); status != http.StatusConflict || storage.deletes != 0 {
		t.Fatal("another staff deleted draft", status)
	}
	storage.deleteErr = errors.New("storage unavailable")
	if status := discard("12"); status != http.StatusInternalServerError || !life.gone {
		t.Fatal("lost retry reservation", status)
	}
	storage.deleteErr = nil
	if status := discard("12"); status != http.StatusNoContent || storage.deletes != 2 {
		t.Fatal("same proof retry failed", status)
	}
}
func TestCourseImageHandlers_UploadAndSigningFailureStayProtected(t *testing.T) {
	for _, failure := range []string{"upload", "sign"} {
		t.Run(failure, func(t *testing.T) {
			svc, repo, life, tx := newDraftTest(t)
			quota := &courseHandlerQuota{}
			storage := &courseHandlerStorage{t: t, repo: repo, life: life, tx: tx, quota: quota}
			if failure == "upload" {
				storage.uploadErr = errors.New("unknown upload result")
			} else {
				storage.signErr = errors.New("sign unavailable")
			}
			h := (&Handler{}).WithImageSupport(ImageDependencies{Uploader: storage, Quota: quota, CourseImages: svc})
			w := httptest.NewRecorder()
			h.UploadCourseImage(courseHandlerContext(w, courseHandlerUploadRequest(t), "12"))
			if w.Code == http.StatusCreated || repo.row == nil || repo.row.SubmittedAt == nil || storage.deletes != 0 {
				t.Fatal("failed upload/sign became discardable", w.Code, w.Body.String())
			}
			if ok, _ := svc.ReserveDiscard(t.Context(), 7, 12, storage.key); ok {
				t.Fatal("unknown upload/sign result discarded")
			}
		})
	}
}
func TestCourseImageHandlers_MasterGateAndGenericRetention(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, allow := range []bool{false, true} {
		h := (&Handler{}).WithImageSupport(ImageDependencies{RequireAnyPermission: func(reqs ...PermissionRequirement) gin.HandlerFunc {
			return func(c *gin.Context) {
				if !allow || len(reqs) != 2 || reqs[0].Resource != "master-trimming" {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
				c.Next()
			}
		}})
		router := gin.New()
		router.POST("/course", h.courseImagePermission(), h.UploadCourseImage)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/course", nil))
		if !allow && w.Code != http.StatusForbidden {
			t.Fatal("view/trimming-only allowed")
		}
		if allow && w.Code != http.StatusUnauthorized {
			t.Fatal("missing identity accepted", w.Code)
		}
	}
	h := &Handler{}
	w := httptest.NewRecorder()
	h.DeleteTrimmingImage(courseHandlerContext(w, httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(`{"key":"trimming-images/7/unknown.jpg"}`)), "12"))
	if w.Code != http.StatusForbidden {
		t.Fatal("unregistered cleanup reopened")
	}
}
