# インフラドキュメント（SSOT）

> **目的**: 現行インフラ（Cloudflare）の構成・運用・規約の入口。**環境ごとに STG / PROD を分離**して管理する。
> **読者**: 全開発者・PO。**タイミング**: インフラ変更・障害対応・環境構築の前。

| ドキュメント | 内容 |
|---|---|
| [architecture.md](architecture.md) | 現行構成の全体像（env 共通） |
| [iac-guidelines.md](iac-guidelines.md) | IaC 運用規約（Terraform/Wrangler 境界・state・token・drift） |
| [staging/runbook.md](staging/runbook.md) | STG 運用手順 |
| [production/setup.md](production/setup.md) | 本番構築手順（#253） |
| [production/runbook.md](production/runbook.md) | 本番運用手順（構築後に整備） |
| [_archive/](_archive/) | 完了した作業の凍結記録（STG 移行全記録・AWS 時代の文書） |

- コードの所在: Terraform = `infra/cloudflare/`（Phase B で `envs/{staging,production}` + `modules/` へ移行予定 — `reorg-plan.md`）／Wrangler = `backend/wrangler*.jsonc`
- デプロイ・seed 等の作業手順は [docs/ops/deploy/](../deploy/README.md) が正本（本ディレクトリからは重複させず参照する）
