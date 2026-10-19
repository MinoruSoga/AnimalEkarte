package support

import (
	"context"
	"fmt"
)

// chat_hybrid_deps.go — 生成 quota・ユーザー inflight gate の注入面。

// WithLLMBudget は生成 quota granter と生成モデル名を注入する。
func (h *Handler) WithLLMBudget(g LLMBudgetGranter, model string) *Handler {
	h.chatBudget = g
	h.chatModel = model
	return h
}

// grantGenerationBudget は user(clinic×staff, 10s) と global(1250ms) の
// grant を取得する。未付与は errNotGranted（429）、DB 失敗は error（503）。
func (h *Handler) grantGenerationBudget(ctx context.Context, clinicID, staffID uint64) error {
	if h.chatBudget == nil || h.chatModel == "" {
		return fmt.Errorf("llm budget dependency missing")
	}
	ok, err := h.chatBudget.GrantAll(ctx, []BudgetGrant{
		{Scope: LLMUserBudgetScope(clinicID, staffID), Model: h.chatModel, Interval: LLMBudgetUserInterval},
		{Scope: LLMBudgetScopeGlobal, Model: h.chatModel, Interval: LLMBudgetGlobalInterval},
	})
	if err != nil {
		return err
	}
	if !ok {
		return errNotGranted
	}
	return nil
}

func (h *Handler) acquireChatInflight(clinicID, staffID uint64) bool {
	key := fmt.Sprintf("%d/%d", clinicID, staffID)
	h.chatInflightMu.Lock()
	defer h.chatInflightMu.Unlock()
	if h.chatInflight == nil {
		h.chatInflight = make(map[string]struct{})
	}
	if _, ok := h.chatInflight[key]; ok {
		return false
	}
	h.chatInflight[key] = struct{}{}
	return true
}

func (h *Handler) releaseChatInflight(clinicID, staffID uint64) {
	key := fmt.Sprintf("%d/%d", clinicID, staffID)
	h.chatInflightMu.Lock()
	delete(h.chatInflight, key)
	h.chatInflightMu.Unlock()
}
