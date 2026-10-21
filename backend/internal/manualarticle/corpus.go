package manualarticle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

// CorpusArticle はサポートチャット検索用のマニュアル記事スナップショット。
// FE manual-index.ts の ManualArticle に対応する決定的フィールドのみ持つ
// （order は canonical hash の対象外。並びは category→slug のバイト順）。
type CorpusArticle struct {
	Category   string   `json:"category"`
	Slug       string   `json:"slug"`
	Title      string   `json:"title"`
	Section    string   `json:"section"`
	Body       string   `json:"body"`
	Keywords   []string `json:"keywords"`
	KeywordsVI []string `json:"keywords_vi"`
	// Hash は canonical JSON（上記フィールドのみ・順序固定）の SHA-256 hex。
	Hash string `json:"-"`
}

// Corpus は権限に応じた実効マニュアルコーパス。Revision は canonical
// article array JSON の SHA-256 hex で、FE/BE が同じ正本から同一値を導く。
type Corpus struct {
	Articles []CorpusArticle
	Revision string
}

var frontmatterRe = regexp.MustCompile(`(?s)\A---\n(.*?)\n---\n(.*)\z`)

var frontmatterQuoteRe = regexp.MustCompile(`^["']|["']$`)

// frontmatterHasKey は frontmatter 内にその key が存在するかを返す
// （明示空値とキー不在を区別するための純粋ヘルパー）。
func frontmatterHasKey(raw, key string) bool {
	m := frontmatterRe.FindStringSubmatch(raw)
	if m == nil {
		return false
	}
	for _, line := range strings.Split(m[1], "\n") {
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		if strings.TrimSpace(line[:idx]) == key {
			return true
		}
	}
	return false
}

// ParseFrontmatter は frontend/src/lib/manual-index.ts parseFrontmatter と
// 同一の挙動: 先頭の --- ブロックを行単位で解析し、最初のコロンで key/value
// を分け、value を trim して前後の単一引用符を1つずつ剥がす。
// keywords / keywords_vi はカンマ（,・、）区切りで空要素を捨てる。
// 本文は frontmatter 直後の文字列をそのまま返す。
func ParseFrontmatter(raw string) (title, section string, keywords, keywordsVI []string, body string) {
	m := frontmatterRe.FindStringSubmatch(raw)
	if m == nil {
		return "", "", nil, nil, raw
	}
	fmBlock, b := m[1], m[2]
	for _, line := range strings.Split(fmBlock, "\n") {
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := frontmatterQuoteRe.ReplaceAllString(strings.TrimSpace(line[idx+1:]), "")
		switch key {
		case "title":
			title = value
		case "section":
			section = value
		case "keywords":
			keywords = splitKeywordList(value)
		case "keywords_vi":
			keywordsVI = splitKeywordList(value)
		}
	}
	return title, section, keywords, keywordsVI, b
}

func splitKeywordList(value string) []string {
	out := make([]string, 0)
	for _, k := range strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == '、' }) {
		if t := strings.TrimSpace(k); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// BaseCorpusArticleFromMarkdown は 1 本の Markdown 原稿からコーパス記事を作る。
// title 未指定は slug、section 未指定は「その他」（manual-index.ts build と同じ）。
func BaseCorpusArticleFromMarkdown(category, slug, raw string) CorpusArticle {
	title, section, keywords, keywordsVI, body := ParseFrontmatter(raw)
	// FE の ?? 意味論と一致: キー不在のみ既定値、明示空値は空のまま残す。
	if !frontmatterHasKey(raw, "title") {
		title = slug
	}
	if !frontmatterHasKey(raw, "section") {
		section = "その他"
	}
	return CorpusArticle{
		Category:   category,
		Slug:       slug,
		Title:      title,
		Section:    section,
		Body:       body,
		Keywords:   nonNilStrings(keywords),
		KeywordsVI: nonNilStrings(keywordsVI),
	}
}

func nonNilStrings(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// canonicalArticleJSON はフィールド順固定・空白なし・HTML/Unicode エスケープなしの
// 単一記事 JSON を返す（FE 側の canonical 実装と byte 一致すること）。
func canonicalArticleJSON(a CorpusArticle) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	payload := struct {
		Category   string   `json:"category"`
		Slug       string   `json:"slug"`
		Title      string   `json:"title"`
		Section    string   `json:"section"`
		Body       string   `json:"body"`
		Keywords   []string `json:"keywords"`
		KeywordsVI []string `json:"keywords_vi"`
	}{
		Category:   a.Category,
		Slug:       a.Slug,
		Title:      a.Title,
		Section:    a.Section,
		Body:       a.Body,
		Keywords:   nonNilStrings(a.Keywords),
		KeywordsVI: nonNilStrings(a.KeywordsVI),
	}
	// Marshal 不可能な値型のみで構成されるため Encode は失敗しない。
	_ = enc.Encode(&payload)
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n"))
}

// CanonicalCorpusJSON は canonical article array JSON を返す（revision 計算の入力）。
func CanonicalCorpusJSON(articles []CorpusArticle) []byte {
	var buf bytes.Buffer
	buf.WriteByte('[')
	for i, a := range articles {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.Write(canonicalArticleJSON(a))
	}
	buf.WriteByte(']')
	return buf.Bytes()
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ApplyCorpusOverrides は DB 公開版行（PublishedArticle = published_version_id の
// 内容）を base 記事へ適用する。base に存在する記事は title/section/body を
// 置き換え keywords は保持、DB のみの記事は keywords 空で追加する。
// 未公開 current・hidden は公開版に現れない。base 配列・map は変更しない（immutability）。
func ApplyCorpusOverrides(base []CorpusArticle, overrides []PublishedArticle) []CorpusArticle {
	merged := make([]CorpusArticle, 0, len(base)+len(overrides))
	index := make(map[string]int, len(base))
	for _, a := range base {
		index[a.Category+"/"+a.Slug] = len(merged)
		merged = append(merged, a)
	}
	for _, o := range overrides {
		key := string(o.Category) + "/" + o.Slug
		if i, ok := index[key]; ok {
			a := merged[i]
			a.Title = o.Title
			a.Section = o.Section
			a.Body = o.BodyMarkdown
			merged[i] = a
			continue
		}
		index[key] = len(merged)
		merged = append(merged, CorpusArticle{
			Category:   string(o.Category),
			Slug:       o.Slug,
			Title:      o.Title,
			Section:    o.Section,
			Body:       o.BodyMarkdown,
			Keywords:   []string{},
			KeywordsVI: []string{},
		})
	}
	sortCorpusArticles(merged)
	return merged
}

func sortCorpusArticles(articles []CorpusArticle) {
	sort.SliceStable(articles, func(i, j int) bool {
		if articles[i].Category != articles[j].Category {
			return articles[i].Category < articles[j].Category
		}
		return articles[i].Slug < articles[j].Slug
	})
}

// BuildCorpus はソート済み記事列に per-article hash と全体 revision を付与する。
// base/overrides が nil のときも決定的（空配列・空配列の hash）。
func BuildCorpus(base []CorpusArticle, overrides []PublishedArticle) Corpus {
	articles := ApplyCorpusOverrides(base, overrides)
	for i := range articles {
		articles[i].Hash = sha256Hex(canonicalArticleJSON(articles[i]))
	}
	return Corpus{
		Articles: articles,
		Revision: sha256Hex(CanonicalCorpusJSON(articles)),
	}
}
