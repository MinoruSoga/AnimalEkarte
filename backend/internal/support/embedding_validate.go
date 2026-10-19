package support

import (
	"database/sql/driver"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/animal-ekarte/backend/internal/manualarticle"
)

// embedding_validate.go — ベクトル検証・正規化・real[] スキャナ/バリュアの
// 共有純粋関数。client(embedding.go)・cache(manual_cache.go) の両境界が使う。

// normalizeEmbedding は厳密 1536 dim・全要素有限・ノルム>0(非overflow)を検証し、
// L2 正規化コピーを返す。元スライスは変更しない。
func normalizeEmbedding(vec []float64) ([]float64, error) {
	if len(vec) != manualarticle.EmbeddingDimensions {
		return nil, fmt.Errorf("invalid embedding dimensions %d", len(vec))
	}
	var sum float64
	out := make([]float64, len(vec))
	for i, v := range vec {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("non-finite embedding element")
		}
		out[i] = v
		sum += v * v
	}
	norm := math.Sqrt(sum)
	if !(norm > 0) || math.IsInf(norm, 0) {
		return nil, fmt.Errorf("invalid embedding norm")
	}
	for i := range out {
		out[i] /= norm
	}
	return out, nil
}

// embeddingArray は PostgreSQL 1次元 real[] の Scan/Value 実装（外部依存なし）。
// NULL・多次元・非有限値は拒否する。
type embeddingArray []float32

// Scan は `{f1,f2,...}` 形式のフラット配列だけを受け付ける。
func (a *embeddingArray) Scan(src any) error {
	if src == nil {
		return fmt.Errorf("embedding column is NULL")
	}
	var raw string
	switch v := src.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	default:
		return fmt.Errorf("unsupported embedding column type %T", src)
	}
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "{") || !strings.HasSuffix(raw, "}") || strings.Contains(raw, "{{") {
		return fmt.Errorf("invalid embedding array literal")
	}
	inner := raw[1 : len(raw)-1]
	if inner == "" {
		return fmt.Errorf("empty embedding array")
	}
	parts := strings.Split(inner, ",")
	out := make(embeddingArray, len(parts))
	for i, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 32)
		if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
			return fmt.Errorf("invalid embedding element")
		}
		out[i] = float32(f)
	}
	*a = out
	return nil
}

// Value は `{f1,...}` literal を返す。
func (a embeddingArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return nil, fmt.Errorf("empty embedding array")
	}
	var b strings.Builder
	b.WriteByte('{')
	for i, f := range a {
		if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
			return nil, fmt.Errorf("non-finite embedding element")
		}
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(f), 'g', -1, 32))
	}
	b.WriteByte('}')
	return b.String(), nil
}

var sha256HexRe = regexp.MustCompile(`^[0-9a-f]{64}$`)

// validatedCacheRows は Load が返す行を検証・正規化する純粋ヘルパー。
// requested キーとの exact intersection のみ返し、無効行・未知キーは採用しない。
func validatedCacheRows(rows []supportManualEmbeddingRow, requested map[string]struct{}, expectedModel string) map[string][]float64 {
	out := make(map[string][]float64, len(rows))
	for _, row := range rows {
		if _, want := requested[row.CacheKey]; !want {
			continue
		}
		if row.Model != expectedModel || expectedModel == "" ||
			row.TaskType != manualarticle.TaskTypeRetrievalDocument ||
			row.ChunkerVersion != manualarticle.ChunkerVersion ||
			row.Dimensions != manualarticle.EmbeddingDimensions {
			continue
		}
		vec := make([]float64, len(row.Embedding))
		for i, f := range row.Embedding {
			vec[i] = float64(f)
		}
		n, err := normalizeEmbedding(vec)
		if err != nil {
			continue
		}
		out[row.CacheKey] = n
	}
	return out
}
