package support

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// llmResponseMaxBytes は上流レスポンスの読み取り上限（異常に大きな応答からの防衛）
const llmResponseMaxBytes = 1 << 20

// llmErrorBodyDiscardBytes は非 2xx 時にボディを捨てるための読み取り上限
const llmErrorBodyDiscardBytes = 64 << 10

// ChatMessage は OpenAI 互換チャット API のメッセージ
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompleter は LLM へのチャット完了要求を抽象化する。
// nil の場合はチャット機能が無効（フロントエンドは検索型ヘルプへフォールバック）。
type ChatCompleter interface {
	Complete(ctx context.Context, messages []ChatMessage) (string, error)
}

// upstreamError は LLM プロバイダの非 2xx 応答を表す。
type upstreamError struct {
	status int
}

func (e *upstreamError) Error() string {
	return fmt.Sprintf("llm upstream returned status %d", e.status)
}

var errUpstreamEmpty = errors.New("llm upstream returned empty content")

// openAICompatibleClient は OpenAI 互換 /chat/completions エンドポイント用クライアント。
// xAI(api.x.ai/v1)や Gemini の OpenAI 互換 API も baseURL 差し替えで利用できる。
type openAICompatibleClient struct {
	baseURL    string
	apiKey     string
	model      string
	maxTokens  int
	httpClient *http.Client
}

// NewChatCompleter は OpenAI 互換クライアントを初期化する。
// apiKey が空の場合は nil を返す（機能無効）。
func NewChatCompleter(baseURL, apiKey, model string, timeout time.Duration) ChatCompleter {
	if strings.TrimSpace(apiKey) == "" {
		return nil
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return &openAICompatibleClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		model:      model,
		maxTokens:  chatMaxCompletionTokens,
		httpClient: &http.Client{Timeout: timeout},
	}
}

type chatCompletionRequest struct {
	Model               string        `json:"model"`
	Messages            []ChatMessage `json:"messages"`
	MaxCompletionTokens int           `json:"max_completion_tokens"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

// Complete は chat/completions を呼び出して回答テキストを返す。
// 非 2xx は *upstreamError で返し、ハンドラ側で 502 にマッピングする。
// temperature は指定しない（gpt-5 系の reasoning モデルは非デフォルト値を拒否するため）。
func (c *openAICompatibleClient) Complete(ctx context.Context, messages []ChatMessage) (string, error) {
	payload, err := json.Marshal(chatCompletionRequest{
		Model:               c.model,
		Messages:            messages,
		MaxCompletionTokens: c.maxTokens,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/chat/completions",
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf("failed to build llm request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // クローズ失敗は復旧不可

	if resp.StatusCode != http.StatusOK {
		// エラーボディにトークン等が含まれうるため読み捨てるのみ（ログ出力しない）
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, llmErrorBodyDiscardBytes))
		return "", &upstreamError{status: resp.StatusCode}
	}

	var parsed chatCompletionResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, llmResponseMaxBytes)).Decode(&parsed); err != nil {
		return "", fmt.Errorf("failed to decode llm response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", errUpstreamEmpty
	}
	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", errUpstreamEmpty
	}
	return content, nil
}
