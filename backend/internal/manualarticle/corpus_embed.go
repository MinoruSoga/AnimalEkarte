package manualarticle

import (
	_ "embed"
	"encoding/json"
)

// corpusGeneratedJSON は cmd/manual-corpus（人間実行 codegen 経路）が生成する
// base コーパス。プレースホルダは revision 空文字で「利用不可」を明示する。
//
//go:embed corpus.generated.json
var corpusGeneratedJSON []byte

type corpusGeneratedFile struct {
	Articles []CorpusArticle `json:"articles"`
	Revision string          `json:"revision"`
}

// EmbeddedBaseCorpus は go:embed 済み base コーパスを返す。
// revision 未生成（プレースホルダ）の場合は ok=false — 意味検索は index_missing
// として縮退し、keyword 経路は Markdown 由来ではなく呼出側の責務で処理する。
// 埋め込み JSON が壊れている場合も ok=false（生成物の信頼境界を越えて解釈しない）。
func EmbeddedBaseCorpus() (articles []CorpusArticle, revision string, ok bool) {
	return decodeEmbeddedCorpus(corpusGeneratedJSON)
}

// decodeEmbeddedCorpus は埋め込み JSON を厳格検証する。revision 空は未生成
// （fail closed）。格納 revision を記事配列から再計算して照合し、改竄された
// 生成物は採用しない。
func decodeEmbeddedCorpus(raw []byte) ([]CorpusArticle, string, bool) {
	var f corpusGeneratedFile
	if err := json.Unmarshal(raw, &f); err != nil || f.Revision == "" {
		return nil, "", false
	}
	for i := range f.Articles {
		f.Articles[i].Hash = sha256Hex(canonicalArticleJSON(f.Articles[i]))
	}
	if sha256Hex(CanonicalCorpusJSON(f.Articles)) != f.Revision {
		return nil, "", false
	}
	return f.Articles, f.Revision, true
}
