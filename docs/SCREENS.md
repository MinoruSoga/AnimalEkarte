# 動物病院管理システム 画面仕様書

> **バージョン**: v3.0（2026-03-12更新）
> **参照元**: `ui-sample/src/SCREENS.md`（フロントエンドプロトタイプ仕様）

本ドキュメントは、全画面（ルート）ごとの仕様を定義します。
各画面のルートパス、目的、構成コンポーネント、データフロー、ユーザー操作を網羅しています。

> **デザイン・UI共通仕様**:
> - 操作ボタンはNotion風に統一（メイン: `#038B94`, プライマリ: `#2383E2`）
> - セレクトボックスは全て `Combobox` (コンボボックス) に統一
> - 各行の操作アイコンは `MoreHorizontal` + `size-5` を使用
> - 内部ID（〇〇No, 〇〇ID等）はUI上非表示とする
> - 会計・入院機能はLocalStorageによるデータ永続化を採用

---

## 凡例

| 記号 | 意味 |
|------|------|
| `[R]` | ルートコンポーネント (`routes/` 配下) |
| `[C]` | 機能固有コンポーネント (`components/` 配下) |
| `[S]` | 共有コンポーネント (`/components/shared/`) |
| `[H]` | フック (`hooks/` 配下) |
| `[M]` | モーダル / ダイアログ |

---

## 1. ダッシュボード

### 1.1 ダッシュボード（カンバンボード）

| 項目 | 内容 |
|------|------|
| **ルート** | `/` |
| **コンポーネント** | `[R] Dashboard` |
| **目的** | 病院全体の来院フローをカンバンボード形式で俯瞰管理する |

**画面構成:**
- カンバンボード（5カラム: 受付予約 → 受付済 → 診療中 → 会計待ち → 会計済）
- 各カラムに患者カード（`[C] AppointmentCard`）を配置
- カード間のドラッグ&ドロップによるステータス遷移（`react-dnd`）
- カードクリックで詳細モーダル（`[C] DashboardDetailModal`）表示

**AppointmentCard 表示項目:**
| 項目 | 表示内容 |
|---|---|
| 時刻 | `appointment.time`（Clock アイコン付き、等幅フォント） |
| 次回予約 | `nextAppointment` バッジ（「次回予約済」=secondary / 「精算未確認」=destructive＋AlertCircle） |
| 飼い主名 | `ownerName`（太字） |
| ペット | `petType - petName`（Dog アイコン付き） |
| 初診/再診 | `visitType` バッジ（初診=青背景 / 再診=スレート背景） |
| 診療区分 | `serviceType` バッジ（キーワード自動アイコン: トリミング→Scissors / 予防接種→Syringe / 手術→Activity / 診療→Stethoscope） |
| 担当医 | `doctor` バッジ（指名時は「指」ラベル＋オレンジ背景、無効スタッフ時は赤背景＋AlertCircle） |

**DashboardDetailModal 表示項目:**
| セクション | 項目 |
|---|---|
| ヘッダー | 初診/再診アイコン（初/再）、診療区分名、ステータスバッジ（カラム別カラー） |
| 時間カード | 時刻（等幅・大文字）、nextAppointment バッジ |
| 患者情報 | ペット名、ペット種、飼い主名 |
| 診療詳細 | 担当医（未定表示あり）、指名バッジ |

**DashboardDetailModal ステータス別アクション:**
| ステータス | アクションボタン |
|---|---|
| 受付予約 | 取消、編集、飼主詳細、「受付済にする」 |
| 受付済 | 飼主詳細、「カルテ作成」（診療時、同時に診療中へ移動）/ 「診察を開始する」（非診療時） |
| 診療中 | 飼主詳細、「診察を終了する」、「カルテ入力」、「検査」（診療時）/ 「施術記録」（トリミング時） |
| 会計待ち | 飼主詳細、「会計へ進む」（精算未確認時はdisabled） |
| 会計済 | 飼主詳細、「完了/リストから削除」 |

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `Dashboard` | `[R]` | メインページ |
| `KanbanColumn` | `[C]` | カラム（受付予約〜会計済） |
| `AppointmentCard` | `[C]` | 患者カード（DnD対応） |
| `DashboardDetailModal` | `[C][M]` | 患者詳細モーダル |
| `DashboardSummaryWidget` | `[C]` | 統計サマリーウィジェット（カラム別件数カード＋ネイティブSVGミニスパークライン、ホバーツールチップ） |
| `HospitalizationAlertWidget` | `[C]` | 入院アラートウィジェット（入院中件数・退院超過アラート・入院管理へのクイックリンク） |
| `useDashboardKanban` | `[H]` | カンバン状態管理 |
| `useDashboardWeeklyStats` | `[H]` | 週次統計データ収集（localStorage永続化、スパークラインデータ提供） |

**データ型:** `Appointment`, `ColumnData`, `DashboardColumnTitle`, `WeeklyChartPoint`, `WeeklyStatsResult`

**ユーザー操作:**
- カードのドラッグ&ドロップでステータスを変更
- カードクリックで詳細モーダルを開く
- 詳細モーダルからカルテ・予約・会計画面へ遷移

---

## 2. 予約管理

### 2.1 予約カレンダー

| 項目 | 内容 |
|------|------|
| **ルート** | `/reservations` |
| **コンポーネント** | `[R] ReservationManagement` |
| **目的** | 診療・トリミング・手術の予約をカレンダー形式で一元管理する |

**画面構成:**
- ヘッダー: 日付ナビゲーション（前/次）、今日ボタン、表示切替（月/週）、新規予約ボタン
- カレンダービュー（`[C] MonthView` / `[C] WeekView` の切替）
- 予約クリックで詳細モーダル → 編集/キャンセル
- 新規作成で予約フォームモーダル

**ツールバー項目:**
| 項目 | 説明 |
|---|---|
| 日付ナビゲーション | 前/次（月単位 or 週単位）、今日ボタン |
| 年月表示 | `yyyy年 M月` 形式 |
| 予約種別カラー凡例 | 動的マスタ連動（`ServiceTypeMaster.color`）。Figma準拠パステルカラー：診療=#dbeafe(bg)/#5b8def(text) / 検診=#dcfce7/#16a34a / 手術=#ffe2e2/#f87171 / ワクチン=#f3e8ff/#a855f7 / 入院=#cefafe/#0891b2 / トリミング=#ffedd4/#f97316。暗い色は極力使用せず、明るく柔らかいパステルトーンで統一 |
| 担当医フィルタ | Stethoscope アイコン付きSelectドロップダウン。「すべての医師」＋全予約データから動的抽出した医師名一覧 |
| 表示切替 | 月表示 / 週表示 Select |

**MonthView カード表示項目:**
| 項目 | 表示内容 |
|---|---|
| 時刻 | `H:mm` 形式（太字） |
| 初診/再診 | visitType バッジ（初=赤背景 / 再=青背景） |
| ペット名 | `petName` |
| 飼い主名 / 担当医 | `ownerName` + ` / ` + `doctor`（2行目、低opacity） |
| 背景色 | 予約種別に対応したカラー（`getReservationTypeColor`） |
| 件数制限 | 1日最大4件表示、超過分は「他 N 件」表示 |

**WeekView カード表示項目:**
| 項目 | 表示内容 |
|---|---|
| 時刻 | `H:mm` 形式（太字） |
| 初診バッジ | visitType が first の場合のみ「初」バッジ（赤背景） |
| ペット名 | `petName`（太字） |
| 飼い主名 | `ownerName`（高さ36px超で表示、低opacity） |
| 担当医 | `doctor`（高さ52px超で表示、低opacity） |
| 予約種別名 | `getReservationTypeName(type)`（高さ68px超で表示） |
| ステータスドット | checked_in=blue / in_consultation=purple / accounting=orange / completed=gray / cancelled=red（右上に丸ドット表示） |
| 背景色 | 予約種別に対応したカラー |
| ドラッグ&ドロップ | Y軸ドラッグで時間変更（15分単位スナップ、`motion/react` 使用） |
| 重複処理 | 同時間帯の予約は横並び分割表示（カラム計算アルゴリズム） |
| 現在時刻線 | 当日列に赤い水平線で現在時刻を表示 |

