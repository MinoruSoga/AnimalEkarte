package lstep

import (
	"context"
	"log/slog"
	"strings"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/model"
)

// conditionTagMapFromMappings は DB レコードを疾患コード→タグ名マップに変換する（純粋関数）。
func conditionTagMapFromMappings(mappings []*model.LstepConditionTagMapping) map[string]string {
	m := make(map[string]string, len(mappings))
	for _, mapping := range mappings {
		m[mapping.ConditionCode] = mapping.TagName
	}
	return m
}

// SyncChronicConditionTags は慢性疾患タグを差分同期する（BE-012）。
func (s *lstepTagSyncService) SyncChronicConditionTags(ctx context.Context, clinicID, ownerID uint64, activeConditionCodes []string) error {
	lineUserID, ok, err := s.resolveSyncTarget(ctx, clinicID, ownerID, "chronic condition")
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	client, err := s.buildClient(ctx, clinicID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to build lstep client for chronic condition tag sync", "error", err)
		return apperrors.Wrap(err, "failed to build lstep client")
	}
	if client == nil {
		return nil
	}

	// 慢性疾患コードマッピングを DB から取得
	var conditionMap map[string]string
	if s.tagConfigRepo != nil {
		mappings, loadErr := s.tagConfigRepo.FindAllConditionTagMappings(ctx)
		if loadErr != nil {
			slog.ErrorContext(ctx, "failed to load condition tag mappings", "error", loadErr)
			return apperrors.Wrap(loadErr, "failed to load condition tag mappings")
		}
		conditionMap = conditionTagMapFromMappings(mappings)
	} else {
		conditionMap = map[string]string{}
	}

	// 目標タグセットを構築
	activeTags := make(map[string]bool, len(activeConditionCodes))
	for _, code := range activeConditionCodes {
		if tag, ok := conditionMap[code]; ok {
			activeTags[tag] = true
		}
	}

	// 現在の chronic_* キャッシュを取得
	cached, err := s.tagCacheRepo.FindByOwner(ctx, clinicID, ownerID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to find cached tags for chronic sync", "error", err)
		return apperrors.Wrap(err, "failed to find cached tags")
	}

	existingSet := make(map[string]bool, len(cached))
	apiFailed := false
	for _, t := range cached {
		if !strings.HasPrefix(t.TagName, tagPrefixChronic) {
			continue
		}
		existingSet[t.TagName] = true

		// アクティブでないタグを解除
		if activeTags[t.TagName] {
			continue
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, t.TagName, "chronic", "", false); err != nil {
			apiFailed = true
			continue
		}
	}

	// 未付与のアクティブタグを付与
	for tagName := range activeTags {
		if existingSet[tagName] {
			continue
		}
		if err := s.applyTagState(ctx, client, clinicID, ownerID, lineUserID, tagName, "chronic", "", true); err != nil {
			apiFailed = true
			continue
		}
	}

	if !apiFailed {
		s.notifyAPISuccess(ctx, client, clinicID, ownerID, lineUserID)
	}
	return nil
}
