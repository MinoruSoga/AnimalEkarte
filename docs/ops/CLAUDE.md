# インフラ・運用ディレクトリ (Infrastructure & Operations)

> **目的**: クラウド基盤（Cloudflare Workers + Containers / Vercel。AWS ECS は Phase 8 完了までのロールバック専用）、デプロイフロー、およびシステムの安定稼働に関する情報の管理。
> **読者**: AI エージェント(Claude Code)。
> **タイミング**: docs/ops/ 配下編集時。

---

## 📂 ディレクトリ構成

ファイル索引の正本は [README.md](README.md)（二重管理を避けるため本書には索引を置かない）。

---

## 🛠 運用の原則

1.  **疎結合の維持**: Cloudflare（バックエンド）と Vercel（フロントエンド）のエッジ配信を独立させ、障害時の影響範囲を最小化する。AWS ECS/RDS はロールバック経路として分離済み。
2.  **機密情報の保護**: 秘密鍵やパスワードはソースコードに含めず、必ず `wrangler secret put`（Cloudflare Secrets）または Vercel Secrets を使用する。AWS ECS ロールバック経路のみ SSM Parameter Store を使用する。
3.  **証跡の記録**: 全てのデプロイと大規模な構成変更は、ランブックに基づき実行ログを残す。

---
