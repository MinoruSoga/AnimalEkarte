# ops/ — 運用系（デプロイ・CI・テスト・インフラ）

> **目的**: システムを「どう動かすか」の運用系ドキュメントの索引を提供する。
> **読者**: 全開発者・DevOps・AI エージェント。
> **タイミング**: デプロイ・CI 変更・テスト実施・インフラ作業の前。

編集時のルールは [CLAUDE.md](CLAUDE.md) を参照。技術設計は [../architecture/](../architecture/README.md)、仕様は [../spec/](../spec/README.md)。

## 索引

| ドキュメント | 内容 | いつ読むか |
|:---|:---|:---|
| [deploy/](deploy/README.md) | デプロイハブ（環境 URL・Cloudflare デプロイ・障害時判断） | デプロイ・リリース作業前 |
| [deploy/runbooks/](deploy/runbooks/README.md) | 個別実行手順書（リリース前検証・scheduler運用・資格情報ローテーション・シークレット棚卸し） | 該当オペレーション実施時 |
| [testing/](testing/README.md) | テスト戦略・手動検証シナリオ・E2E ガイド・プロファイリング | テスト実施・品質検証時 |
| [ci-policy.md](ci-policy.md) | CI ワークフローの決定事項記録（Actions バージョンピン方針等） | .github/workflows/ 変更前 |
| [coverage-policy.md](coverage-policy.md) | テストカバレッジ ratchet 方式の運用ポリシー | カバレッジゲート調整時 |
| [backlog-spreadsheet.md](backlog-spreadsheet.md) | Q&A バックログスプレッドシートの運用ルール | クライアント Q&A シート操作前 |
| [infra/architecture.md](infra/architecture.md) | インフラ構成図・ネットワーク・セキュリティ設計（`../architecture/overview.md` のレイヤード構造とは別物） | インフラ構成の調査・変更前 |
| [infra/_archive/aws-legacy/](infra/_archive/aws-legacy/) | 2026-07-20 に廃止した AWS 基盤の凍結履歴（**実行禁止**） | 過去の判断・実施証跡を調査する時のみ |

## AI エージェント向け注記

- タスク台帳の正本はリポジトリ直下 [`STATUS.md`](../../STATUS.md)（§1 残作業 · §2 Issue · §3 BUG）。USER 実行リストは [`PO-todo.md`](../../PO-todo.md)。運用作業で発見した課題は `STATUS.md` の `## 個別タスク詳細` へ `### TASK-XXX:` 節で起票する（旧 `todo.md` / `bug.md` / `3-session-agent.html` は STATUS へのスタブのみ）。
- migration/seed に触れる作業は `migration-seed-safety` スキル、リリース前チェックは `stg-release-readiness` スキルを先に読むこと。
