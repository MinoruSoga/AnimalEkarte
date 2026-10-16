package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/support"
)

func TestIndexGateBlocksBeforeDB(t *testing.T) {
	called := 0
	saved := openDB
	defer func() { openDB = saved }()
	openDB = func(*config.Config) (*gorm.DB, error) { called++; return nil, nil }
	if err := run(context.Background(), &config.Config{}); err == nil {
		t.Fatal("unapproved must fail")
	}
	if called != 0 {
		t.Fatal("db opener must not run when approval missing")
	}
}

// validateIndexConfig は GLM(SupportLLM)設定とは独立して SupportEmbedding を要求する。
func TestValidateIndexConfig(t *testing.T) {
	valid := &config.Config{
		SupportExternalDataApproved: true,
		SupportEmbeddingBaseURL:     "https://api.sarashina.ai/v1",
		SupportEmbeddingAPIKey:      "dummy-key",
		SupportEmbeddingModel:       "sarashina3-embedding",
		SupportEmbeddingTimeoutMS:   1200,
	}
	if err := validateIndexConfig(valid); err != nil {
		t.Fatalf("valid embedding config must pass: %v", err)
	}
	// GLM 設定があっても embedding 設定が無ければ拒否。
	glmOnly := &config.Config{
		SupportExternalDataApproved: true,
		SupportLLMBaseURL:           "https://api.z.ai/api/paas/v4",
		SupportLLMAPIKey:            "generation-only-test-key",
		SupportLLMModel:             "glm-4.5-flash",
	}
	if err := validateIndexConfig(glmOnly); err == nil {
		t.Fatal("GLM-only config must fail (embedding credentials separate)")
	}
	for name, cfg := range map[string]*config.Config{
		"nil":          nil,
		"no_approval":  {SupportEmbeddingBaseURL: "https://api.sarashina.ai/v1", SupportEmbeddingAPIKey: "k", SupportEmbeddingModel: "m", SupportEmbeddingTimeoutMS: 1200},
		"no_key":       {SupportExternalDataApproved: true, SupportEmbeddingBaseURL: "https://api.sarashina.ai/v1", SupportEmbeddingModel: "m", SupportEmbeddingTimeoutMS: 1200},
		"bad_endpoint": {SupportExternalDataApproved: true, SupportEmbeddingBaseURL: "https://api.openai.com/v1", SupportEmbeddingAPIKey: "k", SupportEmbeddingModel: "m", SupportEmbeddingTimeoutMS: 1200},
		"no_model":     {SupportExternalDataApproved: true, SupportEmbeddingBaseURL: "https://api.sarashina.ai/v1", SupportEmbeddingAPIKey: "k", SupportEmbeddingTimeoutMS: 1200},
		"bad_timeout":  {SupportExternalDataApproved: true, SupportEmbeddingBaseURL: "https://api.sarashina.ai/v1", SupportEmbeddingAPIKey: "k", SupportEmbeddingModel: "m"},
	} {
		if err := validateIndexConfig(cfg); err == nil {
			t.Fatalf("%s must fail", name)
		}
	}
}

// writeIndexFailure は部分集計だけを出し、原因詳細を漏らさない。
func TestWriteIndexFailureClassification(t *testing.T) {
	var buf bytes.Buffer
	if err := writeIndexFailure(&buf, &indexRunFailure{
		result: support.IndexResult{Chunks: 2, Existing: 1, Stored: 1},
		cause:  errors.New("SECRET_URL_PATIENT_TEXT"),
	}); err != nil {
		t.Fatalf("writeIndexFailure: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output must be JSON: %v", err)
	}
	if out["status"] != "failed" || out["reason"] != "index_failed" {
		t.Fatalf("classification missing: %v", out)
	}
	if out["chunks"] != float64(2) || out["stored"] != float64(1) {
		t.Fatalf("partial counts missing: %v", out)
	}
	if bytes.Contains(buf.Bytes(), []byte("SECRET")) {
		t.Fatal("cause details must not leak")
	}
	var buf2 bytes.Buffer
	if err := writeIndexFailure(&buf2, errors.New("unrelated")); err != nil {
		t.Fatalf("writeIndexFailure: %v", err)
	}
	var out2 map[string]any
	if err := json.Unmarshal(buf2.Bytes(), &out2); err != nil {
		t.Fatal(err)
	}
	if out2["chunks"] != float64(0) {
		t.Fatal("generic error must report zero counts")
	}
}

type failingWriter struct {
	err       error
	attempted []byte
}

func (w *failingWriter) Write(p []byte) (int, error) {
	w.attempted = append(w.attempted, p...)
	return 0, w.err
}

// writer 失敗は呼び出し側へ伝播し、payload は依然として集計のみ。
func TestWriteIndexFailureWriterError(t *testing.T) {
	sentinel := errors.New("writer broken")
	for name, cause := range map[string]error{
		"indexRunFailure": &indexRunFailure{
			result: support.IndexResult{Chunks: 2, Existing: 1, Stored: 1},
			cause:  errors.New("SECRET_URL_PATIENT_TEXT"),
		},
		"generic": errors.New("SECRET_URL_PATIENT_TEXT"),
	} {
		w := &failingWriter{err: sentinel}
		err := writeIndexFailure(w, cause)
		if !errors.Is(err, sentinel) {
			t.Fatalf("%s: writer error must propagate, got %v", name, err)
		}
		var out map[string]any
		if jerr := json.Unmarshal(w.attempted, &out); jerr != nil {
			t.Fatalf("%s: attempted payload must be JSON: %v", name, jerr)
		}
		if out["status"] != "failed" || out["reason"] != "index_failed" {
			t.Fatalf("%s: attempted payload missing safe classification: %v", name, out)
		}
		if bytes.Contains(w.attempted, []byte("SECRET")) {
			t.Fatalf("%s: attempted payload must not contain the cause", name)
		}
	}
}
