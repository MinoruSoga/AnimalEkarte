# 動物病院管理システム 画面仕様書

本ドキュメントは、全画面（ルート）ごとの仕様を定義します。
各画面のルートパス、目的、構成コンポーネント、データフロー、ユーザー操作を網羅しています。

> **デザイン・UI共通仕様**:
> - 操作ボタンはNotion風に統一（メイン: `#038B94`, プライマリ: `#2383E2`）
> - 各行の操作アイコンは `MoreHorizontal` + `size-5` を使用
> - 内部ID（〇〇No, 〇〇ID等）はUI上非表示とする

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
- カード間のドラッグ&ドロップによるステータス遷移（`dnd-kit`）
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
| 受付済 | 飼主詳細、���カルテ作成」（診療時、同時に診療中へ移動）/ 「診察を開始する」（非診療時） |
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
| 件数制限 | 1日最大4件表示��超過分は「他 N 件」表示 |

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
| ドラ���グ&ドロップ | Y軸ドラッグで時間変更（15分単位スナップ、`motion/react` 使用） |
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
- 詳細モー��ルから編集・キャンセル・関連レコード作成への遷移
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
- 検索バー（`[S] NotionFilter`）+ 件数表示
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

**画面構成:**
- ヘッダー: タイトル + 保存ボタン
- 飼主情報セクション（2カラムグリッド）:
  - 飼主名 (必須)、飼主名カナ (必須)、会社名、郵便番号、住所1/2、自宅住所1/2、生年月日、電話番号 (必須)、会社電話、メール、備考、危険フラグ (Switch)、割引率、会員種別 (Select)

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

- ペット一覧テーブル:
  - ペット追加ボタン
  - 各ペット行にドロップダウン（カルテ作成、予約、トリミング、入院、会計、編集、削除）
- ペット編集モーダル（`[C] PetEditModal`）
- フォーム離脱保護（`[S] NavigationBlocker`）

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

- 3カラムグリッド（md:2, lg:3）
- 必須フィールド: ペット名、種別、性別、生年月日
- バリデーション: 必須未入力時にトースト通知

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
- 検索バー（`[S] NotionFilter`）
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

> **詳細仕様**: タブごとの詳細仕様は `/SCREENS_DETAILED_TABS.md` を参照してください。

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

**タブ別フォーム項目詳細:**

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
- **治療プラン**: `TreatmentTable`（マスタ検索ダイアログ連携）
  - TreatmentTable 列: 種別、治療内容、メモ、保険、単価(税込)、数量、割引(%)、値引(￥)、小計、操作
  - **種別**: マスタから選択時に自動設定される読み取り専用フィールド（診察/処置/薬剤/検査/予防接種/定期健診）
- **集計**: `TreatmentDetailedSummary`（小計、税、合計、割引率、値引額）

**Tab 3: 治療（`MedicalRecordTreatment`）**
- **治療プランテーブル**: `TreatmentTable`（`onMarkCompleted` チェックボックス列あり。チェック→確認ダイアログ→治療済みテーブルへ移動）
- **治療済みテーブル**: `TreatmentTable`（`onRevertToPlan` 戻すボタンあり。確認ダイアログ→治療プランへ差し戻し）
- **移動確認ダイアログ**: `TreatmentMoveConfirmDialog`（`AlertDialog`ベース、`pendingMove`ステート制御）
- **集計**: `TreatmentDetailedSummary`
- TreatmentTable 列: [済チェックボックス]、種別、治療内容、メモ、保険、単価(税込)、数量、割引(%)、値引(￥)、小計、操作
  - **種別**: マスタから選択時に自動設定される読み取り専用フィールド（診察/処置/薬剤/検査/予防接種/定期健診）

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
  - TreatmentTable 列: 種別、治療内容、メモ、保険、単価(税込)、数量、割引(%)、値引(￥)、小計、操作
- **集計**: `TreatmentDetailedSummary`
- **コメント / 備考**: 2カラム `Textarea`
- **アクション**: 「PDF出力」ボタン

**Tab 8: 会計(医師確認)（`MedicalRecordBillCheck`）**
- **明細テーブル**: `TreatmentTable`（治療タブの completedItems を自動同期）
  - TreatmentTable 列: 種別、治療内容、メモ、保険、単価(税込)、数量、割引(%)、値引(￥)、小計、操作
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
- 入力/履歴のタブ切替、履歴ではトレンドアイコン（↑↓→）表示

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

**アクセシビリティ:**
- 担当医エラー: `PatientInfoCard.staffAriaDescribedBy` → `FormFieldError`（`role="alert"`）と `aria-describedby` 接続
- 保存バリデーション: 担当医未選択時にトースト警告＋スタッフボタンにエラー状態表示

#### 4.3.1 TreatmentTable 詳細仕様

**概要:**
治療項目を管理するテーブルコンポーネント。各タブ（診察/治療プラン、治療、見積書、会計）で使用され、表示カラムや操作ボタンがコンテキストに応じて変化する。

**TreatmentItem型定義:**
| フィールド | 型 | 必須 | 説明 |
|---|---|---|---|
| `id` | `string` | ○ | 一意識別子（`String(Date.now())`で生成） |
| `selected` | `boolean` | - | チェックボックス選択状態（見積書タブで使用） |
| `status` | `TreatmentStatus` | - | 治療ステータス（"未完了" / "完了" / "-"）。治療タブのみ使用 |
| `type` | `TreatmentType` | - | **種別**（"診察" / "処置" / "薬剤" / "検査" / "予防接種" / "定期健診"）。マスタから自動設定 |
| `content` | `string` | ○ | 治療内容（項目名） |
| `memo` | `string` | ○ | メモ・備考 |
| `insurance` | `boolean` | ○ | 保険適用フラグ |
| `unitPrice` | `number` | ○ | 単価（税込、円） |
| `quantity` | `number` | ○ | 数量 |
| `discountRate` | `number` | ○ | 割引率（%、0-100） |
| `discountAmount` | `number` | ○ | 値引額（円） |
| `inventoryId` | `string` | - | 在庫連携ID（将来拡張用） |

**TreatmentType値:**
```typescript
const TREATMENT_TYPE_VALUES = ["診察", "処置", "薬剤", "検査", "予防接種", "定期健診"] as const;
```

**カテゴリ→種別変換マッピング（`getCategoryTreatmentType`）:**
| マスタカテゴリ（英語キー） | 種別（日本語ラベル） |
|---|---|
| `consultation` | 診察 |
| `procedure` | 処置 |
| `medicine` | 薬剤 |
| `examination` | 検査 |
| `vaccine` | 予防接種 |
| `checkup` | 定期健診 |
| `hospitalization` | （種別なし：入院は治療項目外） |

**表示カラム構成（タブ別）:**

