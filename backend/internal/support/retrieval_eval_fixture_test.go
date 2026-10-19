package support

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/animal-ekarte/backend/internal/manualarticle"
)

type retrievalEvalEvidence struct {
	AllowedArticleIDs []string `json:"allowed_article_ids"`
	Span              string   `json:"span"`
	SpanSHA256        string   `json:"span_sha256"`
	RequiredFacts     []string `json:"required_facts"`
	ForbiddenFacts    []string `json:"forbidden_facts"`
}

type retrievalEvalCase struct {
	ID               string   `json:"id"`
	Category         string   `json:"category"`
	Language         string   `json:"language"`
	Question         string   `json:"question"`
	EvidenceID       string   `json:"evidence_id"`
	ExpectedBehavior string   `json:"expected_behavior"`
	ForbiddenFacts   []string `json:"forbidden_facts"`
	PriorExchange    *struct {
		Question string   `json:"question"`
		Reply    string   `json:"reply"`
		Sources  []string `json:"sources"`
	} `json:"prior_exchange"`
}

type retrievalEvalFixture struct {
	SchemaVersion int    `json:"schema_version"`
	Ticket        string `json:"ticket"`
	ReviewStatus  string `json:"review_status"`
	Grading       struct {
		Variants               []string `json:"variants"`
		QuestionsPerVariant    int      `json:"questions_per_variant"`
		PositiveDenominator    int      `json:"positive_denominator"`
		ControlDenominator     int      `json:"control_denominator"`
		PositivePerLanguage    int      `json:"positive_per_language"`
		TailDenominator        int      `json:"tail_denominator"`
		RetrievalMinimumHits   int      `json:"retrieval_minimum_hits"`
		RetrievalLanguageHits  int      `json:"retrieval_minimum_hits_per_language"`
		RetrievalTailHits      int      `json:"retrieval_minimum_tail_hits"`
		RetrievalNonRegression bool     `json:"retrieval_must_not_regress_against_current_search"`
		AnswerMinimumPasses    int      `json:"answer_minimum_passes"`
		AnswerLanguagePasses   int      `json:"answer_minimum_passes_per_language"`
		ControlRequiredPasses  int      `json:"control_required_passes"`
	} `json:"grading"`
	Evidence map[string]retrievalEvalEvidence `json:"evidence"`
	Cases    []retrievalEvalCase              `json:"cases"`
}

