# 動物病院管理システム 仕様定義書

## 1. プロジェクト概要
本プロジェクトは、ReactとTailwind CSSを用いたモダンなWebベースの動物病院向け統合管理システム（PMS: Practice Management System）のプロトタイプです。
予約管理、受付、電子カルテ、入院管理、会計までの一連の業務フローをシームレスに連携させ、獣医療現場の業務効率化と質の高い医療提供を支援することを目的としています。

## 2. 技術スタック

### フロントエンド
- **Core**: React 19, TypeScript 5.7
- **Build Tool**: Vite 6
- **Routing**: React Router (`react-router` パッケージ, `createBrowserRouter` + `RouterProvider` 構成)
  - **注意**: `react-router-dom` ではなく `react-router` からimportすること
- **State Management**: React Hooks (Custom Hooks) + TanStack Query (React Query)

### UI/UX & デザインシステム
- **CSS Framework**: Tailwind CSS v4
- **Component Library**: shadcn/ui (Radix UI Primitives ベース)
    - プロジェクト内にソースコードとして配置 (`/components/ui`)
    - 全セレクトボックスを `Combobox` に統一
- **Design System**: Notion Like Design System (詳細は `docs/DESIGN_SYSTEM.md` を参照)
    - デザインは `lib/design-tokens.ts` に集約し、ハードコードのHex/rgbaをゼロに保つ
    - メインカラー: ティール (`#038B94`)
    - プライマリボタン: Notionブルー (`#2383E2`)
    - 操作アイコン: `MoreHorizontal` + `size-5`
    - 内部ID（冗長なNo管理フィールド）はUIに非表示とし、ユーザーから隠蔽する
- **Icons**: Lucide React
- **Charts**: ネイティブSVG描画（`DashboardSummaryWidget` のミニスパークライン等）。`/components/ui/chart.tsx` にshadcn/ui内部のRecharts統合が残存するが、現在どのfeatureからも参照されておらず、全feature配下のrecharts直接使用はゼロ
- **Drag & Drop**: react-dnd + react-dnd-touch-backend (カンバンボード)
- **Animations**: Motion (`motion/react`) / CSS Transitions
- **Toast**: Sonner (`sonner@2.0.3`)

### フォーム & データハンドリング
- **Form Management**: React Hook Form (`react-hook-form@7.55.0`)
- **Date Utility**: date-fns

### バックエンド
- **Language**: Go 1.25
- **Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL 18
- **Infrastructure**: Docker Compose

## 3. アーキテクチャ
「bulletproof-react」の設計思想に基づいた **Feature-based Architecture** を採用しています。機能（Feature）ごとにディレクトリを分割し、関心の分離とスケーラビリティを確保しています。

### ディレクトリ構造
```
/
├── components/               # 共有コンポーネント
│   ├── ui/                   # shadcn/ui (汎用UIパーツ)
│   ├── shared/               # アプリケーション固有の共有パーツ
│   └── figma/                # Figmaインポート用アセットコンポーネント
├── features/                 # 機能別モジュール (ドメインロジックの中核)
│   ├── [feature_name]/
│   │   ├── api/
│   │   │   ├── index.ts      # barrel re-export
│   │   │   ── mockData.ts   # モックデータ
│   │   │   └── *.ts          # 個別API関数
│   │   ├── components/       # 機能固有のUIコンポーネント
│   │   ├── constants/        # 定数定義 (一部featureのみ)
│   │   ├── hooks/            # ビジネスロジック・状態管理フック
│   │   ├── routes/
│   │   │   ├── index.ts      # barrel re-export
│   │   │   └── *.tsx         # ページコンポーネント
│   │   └── types/
│   │       └── index.ts      # barrel re-export (型定義の単一エントリポイント)
│   └── ...
├── hooks/                    # 共有カスタムフック
├── lib/                      # ユーティリティ、定数、設定
│   ├── design-tokens.ts      # デザイントークン定数
│   ├── status-helpers.ts     # ステータス表示ヘルパー
│   ├── type-utils.ts         # 型安全ユーティリティ (isOneOf, typedSetter, typedKeys, parseLocationState)
│   ├── format.ts             # 通貨フォーマット・内税計算・割引計算ユーティリティ (extractInnerTax, calcLineTotal, calcGrandTotal)
│   ├── suspense-utils.ts     # Suspense関連ユーティリティ
│   └── lazy-route.tsx        # React.lazy() コード分割ユーティリティ (lazyNamed, RouteFallback, SuspenseRoute)
├── styles/
│   └── globals.css           # Tailwind CSS v4 グローバルスタイル / CSS変数
├── types/
│   └── index.ts              # グローバル型定義 (全featureから参照)
└── App.tsx                   # アプリケーションエントリポイント
```

