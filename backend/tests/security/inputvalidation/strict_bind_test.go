package inputvalidation

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// strictBindFixture mirrors the binding-tag shape used by protected write
// handlers: a required bounded field, an enum, and a bounded integer. Handler
// request structs are unexported, so verification uses this mirror plus one
// real exported struct (bind_boundary_test.go).
type strictBindFixture struct {
	Name   string `json:"name"   binding:"required,max=32"`
	Status string `json:"status" binding:"omitempty,oneof=alpha beta"`
	Count  int    `json:"count"  binding:"omitempty,min=0,max=100"`
}

func newStrictBindContext(t *testing.T, body string, contentLength int64) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodPost, "/emr332-u03/strict-bind", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if contentLength >= 0 {
		req.ContentLength = contentLength
	}
	c.Request = req
	return c
}

// Marker strings are SECCHK20261009-style alphanumeric sentinels only —
// no executable attack syntax appears in this file (campaign safety bound).

func TestBindStrictJSON_AcceptsValidBody(t *testing.T) {
	var dst strictBindFixture
	c := newStrictBindContext(t, `{"name":"SECCHK20261009C05R01","status":"alpha","count":3}`, -1)
	if err := BindStrictJSON(c, &dst, 1024); err != nil {
		t.Fatalf("expected valid body to bind, got %v", err)
	}
	if dst.Name != "SECCHK20261009C05R01" || dst.Status != "alpha" || dst.Count != 3 {
		t.Fatalf("unexpected bound values: %+v", dst)
	}
}

func TestBindStrictJSON_RejectsUnknownField(t *testing.T) {
	// C05 gap closure: the same unknown field that ShouldBindJSON silently
	// drops must be rejected here.
	var dst strictBindFixture
	c := newStrictBindContext(t, `{"name":"ok","secchk20261009c05r02":"SECCHK20261009C05R02"}`, -1)
	if err := BindStrictJSON(c, &dst, 1024); err == nil {
		t.Fatal("expected unknown field to be rejected, got nil error")
	}
}

func TestBindStrictJSON_RejectsTrailingJSONValue(t *testing.T) {
	var dst strictBindFixture
	c := newStrictBindContext(t, `{"name":"ok"}{"name":"extra"}`, -1)
	if err := BindStrictJSON(c, &dst, 1024); err == nil {
		t.Fatal("expected trailing JSON value to be rejected, got nil error")
	}
}

func TestBindStrictJSON_RejectsOversizedBody(t *testing.T) {
	var dst strictBindFixture
	body := `{"name":"` + strings.Repeat("a", 64) + `"}`
	c := newStrictBindContext(t, body, int64(len(body)))
	if err := BindStrictJSON(c, &dst, 8); err == nil {
		t.Fatal("expected oversized body to be rejected, got nil error")
	}
}

func TestBindStrictJSON_RejectsOversizedDeclaredLength(t *testing.T) {
	// Declared Content-Length over the cap is refused before reading.
	var dst strictBindFixture
	c := newStrictBindContext(t, `{}`, 4096)
	if err := BindStrictJSON(c, &dst, 64); err == nil {
		t.Fatal("expected oversized declared length to be rejected, got nil error")
	}
}

func TestBindStrictJSON_EnforcesBindingTags(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"missing required", `{"status":"alpha"}`},
		{"required empty", `{"name":""}`},
		{"over max length", `{"name":"` + strings.Repeat("b", 33) + `"}`},
		{"enum violation", `{"name":"ok","status":"SECCHK20261009C05R03"}`},
		{"range violation", `{"name":"ok","count":101}`},
		{"negative range", `{"name":"ok","count":-1}`},
		{"wrong type", `{"name":"ok","count":"SECCHK20261009C05R04"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var dst strictBindFixture
			c := newStrictBindContext(t, tc.body, -1)
			if err := BindStrictJSON(c, &dst, 1024); err == nil {
				t.Fatalf("expected %s to be rejected, got nil error", tc.name)
			}
		})
	}
}

func TestBindStrictJSON_RejectsMalformedAndEmpty(t *testing.T) {
	for _, body := range []string{"", "{", "not-json", `[1,2,3]`} {
		var dst strictBindFixture
		c := newStrictBindContext(t, body, -1)
		if err := BindStrictJSON(c, &dst, 1024); err == nil {
			t.Fatalf("expected body %q to be rejected, got nil error", body)
		}
	}
}
