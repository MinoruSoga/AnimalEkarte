# ops/ — 運用系（デプロイ・CI・テスト・インフラ）

> **目的**: システムを「どう動かすか」の運用系ドキュメントの索引を提供する。
> **読者**: 全開発者・DevOps・AI エージェント。
> **タイミング**: デプロイ・CI 変更・テスト実施・インフラ作業の前。

編集時のルールは [CLAUDE.md](CLAUDE.md) を参照。技術設計は [../architecture/](../architecture/README.md)、仕様は [../spec/](../spec/README.md)。

## 索引

| ドキュメント | 内容 | いつ読むか |
|:---|:---|:---|
| [deploy/](deploy/README.md) | デプロイハブ（環境 URL・コマンド・ロールバック判定。Cloudflare 正系統 / AWS ECS はロールバック専用） | デプロイ・リリース作業前 |
| [deploy/runbooks/](deploy/runbooks/README.md) | 個別実行手順書（リリース前検証・ECS ロールバック・シークレット棚卸し） | 該当オペレーション実施時 |
| [testing/](testing/README.md) | テスト戦略・手動検証シナリオ・E2E ガイド・プロファイリング | テスト実施・品質検証時 |
| [ci-policy.md](ci-policy.md) | CI ワークフローの決定事項記録（Actions バージョンピン方針等） | .github/workflows/ 変更前 |
| [coverage-policy.md](coverage-policy.md) | テストカバレッジ ratchet 方式の運用ポリシー | カバレッジゲート調整時 |
| [backlog-spreadsheet.md](backlog-spreadsheet.md) | Q&A バックログスプレッドシートの運用ルール | クライアント Q&A シート操作前 |
| [infra-architecture.md](infra-architecture.md) | インフラ構成図・ネットワーク・セキュリティ設計（`../architecture/overview.md` のレイヤード構造とは別物） | インフラ構成の調査・変更前 |
| [p2-terraform-plan-runbook.md](p2-terraform-plan-runbook.md) | P2 Terraform（internal ALB + VPC Origin）plan/apply ランブック | P2 インフラ適用時 |
| [stg-aws-change-readiness.md](stg-aws-change-readiness.md) | STG AWS 変更の準備状況・SG 絞り込み手順 | AWS 側変更の検討時 |
| [stg-aws-cost-reduction.md](stg-aws-cost-reduction.md) | STG AWS コスト削減の実施記録と方針 | AWS コスト見直し時 |

## AI エージェント向け注記

- タスク台帳はリポジトリ直下 [`todo.md`](../../todo.md) に一元化されている。運用作業で発見した課題はそこへ起票する。
- migration/seed に触れる作業は `migration-seed-safety` スキル、リリース前チェックは `stg-release-readiness` スキルを先に読むこと。
