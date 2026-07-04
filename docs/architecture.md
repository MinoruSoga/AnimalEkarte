# システムアーキテクチャ (Architecture)

> **目的**: レイヤードアーキテクチャ(handler→service→repository→model)の責務分離を定義する。
> **読者**: 新規参加開発者・アーキテクチャ判断を行う開発者。
> **タイミング**: オンボーディング時、または層をまたぐ設計判断が必要な実装前。

> **Animal Ekarte**: 高信頼・高拡張な動物病院管理システム
> **最新更新**: 2026-06-12 | **ステータス**: Production Ready

---

## 1. 設計思想

本システムは、**「最速の機能提供」と「長期的な保守性」の両立**を目的として、**軽量レイヤードアーキテクチャ**を採用しています。

### 核心となる原則
- **Single Source of Truth (SSOT)**: バックエンドの Go モデルを唯一の真実とし、フロントエンドの型定義を自動生成。
- **カプセル化と依存性の逆転**: 各機能（Feature）を独立させ、Feature 間の直接参照を禁止。
- **データ分離の徹底**: 全エンドポイントで `clinic_id` による厳格なマルチテナント分離。

---

## 2. バックエンド・アーキテクチャ (Go 1.25)

### 層の責務分離

| 層 (ディレクトリ) | 責務 | 依存方向 |
|:---|:---|:---|
| **`handler/`** | HTTP/JSON の受付・返却。権限チェック（RBAC）。 | → `service` |
| **`service/`** | 業務ロジックの核心。バリデーション、他サービス連携、集計。 | → `repository`, `infra` |
| **`repository/`** | DB (GORM) 操作の抽象化。センチネルエラーへの変換。 | → `model` |
| **`model/`** | DB スキーマ定義。SSOT としての構造体。 | (依存なし) |
| **`infra/`** | LINE API、S3 ストレージ、外部サービスとの低レイヤ連携。 | (依存なし) |

### 規模と実績
- **実装規模**: 108 テーブル、88 ハンドラー（`backend/internal/handler/*_handler.go` ファイル数）、15 配信トリガー。

- **エラー処理**: `internal/errors` による統一されたセンチネルエラー体系。

---

## 3. フロントエンド・アーキテクチャ (React 19)

### Feature-Based 構造
`src/features/[feature]/` に以下の要素をカプセル化し、高度な独立性を維持しています（すべての Feature が全要素を持つわけではない）。
- **`api/`**: TanStack Query による API 通信。
- **`components/`**: Feature 内専用の UI 部品。
- **`routes/`**: ルーティング対象のページコンポーネント。
- **`hooks/`**: Feature 特有のステート・ロジック。
- **`types/`**: ドメイン固有の型定義。

### 合成とページ構成
各 Feature は主に **`src/app/routes/`**（機能カテゴリ別のルート定義ファイル群。lazy import による主要な合成点）で合成され、一部の個別ページは **`src/app/pages/`** のラッパー経由で合成されます。これにより、ある Feature の変更が別の Feature に予期せぬ影響を与えることを防ぎます。

### React 19 実装パターン
- **`useActionState`**: サーバーアクション（保存・更新）の標準。
- **`useTransition` / `useDeferredValue`**: 高負荷な一覧表示やフィルタリングの最適化。
- **`ref as prop`**: コンポーネント間の連携を簡素化。

---

## 4. インフラ・デプロイ構成

### 技術スタック
- **Runtime**: ECS Fargate (Go), Vercel (React)
- **Database**: RDS PostgreSQL 18
- **Storage**: AWS S3 (領収書、検査結果、証明書)
- **Messaging**: LINE Messaging API / Lステップ API

### セキュリティ
- **認証**: JWT + httpOnly Cookie によるセキュアなセッション管理。
- **認可**: リソース単位の CRUD 権限管理 (RBAC)。
- **通信**: 全経路 TLS 1.3、S3 署名付き URL。

---

## 5. 将来の拡張方針

- **疎結合の維持**: 将来的なマイクロサービス化やサブシステム（トリミング専用機等）の切り出しを容易にする設計。
- **自動化**: `make codegen` による型同期、CI/CD による自動テストの徹底。

---