### Barrel Index パターン
全17 featureに `api/index.ts` と `types/index.ts` のbarrel indexを整備済みです。
- **型import**: 常に `../../{feature}/types` barrel経由
- **API import**: 常に `../../{feature}/api` barrel経由
- **サブモジュール直接参照禁止**: `../types/diagnosis` のような直接参照は不可

### データ永続化の現状
現在、バックエンドAPIは実装中。**Mock Data**（各 feature ディレクトリ内の `api/mockData.ts`）を使用して動作しています。
一部機能（master, clinic）は `api/store.ts` でインメモリストアを保持しますが、ブラウザのリロードでリセットされます。
**会計・入院機能**にはLocalStorageベースの永続化ストアを採用しており、ブラウザのリロードをまたいでデータが保持されます。

### ストアバージョン管理（`features/hospitalization/api/store.ts`）

`STORE_VERSION` 定数でスキーマ世代を管理します。この文字列を変更するとブラウザの旧 LocalStorage データを自動パージし、`MOCK_HOSPITALIZATIONS` で再初期化します。

| 定数 | 現在値 | 変更タイミング |
|---|---|---|
| `STORE_VERSION` | `"v2026-03"` | モックデータのスキーマ変更・日付刷新・フィールド追加時 |
| `STORE_VERSION_KEY` | `"animal_hospital_hosp_store_version"` | 変更不要（LocalStorageキー名） |

**変更手順**: `store.ts` の `STORE_VERSION` 文字列を `"v2026-XX"` 形式でインクリメント → ブラウザ再アクセス時に自動マイグレーション実行。

**HOSP\_STORE\_EVENT の発火タイミング**:
- `setStoredHospitalizations` — 退院処理・入院情報更新・削除

`useHospitalizations` はこのイベントを `addEventListener` でサブスクライブし、Detail画面からの退院処理がList/Board画面に即時反映されます。

## 4. コード品質規約

### 型安全性
- **`any` 型ゼロ**: 全ファイルで `any` 型の使用を禁止
- **インライン `as const` ゼロ**: 定数配列は `types/index.ts` 等で事前定義
- **`as` 型断言ゼロ (feature/shared/hooks層)**: 型キャストではなく型ガード (`instanceof`)・`typedSetter`・`typedKeys`・`parseLocationState` を使用。`as` 断言は `/lib/type-utils.ts`・`/lib/suspense-utils.ts`・`/lib/lazy-route.tsx` のユーティリティ内部にのみ集約
- **型定義の配置**: コンポーネント/hookファイルからの型exportは完全禁止。`types/index.ts` barrel経由に統一

### 制御フロー
- **switch文ゼロ**: Record マッピングまたは lookup テーブルで代替
- **`else if` ゼロ**: 早期 return、三項演算子、Record lookup で代替
- **値列挙パターン**: `const XXX_VALUES = [...] as const` → `type Xxx = (typeof XXX_VALUES)[number]` → `Record<Xxx, string>` ラベルマップの三点セット

### 型安全ユーティリティ (`/lib/type-utils.ts`)
```typescript
// shadcn の onValueChange (string) 型全に絞り込む
isOneOf<T>(value: string, values: readonly T[]): value is T
typedSetter<T>(setter, validValues): (value: string) => void
typedSetterNonEmpty<T>(setter, validValues): (value: string) => void
// Object.keys() の型安全ラッパー（as K[] 不要）
typedKeys<K>(obj: Record<K, unknown>): K[]
// React Router location.state の安全なパース
parseLocationState<T>(state: unknown): Partial<T>
```

### フォーム保護
- **NavigationBlocker**: `useBlocker` を呼ぶ内部コンポーネント `NavigationBlockerDialog` を分離し、`when` が `true` のときだけマウントする方式
- **未保存変更検知**: `useUnsavedChanges` フックで `isDirty` を追跡

### フォームバリデーション・アクセシビリティ
- **`FormFieldError`**: `role="alert"` + `aria-live="assertive"` で即時通知、一意の `id` 属性を付与
- **`aria-describedby` 統一パターン**: 全フォーム入力フィールドにエラー時のみ `aria-describedby={errorId}` を条件付与（スクリーンリーダーがフォーカス時にエラーメッセージを読み上げ）
- **`PatientInfoCard`**: `staffAriaDescribedBy` prop でスタッフ選択ボタンと `FormFieldError` を接続（カルテ・入院・トリミング共通）
- **対応済みフォーム**: 飼主・ペット編集・マスタ項目・予約・会計入金・在庫・カルテ・入院・トリミング

### トースト構造化
- `sonner@2.0.3` の `toast.success()` / `toast.warning()` を使用
- `description` プロパティで補足情報を付与

