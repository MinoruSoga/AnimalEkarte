package medicalrecord

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	labMappingLLMModel        = "gpt-6-luna"
	labMappingCandidateCap    = 200
	labMappingLLMTimeout      = 20 * time.Second
	labMappingLLMUnconfigured = "llm_unconfigured"
	labMappingLLMHTTP         = "llm_http"
	labMappingLLMParse        = "llm_parse"
	labMappingUnlistedField   = "unlisted_field"
	labMappingNoChartFields   = "no_chart_fields"
)

// LabMappingFieldCandidate is one clinic chart field the model may choose.
type LabMappingFieldCandidate struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Unit string `json:"unit"`
}

// LabMappingLLMRequest is one unmapped item. It has no measured value and no pet name.
type LabMappingLLMRequest struct {
	SourceType string                     `json:"source_type"`
	Species    string                     `json:"species"`
	Code       string                     `json:"code"`
	DeviceUnit string                     `json:"device_unit"`
	Candidates []LabMappingFieldCandidate `json:"candidates"`
}

// LabMappingLLMResult is the request that was built and, on success, the response body.
// ErrorCategory is empty when ExamTypeFieldID may be used.
type LabMappingLLMResult struct {
	ExamTypeFieldID uint64
	RequestJSON     json.RawMessage
	ResponseJSON    json.RawMessage
	ErrorCategory   string
}

// LabItemDrafter proposes one chart field for one device code.
type LabItemDrafter interface {
	DraftItem(ctx context.Context, req LabMappingLLMRequest) (LabMappingLLMResult, error)
}

type openAILabItemDrafter struct {
	client  *http.Client
	apiKey  string
	baseURL string
}

// NewOpenAILabItemDrafter calls the OpenAI Responses API.
// An empty apiKey does not send HTTP. The request JSON is still returned.
func NewOpenAILabItemDrafter(client *http.Client, apiKey, baseURL string) LabItemDrafter {
	if client == nil {
		client = newLabMappingHTTPClient()
	}
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	return &openAILabItemDrafter{client: client, apiKey: strings.TrimSpace(apiKey), baseURL: base}
}

// NewOpenAILabItemDrafterFromEnv reads OPENAI_API_KEY and OPENAI_BASE_URL.
// The key is not logged. An empty key leaves items blank.
func NewOpenAILabItemDrafterFromEnv() LabItemDrafter {
	base := os.Getenv("OPENAI_BASE_URL")
	if strings.TrimSpace(base) == "" {
		base = "https://api.openai.com/v1"
	}
	return NewOpenAILabItemDrafter(newLabMappingHTTPClient(), os.Getenv("OPENAI_API_KEY"), base)
}

