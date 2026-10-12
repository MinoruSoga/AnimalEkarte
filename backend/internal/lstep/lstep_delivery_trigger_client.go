package lstep

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/infra/lstep"
	"github.com/animal-ekarte/backend/internal/model"
)

// addDeliveryTag は provider 呼び出しのみを行う（D1: owner 直列化境界内で実行し、
// 結果記録は境界解放後の recordDeliveryOutcome に委ねる）。
func (s *lstepDeliveryTriggerService) addDeliveryTag(ctx context.Context, client lstep.Client, lineUserID, tagName string) error {
	return client.AddTag(ctx, lineUserID, tagName)
}

// recordDeliveryOutcome は provider 結果をトリガーログへ永続化する。
func (s *lstepDeliveryTriggerService) recordDeliveryOutcome(ctx context.Context, clinicID, logID uint64, tagName string, sendErr error) error {
	if sendErr != nil {
		reason := fmt.Sprintf("lstep_add_tag_failed: %s", tagName)
		// LSA-12 / DEC-35: failed status 更新失敗は silent にしない（monitor が scheduled のまま残る）
		addErr := apperrors.Wrap(sendErr, "failed to add lstep tag")
		if updateErr := s.triggerLogRepo.UpdateStatus(ctx, clinicID, logID, model.TriggerStatusFailed, nil, &reason); updateErr != nil {
			slog.ErrorContext(ctx, "failed to record trigger log failed status", "log_id", logID, "error", updateErr)
			return errors.Join(addErr, apperrors.Wrap(updateErr, "failed to update trigger log status to failed"))
		}
		return addErr
	}
	now := time.Now()
	if err := s.triggerLogRepo.UpdateStatus(ctx, clinicID, logID, model.TriggerStatusFired, &now, nil); err != nil {
		slog.ErrorContext(ctx, "failed to update trigger log status", "error", err, "log_id", logID)
		return apperrors.Wrap(err, "failed to update trigger log status")
	}
	return nil
}

// buildClient はクリニック設定から lstep.Client を構築する。
// 同期無効または API キー未設定の場合は nil, nil を返す（スキップ）。
func (s *lstepDeliveryTriggerService) buildClient(ctx context.Context, clinicID uint64) (lstep.Client, error) {
	if s.clientBuilderFn != nil {
		return s.clientBuilderFn(ctx, clinicID)
	}
	enabled, err := s.settingsSvc.IsSyncEnabled(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to check lstep sync enabled")
	}
	if !enabled {
		return nil, nil
	}
	apiKey, baseURL, _, err := s.settingsSvc.GetRawCredentials(ctx, clinicID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to get lstep credentials")
	}
	if apiKey == "" {
		return nil, nil
	}
	return lstep.NewClient(apiKey, baseURL), nil
}

// applySuppression は Q23 優先順位による配信抑制を判定・適用する。
// 戻り値 suppressed=true なら呼び出し元が抑制ログを作成して return すべき。
// nil prioritySvc の場合は抑制スキップ（後方互換）。