### import規約
- 全て相対パス（`../../feature/api` 等）
- cross-feature importは barrel 経由のみ

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
| `SortableHeader` | ソート可能テーブルヘッダー (`useTableSort` 連携) |
| `Pagination` | ページネーション |
| `StatusBadge` | ステータスバッジ |
| `PatientInfoCard` | 患者(ペット)情報カード |
| `LoadingSkeleton` | ローディング表示 (5バリアント: form / list / detail / table / card) |

### フォーム・入力
| コンポーネント | 説明 |
|---|---|
| `PrimaryButton` | プライマリアクションボタン (`bg-[#2383E2]` Notionアクセントブルー、shadow-none) |
| `SearchFilterBar` | 検索バー + 件数表示 |
| `MasterSelectModal` | マスタ項目選択モーダル |
| `MasterSelectTrigger` | マスタ項目選択トリガーボタン (`MasterSelectModal` と連携、`ariaDescribedBy` prop 対応) |
| `MasterLink` | マスタへのリンクナビゲーション |
| `NotionDatePicker` | Notion風日付ピッカー |
| `NotionPropertyRow` | Notion風プロパティ行（label/required/align）。全Notion風フォームの共通キー・バリュー |
| `NotionSectionLabel` | Notion風セクションラベル（薄字uppercase） |
| `NotionSectionDivider` | Notion風薄罫線ディバイダー（`className`でマージンカスタマイズ可） |
| `NotionCheckbox` | Notion風チェックボックス（`#2383E2`アクセントブルー、ホバー時ブルーティント、`role="checkbox"`+`aria-checked`、`stopPropagation`によるラッパー二重トグル防止対応）。使用箇所: `TreatmentTable`・`Dashboard`・`TrimmingForm` |
| `OwnerQuickViewModal` | 飼主情報クイックビューモーダル（カルテ画面から飼主詳細を参照） |
| `PetSearchForm` | ペット検索フォーム |
| `PetSearchResultsTable` | ペット検索結果テーブル |

### フィードバック・確認
| コンポーネント | 説明 |
|---|---|
| `ConfirmDialog` | 確認ダイアログ (destructive対応) |
| `NavigationBlocker` | フォーム離脱保護ダイアログ (`beforeunload` 統合) |
| `StaffImpactDialog` | スタッフ変更影響確認ダイアログ |
| `ErrorBoundary` | エラー境界 (クラッシュ時のフォールバックUI) |
| `FormFieldError` | フォームフィールドエラー表示 (`role="alert"` + `aria-live="assertive"`) |
| `PrintPreviewDialog` | 印刷プレビューダイアログ (汎用・`usePrint` フック連携) |

### アクション
| コンポーネント | 説明 |
|---|---|
| `RowActionDropdown` | テーブル行ドロップダウンメニュー（`MoreHorizontal` + `size-5` アイコン使用） |

### 履歴・フィルタ
| コンポーネント | 説明 |
|---|---|
| `HistoryPanel` | 履歴一覧パネル |
| `HistoryFilterPanel` | 履歴フィルタリングパネル |

### アクセシビリティ・通知
| コンポーネント | 説明 |
|---|---|
| `LiveAnnouncer` | ライブリージョン通知基盤 (`LiveAnnouncerProvider` + `useAnnounce` フック) |
| `KeyboardShortcutHelp` | キーボードショートカットヘルプパネル (`?` キーで表示) |

### 共有フック (`/hooks/`)
| フック | 説明 |
|---|---|
| `usePagination` | ページネーションロジック |
| `usePetSearch` | ペット検索ロジック |
| `useStaffUsageCount` | スタッフ使用件数カウント |
| `useStaffValidation` | スタッフ関連バリデーション |
| `useTableSort` | ジェネリックテーブルソート（`initialKey`/`initialDirection`/`comparator`/`resetSort`） |
| `useUnsavedChanges` | 未保存変更検知 |
| `useFocusTrap` | 汎用フォーカストラップ (Escape/Tab循環/フォーカス復帰) |
| `usePrint` | 汎用印刷状態管理 (プレビューOpen/Close・ドキュメント種別切替・`data-print-type`属性・`window.print()`トリガー) |
| `useNumericInput` | 数値入力バリデーション・書式整形フック |
| `useReducedMotion` | `prefers-reduced-motion: reduce` メディアクエリ監視（WCAG 2.3.3）。適用箇所: `TreatmentTable`（行アニメーション）、`MasterSidePeek`（サイドパネル開閉）、`WeekView.AppointmentCard`（ドラッグ視覚効果） |

> **注**: `useAnnounce` は `/components/shared/LiveAnnouncer.tsx` から export されるフックです。`/hooks/` ディレクトリには配置せず、`LiveAnnouncerProvider` と共にインポートします。

## 6. 機能一覧 (Features)

