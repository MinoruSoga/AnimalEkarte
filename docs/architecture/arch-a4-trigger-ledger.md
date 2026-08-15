# ARCH-A4 trigger ledger

> **Purpose**: Record when (and whether) A4 subdomain refactors may start. Pain-first; no bulk splits.
> **Related**: `todo.md` ARCH-A4, [composition-root-conventions](composition-root-conventions.md).

## 2026-08-07 measurement

| Subdomain | Trigger from todo | Evidence | Decision |
|-----------|-------------------|----------|----------|
| **lstep** | LINE 再開 #259 直前 | todo-po: #259 **依存待ち**（先方 enable 後） | **blocked** |
| **billing** | estimate/accounting/billing_item PR が広域 | `billing_item_service.go` **993 → 872** LOC; unbilled 凝集が独立; recent unbilled/billing-item commits | **S1 landed** (unbilled extract) |
| **reservation** | intent/service 肥大の痛み | `reservation_service.go` 666, repository 826; `map[string]any` は owner 内 primitive として残存 | **no bulk** — convert on touch only |

## Landed slices

| Slice | Change | Commit |
|-------|--------|--------|
| **A4-billing-S1** | Extract unbilled aggregation helpers to `billing_item_unbilled.go` (same package, behavior unchanged) | `eca651da3` |
| **A4-billing-S2** | Extract unbilled repo queries to `billing_item_repository_unbilled.go` (769 → ~559 + 224, behavior unchanged) | `13d043315` |
| **A4-billing-S3** | Extract estimate successor path to `estimate_service_successor.go` (598 → ~475 + 137, behavior unchanged) | `264777c88` |

## Next candidates (still trigger-gated)

1. **billing_item_service residual** (post-close / create helpers) — only when a feature PR already forces wide touch.  
2. **accounting_*.go** further cohesion — only with accounting feature work.  
3. **reservation_intent_repository.go** (~626) — only when intent API changes already force large touch.  
4. **lstep** batch/tag/delivery — after #259 unblocked.
