# 動物病院管理システム 仕様定義書

**更新日:** 2026-03-12
**バージョン:** 3.0（ui-sample/src/SPECIFICATION.md 統合版）

---

## 1. プロジェクト概要

動物病院向け統合管理システム（PMS: Practice Management System）。
予約管理、受付、電子カルテ、入院管理、会計までの一連の業務フローをシームレスに連携させ、
獣医療現場の業務効率化と質の高い医療提供を支援する。

---

## 2. 技術スタック

### フロントエンド

| 技術 | バージョン | 用途 |
|------|-----------|------|
| React | 19 | UIフレームワーク |
| TypeScript | 5.7 | 型安全性 |
| Vite | 6 | ビルドツール |
| Tailwind CSS | 4 | スタイリング |
| shadcn/ui | latest | コンポーネントライブラリ（Radix UI Primitives） |
| React Router | latest | ルーティング（`react-router` から import） |
| TanStack Query | latest | サーバー状態管理・APIキャッシュ |
| React Hook Form | 7.55.0 | フォーム管理 |
| react-dnd | latest | ドラッグ&ドロップ（カンバンボード） |
| Motion | latest | アニメーション（`motion/react`） |
| Sonner | 2.0.3 | トースト通知 |
| date-fns | latest | 日付ユーティリティ |

> **注意**: `react-router-dom` ではなく `react-router` からimportすること

### バックエンド

| 技術 | バージョン | 用途 |
|------|-----------|------|
| Go | 1.25 | バックエンドランタイム |
| Gin | latest | HTTPフレームワーク |
| GORM | latest | ORM |
| PostgreSQL | 18 | データベース |
| swaggo/swag | latest | Swagger自動生成 |

### インフラ

| 技術 | 用途 |
|------|------|
| Docker Compose | コンテナ管理 |
| PostgreSQL 18 | データベース |

### ポート

| サービス | ポート |
|---------|--------|
| Frontend | 3000 |
| Backend API | 8080 |
| PostgreSQL | 5432 |

---

## 3. アーキテクチャ

### フロントエンド（bulletproof-react準拠）

「bulletproof-react」の設計思想に基づいた **Feature-based Architecture**。

#### ディレクトリ構造

```
frontend/src/
├── app/           # ルーター、プロバイダ
├── components/
│   ├── ui/       # shadcn/ui（汎用UIパーツ）
│   ├── shared/   # アプリケーション固有の共有パーツ
│   ├── layouts/  # Layout, Sidebar
│   └── errors/   # ErrorBoundary
├── features/      # 機能別モジュール（15機能）
│   └── [feature]/
│       ├── api/         # API呼び出し + React Query hooks
│       ├── components/  # Feature固有UI
│       ├── hooks/       # ビジネスロジック・UI状態
│       ├── routes/      # ページコンポーネント
│       ├── types/       # 型定義（barrel export）
│       └── index.ts     # Public API
├── hooks/         # 共有カスタムフック
├── lib/           # ユーティリティ（design-tokens, format, type-utils等）
├── types/         # グローバル型定義
└── testing/       # テスト設定
```

#### Barrel Index パターン

全17 featureに `api/index.ts` と `types/index.ts` のbarrel index整備済み。
- **型import**: 常に `../../{feature}/types` barrel経由
- **API import**: 常に `../../{feature}/api` barrel経由
- **サブモジュール直接参照禁止**: `../types/diagnosis` のような直接参照は不可

### バックエンド（クリーンアーキテクチャ）

```
backend/
├── cmd/api/       # エントリーポイント
├── internal/
│   ├── config/    # 設定
│   ├── errors/    # センチネルエラー定義
│   ├── handler/   # HTTPハンドラ（Gin）
│   ├── logger/    # slog構造化ログ
│   ├── middleware/# ミドルウェア（CORS, logging）
│   ├── model/     # ドメインモデル（GORM）
│   ├── repository/# データアクセス
│   ├── service/   # ビジネスロジック
│   └── validation/# バリデーション
├── migrations/    # DBマイグレーション（001_init.sql, 002_seed.sql）
└── docs/          # Swagger
```

**依存関係**: handler → service → repository → model（上位層は下位層のみに依存）

---

## 4. コード品質規約

### フロントエンド - 型安全性