### 6.1 ダッシュボード (`/features/dashboard`)
- **概要**: 病院全体の稼働状況を俯瞰するホーム画面。カンバンボード形式で患者の来院フローを管理。
- **ワークフロー（カラム構成）**:
    1. **受付予約**: 本日来院予定の予約
    2. **受付済**: 来院済み・待合室待機中
    3. **診療中**: 診察室・処置室で対応中
    4. **会計待ち**: 診察終了・会計計算待ち
    5. **会計済**: 会計完了・帰宅
- **構成**:
    - `AppointmentCard`: 患者カード (DnD対応)
    - `KanbanColumn`: カラム
    - `DashboardDetailModal`: 詳細モーダル
    - `DashboardSummaryWidget`: 統計サマリーウィジェット（カラム別件数カード＋ネイティブSVGミニスパークライン。ホバーツールチップ・アクティブドットもSVG＋CSSで実装、recharts依存ゼロ）
    - `HospitalizationAlertWidget`: 入院アラートウィジェット（入院中件数・退院超過アラート・クイックリンク）
    - `useDashboardKanban`: カンバン状態管理フック
    - `useDashboardWeeklyStats`: 週次統計データ収集フック（localStorage永続化。曜日ごとのカラム別件数スナップショットを記録し、`DashboardSummaryWidget` のスパークラインデータを提供。前日比較トレンド表示にも使用）

### 6.2 予約管理 (`/features/reservations`)
- **概要**: 診療・トリミング・手術の予約を一元管理。
- **予約タイプ**: マスタデータ駆動（`master_items` category=`serviceType`）。初期値: 診療、定期健診、検査、手術、トリミング、予防接種、入院、ホテル
- **機能**:
    - カレンダー表示（月表示 / 週表示）
    - 予約の新規作成・編集・キャンセル
    - 患者（ペット）検索と予約紐付け
    - 担当獣医師の割り当て
- **構成 (フック分離済み)**:
    - `ReservationManagement` (168行) — UI層のみ
    - `useReservationManagement` (298行) — CRUD・モーダル・重複チェック・ナビゲーション
    - `MonthView` / `WeekView`: カレンダービュー
    - `ReservationFormModal`: 予約フォームモーダル
    - `ReservationDetailModal`: 予約詳細モーダル

### 6.3 電子カルテ (`/features/medical-records`)
- **概要**: タブ形式で診療記録の各セクションを管理。
- **ステータス**: 作成中 → 確定済
- **タブ構成**:
    - **問診**: 主訴 (Markdown)、治療方針 (Markdown)、問診履歴
    - **診察/治療プラン**: バイタル入力 (体重, 体温, 心拍数等)、バイタル履歴グラフ、診断登録（カテゴリ+診断名）、治療プラン (TreatmentTable)
    - **治療**: 治療プランテーブル（チェックボックスで治療済みへ移動、確認ダイアログ付き）＋治療済みテーブル（戻すボタンでプランへ差し戻し）、マスタ検索連携
    - **予防接種**: 予防接種記録・フォーム
    - **検査**: 検査オーダーフォーム・結果履歴
    - **画像**: 検査画像のアップロード・フィルタ付きギャラリー
    - **見積書**: 診療内容に基づく概算見積作成
    - **会計(医師確認)**: 算定チェック・会計連携
- **構成 (大型ファイル分割済み)**:
    - `useMedicalRecordForm`: メインフック (158行)
    - `useMedicalRecordInit`: 初期化ロジック (122行) — recordId/petIdベースのデータロード
    - `validateMedicalRecordSave`: バリデーション純粋関数 (16行)
- **型サブモジュール**: `types/diagnosis.ts`, `types/examination.ts`, `types/interview.ts`, `types/vaccination.ts`, `types/vital.ts` → `types/index.ts` でbarrel re-export

### 6.4 入院管理 (`/features/hospitalization`)
- **概要**: 入院患者のケアプラン作成と日々の記録管理。
- **ステータス**: 入院中、退院済、予約
- **タイプ**: 入院、ホテル
- **機能**:
    - **入院ボード**: カード形式の入院患者一覧
    - **リストビュー**: テーブル形式の一覧表示
    - **ケアプラン**: 投薬・処置スケジュールの作成（食事、投薬、処置、指示、アイテム）
    - **デイリーログ**: タスク完了チェック、バイタル・排泄・食事の記録
    - **タイムライン**: 実施記録の時系列表示
    - **コスト管理**: 入院費用の自動計算
- **構成 (大型ファイル分割済み)**:
    - `useHospitalizationForm` (186行) + `useTreatmentPlans` (112行)
    - `components/CarePlan/`: CarePlanDialog, CarePlanItemRow, CarePlanSection
    - `components/DailyRecord/`: DailyRecordSection, Timeline, VitalDialog, LogDialog 等
    - `CageDragPreview`: `useDragLayer` によるカスタムドラッグプレビューオーバーレイ（ケージ間ペット移動時に半透明カードがカーソルに追従）。`HospitalizationBoard` に統合済み
    - `CarePlanPreviewPopover`: ケアプラン概要ポップオーバー（ステータストグル付き。active↔completed の即時切り替えが可能）
    - `HospitalizationSummaryDocument`: 入院サマリー帳票（入院日数自動計算・1日あたり費用表示）
    - デスクトップ/モバイルレスポンシブレイアウト分離