| カラム | 診察/治療プラン | 治療（プラン） | 治療（済） | 見積書 | 会計 |
|---|:---:|:---:|:---:|:---:|:---:|
| **済チェック** | - | ○ | - | - | - |
| **選択チェック** | - | - | - | ○ | - |
| **種別** | ○ | ○ | ○ | ○ | ○ |
| **治療内容** | ○ | ○ | ○ | ○ | ○ |
| **メモ** | ○ | ○ | ○ | ○ | ○ |
| **保険** | ○ | ○ | ○ | ○ | ○ |
| **単価(税込)** | ○ | ○ | ○ | ○ | ○ |
| **数量** | ○ | ○ | ○ | ○ | ○ |
| **割引(%)** | ○ | ○ | ○ | ○ | ○ |
| **値引(￥)** | ○ | ○ | ○ | ○ | ○ |
| **小計** | ○ | ○ | ○ | ○ | ○ |
| **操作** | ○ | ○ | ○ | ○ | ○ |

**カラム詳細:**

1. **種別カラム**
   - 幅: `w-20`（80px）
   - 表示: 読み取り専用テキスト（`text-xs`、`C.textSecondary`）
   - データ: `item.type`（TreatmentType）
   - 未設定時: 空欄表示
   - 編集: 不可（マスタ選択時に自動設定）

2. **治療内容カラム**
   - 幅: `min-w-[160px]`
   - 編集: `Input`（テキスト入力）
   - プレースホルダー: "治療内容を入力"

3. **メモカラム**
   - 幅: `min-w-[120px]`
   - 編集: `Input`（テキスト入力）
   - プレースホルダー: "メモ"

4. **保険カラム**
   - 幅: `w-16`（64px）
   - 編集: `NotionCheckbox`（クリックでトグル）
   - 表示: チェックマーク（`#2383E2` ブルー）

5. **単価(税込)カラム**
   - 幅: `w-24`（96px）
   - 編集: `Input` + `useNumericInput`（数値検証）
   - フォーマット: `formatCurrency`（カンマ区切り）
   - 右揃え（`text-right`）

6. **数量カラム**
   - 幅: `w-16`（64px）
   - 編集: `Input` + `useNumericInput`（整数のみ）
   - 右揃え

7. **割引(%)カラム**
   - 幅: `w-20`（80px）
   - 編集: `Input` + `useNumericInput`（0-100範囲）
   - 右揃え

8. **値引(￥)カラム**
   - 幅: `w-24`（96px）
   - 編集: `Input` + `useNumericInput`
   - フォーマット: `formatCurrency`
   - 右揃え

9. **小計カラム**
   - 幅: `w-28`（112px）
   - 表示: 読み取り専用（`calcLineTotal`で自動計算）
   - フォーマット: `formatCurrency`
   - 右揃え、太字（`font-semibold`）

10. **操作カラム**
    - 幅: `w-12`（48px）
    - ボタン: 削除（`Trash2` アイコン、`size-4`、`text-destructive`）
    - ホバー時のみ表示（`opacity-0 group-hover:opacity-100`）

**マスタ検索連携:**
- **検索ボタン**: テーブル下部に「+ 治療項目を追加」ボタン（`TreatmentSearchDialog`トリガー）
- **検索ダイアログ**: `TreatmentSearchCommand`（`Command`コンポーネント）でマスタ項目を検索
- **フィルタ条件**: 
  - `status === "active"`（有効項目のみ）
  - `category in TREATMENT_CATEGORIES`（consultation, examination, procedure, vaccine, checkup, hospitalization, medicine）
  - `price != null && price > 0`（価格設定済み）
- **選択時動作**:
  1. `TreatmentMasterItem`（code, name, unitPrice, category）を受け取る
  2. `getCategoryTreatmentType(item.category)`で種別を自動設定
  3. 新規`TreatmentItem`を作成（id: `String(Date.now())`、type: 自動設定、content: `item.name`、unitPrice: `item.unitPrice`、memo: ""、insurance: true、quantity: 1、割引0）
  4. テーブル末尾に追加

**集計機能（`TreatmentDetailedSummary`）:**
| 項目 | 計算式 |
|---|---|
| 小計 | Σ `calcLineTotal(unitPrice, quantity, discountRate, discountAmount)` |
| 全体割引 | 割引率（%）入力 / 値引額（円）入力 |
| 税込合計 | `calcGrandTotal(小計, 全体割引率, 全体値引額)` |
| 消費税（内税） | `total * (TAX_RATE / (100 + TAX_RATE))` |

**計算式詳細（`/lib/format.ts`）:**
```typescript
// 行小計 = (単価 × 数量) × (1 - 割引率/100) - 値引額
calcLineTotal(unitPrice, quantity, discountRate, discountAmount): number

// 合計 = 小計 × (1 - 全体割引率/100) - 全体値引額
calcGrandTotal(subtotal, globalDiscountRate, globalDiscountAmount): { tax, total }
```

**行追加・削除:**
- **手動追加**: 「+ 行を追加」ボタン（空行挿入）
- **マスタ追加**: 「+ 治療項目を追加」ボタン→検索ダイアログ→マスタ選択
- **削除**: 各行の削除ボタン（確認なし即削除）

**アニメーション:**
- 行追加時: `motion.div`で`initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }}`
- `useReducedMotion`フックで`prefers-reduced-motion: reduce`を検知し、アニメーション無効化

**治療タブ専用機能（`MedicalRecordTreatment`）:**
- **治療プラン→治療済み移動**:
  1. 「済」チェックボックスをオン
  2. `TreatmentMoveConfirmDialog`で確認（「[項目名] を治療済みに移動しますか？」）
  3. 確認後、`planItems`から削除→`completedItems`に追加
  4. スクリーンリーダー通知（`useAnnounce`）: "治療済みに移動しました"
  5. 治療済みテーブルの見出しにフォーカス移動（`completedHeadingRef`）
- **治療済み→治療プラン戻し**:
  1. 「プランに戻す」ボタンクリック
  2. `TreatmentMoveConfirmDialog`で確認
  3. 確認後、`completedItems`から削除→`planItems`に追加
  4. 治療プランテーブルの見出しにフォーカス移動

**見積書タブ専用機能（`MedicalRecordEstimate`）:**
- チェックボックス列なし
- 戻すボタンなし
- 全項目選択可能（将来拡張用）

**会計タブ専用機能（`MedicalRecordBillCheck`）:**
- 治療タブの`completedItems`を自動同期（読み取り専用）
- チェックボックス列なし
- 操作カラムなし

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
- ��ード/行クリックで詳細画面へ
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

**アクセシビリティ:**
- 担当医エラー: `PatientInfoCard.staffAriaDescribedBy` → `FormFieldError`（`role="alert"`）と `aria-describedby` 接続
- 保存バリデーション: 担当医未選択時にトースト警告＋スタッフボタンにエラー状態表示

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