**ReservationFormFields フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 日付 | `Popover` + `Calendar` | 日付選択 |
| 時間帯（開始） | `Select`（30分刻み 0:00〜23:30） | Clock アイコン付き |
| 時間帯（終了） | `Select`（30分刻み 0:00〜23:30） | ArrowRight で接続 |
| 予約区分 | `Select`（serviceType マスタ連動） | `MasterLink` 付き |
| 初診/再診 | `RadioGroup`（first / revisit） | カスタムラベルUI（カラードット付き） |
| 担当者 | `Select`（staff マスタ連動、active のみ） | `MasterLink` 付き |
| メモ | `Textarea` | |

**ReservationFormModal ウィザード構成:**
- **Step 1: ペット検索・選択**（StepIndicator 1=ペット選択 / 2=予約情報）
  - `PatientSearch`（飼主名・ペット名テキスト検索）
  - `PatientSelectionTable`（検索結果テーブル、行クリックでペット選択）
  - 選択済みペットは `SelectedPetChip`（PawPrint アイコン付き、種バッジ、×ボタン）で表示
  - 編集モード時はStep 1をスキップ
- **Step 2: 予約情報入力**
  - `ReservationFormFields`（上記フォーム項目一式）
  - 新規時: 日付セルクリック日を初期値、デフォルト時間10:00-11:00
  - フッターに「保存」ボタン

**アクセシビリティ:**
- `ReservationFormFields` の日付・予約種別・担当医の各 `SelectTrigger` に `aria-describedby` で `FormFieldError` 接続
- 重複チェックエラー時: 構造化トースト（`toast.warning` + description）

**ReservationDetailModal 表示項目:**
| セクション | 項目 |
|---|---|
| アクセントバー | visitType に応じた色帯（初診=赤 / 再診=青） |
| ヘッダー | 初診/再診バッジ（丸ドット付き）、予約種別名 |
| ステータスセレクター | 現在ステータスの色帯＋6段階ドロップダウン（予約確定/受付済/診療中/会計待ち/完了/キャンセル、各色ドット付き） |
| 日時カード | 日付（yyyy年 M月 d日 (E)）、時間帯（H:mm – H:mm）、Calendar/Clock アイコン |
| 患者情報 | ペット名（太字）、飼い主名 |
| 診療詳細 | 担当医 + 指名バッジ（amber 背景）、予約区分（Tag アイコン付き） |
| メモ | notes（amber 背景カード、FileText アイコン付き、条件表示） |
| フッターアクション | 削除ボタン（ゴミ箱）、編集ボタン、種別別レコード作成ボタン（カルテ作成 / トリミング記録作成 / 入院・ホテル登録） |

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `ReservationManagement` | `[R]` | メインページ（UI層のみ） |
| `useReservationManagement` | `[H]` | 予約CRUD・モーダル・バリデーションロジック |
| `MonthView` | `[C]` | 月表示カレンダー |
| `WeekView` | `[C]` | 週表示カレンダー（`motion/react` アニメーション） |
| `ReservationFormModal` | `[C][M]` | 予約作成/編集モーダル |
| `ReservationFormFields` | `[C]` | フォームフィールド群 |
| `ReservationDetailModal` | `[C][M]` | 予約詳細モーダル |
| `PatientSearch` | `[C]` | 患者検索コンポーネント |
| `PatientSelectionTable` | `[C]` | 患者選択テーブル |
| `ConfirmDialog` | `[S][M]` | キャンセル確認 |

**データ型:** `ReservationAppointment`, `ReservationFormData`, `CalendarView`, `ReservationStatus`

**ユーザー操作:**
- 月/週ビュー切替
- 日付ナビゲーション（前月/翌月、前週/翌週）
- カレンダーセルクリックで新規予約作成
- 既存予約クリックで詳細表示
- 詳細モーダルから編集・キャンセル・関連レコード作成への遷移
- 予約フォームでペット検索・紐付け

---

## 3. 飼主・ペット管理

### 3.1 飼主一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/owners` |
| **コンポーネント** | `[R] OwnersList` |
| **目的** | 飼主・ペット情報の検索・一覧表示 |

**画面構成:**
- ヘッダー: タイトル + 新規登録ボタン
- 検索バー（`[S] SearchFilterBar`）+ 件数表示
- データテーブル（`[S] DataTable`）: 飼主名、ペット名、生死、種、生年月日、体重、環境、前回来院、操作
- ページネーション（`[S] Pagination`、20件/ページ）
- 行アクション: 編集、削除（`[S] RowActionDropdown`）

**データ型:** `Pet`（ペット単位で表示、飼主情報を含む）

**ユーザー操作:**
- テキスト検索（飼主名、ペット名、ID等）
- 行クリックで飼主編集画面へ遷移
- 行ドロップダウンから編集/削除
- 新規登録ボタンで作成画面へ

### 3.2 飼主登録/編集

| 項目 | 内容 |
|------|------|
| **ルート** | `/owners/new`（新規）/ `/owners/:id`（編集） |
| **コンポーネント** | `[R] OwnerForm` |
| **目的** | 飼主情報の入力・更新、配下ペットの管理 |

**OwnerForm フォーム項目（4カラムグリッド）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 郵便番号 | `Input` | placeholder: 123-4567 |
| 会社名 | `Input` | |
| 会員区分 | `Button` グループ | `MEMBERSHIP_TYPE_VALUES`（非会員/会員/退亡者/他診/準） |
| 飼主名 | `Input` | 必須 |
| 住所1 | `Input` | |
| 電話番号(自宅) | `Input` | |
| 危険人物 | `Switch` | 「該当する」ラベル |
| 飼主名(カナ) | `Input` | 必須 |
| 住所2 | `Input` | |
| 自宅住所1 | `Input` | |
| 備考・特記事項 | `Textarea` | rowspan、min-h-[140px] |
| 飼主生年月日 | `NotionDatePicker` | |
| メールアドレス | `Input`（type=email） | |
| 自宅住所2 | `Input` | |
| 電話番号 | `Input` | 必須、placeholder: 090-1234-5678 |
| 会社 電話番号 | `Input` | colspan=2 |
| 値引率 (%) | `Input`（type=number） | |

**PetEditModal フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| ペット名 | `Input` | |
| ペット名カナ | `Input` | |
| 種 | `Select`（犬/猫/その他） | `PET_SPECIES_VALUES` |
| 性別 | `Select`（雄/雌/不明） | `PET_GENDER_VALUES` |
| 生年月日 | `NotionDatePicker` | |
| 品種 | `Input` | |
| 毛色 | `Input` | |
| 避妊去勢日 | `NotionDatePicker` | |
| 入手種別 | `Select`（購入/譲渡/保護/その他） | `ACQUISITION_TYPE_VALUES` |
| 危険度 | `Select`（低/中/高） | `DANGER_LEVEL_VALUES` |
| フード | `Input` | |
| 保険名 | `Select`（アニコム/アイペット/ペット＆ファミリー/楽天/アクサ/SBI/FPC/その他） | `INSURANCE_COMPANY_VALUES` |
| 保険詳細(負担割合) | `Select`（50%/70%/90%/100%/その他） | `PET_INSURANCE_RATIO_VALUES` |
| 備考・特記事項 | `Textarea` | |

**アクセシビリティ:**
- 飼主名・飼主名カナ・電話番号: `aria-invalid` + `aria-describedby` → `FormFieldError`（`role="alert"`）接続
- ペット編集モーダル: ペット名に `aria-invalid` + `aria-describedby` → `FormFieldError` 接続

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `OwnerForm` | `[R]` | メインページ |
| `PetEditModal` | `[C][M]` | ペット追加/編集モーダル |
| `PageLayout` | `[S]` | ページレイアウト |
| `NavigationBlocker` | `[S]` | フォーム離脱保護 |
| `NotionDatePicker` | `[S]` | 日付ピッカー |
| `ConfirmDialog` | `[S][M]` | ペット削除確認 |
| `useOwnerForm` | `[H]` | フォーム状態管理 |
| `useUnsavedChanges` | `[H]` | 未保存検知 |

**データ型:** `OwnerData`, `PetInfo`, `PetFormData`, `PetEditModalData`, `MembershipType`

**ユーザー操作:**
- 飼主情報の入力・保存
- ペット追加/編集（モーダル）
- ペット削除（確認ダイアログ）
- ペット行ドロップダウンから各機能画面へ遷移
- 未保存状態での離脱時に確認ダイアログ表示

