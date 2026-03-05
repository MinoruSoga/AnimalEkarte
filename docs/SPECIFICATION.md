# 動物病院管理システム 仕様定義書

## 1. プロジェクト概要
本プロジェクトは、ReactとTailwind CSSを用いたモダンなWebベースの動物病院向け統合管理システム（PMS: Practice Management System）のプロトタイプです。
予約管理、受付、電子カルテ、入院管理、会計までの一連の業務フローをシームレスに連携させ、獣医療現場の業務効率化と質の高い医療提供を支援することを目的としています。

## 2. 技術スタック

### フロントエンド
- **Core**: React 19, TypeScript 5.7
- **Build Tool**: Vite 6
- **Routing**: React Router (`react-router`, `createBrowserRouter` + `RouterProvider` 構成)
  - **注意**: `react-router-dom` ではなく `react-router` からimportすること
- **State Management**: React Hooks (Custom Hooks) / TanStack Query

### Backend
- **Runtime**: Go 1.25
- **Framework**: Gin
- **Database ORM**: GORM
- **Database**: PostgreSQL 18

### UI/UX & デザインシステム
- **CSS Framework**: Tailwind CSS 4
- **Component Library**: shadcn/ui (Radix UI Primitives ベース)
    - プロジェクト内にソースコードとして配置 (`/components/ui`)
- **Design System**: Notion Like Design System (詳細は `/docs/DESIGN_SYSTEM.md` を参照)
- **Icons**: Lucide React
- **Charts**: Recharts (売上分析、バイタルグラフ等)
- **Drag & Drop**: react-dnd + react-dnd-touch-backend (カンバンボード)
- **Animations**: Motion (`motion/react`) / CSS Transitions
- **Toast**: Sonner (`sonner@2.0.3`)

### フォーム & データハンドリング
- **Form Management**: React Hook Form (`react-hook-form@7.55.0`)
- **Date Utility**: date-fns
- **API Integration**: TanStack Query (React Query)

### Infrastructure
- **Containerization**: Docker Compose
- **Database**: PostgreSQL 18
- **Web Server**: Gin (Go)

## 3. アーキテクチャ

### Frontend
「bulletproof-react」の設計思想に基づいた **Feature-based Architecture** を採用しています。機能（Feature）ごとにディレクトリを分割し、関心の分離とスケーラビリティを確保しています。

#### ディレクトリ構造
```
frontend/src/
├── app/           # ルーター、プロバイダ
├── components/    # ui/, shared/, layouts/, errors/
├── features/      # 機能別モジュール
├── hooks/         # 共有カスタムフック
├── lib/           # ユーティリティ、定数
├── types/         # グローバル型定義
├── styles/        # Tailwind CSS v4
└── testing/       # テスト設定
```

#### Barrel Index パターン
全13 featureに `api/index.ts` と `types/index.ts` のbarrel indexを整備済みです。
- **型import**: 常に `../../{feature}/types` barrel経由
- **API import**: 常に `../../{feature}/api` barrel経由
- **サブモジュール直接参照禁止**: `../types/diagnosis` のような直接参照は不可

### Backend
Go + Gin + GORMを用いたクリーンアーキテクチャ実装

#### ディレクトリ構造
```
backend/
├── cmd/api/       # エントリーポイント
├── internal/
│   ├── config/    # 設定
│   ├── errors/    # センチネルエラー
│   ├── handler/   # HTTPハンドラ
│   ├── logger/    # slog構造化ログ
│   ├── middleware/# ミドルウェア
│   ├── model/     # ドメインモデル
│   ├── repository/# データアクセス
│   ├── service/   # ビジネスロジック
│   └── validation/# バリデーション
├── migrations/    # DBマイグレーション
└── docs/          # Swagger
```

## 4. コード品質規約

### Frontend - 型安全性
- **`any` 型ゼロ**: 全ファイルで `any` 型の使用を禁止
- **インライン `as const` ゼロ**: 定数配列は `types/index.ts` 等で事前定義
- **`as "..."` 型断言ゼロ**: 型キャストではなく型ガードや `typedSetter` を使用
- **型定義の配置**: コンポーネント/hookファイルからの型exportは完全禁止。`types/index.ts` barrel経由に統一

