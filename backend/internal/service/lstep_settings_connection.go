package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	apperrors "github.com/animal-ekarte/backend/internal/errors"
	"github.com/animal-ekarte/backend/internal/model"
)

func (s *lstepSettingsService) TestConnection(ctx context.Context, clinicID uint64) (*LstepConnectionTestResult, error) {
	records, err := s.repo.FindByClinicAndService(ctx, clinicID, model.IntegrationServiceLstep)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find lstep settings for test", "error", err)
		return nil, apperrors.Wrap(err, "failed to load settings for connection test")
	}

	kvMap := make(map[string]string, len(records))
	for _, r := range records {
		val, _ := s.decrypt(r.KeyName, r.KeyValue)
		kvMap[r.KeyName] = val
	}

	result := &LstepConnectionTestResult{}

	// Lステップ疎通確認
	lstepKey := kvMap[model.IntegrationKeyLstepAPIKey]
	lstepBase := kvMap[model.IntegrationKeyLstepBaseURL]
	if lstepBase == "" {
		lstepBase = "https://api.lstep.jp"
	}
	if lstepKey != "" {
		if err := testLstepAPI(ctx, lstepBase, lstepKey); err != nil {
			result.LstepOK = false
			result.LstepError = err.Error()
		} else {
			result.LstepOK = true
		}
	}

	// LINE Messaging API疎通確認
	lineToken := kvMap[model.IntegrationKeyLineChannelAccessToken]
	if lineToken != "" {
		if err := testLineAPI(ctx, lineToken); err != nil {
			result.LineOK = false
			result.LineError = err.Error()
		} else {
			result.LineOK = true
		}
	}

	return result, nil
}

func testLstepAPI(ctx context.Context, baseURL, apiKey string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/tags", http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()               //nolint:errcheck // close error on connectivity probe is not actionable
	_, _ = io.Copy(io.Discard, resp.Body) //nolint:errcheck // body drain failure on connectivity probe is not actionable
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed: HTTP %d", resp.StatusCode)
	}
	return nil
}

func testLineAPI(ctx context.Context, channelToken string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.line.me/v2/bot/info", http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+channelToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()               //nolint:errcheck // close error on connectivity probe is not actionable
	_, _ = io.Copy(io.Discard, resp.Body) //nolint:errcheck // body drain failure on connectivity probe is not actionable
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("authentication failed: HTTP %d", resp.StatusCode)
	}
	return nil
}
