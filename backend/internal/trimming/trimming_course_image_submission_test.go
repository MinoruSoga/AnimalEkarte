package trimming

import (
	"context"
	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestCourseImageSubmission_ProbeCommitsBeforeOtherFieldTypeRejection(t *testing.T) {
	for _, wireKey := range []string{draftKey, "/uploads/" + draftKey + "?v=1", " " + draftKey + " "} {
		t.Run(wireKey, func(t *testing.T) { assertCourseProbeBeforeRejection(t, wireKey, "image_object_key") })
	}
}
func assertCourseProbeBeforeRejection(t *testing.T, wireKey string, field string) {
	gin.SetMode(gin.TestMode)
	svc, repo, _, tx := newDraftTest(t)
	if err := svc.ReserveUpload(t.Context(), 7, 12, draftKey); err != nil {
		t.Fatal(err)
	}
	if err := svc.CompleteUpload(t.Context(), 7, 12, draftKey); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "12")
	c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"`+field+`":"`+wireKey+`","name":false}`))
	c.Request = c.Request.WithContext(persistence.WithTxValue(c.Request.Context(), &gorm.DB{}))
	if !protectCourseImageBody(c, 7, svc, "image_object_key") {
		t.Fatal(w.Body.String())
	}
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err == nil {
		t.Fatal("invalid field unexpectedly accepted")
	}
	if repo.row.SubmittedAt == nil || tx.committed != 3 {
		t.Fatal("business rejection lost committed submission")
	}
	if ok, _ := svc.ReserveDiscard(t.Context(), 7, 12, draftKey); ok {
		t.Fatal("rejected image still eligible")
	}
}

func TestCourseImageSubmission_CaseInsensitiveAcceptedTagsProtected(t *testing.T) {
	for _, field := range []string{"IMAGE_OBJECT_KEY", "STYLE_IMAGE", "COMPLETED_IMAGE"} {
		t.Run(field, func(t *testing.T) {
			// Probe protects the received field before any later struct/business validation.
			svc, repo, _, _ := newDraftTest(t)
			if err := svc.ReserveUpload(t.Context(), 7, 12, draftKey); err != nil {
				t.Fatal(err)
			}
			if err := svc.CompleteUpload(t.Context(), 7, 12, draftKey); err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Set("user_id", "12")
			c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"`+field+`":"`+draftKey+`","name":false}`))
			if !protectCourseImageBody(c, 7, svc, "image_object_key", "style_image", "completed_image") || repo.row.SubmittedAt == nil {
				t.Fatal("case alias escaped protection")
			}
		})
	}
}

type submissionProbeSpy struct {
	CourseImageDrafts
	keys []string
}

func (s *submissionProbeSpy) ProtectSubmission(_ context.Context, _, _ uint64, key string) error {
	s.keys = append(s.keys, key)
	if len(s.keys) == 1 {
		return apperrors.WrapConflict("first key denied")
	}
	return nil
}
func TestCourseImageSubmission_DuplicateAndCaseAliasesAllProtectedBeforeError(t *testing.T) {
	second := "trimming-images/7/456789_abcdef0123456789abcdef0123456789.png"
	spy := &submissionProbeSpy{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "12")
	c.Request = httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(`{"image_object_key":"`+draftKey+`","image_object_key":"`+second+`","IMAGE_OBJECT_KEY":"/uploads/`+draftKey+`?v=1","name":false}`))
	if protectCourseImageBody(c, 7, spy, "image_object_key") {
		t.Fatal("first error was lost")
	}
	if !reflect.DeepEqual(spy.keys, []string{draftKey, second, draftKey}) {
		t.Fatalf("received key escaped protection: %v", spy.keys)
	}
}
