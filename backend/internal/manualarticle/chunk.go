package manualarticle

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

// チャンク分割パラメータ（EMR-410 R3・chunker_version=1 固定設計値。調整禁止）。
const (
	ChunkerVersion            = 1
	chunkBodyMaxRunes         = 400
	chunkContinuationOverlap  = 40
	chunkTitleMaxRunes        = 64
	chunkHeadingMaxRunes      = 64
	embeddingInputMaxBytes    = 2048
	articleBodyMaxBytes       = 128 << 10 // 128KiB / 記事
	corpusBodyMaxBytes        = 2 << 20   // 2MiB / コーパス
	CorpusMaxChunks           = 2048
	TaskTypeRetrievalDocument = "RETRIEVAL_DOCUMENT"
	TaskTypeRetrievalQuery    = "RETRIEVAL_QUERY"
	EmbeddingDimensions       = 1536
)

// ErrCorpusTooLarge は意味検索コーパスが容量上限を超えたことを示す
// （部分索引化はしない。呼出側は keyword-only へ縮退する）。
var ErrCorpusTooLarge = errors.New("manual corpus exceeds capacity limits")

// Chunk は 400 codepoints 以下の意味検索単位。
type Chunk struct {
	ID             string // sha256([category/slug, articleHash, version, ordinal])
	ArticleID      string // category/slug
	Ordinal        int
	Heading        string
	Body           string
	EmbeddingInput string // title + "\n" + heading + "\n" + body
}

var headingRe = regexp.MustCompile(`^(#{1,6})[ \t]+(.*)$`)
var imageOnlyRe = regexp.MustCompile(`(?s)\A\s*(!\[[^\]]*\]\([^)]*\)|https?://\S+\.(?:png|jpe?g|gif|webp|svg)\S*)\s*\z`)

// canonicalTupleJSON は曖昧さのない tuple エンコード（JSON array compact）。
func canonicalTupleJSON(parts []any) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(parts)
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
}

// ChunkID は chunk 内容アドレス ID を導出する。
func ChunkID(articleID, articleHash string, ordinal int) string {
	return sha256Hex(canonicalTupleJSON([]any{articleID, articleHash, ChunkerVersion, ordinal}))
}

// ChunkCacheKey は embedding cache キーを導出する（内容が同じなら再利用可能）。
func ChunkCacheKey(model, taskType, embeddingInput string) string {
	return sha256Hex(canonicalTupleJSON([]any{model, taskType, ChunkerVersion, embeddingInput}))
}

func capRunes(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	count := 0
	for offset := range s {
		if count == maxRunes {
			return s[:offset]
		}
		count++
	}
	return s
}

// isSkippableChunkBody は空・画像単独のチャンク本文を判定する。
func isSkippableChunkBody(body string) bool {
	t := strings.TrimSpace(body)
	if t == "" {
		return true
	}
	return imageOnlyRe.MatchString(t)
}

var sentenceEndRunes = map[rune]bool{
	'。': true, '！': true, '？': true, '.': true, '!': true, '?': true, '；': true, ';': true,
}

// ChunkArticle は 1 記事を意味検索チャンク列へ分割する。
// 記事本文が 128KiB を超える場合は ErrCorpusTooLarge（索引経路の容量ゲート）。
// 実装は IterateArticleChunks（canonical streaming chunker）への委譲。
func ChunkArticle(article CorpusArticle) ([]Chunk, error) {
	if len(article.Body) > articleBodyMaxBytes {
		return nil, ErrCorpusTooLarge
	}
	var chunks []Chunk
	if err := IterateArticleChunks(context.Background(), article, func(c Chunk) bool {
		chunks = append(chunks, c)
		return true
	}); err != nil {
		return nil, err
	}
	return chunks, nil
}

// ChunkCorpus はコーパス全体のチャンク列を返す。合計 2MiB / 2048 chunks /
// 1記事128KiB のいずれかを超えたら ErrCorpusTooLarge（部分索引化しない）。
// ChunkCorpusContext は ctx 打ち切り付きの ChunkCorpus。容量ゲートは完全維持。
func ChunkCorpusContext(ctx context.Context, corpus Corpus) ([]Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	total := 0
	for _, a := range corpus.Articles {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(a.Body) > articleBodyMaxBytes {
			return nil, ErrCorpusTooLarge
		}
		total += len(a.Body)
	}
	if total > corpusBodyMaxBytes {
		return nil, ErrCorpusTooLarge
	}
	var out []Chunk
	for _, a := range corpus.Articles {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := IterateArticleChunks(ctx, a, func(c Chunk) bool {
			out = append(out, c)
			return len(out) <= CorpusMaxChunks && ctx.Err() == nil
		}); err != nil {
			return nil, err
		}
		if len(out) > CorpusMaxChunks {
			return nil, ErrCorpusTooLarge
		}
	}
	return out, nil
}

func ChunkCorpus(corpus Corpus) ([]Chunk, error) {
	return ChunkCorpusContext(context.Background(), corpus)
}