**印刷関連:**
- `PrintPreviewDialog`（`[S][M]`）: 入院サマリープレビューダイアログ
- `HospitalizationSummaryDocument`（`[C]`）: 入院サマリー帳票（入院日数自動計算・1日あたり費用表示）
- `usePrint<HospDocumentType>`（`[H]`）: 印刷状態管理

**ケアプラン（`[C] CarePlan/`）:**
| コンポーネント | 説明 |
|---|---|
| `CarePlanPreviewPopover` | ケアプラン概要ポップオーバー（ステータストグル付き。active↔completed の即時切り替えが可能） |
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
| マスタ連動情報 | Badge 表示（単価(税込)、カテゴリ） | マスタ選択時のみ表示 |
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
| 記録���刻 | `Input`（type=time） | 現在時刻で初期化 |
| 内容・量 | `Input` | 種別依存のプレースホルダー |
| 詳細メモ | `Textarea` | |

**TaskCompleteDialog フォーム項目（タスク完了記録用）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| タスク情報 | 読み取り専用表示 | タスク名 + 詳細（背景カード） |
| 実施時刻 | `Input`（type=time） | 現在時刻で初期化 |
| 実施メモ (任意) | `Textarea` | |

**データ型:** `Hospitalization`, `CarePlanItem`, `DailyRecord`, `VitalRecord`, `CareLogRecord`, `StaffNoteRecord`, `Task`, `TimelineItem`, `HospDocumentType`

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
- 入院サマリーの印刷プレビュー → 印刷（`usePrint<HospDocumentType>` + `HOSP_DOCUMENT_TYPE_LABELS` で動的タイトル）

**印刷機能:**
- ヘッダーアクションに印刷ボタン表示
- `PrintPreviewDialog` でプレビュー表示、`window.print()` で印刷実行
- 印刷エリア（`hidden print:block`、`data-print-area` 属性）に `HospitalizationSummaryDocument` を配置

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
- フィルタ: `[S] NotionFilter`（キーワード検索「飼主名、ペット名...」）+ 日付範囲フィルタ
- データテーブル（`[S] DataTable`）
- ページネーション（`[S] Pagination`、20件/ページ）
- 削除確認ダイアログ（`[S] ConfirmDialog`）

**フィルタ項目:**
| 項目 | 入力部品 | 備考 |
|---|---|---|
| 開始日 | `NotionDatePicker` | `lg:w-[160px]` |
| 終了日 | `NotionDatePicker` | `lg:w-[160px]`、「〜」で接続 |
| キーワード | `NotionFilter` | 飼主名・ペット名で検索 |
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
| `NotionFilter` | `[S]` | 検索フィルタバー |
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
| **画面構成** | `[S] PetSearchForm` + `[S] PetSearchResultsTable` 共通コンポーネント使用 |

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
| コース選択 | `MasterSelectTrigger` → `MasterSelectModal`���trimming_course ��スタ連動） | `MasterLink` 付き、選択時に `charge` 自動反映 |
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
- `MasterSelectModal`（staff マスタ連動、active のみ���
- タイトル: 「担当スタッフを選択」

**バリデーション:**
- 担当スタッフ未選択時はトースト警告（`toast.warning`）で保存ブロック

**アクセシビリティ:**
- 担当医エラー: `PatientInfoCard.staffAriaDescribedBy` → `FormFieldError`（`role="alert"`）と `aria-describedby` 接続
- コース選択エラー: `MasterSelectTrigger.ariaDescribedBy` → `FormFieldError` と `aria-describedby` 接続
- `MasterSelectTrigger`: 選択済み・未選択状態ともに `<button>` 要素（キーボード操作対応）

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
- 検索バー（`[S] NotionFilter`）: 「飼主名、ペット名、検査種別...」
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
| `NotionFilter` | `[S]` | 検索フィルタバー |
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
- 検索バー（`[S] NotionFilter`）: 「飼主名、ペット名...」
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
| 保険 | align:center | 保険名バッジ（保険あり時のみ表示、`bg-[#D3E5EF] text-[#183B56]`） |
| ソース | align:center | 「入院連携」バッジ（`source === "hospitalization"` 時、`bg-cyan-50 text-cyan-700`） |
| ステータス | `w-[100px]` | `StatusBadge`（`getAccountingStatusColor` / `getAccountingStatusLabel`） |
| 操作 | `w-[100px]`, align:right | `RowActionDropdown`（編集 / 削除） |

**保��フィルター:**
- `INSURANCE_FILTER_VALUES` 型の `ToggleGroup`（全て / 保険あり / 保険なし）
- テーブル上部のフィルタバー右端に配置

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
| `NotionFilter` | `[S]` | 検索フィルタバー |
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
| **画面構成** | `[S] PetSearchForm` + `[S] PetSearchResultsTable` 共通コンポーネント使用 |

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
| フィ��ルド | 入力部品 | 備考 |
|---|---|---|
| 区分 | `Select`（`MANUAL_ITEM_CATEGORY_VALUES`: 療法食・フード / 物販・ケア用品 / その他） | |
| 品目名 | `Input` | placeholder: 例: ロイヤルカナン 3kg |
| 単価 (税込) | `Input`（type=number） | placeholder: 0 |

明細テーブル列:
| 列 | className / align | 表示内容 |
|---|---|---|
| 区分 | `w-[100px]` | `Badge`（`getItemCategoryLabel`） |
| 項目名 | - | `item.name` + カルテ連携バッジ（`source === "medical_record"` 時） |
| 単価(税込) | align:right, `w-[100px]` | `¥{unitPrice}` |
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
| 保険負担額 | 読���取り専用 | 緑背景カード、マイナス表示 |

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
- **保険負担内訳**: 保険適用額・自己負担額・保険者負担額の3行セット（保険設定時のみ表示）
- **シミュレーション注釈**: 保険負担割合別の参考金額を注記表示
- **入院連携バッジ**: `source === "hospitalization"` の明細行に「入院連携」バッジ表示

**データ初期化（`useAccountingDetail`）:**
- 新規時: `location.state` から `accountingItems`（カルテ連携）/ `hospitalizationId`（入院連携）または `petId` クエリパラメータから生成
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
- 明細項目の手動追加（ダイアログ: 区分・品目名・単価(税込)）/削除
- 保険適用切替（Switch）、負担割合選択
- 支払方法選択（現金/カード/電子マネー）
- 預り金入力 + クイック入力ボタン → お釣り自動計算
- 会計確定（確認ダイアログ → ステータス completed へ遷移）
- 精算完了後: 領収書/診療明細書のプレビュー → 印刷

**アクセシビリティ:**
- 預り金入力: `aria-invalid` + `aria-describedby` → `FormFieldError`（`role="alert"`）接続（不足額エラー時）
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
- ヘッダー: タイトル（Syringe アイコン）+ 「新規登録」ボタン（`[S] PrimaryButton`、Plus アイコン）→ `/medical-records/select-pet`（state: `{ activeTab: "予防接種" }`）
- 検索バー（`[S] NotionFilter`）: 「飼主名、ペット名、予防接種名...」
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
| `NotionFilter` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `Pagination` | `[S]` | ページネーション |
| `useVaccinations` | `[H]` | 検索・フィルタロジック |
| `useStaffValidation` | `[H]` | スタッフ有効性チェック |
| `usePagination` | `[H]` | ページネーション |