---

## 4. 電子カルテ

### 4.1 カルテ一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/medical-records` |
| **コンポーネント** | `[R] MedicalRecords` |
| **目的** | 全カルテの検索・一覧表示 |

**画面構成:**
- ヘッダー: タイトル + 新規作成ボタン
- 検索バー（`[S] SearchFilterBar`）
- データテーブル: 診療日、飼主名、ペット名、種、主訴、担当医、ステータス、操作
- ページネーション（20件/ページ）
- 行アクション: 編集、削除

**データ型:** `MedicalRecord`
**ステータス:** 作成中 / 確定済

**ユーザー操作:**
- テキスト検索
- 行クリックで編集画面へ
- 新規作成 → ペット選択画面（`/medical-records/select-pet`）へ
- 担当医が無効（マスタで非アクティブ）の場合、警告アイコン表示

### 4.2 カルテ用ペット選択

| 項目 | 内容 |
|------|------|
| **ルート** | `/medical-records/select-pet` |
| **コンポーネント** | `[R] MedicalRecordPetSelection` |
| **目的** | カルテ作成対象のペットを検索・選択する |

**画面構成:**
- `[S] PetSearchForm` + `[S] PetSearchResultsTable` 共通コンポーネントを使用
  - ペット検索フォーム（飼主ID、飼主名、電話、ペット名、種）
  - 検索結果テーブル
  - 選択ボタンで `/medical-records/new?petId=xxx` へ遷移

### 4.3 カルテ入力/編集

| 項目 | 内容 |
|------|------|
| **ルート** | `/medical-records/new`（新規）/ `/medical-records/:id`（編集） |
| **コンポーネント** | `[R] MedicalRecordForm` |
| **目的** | 診療記録の全項目を9タブ構成で入力・編集する |

**画面構成:**
- スティッキーヘッダー:
  - 患者情報カード（`[S] PatientInfoCard`）: ペット名、種、飼主名、担当医、診療区分、バイタル入力ボタン
  - タブバー（9タブ）
- タブコンテンツ（遅延マウント: 一度表示したタブは保持）

**タブ詳細:**

| # | タブ名 | コンポーネント | 説明 |
|---|--------|---------------|------|
| 1 | **問診** | `MedicalRecordInterview` | 主訴（Markdown入力）、問診履歴表示 |
| 2 | **診察/治療プラン** | `MedicalRecordDiagnosisPlan` | バイタル入力・履歴グラフ、診断登録（カテゴリ+診断名）、治療プラン（TreatmentTable） |
| 3 | **治療** | `MedicalRecordTreatment` | 処置完了記録テーブル（TreatmentTable）、検索ダイアログでマスタ連携 |
| 4 | **予防接種** | `MedicalRecordVaccination` | 予防接種フォーム + 接種履歴一覧 |
| 5 | **定期健診** | `MedicalRecordCheckup` | 健診種別・実施日・結果登録フォーム + 健診履歴一覧 |
| 6 | **検査** | `MedicalRecordExamination` | 検査オーダーフォーム + 結果履歴 |
| 7 | **画像** | `MedicalRecordImage` | 画像アップロード + フィルタ付きギャラリー |
| 8 | **見積書** | `MedicalRecordEstimate` | 診療内容に基づく概算見積 |
| 9 | **会計(医師確認)** | `MedicalRecordBillCheck` | 算定チェック・確認 |

**Tab 1: 問診（`MedicalRecordInterview`）**
- 3カラムレイアウト（lg:12グリッド = 3+4+5）
- **左カラム**: 主訴入力（`InterviewChiefComplaint`）
  - Markdown テキストエリア（テンプレート見出し: どんな症状 / どこが / いつから / その他・備考 / フリースペース）
  - テンプレート挿入ボタン（定期検診 / 予防接種 / 下痢・嘔吐 / 皮膚）
- **中カラム**: 治療方針（`InterviewTreatmentPolicy`）
  - Markdown テキストエリア
- **右カラム**: カルテ履歴（`InterviewHistory`）
  - 履歴リスト: 日付、担当医、診療種別バッジ、タイトル、内容

**Tab 2: 診察/治療プラン（`MedicalRecordDiagnosisPlan`）**
- **診断ヘッダー**（`DiagnosisHeader`、3カラム）:
  - `DiagnosisHeaderChiefComplaint`: 主訴の読み取り専用表示
  - `DiagnosisHeaderPhysicalExam`: 身体所見 Markdown エリア
  - `DiagnosisHeaderDiagnosis`: 診断登録フォーム
    - `diagnosisDetails`: 診断詳細（Markdown）
    - `diagnosis1Category` / `diagnosis1Name`: 診断1（カテゴリ Select + 診断名 Select）
    - `diagnosis2Category` / `diagnosis2Name`: 診断2（同上）
- **治療プラン**: `TreatmentTable`（マスタ検索ダイアログ連携）
- **集計**: `TreatmentDetailedSummary`（小計、税、合計、割引率、値引額）

**Tab 3: 治療（`MedicalRecordTreatment`）**
- **治療プランテーブル**: `TreatmentTable`（`onMarkCompleted` チェックボックス列あり。チェック→確認ダイアログ→治療済みテーブルへ移動）
- **治療済みテーブル**: `TreatmentTable`（`onRevertToPlan` 戻すボタンあり。確認ダイアログ→治療プランへ差し戻し）
- **移動確認ダイアログ**: `TreatmentMoveConfirmDialog`（`AlertDialog`ベース、`pendingMove`ステート制御）
- **集計**: `TreatmentDetailedSummary`
- TreatmentTable 列: [済チェックボックス]、治療内容、メモ、保険、単価(税込)、数量、割引(%)、値引(￥)、小計、操作

**Tab 4: 予防接種（`MedicalRecordVaccination`）**
- 2カラムレイアウト（lg:6+6）
- **左カラム** `VaccinationForm` フォーム項目:

| フィールド | 入力部品 | 備考 |
|---|---|---|
| 予防接種名 | `MasterSelectModal`（vaccine マスタ連動） | Syringe アイコン、`MasterLink` 付き |
| 予防接種日 | `NotionDatePicker` | |
| 担当医 | `Select`（staff マスタ連動、active のみ） | `MasterLink` 付き |
| 補助説明 | `Input` | |
| LOT1〜LOT4 | `Input` × 4（4カラムグリッド） | |
| 次回予防接種予定設定 | `RadioGroup`（`NEXT_SCHEDULE_TYPE_VALUES`） | |
| 次回予定日 | `NotionDatePicker` | |
| 備考 | `Textarea` | |

- **右カラム** `VaccinationHistory`: 接種履歴一覧

**Tab 5: 検査（`MedicalRecordExamination`）**
- 2カラムレイアウト（lg:6+6）
- **左カラム** `ExaminationForm` フォーム項目:

| フィールド | 入力部品 | 備考 |
|---|---|---|
| 検査種別 | `MasterSelectModal`（examination マスタ連動） | 必須、FlaskConical アイコン、`MasterLink` 付き |
| 担当医 | `Select`（staff マスタ連動、active のみ） | `MasterLink` 付き |
| 検査項目テーブル | `Table`（マスタ連動で自動生成） | 列: 項目名、測定値（Input）、単位、基準値 |
| 備考・所見 | `Textarea` | |
| アクション | 「結果を登録」ボタン + 「クリア」ボタン | |

- **右カラム** `ExaminationHistory`: 検査結果履歴

**Tab 6: 画像（`MedicalRecordImage`）**
- **フィルタバー**（`ImageGalleryFilter`）: キーワード検索、日付範囲（開始/終了）、ソート順（昇順/降順）、アップロードボタン
- **検査結果セクション**: 日付別画像グループ（`ImageGalleryGroup`）

**Tab 7: 見積書（`MedicalRecordEstimate`）**
- **件名**: `EstimateForm`（件名 Input）
- **明細テーブル**: `TreatmentTable`（チェックボックス列なし・戻すボタンなし）
- **集計**: `TreatmentDetailedSummary`
- **コメント / 備考**: 2カラム `Textarea`
- **アクション**: 「PDF出力」ボタン