- **`any` 型ゼロ**: 全ファイルで `any` 型の使用を禁止
- **インライン `as const` ゼロ**: 定数配列は `types/index.ts` 等で事前定義
- **`as` 型断言ゼロ（feature/shared/hooks層）**: `lib/type-utils.ts`・`lib/suspense-utils.ts` 内部にのみ集約
- **型定義の配置**: コンポーネント/hookファイルからの型exportは完全禁止。`types/index.ts` barrel経由に統一

### フロントエンド - 制御フロー

- **switch文ゼロ**: Record マッピングまたは lookup テーブルで代替
- **`else if` ゼロ**: 早期 return、三項演算子、Record lookup で代替
- **値列挙パターン**: `const XXX_VALUES = [...] as const` → `type Xxx = (typeof XXX_VALUES)[number]` → `Record<Xxx, string>` ラベルマップ

### フロントエンド - コンポーネント

```typescript
// ✅ React 19 正しい書き方
interface PatientCardProps {
  patient: Patient;
  onSelect?: (id: string) => void;
  ref?: React.Ref<HTMLDivElement>;  // ref as prop
}
export function PatientCard({ patient, onSelect, ref }: PatientCardProps) {
  return <div ref={ref}>...</div>;
}

// ❌ 禁止
export const PatientCard: FC<Props> = () => {};       // FC禁止
export const PatientCard = forwardRef(() => {});       // forwardRef禁止
```

### フロントエンド - フォーム保護

- **NavigationBlocker**: `useBlocker` + `NavigationBlockerDialog` で未保存変更を検知
- **FormFieldError**: `role="alert"` + `aria-live="assertive"` で即時通知
- **`aria-describedby` 統一**: 全フォーム入力フィールドにエラー時のみ条件付与

### バックエンド - Context伝播（必須）

```go
// 全関数の第一引数に context.Context
func (s *Service) GetPet(ctx context.Context, id string) (*Pet, error)
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*Pet, error)

// GORMでもContext使用
r.db.WithContext(ctx).First(&pet, "id = ?", id)
```

### バックエンド - エラーハンドリング

```go
// センチネルエラー定義
var (
    ErrNotFound     = errors.New("resource not found")
    ErrInvalidInput = errors.New("invalid input")
)

// エラーラッピング
return fmt.Errorf("getPet: %w", err)

// エラー判定
if errors.Is(err, ErrNotFound) { /* 404レスポンス */ }
```

### バックエンド - slog構造化ログ

```go
slog.InfoContext(ctx, "pet created",
    slog.String("pet_id", pet.ID.String()),
    slog.String("name", pet.Name))
```

---

## 5. 共有コンポーネント一覧 (`/components/shared/`)

### レイアウト

| コンポーネント | 説明 |
|---|---|
| `PageLayout` | ページ全体のレイアウト（ヘッダー + スクロール可能コンテンツ） |
| `FormHeader` | ページヘッダー（タイトル、戻るボタン、アクション） |

### データ表示

| コンポーネント | 説明 |
|---|---|
| `DataTable` | 汎用テーブル（カラム定義 + renderRow） |
| `DataTableRow` | テーブル行（ホバー/クリック対応） |
| `SortableHeader` | ソート可能テーブルヘッダー（`useTableSort` 連携） |
| `Pagination` | ページネーション |
| `StatusBadge` | ステータスバッジ |
| `PatientInfoCard` | 患者(ペット)情報カード（`staffAriaDescribedBy` prop対応） |
| `LoadingSkeleton` | ローディング表示（5バリアント: form/list/detail/table/card） |

### フォーム・入力

| コンポーネント | 説明 |
|---|---|
| `PrimaryButton` | プライマリアクションボタン（`bg-[#2383E2]` Notionブルー） |
| `SearchFilterBar` | 検索バー + 件数表示 |
| `MasterSelectModal` | マスタ項目選択モーダル |
| `MasterSelectTrigger` | マスタ項目選択トリガーボタン（`ariaDescribedBy` prop対応） |
| `MasterLink` | マスタへのリンクナビゲーション |
| `NotionDatePicker` | Notion風日付ピッカー |
| `NotionPropertyRow` | Notion風プロパティ行（label/required/align） |
| `NotionSectionLabel` | Notion風セクションラベル（薄字uppercase） |
| `NotionSectionDivider` | Notion風薄罫線ディバイダー |
| `NotionCheckbox` | Notion風チェックボックス（`#2383E2`アクセントブルー） |
| `OwnerQuickViewModal` | 飼主情報クイックビューモーダル |
| `PetSearchForm` | ペット検索フォーム |
| `PetSearchResultsTable` | ペット検索結果テーブル |

