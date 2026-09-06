# SKILL-REEVAL — 代表タスクでの再評価

更新日: 2026-09-06  
前提: GO-REFS → API-EXAMPLES → REVIEW-EVIDENCE → ENTRY-SLIM → DEDUP → CODEX-MIRROR の正本修正後。

## 代表タスクと停止判断

| タスク | 開く資料 | 止まる条件 |
|--------|----------|------------|
| docs 誤字 | 対象 md のみ。Go/Gin guidelines は不要 | 秘密・行値・credential を書こうとしたとき |
| 局所 Go | `golang-testing`、ADR-006、変更 package。lint は `scoped-verification-gates` | 存在しない `internal/service` や `go-linting` を必須にしたとき。全体 `go test ./...` |
| FE テスト | `test-generation`、変更 spec。E2E は `E2E_TESTING_GUIDE` | full clinical suite を auth smoke と混ぜたとき。実 LINE |
| migration | `migration-seed-safety`。apply は USER | エージェントが `make migrate` / shared STG apply を始めたとき |
| STG 準備 | `stg-release-readiness`、`STG_PLANETSCALE_SEED_RUNBOOK` の stop gates | 八王子 CSV 生成、PlanetScale apply、`make reset` |

## 整合

- Session A/B を実行手順にしない。worktree + claim で足りる。
- CSRF 例は Cookie + `X-Requested-With`。meta token を正本にしない。
- `apperrors.Wrap` は `(err, message)`。`WrapUnauthorized` は message のみ。
- 未実行テスト / 未測定 coverage を PASS と書かない。
- push / dispatch は `deployment` の承認境界。エージェントは実行しない。
- export 削除は Feature Indexing の公開契約を先に確認する。

ミラー（`.agents/` / `.codex/`）は gitignore。正本は `.claude/`。同期は `sync-agents-skills.sh` または commit hook。
