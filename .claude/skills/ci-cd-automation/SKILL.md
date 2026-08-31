---
name: ci-cd-automation
description: このプロジェクトの GitHub Actions CI/CD 構成の把握と、CI 失敗調査・ローカル検証の手順。CI が赤い時、ワークフロー変更時、push 後の結論確認時に使用。
---

# CI/CD Pipeline — AnimalEkarte 実構成と失敗調査

このスキルは**実在するワークフローのみ**を記述する。架空のテンプレートを書かないこと。
ワークフロー定義の正本は `.github/workflows/` であり、変更時は必ず実ファイルを読む。

## 実在するワークフロー一覧

| ファイル | 役割 |
|---------|------|
| `ci.yml` | メイン CI（下記ジョブ構成） |
| `backend-deploy.yml` | Backend の Cloudflare Workers + Containers デプロイ |
| `frontend-deploy.yml` | Frontend デプロイ |
| `e2e.yml` | E2E テスト |
| `security-scan.yml` | agentshield（エージェント設定の監査。Go コードスキャナではない） |
| `performance-tests.yml` | パフォーマンステスト（push 後のみ） |
| `actionlint.yml` | ワークフロー自体の lint（`paths: .github/workflows/**` フィルタ） |
| `stg-smoke.yml` | STG ヘルススモーク |
| `worker-secret-sync.yml` | GitHub Secrets から Cloudflare Worker secrets への明示同期 |

`backend-deploy-ecs.yml` と `staging-stop.yml` は AWS 廃止時に削除済み。AWS はホットスタンバイではなく、
これらの workflow 名を復旧手順として案内・実行しない。現行インフラと障害初動の正本は
`docs/ops/infra/architecture.md` と `docs/ops/infra/staging/runbook.md`。

## ci.yml の実ジョブ構成

- **トリガー**: `pull_request: branches: [main, staging, production]` + `push: branches: [main]`
- **changes**: paths-filter で backend / frontend / migration 変更を判定。openな`main→staging` PRが同じhead SHAを検証する場合だけpush側の重いjobをskip（API障害時はfail-open）
- **Backend Build**: DB不要の`go build ./...`。test shardsと並列
- **Backend Test matrix**: 4つの独立PostgreSQLで`medicalrecord` / `auth` / `staff+billing+reservation` / remainingを並列実行。各shard内は`-race -coverpkg=./internal/... -p 1`
- **Backend**: 4 profileをblock単位で統合し、coverage summary / ratchetを実行する集約check
- **Frontend Build**: `pnpm install --frozen-lockfile` → audit → build
- **Frontend Test matrix**: Vitest 2 shard。coverage payloadをblob reporterへ保存
- **Frontend**: `vitest --mergeReports`でcoverageをnative mergeし、ratchetを実行する集約check
- **codegen-check**: Go モデル ↔ models.ts の同期検証
- **migration-verify**: PR→mainのマイグレーション検証

使用アクションの正: `actions/checkout@v7` / `actions/setup-go@v7` / `actions/setup-node@v7` / `actions/upload-artifact@v7` / `actions/download-artifact@v7` / `pnpm/action-setup@v6`。
（`node-actions/setup-node` や `go-actions/setup-go` というアクションは存在しない — 過去にこのスキルが記載していた誤り）

## CI 失敗調査の手順（実績由来）

過去の実害から抽出した盲点。この順で確認する。

### 1. run の特定は `--branch` 軸（`--commit` は使わない）

```bash
# ❌ このリポでは空を返す
gh run list --commit <sha>
# ✅ branch + headSha で自前フィルタ
gh run list --branch main --json databaseId,event,conclusion,headSha --limit 10
```
（出典: memory ops_gh_run_list_commit_empty_use_branch）

### 2. job / step 単位で確認（conclusion=success を額面で信じない）

```bash
gh run view <run-id> --json jobs
```