- **スタイリングポリシー（NotionライクUI vs. Kanban compact）**:

  | スコープ | 使用トークン | 主な特徴 |
  |---|---|---|
  | **Notionページ層** (Detail/Form/Mobile) | `C.*` + `STYLE.sectionLabel` + `p-4` | shadow なし・border のみ・セクションラベル uppercase |
  | **Kanban compact層** (Board/CageDragPreview) | `H_STYLES.*` + `C.*` | `p-2`・`text-lg font-bold`（ペット名）・`hover:ring-1 hover:ring-[#37352F]/20` |

  - `H_STYLES` (`/features/hospitalization/styles.ts`) は **Board専 compact トークン** として維持。Detail/Form/Mobile 系コンポーネントからの import は禁止。
  - Dialog 系コンポーネント (VitalDialog / LogDialog / CarePlanDialog / TaskCompleteDialog) はローカル定数 `DIALOG_BTN = "h-10 px-4 text-sm"` を使用し、`H_STYLES.button.action` に依存しない。
  - `CarePlanPreviewPopover`: `HOSP_STORE_EVENT` をサブスクライブし、Detail 画面でのプラン変更後に Board 上のポップオーバーが自動リフレッシュされる。

### 6.5 会計 (`/features/accounting`)
- **概要**: 診療費の計算と請求書発行。
- **ステータス**: 未収(waiting)、収済(completed)、キャンセル(cancelled)、保留(pending)
- **機能**:
    - 患者選択から会計情報への連携
    - 診療明細の作成・編集
    - 請求書・領収書プレビュー
    - 入金手段管理（現金、クレジットカード、電子マネー）
    - **入院連携**: `ITEM_SOURCE_VALUES` に `"hospitalization"` を追加。入院→退院→会計への明細引き渡しパイプライン完成。会計一覧・明細テーブル・領収書・診療明細書すべてで「入院連携」バッジ表示
    - **保険フィルター**: 会計一覧に保険列と `INSURANCE_FILTER_VALUES` 型の `ToggleGroup` フィルター（全て / 保険あり / 保険なし）を追加
    - **保険負担内訳**: 領収書・診療明細書に保険負担内訳3行セット（保険適用額・自己負担額・保険者負担額）および保険負担割合別シミュレーション注釈を追加
    - **割引計算**: `/lib/format.ts` の `calcLineTotal`（行レベル割引）・`calcGrandTotal`（グローバル割引）に集約。カルテ4タブと入院側の両方で使用
- **構成 (大型ファイル分割済み)**:
    - `AccountingDetail` (135行) — メインページ
    - `useAccountingDetail` — 状態管理フック
    - `AccountingItemTable` — 明細テーブル
    - `AccountingPaymentPanel` — 入金パネル
    - `AccountingDocument` — 請求書・領収書レンダリング
    - `AccountingDocumentPreview` — 書類プレビューダイアログ

### 6.6 検査管理 (`/features/examinations`)
- **概要**: 院内・院外検査のオーダーと結果管理。
- **ステータス**: 依頼中、検査中、完了
- **機能**:
    - 検査フォームによるオーダー作成
    - 結果入力とサマリー記録

### 6.7 顧客・ペット管理 (`/features/owners`, `/features/pets`)
- **概要**: 顧客とペットの基本情報管理。
- **機能**:
    - 飼い主情報のCRUD
    - ペット情報のCRUD（画像設定含）
    - ペット編集モーダル (`PetEditModal`)
    - ステータス管理（生存/死亡）

### 6.8 トリミング (`/features/trimming`)
- **概要**: トリミング業務の管理。
- **ステータス**: 予約、進行中、完了
- **機能**:
    - トリミング予約リスト
    - コース選択とオプション管理
    - ペット選択フロー

### 6.9 予防接種管理 (`/features/vaccinations`)
- **概要**: 予防接種の管理。
- **機能**:
    - 予防接種記録リスト
    - 次回接種予定の管理

### 6.10 定期健診管理 (`/features/checkups`)
- **概要**: 定期健康診断の記録管理。
- **機能**:
    - 健診記録リスト（年次健康診断・シニア健康診断・パピー健診）
    - 結果概要の表示
    - 次回健診予定の管理
    - カルテへの遷移