**データ型:** `VaccinationRecord`, `DataTableColumn`

---

## 10. 定期健診管理

### 10.1 定期健診一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/checkups` |
| **コンポーネント** | `[R] CheckupList` |
| **目的** | 定期健診記録の一覧管理 |

**画面構成:**
- ヘッダー: タイトル（ClipboardCheck アイコン）+ 「新規登録」ボタン（`[S] PrimaryButton`、Plus アイコン）→ `/medical-records/select-pet`（state: `{ activeTab: "定期健診" }`）
- 検索バー（`[S] NotionFilter`）: 「飼主名、ペット名、健診種別...」
- データテーブル（`[S] DataTable`）
- ページネーション（`[S] Pagination`、20件/ページ）

**テーブル列:**
| 列 | className / align | 表示内容 |
|---|---|---|
| 実施日 | `w-[120px]` | `r.date`（等幅フォント） |
| 飼主名 | - | `r.ownerName` |
| ペット名 | - | `r.petName` |
| 健診種別 | - | `r.checkupType`（太字） |
| 結果概要 | - | `r.result`（未設定時「-」） |
| 担当医 | `w-[100px]` | `r.doctor`（無効スタッフ時は赤文字＋AlertTriangle、未設定時「-」） |
| 次回予定 | `w-[140px]` | `r.nextDate`（等幅フォント） |
| 操作 | `w-[100px]`, align:right | `RowActionDropdown`（「カルテを開く」のみ） |

**行アクション:**
| アクション | アイコン | 動作 |
|---|---|---|
| カルテを開く | FileText | `/medical-records/{medicalRecordId}` へ遷移（state: `{ from: "/checkups" }`） |

**特記:** 定期健診の新規登録はカルテから行う。一覧画面は参照+カルテ遷移のみ。行クリックでもカルテへ遷移する。

**健診種別（モックデータ）:**
| 種別名 | 説明 |
|---|---|
| 年次健康診断 | 1年ごとの定期健康チェック |
| シニア健康診断 | 高齢ペット向け（半年ごと推奨） |
| パピー健診 | 幼齢ペット向け成長確認（3ヶ月ごと推奨） |

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `CheckupList` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `NotionFilter` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `Pagination` | `[S]` | ページネーション |
| `useCheckups` | `[H]` | 検索・フィルタロジック |
| `useStaffValidation` | `[H]` | スタッフ有効性チェック |
| `usePagination` | `[H]` | ページネーション |

**データ型:** `CheckupRecord`, `DataTableColumn`

---

## 11. 設定・マスタ管理

### 11.1 マスタ設定トップ

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

### 11.2 病院情報設定

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
| `NotionPropertyRow` | `[S]` | Notion風プロパティ行（label/required/align） |
| `NotionSectionLabel` | `[S]` | Notion風セクションラベル |
| `NotionSectionDivider` | `[S]` | Notion風薄罫線ディバイダー |
| `useClinicInfo` | `[H]` | 病院情報CRUD |

**データ型:** `ClinicInfo`（name, branchName, postalCode, address, phoneNumber, faxNumber?, registrationNumber?, directorName?, email?, website?, logoUrl?）

### 11.3 診療項目マスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/treatment-items` |
| **コンポーネント** | `[R] TreatmentItemsSettings` |
| **目的** | 診察・検査・処置・予防接種・定期健診の5カテゴリを1ページで管理 |

**タブ構成:**
| タブラベル | カテゴリキー | showPrice | showCode | showCategory | showParentItem |
|---|---|---|---|---|---|
| 診察 | `consultation` | true | false | false | true |
| 検査 | `examination` | true | false | false | true |
| 処置 | `procedure` | true | false | false | true |
| 予防接種 | `vaccine` | true | false | false | true |
| 定期健診 | `checkup` | true | false | false | true |

- ヘッダー: 「診療項目マスタ」（Stethoscope アイコン）+ 戻るボタン → `/settings`
- タブ切替: `Tabs` / `TabsList` / `TabsTrigger` / `TabsContent`（Notion風スタイル）
- 各タブ内は `[C] SettingsContent`（`embedded` モード）で一覧/編集を描画（ツリー表示対応）
- 編集モード時は `PageLayout` 付きの独立フォームに遷移

### 11.4 診断マスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/diagnosis` |
| **コンポーネント** | `[R] DiagnosisSettings` |
| **目的** | 診断カテゴリと診断名の2カテゴリを1ページで管理 |

**タブ構成:**
| タブラベル | カテゴリキー | showPrice | showCode | showCategory |
|---|---|---|---|---|
| カテゴリ | `diagnosis_category` | false | false | true |
| 診断名 | `diagnosis_name` | false | false | true |

- ヘッダー: 「診断マスタ」（FolderTree アイコン）+ 戻るボタン → `/settings`
- タブ切替: `Tabs` / `TabsList` / `TabsTrigger` / `TabsContent`（Notion風スタイル）
- 各タブ内は `[C] SettingsContent`（`embedded` モード）で一覧/編集を描画

### 11.5 トリミングマスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/trimming` |
| **コンポーネント** | `[R] TrimmingSettings` |
| **目的** | トリミングコースとオプションの2カテゴリを1ページで管理 |

**タ��構成:**
| タブラベル | カテゴリキー | showPrice | showCode | showCategory | showParentItem |
|---|---|---|---|---|---|
| コース | `trimming_course` | true | false | false | true |
| オプション | `trimming_option` | true | false | false | true |

- ヘッダー: 「トリミングマスタ」（Scissors アイコン）+ 戻るボタン → `/settings`
- タブ切替: `Tabs` / `TabsList` / `TabsTrigger` / `TabsContent`（Notion風スタイル）
- 各タブ内は `[C] SettingsContent`（`embedded` モード）で一覧/編集を描画（ツリー表示対応）

### 11.6 マスタカテゴリ設定（個別ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/{category-slug}`（6パターン） |
| **コンポーネント** | `[R] Settings` → `[C] SettingsContent` |
| **目的** | 各マスタカテゴリのアイテムCRUD |

**ルートマッピング（6カテゴリ）:**
| スラグ | カテゴリキー | ラベル | アイコン | showPrice | showCode | showCategory | showParentItem |
|---|---|---|---|---|---|---|---|
| `service-type` | `serviceType` | 予約区分マスタ | Activity | false | false | false | false |
| `medicine` | `medicine` | 薬剤マスタ | Pill | true | false | false | true |
| `hospitalization` | `hospitalization` | 入院マスタ | Bed | true | false | false | true |
| `cage` | `cage` | ケージマスタ | Building2 | false | true | true | false |
| `staff` | `staff` | スタッフマスタ | Users | false | true | true | false |
| `insurance` | `insurance` | 保険マスタ | ShieldCheck | false | false | true | false |

