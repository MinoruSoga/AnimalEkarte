# 動物病院管理システム 画面仕様書

本ドキュメントは、全画面（ルート）ごとの仕様を定義します。
各画面のルートパス、目的、構成コンポーネント、データフロー、ユーザー操作を網羅しています。

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
| 診療区分 | `serviceType` バッジ（キーワード自動アイコン: トリミング→Scissors / ワクチン→Syringe / 手術→Activity / 診療→Stethoscope） |
| 担当医 | `doctor` バッジ（指名時は「指」ラベル＋オレンジ背景、無効スタッフ時は赤背景＋AlertCircle） |

**DashboardDetailModal 表示項目:**
| セクション | 項目 |
|---|---|
| ヘッダー | 初診/再診アイコン（初/再）、診療区分名、予約ID、ステースバッジ（カラム別カラー） |
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
| `useDashboardKanban` | `[H]` | カンバン状態管理 |

**データ型:** `Appointment`, `ColumnData`, `DashboardColumnTitle`

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
| 予約種別カラー凡例 | 8種別のカラードット＋ラベル（診療=blue / 検診=green / 検査=green / 手術=red / トリミング=orange / ワクチン=purple / 入院=cyan / ホテル=cyan） |
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
  - 選択済みペットは `SelectedPetChip`（PawPrint アイコン付き、種バッジ、飼主No表示、×ボタン）で表示
  - 編集モード時はStep 1をスキップ
- **Step 2: 予約情報入力**
  - `ReservationFormFields`（上記フォーム項目一式）
  - 新規時: 日付セルクリック日を初期値、デフォルト時間10:00-11:00
  - フッターに「保存」ボタン

**ReservationDetailModal 表示項目:**
| セクション | 項目 |
|---|---|
| アクセントバー | visitType に応じた色帯（初診=赤 / 再診=青） |
| ヘッダー | 初診/再診バッジ（丸ドット付き）、予約種別名 |
| ステータスセレクター | 現在ステータスの色帯＋6段階ドロップダウン（予約確定/受付済/診療中/会計待ち/完了/キャンセル、各色ドット付き） |
| 日時カード | 日付（yyyy年 M月 d日 (E)）、時間帯（H:mm – H:mm）、Calendar/Clock アイコン |
| 患者情報 | ペット名（太字）、飼い主名、カルテNo（petId、等幅フォント） |
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

**データ型:** `ReservationAppointment`, `ReservationFormData`, `CalendarView`, `ReservationStatus`, `ReservationType`

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
- データテーブル（`[S] DataTable`）: 飼主No、飼主名、ペット番号、ペット名、生死、種、生年月日、体重、環境、前回来院、操作
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

**画面構成:**
- ヘッダー: タイトル + 保存ボタン
- 飼主情報セクション（2カラムグリッド）:
  - 飼主ID、飼主名 (必須)、飼主名カナ (必須)、会社名、郵便番号、住所1/2、自宅住所1/2、生年月日、電話番号 (必須)、会社電話、メール、備考、危険フラグ (Switch)、割引率、会員種別 (Select)

**OwnerForm フォーム項目（4カラムグリッド）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 飼主No | `Input` | |
| 郵便番号 | `Input` | placeholder: 123-4567 |
| 会社名 | `Input` | |
| 会員区分 | `Button` グループ | `MEMBERSHIP_TYPE_VALUES`（非会員/会員/退亡者/他診/準） |
| 飼主名 | `Input` | 必須 |
| 住所1 | `Input` | |
| 郵便番号(自宅) | `Input` | |
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

- ペット一覧テーブル:
  - ペット追加ボタン
  - 各ペット行にドロップダウン（カルテ作成、予約、トリミング、入院、会計、編集、削除）
- ペット編集モーダル（`[C] PetEditModal`）
- フォーム離脱保護（`[S] NavigationBlocker`）

**PetEditModal フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| ペット番号 | `Input` | |
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

- 3カラムグリッド（md:2, lg:3）
- 必須フィールド: ペット名、種別、性別、生年月日
- バリデーション: 必須未入力時にトースト通知

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
- `[S] PetSelection` 共通コンポーネントを使用
  - ペット検索フォーム（飼主ID、飼主名、電話、ペット名、種）
  - 検索結果テーブル
  - 選択ボタンで `/medical-records/new?petId=xxx` へ遷移

### 4.3 カルテ入力/編集

| 項目 | 内容 |
|------|------|
| **ルート** | `/medical-records/new`（新規）/ `/medical-records/:id`（編集） |
| **コンポーネント** | `[R] MedicalRecordForm` |
| **目的** | 診療記録の全項目を8タブ構成で入力・編集する |

**画面構成:**
- スティッキーヘッダー:
  - 患者情報カード（`[S] PatientInfoCard`）: ペット名、種、飼主名、担当医、診療区分、バイタル入力ボタン
  - タブバー（8タブ）
- タブコンテンツ（遅延マウント: 一度表示したタブは保持）

**タブ詳細:**

| # | タブ名 | コンポーネント | 説明 |
|---|--------|---------------|------|
| 1 | **問診** | `MedicalRecordInterview` | 主訴（Markdown入力）、問診履歴表示 |
| 2 | **診察/治療プラン** | `MedicalRecordDiagnosisPlan` | バイタル入力・履歴グラフ、診断登録（カテゴリ+診断名）、治療プラン（TreatmentTable） |
| 3 | **治療** | `MedicalRecordTreatment` | 処置完了記録テーブル（TreatmentTable）、検索ダイアログでマスタ連携 |
| 4 | **予防接種** | `MedicalRecordVaccination` | ワクチンフォーム + 接種履歴一覧 |
| 5 | **検査** | `MedicalRecordExamination` | 検査オーダーフォーム + 結果履歴 |
| 6 | **画像** | `MedicalRecordImage` | 画像アップロード + フィルタ付きギャラリー |
| 7 | **見積書** | `MedicalRecordEstimate` | 診療内容に基づく概算見積 |
| 8 | **会計(医師確認)** | `MedicalRecordBillCheck` | 算定チェック・確認 |

**タブ別フォーム項目詳細:**

