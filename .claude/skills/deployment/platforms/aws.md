# AWS 基盤の退役記録（実行禁止）

> **状態**: AWS ECS/RDS 基盤は 2026-07-20 に廃止済み。関連 workflow、Terraform、
> ECS task definition、停止スケジューラも削除済み。
>
> **重要**: AWS は切り戻し先・ホットスタンバイではない。本ファイルに実行可能な
> deploy / rollback / monitoring コマンドを置かない。

## 現行の正本

- [インフラ構成](../../../../docs/ops/infra/architecture.md)
- [STG 運用 Runbook](../../../../docs/ops/infra/staging/runbook.md)
- [PROD 運用 Runbook](../../../../docs/ops/infra/production/runbook.md)

現行バックエンドは Cloudflare Workers + Containers、DB は PlanetScale Postgres。
障害時は Cloudflare 側の修正・再デプロイ、またはスナップショットと現行 IaC からの再建で復旧する。

## 凍結履歴

AWS 時代の構成・判断・廃止証跡を調査する場合だけ
[`docs/ops/infra/_archive/aws-legacy/`](../../../../docs/ops/infra/_archive/aws-legacy/) と
[`migration-cloudflare.md`](../../../../docs/ops/infra/_archive/migration-cloudflare.md) を参照する。
archive 内の CLI / Terraform 手順は当時の記録であり、現在の環境へ実行してはならない。