**画面構成（リストモード）:**
- ヘッダー: カテゴリ名（`config.IconComponent` 付き）+ 戻るボタン → `/settings` + 新規登録ボタン
- 検索バー（`[S] NotionFilter`）: `showCode` 時は `{config.labels.code}、{config.labels.name}で検索...`、それ以外は `{config.labels.name}で検索...`
- データテーブル（`[S] DataTable`）

**ツリー表示（`showParentItem: true` のカテゴリ）:**
- ドラッグ中にテーブル上部に「トップレベルに移動」ドロップゾーンが出現
- 各行にドラッグハンドル（`GripVertical` アイコン）付き
- トップレベル項目＝「カテゴリ」として機能、Chevronで展開/折りたたみ
- 子項目数をカウントバッジで表示
- 最下層のみ金額表示、親は金額欄を空欄表示
- 操作列: 「+」ボタン（子項目インライン追加）+ 編集ボタン
- D&D操作: ドラッグで項目を別の親にドロップして所属カテゴリ変更（自己参照・子孫へのドロップは防止）
- D&D並び順変更: 行の上端25%にドロップで「前に挿入」、下端25%で「後に挿入」、中央50%で「子項目化」。`sortOrder`フィールドで永続化、`bulkUpdate`で兄弟全体のsortOrderを一括更新
- カスタムドラッグプレビュー: `setDragImage`でNotionライクなピル型ゴースト（GripVerticalアイコン＋項目名）を表示
- ホバー自動展開: 折りたたまれた親ノードの中央ゾーンに600msホバーで自動展開
- キーボードアクセシビリティ: ドラッグハンドルに `role="button"` / `tabIndex={0}` / `aria-roledescription` / `aria-label` を付与。`aria-grabbed` / `aria-dropeffect="move"` で状態通知。`Alt+ArrowUp/Down` でキーボードのみの並び替え、`Alt+ArrowLeft` で親の階層に昇格、`Alt+ArrowRight` で直前の兄弟の子に降格（循環参照防止付き）。移動後はハンドルにフォーカス自動復帰。`useAnnounce` で全D&D操作結果を `aria-live="polite"` リージョンにリアルタイム通知
- インライン追加行: `Enter` で追加（連打で複数登録可）、`Esc` で閉じる

リストモード テーブル列:
| 列 | className / align | 表示内容 | 条件 |
|---|---|---|---|
| コード | `w-[120px]` | `item.code`（等幅フォント） | `showCode` 時のみ |
| 名称 | - | `item.name`（太字）、ドラッグハンドル付き。ツリー時はインデント+Chevron追加。フラット時も`GripVertical`+D&D並び替え対応 | 常時 |
| 所属カテゴリ | `w-[130px]` | `diagnosisCategories` からの名前解決 | `diagnosis_name` のみ |
| 分類 | `w-[100px]` | `item.category` | `showCategory` 時のみ（`showParentItem: true` のカテゴリでは非表示） |
| 単価(税込) | `w-[100px]`, align:right | `¥{price}` or 「-」（等幅フォント） | `showPrice` 時のみ |
| ステータス | `w-[100px]`, align:center | `StatusBadge`（`getMasterStatusColor` / `getMasterStatusLabel`） | 常時 |
| 操作 | `w-[80px]`, align:right | 編集ボタン（行クリックでも編集モードへ遷移） | 常時 |

**画面構成（編集モード）:**
- ヘッダー: 「{カテゴリ名} 編集/新規登録」、戻るボタン → リストモードへ
- カードコンテナ: `bg-white p-6 rounded-lg border shadow-sm space-y-4`

共通フォーム項目（全カテゴリ共通）:
| フィールド | 入力部品 | グリッド | 備考 |
|---|---|---|---|
| カテゴリ | `Select`（カテゴリ候補リスト） | full | `showParentItem` 時のみ表示、循環参照防止済み。「なし（トップレベル）」選択可 |
| コード | `Input` | 2cols-左 | `showCode` 時のみ表示、必須（`*`マーク）、placeholder: `config.codePlaceholder` |
| 名称 | `Input` | 2cols-右 | 必須（`*`マーク）、placeholder: `config.namePlaceholder` |
| 分類 | `Input` | 2cols-左 | `showCategory` 時のみ表示、placeholder: `config.categoryPlaceholder` |
| 単価(税込) | `Input`（type=number）/ 「子項目で金額を設定」表示 | 2cols-右 | `showPrice` 時のみ表示。子を持つ親は入力不可 |
| [カテゴリ固有セクション] | `MasterItemFormSections` | - | 下記参照 |
| 備考 / 詳細 | `Input` | full | placeholder: 補足情報など |
| ステータス | `RadioGroup`（有効 / 無効） | full | radio ボタン2つ |

**カテゴリ固有セクション（`[C] MasterItemFormSections` → `sections/`）:**

**examination（検査マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 検査項目リスト | 動的リスト（追加/削除） | 3カラムグリッド per 行 |
| └ 項目名 | `Input` | placeholder: RBC |
| └ ��位 | `Input` | placeholder: 例: mg/dL |
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
| 職種 | `Combobox`（job_title マスタ連動、active のみ） | `MasterLink` 付き |
| 資格番号 | `Input` | placeholder: 例: 獣医第12345号 |
| 所属医院 | `DropdownMenu`（`DropdownMenuCheckboxItem`） | 複数選択可、MOCK_CLINICS 連動 |
| メールアドレス | `Input`（type=email） | placeholder: 例: user@example.com |
| パスワード | `Input`（type=password） | 未入力で変更なし（新規登録時は必須） |
| ユーザー種別 | `Combobox`（`USER_TYPE_VALUES`: システム管理者/医院管理者/スタッフ） | デフォルト: staff |

**スタッフマスタ 一覧表示列:**
| 列 | 表示内容 |
|---|---|
| 名称 | `item.name`（ドラッグハンドル付き） |
| 職種 | job_title マスタから解決した名称 |
| 所属医院 | clinics配列から解決した医院名（カンマ区切り） |
| メールアドレス | `item.email` |
| 最終ログイン | `item.lastLoginAt`（yyyy/MM/dd HH:mm 形式） |
| ステータス | StatusBadge（有効/無効） |
| 操作 | RowActionDropdown（編集） |

**特記事項:**
- スタッフマスタでは社員番号フィールドは存在せず、コード列も非表示
- 職種は job_title マスタから動的に取得（単一選択、Combobox形式）
- 所属医院は複数選択可能（チェックボックス形式のドロップダウンメニュー）
- ユーザー種別により権限管理を実施（システム管理者/医院管理者/スタッフ）
- セクション構成: スタッフ詳細（職種・資格番号） / 所属情報（所属医院） / アカウント発行・権限設定（メールアドレス・パスワード・ユーザー種別）

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

