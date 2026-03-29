# 📚 ドキュメント目次

Animal Ekarte（動物病院電子カルテシステム）の技術ドキュメント統一版です。

このディレクトリは**単一の信頼できる情報源（SSOT）**として機能し、プロジェクト全体の仕様・設計・ガイドラインを統一的に管理しています。

---

## 📖 公式ドキュメント

### 仕様・設計

| ドキュメント | 説明 | 対象者 |
|-------------|------|--------|
| **[SPECIFICATION.md](./SPECIFICATION.md)** | 🎯 技術仕様書（技術スタック、アーキテクチャ、Feature一覧、ルーティング）⚠️初期設計版 | エンジニア全員 |
| **[DESIGN_SYSTEM.md](./DESIGN_SYSTEM.md)** | 🎨 デザインシステム v3.0（Notionトークン、D&D、アクセシビリティ） | フロントエンド開発者 |
| **[AUTH.md](./AUTH.md)** | 🔐 認証・認可設計（RBAC、マルチクリニック、権限マトリクス） | エンジニア全員 |
| **[ERD.md](./ERD.md)** | 🗄️ ER図 v27.0（全57テーブル） | バックエンド開発者 |
| **[ATTRIBUTIONS.md](./ATTRIBUTIONS.md)** | 📄 ライセンス・帰属情報 | プロジェクト管理者 |

### 画面仕様

> **主要参照先**: `docs/screens/` 個別ファイル（アクティブ更新中）
> 以下の旧モノリシック仕様書はアーキテクチャ参照用として保持しています。

| ドキュメント | 説明 | 対象者 |
|-------------|------|--------|
| **[screens/](./screens/)** | 📱 ★ **現行** 画面仕様（画面ごとの個別ファイル、API連携含む） | フロントエンド開発者 |
| **[SCREENS.md](./SCREENS.md)** | 📱 画面コンポーネント構成（`[R]`/`[S]`/`[C]`記法、アーキテクチャ参照用） | フロントエンド開発者 |
| **[SCREENS_MASTER.md](./SCREENS_MASTER.md)** | 📱 マスタ管理画面仕様 v3.0（全16カテゴリ詳細） | フロントエンド開発者 |
| **[SCREENS_DETAILED_TABS.md](./SCREENS_DETAILED_TABS.md)** | 📋 電子カルテタブ詳細仕様（TreatmentTable等の実装詳細） | フロントエンド開発者 |
| **[SCREENS_VALIDATION.md](./SCREENS_VALIDATION.md)** | ✅ バリデーション・エラーメッセージ詳細 ⚠️TS型定義はmodels.ts参照 | フロントエンド開発者 |
| **[FORMS_SPECIFICATION.md](./FORMS_SPECIFICATION.md)** | 📝 フォームフィールド詳細仕様（必須/任意・バリデーション・デフォルト値） | フロントエンド開発者 |

### 実装ガイド

| ドキュメント | 説明 | 対象者 |
|-------------|------|--------|
| **[infra/deploy/](./infra/deploy/)** | 🚀 デプロイメント（CI/CD、環境設定） | DevOps / 本番管理者 |
| **[infra/](./infra/)** | ☁️ インフラストラクチャ（AWS、Terraform） | インフラエンジニア |

---

## 🔗 API 仕様

実装済みAPIの完全な仕様は OpenAPI 仕様書で確認できます：

- **ファイル:** `backend/docs/api.yaml`
- **サンプル:** `backend/docs/api-examples.md`
- **Postman:** `backend/docs/postman-collection.json`

---

## 🎨 画面一覧

画面ごとの仕様は `docs/screens/` の個別ファイルで管理しています。

| 機能 | ルート | 説明 |
|------|--------|------|
| **ログイン** | `/login` | 認証（デモアカウント対応、PermissionGate） |
| **ダッシュボード** | `/` | 来院フローの俯瞰管理（カンバンボード） |
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

詳細は [screens/](./screens/) の個別ファイルを参照。コンポーネント構成は [SCREENS.md](./SCREENS.md) を参照。

---

## 🗄️ データベース設計

### 概要

- **ER図**: [ERD.md](./ERD.md)
- **テーブル数**: 57（詳細は ERD.md v27.0 参照）
- **マスタカテゴリ**: 16種類

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
| **在庫** | 1 | `inventory_items` |
| **シフト** | 1 | `shift_entries` |
| **認証** | 8 | `company`, `clinics`, `user_accounts`, `user_clinic_memberships`, `job_titles`, `permission_groups`, `permission_group_rules`, `user_permission_groups` |
| **マスタ（16テーブル）** | 16 | `examination_types`, `vaccines`, `medicines`, `staffs`, `insurances`, `cages`, `service_types`, `consultations`, `procedures`, `hospitalization_plans`, `trimming_courses`, `trimming_options`, `diagnosis_categories`, `diagnosis_names`, `checkup_types`, `examination_type_items` |

詳細は [ERD.md](./ERD.md) を参照。

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
│   │   ├── features/        # 機能別モジュール
│   │   ├── hooks/           # 共有フック
│   │   ├── lib/             # ユーティリティ
│   │   └── types/           # グローバル型定義
│   ├── CODING_RULES.md      # React 19 詳細ルール
│   └── ...
├── backend/                  # Go + Gin + GORM
│   ├── cmd/api/             # エントリーポイント
│   ├── internal/            # ビジネスロジック
│   ├── migrations/          # DB マイグレーション（SQL）
│   ├── docs/                # API仕様（api.yaml、postman）
│   ├── CODING_RULES.md      # Go 詳細ルール
│   └── ...
├── docs/                     # 📍 このディレクトリ
│   ├── SPECIFICATION.md     # 技術仕様書
│   ├── DESIGN_SYSTEM.md     # UI デザイン（v3.0）
│   ├── SCREENS.md           # 画面仕様（v3.0、全42ルート）
│   ├── AUTH.md              # 認証・認可設計（RBAC）
│   ├── ERD.md               # データベース ER 図（v27.0、57テーブル）
│   ├── SCREENS_MASTER.md    # マスタ管理画面仕様
│   ├── SCREENS_VALIDATION.md # バリデーション・データ型仕様
│   ├── SCREENS_README.md    # 画面仕様書ガイド
│   ├── FORMS_SPECIFICATION.md # フォーム仕様書
│   ├── ATTRIBUTIONS.md      # ライセンス・帰属情報
│   └── infra/               # インフラ・デプロイ
│       └── deploy/          # CI/CD・デプロイ手順
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

---

## 📞 サポート・問い合わせ

### ドキュメント不足について

- **仕様明確化**: [SPECIFICATION.md](./SPECIFICATION.md) を確認
- **デザイン詳細**: [DESIGN_SYSTEM.md](./DESIGN_SYSTEM.md) を確認
- **画面仕様**: [screens/](./screens/) の該当画面ファイルを確認
- **API詳細**: `backend/docs/api.yaml` を確認

### 関連ドキュメント

| リンク | 説明 |
|--------|------|
| [CLAUDE.md](../CLAUDE.md) | プロジェクト全体の指示・ルール |
| [CODING_RULES.md](../CODING_RULES.md) | コーディング規約（全体） |
| [frontend/CODING_RULES.md](../frontend/CODING_RULES.md) | React 19 詳細ルール |
| [backend/CODING_RULES.md](../backend/CODING_RULES.md) | Go / Gin 詳細ルール |

---

**最終更新**: 2026年3月14日 | **バージョン**: 3.2
