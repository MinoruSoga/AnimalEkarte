# scripts/ — repo ユーティリティ索引

CI 契約チェック・DB リセット/移行ツール・セキュリティ/ポリシー検査・リハーサル用スクリプトの置き場。
多くの `check-*` / `verify-*` は `make ci` (`run-local-ci.sh`) または GitHub Actions から呼ばれる**自動ゲート**であり、
手動ユーティリティ（`stage-*` / `link-*` / `apply-*` 等）は開発者が必要時に実行する。

## レイアウト

| パス | 内容 |
|------|------|
| `lib/` | `check-a4-*` / `run-f8-g4-*` が共有する mjs モジュール（A4/F8-G4 リハーサル基盤） |
| `sql/` | `apply-*.py` / `link-*.py` が流す companion SQL（old_db スタッフ紐付け系） |
| `tmp/` | **gitignore 済みスクラッチ置き場**（ローカル seed スクリプト等）。CI/ゲート対象外 |

## 規約

- 新規 `.sh` は `shellcheck-scripts.sh`（severity=warning）の自動検査対象になる（対象は `scripts/*.sh` 直下のみ、サブdirは無検査）
- 契約チェック系は `<name>.test.sh` / `<name>.test.mjs` のペアで退行テストを持つものが多い。ゲート追加時はテスト同梱を推奨
- ファイル全体への `shellcheck disable` は reject される（個別行 disable は可）

## カタログ（目的別）

### CI / ワークフロー契約

| ファイル | 用途 |
|------|------|
| `run-local-ci.sh` | `make ci` 本体。30ステップのローカルCIランナー（backend test は4シャード並列） |
| `check-workflow-contracts.test.mjs` | workflow/script/ドキュメント間の契約一貫性テスト（CI step 2） |
| `check-ci-step-order.sh` (+`.test.sh`) | run-local-ci.sh のステップ順序契約 |
| `ci_scope_plan.py` (+`_test.py`) | CIスコープ計画（変更範囲→必要検証の決定） |
| `ci_tygo_hosp_pass.py` | tygo 設定生成（hospitalization-responses 限定） |
| `ci-migration-verify.sh` | migration 適用検証 |
| `merge_go_coverprofiles.py` (+`_test.py`) | シャード別 Go coverage マージ（重複ブロック除外） |
| `check-test-worker-makefile.test.sh` | test worker の Makefile 契約 |
| `check-actions-version-drift.sh` (+`.test.sh`) | GitHub Actions のバージョンドリフト検査 |

### DB リセット / old_db 移行パイプライン

| ファイル | 用途 |
|------|------|
| `local-db-reset-contract.sh` (+`.test.sh`) | `make reset` の postflight 契約チェック |
| `check-reset-wait-services.sh` (+`.test.sh`) | reset 後のサービス待機契約 |
| `stage-old-db-handoff.sh` / `check-old-db-handoff.sh` (+`layout.test.mjs`) | old_db 21テーブル CSV の PHI quarantine ステージ/検証 |
| `import-old-db-handoffs-on-reset.sh` | reset 後に全 staged handoff を自動 import |
| `stg-uat-old-db-handoff.sh` (+`.test.sh` + `.local.env*`) | staged handoff を STG 共有環境へ bind/import（UAT用） |
| `verify_seed.py` / `verify_seed_matches_stg_dump_full.sh` (+`.test.sh`) | seed 適用結果の検証（STG dump との照合含む） |
| `staff-activity-map-apply.sh` + `apply-old-db-staff-activity-map.py` (+`sql/`) | 移行スタッフの activity map に基づく有効化 |
| `staff-identity-map-link.sh` + `link-old-db-staff-identity-map.py` / `link-old-db-cross-clinic-staff-accounts.py` (+`sql/`) | 移行スタッフの cross-clinic 同一人物紐付け |
| `test_account_csv_layout.py` / `check-csv-import-account-source.test.sh` | CSV import レイアウト/アカウント出自の契約 |

### セキュリティ / エージェントポリシー

| ファイル | 用途 |
|------|------|
| `check-agent-security-policy.sh` (+`.test.mjs`) | エージェントセキュリティポリシー契約 |
| `check-workflow-remote-exec-policy.sh` | ワークフローのリモート実行ポリシー検査 |
| `verify-agent-task.py` (+`test_verify_agent_task.py` + `test_agent_scope_contracts.py`) | 委譲タスクのスコープ/成果物検証（offline） |
| `export-codex-security-latest.sh` | Codex セキュリティレポート出力（既定 archive） |

### ドキュメント / ドリフト検査

| ファイル | 用途 |
|------|------|
| `generate-agent-doc-sections.sh` | `.claude/CLAUDE.md` の `AUTO:*` セクション再生成（このディレクトリの CLAUDE.md 追加時は rows へ登録して再実行） |
| `check-docs-symbol-drift.sh` (+`.test.sh`) | docs 内シンボル参照のドリフト検査 |
| `yaml-frontmatter-description.mjs` (+`.test.mjs`) | frontmatter description 契約 |
| `check-design-primary-cta.mjs` (+`.test.sh`) | デザイン規約（Primary CTA = brand）検査 |
| `manual-sync.py` | manual コンテンツ同期 |
| `verify_research_html.py` | research HTML 成果物の数値突合検証 |

### A4 / F8-G4 リハーサル基盤

| ファイル | 用途 |
|------|------|
| `run-f8-g4-rehearsal.mjs` / `check-f8-g4-resources.mjs` | F8-G4 リハーサル実行/リソース検査 |
| `check-a4-rehearsal-compose.mjs` / `check-a4-resource-boundary.mjs` / `check-a4-env-file.test.mjs` / `write-a4-runtime-report.mjs` (+各`.test.*`) | A4 リハーサルの compose/境界/env/レポート契約 |

### lab-device エージェント

| ファイル | 用途 |
|------|------|
| `build-lab-device-agent-bundle.sh` (+`.test.mjs`) / `install-lab-device-agent.sh` / `canonicalize-lab-device-origin.mjs` | 検査機器連携エージェントの macOS パッケージング/インストール/origin 正規化 |

### その他

| ファイル | 用途 |
|------|------|
| `shellcheck-scripts.sh` (+`.test.sh`) | `scripts/*.sh` 一括 shellcheck ゲート（make ci 実行） |
| `validate-k6-summary.mjs` (+`.test.mjs`) | k6 負荷試験サマリ検証 |
| `auth-d5-dualprocess.sh` | D5 二重プロセス即時失効/レイテンシハーネス |
