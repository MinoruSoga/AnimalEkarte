package manualarticle

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

// IterateArticleChunks shares one layout algorithm between semantic indexing and
// keyword fallback. Only the callers enforce corpus capacity; retained working
// text here is bounded by one chunk, apart from views into the original body.
// Returning false from emit stops normally. Cancellation always returns ctx.Err().
func IterateArticleChunks(ctx context.Context, article CorpusArticle, emit func(Chunk) bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !utf8.ValidString(article.Body) {
		return errors.New("invalid manual UTF-8")
	}
	_, _, _, _, body := ParseFrontmatter(article.Body)
	articleID := article.Category + "/" + article.Slug
	title := capRunes(article.Title, chunkTitleMaxRunes)
	var headings []string
	var pending, currentHeading string
	ordinal := 0
	emitBody := func(piece string) bool {
		if ctx.Err() != nil {
			return false
		}
		h := currentHeading
		keepGoing := emit(Chunk{
			ID: ChunkID(articleID, article.Hash, ordinal), ArticleID: articleID,
			Ordinal: ordinal, Heading: h, Body: piece,
			EmbeddingInput: title + "\n" + h + "\n" + piece,
		})
		ordinal++
		return keepGoing && ctx.Err() == nil
	}
	flushPending := func() bool {
		if pending == "" {
			return true
		}
		piece := pending
		pending = ""
		return emitBody(piece)
	}
	pushParagraph := func(paragraph string) bool {
		if ctx.Err() != nil {
			return false
		}
		if isSkippableChunkBody(paragraph) {
			return true
		}
		budget := embeddingInputMaxBytes - len(title) - len(currentHeading) - 2
		fits := func(text string) bool {
			return len(text) <= budget && utf8.RuneCountInString(text) <= chunkBodyMaxRunes
		}
		if pending != "" {
			if len(pending)+2+len(paragraph) <= budget &&
				utf8.RuneCountInString(pending)+2+utf8.RuneCountInString(paragraph) <= chunkBodyMaxRunes {
				pending += "\n\n" + paragraph
				return true
			}
			if !flushPending() {
				return false
			}
		}
		if fits(paragraph) {
			pending = paragraph
			return true
		}
		rest, overlap := paragraph, ""
		for rest != "" {
			if ctx.Err() != nil {
				return false
			}
			window := chunkPrefix(rest, chunkBodyMaxRunes-utf8.RuneCountInString(overlap), budget-len(overlap))
			piece := overlap + window
			if !emitBody(piece) {
				return false
			}
			rest = rest[len(window):]
			previous := []rune(piece) // At most 400 code points, never the unconsumed article.
			start := max(0, len(previous)-chunkContinuationOverlap)
			overlap = string(previous[start:])
		}
		return true
	}

	paragraphStart := -1
	flushParagraph := func(end int) bool {
		if paragraphStart < 0 {
			return true
		}
		paragraph := strings.TrimRight(body[paragraphStart:end], "\r\n")
		paragraphStart = -1
		return pushParagraph(paragraph)
	}
	var fence byte
	fenceLength := 0
	for offset := 0; offset < len(body); {
		if err := ctx.Err(); err != nil {
			return err
		}
		end := len(body)
		if relative := strings.IndexByte(body[offset:], '\n'); relative >= 0 {
			end = offset + relative
		}
		line := body[offset:end]
		next := end
		if next < len(body) {
			next++
		}
		trimmed := strings.TrimSpace(line)
		marker, markerLength := chunkFence(trimmed)
		switch {
		case fence != 0:
			if marker == fence && markerLength >= fenceLength && strings.TrimSpace(trimmed[markerLength:]) == "" {
				fence = 0
			}
		case marker != 0:
			fence, fenceLength = marker, markerLength
			if paragraphStart < 0 {
				paragraphStart = offset
			}
		default:
			if match := headingRe.FindStringSubmatch(line); match != nil {
				if !flushParagraph(offset) || !flushPending() {
					return ctx.Err()
				}
				level := len(match[1])
				if level-1 < len(headings) {
					headings = headings[:level-1]
				}
				for len(headings) < level-1 {
					headings = append(headings, "")
				}
				headings = append(headings, strings.TrimSpace(match[2]))
				var path []string
				for _, part := range headings {
					if part != "" {
						path = append(path, part)
					}
				}
				currentHeading = capRunes(strings.Join(path, " > "), chunkHeadingMaxRunes)
			} else if trimmed == "" {
				if !flushParagraph(offset) {
					return ctx.Err()
				}
			} else if paragraphStart < 0 {
				paragraphStart = offset
			}
		}
		offset = next
	}
	if !flushParagraph(len(body)) || !flushPending() {
		return ctx.Err()
	}
	return ctx.Err()
}

// chunkPrefix advances only through the bounded window. All offsets remain byte
// offsets, including newline and punctuation cuts in multibyte text.
func chunkPrefix(text string, maxRunes, maxBytes int) string {
	end, count := 0, 0
	for offset, r := range text {
		size := utf8.RuneLen(r)
		if count == maxRunes || offset+size > maxBytes {
			break
		}
		end = offset + size
		count++
	}
	if end == len(text) {
		return text
	}
	window := text[:end]
	if newline := strings.LastIndexByte(window, '\n'); newline >= 0 {
		return window[:newline+1]
	}
	lastSentence := 0
	for offset, r := range window {
		if sentenceEndRunes[r] {
			lastSentence = offset + utf8.RuneLen(r)
		}
	}
	if lastSentence != 0 {
		return window[:lastSentence]
	}
	return window
}

func chunkFence(line string) (byte, int) {
	if len(line) < 3 || (line[0] != '`' && line[0] != '~') {
		return 0, 0
	}
	length := 1
	for length < len(line) && line[length] == line[0] {
		length++
	}
	if length < 3 {
		return 0, 0
	}
	return line[0], length
}
