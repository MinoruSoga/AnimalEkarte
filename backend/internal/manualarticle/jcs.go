package manualarticle

// jcs.go — RFC 8785 (JSON Canonicalization Scheme) serialization of the two
// fixed digest input objects, and the SHA-256 lower-hex digest projections.
//
// 採用済み設計（plane-EMR-502 §6 exact projection addendum）:
//   content_digest = SHA-256 lower-hex of UTF-8 JCS over
//     {"schema":"manual-article-content-v1","title":<exact>,"order_value":<num>,
//      "section":<exact>,"body_markdown":<exact>}
//   command_digest = SHA-256 lower-hex of UTF-8 JCS over
//     {"schema":"manual-article-edit-v1","article_id":<decimal string>,
//      "category":<enum>,"slug":<exact>,"content_digest":<64hex>}
// オブジェクト member は UTF-16 code unit の昇順（上記記述順は既に昇順）。
// 文字列は trim・NFC 正規化・改行正規化を一切しない。

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const (
	// manualArticleContentSchema is the fixed schema marker of the content digest object.
	manualArticleContentSchema = "manual-article-content-v1"
	// manualArticleEditSchema is the fixed schema marker of the edit command digest object.
	manualArticleEditSchema = "manual-article-edit-v1"
)

// jcsAppendString は RFC 8785 の文字列 literal を追記する。
// 入力文字列は strict parse 済みで valid UTF-8 が前提。lone surrogate は
// 受理しない（strict parse が raw byte 走査で先に拒否する）。
func jcsAppendString(sb *strings.Builder, s string) {
	const hexdigits = "0123456789abcdef"
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\b':
			sb.WriteString(`\b`)
		case '\f':
			sb.WriteString(`\f`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			if r < 0x20 {
				sb.WriteString(`\u00`)
				sb.WriteByte(hexdigits[byte(r)>>4])
				sb.WriteByte(hexdigits[byte(r)&0x0f])
			} else {
				sb.WriteRune(r)
			}
		}
	}
	sb.WriteByte('"')
}

// jcsManualContentObject は content digest 入力の JCS object を組み立てる。
// member 順: body_markdown < order_value < schema < section < title。
func jcsManualContentObject(title, orderText, section, body string) string {
	var sb strings.Builder
	sb.WriteString(`{"body_markdown":`)
	jcsAppendString(&sb, body)
	sb.WriteString(`,"order_value":`)
	sb.WriteString(orderText)
	sb.WriteString(`,"schema":`)
	jcsAppendString(&sb, manualArticleContentSchema)
	sb.WriteString(`,"section":`)
	jcsAppendString(&sb, section)
	sb.WriteString(`,"title":`)
	jcsAppendString(&sb, title)
	sb.WriteByte('}')
	return sb.String()
}

// jcsManualEditCommandObject は command digest 入力の JCS object を組み立てる。
// member 順: article_id < category < content_digest < schema < slug。
// article_id は canonical 正数十進文字列として JCS 文字列値に入る。
func jcsManualEditCommandObject(articleIDText, category, slug, contentDigestHex string) string {
	var sb strings.Builder
	sb.WriteString(`{"article_id":`)
	jcsAppendString(&sb, articleIDText)
	sb.WriteString(`,"category":`)
	jcsAppendString(&sb, category)
	sb.WriteString(`,"content_digest":`)
	jcsAppendString(&sb, contentDigestHex)
	sb.WriteString(`,"schema":`)
	jcsAppendString(&sb, manualArticleEditSchema)
	sb.WriteString(`,"slug":`)
	jcsAppendString(&sb, slug)
	sb.WriteByte('}')
	return sb.String()
}

func sha256LowerHex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// manualContentDigestHex は content 値から content_digest（64 lower-hex）を算出する。
func manualContentDigestHex(title, orderText, section, body string) string {
	return sha256LowerHex([]byte(jcsManualContentObject(title, orderText, section, body)))
}

// manualEditCommandDigestHex は 対象 identity + content_digest から
// command_digest（64 lower-hex）を算出する。
func manualEditCommandDigestHex(articleIDText, category, slug, contentDigestHex string) string {
	return sha256LowerHex([]byte(jcsManualEditCommandObject(articleIDText, category, slug, contentDigestHex)))
}
