# 📚 ドキュメント目次

Animal Ekarte（動物病院電子カルテシステム）の技術ドキュメント統一版です。

このディレクトリは**単一の信頼できる情報源（SSOT）**として機能し、プロジェクト全体の仕様・設計・ガイドラインを統一的に管理しています。

---

## 📖 公式ドキュメント

### 仕様・設計

| ドキュメント | 説明 | 対象者 |
|-------------|------|--------|
| **[SPECIFICATION.md](./SPECIFICATION.md)** | 🎯 技術仕様書（技術スタック、アーキテクチャ、機能一覧） | エンジニア全員 |
| **[DESIGN_SYSTEM.md](./DESIGN_SYSTEM.md)** | 🎨 デザインシステム（Notion準拠トークン、コンポーネント仕様） | フロントエンド開発者 |
| **[SCREENS.md](./SCREENS.md)** | 📱 画面仕様書（全42ルートのコンポーネント・UX仕様） | フロントエンド開発者 |
| **[AUTH.md](./AUTH.md)** | 🔐 認証・認可設計（RBAC、マルチクリニック、権限マトリクス） | エンジニア全員 |
| **[ERD.md](./ERD.md)** | 🗄️ ER図（データベース設計、エンティティ定義） | バックエンド開発者 |
| **[DB_DEFINITION.md](./DB_DEFINITION.md)** | 📊 DB定義書（全31テーブル、RLSポリシー、インデックス） | バックエンド開発者 |
| **[ATTRIBUTIONS.md](./ATTRIBUTIONS.md)** | 📄 ライセンス・帰属情報 | プロジェクト管理者 |

### 実装ガイド

| ドキュメント | 説明 | 対象者 |
|-------------|------|--------|
| **[MIGRATION.md](./MIGRATION.md)** | 🔧 マイグレーション（DB作成・管理） | バックエンド開発者 |
| **[API-ROADMAP.md](./API-ROADMAP.md)** | 🛣️ API設計ロードマップ（未実装エンドポイント） | バックエンド開発者 |
| **[deploy/](./deploy/)** | 🚀 デプロイメント（CI/CD、環境設定） | DevOps / 本番管理者 |
| **[infra/](./infra/)** | ☁️ インフラストラクチャ（AWS、Terraform） | インフラエンジニア |

---

## 🔗 API 仕様

### ✅ 実装済みAPI（Swagger）

実装済みAPIの完全な仕様は Swagger UI で確認できます：

- **URL:** http://localhost:8080/swagger/index.html
- **ファイル:** `backend/docs/swagger.yaml`

```bash
# Swagger 再生成（Go code を変更後）
docker compose exec backend swag init -g cmd/api/main.go
```

### 🗺️ 未実装API（設計のみ）

将来実装予定のエンドポイント一覧：

- [API-ROADMAP.md](./API-ROADMAP.md) - エンドポイント設計書

---

## 🎨 画面一覧

全42ルートの仕様を `SCREENS.md` で統合管理しています。

| 機能 | ルート | 説明 |
|------|--------|------|
| **ログイン** | `/login` | 認証（デモアカウント対応、PermissionGate） |
| **ダッシュボード** | `/` | 来院フローの俯瞰管理（カンバンボード + DashboardSummaryWidget） |
| **予約管理** | `/reservations/*` | MonthView / WeekView カレンダー（D&D対応） |
| **飼主・ペット管理** | `/owners`, `/owners/:id` | 顧客とペット情報 |
| **電子カルテ** | `/medical-records/*` | 9タブのMedicalRecordForm |
| **入院管理** | `/hospitalization/*` | Board/List表示、CarePlanDialog、DailyRecord |
| **会計** | `/accounting/*` | InsuranceRatio、PaymentPanel |
| **トリミング** | `/trimming/*` | 3カラムレイアウト、HistoryFilterPanel |
| **検査管理** | `/examinations` | 読み取り専用リスト |
| **予防接種** | `/vaccinations` | 読み取り専用リスト |
| **定期健診** | `/checkups` | 3種健診タイプ |
| **在庫管理** | `/inventory` | SortableHeader、InventoryForm |
| **シフト管理** | `/shifts` | ShiftCalendar、ShiftEditPopover |
| **マスタ設定** | `/settings/*` | 15カテゴリのマスタデータ管理 |

詳細は [SCREENS.md](./SCREENS.md) を参照。

---

## 🗄️ データベース設計

### 概要

- **ER図**: [ERD.md](./ERD.md)
- **DB定義書**: [DB_DEFINITION.md](./DB_DEFINITION.md)
- **テーブル数**: 31（コア 27 + 認証 4）
- **マスタカテゴリ**: 15種類

### テーブル構成

| 分類 | テーブル数 | 主なテーブル |
|------|-----------|------|
| **飼主・ペット** | 2 | `owners`, `pets` |
| **予約** | 1 | `reservation_appointments` |
| **カルテ** | 5 | `medical_records`, `treatment_items`, `vital_entries`, `examination_records`, `examination_record_items` |
| **入院** | 6 | `hospitalizations`, `care_plan_items`, `daily_records`, `vital_records`, `care_log_records`, `staff_note_records` |
| **トリミング** | 2 | `trimming_records`, `trimming_record_options` |
| **会計** | 3 | `accountings`, `accounting_items`, `payment_infos` |
| **予防接種・健診** | 3 | `vaccination_records`, `checkup_records`, `treatment_plans` |
| **マスタ・在庫** | 3 | `master_items`, `master_item_inspections`, `inventory_items` |
| **シフト** | 1 | `shift_entries` |
| **クリニック** | 1 | `clinic_info` |
| **認証（Phase 4）** | 4 | `clinics`, `user_accounts`, `user_clinic_memberships`, `user_permissions` |