**Tab 1: 問診（`MedicalRecordInterview`）**
- 3カラムレイアウト（lg:12グリッド = 3+4+5）
- **左カラム**: 主訴入力（`InterviewChiefComplaint`）
  - Markdown テキストエリア（テンプレート見出し: どんな症状 / どこが / いつから / その他・備考 / フリースペース）
  - テンプレート挿入ボタン（定期検診 / ワクチン / 下痢・嘔吐 / 皮膚）
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
- **治療プランテーブル**: `TreatmentTable`（showStatus=true、ステータス列あり）
- **治療済みテーブル**: `TreatmentTable`（showStatus=false）
- **集計**: `TreatmentDetailedSummary`
- TreatmentTable 列: [ステータス]、治療内容、メモ、保険、単価、数量、割引(%)、値引(￥)、小計、操作

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
- **明細テーブル**: `TreatmentTable`（showStatus=false）
- **集計**: `TreatmentDetailedSummary`
- **コメント / 備考**: 2カラム `Textarea`
- **アクション**: 「PDF出力」ボタン

**Tab 8: 会計(医師確認)（`MedicalRecordBillCheck`）**
- **明細テーブル**: `TreatmentTable`（治療タブの completedItems を自動同期）
- **集計**: `TreatmentDetailedSummary`
- **固定フローティングアクション**: 「チェック完了」ボタン + 「会計へ進む」ボタン（Receipt アイコン付き、items 空時は disabled）
- **会計遷移**: カルテの明細を `AccountingItem[]` に自動変換し、カテゴリを自動推定（検査/処方/手術/処置/フード等のキーワードマッチ）して会計画面へ state 経由で渡す

**VitalInputDialog フォーム項目（カルテ共通）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 体温 | `Input`（number, step=0.1） | Thermometer アイコン、単位: ℃ |
| 心拍数 | `Input`（number） | Heart アイコン、単位: /min |
| 呼吸数 | `Input`（number） | Wind アイコン、単位: /min |
| 体重 | `Input`（number, step=0.01） | Weight アイコン、単位: kg |
| メモ | `Textarea` | StickyNote アイコン |
- 入力/履歴のタブ切替、履歴ではトレンドアイコン（↑↓→）表示

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `MedicalRecordForm` | `[R]` | メインページ |
| `PatientInfoCard` | `[S]` | 患者情報カード |
| `MasterSelectModal` | `[S][M]` | 診療区分・担当医選択 |
| `VitalInputDialog` | `[C][M]` | バイタル入力ダイアログ |
| `TreatmentTable` | `[C]` | 治療項目テーブル |
| `TreatmentSearchDialog` | `[C][M]` | マスタ検索ダイアログ |
| `InterviewChiefComplaint` | `[C]` | 主訴入力（Markdown） |
| `InterviewHistory` | `[C]` | 問診履歴 |
| `DiagnosisHeader` | `[C]` | 診断ヘッダー |
| `ExaminationForm` | `[C]` | 検査フォーム |
| `ExaminationHistory` | `[C]` | 検査結果履歴 |
| `VaccinationForm` | `[C]` | ワクチンフォーム |
| `VaccinationHistory` | `[C]` | 接種履歴 |
| `EstimateForm` | `[C]` | 見積フォーム |
| `NavigationBlocker` | `[S]` | フォーム離脱保護 |
| `useMedicalRecordForm` | `[H]` | メインフォームフック |
| `useMedicalRecordInit` | `[H]` | 初期化ロジック |

**データ型:** `MedicalRecord`, `TreatmentItem`, `VitalEntry`, `DiagnosisFormData`, `ExaminationFormData`, `ExaminationResultGroup`, `VaccinationFormData`, `VaccinationHistoryItem`, `InterviewHistoryItem`

**ユーザー操作:**
- タブ切替で各セクションの入力
- 患者情報カードから診療区分・担当医をマスタ選択
- バイタル入力ダイアログでバイタル記録
- 治療テーブルでマスタ検索→項目追加
- 保存ボタンで確定（バリデーション実行）
- 削除ボタンでカルテ削除（確認ダイアログ）
- 未保存離脱時の保護ダイアログ

---

## 5. 入院管理

### 5.1 入院一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/hospitalization` |
| **コンポーネント** | `[R] HospitalizationList` |
| **目��** | 入院・ホテル患者の一覧管理、ボード/リスト表示切替 |

**画面構成:**
- ヘッダー: タイトル + 新規登録ボタン
- フィルタ: ステータス（全て/入院中/退院済/予約）タブ、ボード/リスト表示切替
- 検索バー
- **ボードビュー**（`[C] HospitalizationBoard`）: ケージごとのカード表示
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

**画面構成:** `[S] PetSelection` 共通コンポーネント使用

### 5.3 入院登録/編集

| 項目 | 内容 |
|------|------|
| **ルート** | `/hospitalization/new`（新規）/ `/hospitalization/:id/edit`（編集） |
| **コンポーネント** | `[R] HospitalizationForm` |
| **目的** | 入院情報の入力・治療プラン管理 |

**画面構成:**
- 患者情報カード（`[S] PatientInfoCard`）
- 基本情報セクション（`[C] HospitalizationBasicInfo`）: 入院区分、ケージ選択、担当医、日付
- メモ・指示セクション（`[C] HospitalizationNoteCard`）: オーナー要望、スタッフメモ
- 処置テーブル（`[C] HospitalizationTreatmentTable`）: 治療項目の管理
- コスト集計（`[C] HospitalizationCostSummary`）
- 担当医選択モーダル（`[S] MasterSelectModal`）
- フォーム離脱保護

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

**使用フック:**
| フック | 説明 |
|---|---|
| `useHospitalizationForm` | フォーム状態管理（186行） |
| `useTreatmentPlans` | 治療プラン管理（112行） |

**データ型:** `HospitalizationFormData`, `TreatmentPlan`, `CreateHospitalizationDTO`, `UpdateHospitalizationDTO`

### 5.4 入院詳細

| 項目 | 内容 |
|------|------|
| **ルート** | `/hospitalization/:id` |
| **コンポーネント** | `[R] HospitalizationDetail` |
| **目的** | 入院患者のケアプラン管理、デイリーログ記録 |

**画面構成（レスポンシブ: デスクトップ/モバイル分離）:**

**デスクトップレイアウト（`HospitalizationDesktopLayout`）:**
- 左カラム: 患者ヘッダー + アクションバー + ケアプランセクション
- 右カラム: デイリーレコードセクション