### フィードバック・確認

| コンポーネント | 説明 |
|---|---|
| `ConfirmDialog` | 確認ダイアログ（destructive対応） |
| `NavigationBlocker` | フォーム離脱保護ダイアログ（`beforeunload` 統合） |
| `StaffImpactDialog` | スタッフ変更影響確認ダイアログ |
| `ErrorBoundary` | エラー境界（クラッシュ時のフォールバックUI） |
| `FormFieldError` | フォームフィールドエラー表示（`role="alert"` + `aria-live="assertive"`） |
| `PrintPreviewDialog` | 印刷プレビューダイアログ（`usePrint` フック連携） |

### アクション・履歴

| コンポーネント | 説明 |
|---|---|
| `RowActionDropdown` | テーブル行ドロップダウンメニュー（`MoreHorizontal` + `size-5`） |
| `HistoryPanel` | 履歴一覧パネル |
| `HistoryFilterPanel` | 履歴フィルタリングパネル |

### アクセシビリティ・通知

| コンポーネント | 説明 |
|---|---|
| `LiveAnnouncer` | ライブリージョン通知基盤（`LiveAnnouncerProvider` + `useAnnounce`） |
| `KeyboardShortcutHelp` | キーボードショートカットヘルプパネル（`?` キーで表示） |

### 共有フック (`/hooks/`)

| フック | 説明 |
|---|---|
| `usePagination` | ページネーションロジック |
| `usePetSearch` | ペット検索ロジック |
| `useStaffUsageCount` | スタッフ使用件数カウント |
| `useStaffValidation` | スタッフ関連バリデーション |
| `useTableSort` | ジェネリックテーブルソート（initialKey/initialDirection/comparator） |
| `useUnsavedChanges` | 未保存変更検知 |
| `useFocusTrap` | 汎用フォーカストラップ（Escape/Tab循環/フォーカス復帰） |
| `usePrint` | 汎用印刷状態管理（プレビューOpen/Close・ドキュメント種別切替） |
| `useNumericInput` | 数値入力バリデーション・書式整形 |
| `useReducedMotion` | `prefers-reduced-motion` メディアクエリ監視（WCAG 2.3.3） |

---

## 6. 機能一覧 (Features)

### 6.1 ダッシュボード (`/features/dashboard`)

- **概要**: 病院全体の稼働状況を俯瞰するホーム画面。カンバンボード形式。
- **ワークフロー（5カラム）**:
  1. 受付予約 → 2. 受付済 → 3. 診療中 → 4. 会計待ち → 5. 会計済
- **主要コンポーネント**:
  - `AppointmentCard`: 患者カード（react-dnd対応）
  - `KanbanColumn`: カラム
  - `DashboardDetailModal`: 詳細モーダル（各ステータス別アクションボタン付き）
  - `DashboardSummaryWidget`: 統計サマリーウィジェット（ネイティブSVGミニスパークライン）
  - `HospitalizationAlertWidget`: 入院アラートウィジェット
  - `useDashboardKanban`: カンバン状態管理
  - `useDashboardWeeklyStats`: 週次統計（localStorage永続化）

### 6.2 予約管理 (`/features/reservations`)

- **概要**: 診療・トリミング・手術の予約を一元管理。
- **予約タイプ**: マスタデータ駆動（`master_items.category=serviceType`）
- **機能**: カレンダー表示（月/週）、予約CRUD、患者検索・紐付け、担当医割り当て
- **主要コンポーネント**: `MonthView`, `WeekView`（Y軸D&D時間変更、現在時刻線）, `ReservationFormModal`, `ReservationDetailModal`
- **カラー凡例**: Figma準拠パステルカラー（診療=#dbeafe / 検診=#dcfce7 / 手術=#ffe2e2 等）

### 6.3 電子カルテ (`/features/medical-records`)

- **概要**: タブ形式で診療記録の各セクションを管理。
- **ステータス**: 作成中 → 確定済
- **タブ構成（8タブ）**:
  1. 問診（主訴・治療方針・問診履歴）
  2. 診察/治療プラン（バイタル・バイタル履歴グラフ・診断・治療プランTable）
  3. 治療（TreatmentTable、チェックボックスで完了移動）
  4. 予防接種（記録・フォーム）
  5. 検査（オーダーフォーム・結果履歴）
  6. 画像（アップロード・フィルタ付きギャラリー）
  7. 見積書（概算見積作成）
  8. 会計(医師確認)（算定チェック・会計連携）
