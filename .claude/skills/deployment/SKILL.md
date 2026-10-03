---
name: deployment
description: バックエンド(Cloudflare Workers + Containers)・フロントエンド(Cloudflare Workers Static Assets)へのアプリケーションデプロイ。両方とも GitHub Actions 経由で自動化。
---

# デプロイメントスキル

> AnimalEkarte は **バックエンドを Cloudflare Workers + Containers**、**フロントエンドを Cloudflare Workers Static Assets**（EMR-255 で Vercel から移行）にデプロイ。両方とも GitHub Actions ワークフロー経由で自動化（`backend-deploy.yml` / `frontend-deploy.yml`）。
> AWS ECS/RDS は廃止済みで、切り戻し先やホットスタンバイではない。

## このスキルを使用するタイミング

- 新しいバージョンのデプロイ
- インフラストラクチャの更新
- 環境変数の管理

## デプロイフロー（実際）

```
Backend:  git push → backend-deploy.yml → wrangler deploy → migrate（Cloudflare Workers + Containers）
Frontend: git push → frontend-deploy.yml → APP_ENV付き pnpm build → wrangler deploy（CLOUDFLARE_API_TOKEN使用）
```

## デプロイ方法

デプロイは **GitHub Actions で自動実行**（`scripts/deploy.sh` は存在しない）。

```bash
# 参照のみ。push / workflow_dispatch は USER の明示承認後。エージェントは自動実行しない
# ステージング: staging ブランチへの push で自動デプロイ（backend/** / frontend/** 変更時）
# git push origin staging

# 手動トリガー (workflow_dispatch) も USER 承認後
# gh workflow run backend-deploy.yml

# CI ステータス確認（読み取り）
gh run list --workflow=backend-deploy.yml
gh run list --workflow=frontend-deploy.yml
```

ジョブ構成の正本は `ci-cd-automation`。デプロイ手順の正本はこのスキル。二重にジョブ一覧を増やさない。

## CI パイプライン（ci.yml）

実ジョブ構成は `ci-cd-automation` スキルを参照（最小構成: `Workflow Contracts` + `Gitleaks Secret Scan` のみ。build/test/coverage/lint/migration/worker/codegen/audit は `make ci` = `scripts/run-local-ci.sh` 側に集約済み）。

## 重要な注意事項

- 必ずステージングで検証してから本番へ
- バックエンド障害時は `docs/ops/infra/staging/runbook.md` に従う。AWS への切り戻しはできないため、Cloudflare 側の修正・再デプロイ、またはスナップショット + 現行 IaC からの再建で復旧する
- デプロイ後はモニタリングダッシュボードを確認

## プラットフォーム参照

- 現行インフラ: [`docs/ops/infra/architecture.md`](../../../docs/ops/infra/architecture.md)
- STG 運用: [`docs/ops/infra/staging/runbook.md`](../../../docs/ops/infra/staging/runbook.md)
- AWS 退役記録（実行禁止）: [platforms/aws.md](./platforms/aws.md)
- フロントエンドは Cloudflare Workers Static Assets（`.github/workflows/frontend-deploy.yml` → `frontend/wrangler.jsonc`）にデプロイ。AWS 対象外。`/api/*` は Worker の service binding が backend へ同一オリジン中継する。