**モバイルレイアウト（`HospitalizationMobileLayout`）:**
- シングルカラム: 患者ヘッダー → ケアプラン → デイリーレコード

**ケアプラン（`[C] CarePlan/`）:**
| コンポーネント | 説明 |
|---|---|
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
| 名称 | `Input` | 例: ロイヤルカナン消化器サポート |
| マスタ連動情報 | Badge 表示（単価、カテゴリ） | マスタ選択時のみ表示 |
| 詳細・指示量 | `Input` | 例: 30g / 1錠 / 左前肢 |
| タイミング | トグルボタン（`PLAN_TIMING_VALUES`: 朝/昼/夜） | 複数選択可 |
| メモ・特記事項 | `Textarea` | 例: ふやかして与える |
| ステータス | `Select`（`CARE_PLAN_STATUS_VALUES`） | |

**VitalDialog フォーム項目（入院デイリーレコード用）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 記録時刻 | `Input`（type=time） | 現在時刻で初期化 |
| 体温 (℃) | `Input`（number, step=0.1） | |
| 体重 (kg) | `Input`（number, step=0.01） | |
| 心拍数 (/min) | `Input`（number） | |
| 呼吸数 (/min) | `Input`（number） | |
| メモ | `Textarea` | |

**LogDialog フォーム項目（入院ケアログ用）:**
- ログ種別に応じてタイトル・説明・プレースホルダーが変化
  - food: 「食事記録」（完食、1/2など）
  - excretion: 「排泄記録」（良便、軟便など）
  - medicine / other: 「活動・メモ」（内容）

| フィールド | 入力部品 | 備考 |
|---|---|---|
| 記録時刻 | `Input`（type=time） | 現在時刻で初期化 |
| 内容・量 | `Input` | 種別依存のプレースホルダー |
| 詳細メモ | `Textarea` | |

**TaskCompleteDialog フォーム項目（タスク完了記録用）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| タスク情報 | 読み取り専用表示 | タスク名 + 詳細（背景カード） |
| 実施時刻 | `Input`（type=time） | 現在時刻で初期化 |
| 実施メモ (任意) | `Textarea` | |

**データ型:** `Hospitalization`, `CarePlanItem`, `DailyRecord`, `VitalRecord`, `CareLogRecord`, `StaffNoteRecord`, `Task`, `TimelineItem`

**ユーザー操作:**
- ケアプランの追加/編集/削除（カテゴリ: 食事、投薬、処置・検査、処置・指示、持ち物・その他）
- デイリーレコード: 日付ナビゲーション
- タスク完了チェック（朝/昼/夜のタイミングごと）
- バイタル記録入力
- ケアログ記録（食事、排泄、投薬、処置、その他）
- スタッフメモ追加
- タイムラインで時系列確認
- 退院処理（確認ダイアログ）
- 編集画面への遷移

---

## 6. トリミング

### 6.1 トリミング一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/trimming` |
| **コンポーネント** | `[R] TrimmingList` |
| **目的** | トリミング予約の検索・一覧管理 |

**画面構成:**
- ヘッダー: タイトル（Scissors アイコン）+ 新規登録ボタン（`[S] PrimaryButton`）
- フィルタ: `[S] SearchFilterBar`（キーワード検索「飼主名、ペット名...」）+ 日付範囲フィルタ
- データテーブル（`[S] DataTable`）
- ページネーション（`[S] Pagination`、20件/ページ）
- 削除確認ダイアログ（`[S] ConfirmDialog`）

**フィルタ項目:**
| 項目 | 入力部品 | 備考 |
|---|---|---|
| 開始日 | `NotionDatePicker` | `lg:w-[160px]` |
| 終了日 | `NotionDatePicker` | `lg:w-[160px]`、「〜」で接続 |
| キーワード | `SearchFilterBar` | 飼主名・ペット名で検索 |
| クリア | `Button`（outline） | 全フィルタリセット |

**テーブル列:**
| 列 | className | 表示内容 |
|---|---|---|
| 診療日 | `w-[120px]` | `record.date`（等幅フォント） |
| 飼主名 | - | `record.ownerName` |
| ペット名 | - | `record.petName` + `record.petNumber`（2行表示） |
| 種 | `w-[80px]` | `record.species` |
| 体重 | `w-[80px]` | `record.weight` |
| スタイル希望 | - | `record.styleRequest`（truncate、max-w-[200px]） |
| 担当 | `w-[100px]` | `record.staff`（無効スタッフ時は赤文字＋AlertTriangle） |
| ステータス | `w-[100px]` | `StatusBadge`（`getTrimmingStatusColor`） |
| 操作 | `w-[100px]`, align:right | `RowActionDropdown`（編集 / 削除） |

**行アクション:**
| アクション | アイコン | 動作 |
|---|---|---|
| 編集 | Edit | `/trimming/{id}` へ遷移 |
| 削除 | Trash2 | `ConfirmDialog` → `deleteRecord`、構造化トースト |

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `TrimmingList` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `SearchFilterBar` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `NotionDatePicker` | `[S]` | 日付ピッカー（×2） |
| `StatusBadge` | `[S]` | ステータスバッジ |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `ConfirmDialog` | `[S][M]` | 削除確認 |
| `Pagination` | `[S]` | ページネーション |
| `useTrimmingRecords` | `[H]` | 検索・フィルタ・削除ロジック |
| `useStaffValidation` | `[H]` | スタッフ有効性チェック |
| `usePagination` | `[H]` | ページネーション（resetKey 連動） |

**データ型:** `TrimmingRecord`, `DataTableColumn`
**ステータス:** 予約 / 進行中 / 完了

### 6.2 トリミング用ペット選択

| 項目 | 内容 |
|------|------|
| **ルート** | `/trimming/select-pet` |
| **コンポーネント** | `[R] TrimmingPetSelection` |
| **画面構成** | `[S] PetSelection` 共通コンポーネント使用 |

### 6.3 トリミング登録/編集

| 項目 | 内容 |
|------|------|
| **ルート** | `/trimming/new`（新規）/ `/trimming/:id`（編集） |
| **コンポーネント** | `[R] TrimmingForm` |
| **目的** | トリミング情報の入力・編集 |

**画面構成:**
- 患者情報カード（`[S] PatientInfoCard`）: 担当スタッフ・診療区分「トリミング」表示
- 3カラムレイアウト（`grid-cols-1 md:2 lg:3 gap-3`）
- フォーム離脱保護（`[S] NavigationBlocker`）
- ヘッダーに削除ボタン（編集時のみ）+ 保存ボタン

