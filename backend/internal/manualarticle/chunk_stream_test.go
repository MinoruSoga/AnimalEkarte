package manualarticle

import (
	"context"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestChunkUnicodeNewlineReconstruction(t *testing.T) {
	body := strings.Repeat("甲", 350) + "\n" + strings.Repeat("乙", 100)
	chunks, err := ChunkArticle(CorpusArticle{Category: "screens", Slug: "stream", Title: "T", Body: body, Hash: "h"})
	require.NoError(t, err)
	require.Equal(t, body, reconstructContinuation(t, chunks))
}

func TestChunkPacksOnlyFittingParagraphs(t *testing.T) {
	chunks, err := ChunkArticle(CorpusArticle{Category: "screens", Slug: "stream", Body: "## A\nfirst\n\nsecond\n## B\nthird", Hash: "h"})
	require.NoError(t, err)
	require.Len(t, chunks, 2)
	require.Equal(t, "first\n\nsecond", chunks[0].Body)
	require.Equal(t, "A", chunks[0].Heading)
	require.Equal(t, "third", chunks[1].Body)
	require.Equal(t, "B", chunks[1].Heading)

	a, b := strings.Repeat("a", 300), strings.Repeat("b", 300)
	chunks, err = ChunkArticle(CorpusArticle{Category: "screens", Slug: "stream", Body: a + "\n\n" + b, Hash: "h"})
	require.NoError(t, err)
	require.Len(t, chunks, 2)
	require.Equal(t, a, chunks[0].Body)
	require.Equal(t, b, chunks[1].Body, "ordinary paragraph boundaries do not create continuation overlap")
}

func TestChunkEmbeddingByteSplitRetainsOverlap(t *testing.T) {
	body := strings.Repeat("😀", 400)
	chunks, err := ChunkArticle(CorpusArticle{
		Category: "screens", Slug: "stream", Title: strings.Repeat("😀", 64),
		Body: "# " + strings.Repeat("😀", 64) + "\n" + body, Hash: "h",
	})
	require.NoError(t, err)
	require.Greater(t, len(chunks), 1)
	require.Equal(t, body, reconstructContinuation(t, chunks))
}

func reconstructContinuation(t *testing.T, chunks []Chunk) string {
	t.Helper()
	var out strings.Builder
	var previous []rune
	for i, chunk := range chunks {
		require.True(t, utf8.ValidString(chunk.Body))
		require.LessOrEqual(t, utf8.RuneCountInString(chunk.Body), 400)
		require.LessOrEqual(t, len(chunk.EmbeddingInput), 2048)
		runes := []rune(chunk.Body)
		if i > 0 {
			overlap := min(40, len(previous))
			require.GreaterOrEqual(t, len(runes), overlap)
			require.Equal(t, string(previous[len(previous)-overlap:]), string(runes[:overlap]))
			out.WriteString(string(runes[overlap:]))
		} else {
			out.WriteString(chunk.Body)
		}
		previous = runes
	}
	return out.String()
}

func TestChunkIteratorCancellationAndStop(t *testing.T) {
	article := CorpusArticle{Category: "screens", Slug: "stream", Body: strings.Repeat("甲", 50000), Hash: "h"}
	for _, test := range []struct {
		name                 string
		before, during, stop bool
	}{
		{name: "already_cancelled", before: true},
		{name: "cancel_in_callback", during: true},
		{name: "stop_in_callback", stop: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if test.before {
				cancel()
			}
			calls := 0
			err := IterateArticleChunks(ctx, article, func(Chunk) bool {
				calls++
				if test.during {
					cancel()
				}
				return !test.stop
			})
			if test.before || test.during {
				require.ErrorIs(t, err, context.Canceled)
			} else {
				require.NoError(t, err)
			}
			if test.before {
				require.Zero(t, calls)
			} else {
				require.Equal(t, 1, calls)
			}
		})
	}
}