### 6.11 在庫管理 (`/features/inventory`)
- **概要**: 在庫品目の一覧管理・登録・編集。ルート有効化済み、サイドバーにナビゲーション追加済み。
- **機能**:
    - 在庫一覧表示（検索・カテゴリフィルタ・状態フィルタ・ソート・ページネーション）
    - 在庫追加・編集フォーム（バリデーション・`FormFieldError`・`NavigationBlocker`・`useUnsavedChanges`・`LoadingSkeleton`）
    - カルテ保存時の在庫消費連動 (`consumeStock`)
    - ソート: `SortableHeader` + `useTableSort`（数値 `comparator` 適用、7列ソート対応）
    - 負数入力防止（`min={0}` + `Math.max(0, ...)` ガード）
- **カテゴリ**: 医薬品、消耗品、フード、その他
- **ステータス**: 在庫あり / 残りわずか / 在庫切れ（在庫数と発注点から自動判定）

### 6.12 シフト管理 (`/features/shifts`)
- **概要**: スタッフの勤務シフトを週間・月間カレンダーで管理。
- **シフトタイプ**: 通常勤務、午前のみ、午後のみ、休み、有給
- **機能**:
    - **週間ビュー**: スタッフ×曜日グリッドでシフト編集（`ShiftWeekView`）
    - **月間ビュー**: カレンダー形式の月間俯瞰（`ShiftMonthView`）
    - シフトセルクリックによるPopover編集（`ShiftEditPopover`）
    - ロールフィルタ（全員・医師・スタッフ・トリマー）
    - 週計労働時間の自動計算・40時間超過警告
    - Notion風カラーパレットによるシフトタイプ別色分け
    - シフト凡例表示（`ShiftLegend`）
- **アクセシビリティ**:
    - `ShiftEditPopover`: `useFocusTrap` によるフォーカストラップ（Escape/Tab循環/フォーカス復帰）、`role="dialog"` + `aria-modal="true"`、クリック外閉じ
    - ツールバーボタングループ: `role="group"` + `aria-label`、トグルボタンに `aria-pressed`
    - ナビゲーションボタン: `size-10`（40px）+ `after:absolute after:-inset-2` ヒットエリア拡張
    - トグルボタン: `py-2` 高さ確保、Popover内ボタンに `after:absolute after:-inset-*` ヒットエリア拡張
    - シフトセル: `role="button"` + `tabIndex={0}` + キーボードハンドラ（Enter/Space）、`min-h-[56px]` タッチターゲット
    - 週間テーブル: `role="grid"` + `aria-label="週間シフト表"`
    - 凡例: `role="list"` / `role="listitem"` セマンティクス
- **構成**:
    - `ShiftCalendar`: メインページコンポーネント（`PageLayout` ラップ）
    - `ShiftWeekView` / `ShiftMonthView`: ビューコンポーネント
    - `ShiftCell`: セル表示コンポーネント（`<div>` ベース、親 `<td>` 内に配置）
    - `ShiftEditPopover`: シフト編集Popover（`useFocusTrap` 統合）
    - `ShiftLegend`: 凡例コンポーネント
    - `useShiftManagement`: 状態管理フック（ビュー切替・ナビゲーション・CRUD・労働時間計算）
- **型定義**: `SHIFT_TYPE_VALUES` / `SHIFT_VIEW_VALUES` の値列挙パターン、`ShiftEntry`・`ShiftStaffInfo`・`DayShiftSummary`・`ShiftColorConfig` インターフェース

### 6.13 マスタ管理 (`/features/master`)
- **概要**: 診療・トリミング・入院・スタッフなど、システム全体で使用するマスタデータを一元管理。
- **マスタカテゴリ** (16種類):
    - **診療関連**: serviceType (予約区分), consultation (診察), examination (検査), procedure (処置), vaccine (予防接種), medicine (薬剤), checkup (定期健診)
    - **診断**: diagnosis_category (診断カテゴリ), diagnosis_name (診断名)
    - **入院**: hospitalization (入院), cage (ケージ)
    - **トリミング**: trimming_course (コース), trimming_option (オプション)
    - **その他**: staff (スタッフ), insurance (保険), job_title (職種)
- **データ構造**: 汎用テーブル`MasterItem`（15カテゴリ統合、LocalStorage実装）
- **UI形式**:
    - **フラットテーブル**: 大半のマスタ（staff, vaccine, medicine, cage, insurance等）
    - **ツリーテーブル**: 階層構造を持つマスタ（diagnosis_category, diagnosis_name, serviceType）
    - **インライン追加**: 名前のみで完結するマスタ（job_title, diagnosis_category, serviceType等）のみ下部にインライン追加フォームを表示。複数必須項目がある場合は非表示
