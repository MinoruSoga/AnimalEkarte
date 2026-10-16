package support

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/animal-ekarte/backend/internal/manualarticle"
)

func vecBody(model string, idx any, n int, val float64) string {
	vec := strings.Repeat(fmt.Sprintf("%v,", val), n-1) + fmt.Sprintf("%v", val)
	return fmt.Sprintf(`{"model":%q,"data":[{"index":%v,"embedding":[%s]}]}`, model, idx, vec)
}

func okServer(t *testing.T, body string, calls *int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls != nil {
			*calls++
		}
		if r.URL.Path != "/embeddings" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Write([]byte(body))
	}))
}

func TestEmbed_StrictContract(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{"ok", vecBody("m", 0, 1536, 0.5), true},
		{"index_missing", fmt.Sprintf(`{"model":"m","data":[{"embedding":[%s]}]}`, strings.Repeat("0.5,", 1535)+"0.5"), false},
		{"index_1", vecBody("m", 1, 1536, 0.5), false},
		{"data0", `{"model":"m","data":[]}`, false},
		{"wrong_model", vecBody("other", 0, 1536, 0.5), false},
		{"missing_model", fmt.Sprintf(`{"data":[{"index":0,"embedding":[%s]}]}`, strings.Repeat("0.5,", 1535)+"0.5"), false},
		{"len1535", vecBody("m", 0, 1535, 0.5), false},
		{"zero_vector", vecBody("m", 0, 1536, 0), false},
		{"null_elem", strings.Replace(vecBody("m", 0, 1536, 0.5), "0.5,", "null,", 1), false},
		{"huge_finite", vecBody("m", 0, 1536, 1e308), false},
		{"trailing", vecBody("m", 0, 1536, 0.5) + ` {"x":1}`, false},
	}
	for _, tc := range cases {
		calls := 0
		srv := okServer(t, tc.body, &calls)
		c := NewSarashinaEmbedder(srv.URL, "k", "m", time.Second)
		_, err := c.Embed(context.Background(), manualarticle.TaskTypeRetrievalQuery, "q")
		srv.Close()
		if tc.want && err != nil {
			t.Fatalf("%s: unexpected err %v", tc.name, err)
		}
		if !tc.want && err == nil {
			t.Fatalf("%s: expected rejection", tc.name)
		}
		if calls != 1 {
			t.Fatalf("%s: expected exactly 1 upstream call, got %d", tc.name, calls)
		}
	}
}

func TestEmbed_InputBoundsAndTaskType(t *testing.T) {
	c := NewSarashinaEmbedder("http://127.0.0.1:1", "k", "m", time.Second)
	for _, bad := range []string{"", strings.Repeat("x", 2049)} {
		if _, err := c.Embed(context.Background(), manualarticle.TaskTypeRetrievalQuery, bad); err == nil {
			t.Fatal("expected input bound rejection without HTTP")
		}
	}
	if _, err := c.Embed(context.Background(), "BAD", "x"); err == nil {
		t.Fatal("expected task_type rejection")
	}
}

func TestEmbed_NormAndStatus(t *testing.T) {
	for _, code := range []int{400, 401, 429, 500} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(code) }))
		c := NewSarashinaEmbedder(srv.URL, "k", "m", time.Second)
		if _, err := c.Embed(context.Background(), manualarticle.TaskTypeRetrievalQuery, "q"); err == nil {
			t.Fatalf("status %d must fail", code)
		}
		srv.Close()
	}
	srv := okServer(t, vecBody("m", 0, 1536, 2.0), nil)
	defer srv.Close()
	v, err := NewSarashinaEmbedder(srv.URL, "k", "m", time.Second).Embed(context.Background(), manualarticle.TaskTypeRetrievalDocument, "d")
	if err != nil {
		t.Fatal(err)
	}
	var n float64
	for _, x := range v {
		n += x * x
	}
	if n < 0.999 || n > 1.001 {
		t.Fatalf("not normalized: %v", n)
	}
}