**Tab 8: 会計(医師確認)（`MedicalRecordBillCheck`）**
- **明細テーブル**: `TreatmentTable`（治療タブの completedItems を自動同期）
- **集計**: `TreatmentDetailedSummary`
- **固定フローティングアクション**:
  - 「チェック完了」ボタン: クリックでトグル動作。チェック完了状態では「未チェックに戻す」に変わり、再クリックで「チェック完了」に戻る。トースト通知でステータス変更を表示
  - 「会計へ進む」ボタン / 「会計を確認」ボタン:
    - 新規カルテ（linkedAccountingId なし）: 「会計へ進む」（Receipt アイコン付き、items 空時は disabled）
    - 既存カルテ（linkedAccountingId あり）: 「会計を確認」（緑背景、既存会計詳細へ遷移）
- **会計遷移**: カルテの明細を `AccountingItem[]` に自動変換し、カテゴリを自動推定（検査/処方/手術/処置/フード等のキーワードマッチ）して会計画面へ state 経由で渡す

**VitalInputDialog フォーム項目（カルテ共通）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 体温 | `Input`（number, step=0.1） | Thermometer アイコン、単位: ℃ |
| 心拍数 | `Input`（number） | Heart アイコン、単位: /min |
| 呼吸数 | `Input`（number） | Wind アイコン、単位: /min |
| 体重 | `Input`（number, step=0.01） | Weight アイコン、単位: kg |
| メモ | `Textarea` | StickyNote アイコン |

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `MedicalRecordForm` | `[R]` | メインページ |
| `PatientInfoCard` | `[S]` | 患者情報カード |
| `OwnerQuickViewModal` | `[S][M]` | 飼主情報クイックビュー |
| `MasterSelectModal` | `[S][M]` | 診療区分・担当医選択 |
| `VitalInputDialog` | `[C][M]` | バイタル入力ダイアログ |
| `TreatmentTable` | `[C]` | 治療項目テーブル |
| `TreatmentSearchDialog` | `[C][M]` | マスタ検索ダイアログ |
| `InterviewChiefComplaint` | `[C]` | 主訴入力（Markdown） |
| `InterviewHistory` | `[C]` | 問診履歴 |
| `DiagnosisHeader` | `[C]` | 診断ヘッダー |
| `ExaminationForm` | `[C]` | 検査フォーム |
| `ExaminationHistory` | `[C]` | 検査結果履歴 |
| `VaccinationForm` | `[C]` | 予防接種フォーム |
| `VaccinationHistory` | `[C]` | 接種履歴 |
| `EstimateForm` | `[C]` | 見積フォーム |
| `NavigationBlocker` | `[S]` | フォーム離脱保護 |
| `PrintPreviewDialog` | `[S][M]` | 印刷プレビューダイアログ |
| `useMedicalRecordForm` | `[H]` | メインフォームフック |
| `useMedicalRecordInit` | `[H]` | 初期化ロジック |
| `usePrint` | `[H]` | 印刷状態管理フック |

**データ型:** `MedicalRecord`, `TreatmentItem`, `VitalEntry`, `DiagnosisFormData`, `ExaminationFormData`, `ExaminationResultGroup`, `VaccinationFormData`, `VaccinationHistoryItem`, `InterviewHistoryItem`, `MrDocumentType`

**ユーザー操作（要medical権限）:**
- タブ切替で各セクションの入力
- 患者情報カードから診療区分・担当医をマスタ選択
- バイタル入力ダイアログでバイタル記録
- 治療テーブルでマスタ検索→項目追加
- 保存ボタンで確定（バリデーション実行）
- 削除ボタンでカルテ削除（確認ダイアログ）
- 未保存離脱時の保護ダイアログ
- 確定済みカルテ: 印刷（`usePrint<MrDocumentType>` + `MR_DOCUMENT_TYPE_LABELS` で動的タイトル）

**印刷機能:**
- ヘッダーアクションに印刷ボタン表示（確定済み時のみ）
- `PrintPreviewDialog` でプレビュー表示、`window.print()` で印刷実行
- 印刷エリア（`hidden print:block`、`data-print-area` 属性）に帳票を配置

---

## 5. 入院管理

### 5.1 入院一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/hospitalization` |
| **コンポーネント** | `[R] HospitalizationList` |
| **目的** | 入院・ホテル患者の一覧管理、ボード/リスト表示切替 |

**画面構成:**
- ヘッダー: タイトル + 新規登録ボタン
- フィルタ: ステータス（全て/入院中/退院済/予約）タブ、ボード/リスト表示切替
- 検索バー
- **ボードビュー**（`[C] HospitalizationBoard`）: ケージごとのカード表示
  - `[C] CageDragPreview`: `useDragLayer` によるカスタムドラッグプレビューオーバーレイ（ケージ間移動時に半透明カードがカーソルに追従）
- **リストビュー**（`[C] HospitalizationListView`）: データテーブル表示
- ページネーション（リストビュー時、20件/ページ）

**データ型:** `Hospitalization`, `HospitalizationFilterStatus`, `HospitalizationViewMode`

**ユーザー操作:**
- ステータスフィルタ切替
- ボード/リスト表示切替
- カード/行クリックで詳細画面へ
- ボードビューでケージ間のペット移動

### 5.2 入院用ペット選択

| 項目 | 内容 |
|------|------|
| **ルート** | `/hospitalization/select-pet` |
| **コンポーネント** | `[R] HospitalizationPetSelection` |
| **目的** | 入院対象ペットの検索・選択 |

**画面構成:** `[S] PetSearchForm` + `[S] PetSearchResultsTable` 共通コンポーネント使用

### 5.3 入院登録/編集

| 項目 | 内容 |
|------|------|
| **ルート** | `/hospitalization/new`（新規）/ `/hospitalization/:id/edit`（編集） |
| **コンポーネント** | `[R] HospitalizationForm` |
| **目的** | 入院情報の入力・治療プラン管理 |

**HospitalizationBasicInfo フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 入院タイプ | `RadioGroup`（入院 / ホテル） | `HOSPITALIZATION_TYPE_VALUES` |
| 期間（開始日） | `NotionDatePicker` | Calendar アイコン付き |
| 期間（終了日） | `NotionDatePicker` | |
| ケージ・個室 | `Select`（cage マスタ連動） | `MasterLink` 付き |
| メモ | `Textarea` | |

**HospitalizationNoteCard（2枚）:**
| カード | アイコン | プレースホルダー |
|---|---|---|
| 飼主からのリクエスト | MessageSquare | 「リクエストを入力...」 |
| スタッフへの連絡事項 | AlertCircle | 「連絡事項を入力...」 |

**アクセシビリティ:**
- 担当医エラー: `PatientInfoCard.staffAriaDescribedBy` → `FormFieldError`（`role="alert"`）と `aria-describedby` 接続

### 5.4 入院詳細

| 項目 | 内容 |
|------|------|
| **ルート** | `/hospitalization/:id` |
| **コンポーネント** | `[R] HospitalizationDetail` |
| **目的** | 入院患者のケアプラン管理、デイリーログ記録 |

**画面構成（レスポンシブ: デスクトップ/モバイル分離）:**
- **デスクトップ**: 左カラム（患者ヘッダー＋アクションバー＋ケアプラン）+ 右カラム（デイリーレコード）
- **モバイル**: シングルカラム（患者ヘッダー → ケアプラン → デイリーレコード）

**ケアプラン（`[C] CarePlan/`）:**
| コンポーネント | 説明 |
|---|---|
| `CarePlanPreviewPopover` | ケアプラン概要ポップオーバー（ステータストグル付き） |
| `CarePlanSection` | ケアプラン一覧表示 |
| `CarePlanItemRow` | ケアプラン項目行 |
| `CarePlanDialog` | ケアプラン追加/編集ダイアログ |

**デイリーレコード（`[C] DailyRecord/`）:**
| コンポーネント | 説明 |
|---|---|
| `DailyRecordSection` | デイリーレコードメインセクション |
| `DateNavigation` | 日付ナビゲーション |
| `TimingSection` | 朝/昼/夜のタスク区分表示 |
| `TaskCompleteDialog` | タスク完了ダイアログ |
| `VitalDialog` | バイタル入力ダイアログ |
| `LogDialog` | ケアログ入力ダイアログ |
| `SimpleNoteForm` | スタッフメモ入力 |
| `Timeline` | 実施記録の時系列表示 |