詳細は [DB_DEFINITION.md](./DB_DEFINITION.md) および [ERD.md](./ERD.md) を参照。

---

## 📋 コーディング規約

### フロントエンド

- **ファイル**: `frontend/CODING_RULES.md`
- **スタイル**: React 19 (FC 禁止、ref as prop)
- **型安全**: `any` 型禁止、型ガード必須
- **制御フロー**: switch 文禁止、Record lookup 推奨

### バックエンド

- **ファイル**: `backend/CODING_RULES.md`
- **エラーハンドリング**: センチネルエラー + ラッピング
- **Context**: 全関数の第一引数に `context.Context`
- **ログ**: slog 構造化ログ必須

### 全体

- **ファイル**: `CODING_RULES.md`
- **命名規則**: Go は PascalCase / snake_case、TypeScript は camelCase / kebab-case
- **コミット**: 英語 / 日本語可、フォーマット統一

---

## 🏗️ プロジェクト構成

```
AnimalEkarte/
├── frontend/                 # React 19 + TypeScript
│   ├── src/
│   │   ├── app/             # ルーター・プロバイダ
│   │   ├── components/      # UI・共有コンポーネント
│   │   ├── features/        # 機能別モジュール（15個 + auth予定）
│   │   ├── hooks/           # 共有フック
│   │   ├── lib/             # ユーティリティ（design-tokens.ts含む）
│   │   └── types/           # グローバル型定義
│   ├── CODING_RULES.md      # React 19 詳細ルール
│   └── ...
├── backend/                  # Go + Gin + GORM
│   ├── cmd/api/             # エントリーポイント
│   ├── internal/            # ビジネスロジック（再実装予定）
│   ├── migrations/          # DB マイグレーション
│   ├── CODING_RULES.md      # Go 詳細ルール
│   └── ...
├── docs/                     # 📍 このディレクトリ
│   ├── SPECIFICATION.md     # 技術仕様書
│   ├── DESIGN_SYSTEM.md     # UI デザイン（v3.0）
│   ├── SCREENS.md           # 画面仕様（v3.0、全42ルート）
│   ├── AUTH.md              # 認証・認可設計（RBAC）
│   ├── ERD.md               # データベース ER 図
│   ├── DB_DEFINITION.md     # DB 定義書（全31テーブル）
│   ├── MIGRATION.md         # マイグレーション
│   ├── API-ROADMAP.md       # API設計
│   ├── deploy/              # デプロイメント
│   ├── infra/               # インフラ
│   └── README.md            # このファイル
├── CLAUDE.md                # プロジェクト指示
├── CODING_RULES.md          # 全体ルール
└── docker-compose.yml       # Docker設定
```

---

## 🚀 クイックスタート

### 環境起動

```bash
# コンテナ起動
make up

# ログ確認
make logs

# 停止
make down
```

### 開発

```bash
# Frontend ビルド
docker compose exec frontend npm run build

# Frontend Lint
docker compose exec frontend npm run lint

# Backend テスト
docker compose exec backend go test ./... -v

# Backend Lint
docker compose exec backend golangci-lint run ./...
```

### データベース

```bash
# DB 接続
make db

# マイグレーション確認
\d  # psql コマンド
```

詳細は [MIGRATION.md](./MIGRATION.md) を参照。

---

## 📞 サポート・問い合わせ

### ドキュメント不足について

- **仕様明確化**: [SPECIFICATION.md](./SPECIFICATION.md) を確認
- **デザイン詳細**: [DESIGN_SYSTEM.md](./DESIGN_SYSTEM.md) を確認
- **画面仕様**: [SCREENS.md](./SCREENS.md) から該当画面を検索
- **API詳細**: Swagger UI ( http://localhost:8080/swagger/index.html ) を確認

### 関連ドキュメント

| リンク | 説明 |
|--------|------|
| [CLAUDE.md](../CLAUDE.md) | プロジェクト全体の指示・ルール |
| [CODING_RULES.md](../CODING_RULES.md) | コーディング規約（全体） |
| [frontend/CODING_RULES.md](../frontend/CODING_RULES.md) | React 19 詳細ルール |
| [backend/CODING_RULES.md](../backend/CODING_RULES.md) | Go / Gin 詳細ルール |
| [backend/README.md](../backend/README.md) | バックエンド README |

---

## 📅 更新履歴

- **2026-03-12**: ui-sample/src から v3.0 ドキュメントへ全面更新
  - SCREENS.md（v3.0 — 全42ルート、15セクション）
  - DESIGN_SYSTEM.md（v3.0 — Notionトークン、D&D、アクセシビリティ）
  - AUTH.md（新規作成 — RBAC、マルチクリニック、権限マトリクス）
  - README.md（v3.0 対応に更新）

- **2026-03-11**: DB設計見直し・バックエンド全層削除・再構築
  - DB_DEFINITION.md（31テーブル完全スキーマ）
  - ERD.md（Mermaid ERD 更新）

- **2026-03-06**: ui-sample/src から正式版ドキュメントを統合作成
  - SPECIFICATION.md（技術仕様）
  - DESIGN_SYSTEM.md（UIデザイン v1.0）
  - SCREENS.md（画面仕様 v1.0）
  - ERD.md（データベース v1.0）
  - ATTRIBUTIONS.md（ライセンス）

---

**最終更新**: 2026年3月12日 | **バージョン**: 3.0
