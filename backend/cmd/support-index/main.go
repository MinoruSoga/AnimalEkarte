// cmd/support-index は base + 最新 DB 上書きコーパスの不足 cache_key を
// 逐次 embedding して support_manual_embeddings へ保存する索引 CLI（R3）。
// 外部 API/DB 呼出は外部データ承認・API key・Sarashina URL 検証後のみ。
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/animal-ekarte/backend/internal/config"
	"github.com/animal-ekarte/backend/internal/dbconn"
	"github.com/animal-ekarte/backend/internal/manualarticle"
	"github.com/animal-ekarte/backend/internal/support"
)

// 注入可能な opener（テストは config 境界だけを検証し、実 DB/network を呼ばない）。
var openDB = dbconn.OpenGORM

// validateIndexConfig は DB/network 到達前の純粋ゲート。key/URL を戻さない。
func validateIndexConfig(cfg *config.Config) error {
	if cfg == nil || !cfg.SupportExternalDataApproved {
		return fmt.Errorf("external data approval required")
	}
	if cfg.SupportEmbeddingAPIKey == "" {
		return fmt.Errorf("embedding API key required")
	}
	if cfg.SupportEmbeddingBaseURL != "https://api.sarashina.ai/v1" {
		return fmt.Errorf("embedding endpoint must be Sarashina")
	}
	if cfg.SupportEmbeddingModel == "" {
		return fmt.Errorf("embedding model required")
	}
	if cfg.SupportEmbeddingTimeoutMS <= 0 {
		return fmt.Errorf("embedding timeout must be positive")
	}
	return nil
}

func run(ctx context.Context, cfg *config.Config) error {
	if err := validateIndexConfig(cfg); err != nil {
		return err
	}
	base, _, ok := manualarticle.EmbeddedBaseCorpus()
	if !ok {
		return fmt.Errorf("embedded corpus unavailable (run cmd/manual-corpus first)")
	}
	db, err := openDB(cfg)
	if err != nil {
		return fmt.Errorf("db open failed")
	}
	if sqlDB, derr := db.DB(); derr == nil {
		defer func() {
			if closeErr := sqlDB.Close(); closeErr != nil {
				fmt.Fprintln(os.Stderr, `{"event":"support_index","status":"warning","reason":"db_close_failed"}`)
			}
		}()
	}
	overrides, err := manualarticle.New(db).FindAll(ctx)
	if err != nil {
		return fmt.Errorf("corpus load failed")
	}
	deps := support.IndexDeps{
		Embedder: support.NewSarashinaEmbedder(cfg.SupportEmbeddingBaseURL, cfg.SupportEmbeddingAPIKey, cfg.SupportEmbeddingModel, time.Duration(cfg.SupportEmbeddingTimeoutMS)*time.Millisecond),
		Cache:    support.NewManualEmbeddingCache(db, cfg.SupportEmbeddingModel),
		Budget:   support.NewLLMBudgetGranter(db),
		Model:    cfg.SupportEmbeddingModel,
	}
	res, err := support.RunIndex(ctx, base, overrides, deps)
	if err != nil {
		return &indexRunFailure{result: res, cause: err}
	}
	fmt.Fprintf(os.Stderr, `{"ok":true,"chunks":%d,"existing":%d,"stored":%d}`+"\n", res.Chunks, res.Existing, res.Stored)
	return nil
}

// indexRunFailure は RunIndex 失敗を部分集計とともに運ぶ。Error() は静的文言のみ。
type indexRunFailure struct {
	result support.IndexResult
	cause  error
}

func (e *indexRunFailure) Error() string { return "manual index execution failed" }
func (e *indexRunFailure) Unwrap() error { return e.cause }

// writeIndexFailure は失敗を構造化 JSON で stderr へ書く。
// 原因の詳細（URL/key/本文/DB エラー）は一切出力しない。
func writeIndexFailure(w io.Writer, err error) error {
	var failure *indexRunFailure
	var res support.IndexResult
	if errors.As(err, &failure) {
		res = failure.result
	}
	_, writeErr := fmt.Fprintf(w, `{"event":"support_index","status":"failed","reason":"index_failed","chunks":%d,"existing":%d,"stored":%d}`+"\n", res.Chunks, res.Existing, res.Stored)
	return writeErr
}

func mainExitCode() int {
	timeout := flag.Duration("timeout", 30*time.Minute, "overall index timeout (<=24h)")
	flag.Parse()
	if *timeout <= 0 || *timeout > 24*time.Hour {
		fmt.Fprintln(os.Stderr, `{"error":"invalid timeout"}`)
		return 1
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, *timeout)
	defer cancel()
	cfg := config.Load()
	if cfg == nil {
		fmt.Fprintln(os.Stderr, `{"error":"config load failed"}`)
		return 1
	}
	if err := run(ctx, cfg); err != nil {
		if writeErr := writeIndexFailure(os.Stderr, err); writeErr != nil {
			fmt.Fprintln(os.Stderr, `{"error":"index failure reporting failed"}`)
		}
		return 1
	}
	return 0
}

func main() {
	os.Exit(mainExitCode())
}