**trimming_option（トリミング���プションマスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 追加所要時間 (分) | `Input`（type=number, Clock アイコン付き） | placeholder: 例: 15 |
| 併用可否 | `Select`（`COMBINABLE_VALUES`: 併用可/単独のみ） | デフォルト: yes |

**hospitalization（入院マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 対象体格 | `Select`（`BODY_SIZE_VALUES`: 小型/中型/大型） | デフォルト: small |
| 料金単位 | `Select`（`BILLING_UNIT_VALUES`: 1日あたり/1泊あたり） | デフォルト: per_day |

**consultation（診察マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 適用区分 | `Select`（`CONSULTATION_TIME_VALUES`: 常時/初診/再診/時間外/緊急） | デフォルト: anytime |
| 標準診察時間 | `Input` | placeholder: 例: 15分 |

**procedure（処置マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 所要時間(目安) | `Input` | placeholder: 例: 30分 |
| 麻酔要否 | `Select`（`ANESTHESIA_VALUES`: 不要/局所麻酔/鎮静/全身麻酔） | デフォルト: none |

**checkup（定期健診マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 推奨受診間隔 | `Input` | placeholder: 例: 1年 |
| 対象年齢 | `Select`（`CHECKUP_TARGET_AGE_VALUES`: 全年齢/幼齢(〜1歳)/成年(1〜7歳)/シニア(7歳〜)） | デフォルト: all |

**serviceType（予約区分マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 表示カラー | カラーピッカー（`SERVICE_TYPE_COLOR_VALUES` 丸ボタン群、選択時チェック+ring） | デフォルト: default |
| プレビュー | 読み取り専用バッジ | 選択色+名称のプレビュー表示 |

**diagnosis_name（診断名マスタ）:**
| フィールド | 入力部品 | 備考 |
|---|---|---|
| 診断カテゴリ | `Select`（diagnosis_category マスタ連動、active のみ） | 必須（`*`マーク）、`MasterLink` 付き、未登録時は警告メッセージ |

**diagnosis_category（固有セクションなし）:**
- 共通フィールド（コード、名称、カテゴリ、[単価(税込)]）のみ

**編集モード アクション（`pt-4 border-t`）:**
| ボタン | 位置 | 動作 | 備考 |
|---|---|---|---|
| 削除 | 左（編集時のみ） | `handleDelete` → staff 時は `StaffImpactDialog` / 他は `ConfirmDialog` | Trash2 アイコン、赤文字 |
| キャンセル | 右 | `handleCloseEdit` → リストモードへ | outline |
| 保存 | 右 | `handleSave` | Save アイコン、`[S] PrimaryButton` |

**スタッフ固有ダイアログ（`[S] StaffImpactDialog`）:**
- ステータス変更・名称変更・削除時に影響範囲を確認
- `staffName`, `action`（rename / deactivate / delete）, `usage`（使用箇所数）を表示

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
| `ConsultationSection` | `[C]` | 診察マスタ固有: 適用区分・標準診察時間 |
| `ProcedureSection` | `[C]` | 処置マスタ固有: 所要時間・麻酔要否 |
| `CheckupSection` | `[C]` | 定期健診マスタ固有: 推奨受診間隔・対象年齢 |
| `ServiceTypeSection` | `[C]` | 予約区分マスタ固有: 表示カラーピッカー・プレビュー |
| `SectionWrapper` | `[C]` | セクション共通ラッパー（`SectionPropertyRow` = `NotionPropertyRow` re-export） |
| `NotionPropertyRow` | `[S]` | Notion風プロパティ行（MasterItemEditForm・全セクション共通） |
| `PageLayout` | `[S]` | ページレイアウト |
| `NotionFilter` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `PrimaryButton` | `[S]` | 新規登録・保存ボタン |
| `StatusBadge` | `[S]` | ステータスバッジ |
| (行クリック) | — | 行クリックで編集モードへ遷移 |
| `StaffImpactDialog` | `[S][M]` | スタッフ変更影響確認ダイアログ |
| `ConfirmDialog` | `[S][M]` | 削除確認 |
| `MasterLink` | `[S]` | マスタ設定リンク |
| `useMasterItemEditor` | `[H]` | CRUD操作フック |
| `useMasterItems` | `[H]` | マスタデータ取得 |

**データ型:** `MasterItem`, `MasterFormData`, `MasterCategory`, `MasterSectionProps`, `CreateMasterItemDTO`, `UpdateMasterItemDTO`, `CategoryConfig` + 各カテゴリ固有型（`ConsultationTime`, `Anesthesia`, `CheckupTargetAge`, `VaccineSpecies`, `DosageForm`, `MedicineUnit`, `StaffRole`, `CageType`, `CageSize`, `CoverageRate`, `TargetSize`, `Combinable`, `BodySize`, `BillingUnit`）

**ユーザー操作:**
- リストモード: 検索、行クリックで編集モードへ、新規登録ボタン
- 編集モード: 共通フィールド + カテゴリ固有フィールド入力
- ステータス切替（有効/無効ラジオ）
- 保存 / キャンセル / 削除
- staff カテゴリ: 名称変更・ステータス変更・削除時に影響確認ダイアログ
- diagnosis_name カテゴリ: 親カテゴリ（diagnosis_category マスタ）からの選択

**アクセシビリティ:**
- 名称フィールド: `aria-invalid` + `aria-describedby` → `FormFieldError`（`role="alert"`）接続
- D&D操作: `useAnnounce` でキーボード・マウス全操作結果をスクリーンリーダーに通知
- `MasterSelectModal` 内アイテムリスト: `<button>` 要素でキーボード操作対応

---

## 12. 在庫管理

### 12.1 在庫一覧

| 項目 | 内容 |
|------|------|
| **ルート** | `/inventory` |
| **コンポーネント** | `[R] InventoryList` |
| **目的** | 在庫品目の一覧表示・検索・フィルタリング |

**画面構成:**
- ヘッダー: タイトル「在庫管理」（Package アイコン）+ 更新ボタン（RefreshCw アイコン、outline���+ 在庫登録ボタン（`[S] PrimaryButton`）
- 検索バー（`[S] NotionFilter`）: 「品名、保管場所、仕入先で検索...」
- カテゴリフィルタ（`Select`）: 全カテゴリー / 医薬品 / 消耗品 / フード / その他
- 状態フィルタ（`Select`）: 全ての状態 / 在庫あり / 残りわずか / 在庫切れ
- データテーブル（`[S] DataTable`）
- ページネーション（`[S] Pagination`、20件/ページ）
- 削除確認ダイアログ（`[S] ConfirmDialog`）
- ローディング: `[S] LoadingSkeleton`（variant="table", rows=8, columns=7）

