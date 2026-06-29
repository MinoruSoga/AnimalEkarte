# ADR-001: システムアーキテクチャ基本設計

**Status**: Accepted
**Date**: 2026-04-01
**Deciders**: MinoruSoga

## Context

動物病院向け電子カルテシステム (AnimalEkarte) のアーキテクチャ基盤を決定する必要があった。
マルチクリニック対応・LINE連携・会計機能・CPM(顧客育成)を単一システムで運用する。

## Decision

以下の技術スタックを採用する。

| レイヤ | 技術選定 | 理由 |
|--------|---------|------|
| フロントエンド | React 19 / TypeScript 5.7 / Vite 6 / Tailwind CSS 4 | Server Actions / useActionState による型安全フォーム。shadcn/ui で一貫した UI コンポーネント |
| バックエンド | Go 1.25 / Gin / GORM | 型安全性・静的解析・高スループット。P1-P18 の handler→service→repository 3層分離 |
| データベース | PostgreSQL 18 | clinic_id による完全マルチテナント隔離。JSONB・行レベルセキュリティ対応 |
| インフラ | Docker Compose (dev) / AWS ECS (staging/prod) / Vercel (frontend) | 開発環境の再現性とステージング/本番の分離 |
| CI | GitHub Actions | backend (Go build/test/lint) + frontend (pnpm lint/build) を PR ごとに検証 |

## Consequences

**ポジティブ:**
- clinic_id スコープをリポジトリ層の `clinicScope` (GORM Scopes) で強制し、クロステナントデータ漏洩を構造的に防止（詳細は ADR-002）
- Go の静的型付けにより API レスポンス型の TypeScript 自動生成 (tygo) が可能
- Vercel Preview Deployments で PR ごとのフロントエンド確認が容易

**ネガティブ:**
- Docker 必須の開発環境はローカル直実行より起動が遅い
- Go と TypeScript の 2 言語体制でコンテキストスイッチコストが発生

## References

- [docs/architecture.md](../architecture.md)
- [docs/ERD.md](../ERD.md)
- [backend/CLAUDE.md](../../backend/CLAUDE.md)
- [frontend/CLAUDE.md](../../frontend/CLAUDE.md)
