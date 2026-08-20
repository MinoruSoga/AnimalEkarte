# インフラ・運用ディレクトリ (Infrastructure & Operations)

> **目的**: クラウド基盤（Cloudflare Workers + Containers / PlanetScale / Vercel）、デプロイフロー、およびシステムの安定稼働に関する情報の管理。
> **読者**: AI エージェント(Claude Code)。
> **タイミング**: docs/ops/ 配下編集時。
>
> **インフラ SSOT**: [`infra/architecture.md`](infra/architecture.md) /
> [`infra/staging/runbook.md`](infra/staging/runbook.md) /
> [`infra/production/runbook.md`](infra/production/runbook.md)。
> AWS ECS/RDS は 2026-07-20 に廃止済みで、切り戻し先・ホットスタンバイとして使用できない。
> AWS 時代の文書は git 履歴のみ（2026-08-20 削除。`git show e0260d32f^:docs/ops/infra/_archive/aws-legacy/` 配下）。実行手順として使用しない。

---

## 📂 ディレクトリ構成

ファイル索引の正本は [README.md](README.md)（二重管理を避けるため本書には索引を置かない）。

---

## 🛠 運用の原則

1.  **疎結合の維持**: Cloudflare（バックエンド）と Vercel（フロントエンド）のエッジ配信を独立させ、障害時の影響範囲を最小化する。
2.  **機密情報の保護**: 秘密鍵やパスワードはソースコードに含めず、必ず `wrangler secret put`（Cloudflare Secrets）または Vercel Secrets を使用する。
3.  **証跡の記録**: 全てのデプロイと大規模な構成変更は、ランブックに基づき実行ログを残す。

---