- **主要フック**: `useMedicalRecordForm`（158行）+ `useMedicalRecordInit`（122行）+ `validateMedicalRecordSave`

### 6.4 入院管理 (`/features/hospitalization`)

- **概要**: 入院患者のケアプラン作成と日々の記録管理。
- **ステータス**: 入院中、退院済、予約
- **タイプ**: 入院、ホテル
- **機能**:
  - 入院ボード（DragPreviewオーバーレイ付き）
  - リストビュー
  - ケアプラン（5タイプ: 食事/投薬/処置/指示/アイテム）
  - デイリーログ（バイタル・排泄・食事記録）
  - タイムライン（実施記録時系列）
  - コスト管理（入院費用自動計算）
  - `HospitalizationSummaryDocument`: 入院サマリー帳票
- **主要フック**: `useHospitalizationForm`（186行）+ `useTreatmentPlans`（112行）
- **スタイリング**: NotionUI層（Detail/Form）と Kanban compact層（Board）を分離

### 6.5 会計 (`/features/accounting`)

- **概要**: 診療費の計算と請求書発行。
- **ステータス**: 未収(waiting)、収済(completed)、キャンセル(cancelled)、保留(pending)
- **機能**:
  - 入院連携（`item_source='hospitalization'` で入院治療プランを会計に引き渡し）
  - 保険フィルター（全て/保険あり/保険なし の `ToggleGroup`）
  - 保険負担内訳表示（領収書・診療明細書）
  - 割引計算（`calcLineTotal`, `calcGrandTotal`）
- **主要コンポーネント**: `AccountingDetail`（135行）、`AccountingItemTable`、`AccountingPaymentPanel`、`AccountingDocument`、`AccountingDocumentPreview`

### 6.6 検査管理 (`/features/examinations`)

- **概要**: 院内・院外検査のオーダーと結果管理。
- **ステータス**: 依頼中、検査中、完了
- **機能**: 検査フォームによるオーダー作成、結果入力・サマリー記録

### 6.7 顧客・ペット管理 (`/features/owners`, `/features/pets`)

- **概要**: 顧客とペットの基本情報管理。
- **機能**: 飼い主/ペット情報CRUD、ペット編集モーダル（`PetEditModal`）、ステータス管理（生存/死亡）

### 6.8 トリミング (`/features/trimming`)

- **概要**: トリミング業務の管理。
- **ステータス**: 予約、進行中、完了
- **機能**: トリミング予約リスト、コース選択・オプション管理、ペット選択フロー

### 6.9 予防接種管理 (`/features/vaccinations`)

- **機能**: 予防接種記録リスト、次回接種予定管理

### 6.10 定期健診管理 (`/features/checkups`)

- **機能**: 健診記録リスト（年次・シニア・パピー）、次回健診予定管理、カルテ遷移

### 6.11 在庫管理 (`/features/inventory`)

- **機能**: 在庫一覧（検索・カテゴリ/状態フィルタ・ソート・ページネーション）、在庫追加・編集、カルテ保存時の在庫消費連動（`consumeStock`）
- **カテゴリ**: 医薬品、消耗品、フード、その他
- **ステータス**: 在庫あり/残りわずか/在庫切れ（数量と発注点から自動判定）

### 6.12 シフト管理 (`/features/shifts`)

- **概要**: スタッフの勤務シフトを週間・月間カレンダーで管理。
- **シフトタイプ**: 通常勤務、午前のみ、午後のみ、休み、有給
- **機能**: 週間ビュー（スタッフ×曜日グリッド編集）、月間ビュー、Popover編集、ロールフィルタ、週計40h超過警告
- **アクセシビリティ**: `useFocusTrap`、`role="grid"` + `aria-label`、キーボード操作（Enter/Space）

### 6.13 マスタ管理 (`/features/master`)

- **概要**: 16カテゴリのマスタデータを一元管理。
- **カテゴリ**: serviceType, consultation, examination, procedure, vaccine, medicine, checkup, diagnosis_category, diagnosis_name, hospitalization, cage, trimming_course, trimming_option, staff, insurance, job_title（16種）
- **機能**:
  - フラットテーブル（大半のマスタ）
  - ツリーテーブル（階層構造マスタ）
  - D&Dソート（`useDragAndDrop` + `useKeyboardReorder`）
  - インライン追加（名前のみで完結するマスタ）
  - スタッフ変更影響確認ダイアログ