**テーブル列:**
| 列 | className / align | 表示内容 |
|---|---|---|
| 品名 | - | `item.name`（`SortableHeader` 対応） |
| カテゴリー | `w-[100px]` | `getInventoryCategoryLabel(item.category)`（`SortableHeader` 対応） |
| 在庫数 | `w-[90px]`, align:right | `item.quantity`（等幅フォント、発注点以下は赤文字）（`SortableHeader` 対応、数値 `comparator`） |
| 単位 | `w-[80px]` | `item.unit` |
| 発注点 | `w-[80px]`, align:right | `item.minStockLevel`（等幅フォント） |
| 状態 | `w-[110px]`, align:center | `StatusBadge`（`getInventoryStatusColor` / `getInventoryStatusLabel`）（`SortableHeader` 対応） |
| 保管場所 | `w-[120px]` | `item.location`（`SortableHeader` 対応） |
| 期限 | `w-[110px]` | `item.expiryDate`（等幅フォント、yyyy/MM/dd）（`SortableHeader` 対応） |
| 仕入先 | `w-[130px]` | `item.supplier`（truncate）（`SortableHeader` 対応） |
| 操作 | `w-[50px]`, align:right | `RowActionDropdown`（編集 / 削除） |

**行アクション:**
| アクション | アイコン | 動作 |
|---|---|---|
| 編集 | Edit | `/inventory/{id}` へ遷移 |
| 削除 | Trash2 | `ConfirmDialog` → 削除、構造化トースト |

**使用コンポーネ���ト:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `InventoryList` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `NotionFilter` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `SortableHeader` | `[S]` | ソート可能ヘッダー（9列中7列） |
| `StatusBadge` | `[S]` | ステータスバッジ |
| `RowActionDropdown` | `[S]` | 行アクションメニュー |
| `ConfirmDialog` | `[S][M]` | 削除確認 |
| `Pagination` | `[S]` | ページネーション |
| `LoadingSkeleton` | `[S]` | ローディング表示 |
| `useTableSort` | `[H]` | ソートロジック（数値 `comparator` 適用済み） |
| `usePagination` | `[H]` | ページネーション |

**データ型:** `InventoryItem`, `InventoryCategory`, `InventoryStatus`, `DataTableColumn`

**ユーザー操作:**
- テキスト検索（品名、保管場所、仕入先）
- カテゴリーフィルタ / 状態フィルタ
- 列ヘッダークリックでソート（3状態サイクル: 昇順→降順→なし）
- 行クリックで編集画面へ遷移
- 行ドロップダウンから編集/削除
- 更新ボタンでデータリフレッシュ
- 新規登録ボタンで作成画面へ

### 12.2 在庫登録/編集

| 項目 | 内容 |
|------|------|
| **ルート** | `/inventory/new`（新規）/ `/inventory/:id`（編集） |
| **コンポーネント** | `[R] InventoryForm` |
| **目的** | 在庫品目の追加・編集 |

**画面構成:**
- ヘッダー: タイトル「在庫登録」/「在庫編集」（Package アイコン）+ 戻るボタン → `/inventory`
- フォームカード（`STYLE.formCard`）
- ローディング: `[S] LoadingSkeleton`（variant="form", rows=5）
- フォーム離脱保護（`[S] NavigationBlocker`）

**フォーム項目:**
| フィールド | 入力部品 | グリッド | 備考 |
|---|---|---|---|
| 品名 | `Input` | full | 必須（`*`マーク）、placeholder: 例: アモキシシリン 250mg |
| カテゴリー | `Select`（`INVENTORY_CATEGORY_VALUES`） | 2cols-左 | 医薬品/消耗品/フード/その他 |
| 単位 | `Input` | 2cols-右 | 必須（`*`マーク）、placeholder: 例: 錠, 箱, 本 |
| 現在在庫数 | `Input`（type=number, min=0） | 2cols-左 | 負数入力防止（`Math.max(0, ...)` ガード） |
| 発注点 (アラート基準) | `Input`（type=number, min=0） | 2cols-右 | 負数入力防止（`Math.max(0, ...)` ガード） |
| 保管場所 | `Input` | 2cols-左 | placeholder: 例: 薬品棚 A-1 |
| 使用期限 | `NotionDatePicker` | 2cols-右 | |
| 仕入先 | `Input` | full | placeholder: 例: 動物薬品工業 |
| ステータス（自動判定） | 読み取り専用 | full | 編集時のみ表示、在庫数と発注点から自動計算 |

**バリデーション:**
- 品名: 空チェック → `FormFieldError`（`role="alert"`, `aria-describedby`）
- 単位: 空チェック → `FormFieldError`（`role="alert"`, `aria-describedby`）
- エラー時: 構造化トースト（`toast.error` + description）

**アクション（`pt-4 border-t`）:**
| ボタン | 位置 | 動作 | 備考 |
|---|---|---|---|
| 削除 | 左（編集時のみ） | `ConfirmDialog` → `deleteInventoryItem` → 一覧へ遷移 | Trash2 アイコン、赤色 ghost |
| キャンセル | 右 | `/inventory` へ遷移 | outline |
| 保存 / 更新 | 右 | `handleSubmit` → バリデーション → API → `markClean` → 遷移 | Save アイコン、送信中は Loader2 スピナー |

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `InventoryForm` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `PrimaryButton` | `[S]` | 保存ボタン |
| `NavigationBlocker` | `[S]` | フォーム離脱保護 |
| `LoadingSkeleton` | `[S]` | ローディング表示 |
| `NotionDatePicker` | `[S]` | 日付ピッカー |
| `FormFieldError` | `[S]` | フィールドエラー表示 |
| `ConfirmDialog` | `[S][M]` | 削除確認 |
| `useUnsavedChanges` | `[H]` | 未保存検知 |

**データ型:** `InventoryItem`, `InventoryCategory`, `InventoryStatus`
**カテゴリ:** 医薬品 / 消耗品 / フード / その他
**ステータス:** 在庫あり / 残りわずか / 在庫切れ
**特記:** カルテ保存時の在庫消費連動 (`consumeStock`) は `useInventory` フック経由で実装済み。

**ユーザー操作:**
- 各フィールドの入力（全変更で `markDirty` 自動呼出）
- カテゴリー選択（Select）
- 使用期限選択（NotionDatePicker）
- 保存→バリデーション→構造化トースト→一覧へ遷移
- 削除→確認ダイアログ→構造化トースト→一覧へ遷移（編集時のみ）
- キャンセルで一覧へ戻る
- 未保存離脱時の保護ダイアログ（SPA内ナビゲーション + ブラウザ閉じ/リロード）

---

## 13. シフト管理

### 13.1 シフト管理カレンダー

| 項目 | 内容 |
|------|------|
| **ルート** | `/shifts` |
| **コンポーネント** | `[R] ShiftCalendar` |
| **目的** | スタッフの勤務シフトを週間・月間カレンダーで管理する |