### Frontend - 制御フロー
- **switch文ゼロ**: Record マッピングまたは lookup テーブルで代替
- **`else if` ゼロ**: 早期 return、三項演算子、Record lookup で代替
- **値列挙パターン**: `const XXX_VALUES = [...] as const` → `type Xxx = (typeof XXX_VALUES)[number]` → `Record<Xxx, string>` ラベルマップの三点セット

### Frontend - Component
- **React 19**: `FC`禁止、`forwardRef`禁止、ref as prop
- **bulletproof-react**: feature間import禁止、barrel file禁止
- **フォーム保護**: `useBlocker` + `NavigationBlockerDialog` で未保存変更検知

### Backend - エラーハンドリング
- **センチネルエラー**: センチネルエラー + ラッピング
- **Context伝播**: 全関数の第一引数に`context.Context`
- **GORMでのContext使用**: `r.db.WithContext(ctx).First(&pet, "id = ?", id)`

### Backend - ログ
- **slog構造化ログ**: `slog.InfoContext()`, `slog.ErrorContext()` を使用
- **ログ出力**: 敏感情報（パスワード、トークン）を出力しない

## 5. 共有コンポーネント一覧 (`/components/shared/`)

### レイアウト
| コンポーネント | 説明 |
|---|---|
| `PageLayout` | ページ全体のレイアウト (ヘッダー + スクロール可能コンテンツ) |
| `FormHeader` | ページヘッダー (タイトル, 戻るボタン, アクション) |

### データ表示
| コンポーネント | 説明 |
|---|---|
| `DataTable` | 汎用テーブル (カラム定義 + renderRow) |
| `DataTableRow` | テーブル行 (ホバー/クリック対応) |
| `Pagination` | ページネーション |
| `StatusBadge` | ステータスバッジ |
| `PatientInfoCard` | 患者(ペット)情報カード |

### フォーム・入力
| コンポーネント | 説明 |
|---|---|
| `PrimaryButton` | プライマリアクションボタン (`bg-[#37352F]`) |
| `SearchFilterBar` | 検索バー + 件数表示 |
| `MasterSelectModal` | マスタ項目選択モーダル |
| `MasterLink` | マスタへのリンクナビゲーション |
| `NotionDatePicker` | Notion風日付ピッカー |
| `PetSearchForm` | ペット検索フォーム |
| `PetSearchResultsTable` | ペット検索結果テーブル |
| `PetSelection` | ペット選択ページ共通コンポーネント |

### フィードバック・確認
| コンポーネント | 説明 |
|---|---|
| `ConfirmDialog` | 確認ダイアログ (destructive対応) |
| `NavigationBlocker` | フォーム離脱保護ダイアログ |
| `StaffImpactDialog` | スタッフ変更影響確認ダイアログ |

### アクション
| コンポーネント | 説明 |
|---|---|
| `RowActionButton` | テーブル行アクションボタン |
| `RowActionDropdown` | テーブル行ドロップダウンメニュー |

### 履歴・フィルタ
| コンポーネント | 説明 |
|---|---|
| `HistoryPanel` | 履歴一覧パネル |
| `HistoryFilterPanel` | 履歴フィルタリングパネル |

### 共有フック (`/hooks/`)
| フック | 説明 |
|---|---|
| `usePagination` | ページネーションロジック |
| `usePetSearch` | ペット検索ロジック |
| `useStaffUsageCount` | スタッフ使用件数カウント |
| `useStaffValidation` | スタッフ関連バリデーション |
| `useUnsavedChanges` | 未保存変更検知 |

## 6. 機能一覧 (Features)

### 6.1 ダッシュボード (`/features/dashboard`)
- **概要**: 病院全体の稼働状況を俯瞰するホーム画面。カンバンボード形式で患者の来院フローを管理。
- **ワークフロー（カラム構成）**:
    1. **受付予約**: 本日来院予定の予約
    2. **受付済**: 来院済み・待合室待機中
    3. **診療中**: 診察室・処置室で対応中
    4. **会計待ち**: 診察終了・会計計算待ち
    5. **会計済**: 会計完了・帰宅

