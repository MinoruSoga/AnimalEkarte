package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/animal-ekarte/backend/internal/manualarticle"
)

// embedding.go — Sarashina embeddings API クライアント（R4）。
// 単一 string input・task_type 必須・1536 dim・非正規化 float を厳格検証し、
// L2 正規化したベクトルを返す。対話側では再試行しない（CLI のみ別途再試行）。

const (
	// embeddingResponseMaxBytes は応答ボディ上限（128KiB）。
	embeddingResponseMaxBytes = 128 << 10
	embeddingDimensions       = manualarticle.EmbeddingDimensions
)

// EmbeddingUnavailableError は意味検索を利用不可とする失敗分類
// （上流 4xx/5xx・timeout・不正応答・過大応答・依存欠落を含む）。
type EmbeddingUnavailableError struct {
	Status int
	Reason string
	Err    error
}

func (e *EmbeddingUnavailableError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("embedding unavailable (%s): %v", e.Reason, e.Err)
	}
	return fmt.Sprintf("embedding unavailable (%s)", e.Reason)
}
func (e *EmbeddingUnavailableError) Unwrap() error { return e.Err }

// ManualEmbedder は質問/文書の embedding を 1 件ずつ生成する。
type ManualEmbedder interface {
	Embed(ctx context.Context, taskType, input string) ([]float64, error)
}

type sarashinaEmbedder struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewSarashinaEmbedder は embeddings クライアントを初期化する。
// apiKey 空・baseURL 空・timeout<=0 は nil（=意味検索無効）を返す。
// リダイレクトは無効化する（承認外エンドポイントへの誘導を防ぐ）。
func NewSarashinaEmbedder(baseURL, apiKey, model string, timeout time.Duration) ManualEmbedder {
	if apiKey == "" || baseURL == "" || timeout <= 0 {
		return nil
	}
	return &sarashinaEmbedder{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

type embeddingRequestBody struct {
	Model    string `json:"model"`
	TaskType string `json:"task_type"`
	Input    string `json:"input"`
}

type embeddingResponseBody struct {
	Model string `json:"model"`
	Data  []struct {
		Index     *int              `json:"index"`
		Embedding []json.RawMessage `json:"embedding"`
	} `json:"data"`
}

// Embed は /embeddings を 1 回呼び、厳格検証済みの L2 正規化ベクトルを返す。
func (c *sarashinaEmbedder) Embed(ctx context.Context, taskType, input string) ([]float64, error) {
	if taskType != manualarticle.TaskTypeRetrievalQuery && taskType != manualarticle.TaskTypeRetrievalDocument {
		return nil, &EmbeddingUnavailableError{Reason: "task_type"}
	}
	if input == "" || len(input) > 2048 {
		return nil, &EmbeddingUnavailableError{Reason: "input"}
	}
	// クライアント境界で送信テキストを screen（既存パターン再利用）。
	if err := screenOutboundTexts(input); err != nil {
		return nil, &EmbeddingUnavailableError{Reason: "screening", Err: err}
	}
	payload, err := json.Marshal(embeddingRequestBody{Model: c.model, TaskType: taskType, Input: input})
	if err != nil {
		return nil, &EmbeddingUnavailableError{Reason: "marshal", Err: err}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embeddings", bytes.NewReader(payload))
	if err != nil {
		return nil, &EmbeddingUnavailableError{Reason: "request", Err: err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &EmbeddingUnavailableError{Reason: "transport", Err: err}
	}
	defer resp.Body.Close() //nolint:errcheck // close failure cannot recover the consumed upstream response

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, llmErrorBodyDiscardBytes))
		return nil, &EmbeddingUnavailableError{Status: resp.StatusCode, Reason: "status"}
	}

	// 応答上限 128KiB+1 byte で検出（超過は malformed として失敗）。
	raw, err := io.ReadAll(io.LimitReader(resp.Body, embeddingResponseMaxBytes+1))
	if err != nil {
		return nil, &EmbeddingUnavailableError{Reason: "read", Err: err}
	}
	if len(raw) > embeddingResponseMaxBytes {
		return nil, &EmbeddingUnavailableError{Reason: "response_too_large"}
	}
	var parsed embeddingResponseBody
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&parsed); err != nil {
		return nil, &EmbeddingUnavailableError{Reason: "decode", Err: err}
	}
	// trailing data は malformed として失敗（厳密 1 JSON）。
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, &EmbeddingUnavailableError{Reason: "trailing_data"}
	}
	if len(parsed.Data) != 1 {
		return nil, &EmbeddingUnavailableError{Reason: "data_count"}
	}
	item := parsed.Data[0]
	if item.Index == nil || *item.Index != 0 {
		return nil, &EmbeddingUnavailableError{Reason: "index"}
	}
	if parsed.Model != c.model {
		return nil, &EmbeddingUnavailableError{Reason: "model_mismatch"}
	}
	if len(item.Embedding) != embeddingDimensions {
		return nil, &EmbeddingUnavailableError{Reason: "dimensions"}
	}
	vec := make([]float64, embeddingDimensions)
	for i, raw := range item.Embedding {
		var v *float64
		if err := json.Unmarshal(raw, &v); err != nil || v == nil {
			return nil, &EmbeddingUnavailableError{Reason: "non_finite"}
		}
		vec[i] = *v
	}
	var norm float64
	for _, v := range vec {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, &EmbeddingUnavailableError{Reason: "non_finite"}
		}
		norm += v * v
	}
	norm = math.Sqrt(norm)
	if !(norm > 0) || math.IsInf(norm, 0) {
		return nil, &EmbeddingUnavailableError{Reason: "zero_norm"}
	}
	for i := range vec {
		vec[i] /= norm
	}
	return vec, nil
}

// CosineSimilarity は L2 正規化済みベクトル同士の内積を返し [-1,1] へ clamp する。
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot float64
	for i := range a {
		dot += a[i] * b[i]
	}
	if dot > 1 {
		return 1
	}
	if dot < -1 {
		return -1
	}
	return dot
}
