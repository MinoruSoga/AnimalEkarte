package support

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/animal-ekarte/backend/internal/manualarticle"
)

var key64 = strings.Repeat("a", 64)

func unitVec() []float64 {
	v := make([]float64, 1536)
	for i := range v {
		v[i] = 0.025
	}
	return v
}

func TestNormalizeEmbedding(t *testing.T) {
	v, err := normalizeEmbedding(unitVec())
	if err != nil {
		t.Fatal(err)
	}
	var n float64
	for _, x := range v {
		n += x * x
	}
	if math.Abs(n-1) > 1e-6 {
		t.Fatalf("norm %v", n)
	}
	for name, bad := range map[string][]float64{
		"dim1535": make([]float64, 1535),
		"zero":    make([]float64, 1536),
	} {
		if _, err := normalizeEmbedding(bad); err == nil {
			t.Fatalf("%s must fail", name)
		}
	}
	nan := unitVec()
	nan[0] = math.NaN()
	if _, err := normalizeEmbedding(nan); err == nil {
		t.Fatal("NaN must fail")
	}
	big := make([]float64, 1536)
	for i := range big {
		big[i] = 1e308
	}
	if _, err := normalizeEmbedding(big); err == nil {
		t.Fatal("overflowing norm must fail")
	}
}

func TestEmbeddingArrayScanValueRoundtrip(t *testing.T) {
	src := embeddingArray{1.5, -2.25, 0.001}
	val, err := src.Value()
	if err != nil {
		t.Fatal(err)
	}
	var dst embeddingArray
	if err := dst.Scan(val); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 3 || dst[0] != 1.5 {
		t.Fatalf("roundtrip failed: %v", dst)
	}
	for _, bad := range []struct {
		n string
		s any
	}{{"null", nil}, {"nested", "{{1,2},{3}}"}, {"empty", "{}"}, {"nan", "{NaN}"}, {"multi", "{1,2},{3}"}} {
		var a embeddingArray
		if err := a.Scan(bad.s); err == nil {
			t.Fatalf("case %s must fail", bad.n)
		}
	}
}

func TestValidatedCacheRows(t *testing.T) {
	good := supportManualEmbeddingRow{CacheKey: key64, Model: "m", TaskType: manualarticle.TaskTypeRetrievalDocument, ChunkerVersion: 1, Dimensions: 1536, Embedding: embeddingArray(make([]float32, 1536))}
	for i := range good.Embedding {
		good.Embedding[i] = 0.02
	}
	req := map[string]struct{}{key64: {}, strings.Repeat("b", 64): {}}
	out := validatedCacheRows([]supportManualEmbeddingRow{good}, req, "m")
	if len(out) != 1 {
		t.Fatalf("expected 1 valid row, got %d", len(out))
	}
	var n float64
	for _, x := range out[key64] {
		n += x * x
	}
	if math.Abs(n-1) > 1e-3 {
		t.Fatalf("loaded vector must be re-normalized: %v", n)
	}
	badModel := good
	badModel.Model = "other"
	badTask := good
	badTask.TaskType = "RETRIEVAL_QUERY"
	unknown := good
	unknown.CacheKey = strings.Repeat("c", 64)
	out = validatedCacheRows([]supportManualEmbeddingRow{badModel, badTask, unknown}, req, "m")
	if len(out) != 0 {
		t.Fatal("invalid/unknown rows must be skipped")
	}
}

func TestCacheNilDB(t *testing.T) {
	c := NewManualEmbeddingCache(nil, "m")
	if _, err := c.Load(context.Background(), []string{key64}); err == nil {
		t.Fatal("nil DB must error")
	}
	if err := c.StoreWithKey(context.Background(), key64, "m", manualarticle.TaskTypeRetrievalDocument, 1, 1536, unitVec()); err == nil {
		t.Fatal("nil DB must error")
	}
}
