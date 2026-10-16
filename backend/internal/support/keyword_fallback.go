package support

import (
	"context"
	"sort"

	"github.com/animal-ekarte/backend/internal/manualarticle"
)

// rankKeywordArticles は容量超過 fallback 用の有界 keyword ランカー。
// IterateArticleChunks で逐次評価し、topK 件のみ保持（全チャンクをメモリに残さない）。
// スコア規則は rankKeywordChunks と同一（min(2,len(groups))、matched DESC→ordinal ASC）。
func rankKeywordArticles(ctx context.Context, articles []manualarticle.CorpusArticle, groups [][]string, topK int) ([]manualarticle.Chunk, error) {
	if len(groups) == 0 || topK <= 0 {
		return nil, nil
	}
	required := 2
	if len(groups) < 2 {
		required = len(groups)
	}
	type scored struct {
		chunk   manualarticle.Chunk
		matched int
		artIdx  int
	}
	best := []scored{}
	less := func(i, j int) bool {
		if best[i].matched != best[j].matched {
			return best[i].matched > best[j].matched
		}
		if best[i].chunk.Ordinal != best[j].chunk.Ordinal {
			return best[i].chunk.Ordinal < best[j].chunk.Ordinal
		}
		return best[i].artIdx < best[j].artIdx
	}
	consider := func(c manualarticle.Chunk, a manualarticle.CorpusArticle, ai int) bool {
		norm, fold := articleSearchHaystack(a.Title, a.Section, c.Heading+"\n"+c.Body, a.Keywords, a.KeywordsVI)
		matched := 0
		for _, g := range groups {
			if groupHasTextEvidence(norm, fold, g) {
				matched++
			}
		}
		if matched < required {
			return ctx.Err() == nil
		}
		best = append(best, scored{c, matched, ai})
		sort.SliceStable(best, less)
		if len(best) > topK {
			best = best[:topK]
		}
		return ctx.Err() == nil
	}
	for ai, a := range articles {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := manualarticle.IterateArticleChunks(ctx, a, func(c manualarticle.Chunk) bool {
			return consider(c, a, ai)
		}); err != nil {
			return nil, err
		}
	}
	out := make([]manualarticle.Chunk, len(best))
	for i, b := range best {
		out[i] = b.chunk
	}
	return out, nil
}