**CarePlanDialog フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| マスタ引用 | 「マスタ検索」ボタン → `TreatmentSearchDialog` | 処置・検査・薬をマスタから検索して自動入力 |
| 種類 | `Select`（`CARE_PLAN_TYPE_VALUES`: 食事/投薬/処置・検査/処置・指示/持ち物・その他） | |
| 名称 | `Input` | |
| 詳細・指示量 | `Input` | 例: 30g / 1錠 / 左前肢 |
| タイミング | トグルボタン（`PLAN_TIMING_VALUES`: 朝/昼/夜） | 複数選択可 |
| メモ・特記事項 | `Textarea` | |
| ステータス | `Select`（`CARE_PLAN_STATUS_VALUES`） | |

**データ型:** `Hospitalization`, `CarePlanItem`, `DailyRecord`, `VitalRecord`, `CareLogRecord`, `StaffNoteRecord`, `Task`, `TimelineItem`, `HospDocumentType`

**ユーザー操作:**
- ケアプランの追加/編集/削除（カテゴリ: 食事、投薬、処置・検査、処置・指示、持ち物・その他）
- デイリーレコード日付ナビゲーション
- タスク完了チェック（朝/昼/夜のタイミングごと）
- バイタル記録・ケアログ記録（食事、排泄、投薬、処置、その他）
- スタッフメモ追加
- タイムラインで時系列確認
- 退院処理（確認ダイアログ）
- 入院サマリーの印刷プレビュー → 印刷

---

## 6. トリミング

### 6.1 トリミング一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/trimming` |
| **コンポーネント** | `[R] TrimmingList` |
| **目的** | トリミング予約の検索・一覧管理 |

**フィルタ項目:**
| 項目 | 入力部品 | 備考 |
|---|---|---|
| 開始日 | `NotionDatePicker` | `lg:w-[160px]` |
| 終了日 | `NotionDatePicker` | `lg:w-[160px]`、「〜」で接続 |
| キーワード | `SearchFilterBar` | 飼主名・ペット名で検索 |
| クリア | `Button`（outline） | 全フィルタリセット |

**テーブル列:**
| 列 | 表示内容 |
|---|---|
| 診療日 | `record.date`（等幅フォント） |
| 飼主名 | `record.ownerName` |
| ペット名 | `record.petName` + `record.petNumber`（2行表示） |
| 種 | `record.species` |
| 体重 | `record.weight` |
| スタイル希望 | `record.styleRequest`（truncate、max-w-[200px]） |
| 担当 | `record.staff`（無効スタッフ時は赤文字＋AlertTriangle） |
| ステータス | `StatusBadge`（`getTrimmingStatusColor`） |
| 操作 | `RowActionDropdown`（編集 / 削除） |

**データ型:** `TrimmingRecord`, `DataTableColumn`
**ステータス:** 予約 / 進行中 / 完了

### 6.2 トリミング用ペット選択

| 項目 | 内容 |
|------|------|
| **ルート** | `/trimming/select-pet` |
| **コンポーネント** | `[R] TrimmingPetSelection` |
| **画面構成** | `[S] PetSearchForm` + `[S] PetSearchResultsTable` 共通コンポーネント使用 |

### 6.3 トリミング登録/編集

| 項目 | 内容 |
|------|------|
| **ルート** | `/trimming/new`（新規）/ `/trimming/:id`（編集） |
| **コンポーネント** | `[R] TrimmingForm` |
| **目的** | トリミング情報の入力・編集 |

**左カラム フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| コース選択 | `MasterSelectTrigger` → `MasterSelectModal`（trimming_course マスタ連動） | `MasterLink` 付き、選択時に `charge` 自動反映 |
| スタイルの希望 | `Textarea` | min-h-[80px] |
| メモ | `Textarea` | min-h-[80px] |
| オプション | `Checkbox` グリッド（2cols） | trimming_option マスタ連動、複数選択可、各項目に `+¥{price}` 表示 |
| 希望スタイル画像 | `file input` + プレビュー | ドラッグ&ドロップUI、h-[180px] |

**中カラム フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| BW（体重） | `Input` + `radio`（Kg/g） | `BODY_WEIGHT_UNIT_VALUES` |
| BT（体温） | `Input` | |
| USED SHAMPOO | `Input` | |
| USED RIBBON | `Input` | |
| TREATMENT | `Input` | |
| 備考 | `Input` | |
| 完成画像 | `file input` + プレビュー | h-[180px] |

**右カラム（トリミング履歴）:**
- `[S] HistoryFilterPanel`: 日付範囲、キーワード検索、ソート順（昇順/降順）、クリアボタン
- 履歴カード: 診療日、コース名バッジ、作成者/更新者・日時、スタイルの希望、メモ、画像セクション

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `TrimmingForm` | `[R]` | メインページ |
| `PatientInfoCard` | `[S]` | 患者情報カード |
| `MasterSelectModal` | `[S][M]` | コース・スタッフ選択 |
| `MasterSelectTrigger` | `[S]` | コース選択トリガー |
| `HistoryFilterPanel` | `[S]` | 履歴フィルタパネル |
| `MasterLink` | `[S]` | マスタ設定リンク |
| `NavigationBlocker` | `[S]` | フォーム離脱保護 |
| `ConfirmDialog` | `[S][M]` | 削除確認 |
| `useTrimmingForm` | `[H]` | フォーム状態管理 |

**データ型:** `TrimmingFormData`, `TrimmingParts`, `TrimmingHistoryItem`, `BodyWeightUnit`, `SortOrder`

---

## 7. 検査管理

### 7.1 検査一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/examinations` |
| **コンポーネント** | `[R] Examinations` |
| **目的** | 検査オーダー・結果の一覧管理 |

**テーブル列:**
| 列 | 表示内容 |
|---|---|
| 日時 | `r.date`（等幅フォント） |
| 飼主名 | `r.ownerName` |
| ペット名 | `r.petName` |
| 検査種別 | `r.testType` |
| 結果概要 | `r.resultSummary`（truncate、未入力時「-」） |
| 担当医 | `r.doctor`（無効スタッフ時は赤文字＋AlertTriangle） |
| ステータス | `StatusBadge`（`getExaminationStatusColor`） |
| 操作 | `RowActionDropdown`（「カルテを開く」のみ） |

**特記:** 検査の新規作成はカルテ内の検査タブから行う。一覧画面は参照+カルテ遷移のみ。行クリックでカルテの検査タブへ遷移する。

**データ型:** `ExaminationRecord`, `ExaminationRecordItem`, `DataTableColumn`
**ステータス:** 依頼中 / 検査中 / 完了

---

## 8. 会計

### 8.1 会計一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/accounting` |
| **コンポーネント** | `[R] Accounting` |
| **目的** | 会計レコードの一覧管理 |

**テーブル列:**
| 列 | 表示内容 |
|---|---|
| 日時 | `r.scheduledDate`（等幅フォント） |
| 飼主名 | `r.ownerName` |
| ペット名 | `r.petName` + カルテ連携バッジ |
| 請求金額 | `formatCurrency(calculateTotal(r))`（等幅・太字） |
| 支払方法 | `getPaymentMethodLabel(r.payment?.method)` |
| 保険 | 保険名バッジ（保険あり時のみ） |
| ソース | 「入院連携」バッジ（`source === "hospitalization"` 時） |
| ステータス | `StatusBadge` |
| 操作 | `RowActionDropdown`（編集 / 削除） |

**データ型:** `Accounting`, `AccountingStatus`, `PaymentMethod`, `DataTableColumn`
**ステータス:** 未収(`waiting`) / 収済(`completed`) / キャンセル(`cancelled`) / 保留(`pending`)

### 8.2 会計用ペット選択

| 項目 | 内容 |
|------|------|
| **ルート** | `/accounting/select-pet` |
| **コンポーネント** | `[R] AccountingPetSelection` |
| **画面構成** | `[S] PetSearchForm` + `[S] PetSearchResultsTable` 共通コンポーネント使用 |

### 8.3 会計精算（新規/編集）

| 項目 | 内容 |
|------|------|
| **ルート** | `/accounting/new`（新規）/ `/accounting/:id`（編集） |
| **コンポーネント** | `[R] AccountingDetail` |
| **目的** | 診療費の計算、支払い処理、書類発行 |