- **aggregate マスク**: `Backend` / `Frontend` はbranch protection互換の集約check。赤い場合は集約stepだけでなく、`Backend Build`・4 test shards、`Frontend Build`・2 test shardsのどれが先に失敗したか確認する。matrixは`fail-fast: false`なので全shardの結論を列挙する
- **paths-filter silent green**: changes ジョブで skip されたジョブは実行されていないのに全体は success に見える。skip されたジョブがあれば当該層は「未検証」として扱う
（出典: memory feedback_ci_step_order_masks_lint / feedback_paths_filter_silent_green / ops_golangci_lint_cap_and_reconcile_20260630）

### 3. golangci-lint の件数 cap に注意

現行の `backend/.golangci.yml` は `max-issues-per-linter: 0 / max-same-issues: 0`（cap 解除済み・2026-06-30対応）。
ただし cap が再導入された場合や古いブランチでは `max-same-issues` / `max-issues-per-linter` により**件数が過少表示される**（11件目以降が隠れる）ため、完全な件数確認では cap 解除を明示付与する:

```bash
--max-same-issues 0 --max-issues-per-linter 0
```
（出典: memory ops_golangci_lint_cap_and_reconcile_20260630、commit 7d103994）

### 4. ローカル再現（スコープ限定・禁止コマンド回避）

全体 `go test ./...` / `golangci-lint run ./...` は CLAUDE.md の自動実行禁止コマンド。スコープ限定 + 以下の罠回避で再現する:

```bash
# lint: entrypoint.sh がコマンドを無視するため --entrypoint 上書き必須
docker compose run --rm --no-deps -T --entrypoint golangci-lint backend run ./internal/repository/...

# キャッシュ偽0件の回避（stale cache で 0 issues に見える）
docker compose exec -T backend sh -c 'GOLANGCI_LINT_CACHE=/tmp/glc-$RANDOM golangci-lint run ./internal/handler/...'
```
（出典: memory ops_backend_scoped_lint_entrypoint_override / ops_golangci_lint_stale_cache_false_zero）

### 5. DB 依存テストの fresh-DB ゲート

warm-DB（前 run のテーブル残存）でローカル PASS しても CI の fresh DB で FAIL する。seed / migration / ENUM を変更した場合はテスト DB を作り直して1回走らせる（DB再作成は自動実行禁止コマンドに該当するためユーザーに手動実行を依頼する）。
（出典: memory ops_golangci_lint_cap_and_reconcile_20260630）

## ワークフロー変更時の注意

- **バージョン統一**: Actions のバージョンはリポジトリ全体で統一されている（現在は setup-node v7 等）。1ファイルだけ上げない（出典: memory infra002_github_actions_unification_complete — 「value の DRY ≠ actions の DRY」）
- **drift スキャンは `with:` / `env:` / `working-directory:` も対象**（出典: memory feedback_workflow_with_param_drift）
- **actionlint はバージョンドリフトを検出しない**。「actionlint が通った＝Actions バージョン統一が保たれている」は誤り。ドリフト検査の正本は `scripts/check-actions-version-drift.sh`（actionlint.yml に組込済）
- **新規ワークフロー追加は既存の「統一済み」状態を静かに壊す**。新規 yml 追加時は、`scripts/check-actions-version-drift.sh`で既存workflowの現在のactionバージョンと必ず突合する（backend-deploy.yml に setup-node@v4 が混入し #195 が回帰した実例）

  （出典: memory closed_issue_reaudit_20260707 / commit 2d8ab41d）
- **production 起動条件（env）変更は CI workflow にも波及**する。`.github/workflows/` の env を同時更新（出典: memory feedback_config_change_ci_propagation）
- workflow ファイル変更後は actionlint.yml が走る（paths フィルタあり — 見かけ green に注意）

## 完了条件（CI 調査タスクの合格基準）

- [ ] 対象 commit の run を headSha 一致で特定した
- [ ] 全ジョブの conclusion を job 単位で確認し、skip されたジョブを列挙した
- [ ] 失敗時: 失敗 step と未実行 step を区別して報告した
- [ ] 「CI green」と報告する場合、skip ではなく実 success であることを確認済み

## 関連スキル

- `docker-patterns` — Docker Compose 開発環境・イメージ最適化（docker-optimization は削除済み・本スキルに統合）
- `deployment` — デプロイメント（backend-deploy.yml / frontend-deploy.yml）
- `migration-seed-safety` — migration / seed 変更時の安全ガード