**左カラム フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| コース選択 | `MasterSelectTrigger` → `MasterSelectModal`（trimming_course マスタ連動） | `MasterLink` 付き、選択時に `charge` 自動反映 |
| スタイルの希望 | `Textarea` | min-h-[80px] |
| メモ | `Textarea` | min-h-[80px] |
| オプション | `Checkbox` グリッド（2cols） | trimming_option マスタ連動、`MasterLink` 付き、複数選択可、各項目に `+¥{price}` 表示 |
| 希望スタイル画像 | `file input` + プレビュー | ドラッグ&ドロップUI、Upload アイコン、×ボタンで削除、h-[180px] |

**中カラム フォーム項目:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| BW（体重） | `Input` + `radio`（Kg/g） | `BODY_WEIGHT_UNIT_VALUES`、2カラムグリッド |
| BT（体温） | `Input` | |
| USED SHAMPOO | `Input` | placeholder: 使用したシャンプーを入力... |
| USED RIBBON | `Input` | placeholder: 使用したリボンを入力... |
| TREATMENT | `Input` | placeholder: 処置内容を入力... |
| 備考 | `Input` | |
| 完成画像 | `file input` + プレビュー | Upload アイコン、×ボタンで削除、h-[180px] |

**右カラム（トリミング履歴）:**
- タイトル: 「トリミング履歴」
- `[S] HistoryFilterPanel`: 日付範囲、キーワード検索（「コース名・メモで検索...」）、ソート順（昇順/降順）、クリアボタン
- 履歴カード: 診療日、コース名バッジ、作成者/更新者・日時、スタイルの希望、メモ、画像セクション
- 空状態: 「該当するトリミング履歴がありません」

**担当スタッフ選択（PatientInfoCard 経由）:**
- `MasterSelectModal`（staff マスタ連動、active のみ）
- タイトル: 「担当スタッフを選択」

**バリデーション:**
- 担当スタッフ未選択時はトースト警告（`toast.warning`）で保存ブロック

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
| `useUnsavedChanges` | `[H]` | 未保存検知 |
| `useMasterItems` | `[H]` | マスタデータ取得 |

**データ型:** `TrimmingFormData`, `TrimmingParts`, `TrimmingHistoryItem`, `BodyWeightUnit`, `SortOrder`

**ユーザー操作:**
- コース選択（マスタモーダル）
- オプション複数選択（チェックボックス）
- 体重単位切替（Kg/g ラジオ）
- スタイル画像・完成画像のアップロード/削除
- 担当スタッフ選択（PatientInfoCard クリック）
- トリミング履歴の検索・フィルタ・ソート
- 保存（バリデーション→トースト→一覧へ遷移）
- 削除（確認ダイアログ→一覧へ遷移）
- 未保存離脱時の保護ダイアログ

---

## 7. 検査管理

### 7.1 検査一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/examinations` |
| **コンポーネント** | `[R] Examinations` |
| **目的** | 検査オーダー・結果の一覧管理 |

**画面構成:**
- ヘッダー: タイトル（TestTube アイコン）+ 「検査データ取込」ボタン（FileSpreadsheet アイコン、outline）
- 検索バー（`[S] SearchFilterBar`）: 「飼主名、ペット名、検査種別...」
- データテーブル（`[S] DataTable`）
- ページネーション（`[S] Pagination`、20件/ページ）

**テーブル列:**
| 列 | className | 表示内容 |
|---|---|---|
| 日時 | `w-[120px]` | `r.date`（等幅フォント） |
| 飼主名 | - | `r.ownerName` |
| ペット名 | - | `r.petName` |
| 検査種別 | - | `r.testType` |
| 結果概要 | - | `r.resultSummary`（truncate、max-w-[200px]、未入力時「-」） |
| 担当医 | `w-[100px]` | `r.doctor`（無効スタッフ時は赤文字＋AlertTriangle） |
| ステータス | `w-[80px]` | `StatusBadge`（`getExaminationStatusColor`） |
| 操作 | `w-[80px]`, align:right | `RowActionDropdown`（「カルテを開く」のみ） |

**行アクション:**
| アクション | アイコン | 動作 |
|---|---|---|
| カルテを開く | FileText | `/medical-records/{medicalRecordId}` へ遷移（state: `{ activeTab: "検査", from: "/examinations" }`） |

**特記:** 検査の新規作成はカルテ内の検査タブから行う。一覧画面は参照+カルテ遷移のみ。行クリックでもカルテの検査タブへ遷移する。

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `Examinations` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `SearchFilterBar` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `StatusBadge` | `[S]` | ステータスバッジ |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `Pagination` | `[S]` | ページネーション |
| `useExaminationRecords` | `[H]` | 検索・フィルタロジック |
| `useStaffValidation` | `[H]` | スタッフ有効性チェック |
| `usePagination` | `[H]` | ページネーション |

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

**画面構成:**
- ヘッダー: タイトル（CreditCard アイコン）+ 新規会計登録ボタン（`[S] PrimaryButton`）
- 検索バー（`[S] SearchFilterBar`）: 「飼主名、ペット名...」
- データテーブル（`[S] DataTable`）
- ページネーション（`[S] Pagination`、20件/ページ）
- 削除確認ダイアログ（`[S] ConfirmDialog`）

**テーブル列:**
| 列 | className / align | 表示内容 |
|---|---|---|
| 日時 | `w-[140px]` | `r.scheduledDate`（等幅フォント） |
| 飼主名 | - | `r.ownerName` |
| ペット名 | - | `r.petName` + カルテ連携バッジ（`medicalRecordId` 存在時、青背景「カルテ連携」） |
| 請求金額 | align:right | `formatCurrency(calculateTotal(r))`（等幅・太字） |
| 支払方法 | align:center | `getPaymentMethodLabel(r.payment?.method)` |
| ステータス | `w-[100px]` | `StatusBadge`（`getAccountingStatusColor` / `getAccountingStatusLabel`） |
| 操作 | `w-[100px]`, align:right | `RowActionDropdown`（編集 / 削除） |

