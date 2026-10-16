package manualarticle

import (
	"context"
	"errors"
)

// ErrEmbeddedCorpusUnavailable は埋め込み base コーパスが未生成であることを示す
// （corpus.generated.json がプレースホルダの状態。人間実行 codegen 待ち）。
var ErrEmbeddedCorpusUnavailable = errors.New("embedded manual corpus is not generated")

// CorpusSource は権限に応じた実効コーパスを供給する。
// permitted=true のとき FindAll の最新 DB 行を base へ上書き適用し、
// DB 取得失敗は error（古い base で回答しない fail-closed）。
// permitted=false では DB を読まず base のみを返す。
type CorpusSource struct {
	svc ManualArticleService
}

// NewCorpusSource は CorpusSource を初期化する（svc は nil 可 = DB 無しで base のみ）。
func NewCorpusSource(svc ManualArticleService) *CorpusSource {
	return &CorpusSource{svc: svc}
}

// EffectiveCorpus は権限に応じた実効コーパスを返す。
func (s *CorpusSource) EffectiveCorpus(ctx context.Context, permitted bool) (Corpus, error) {
	base, _, ok := EmbeddedBaseCorpus()
	if !ok {
		return Corpus{}, ErrEmbeddedCorpusUnavailable
	}
	if !permitted {
		return BuildCorpus(base, nil), nil
	}
	if s.svc == nil {
		return Corpus{}, ErrEmbeddedCorpusUnavailable
	}
	rows, err := s.svc.FindAll(ctx)
	if err != nil {
		return Corpus{}, err
	}
	return BuildCorpus(base, rows), nil
}