func newLabMappingHTTPClient() *http.Client {
	return &http.Client{
		Timeout: labMappingLLMTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

type labMappingLLMBody struct {
	Model        string          `json:"model"`
	Store        bool            `json:"store"`
	Reasoning    labMappingThink `json:"reasoning"`
	Instructions string          `json:"instructions"`
	Input        string          `json:"input"`
	Text         labMappingText  `json:"text"`
}

type labMappingThink struct {
	Effort string `json:"effort"`
}

type labMappingText struct {
	Format labMappingFormat `json:"format"`
}

type labMappingFormat struct {
	Type   string          `json:"type"`
	Name   string          `json:"name"`
	Strict bool            `json:"strict"`
	Schema json.RawMessage `json:"schema"`
}

func (d *openAILabItemDrafter) DraftItem(ctx context.Context, req LabMappingLLMRequest) (LabMappingLLMResult, error) {
	if err := ctx.Err(); err != nil {
		return LabMappingLLMResult{}, err
	}
	body, err := labMappingLLMRequestBody(req)
	if err != nil {
		return LabMappingLLMResult{ErrorCategory: labMappingLLMParse}, nil
	}
	result := LabMappingLLMResult{RequestJSON: labMappingResponseJSON(body)}
	if d.apiKey == "" {
		result.ErrorCategory = labMappingLLMUnconfigured
		return result, nil
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, d.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		result.ErrorCategory = labMappingLLMHTTP
		return result, nil
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+d.apiKey)
	resp, err := d.client.Do(httpReq)
	if err != nil {
		if ctx.Err() != nil {
			return LabMappingLLMResult{}, ctx.Err()
		}
		result.ErrorCategory = labMappingLLMHTTP
		return result, nil
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		result.ErrorCategory = labMappingLLMHTTP
		return result, nil
	}
	result.ResponseJSON = labMappingResponseJSON(raw)
	if resp.StatusCode != http.StatusOK {
		result.ErrorCategory = labMappingLLMHTTP
		return result, nil
	}
	fieldID, ok := labMappingFieldIDFromResponse(raw)
	if !ok {
		result.ErrorCategory = labMappingLLMParse
		return result, nil
	}
	result.ExamTypeFieldID = fieldID
	return result, nil
}

func labMappingResponseJSON(raw []byte) json.RawMessage {
	// jsonb rejects NUL and unpaired surrogate escapes even in syntactically valid JSON.
	if bytes.Contains(raw, []byte(`\u0000`)) || bytes.IndexByte(raw, 0) >= 0 || !utf8.Valid(raw) ||
		(json.Valid(raw) && labMappingHasUnpairedSurrogate(raw)) {
		wrapped, _ := json.Marshal(struct {
			Encoding string `json:"encoding"`
			Body     []byte `json:"body"`
		}{Encoding: "base64", Body: raw})
		return wrapped
	}
	if json.Valid(raw) {
		return raw
	}
	wrapped, _ := json.Marshal(string(raw))
	return wrapped
}

// raw must be valid JSON; inspect escapes before decoding can replace lone surrogates.
func labMappingHasUnpairedSurrogate(raw []byte) bool {
	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		if raw[i+1] != 'u' {
			i++ // Skip the escaped character, including a literal backslash.
			continue
		}
		if raw[i+2]|0x20 == 'd' {
			switch raw[i+3] | 0x20 {
			case 'c', 'd', 'e', 'f':
				return true
			case '8', '9', 'a', 'b':
				if i+11 >= len(raw) || raw[i+6] != '\\' || raw[i+7] != 'u' ||
					raw[i+8]|0x20 != 'd' || raw[i+9]|0x20 < 'c' || raw[i+9]|0x20 > 'f' {
					return true
				}
				i += 6 // Consume the matching low-surrogate escape too.
			}
		}
		i += 5
	}
	return false
}

func labMappingLLMRequestBody(req LabMappingLLMRequest) (json.RawMessage, error) {
	input, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	schema := json.RawMessage(`{"type":"object","additionalProperties":false,"properties":{"exam_type_field_id":{"type":"integer"},"field_name":{"type":"string"}},"required":["exam_type_field_id","field_name"]}`)
	body := labMappingLLMBody{
		Model: labMappingLLMModel,
		Store: false,
		Reasoning: labMappingThink{
			Effort: "none",
		},
		Instructions: "Choose one candidate id for this device code. Return 0 when none match. Do not convert units.",
		Input:        string(input),
		Text: labMappingText{Format: labMappingFormat{
			Type:   "json_schema",
			Name:   "lab_item_mapping",
			Strict: true,
			Schema: schema,
		}},
	}
	return json.Marshal(body)
}

func labMappingFieldIDFromResponse(raw []byte) (uint64, bool) {
	text, ok := labMappingOutputText(raw)
	if !ok || text == "" {
		return 0, false
	}
	var parsed struct {
		ExamTypeFieldID uint64 `json:"exam_type_field_id"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return 0, false
	}
	return parsed.ExamTypeFieldID, true
}

func labMappingOutputText(raw []byte) (string, bool) {
	var direct struct {
		OutputText string `json:"output_text"`
		Output     []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(raw, &direct); err != nil {
		return "", false
	}
	if direct.OutputText != "" {
		return direct.OutputText, true
	}
	for _, item := range direct.Output {
		for _, content := range item.Content {
			if content.Text != "" && (content.Type == "" || content.Type == "output_text") {
				return content.Text, true
			}
		}
	}
	return "", false
}