- **スタッフマスタ詳細**:
    - **職種**: Combobox形式、job_titleマスタから動的取得（active項目のみ、単一選択）
    - **所属医院**: DropdownMenuのCheckboxItem形式で複数選択可能（MOCK_CLINICS連動）
    - **アカウント情報**: メールアドレス、パスワード、ユーザー種別（システム管理者/医院管理者/スタッフ）
    - **その他**: 資格番号（任意）
    - **一覧表示列**: 名称、職種、所属医院、メールアドレス、最終ログイン、ステータス、操作
    - **特記事項**: 社員番号フィールドは存在せず、コード列も非表示
- **操作機能**:
    - **ドラッグ&ドロップ並び替え**: 全マスタで対応（`useDragAndDrop`）
    - **キーボード並び替え**: Alt+矢印キーで項目移動（`useKeyboardReorder`）
    - **検索・フィルタ**: `SearchFilterBar` による全文検索
    - **編集**: 行クリックでインライン編集モードへ移行
    - **削除**: スタッフは影響範囲確認ダイアログ（`StaffImpactDialog`）、他は通常確認ダイアログ
- **権限管理**:
    - ユーザー種別により機能制限を実施
    - システム管理者: 全マスタ編集可能
    - 医院管理者: 所属医院のマスタ編集可能
    - スタッフ: 閲覧のみ
