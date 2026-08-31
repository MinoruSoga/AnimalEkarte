# ADR-001: システムアーキテクチャ基本設計

**Status**: Accepted; backend package-architecture portion superseded by [ADR-005](005-go-gin-backend-guidelines.md)

**Date**: 2026-04-01
**Deciders**: MinoruSoga

## Context

動物病院向け電子カルテシステム (AnimalEkarte) のアーキテクチャ基盤を決定する必要があった。
マルチクリニック対応・LINE連携・会計機能・CPM(顧客育成)を単一システムで運用する。

## Decision

以下は **2026-04-01 の採用時 snapshot**。現行 version は各 package manifest / toolchain 設定を正本とする。

| レイヤ | 技術選定 | 理由 |
|--------|---------|------|
| フロントエンド | React 19 / TypeScript 5.7 / Vite 6 / Tailwind CSS 4 | Server Actions / useActionState による型安全フォーム。shadcn/ui で一貫した UI コンポーネント |
| バックエンド | Go 1.25 / Gin / GORM | 型安全性・静的解析・高スループット。当時は project 固有の3層構成も同時採用した |
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

## Supersession note

技術スタックの選定履歴は有効だが、Go/Gin公式要件として project 固有の3層構成を強制する判断は ADR-005 により廃止した。package architecture は公式ガイドが規定するものと、application 固有の判断を分離して扱う。

インフラ欄は当時の決定記録であり、AWS ECS/RDSは2026-07-20に廃止済み。現行構成の正本は [docs/ops/infra/architecture.md](../../ops/infra/architecture.md) とする。

## References

- [docs/architecture/overview.md](../overview.md)
- [docs/architecture/erd.md](../erd.md)
- [backend/CLAUDE.md](../../../backend/CLAUDE.md)
- [frontend/CLAUDE.md](../../../frontend/CLAUDE.md)