**行アクション:**
| アクション | アイコン | 動作 |
|---|---|---|
| 編集 | Edit | `/accounting/{id}` へ遷移 |
| 削除 | Trash2 | `ConfirmDialog` → `deleteRecord`、構造化トースト |

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `Accounting` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `SearchFilterBar` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `PrimaryButton` | `[S]` | 新規作成ボタン |
| `StatusBadge` | `[S]` | ステータスバッジ |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `ConfirmDialog` | `[S][M]` | 削除確認 |
| `Pagination` | `[S]` | ページネーション |
| `useAccountingRecords` | `[H]` | 検索・フィルタ・削除ロジック |
| `usePagination` | `[H]` | ページネーション |

**データ型:** `Accounting`, `AccountingStatus`, `PaymentMethod`, `DataTableColumn`
**ステータス:** 未収(`waiting`) / 収済(`completed`) / キャンセル(`cancelled`) / 保留(`pending`)

### 8.2 会計用ペット選択

| 項目 | 内容 |
|------|------|
| **ルート** | `/accounting/select-pet` |
| **コンポーネント** | `[R] AccountingPetSelection` |
| **画面構成** | `[S] PetSelection` 共通コンポーネント使用 |

### 8.3 会計精算（新規/編集）

| 項目 | 内容 |
|------|------|
| **ルート** | `/accounting/new`（新規）/ `/accounting/:id`（編集） |
| **コンポーネント** | `[R] AccountingDetail` |
| **目的** | 診療費の計算、支払い処理、書類発行 |

**画面構成:**
- ヘッダー: タイトル「会計精算」+ 戻るボタン + 精算完了時のみ書類発行ボタン群
- 受付情報バナー: `受付No: {id} | {ownerName}様 - {petName}ちゃん` + カルテ確認リンク（`medicalRecordId` 存在時）
- 2カラムレイアウト（`flex-col lg:flex-row gap-6`）
- 書類プレビューダイアログ（`[C] AccountingDocumentPreview`）
- 会計確定確認ダイアログ（`[S] ConfirmDialog`）
- 印刷エリア（hidden、print:block）

**精算完了時ヘッダーアクション:**
| ボタン | アイコン | 動作 |
|---|---|---|
| 診療明細書 | FileText | `handlePrint("statement")` → プレビューモーダル |
| 領収書発行 | Printer | `handlePrint("receipt")` → プレビューモーダル |

**左カラム: 明細テーブル（`[C] AccountingItemTable`）**

明細追加ダイアログ フォーム項目:
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 区分 | `Select`（`MANUAL_ITEM_CATEGORY_VALUES`: 療法食・フード / 物販・ケア用品 / その他） | |
| 品目名 | `Input` | placeholder: 例: ロイヤルカナン 3kg |
| 単価 (税込) | `Input`（type=number） | placeholder: 0 |

明細テーブル列:
| 列 | className / align | 表示内容 |
|---|---|---|
| 区分 | `w-[100px]` | `Badge`（`getItemCategoryLabel`） |
| 項目名 | - | `item.name` + カルテ連携バッジ（`source === "medical_record"` 時） |
| 単価 | align:right, `w-[100px]` | `¥{unitPrice}` |
| 数量 | align:center, `w-[80px]` | `item.quantity` |
| 保険 | align:center, `w-[80px]` | 適用時: 緑●、非適用時: グレー「-」 |
| 金額 | align:right, `w-[120px]` | `¥{unitPrice × quantity}` |
| 削除 | `w-[50px]` | 手動追加項目のみ Trash2 ボタン |

明細フッター: 税抜小計 / 消費税 / 合計（太字・大文字）

**右カラム: 入金パネル（`[C] AccountingPaymentPanel`）**

ペット保険カード:
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 保険ON/OFF | `Switch` | CardHeader 内トグル |
| 負担割合 | `Select`（`INSURANCE_RATIO_VALUES`: 50%/70%/90%/100%） | 保険ON時のみ表示 |
| 保険負担額 | 読み取り専用 | 緑背景カード、マイナス表示 |

決済情報カード:
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 請求金額 | 読み取り専用 | 中央配置、4xl太字 |
| 支払方法 | `Button` グリッド（3cols） | `PAYMENT_METHOD_VALUES`: 現金/カード/電子マネー、選択時は `bg-[#37352F]` |
| お預かり金額 | `Input`（type=number） | 右寄せ大文字、「円」ラベル付き |
| クイック入力 | `Button` × 3 | 「丁度」「千円単位」「一万単位」 |
| お釣り | 読み取り専用 | `bg-[#F7F6F3]` カード、不足時は赤文字 |
| 会計確定 | `Button`（full-width） | Save アイコン、`changeAmount < 0` or `receivedAmount` 空 or 完了済で disabled |

**書類プレビューダイアログ（`[C] AccountingDocumentPreview`）:**
- 領収書 / 診療明細書の切替
- プレビュー表示（`[C] AccountingDocument`）
- 印刷ボタン（`window.print()`）

**データ初期化（`useAccountingDetail`）:**
- 新規時: `location.state` から `accountingItems`（カルテ連携）または `petId` クエリパラメータから生成
- 編集時: `findAccountingById` でロード、`payment` 存在時は保険・支払い情報を復元

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `AccountingDetail` | `[R]` | メインページ |
| `AccountingItemTable` | `[C]` | 明細テーブル |
| `AccountingPaymentPanel` | `[C]` | 入金パネル |
| `AccountingDocumentPreview` | `[C][M]` | 書類プレビューダイアログ |
| `AccountingDocument` | `[C]` | 印刷用書類本体 |
| `PageLayout` | `[S]` | ページレイアウト |
| `ConfirmDialog` | `[S][M]` | 会計確定確認 |
| `useAccountingDetail` | `[H]` | 会計CRUD・計算・書類ロジック |

**データ型:** `Accounting`, `AccountingItem`, `PaymentInfo`, `AccountingCalculation`, `DocumentType`, `ItemCategory`, `ManualItemCategory`, `PaymentMethod`, `InsuranceRatio`, `ItemSource`, `TaxRate`

**ユーザー操作:**
- 明細項目の手動追加（ダイアログ: 区分・品目名・単価）/削除
- 保険適用切替（Switch）、負担割合選択
- 支払方法選択（現金/カード/電子マネー）
- 預り金入力 + クイック入力ボタン → お釣り自動計算
- 会計確定（確認ダイアログ → ステータス completed へ遷移）
- 精算完了後: 領収書/診療明細書のプレビュー → 印刷
- カルテへのリンク遷移（受付バナー内）