- **構成**:
    - `Settings`: マスタ編集メインページ（リスト/編集切替）
    - `MasterSettingsIndex`: マスタカテゴリ一覧（カード表示）
    - `MasterFlatDataTable`: フラットテーブル表示
    - `MasterTreeDataTable`: ツリーテーブル表示
    - `MasterItemEditForm`: マスタ項目編集フォーム
    - `MasterItemFormSections`: カテゴリ固有フォームセクションディスパッチャー
    - **セクションコンポーネント**: `StaffSection`, `VaccineSection`, `MedicineSection` 等（各マスタカテゴリ固有のフォーム）
    - `useMasterItemEditor`: CRUD操作フック
    - `useMasterItems`: マスタデータ取得フック

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
/login                         → ログイン (サイドバーなし、AUTH.md Phase 1〜3 実装済み)
```

## 8. 大型ファイル分割の進捗

| 元ファイル | 元行数 | 分割後 | 行数 |
|---|---|---|---|
| `AccountingDetail.tsx` | 587行 | `AccountingDetail.tsx` | 135行 |
| | | `useAccountingDetail.ts` |  |
| | | `AccountingItemTable.tsx` | — |
| | | `AccountingPaymentPanel.tsx` | — |
| | | `AccountingDocument.tsx` | — |
| | | `AccountingDocumentPreview.tsx` | — |
| `useHospitalizationForm.ts` | 281行 | `useHospitalizationForm.ts` | 186行 |
| | | `useTreatmentPlans.ts` | 112行 |
| `master/Settings.tsx` | 1147行 | `Settings.tsx` | 262行 |
| | | `category-config.ts` | 187行 |
| | | `MasterItemFormSections.tsx` | 55行 (ディスパッチャー) |
| | | `sections/` (14コンポーネント) | 各40-90行 |
| | | `useMasterItemEditor.ts` | 195行 |
| `MasterItemFormSections.tsx` | 582行 | `MasterItemFormSections.tsx` | 55行 |
| | | `sections/SectionWrapper.tsx` | 33行 |
| | | `sections/ExaminationSection.tsx` | 90行 |
| | | `sections/VaccineSection.tsx` | 48行 |
| | | `sections/MedicineSection.tsx` | 58行 |
| | | `sections/StaffSection.tsx` | 48行 |
| | | `sections/CageSection.tsx` | 56行 |
| | | `sections/InsuranceSection.tsx` | 48行 |
| | | `sections/TrimmingCourseSection.tsx` | 52行 |
| | | `sections/TrimmingOptionSection.tsx` | 52行 |
| | | `sections/HospitalizationSection.tsx` | 56行 |
| | | `sections/DiagnosisNameSection.tsx` | 60行 |
| `useMedicalRecordForm.ts` | 221行 | `useMedicalRecordForm.ts` | 158行 |
| | | `useMedicalRecordInit.ts` | 122行 |
| | | `validateMedicalRecordSave.ts` | 16行 |
| `MasterSettingsIndex.tsx` | 277行 | `MasterSettingsIndex.tsx` | 187行 |
| (DRY化) | | (`CATEGORY_CONFIG` 統合) | — |
| `ReservationManagement.tsx` | 424行 | `ReservationManagement.tsx` | 168行 |
| | | `useReservationManagement.ts` | 298行 |
| `SettingsContent.tsx` | 693行 | `SettingsContent.tsx` | 260行 |
| | | `MasterStatusDot.tsx` | 23行 |
| | | `MasterInlineAdd.tsx` | 130行 |
| | | `MasterSidePeek.tsx` | 45行 |
| | | `MasterTreeDataTable.tsx` | 230行 |
| | | `MasterFlatDataTable.tsx` | 130行 |
| | | `master-table-types.ts` | 2行 |
| `useDragAndDrop.ts` | 525行 | `useDragAndDrop.ts` | 290行 |
| | | `useKeyboardReorder.ts` | 180行 |
| | | `drag-preview.ts` | 45行 |

## 9. 今後のロードマップ
1. **バックエンド連携**: Go 1.25/Gin/GORM/PostgreSQL 18による全45テーブルの本番API実装（現在進行中）
2. ~~**在庫管理の再有効化**~~ **完了**（ルート有効化・サイドバー追加・NavigationBlocker統合・LoadingSkeleton適用・ソート機能・フォームバリデーション・SCREENS.md仕様展開済み）
3. **さらなるリファクタリング候補**:
    - ~~`MasterItemFormSections.tsx` (582行) の各カテゴリセクションを個別コンポーネントに分割~~ **完了**
    - ~~shared components の Props 型を `types/` barrel に集約整理~~ **完了**
    - ~~`as` 型断言の完全排除~~ **完了**（`typedKeys`・`parseLocationState`・`instanceof` ガードで feature/shared/hooks 層の全 `as` 断言を排除、ユーティリティ内部にのみ集約）
    - ~~`aria-describedby` 統一パターン適用~~ **完了**（全フォーム入力フィールドに `FormFieldError` との `aria-describedby` 接続、`PatientInfoCard.staffAriaDescribedBy`・`MasterSelectTrigger.ariaDescribedBy` prop 追加）
    - ~~`MasterSelectTrigger`/`MasterSelectModal` の `<div onClick>` → `<button>` 変換~~ **完了**（キーボードアクセシビリティ改善）
    - ~~`useDragAndDrop` の `useAnnounce` 統合~~ **完了**（キーボード・マウス全D&D操作にスクリーンリーダー通知を追加）
    - ~~`SettingsContent.tsx` (693行) のコンポーネント分割~~ **完了**
    - ~~`useDragAndDrop.ts` (525行) のフック分割~~ **完了**
    - ~~不要コード除去の再検証~~ **完了**
    - ~~**Notion風UI統一（全フォーム）**~~ **完了**（`NotionPropertyRow`・`NotionSectionLabel`・`NotionSectionDivider` の3共通コンポーネントを `/components/shared/NotionPropertyRow.tsx` に作成。全フォームで `STYLE.propertyRow` + `STYLE.propertyInput` + `STYLE.selectCompact` パターンに統一）
    - ~~**React.lazy() コード分割**~~ **完了**（30ルートコンポーネントを `lazyNamed` ユーティリティ経由で遅延読み込み化。`/lib/lazy-route.tsx` に `lazyNamed`・`RouteFallback`・`SuspenseRoute` を作成。初期バンドルサイズを大幅削減）
4. ~~**シフト管理**~~ **完了**（週間・月間カレンダー切替ビュー、スタッフ×曜日グリッド編集、ロールフィルタ、週計労働時間40h超過警告、Popover編集、ルート・サイドバー接続済み）
5. ~~**認証機能**~~ **Phase 1〜3 完了**（詳細設計は `docs/AUTH.md` を参照。3層モデル: ユーザー種別・職種・権限、マルチクリニック対応、RLSポリシー設計済み。Phase 1: `AuthProvider`+`LoginForm`+モック認証+6デモアカウント、Phase 2: `ClinicSwitcher`サイドバー統合+ユーザー情報表示、Phase 3: サイドバー13メニュー項目の権限フィルタ+全41対象ルートに`ProtectedRoute`権限ガード適用。Phase 4: バックエンド接続時のRLSポリシー適用は未実施）
6. ~~**印刷機能**~~ **完了**（`usePrint` 汎用フック・`PrintPreviewDialog` 共通コンポーネント・入院サマリ/会計帳票・`data-print-area` 属性による印刷分離・`@media print` CSS拡充・印刷ボタン統合）

## 10. 関連ドキュメント

| ドキュメント | パス | 説明 |
|---|---|---|
| **デザインシステム** | `docs/DESIGN_SYSTEM.md` | カラーパレット、タイポグラフィ、コンポーネントスタイリング規約、予約カレンダーFigma準拠パステルカラー |
| **画面仕様書** | `docs/SCREENS.md` | 全42ルートの画面仕様・構成・データフロー・操作仕様 |
| **ER図** | `docs/ERD.md` | 全45テーブルの定義・リレーション・列挙型一覧（Mermaid記法） |
| **ER図** | `docs/ERD.md` | 全45テーブルの定義・リレーション（v5.0 最新版） |
| **認証・認可設計書** | `docs/AUTH.md` | 認証フロー・RBAC（3層モデル: ユーザー種別/職種/権限）・マルチクリニック設計・RLSポリシー・フロントエンド実装方針 |
