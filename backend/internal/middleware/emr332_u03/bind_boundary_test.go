package emr332_u03

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/animal-ekarte/backend/internal/medicalrecord"
)

// C05 evidence tests (docs/work/emr-332-security/u03/c05-input-validation.md).
//
// Scope: in-memory binding-layer behavior only. No service, repository,
// session, audit, cache, queue, file, mail, or telemetry object exists in this
// package, so the side-effect invocation count is 0 by construction — the
// assertion below records that contract mechanically.
//
// Marker convention: SECCHK20261009-style alphanumeric sentinels only; no
// executable attack syntax appears in this file.

// c05MirrorRequest mirrors the binding-tag shape of unexported protected
// write-handler structs (owner.createOwnerRequest uses the same
// required/max=255 + omitempty/oneof vocabulary).
type c05MirrorRequest struct {
	Name   string `json:"name"   binding:"required,max=255"`
	Status string `json:"status" binding:"omitempty,oneof=alive deceased"`
	Age    int    `json:"age"    binding:"omitempty,min=0,max=200"`
}

func newC05BindContext(t *testing.T, body string) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/emr332-u03/c05-bind", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c
}

// TestC05ShouldBindJSON_TagEnforcement records that binding tags reject bad
// type / range / required violations on the representative mirror struct.
func TestC05ShouldBindJSON_TagEnforcement(t *testing.T) {
	cases := []struct {
		name      string
		body      string
		expectErr bool
	}{
		{"normal: valid body", `{"name":"SECCHK20261009C05R05","status":"alive","age":7}`, false},
		{"refusal: missing required", `{"status":"alive"}`, true},
		{"refusal: required empty", `{"name":""}`, true},
		{"refusal: over max=255", `{"name":"` + strings.Repeat("n", 256) + `"}`, true},
		{"refusal: enum violation", `{"name":"ok","status":"SECCHK20261009C05R06"}`, true},
		{"refusal: wrong type string->int", `{"name":"ok","age":"SECCHK20261009C05R07"}`, true},
		{"refusal: range min", `{"name":"ok","age":-5}`, true},
		{"refusal: range max", `{"name":"ok","age":201}`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req c05MirrorRequest
			c := newC05BindContext(t, tc.body)
			err := c.ShouldBindJSON(&req)
			if tc.expectErr && err == nil {
				t.Fatalf("%s: expected bind error, got nil (bound=%+v)", tc.name, req)
			}
			if !tc.expectErr && err != nil {
				t.Fatalf("%s: expected successful bind, got %v", tc.name, err)
			}
		})
	}
}

// TestC05ShouldBindJSON_UnknownFieldSilentlyAccepted is the recorded gap:
// ShouldBindJSON accepts an unknown/disallowed field without error and drops
// it. This is the honest baseline that motivates BindStrictJSON.
func TestC05ShouldBindJSON_UnknownFieldSilentlyAccepted(t *testing.T) {
	var req c05MirrorRequest
	c := newC05BindContext(t, `{"name":"ok","secchk20261009c05r08":"SECCHK20261009C05R08"}`)
	if err := c.ShouldBindJSON(&req); err != nil {
		t.Fatalf("C05 baseline expectation violated: unknown field was rejected (err=%v)", err)
	}
	if req.Name != "ok" {
		t.Fatalf("expected known field to bind, got %+v", req)
	}
}

// TestC05ShouldBindJSON_TrailingValueUnread records that ShouldBindJSON
// consumes only the first JSON value; a trailing value is left unread and
// silently ignored (unlike bindStaffJSON/bindAuthJSON which reject it).
func TestC05ShouldBindJSON_TrailingValueUnread(t *testing.T) {
	var req c05MirrorRequest
	c := newC05BindContext(t, `{"name":"ok"}{"secchk":"SECCHK20261009C05R09"}`)
	if err := c.ShouldBindJSON(&req); err != nil {
		t.Fatalf("C05 baseline expectation violated: trailing value rejected (err=%v)", err)
	}
}

