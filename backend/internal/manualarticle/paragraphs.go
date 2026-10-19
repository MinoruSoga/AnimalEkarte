package manualarticle

import (
	"context"
	"strings"
)

// paragraphs.go — 意味検索容量超過時の keyword fallback 用の
// キャンセル可能な段落イテレータ。メモリにチャンク全体を持たず、
// ctx を毎段落チェックする。意味検索の容量ゲートは変更しない。

// IterateParagraphs は記事本文（frontmatter 除去・フェンス内 # 非見出し）を
// 段落単位で fn へ渡す。fn が false または ctx 終了で打ち切る。
func IterateParagraphs(ctx context.Context, article CorpusArticle, fn func(heading, paragraph string) bool) {
	// EMR-438: Body は正規化済み本文のため再パースしない（chunk_iter.go の EMR-438 参照）。
	body := article.Body
	var headingStack []string
	var paragraph strings.Builder
	fenceMarker := ""
	flush := func() bool {
		p := paragraph.String()
		paragraph.Reset()
		if strings.TrimSpace(p) == "" {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		default:
		}
		return fn(strings.Join(headingStack, " > "), strings.TrimRight(p, "\n"))
	}
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if fenceMarker == "" && (strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")) {
			fenceMarker = trimmed[:3]
			paragraph.WriteString(line)
			paragraph.WriteByte('\n')
			continue
		}
		if fenceMarker != "" {
			paragraph.WriteString(line)
			paragraph.WriteByte('\n')
			if strings.HasPrefix(trimmed, fenceMarker) {
				fenceMarker = ""
			}
			continue
		}
		if m := headingRe.FindStringSubmatch(line); m != nil {
			if !flush() {
				return
			}
			level := len(m[1])
			if level-1 < len(headingStack) {
				headingStack = headingStack[:level-1]
			}
			for len(headingStack) < level-1 {
				headingStack = append(headingStack, "")
			}
			headingStack = append(headingStack, strings.TrimSpace(m[2]))
			continue
		}
		if trimmed == "" {
			if !flush() {
				return
			}
			continue
		}
		paragraph.WriteString(line)
		paragraph.WriteByte('\n')
	}
	_ = flush()
}