**明細追加ダイアログ フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 区分 | `Select`（`MANUAL_ITEM_CATEGORY_VALUES`: 療法食・フード / 物販・ケア用品 / その他） | |
| 品目名 | `Input` | |
| 単価 (税込) | `Input`（type=number） | |

**入金パネル（`[C] AccountingPaymentPanel`）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 保険ON/OFF | `Switch` | CardHeader 内トグル |
| 負担割合 | `Select`（50%/70%/90%/100%） | 保険ON時のみ |
| 支払方法 | `Button` グリッド（3cols） | 現金/カード/電子マネー |
| お預かり金額 | `Input`（type=number） | 右寄せ大文字 |
| クイック入力 | `Button` × 3 | 「丁度」「千円単位」「一万単位」 |
| お釣り | 読み取り専用 | 不足時は赤文字 |
| 会計確定 | `Button`（full-width） | 条件付きdisabled |

**データ型:** `Accounting`, `AccountingItem`, `PaymentInfo`, `AccountingCalculation`, `DocumentType`, `ItemCategory`, `ManualItemCategory`, `PaymentMethod`, `InsuranceRatio`, `ItemSource`, `TaxRate`

**ユーザー操作:**
- 明細項目の手動追加/削除
- 保険適用切替（Switch）、負担割合選択
- 支払方法選択（現金/カード/電子マネー）
- 預り金入力 + クイック入力ボタン → お釣り自動計算
- 会計確定（確認ダイアログ → ステータス completed へ遷移）
- 精算完了後: 領収書/診療明細書のプレビュー → 印刷

---

## 9. 予防接種管理

### 9.1 予防接種一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/vaccinations` |
| **コンポーネント** | `[R] VaccinationList` |
| **目的** | 予防接種記録の一覧管理 |

**テーブル列:**
| 列 | 表示内容 |
|---|---|
| 実施日 | `r.date`（等幅フォント） |
| 飼主名 | `r.ownerName` |
| ペット名 | `r.petName` |
| 予防接種名 | `r.vaccineName`（太字） |
| 担当医 | `r.doctor`（無効スタッフ時は赤文字＋AlertTriangle） |
| 次回予定 | `r.nextDate`（等幅フォント） |
| 操作 | `RowActionDropdown`（「カルテを開く」のみ） |

**特記:** 予防接種の新規登録はカルテ内の予防接種タブから行う。一覧画面は参照+カルテ遷移のみ。

**データ型:** `VaccinationRecord`, `DataTableColumn`

---

## 10. 定期健診管理

### 10.1 定期健診一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/checkups` |
| **コンポーネント** | `[R] CheckupList` |
| **目的** | 定期健診記録の一覧管理 |

**テーブル列:**
| 列 | 表示内容 |
|---|---|
| 実施日 | `r.date`（等幅フォント） |
| 飼主名 | `r.ownerName` |
| ペット名 | `r.petName` |
| 健診種別 | `r.checkupType`（太字） |
| 結果概要 | `r.result`（未設定時「-」） |
| 担当医 | `r.doctor`（無効スタッフ時は赤文字＋AlertTriangle） |
| 次回予定 | `r.nextDate`（等幅フォント） |
| 操作 | `RowActionDropdown`（「カルテを開く」のみ） |

**特記:** 定期健診の新規登録はカルテから行う。一覧画面は参照+カルテ遷移のみ。

**健診種別:**
| 種別名 | 説明 |
|---|---|
| 年次健康診断 | 1年ごとの定期健康チェック |
| シニア健康診断 | 高齢ペット向け（半年ごと推奨） |
| パピー健診 | 幼齢ペット向け成長確認（3ヶ月ごと推奨） |

**データ型:** `CheckupRecord`, `DataTableColumn`

---

## 11. 設定・マスタ管理

### 11.1 マスタ設定トップ

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings` |
| **コンポーネント** | `[R] MasterSettingsIndex` |
| **目的** | マスタカテゴリ一覧をカード形式で表示 |

**セクション構成:**
| # | セクション名 | カテゴリキー |
|---|---|---|
| 1 | 基本設定 | `clinic` |
| 2 | 診療関連マスタ | `serviceType`, `consultation`, `examination`, `procedure`, `vaccine`, `medicine`, `diagnosis_category`, `diagnosis_name` |
| 3 | 入院・ケージ管理 | `hospitalization`, `cage` |
| 4 | トリミング関連 | `trimming_course`, `trimming_option` |
| 5 | スタッフ・保険 | `staff`, `insurance` |

**カテゴリカード（`CategoryCard`）表示項目:**
| 項目 | 表示内容 |
|---|---|
| アイコン | `CATEGORY_CONFIG[key].IconComponent`（ホバー時 bg-[#37352F] text-white） |
| カテゴリ名 | `cfg.label`（truncate） |
| 説明 | `cfg.description`（line-clamp-2） |
| 件数 | `{count}件登録済`（マスタカテゴリのみ） |
| 矢印 | ChevronRight（ホバー時 opacity 変化） |

**データ型:** `MasterCategory`, `MasterCardKey`, `MasterCategoryCard`, `MasterSection`, `CategoryConfig`

### 11.2 病院情報設定

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/clinic` |
| **コンポーネント** | `[R] ClinicSettings` |
| **目的** | 病院の基本情報を管理する |

**フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 病院名 | `Input` | 必須 |
| 支店名 | `Input` | |
| 郵便番号 | `Input` | |
| 住所 | `Input` | |
| 電話番号 | `Input` | |
| FAX番号 | `Input` | |
| 登録番号 | `Input` | |
| 院長名 | `Input` | |
| メールアドレス | `Input`（type=email） | |
| WebサイトURL | `Input` | |

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `ClinicSettings` | `[R]` | メインページ |
| `NotionPropertyRow` | `[S]` | Notion風プロパティ行 |
| `NotionSectionLabel` | `[S]` | Notion風セクションラベル |
| `NotionSectionDivider` | `[S]` | Notion風薄罫線ディバイダー |
| `useClinicInfo` | `[H]` | 病院情報CRUD |

**データ型:** `ClinicInfo`

### 11.3 診療項目マスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/treatment-items` |
| **コンポーネント** | `[R] TreatmentItemsSettings` |
| **目的** | 診察・検査・処置・予防接種・定期健診の5カテゴリを1ページで管理 |

**タブ構成:**
| タブ | カテゴリキー | showPrice | showParentItem |
|---|---|---|---|
| 診察 | `consultation` | true | true |
| 検査 | `examination` | true | true |
| 処置 | `procedure` | true | true |
| 予防接種 | `vaccine` | true | true |
| 定期健診 | `checkup` | true | true |

### 11.4 診断マスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/diagnosis` |
| **コンポーネント** | `[R] DiagnosisSettings` |
| **目的** | 診断カテゴリと診断名の2カテゴリを1ページで管理 |

**タブ構成:**
| タブ | カテゴリキー |
|---|---|
| カテゴリ | `diagnosis_category` |
| 診断名 | `diagnosis_name` |

### 11.5 トリミングマスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/trimming` |
| **コンポーネント** | `[R] TrimmingSettings` |
| **目的** | トリミングコースとオプションの2カテゴリを1ページで管理 |

**タブ構成:**
| タブ | カテゴリキー | showPrice |
|---|---|---|
| コース | `trimming_course` | true |
| オプション | `trimming_option` | true |

### 11.6 マスタカテゴリ設定（個別ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/{category-slug}`（6パターン） |
| **コンポーネント** | `[R] Settings` → `[C] SettingsContent` |
| **目的** | 各マスタカテゴリのアイテムCRUD |

**ルートマッピング（6カテゴリ）:**
| スラグ | カテゴリキー | ラベル |
|---|---|---|
| `service-type` | `serviceType` | 予約区分マスタ |
| `medicine` | `medicine` | 薬剤マスタ |
| `hospitalization` | `hospitalization` | 入院マスタ |
| `cage` | `cage` | ケージマスタ |
| `staff` | `staff` | スタッフマスタ |
| `insurance` | `insurance` | 保険マスタ |