**画面構成:**
- ヘッダー: タイトル「シフト管理」（CalendarDays アイコン）
- ツールバー:
  - ビュー切替トグル（週表示 / 月表示）
  - ロールフィルタ（全員 / 医師 / スタッフ / トリマー）
  - ナビゲーション（前週/次週 or 前月/次月）+ 今日ボタン
  - 現在の期間ラベル
- **週表示** (`[C] ShiftWeekView`):
  - スタッフ×曜日の7列グリッド
  - 各セルにシフトタイプ別色分けバッジ（`[C] ShiftCell`）
  - セルクリックで `[C] ShiftEditPopover` 表示（シフトタイプ選択・時間入力・メモ）
  - 行末に週計労働時間表示（40時間超過で警告色）
- **月表示** (`[C] ShiftMonthView`):
  - カレンダーグリッドに日ごとの勤務人数サマリー
  - シフトタイプ別の内訳表示
- 凡例バー（`[C] ShiftLegend`）: 全シフトタイプの色とラベル

**コンポーネント構成:**

| コンポーネント | 種別 | 役割 |
|---|---|---|
| `ShiftCalendar` | `[R]` | メインページ |
| `ShiftWeekView` | `[C]` | 週間グリッドビュー |
| `ShiftMonthView` | `[C]` | 月間カレンダービュー |
| `ShiftCell` | `[C]` | シフトセル表示 |
| `ShiftEditPopover` | `[C]` | シフト編集Popover |
| `ShiftLegend` | `[C]` | 凡例コンポーネント |
| `PageLayout` | `[S]` | ページレイアウト |
| `useShiftManagement` | `[H]` | 状態管理（ビュー切替・ナビゲーション・CRUD・労働時間計算） |

**データ型:** `ShiftEntry`, `ShiftType`, `ShiftView`, `ShiftStaffInfo`, `DayShiftSummary`
**シフトタイプ:** 通常勤務(full) / 午前のみ(morning) / 午後のみ(afternoon) / 休み(off) / 有給(paid_leave)

**ユーザー操作:**
- ビュー切替（週表示 ↔ 月表示）
- ロールフィルタでスタッフを絞り込み
- 前/次の期間へナビゲーション、今日ボタンで現在週/月へ復帰
- セルクリック → Popover でシフトタイプ・時間・メモを編集 ��� 保存でトースト通知
- セルクリック → Popover で「削除」ボタン → シフト削除

**アクセシビリティ:**
- `ShiftEditPopover`: `useFocusTrap` 統合（Escape/Tab循環/フォーカス復帰）、`role="dialog"` + `aria-modal="true"`
- ツールバー: `role="group"` + `aria-label`、トグルボタンに `aria-pressed`
- ナビゲーション: `size-10`（40px）+ `after:-inset-2` ヒットエリア拡張
- シフトセル: `role="button"` + `tabIndex={0}`、Enter/Space キーボードハンドラ

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
- 検索ロジックは `usePetSearch` 共有フック経由

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

Inventory(/inventory) ──→ InventoryForm(/inventory/new, /inventory/:id)

Shifts(/shifts) ──→ 週間/月間ビュー切替・シフト編集Popover

Settings(/settings) ──→ ClinicSettings / Settings({category})
```

---

## 15. ログイン画面（AUTH.md Phase 1 実装済み）

### `/login` — ログイン

| 項目 | 内容 |
|------|------|
| **ルート** | `/login` |
| **ページコンポーネント** | `Login` |
| **feature** | `auth` |
| **アクセス制御** | 公開ルート（未認証ユーザーのみアクセス可、認証済みの場合は `/` にリダイレクト） |

#### 画面構成

- **レイアウト**: サイドバーなし・フルスクリーン中央配置
- **ロゴ**: クリニックロゴ（`ClinicInfo.logoUrl` または デフォルトアイコン）
- **フォーム要素**:
  - メールアドレス入力（`type="email"`、バリデーション: 必須・メール形式）
  - パスワード入力（`type="password"`、バリデーション: 必須・最小8文字）
  - ログインボタン（`NotionButton` variant="primary"）
  - エラーメッセージ表示（認証失敗時）
- **デザイン**: Notion風カラーパレット準拠、WCAG AA アクセシビリティ対応
- **状態管理**: `useAuth` フック（`features/auth/hooks/useAuth.ts` に実装済み）

#### データフロー

```
Login
  └── onSubmit(email, password)
        └── authApi.login(email, password)
              ├── 成功 → navigate('/') + AuthContext更新
              └── 失敗 → エラーメッセージ表示
```

> **実装状況**: AUTH.md Phase 1 で実装済み。`LoginForm`（Notion風WCAG AA準拠）、6デモアカウント、`AuthProvider`+`useAuth`によるセッション永続化、`AuthGuard`による未認証時リダイレクトが動作中。

---

## 全ルート一覧��42ルート + Fallback）

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
| 30 | `/estimates` | 見積書一覧 | `EstimateList` |
| 31 | `/estimates/new` | 見積書新規作成 | `EstimateForm` |
| 32 | `/estimates/:id` | 見積書詳細 | `EstimateDetail` |
| 33 | `/estimates/:id/edit` | 見積書編集 | `EstimateForm` |
| 34 | `/settings` | マスタ設定トップ | `MasterSettingsIndex` |
| 35 | `/settings/clinic` | 病院情報設定 | `ClinicSettings` |
| 36 | `/settings/treatment-items` | 診療項目マスタ (5カテゴリ統合) | `TreatmentItemsSettings` |
| 37 | `/settings/diagnosis` | 診断マスタ (2カテゴリ統合) | `DiagnosisSettings` |
| 38 | `/settings/trimming` | トリミングマスタ (2カテゴリ統合) | `TrimmingSettings` |
| 39 | `/settings/service-type` | 予約区分マスタ | `Settings` |
| 40 | `/settings/medicine` | 薬剤マスタ | `Settings` |
| 41 | `/settings/staff` | スタッフマスタ | `Settings` |
| 42 | `/settings/insurance` | 保険マスタ | `Settings` |
| 43 | `/settings/hospitalization` | 入院マスタ | `Settings` |
| 44 | `/settings/cage` | ケージマスタ | `Settings` |
| 45 | `/dev/tests` | フォーマットテスト (開発用) | `FormatTestRunner` |
| 46 | `/login` | ログイン | `Login` |
| — | `*` | 404ページ | インライン |

> **注**: マスタ設定ルートは3統合ページ（`TreatmentItemsSettings`/`DiagnosisSettings`/`TrimmingSettings`）+ 6個別ページ（`Settings` コンポーネント, category prop切替）= 9ルートで15カテゴリをカバー。`/login` は認証実装（AUTH.md Phase 1〜3 完了済み）に伴い追加された公開ルート。`/dev/tests` は開発用フォーマットテストランナー。