---

## 9. 予防接種管理

### 9.1 予防接種一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/vaccinations` |
| **コンポーネント** | `[R] VaccinationList` |
| **目的** | 予防接種記録の一覧管理 |

**画面構成:**
- ヘッダー: タイトル（Syringe アイコン）+ 「データ取込」ボタン（FileSpreadsheet アイコン、outline）
- 検索バー（`[S] SearchFilterBar`）: 「飼主名、ペット名、予防接種名...」
- データテーブル（`[S] DataTable`）
- ページネーション（`[S] Pagination`、20件/ページ）

**テーブル列:**
| 列 | className / align | 表示内容 |
|---|---|---|
| 実施日 | `w-[120px]` | `r.date`（等幅フォント） |
| 飼主名 | - | `r.ownerName` |
| ペット名 | - | `r.petName` |
| 予防接種名 | - | `r.vaccineName`（太字） |
| 担当医 | `w-[100px]` | `r.doctor`（無効スタッフ時は赤文字＋AlertTriangle、未設定時「-」） |
| 次回予定 | `w-[140px]` | `r.nextDate`（等幅フォント） |
| 操作 | `w-[100px]`, align:right | `RowActionDropdown`（「カルテを開く」のみ） |

**行アクション:**
| アクション | アイコン | 動作 |
|---|---|---|
| カルテを開く | FileText | `/medical-records/{medicalRecordId}` へ遷移（state: `{ activeTab: "予防接種", from: "/vaccinations" }`） |

**特記:** 予防接種の新規登録はカルテ内の予防接種タブから行う。一覧画面は参照+カルテ遷移のみ。行クリックでもカルテの予防接種タブへ遷移する。

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `VaccinationList` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `SearchFilterBar` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `Pagination` | `[S]` | ページネーション |
| `useVaccinations` | `[H]` | 検索・フィルタロジック |
| `useStaffValidation` | `[H]` | スタッフ有効性チェック |
| `usePagination` | `[H]` | ページネーション |

**データ型:** `VaccinationRecord`, `DataTableColumn`

---

## 10. 設定・マスタ管理

### 10.1 マスタ設定トップ

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings` |
| **コンポーネント** | `[R] MasterSettingsIndex` |
| **目的** | マスタカテゴリ一覧をカード形式で表示 |

**画面構成:**
- ヘッダー: タイトル「マスタ設定」（Settings アイコン）
- セクション分類（5セクション）
- カードグリッド（`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3`）
- 最終セクションに `pb-16` 付き

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
| アイコン | `CATEGORY_CONFIG[key].IconComponent`（p-2 rounded-lg bg-[#F7F6F3]、ホバー時 bg-[#37352F] text-white） |
| カテゴリ名 | `cfg.label`（truncate） |
| 説明 | `cfg.description`（line-clamp-2） |
| 件数 | `{count}件登録済`（マスタカテゴリのみ、clinic は undefined） |
| 矢印 | ChevronRight（ホバー時 opacity 変化） |

**特記:**
- `CATEGORY_CONFIG` から全カードを自動導出
- コンパイル時に全カテゴリの網羅性を型チェック（`ExhaustiveKeyMap`）
- `useMasterItems()` で全マスタアイテムを取得し、カテゴリ別にカウント

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `MasterSettingsIndex` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `CategoryCard` | ローカル | カテゴリカード |
| `useMasterItems` | `[H]` | マスタデータ取得 |

**データ型:** `MasterCategory`, `MasterCardKey`, `MasterCategoryCard`, `MasterSection`, `CategoryConfig`

### 10.2 病院情報設定

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/clinic` |
| **コンポーネント** | `[R] ClinicSettings` |
| **目的** | 病院の基本情報を管理する |

**画面構成:**
- ヘッダー: タイトル「病院情報設定」（Building2 アイコン）、戻るボタン → `/settings`
- フォーム（`react-hook-form@7.55.0`）
- ローディング時: スケルトン表示（6行の `animate-pulse`）
- カードコンテナ: `bg-white p-6 rounded-lg shadow-sm border`

**フォーム項目:**
| フィールド | 入力部品 | グリッド | 備考 |
|---|---|---|---|
| 病院名 | `Input`（register: `name`） | 2cols-左 | 必須（`*`マーク）、placeholder: 例: わんにゃん動物病院 |
| 支店名 | `Input`（register: `branchName`） | 2cols-右 | placeholder: 例: 八王子院 |
| 郵便番号 | `Input`（register: `postalCode`） | 3cols-左 | placeholder: 例: 100-0001 |
| 住所 | `Input`（register: `address`） | 3cols-中右（col-span-2） | placeholder: 例: 東京都千代田区千代田1-1 |
| 電話番号 | `Input`（register: `phoneNumber`） | 2cols-左 | placeholder: 例: 03-1234-5678 |
| FAX番号 | `Input`（register: `faxNumber`） | 2cols-右 | placeholder: 例: 03-1234-5679 |
| 登録番号 | `Input`（register: `registrationNumber`） | full | placeholder: 例: 東京都獣医師会 第12345号、ヘルプテキスト付き |
| 院長名 | `Input`（register: `directorName`） | full | placeholder: 例: 山田 太郎 |
| メールアドレス | `Input`（type=email, register: `email`） | 2cols-左 | placeholder: info@example.com |
| WebサイトURL | `Input`（register: `website`） | 2cols-右 | placeholder: https://example.com |

**アクション（`pt-4 border-t`）:**
| ボタン | 動作 | 備考 |
|---|---|---|
| キャンセル | `/settings` へ遷移 | outline |
| 設定を保存 | `handleSubmit(onSubmit)` → `updateClinicInfo` + `reset` | Save アイコン、`isDirty` false 時 disabled |

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `ClinicSettings` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `PrimaryButton` | `[S]` | 保存ボタン |
| `useClinicInfo` | `[H]` | 病院情報CRUD |

**データ型:** `ClinicInfo`（name, branchName, postalCode, address, phoneNumber, faxNumber?, registrationNumber?, directorName?, email?, website?, logoUrl?）