**ツリー表示（`showParentItem: true` のカテゴリ）:**
- ドラッグ中にテーブル上部に「トップレベルに移動」ドロップゾーンが出現
- 各行にドラッグハンドル（`GripVertical` アイコン）付き
- トップレベル項目＝「カテゴリ」として機能、Chevronで展開/折りたたみ
- 操作列: 「+」ボタン（子項目インライン追加）+ 編集ボタン
- D&D並び順変更: `sortOrder`フィールドで永続化、`bulkUpdate`で兄弟全体を一括更新
- カスタムドラッグプレビュー: `setDragImage`でNotionライクなピル型ゴーストを表示
- ホバー自動展開: 折りたたまれた親ノードの中央ゾーンに600msホバーで自動展開
- キーボードアクセシビリティ: `Alt+ArrowUp/Down` で並び替え、`Alt+ArrowLeft/Right` で階層変更

**カテゴリ固有セクション（`[C] MasterItemFormSections` → `sections/`）:**

| カテゴリ | 固有フィールド |
|---|---|
| `examination` | 検査項目リスト（動的: 項目名・単位・正常値） |
| `vaccine` | 対象種別（犬/猫/共通）、標準接種間隔 |
| `medicine` | 剤形（錠剤/液剤/注射/外用薬/粉末）、単位 |
| `staff` | 職種（job_title連動）、資格番号、所属医院（複数可）、メール、パスワード、ユーザー種別 |
| `cage` | ケージタイプ（ICU/犬舎/猫舎/共用）、サイズ（小/中/大） |
| `insurance` | 補償割合（50%/70%/80%/100%）、請求先電話番号 |
| `trimming_course` | 対象サイズ（小型犬/中型犬/大型犬/猫）、所要時間 |
| `trimming_option` | 追加所要時間、併用可否（併用可/単独のみ） |
| `hospitalization` | 対象体格（小/中/大）、料金単位（1日/1泊） |
| `consultation` | 適用区分（常時/初診/再診/時間外/緊急）、標準診察時間 |
| `procedure` | 所要時間、麻酔要否（不要/局所/鎮静/全身） |
| `checkup` | 推奨受診間隔、対象年齢（全/幼齢/成年/シニア） |
| `serviceType` | 表示カラー（カラーピッカー＋プレビュー） |
| `diagnosis_name` | 診断カテゴリ（diagnosis_category連動） |
| `diagnosis_category` | 固有セクションなし |

**スタッフ固有ダイアログ（`[S] StaffImpactDialog`）:**
- ステータス変更・名称変更・削除時に影響範囲を確認

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `Settings` | `[R]` | メインページ（リスト/編集切替） |
| `MasterItemFormSections` | `[C]` | カテゴリ固有フォームセクション（ディスパッチャー） |
| `NotionPropertyRow` | `[S]` | Notion風プロパティ行 |
| `StaffImpactDialog` | `[S][M]` | スタッフ変更影響確認ダイアログ |
| `ConfirmDialog` | `[S][M]` | 削除確認 |
| `MasterLink` | `[S]` | マスタ設定リンク |
| `useMasterItemEditor` | `[H]` | CRUD操作フック |
| `useMasterItems` | `[H]` | マスタデータ取得 |

**データ型:** `MasterItem`, `MasterFormData`, `MasterCategory`, `CreateMasterItemDTO`, `UpdateMasterItemDTO`, `CategoryConfig` + 各カテゴリ固有型

---

## 12. 在庫管理

### 12.1 在庫一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/inventory` |
| **コンポーネント** | `[R] InventoryList` |
| **目的** | 在庫品目の一覧表示・検索・フィルタリング |

**フィルタ:**
- キーワード検索（品名、保管場所、仕入先）
- カテゴリフィルタ（Select）: 全カテゴリー / 医薬品 / 消耗品 / フード / その他
- 状態フィルタ（Select）: 全ての状態 / 在庫あり / 残りわずか / 在庫切れ

**テーブル列（SortableHeader対応）:**
| 列 | 表示内容 |
|---|---|
| 品名 | `item.name` |
| カテゴリー | `getInventoryCategoryLabel(item.category)` |
| 在庫数 | `item.quantity`（発注点以下は赤文字） |
| 単位 | `item.unit` |
| 発注点 | `item.minStockLevel` |
| 状態 | `StatusBadge` |
| 保管場所 | `item.location` |
| 期限 | `item.expiryDate` |
| 仕入先 | `item.supplier` |
| 操作 | `RowActionDropdown`（編集 / 削除） |

**データ型:** `InventoryItem`, `InventoryCategory`, `InventoryStatus`, `DataTableColumn`

### 12.2 在庫登録/編集

| 項目 | 内容 |
|------|------|
| **ルート** | `/inventory/new`（新規）/ `/inventory/:id`（編集） |
| **コンポーネント** | `[R] InventoryForm` |
| **目的** | 在庫品目の追加・編集 |

**フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 品名 | `Input` | 必須 |
| カテゴリー | `Select`（`INVENTORY_CATEGORY_VALUES`） | |
| 単位 | `Input` | 必須 |
| 現在在庫数 | `Input`（type=number, min=0） | |
| 発注点 | `Input`（type=number, min=0） | |
| 保管場所 | `Input` | |
| 使用期限 | `NotionDatePicker` | |
| 仕入先 | `Input` | |
| ステータス（自動判定） | 読み取り専用 | 編集時のみ表示 |

**データ型:** `InventoryItem`, `InventoryCategory`, `InventoryStatus`
**カテゴリ:** 医薬品 / 消耗品 / フード / その他
**ステータス:** 在庫あり / 残りわずか / 在庫切れ

---

## 13. シフト管理

### 13.1 シフト管理カレンダー

| 項目 | 内容 |
|------|------|
| **ルート** | `/shifts` |
| **コンポーネント** | `[R] ShiftCalendar` |
| **目的** | スタッフの勤務シフトを週間・月間カレンダーで管理する |

**画面構成:**
- ツールバー: ビュー切替（週/月）、ロールフィルタ（全員/医師/スタッフ/トリマー）、ナビゲーション、今日ボタン
- **週表示** (`[C] ShiftWeekView`): スタッフ×曜日の7列グリッド、シフトタイプ別色分けバッジ
  - セルクリックで `[C] ShiftEditPopover`（シフトタイプ選択・時間入力・メモ）
  - 行末に週計労働時間表示（40時間超過で警告色）
- **月表示** (`[C] ShiftMonthView`): カレンダーグリッドに日ごとの勤務人数サマリー
- 凡例バー（`[C] ShiftLegend`）

**コンポーネント構成:**
| コンポーネント | 種別 | 役割 |
|---|---|---|
| `ShiftCalendar` | `[R]` | メインページ |
| `ShiftWeekView` | `[C]` | 週間グリッドビュー |
| `ShiftMonthView` | `[C]` | 月間カレンダービュー |
| `ShiftCell` | `[C]` | シフトセル表示 |
| `ShiftEditPopover` | `[C]` | シフト編集Popover |
| `ShiftLegend` | `[C]` | 凡例コンポーネント |
| `useShiftManagement` | `[H]` | 状態管理（ビュー切替・ナビゲーション・CRUD・労働時間計算） |

**データ型:** `ShiftEntry`, `ShiftType`, `ShiftView`, `ShiftStaffInfo`, `DayShiftSummary`
**シフトタイプ:** 通常勤務(full) / 午前のみ(morning) / 午後のみ(afternoon) / 休み(off) / 有給(paid_leave)

**アクセシビリティ:**
- `ShiftEditPopover`: `useFocusTrap` 統合（Escape/Tab循環/フォーカス復帰）、`role="dialog"` + `aria-modal="true"`
- ツールバー: `role="group"` + `aria-label`、トグルボタンに `aria-pressed`

---

## 14. 共通ペット選択画面

以下の4画面は `[S] PetSearchForm` + `[S] PetSearchResultsTable` 共通コンポーネントを使用:

| ルート | コンポーネント | 選択後の遷移先 |
|---|---|---|
| `/medical-records/select-pet` | `MedicalRecordPetSelection` | `/medical-records/new?petId=xxx` |
| `/hospitalization/select-pet` | `HospitalizationPetSelection` | `/hospitalization/new?petId=xxx` |
| `/trimming/select-pet` | `TrimmingPetSelection` | `/trimming/new?petId=xxx` |
| `/accounting/select-pet` | `AccountingPetSelection` | `/accounting/new?petId=xxx` |