- **主要コンポーネント**: `Settings`（262行）+ `category-config.ts`（187行）+ 14セクションコンポーネント

### 6.14 認証 (`/features/auth`)

- **Phase 1〜3 完了**: `AuthProvider` + `LoginForm` + モック認証 + 6デモアカウント
- **Phase 2**: `ClinicSwitcher`サイドバー統合 + ユーザー情報表示
- **Phase 3**: サイドバー13メニュー項目の権限フィルタ + 41対象ルートに `ProtectedRoute` 権限ガード
- **3層モデル**: ユーザー種別（system_admin/clinic_admin/staff）+ 職種 + 権限（10種）
- **Phase 4（未実施）**: バックエンド接続時のRLSポリシー適用

---

## 7. ルーティング構成（42ルート）

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
/checkups                      → 定期健診一覧
/inventory                     → 在庫一覧
/inventory/new                 → 在庫新規登録
/inventory/:id                 → 在庫編集
/shifts                        → シフト管理 (週間・月間カレンダー)
/settings                      → マスタ設定トップ (カテゴリカード一覧)
/settings/clinic               → 病院情報設定
/settings/treatment-items      → 診療項目マスタ (診察・検査・処置・予防接種・定期健診のタブ統合)
/settings/diagnosis            → 診断マスタ (診断カテゴリ・診断名のタブ統合)
/settings/trimming             → トリミングマスタ (コース・オプションのタブ統合)
/settings/{category-slug}      → 各マスタカテゴリ設定 (6パターン)
/dev/tests                     → フォーマットテスト (開発用)
/login                         → ログイン (サイドバーなし)
```

---

## 8. 実装済みの主要技術パターン

### React.lazy() コード分割

30ルートコンポーネントを `lazyNamed` ユーティリティ経由で遅延読み込み化済み。
`/lib/lazy-route.tsx` に `lazyNamed`・`RouteFallback`・`SuspenseRoute` を配置。

### React 19 新機能

```typescript
// useActionState: フォームアクション管理
const [state, formAction, isPending] = useActionState(submitAction, initialState);

// useOptimistic: 楽観的UI更新
const [optimisticItems, addOptimisticItem] = useOptimistic(items, updateFn);

// use(): Promise/Context直接読み取り
const data = use(fetchPromise);
```

### lib ユーティリティ

| ファイル | 説明 |
|---------|------|
| `design-tokens.ts` | Tailwind CSS定数・色パレット |
| `status-helpers.ts` | ステータス表示ヘルパー |
| `type-utils.ts` | 型安全ユーティリティ（isOneOf, typedSetter, typedKeys, parseLocationState） |
| `format.ts` | 通貨フォーマット・内税計算・割引計算（extractInnerTax, calcLineTotal, calcGrandTotal） |
| `suspense-utils.ts` | Suspense関連ユーティリティ |
| `lazy-route.tsx` | React.lazy() コード分割ユーティリティ |

### データベース

- 31テーブル、30+ENUM型
- UUID主キー（`gen_random_uuid()`）
- `TIMESTAMPTZ` でタイムゾーン対応
- インデックス設計済み（詳細は `docs/DB_DEFINITION.md` 参照）

---

## 9. 関連ドキュメント

| ドキュメント | パス | 説明 |
|---|---|---|
| **ER図** | `docs/ERD.md` | 全31テーブルの定義・リレーション・列挙型一覧（Mermaid） |
| **DB定義書** | `docs/DB_DEFINITION.md` | PostgreSQL DDL・インデックス・型マッピング |
| **画面仕様書** | `docs/SCREENS.md` | 全42ルートの画面仕様・構成・データフロー |
| **マスタ設定仕様書** | `docs/screens/20-master-settings.md` | マスタ管理の詳細仕様 |
| **デザインシステム** | `docs/DESIGN_SYSTEM.md` | カラーパレット、タイポグラフィ、スタイリング規約 |
| **API設計** | `docs/API-ROADMAP.md` | APIロードマップ |
| **コーディング規約** | `CODING_RULES.md` | 全体コーディング規約 |
| **Frontend規約** | `frontend/CODING_RULES.md` | React 19 / TypeScript詳細 |
| **Backend規約** | `backend/CODING_RULES.md` | Go / Gin / GORM詳細 |
| **Swagger UI** | `http://localhost:8080/swagger/index.html` | API仕様（バックエンド起動時） |
