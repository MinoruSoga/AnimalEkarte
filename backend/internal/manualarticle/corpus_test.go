package manualarticle

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCanonicalJSON_EscapesU2028StandardAndPreservesLiteralBackslash(t *testing.T) {
	body := "a\u2028b" + "\\u2028" + "<>&😀é\u2029"
	a := CorpusArticle{Category: "s", Slug: "x", Title: "t", Section: "sec", Body: body, Keywords: []string{}, KeywordsVI: []string{}}
	j := canonicalArticleJSON(a)
	if !strings.Contains(string(j), `\u2028`) {
		t.Fatal("U+2028 must stay Go-standard escaped \\u2028")
	}
	if !strings.Contains(string(j), `\\u2028`) {
		t.Fatal("literal backslash-u2028 must remain double-escaped")
	}
	var rt struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(j, &rt); err != nil || rt.Body != body {
		t.Fatalf("roundtrip failed: %v", err)
	}
}

func TestFrontmatterExplicitEmptyVsAbsent(t *testing.T) {
	title, _, _, _, _ := ParseFrontmatter("---\ntitle: \"\"\n---\nbody")
	if title != "" {
		t.Fatal("explicit empty title must stay empty (FE ?? semantics)")
	}
	title, _, _, _, body := ParseFrontmatter("no fm")
	if title != "" || body != "no fm" {
		t.Fatal("absent frontmatter")
	}
	a := BaseCorpusArticleFromMarkdown("screens", "myslug", "---\n---\nx")
	if a.Title != "myslug" {
		t.Fatal("absent title defaults slug")
	}
}