**共通画面構成:**
- `[S] PageLayout` でラップ
- ペット検索フォーム（`[S] PetSearchForm`）: 飼主ID、飼主名、電話、ペット名、種
- 検索結果テーブル（`[S] PetSearchResultsTable`）
- 選択ボタンクリックでクエリパラメータ付き遷移

**データ型:** `PetSearchParams`, `Pet`

---

## 15. ログイン画面

### 15.1 ログイン

| 項目 | 内容 |
|------|------|
| **ルート** | `/login` |
| **コンポーネント** | `[R] Login` |
| **目的** | スタッフ認証・セッション開始 |
| **アクセス制御** | 公開ルート（認証済みの場合は `/` にリダイレクト） |

**フォーム要素:**
- メールアドレス入力（`type="email"`、バリデーション: 必須・メール形式）
- パスワード入力（`type="password"`、バリデーション: 必須・最小8文字）
- ログインボタン
- パスワード表示トグル（`Eye` / `EyeOff` アイコン）
- エラーメッセージ表示（認証失敗時）
- デモアカウントパネル（開発用）

**デモアカウント一覧:**
| メール | 表示名 | ユーザー種別 | 職種 |
|---|---|---|---|
| `admin@example.com` | 田中 太郎 | `clinic_admin` | 医師 |
| `vet@example.com` | 山田 花子 | `staff` | 医師 |
| `nurse@example.com` | 佐藤 美咲 | `staff` | 看護師 |
| `reception@example.com` | 鈴木 一郎 | `staff` | 受付 |
| `trimmer@example.com` | 高橋 さくら | `staff` | トリマー |
| `system@example.com` | 本部 管理者 | `system_admin` | — |

**関連コンポーネント:**
| コンポーネント | パス | 役割 |
|---|---|---|
| `Login` | `/features/auth/routes/Login.tsx` | 認証状態チェック + リダイレクト |
| `LoginForm` | `/features/auth/components/LoginForm.tsx` | フォームUI + バリデーション |
| `ClinicSwitcher` | `/features/auth/components/ClinicSwitcher.tsx` | サイドバーのクリニック切替 |
| `ProtectedRoute` | `/features/auth/components/ProtectedRoute.tsx` | ルートレベル認証ガード |
| `PermissionGate` | `/features/auth/components/PermissionGate.tsx` | コンポーネントレベル権限ゲート |

---

## 画面遷移図（概要）

```
Dashboard(/) ──→ カルテ / 予約 / 会計（モーダル内リンク）

Owners(/owners) ──→ OwnerForm(/owners/new, /owners/:id)
  └── PetEditModal
  └── ペット行ドロップダウン ──→ カルテ/予約/トリミング/入院/会計

MedicalRecords(/medical-records) ──→ PetSelection ──→ MedicalRecordForm
  └── タブ内から検査/予防接種レコード作成

Hospitalization(/hospitalization) ──→ PetSelection ──→ HospitalizationForm
  └── HospitalizationDetail（ケアプラン/デイリーレコード）

Trimming(/trimming) ──→ PetSelection ──→ TrimmingForm

Examinations(/examinations) ──→ MedicalRecordForm（検査タブ）

Accounting(/accounting) ──→ PetSelection ──→ AccountingDetail
  └── カルテリンク ──→ MedicalRecordForm

Vaccinations(/vaccinations) ──→ MedicalRecordForm（予防接種タブ）

Inventory(/inventory) ──→ InventoryForm(/inventory/new, /inventory/:id)

Shifts(/shifts) ──→ 週間/月間ビュー切替・シフト編集Popover

Settings(/settings) ──→ ClinicSettings / Settings({category})
```

---

## 全ルート一覧（42ルート + Fallback）

| # | ルート | 画面名 | ページコンポーネント |
|---|--------|--------|---------------------|
| 1 | `/` | ダッシュボード | `Dashboard` |
| 2 | `/owners` | 飼主一覧 | `OwnersList` |
| 3 | `/owners/new` | 飼主新規登録 | `OwnerForm` |
| 4 | `/owners/:id` | 飼主編集 | `OwnerForm` |
| 5 | `/reservations` | 予約管理 | `ReservationManagement` |
| 6 | `/medical-records` | カルテ一覧 | `MedicalRecords` |
| 7 | `/medical-records/select-pet` | カルテ用ペット選択 | `MedicalRecordPetSelection` |
| 8 | `/medical-records/new` | カルテ新規作成 | `MedicalRecordForm` |
| 9 | `/medical-records/:id` | カルテ編集 | `MedicalRecordForm` |
| 10 | `/hospitalization` | 入院一覧 | `HospitalizationList` |
| 11 | `/hospitalization/select-pet` | 入院用ペット選択 | `HospitalizationPetSelection` |
| 12 | `/hospitalization/new` | 入院新規登録 | `HospitalizationForm` |
| 13 | `/hospitalization/:id` | 入院詳細 | `HospitalizationDetail` |
| 14 | `/hospitalization/:id/edit` | 入院編集 | `HospitalizationForm` |
| 15 | `/trimming` | トリミング一覧 | `TrimmingList` |
| 16 | `/trimming/select-pet` | トリミング用ペット選択 | `TrimmingPetSelection` |
| 17 | `/trimming/new` | トリミング新規登録 | `TrimmingForm` |
| 18 | `/trimming/:id` | トリミング編集 | `TrimmingForm` |
| 19 | `/examinations` | 検査管理 | `Examinations` |
| 20 | `/accounting` | 会計一覧 | `Accounting` |
| 21 | `/accounting/select-pet` | 会計用ペット選択 | `AccountingPetSelection` |
| 22 | `/accounting/new` | 会計新規作成 | `AccountingDetail` |
| 23 | `/accounting/:id` | 会計精算 | `AccountingDetail` |
| 24 | `/vaccinations` | 予防接種一覧 | `VaccinationList` |
| 25 | `/checkups` | 定期健診一覧 | `CheckupList` |
| 26 | `/inventory` | 在庫一覧 | `InventoryList` |
| 27 | `/inventory/new` | 在庫新規登録 | `InventoryForm` |
| 28 | `/inventory/:id` | 在庫編集 | `InventoryForm` |
| 29 | `/shifts` | シフト管理 | `ShiftCalendar` |
| 30 | `/settings` | マスタ設定トップ | `MasterSettingsIndex` |
| 31 | `/settings/clinic` | 病院情報設定 | `ClinicSettings` |
| 32 | `/settings/treatment-items` | 診療項目マスタ (5カテゴリ統合) | `TreatmentItemsSettings` |
| 33 | `/settings/diagnosis` | 診断マスタ (2カテゴリ統合) | `DiagnosisSettings` |
| 34 | `/settings/trimming` | トリミングマスタ (2カテゴリ統合) | `TrimmingSettings` |
| 35 | `/settings/service-type` | 予約区分マスタ | `Settings` |
| 36 | `/settings/medicine` | 薬剤マスタ | `Settings` |
| 37 | `/settings/staff` | スタッフマスタ | `Settings` |
| 38 | `/settings/insurance` | 保険マスタ | `Settings` |
| 39 | `/settings/hospitalization` | 入院マスタ | `Settings` |
| 40 | `/settings/cage` | ケージマスタ | `Settings` |
| 41 | `/dev/tests` | フォーマットテスト (開発用) | `FormatTestRunner` |
| 42 | `/login` | ログイン | `Login` |
| — | `*` | 404ページ | インライン |

> **注**: マスタ設定ルートは3統合ページ（`TreatmentItemsSettings`/`DiagnosisSettings`/`TrimmingSettings`）+ 6個別ページ（`Settings` コンポーネント, category prop切替）= 9ルートで15カテゴリをカバー。

---

## 関連ドキュメント

| ドキュメント | 説明 |
|---|---|
| [DESIGN_SYSTEM.md](./DESIGN_SYSTEM.md) | デザインシステム・コンポーネントスタイリング |
| [SPECIFICATION.md](./SPECIFICATION.md) | システム仕様・アーキテクチャ |
| [ERD.md](./ERD.md) | データベース設計・ER図 |
| [DB_DEFINITION.md](./DB_DEFINITION.md) | DB定義書・DDL |
| [screens/20-master-settings.md](./screens/20-master-settings.md) | マスタ設定詳細仕様 |
