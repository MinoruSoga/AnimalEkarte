package manualarticle

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestChunkReconstruction_NoLostRunes(t *testing.T) {
	body := strings.Repeat("甲乙丙丁", 350) + "末尾" // 1402 runes
	a := CorpusArticle{Category: "screens", Slug: "s1", Title: "t", Body: body, Hash: "h"}
	chunks, err := ChunkArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range chunks {
		if utf8.RuneCountInString(c.Body) > 400 {
			t.Fatalf("chunk ordinal %d len=%d >400", c.Ordinal, utf8.RuneCountInString(c.Body))
		}
	}
	// remove 40-rune overlap from each continuation and concatenate == input
	var b strings.Builder
	for i, c := range chunks {
		r := []rune(c.Body)
		if i > 0 {
			r = r[40:]
		}
		b.WriteString(string(r))
	}
	if b.String() != body {
		t.Fatalf("reconstruction mismatch: got %d runes want %d", utf8.RuneCountInString(b.String()), utf8.RuneCountInString(body))
	}
}

func TestChunkFencePreservesHashAndBlankLines(t *testing.T) {
	// EMR-438: Body は正規化済み本文（frontmatter は BaseCorpusArticleFromMarkdown のみで
	// 剥ぐ）。`---` で始まる有効な Markdown 本文は metadata 誤認識せずそのまま保持する。
	body := "---\ntitle: X\n---\npara1\n\n```\n# not heading\n\nline2\n```\nafter"
	a := CorpusArticle{Category: "screens", Slug: "s2", Title: "t", Body: body, Hash: "h"}
	chunks, err := ChunkArticle(a)
	if err != nil {
		t.Fatal(err)
	}
	var all strings.Builder
	for _, c := range chunks {
		all.WriteString(c.Body)
		all.WriteString("\n")
	}
	if !strings.Contains(all.String(), "# not heading") {
		t.Fatalf("heading inside fence dropped: %q", all.String())
	}
	if !strings.Contains(all.String(), "title: X") {
		t.Fatalf("valid markdown section dropped by frontmatter reparse: %q", all.String())
	}
}

func TestChunkIDDeterministicAndHashDependent(t *testing.T) {
	a := CorpusArticle{Category: "screens", Slug: "s", Title: "t", Body: "hello", Hash: "h1"}
	c1, _ := ChunkArticle(a)
	a.Hash = "h2"
	c2, _ := ChunkArticle(a)
	if len(c1) == 0 || len(c2) == 0 || c1[0].ID == c2[0].ID {
		t.Fatal("chunk ID must change with article hash")
	}
}

func TestChunkCorpusCapacity(t *testing.T) {
	a := CorpusArticle{Category: "s", Slug: "x", Title: "t", Body: strings.Repeat("a", (128<<10)+1)}
	c := Corpus{Articles: []CorpusArticle{a}}
	if _, err := ChunkCorpus(c); !errors.Is(err, ErrCorpusTooLarge) {
		t.Fatal("expected ErrCorpusTooLarge")
	}
}

func TestChunkEmptyAndImageOnlySkipped(t *testing.T) {
	a := CorpusArticle{Category: "s", Slug: "x", Title: "t", Body: "\n\n![img](x.png)\n\n"}
	chunks, _ := ChunkArticle(a)
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks, got %d", len(chunks))
	}
}

func TestChunkBoundaryLengths(t *testing.T) {
	for _, n := range []int{401, 760, 761, 800, 801, 1200} {
		body := strings.Repeat("あ", n)
		a := CorpusArticle{Category: "s", Slug: "x", Title: "t", Body: body, Hash: "h"}
		chunks, err := ChunkArticle(a)
		if err != nil {
			t.Fatal(err)
		}
		var b strings.Builder
		for i, c := range chunks {
			rc := utf8.RuneCountInString(c.Body)
			if rc > 400 {
				t.Fatalf("n=%d chunk %d len=%d", n, i, rc)
			}
			r := []rune(c.Body)
			if i > 0 {
				if rc <= 40 {
					t.Fatalf("n=%d continuation chunk too short", n)
				}
				r = r[40:]
			}
			b.WriteString(string(r))
		}
		if b.String() != body {
			t.Fatalf("n=%d reconstruction lost %d runes", n, n-utf8.RuneCountInString(b.String()))
		}
	}
}

func TestFenceMismatchedMarkerDoesNotClose(t *testing.T) {
	body := "p\n\n```\n# code heading\n~~~\nstill code\n```\nafter"
	a := CorpusArticle{Category: "s", Slug: "x", Title: "t", Body: body, Hash: "h"}
	chunks, _ := ChunkArticle(a)
	var all strings.Builder
	for _, c := range chunks {
		all.WriteString(c.Body)
		all.WriteString("\n")
	}
	for _, want := range []string{"# code heading", "still code", "after"} {
		if !strings.Contains(all.String(), want) {
			t.Fatalf("missing %q in %q", want, all.String())
		}
	}
}

func TestBaseArticleEmptyVsAbsentTitle(t *testing.T) {
	a := BaseCorpusArticleFromMarkdown("screens", "myslug", "---\ntitle: \"\"\nsection: \"\"\n---\nbody")
	if a.Title != "" || a.Section != "" {
		t.Fatalf("explicit empty must stay empty: %+v", a)
	}
	a = BaseCorpusArticleFromMarkdown("screens", "myslug", "---\n---\nbody")
	if a.Title != "myslug" || a.Section != "その他" {
		t.Fatalf("absent keys must default: %+v", a)
	}
}

func TestDecodeEmbeddedCorpusTamperRejected(t *testing.T) {
	base := []CorpusArticle{{Category: "s", Slug: "x", Title: "t", Section: "s", Body: "b", Keywords: []string{}, KeywordsVI: []string{}}}
	corpus := BuildCorpus(base, nil)
	good := `{"articles":` + string(CanonicalCorpusJSON(corpus.Articles)) + `,"revision":"` + corpus.Revision + `"}`
	if _, rev, ok := decodeEmbeddedCorpus([]byte(good)); !ok || rev != corpus.Revision {
		t.Fatal("valid corpus must decode")
	}
	tampered := strings.Replace(good, corpus.Revision, strings.Repeat("0", 64), 1)
	if _, _, ok := decodeEmbeddedCorpus([]byte(tampered)); ok {
		t.Fatal("tampered revision must be rejected")
	}
	if _, _, ok := decodeEmbeddedCorpus([]byte(`{"articles":[],"revision":""}`)); ok {
		t.Fatal("placeholder must be unavailable")
	}
}
