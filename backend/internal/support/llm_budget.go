package support

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"

	"github.com/animal-ekarte/backend/internal/apperrors"
	"github.com/animal-ekarte/backend/internal/persistence"
)

// llm_budget.go — support_llm_budgets による全 container / CLI 共有の
// 最小間隔 quota（R7）。課金記録ではなく派生的 quota 状態であり clinic scope
// を持たない。原子性は INSERT ... ON CONFLICT ... WHERE next_allowed_at<=now()
// RETURNING の単一行条件付き更新で担保する。

// budget scopes
const (
	LLMBudgetScopeGlobal = "global"
	// LLMBudgetScopeUserPrefix + "clinic:<id>/staff:<id>"
	LLMBudgetScopeUserPrefix = "user:"
	LLMBudgetScopeIndex      = "index"
)

// 固定間隔（設計値。global=48RPM 相当 / index=12RPM / user=6RPM）。
const (
	LLMBudgetGlobalInterval = 1250 * time.Millisecond
	LLMBudgetUserInterval   = 10 * time.Second
	LLMBudgetIndexInterval  = 5 * time.Second
)

// LLMBudgetGranter は (scope, model) 単位の最小間隔 grant を原子的に取得する。
// 複数 scope の同時確保は必ず単一トランザクションの GrantAll を使う。
type LLMBudgetGranter interface {
	GrantAll(ctx context.Context, grants []BudgetGrant) (granted bool, err error)
}

// LLMUserBudgetScope は clinic×staff 単位 scope を組み立てる。
func LLMUserBudgetScope(clinicID, staffID uint64) string {
	return fmt.Sprintf("%sclinic:%d/staff:%d", LLMBudgetScopeUserPrefix, clinicID, staffID)
}

type llmBudgetGranter struct {
	db *gorm.DB
}

// NewLLMBudgetGranter は LLMBudgetGranter を初期化する。
func NewLLMBudgetGranter(db *gorm.DB) LLMBudgetGranter {
	return &llmBudgetGranter{db: db}
}

// BudgetGrant は 1 scope の grant 要求。
type BudgetGrant struct {
	Scope, Model string
	Interval     time.Duration
}

// GrantAll は複数 scope の grant を単一トランザクションで all-or-none に
// 取得する。deadlock 防止のため scope/model の正準順で処理する。
func (g *llmBudgetGranter) GrantAll(ctx context.Context, grants []BudgetGrant) (bool, error) {
	if g.db == nil {
		return false, apperrors.WrapInvalidInput("llm budget requires a db")
	}
	if len(grants) == 0 {
		return false, apperrors.WrapInvalidInput("no budget grants")
	}
	seen := map[string]struct{}{}
	for _, gr := range grants {
		if gr.Scope == "" || gr.Model == "" || gr.Interval <= 0 {
			return false, apperrors.WrapInvalidInput("invalid budget grant")
		}
		key := gr.Scope + "/" + gr.Model
		if _, dup := seen[key]; dup {
			return false, apperrors.WrapInvalidInput("duplicate budget grant scope")
		}
		seen[key] = struct{}{}
	}
	sorted := append([]BudgetGrant(nil), grants...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Scope != sorted[j].Scope {
			return sorted[i].Scope < sorted[j].Scope
		}
		return sorted[i].Model < sorted[j].Model
	})
	granted := true
	err := persistence.DBOrTx(ctx, g.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, gr := range sorted {
			var next time.Time
			res := tx.Raw(`INSERT INTO support_llm_budgets (scope, model, next_allowed_at)
				VALUES (?, ?, now() + ?::interval)
				ON CONFLICT (scope, model) DO UPDATE SET next_allowed_at = EXCLUDED.next_allowed_at
				WHERE support_llm_budgets.next_allowed_at <= now()
				RETURNING next_allowed_at`,
				gr.Scope, gr.Model, fmt.Sprintf("%d milliseconds", gr.Interval.Milliseconds())).Scan(&next)
			if res.Error != nil {
				return apperrors.Wrap(res.Error, "failed to grant llm budget")
			}
			if res.RowsAffected == 0 {
				granted = false
				return errNotGranted // rollback all
			}
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errNotGranted) {
			return false, nil
		}
		return false, err
	}
	return granted, nil
}