### 6.2 予約管理 (`/features/reservations`)
- **概要**: 診療・トリミング・手術の予約を一元管理。
- **予約タイプ**: 診療、検診、検査、手術、トリミング、ワクチン、入院、ホテル
- **機能**:
    - カレンダー表示（月表示 / 週表示）
    - 予約の新規作成・編集・キャンセル
    - 患者（ペット）検索と予約紐付け
    - 担当獣医師の割り当て

### 6.3 電子カルテ (`/features/medical-records`)
- **概要**: タブ形式で診療記録の各セクションを管理。
- **ステータス**: 作成中 → 確定済
- **タブ構成**: 問診、診察/治療プラン、治療、予防接種、検査、画像、見積書、会計(医師確認)

### 6.4 入院管理 (`/features/hospitalization`)
- **概要**: 入院患者のケアプラン作成と日々の記録管理。
- **機能**: 入院ボード、リストビュー、ケアプラン、デイリーログ、タイムライン、コスト管理

### 6.5 会計 (`/features/accounting`)
- **概要**: 診療費の計算と請求書発行。
- **ステータス**: 未収(waiting)、収済(completed)、キャンセル(cancelled)、保留(pending)

### 6.6 その他の主要機能
- **検査管理** (`/features/examinations`): 院内・院外検査のオーダーと結果管理
- **顧客・ペット管理** (`/features/owners`, `/features/pets`): 飼い主とペットの基本情報管理
- **トリミング** (`/features/trimming`): トリミング業務の管理
- **ワクチン管理** (`/features/vaccinations`): 予防接種の管理
- **設定・マスタ管理** (`/features/master`, `/features/clinic`): システム全体設定とマスタデータ管理（14カテゴリ）

## 7. ルーティング構成

```
/                              → Dashboard (カンバンボード)
/owners                        → 飼主一覧
/owners/new                    → 飼主新規登録
/owners/:id                    → 飼主編集
/reservations                  → 予約管理 (カレンダー)
/medical-records               → カルテ一覧
/medical-records/select-pet    → カルテ新規 - ペット選択
/medical-records/new           → カルテ新規作成
/medical-records/:id           → カルテ編集
/hospitalization               → 入院一覧
/hospitalization/select-pet    → 入院新規 - ペット選択
/hospitalization/new           → 入院新規登録
/hospitalization/:id           → 入院詳細
/hospitalization/:id/edit      → 入院編集
/trimming                      → トリミング一覧
/trimming/select-pet           → トリミング新規 - ペット選択
/trimming/new                  → トリミング新規登録
/trimming/:id                  → トリミング編集
/examinations                  → 検査管理
/accounting                    → 会計一覧
/accounting/select-pet         → 会計新規 - ペット選択
/accounting/new                → 会計新規作成
/accounting/:id                → 会計詳細
/vaccinations                  → 予防接種一覧
/settings                      → マスタ設定トップ (カテゴリカード一覧)
/settings/clinic               → 病院情報設定
/settings/{category-slug}      → 各マスタカテゴリ設定
```

## 8. 関連ドキュメント

| ドキュメント | パス | 説明 |
|---|---|---|
| **デザインシステム** | `/docs/DESIGN_SYSTEM.md` | カラーパレット、タイポグラフィ、コンポーネントスタイリング規約 |
| **画面仕様書** | `/docs/SCREENS.md` | 全画面のルート・構成・データフロー・操作仕様 |
| **ER図** | `/docs/ERD.md` | 全26エンティティの定義・リレーション・列挙型一覧 |
| **ライセンス** | `/docs/ATTRIBUTIONS.md` | 使用ライブラリのクレジット |
| **コーディング規約** | `/CODING_RULES.md` | 全体コーディング規約 |
| **Frontend規約** | `/frontend/CODING_RULES.md` | React 19 / TypeScript詳細 |
| **Backend規約** | `/backend/CODING_RULES.md` | Go / Gin / GORM詳細 |
