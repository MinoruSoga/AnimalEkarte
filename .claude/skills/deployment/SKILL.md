---
name: deployment
description: バックエンド(Cloudflare Workers + Containers)・フロントエンド(Vercel)へのアプリケーションデプロイ。両方とも GitHub Actions 経由で自動化。
---

# デプロイメントスキル

> AnimalEkarte は **バックエンドを Cloudflare Workers + Containers**、**フロントエンドを Vercel** にデプロイ。両方とも GitHub Actions ワークフロー経由で自動化（`backend-deploy.yml` / `frontend-deploy.yml`）。
> AWS ECS/RDS は廃止済みで、切り戻し先やホットスタンバイではない。

## このスキルを使用するタイミング

- 新しいバージョンのデプロイ
- インフラストラクチャの更新
- 環境変数の管理

## デプロイフロー（実際）

```
Backend:  git push → backend-deploy.yml → wrangler deploy → migrate（Cloudflare Workers + Containers）
Frontend: git push → frontend-deploy.yml → vercel pull → vercel build → vercel deploy（VERCEL_TOKEN使用）
```

## デプロイ方法

デプロイは **GitHub Actions で自動実行**（`scripts/deploy.sh` は存在しない）。

```bash
# ステージング: staging ブランチへの push で自動デプロイ（backend/** / frontend/** 変更時。frontend は production push でも発火）
git push origin staging

# 手動トリガー (workflow_dispatch)
# GitHub Actions → backend-deploy.yml / frontend-deploy.yml → Run workflow

# CI ステータス確認
gh run list --workflow=backend-deploy.yml
gh run list --workflow=frontend-deploy.yml
gh run watch
```

## CI パイプライン（ci.yml）

実ジョブ構成は `ci-cd-automation` スキルを参照（changes判定 + 6本のインベントリlint + backend(build/lint/test/schema-drift) + frontend(audit/type-check/test/lint/build の4ゲート) + codegen-check + migration-verify）。

## 重要な注意事項

- 必ずステージングで検証してから本番へ
- バックエンド障害時は `docs/ops/infra/staging/runbook.md` に従う。AWS への切り戻しはできないため、Cloudflare 側の修正・再デプロイ、またはスナップショット + 現行 IaC からの再建で復旧する
- デプロイ後はモニタリングダッシュボードを確認

## プラットフォーム参照

- 現行インフラ: [`docs/ops/infra/architecture.md`](../../../docs/ops/infra/architecture.md)
- STG 運用: [`docs/ops/infra/staging/runbook.md`](../../../docs/ops/infra/staging/runbook.md)
- AWS 退役記録（実行禁止）: [platforms/aws.md](./platforms/aws.md)
- フロントエンドは Vercel（`.github/workflows/frontend-deploy.yml`）にデプロイ。AWS 対象外。