// TestRetrievalEvalFixture checks label integrity, not model quality. It does not
// execute a model, approve external transmission, or mark a label review complete.
func TestRetrievalEvalFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/retrieval-eval-cases.json")
	require.NoError(t, err)
	var fixture retrievalEvalFixture
	require.NoError(t, json.Unmarshal(raw, &fixture))
	require.Equal(t, 1, fixture.SchemaVersion)
	require.Equal(t, "EMR-410", fixture.Ticket)
	require.Contains(t, []string{"pending_independent_review", "reviewed"}, fixture.ReviewStatus)
	require.Len(t, fixture.Cases, 80)
	require.NotEmpty(t, fixture.Evidence)
	require.ElementsMatch(t, []string{"current_search_glm", "current_search_nano", "hybrid_nano"}, fixture.Grading.Variants)
	require.Equal(t, 80, fixture.Grading.QuestionsPerVariant)
	require.Equal(t, 60, fixture.Grading.PositiveDenominator)
	require.Equal(t, 20, fixture.Grading.ControlDenominator)
	require.Equal(t, 30, fixture.Grading.PositivePerLanguage)
	require.Equal(t, 10, fixture.Grading.TailDenominator)
	require.Equal(t, 54, fixture.Grading.RetrievalMinimumHits)
	require.Equal(t, 27, fixture.Grading.RetrievalLanguageHits)
	require.Equal(t, 9, fixture.Grading.RetrievalTailHits)
	require.True(t, fixture.Grading.RetrievalNonRegression)
	require.Equal(t, 54, fixture.Grading.AnswerMinimumPasses)
	require.Equal(t, 27, fixture.Grading.AnswerLanguagePasses)
	require.Equal(t, 20, fixture.Grading.ControlRequiredPasses)

	positiveCounts := map[string]int{"formal": 0, "paraphrase": 0, "article_tail": 0, "conversation_ellipsis": 0}
	controlCounts := map[string]int{"not_covered": 0, "medical_refusal": 0, "instruction_refusal": 0, "weak_hit": 0}
	languages := map[string]map[string]int{}
	controlLanguages := map[string]int{}
	seenIDs, seenQuestions, usedEvidence := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for i, c := range fixture.Cases {
		require.NotEmpty(t, strings.TrimSpace(c.ID), "case %d", i)
		require.False(t, seenIDs[c.ID], "duplicate case ID at %d", i)
		require.False(t, seenQuestions[c.Question], "duplicate question at %d", i)
		seenIDs[c.ID], seenQuestions[c.Question] = true, true
		require.NotEmpty(t, strings.TrimSpace(c.Question), "case %d", i)
		require.LessOrEqual(t, len(c.Question), 2000, "question UTF-8 bytes at %d", i)
		require.Contains(t, []string{"ja", "vi"}, c.Language, "case %d", i)
		if c.EvidenceID == "" {
			require.Contains(t, controlCounts, c.Category, "control %d", i)
			require.NotEmpty(t, strings.TrimSpace(c.ExpectedBehavior), "control %d", i)
			require.NotEmpty(t, c.ForbiddenFacts, "control %d", i)
			require.Nil(t, c.PriorExchange, "control %d", i)
			controlCounts[c.Category]++
			controlLanguages[c.Language]++
			continue
		}
		require.Contains(t, positiveCounts, c.Category, "positive %d", i)
		evidence, ok := fixture.Evidence[c.EvidenceID]
		require.True(t, ok, "unknown evidence for case %d", i)
		require.Empty(t, c.ExpectedBehavior, "positive %d must inherit evidence labels", i)
		require.Empty(t, c.ForbiddenFacts, "positive %d must inherit evidence labels", i)
		usedEvidence[c.EvidenceID] = true
		positiveCounts[c.Category]++
		if languages[c.Category] == nil {
			languages[c.Category] = map[string]int{}
		}
		languages[c.Category][c.Language]++
		if c.Category == "conversation_ellipsis" {
			require.NotNil(t, c.PriorExchange, "conversation %d", i)
			require.NotEmpty(t, c.PriorExchange.Question, "conversation %d", i)
			require.NotEmpty(t, c.PriorExchange.Reply, "conversation %d", i)
			require.NotEmpty(t, c.PriorExchange.Sources, "conversation %d", i)
			require.LessOrEqual(t, len(c.PriorExchange.Sources), 3)
			for _, source := range c.PriorExchange.Sources {
				require.Contains(t, evidence.AllowedArticleIDs, source, "conversation %d", i)
			}
		} else {
			require.Nil(t, c.PriorExchange, "non-conversation %d", i)
		}
	}
	require.Equal(t, map[string]int{"formal": 20, "paraphrase": 20, "article_tail": 10, "conversation_ellipsis": 10}, positiveCounts)
	require.Equal(t, map[string]int{"not_covered": 5, "medical_refusal": 5, "instruction_refusal": 5, "weak_hit": 5}, controlCounts)
	require.Equal(t, map[string]int{"ja": 10, "vi": 10}, controlLanguages)
	for category, count := range positiveCounts {
		require.Equal(t, map[string]int{"ja": count / 2, "vi": count / 2}, languages[category], category)
	}

	base, _, available := manualarticle.EmbeddedBaseCorpus()
	require.True(t, available, "generated corpus is required; make codegen remains a human execution boundary")
	articles := map[string]manualarticle.CorpusArticle{}
	for _, article := range base {
		articles[article.Category+"/"+article.Slug] = article
	}
	idPattern := regexp.MustCompile(`^(screens|workflows)/[A-Za-z0-9][A-Za-z0-9_-]{0,99}$`)
	for id, evidence := range fixture.Evidence {
		require.True(t, usedEvidence[id], "unused evidence label")
		require.NotEmpty(t, evidence.AllowedArticleIDs, id)
		require.NotEmpty(t, strings.TrimSpace(evidence.Span), id)
		require.NotEmpty(t, evidence.RequiredFacts, id)
		require.NotEmpty(t, evidence.ForbiddenFacts, id)
		for _, fact := range append(append([]string{}, evidence.RequiredFacts...), evidence.ForbiddenFacts...) {
			require.NotEmpty(t, strings.TrimSpace(fact), id)
		}
		hash := sha256.Sum256([]byte(evidence.Span))
		require.Equal(t, hex.EncodeToString(hash[:]), evidence.SpanSHA256, id)
		seenArticles := map[string]bool{}
		for _, articleID := range evidence.AllowedArticleIDs {
			require.Regexp(t, idPattern, articleID)
			require.False(t, seenArticles[articleID], id)
			seenArticles[articleID] = true
			article, ok := articles[articleID]
			require.True(t, ok, id)
			require.Equal(t, 1, strings.Count(article.Body, evidence.Span), "evidence %s must match a unique source span", id)
		}
	}
	t.Logf("fixture integrity only: 80 cases, 60 positive, 20 controls; label review=%s; API quality NOT evaluated", fixture.ReviewStatus)
}