### 10.3 マスタカテゴリ設定

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/{category-slug}`（14パターン） |
| **コンポーネント** | `[R] Settings` |
| **目的** | 各マスタカテゴリのアイテムCRUD |

**ルートマッピング（14カテゴリ）:**
| スラグ | カテゴリキー | ラベル | アイコン | showPrice |
|---|---|---|---|---|
| `service-type` | `serviceType` | 予約区分マスタ | Activity | false |
| `consultation` | `consultation` | 診察マスタ | Stethoscope | true |
| `examination` | `examination` | 検査マスタ | TestTube | true |
| `procedure` | `procedure` | 処置マスタ | Activity | true |
| `vaccine` | `vaccine` | 予防接種マスタ | Syringe | true |
| `medicine` | `medicine` | 薬剤マスタ | Pill | true |
| `diagnosis-category` | `diagnosis_category` | 診断カテゴリマスタ | FolderTree | false |
| `diagnosis-name` | `diagnosis_name` | 診断名マスタ | FileText | false |
| `hospitalization` | `hospitalization` | 入院マスタ | Bed | true |
| `cage` | `cage` | ケージマスタ | Building2 | false |
| `trimming-course` | `trimming_course` | トリミングコースマスタ | Scissors | true |
| `trimming-option` | `trimming_option` | トリミングオプションマスタ | Sparkles | true |
| `staff` | `staff` | スタッフマスタ | Users | false |
| `insurance` | `insurance` | 保険マスタ | ShieldCheck | false |

**画面構成（リストモード）:**
- ヘッダー: カテゴリ名（`config.IconComponent` 付き）+ 戻るボタン → `/settings` + 新規登録ボタン
- 検索バー（`[S] SearchFilterBar`）: `{config.labels.code}、{config.labels.name}で検索...`
- データテーブル（`[S] DataTable`）

リストモード テーブル列:
| 列 | className / align | 表示内容 | 条件 |
|---|---|---|---|
| コード | `w-[120px]` | `item.code`（等幅フォント） | 常時 |
| 名称 | - | `item.name`（太字） | 常時 |
| 所属カテゴリ | `w-[130px]` | `diagnosisCategories` からの名前解決 | `diagnosis_name` のみ |
| カテゴリ | `w-[100px]` | `item.category` | 常時 |
| 単価 | `w-[100px]`, align:right | `¥{price}` or 「-」（等幅フォント） | `showPrice` 時のみ |
| ステータス | `w-[100px]`, align:center | `StatusBadge`（`getMasterStatusColor` / `getMasterStatusLabel`） | 常時 |
| 操作 | `w-[80px]`, align:right | `RowActionButton`（編集） | 常時 |

**画面構成（編集モード）:**
- ヘッダー: 「{カテゴリ名} 編集/新規登録」、戻るボタン → リストモードへ
- カードコンテナ: `bg-white p-6 rounded-lg border shadow-sm space-y-4`

共通フォーム項目（全カテゴリ共通）:
| フィールド | 入力部品 | グリッド | 備考 |
|---|---|---|---|
| コード | `Input` | 2cols-左 | 必須（`*`マーク）、placeholder: `config.codePlaceholder` |
| 名称 | `Input` | 2cols-右 | 必須（`*`マーク）、placeholder: `config.namePlaceholder` |
| カテゴリ | `Input` | 2cols-左 | placeholder: `config.categoryPlaceholder` |
| 単価 (円) | `Input`（type=number） | 2cols-右 | `showPrice` 時のみ表示 |
| [カテゴリ固有セクション] | `MasterItemFormSections` | - | 下記参照 |
| 備考 / 詳細 | `Input` | full | placeholder: 補足情報など |
| ステータス | `RadioGroup`（有効 / 無効） | full | radio ボタン2つ |

**カテゴリ固有セクション（`[C] MasterItemFormSections` → `sections/`）:**

**examination（検査マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 検査項目リスト | 動的リスト（追加/削除） | 3カラムグリッド per 行 |
| └ 項目名 | `Input` | placeholder: RBC |
| └ 単位 | `Input` | placeholder: 例: mg/dL |
| └ 正常値 | `Input` | placeholder: 550-850 |
| └ 削除 | Trash2 ボタン | 赤色 ghost |
| 項目追加 | `Button`（Plus アイコン） | ヘッダー右上に配置 |

**vaccine（予防接種マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 対象種別 | `Select`（`VACCINE_SPECIES_VALUES`: 犬/猫/共通） | デフォルト: dog |
| 標準接種間隔 | `Input` | placeholder: 例: 1年 |

**medicine（薬剤マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 剤形 | `Select`（`DOSAGE_FORM_VALUES`: 錠剤/液剤/注射/外用薬/粉末） | デフォルト: tablet |
| 単位 | `Select`（`MEDICINE_UNIT_VALUES`: 1錠あたり/1mLあたり/1回分/1gあたり） | デフォルト: per_tablet |

**staff（スタッフマスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 役職 | `Select`（`STAFF_ROLE_VALUES`: 獣医師/動物看護師/トリマー/受付/管理者） | デフォルト: veterinarian |
| 資格番号 | `Input` | placeholder: 例: 獣医第12345号 |

**cage（ケージマスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| ケージタイプ | `Select`（`CAGE_TYPE_VALUES`: ICU（酸素室）/犬舎/猫舎/共用） | デフォルト: general |
| サイズ | `Select`（`CAGE_SIZE_VALUES`: 小型/中型/大型） | デフォルト: medium |

**insurance（保険マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 補償割合 (%) | `Select`（`COVERAGE_RATE_VALUES`: 50%/70%/80%/100%） | デフォルト: 70 |
| 請求先電話番号 | `Input` | placeholder: 例: 0120-XXX-XXX |

**trimming_course（トリミングコースマスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 対象サイズ | `Select`（`TARGET_SIZE_VALUES`: 小型犬/中型犬/大型犬/猫） | デフォルト: small |
| 所要時間 (分) | `Input`（type=number, Clock アイコン付き） | placeholder: 例: 60 |

**trimming_option（トリミングオプションマスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 追加所要時間 (分) | `Input`（type=number, Clock アイコン付き） | placeholder: 例: 15 |
| 併用可否 | `Select`（`COMBINABLE_VALUES`: 併用可/単独のみ） | デフォルト: yes |

**hospitalization（入院マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 対象体格 | `Select`（`BODY_SIZE_VALUES`: 小型/中型/大型） | デフォルト: small |
| 料金単位 | `Select`（`BILLING_UNIT_VALUES`: 1日あたり/1泊あたり） | デフォルト: per_day |

**diagnosis_name（診断名マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 診断カテゴリ | `Select`（diagnosis_category マスタ連動、active のみ） | 必須（`*`マーク）、`MasterLink` 付き、未登録時は警告メッセージ |

**diagnosis_category / その他（固有セクションなし）:**
- 共通フィールド（コード、名称、カテゴリ、[単価]）のみ

**編集モード アクション（`pt-4 border-t`）:**
| ボタン | 位置 | 動作 | 備考 |
|---|---|---|---|
| 削除 | 左（編集時のみ） | `handleDelete` → staff 時は `StaffImpactDialog` / 他は `ConfirmDialog` | Trash2 アイコン、赤文字 |
| キャンセル | 右 | `handleCloseEdit` → リストモードへ | outline |
| 保存 | 右 | `handleSave` | Save アイコン、`[S] PrimaryButton` |

**スタッフ固有ダイアログ（`[S] StaffImpactDialog`）:**
- ステータス変更・名称変更・削除時に影響範囲を確認
- `staffName`, `action`（rename / statusChange / delete）, `usage`（使用箇所数）を表示

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `Settings` | `[R]` | メインページ（リスト/編集切替） |
| `MasterItemFormSections` | `[C]` | カテゴリ固有フォームセクション（ディスパッチャー） |
| `ExaminationSection` | `[C]` | 検査マスタ固有: 検査項目リスト |
| `VaccineSection` | `[C]` | 予防接種マスタ固有: 対象種別・接種間隔 |
| `MedicineSection` | `[C]` | 薬剤マスタ固有: 剤形・単位 |
| `StaffSection` | `[C]` | スタッフマスタ固有: 役職・資格番号 |
| `CageSection` | `[C]` | ケージマスタ固有: タイプ・サイズ |
| `InsuranceSection` | `[C]` | 保険マスタ固有: 補償割合・電話 |
| `TrimmingCourseSection` | `[C]` | コースマスタ固有: 対象サイズ・所要時間 |
| `TrimmingOptionSection` | `[C]` | オプションマスタ固有: 追加時間・併用可否 |
| `HospitalizationSection` | `[C]` | 入院マスタ固有: 体格・料金単位 |
| `DiagnosisNameSection` | `[C]` | 診断名マスタ固有: 親カテゴリ選択 |
| `SectionWrapper` | `[C]` | セクション共通ラッパー |
| `PageLayout` | `[S]` | ページレイアウト |
| `SearchFilterBar` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `PrimaryButton` | `[S]` | 新規登録・保存ボタン |
| `StatusBadge` | `[S]` | ステータスバッジ |
| `RowActionButton` | `[S]` | 行アクションボタン |
| `StaffImpactDialog` | `[S][M]` | スタッフ変更影響確認ダイアログ |
| `ConfirmDialog` | `[S][M]` | 削除確認 |
| `MasterLink` | `[S]` | マスタ設定リンク |
| `useMasterItemEditor` | `[H]` | CRUD操作フック |
| `useMasterItems` | `[H]` | マスタデータ取得 |

**データ型:** `MasterItem`, `MasterFormData`, `MasterCategory`, `MasterSectionProps`, `CreateMasterItemDTO`, `UpdateMasterItemDTO`, `CategoryConfig` + 各カテゴリ固有型（`VaccineSpecies`, `DosageForm`, `MedicineUnit`, `StaffRole`, `CageType`, `CageSize`, `CoverageRate`, `TargetSize`, `Combinable`, `BodySize`, `BillingUnit`）

**ユーザー操作:**
- リストモード: 検索、行クリックで編集モードへ、新規登録ボタン
- 編集モード: 共通フィールド + カテゴリ固有フィールド入力
- ステータス切替（有効/無効ラジオ）
- 保存 / キャンセル / 削除
- staff カテゴリ: 名称変更・ステータス変更・削除時に影響確認ダイアログ
- diagnosis_name カテゴリ: 親カテゴリ（diagnosis_category マスタ）からの選択

---

## 11. 在庫管理（現在非表示）

### 11.1 在庫一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/inventory`（コメントアウト） |
| **コンポーネント** | `[R] InventoryList` |
| **目的** | 在庫品目の一覧表示・検索 |

