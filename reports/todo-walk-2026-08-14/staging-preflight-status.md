# staging preflight status（2026-08-14 · agent 着手継続）

## 状態サマリ

| 項目 | 結果 |
|------|------|
| left-right | **4 / 1346** |
| staging-only disposition | **KEEP×4**（完了 · §4.1） |
| Draft PR | **#299** https://github.com/MinoruSoga/AnimalEkarte/pull/299 （**draft · 未 merge**） |
| merge | **未** |

## Required CI（この PR で期待）

| Workflow | 用途 |
|----------|------|
| `CI` (ci.yml) | 必須 green |
| `Security Scan` | 必須 green |
| `Backend Deploy` | staging 向け · 失敗時 merge しない（過去 staging push で failure 例あり 2026-07-17） |
| `Frontend Deploy (Vercel)` | staging 向け |
| `stg-smoke` | あれば必須 |
| Performance Tests | schedule 失敗があっても本 PR の merge gate にしない（別監視） |

main tip 直近: CI **success** on docs drain（2026-08-13）。

## Migration

| 項目 | 結果 |
|------|------|
| `*.sql` name-only diff staging...main | **`backend/migrations/001_init.sql`**（1 path） |
| seeds 等を含む migrations/ 配下 | 多数（以前 92 paths 集計） |
| checksum / PlanetScale ownership | **未（STG 運用が記入）** |
| agent migrate | **しない** |

## Backup / rollback（文書ポインタ · owner は人）

| 項目 | 正本 |
|------|------|
| CI/CD · rollback 方針 | [`docs/ops/deploy/CI-CD-PIPELINE.md`](../../docs/ops/deploy/CI-CD-PIPELINE.md) — last known good CF 再デプロイ |
| local fresh のみ | [`LOCAL_DB_RESET.md`](../../docs/ops/deploy/LOCAL_DB_RESET.md) |
| PROD runbook | [`docs/ops/infra/production/runbook.md`](../../docs/ops/infra/production/runbook.md) |
| STG seed | [`STG_PLANETSCALE_SEED_RUNBOOK.md`](../../docs/ops/deploy/STG_PLANETSCALE_SEED_RUNBOOK.md) |
| rollback owner role | `[ ]` **人が記入** |
| backup 実施証跡 opaque ref | `[ ]` **人が記入** |

## §E-7 チェック

- [x] remote refs
- [x] 4/1346
- [x] staging-only KEEP
- [x] reset 禁止 · merge-commit 方針
- [x] migration 差分把握（sql 1 + seeds）
- [x] draft PR #299
- [x] required CI 一覧を定義
- [x] rollback 文書ポインタ
- [ ] checksum / ownership（人）
- [ ] backup 実施 + owner role（人）
- [ ] PR CI 全 green（人・CI）
- [ ] draft 解除 + merge commit（人）

```
staging_preflight=PARTIAL draft_pr=299 disposition=KEEP_ALL_4 ci_matrix=DEFINED merge=NOT_DONE opaque_ref=reports/todo-walk-2026-08-14/staging-preflight-status.md
```