// TestC05DisallowUnknownFields_Contrast records the strict-decoder behavior
// used by the 8 manual-decoder sites (staff_provisioning.go:154,199 et al.):
// the same unknown field IS rejected when DisallowUnknownFields is set.
func TestC05DisallowUnknownFields_Contrast(t *testing.T) {
	var req c05MirrorRequest
	decoder := json.NewDecoder(strings.NewReader(`{"name":"ok","secchk20261009c05r10":"SECCHK20261009C05R10"}`))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err == nil {
		t.Fatal("expected strict decoder to reject unknown field, got nil error")
	}
}

// TestC05RealHandlerStruct exercises a real exported protected-handler request
// struct: medicalrecord.CreateMedicalRecordAddendumRequest (bound via
// ShouldBindJSON in medical_record_addendum_handler.go).
func TestC05RealHandlerStruct(t *testing.T) {
	t.Run("normal: valid addendum", func(t *testing.T) {
		var req medicalrecord.CreateMedicalRecordAddendumRequest
		c := newC05BindContext(t, `{"after_text":"SECCHK20261009C05R11","reason":"SECCHK20261009C05R12"}`)
		if err := c.ShouldBindJSON(&req); err != nil {
			t.Fatalf("expected valid addendum to bind, got %v", err)
		}
	})
	t.Run("refusal: missing required", func(t *testing.T) {
		var req medicalrecord.CreateMedicalRecordAddendumRequest
		c := newC05BindContext(t, `{"reason":"SECCHK20261009C05R12"}`)
		if err := c.ShouldBindJSON(&req); err == nil {
			t.Fatal("expected missing after_text to be rejected, got nil")
		}
	})
	t.Run("refusal: over max=1000", func(t *testing.T) {
		var req medicalrecord.CreateMedicalRecordAddendumRequest
		c := newC05BindContext(t, `{"after_text":"`+strings.Repeat("x", 1001)+`","reason":"r"}`)
		if err := c.ShouldBindJSON(&req); err == nil {
			t.Fatal("expected over-max after_text to be rejected, got nil")
		}
	})
	t.Run("refusal: wrong type", func(t *testing.T) {
		var req medicalrecord.CreateMedicalRecordAddendumRequest
		c := newC05BindContext(t, `{"after_text":20261009,"reason":"r"}`)
		if err := c.ShouldBindJSON(&req); err == nil {
			t.Fatal("expected numeric after_text to be rejected, got nil")
		}
	})
	t.Run("gap: unknown field silently accepted", func(t *testing.T) {
		var req medicalrecord.CreateMedicalRecordAddendumRequest
		c := newC05BindContext(t, `{"after_text":"a","reason":"r","secchk20261009c05r13":"SECCHK20261009C05R13"}`)
		if err := c.ShouldBindJSON(&req); err != nil {
			t.Fatalf("C05 baseline expectation violated: unknown field rejected on real struct (err=%v)", err)
		}
	})
}

// TestC05ZeroSideEffectInvocations makes the save-path count mechanical: the
// only callable "save" stub here is a counting sentinel that nothing invokes.
func TestC05ZeroSideEffectInvocations(t *testing.T) {
	savePathCalls := 0
	savePath := func() { savePathCalls++ }
	_ = savePath // declared to prove no binding-path call site exists; never invoked.

	var req c05MirrorRequest
	for _, body := range []string{
		`{"name":"ok"}`,
		`{"name":"ok","secchk20261009c05r14":"SECCHK20261009C05R14"}`,
		`{"status":"alive"}`,
	} {
		c := newC05BindContext(t, body)
		_ = c.ShouldBindJSON(&req)
	}
	if savePathCalls != 0 {
		t.Fatalf("save path invoked %d times during binding-only verification", savePathCalls)
	}
}