### 11.2 在庫登録/編集

| 項目 | 内容 |
|------|------|
| **ルート** | `/inventory/new`、`/inventory/:id`（コメントアウト） |
| **コンポーネント** | `[R] InventoryForm` |
| **目的** | 在庫品目の追加・編集 |

**データ型:** `InventoryItem`, `InventoryCategory`, `InventoryStatus`
**カテゴリ:** 医薬品 / 消耗品 / フード / その他
**ステータス:** 在庫あり / 残りわずか / 在庫切れ
**特記:** カルテ保存時の在庫消費連動 (`consumeStock`) は `useInventory` フック経由で実装済み。

---

## 12. 共通ペット選択画面

以下の4画面は `[S] PetSelection` 共通コンポーネントを使用:

| ルート | コンポーネント | 選択後の遷移先 |
|---|---|---|
| `/medical-records/select-pet` | `MedicalRecordPetSelection` | `/medical-records/new?petId=xxx` |
| `/hospitalization/select-pet` | `HospitalizationPetSelection` | `/hospitalization/new?petId=xxx` |
| `/trimming/select-pet` | `TrimmingPetSelection` | `/trimming/new?petId=xxx` |
| `/accounting/select-pet` | `AccountingPetSelection` | `/accounting/new?petId=xxx` |

**共通画面構成:**
- ペット検索フォーム（`[S] PetSearchForm`）: 飼主ID、飼主名、電話、ペット名、種
- 検索結果テーブル（`[S] PetSearchResultsTable`）
- 選択ボタンクリックでクエリパラメータ付き遷移

**データ型:** `PetSearchParams`, `Pet`

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

Settings(/settings) ──→ ClinicSettings / Settings({category})
```

---

## 全ルート一覧（28ルート + Fallback）

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
| 25 | `/settings` | マスタ設定トップ | `MasterSettingsIndex` |
| 26 | `/settings/clinic` | 病院情報設定 | `ClinicSettings` |
| 27 | `/settings/{category-slug}` | マスタカテゴリ設定 (×14) | `Settings` |
| 28 | `*` | 404ページ | インライン |