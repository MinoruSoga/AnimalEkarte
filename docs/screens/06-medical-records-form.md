# カルテ入力/編集 仕様書

## 概要
- **画面の目的**: 9タブ構成の診療記録（電子カルテ）を作成・編集する
- **URLパターン**:
  - 新規作成: `/medical-records/new?petId=xxx`
  - 編集: `/medical-records/:id`
- **アクセス権限**: 認証済ユーザー全員（要 medical 権限）

## 画面レイアウト
```
┌──────────────────────────────────────────────────────────────┐
│ ← 戻る  [削除ボタン] [印刷ボタン(確定済のみ)] [会計確認/進む]│
├──────────────────────────────────────────────────────────────┤
│ [スティッキーヘッダー]                                        │
│ PatientInfoCard: ペット名 / 種 / 飼主名 /                    │
│   [診療区分 MasterSelectModal] [担当医 MasterSelectModal]    │
│   [バイタル入力ボタン]                                        │
├──────────────────────────────────────────────────────────────┤
│ タブバー:                                                     │
│ [問診][診察/治療プラン][治療][予防接種][定期健診][検査]      │
│ [画像][見積書][会計(医師確認)]  ← 9タブ（旧: 6タブから拡張）│
├──────────────────────────────────────────────────────────────┤
│ 各タブのコンテンツエリア                                      │
└──────────────────────────────────────────────────────────────┘
```

## タブ構成（9タブ）

| # | タブ名 | コンポーネント | 説明 |
|---|--------|--------------|------|
| 1 | **問診** | `MedicalRecordInterview` | 主訴（Markdown）、治療方針、カルテ履歴 |
| 2 | **診察/治療プラン** | `MedicalRecordDiagnosisPlan` | バイタル・診断登録・治療プランテーブル |
| 3 | **治療** | `MedicalRecordTreatment` | 処置完了記録（プラン→済の移動管理） |
| 4 | **予防接種** | `MedicalRecordVaccination` | 予防接種フォーム + 接種履歴一覧 |
| 5 | **定期健診** | `MedicalRecordCheckup` | 健診プラン・実施済み管理 + 健診フォーム + 履歴 |
| 6 | **検査** | `MedicalRecordExamination` | 検査オーダーフォーム + 結果履歴 |
| 7 | **画像** | `MedicalRecordImage` | 画像アップロード + フィルタ付きギャラリー |
| 8 | **見積書** | `MedicalRecordEstimate` | 診療内容に基づく概算見積 |
| 9 | **会計(医師確認)** | `MedicalRecordBillCheck` | 算定チェック・確認・会計連携 |

> タブは遅延マウント: 一度表示したタブはアンマウントせず状態を保持する。

## Tab 1: 問診（`MedicalRecordInterview`）

**レイアウト:** 3カラム（lg:12グリッド = 3+4+5）

| カラム | コンポーネント | 説明 |
|--------|-------------|------|
| 左 | `InterviewChiefComplaint` | 主訴入力（Markdown テキストエリア）。`chiefComplaintCategoryId`（カテゴリマスタ）選択。テンプレート挿入ボタン付き。 |
| 中 | `InterviewTreatmentPolicy` | 治療方針（Markdown テキストエリア） |
| 右 | `InterviewHistory` | カルテ履歴リスト（日付・担当医・診療種別バッジ・タイトル・内容） |

## Tab 2: 診察/治療プラン（`MedicalRecordDiagnosisPlan`）

**レイアウト:** 診断ヘッダー（3カラム）+ 治療プランテーブル

| コンポーネント | 内容 |
|-------------|------|
| `DiagnosisHeaderChiefComplaint` | 主訴の読み取り専用表示 |
| `DiagnosisHeaderPhysicalExam` | 身体所見 Markdown エリア（`assessment`） |
| `DiagnosisHeaderDiagnosis` | 診断登録フォーム（診断詳細 Markdown・診断1/2カテゴリ Select・診断1/2名称 Select） |
| `TreatmentTable` | 治療プランテーブル（マスタ検索ダイアログ連携） |
| `TreatmentDetailedSummary` | 集計（小計・税・合計・割引率・値引額） |

## Tab 3: 治療（`MedicalRecordTreatment`）

**レイアウト:** 治療テーブル（`TreatmentTable`）単一構成

| フィールド | 入力部品 | 備考 |
|-----------|----------|------|
| 項目名 | `Input` | `content` |
| メモ | `Input` | `memo` |
| 保険 | `Checkbox` | `insurance` |
| 単価 | `Input(number)` | `unitPrice` |
| 数量 | `Input(number)` | `quantity` |
| 値引(率) | `Input(number)` | `discountRate` |
| 値引(額) | `Input(number)` | `discountAmount` |
| ステータス | `Select` | 実施中 / 完了 / 中止 |

## Tab 4: 予防接種（`MedicalRecordVaccination`）

`VaccinationForm` フォーム項目:

| フィールド | 入力部品 | 備考 |
|----------|---------|------|
| 予防接種名 | `Select` | `esophagitis` / `rabies` / `distemper` 等 |
| 予防接種日 | `NotionDatePicker` | |
| 補助説明 | `Input` | |
| LOT1〜LOT4 | `Input` × 4 | 4つのLOT番号入力 |
| 次回予防接種予定設定 | `RadioGroup` | 3週間後 / 4週間後 / 1年後 / 以外 |
| 次回予定日 | `NotionDatePicker` | |
| 備考 | `Textarea` | |

## Tab 5: 定期健診（`MedicalRecordCheckup`）

**レイアウト:** インライン編集テーブル（`CheckupsTab`）

| フィールド | 入力部品 | 備考 |
|-----------|----------|------|
| 日付 | `Input(date)` | |
| 健診種別 | `Select` | `checkup` マスタ連動 |
| 次回予定日 | `Input(date)` | |
| 結果 | `Input(text)` | |

## Tab 6: 検査（`MedicalRecordExamination`）

**レイアウト:** フィルタバー + 検査結果グループ（`ExaminationGroup`）

| 表示項目 | 説明 |
|---------|------|
| 実施日時 | 検査が行われた日時 |
| 検査機器 | 使用された機器名 |
| 検査項目テーブル| 項目名、測定値、単位、基準値、判定（H/L） |

## Tab 7: 画像（`MedicalRecordImage`）

**レイアウト:** フィルタバー + 画像ギャラリー（`ImageGalleryGroup`）

| 表示項目 | 説明 |
|---------|------|
| 撮影日時 | 画像が保存された日時 |
| 画像名 | アップロードされたファイル名 |
| 画像プレビュー| サムネイル表示、拡大表示対応 |

## Tab 8: 見積書（`MedicalRecordEstimate`）

**レイアウト:** 治療テーブル + 集計サマリー

※表示項目は Tab 3 の治療テーブルと共通。ステータス列は非表示。

## Tab 9: 会計(医師確認)（`MedicalRecordBillCheck`）

**レイアウト:** 治療テーブル + 詳細サマリー（全体割引対応）

| 項目 | 入力・表示 |
|------|-----------|
| 全体割引率 (%) | `Input(number)` |
| 全体値引額 (￥) | `Input(number)` |
| チェック完了 | ボタン（会計ステータスを更新） |

## VitalInputDialog（バイタル入力ダイアログ）

| フィールド | 入力部品 | 単位 |
|----------|---------|------|
| 体温 | `Input`（number, step=0.1）、Thermometer アイコン | ℃ |
| 心拍数 | `Input`（number）、Heart アイコン | /min |
| 呼吸数 | `Input`（number）、Wind アイコン | /min |
| 体重 | `Input`（number, step=0.01）、Weight アイコン | kg |
| メモ | `Textarea`、StickyNote アイコン | |

- 入力/履歴のタブ切替。履歴ではトレンドアイコン（↑↓→）表示

## TreatmentTable 列構成

| カラム | 幅 | 編集 |
|--------|-----|------|
| 種別 | `w-20` | 読み取り専用（マスタから自動設定） |
| 治療内容 | `min-w-[160px]` | `Input` |
| メモ | `min-w-[120px]` | `Input` |
| 保険 | `w-16` | `NotionCheckbox` |
| 単価(税込) | `w-24` | `Input` + `useNumericInput`、`formatCurrency` |
| 数量 | `w-16` | `Input` + `useNumericInput` |
| 割引(%) | `w-20` | `Input` + `useNumericInput` |
| 値引(￥) | `w-24` | `Input` + `useNumericInput`、`formatCurrency` |
| 小計 | `w-28` | 読み取り専用（`calcLineTotal` 自動計算） |
| 操作 | `w-12` | 削除ボタン（ホバー時のみ表示） |

**TreatmentType 種別:** 診察 / 処置 / 薬剤 / 検査 / 予防接種 / 定期健診

## UI コンポーネント

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `MedicalRecordForm` | `[R]` | メインページ |
| `PatientInfoCard` | `[S]` | 患者情報カード（スティッキーヘッダー） |
| `OwnerQuickViewModal` | `[S][M]` | 飼主情報クイックビュー |
| `MasterSelectModal` | `[S][M]` | 診療区分・担当医選択 |
| `VitalInputDialog` | `[C][M]` | バイタル入力ダイアログ |
| `TreatmentTable` | `[C]` | 治療項目テーブル（複数タブで共用） |
| `TreatmentSearchDialog` | `[C][M]` | マスタ検索ダイアログ |
| `TreatmentDetailedSummary` | `[C]` | 集計エリア |
| `TreatmentMoveConfirmDialog` | `[C][M]` | 治療移動確認（`AlertDialog` ベース） |
| `InterviewChiefComplaint` | `[C]` | 主訴入力（Markdown） |
| `InterviewHistory` | `[C]` | 問診履歴 |
| `DiagnosisHeader` | `[C]` | 診断ヘッダー（3カラム） |
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
| `useUnsavedChanges` | `[H]` | 未保存変更検知 |

## 機能詳細

### 1. 治療ステータスのライフサイクル
- **プランから済みへ**: 「治療プラン」タブで実施した項目にチェックを入れると、確認ダイアログ（`TreatmentMoveConfirmDialog`）を経て「治療（実施済み）」タブへ自動移動する。
- **差し戻し**: 誤って済みにしてしまった項目は、治療済みテーブルからプランへ差し戻すことが可能。

### 2. 精緻な自動集計（`useMedicalRecordCalculation`）
- **リアルタイム計算**: 単価、数量、値引率（または値引額）、保険適用の有無を即座に計算し、集計サマリー（小計、消費税、合計）を更新する。
- **値引ロジック**: 飼主マスタに設定されたデフォルト値引率を初期値として読み込みつつ、項目ごとに個別の調整が可能。

### 3. 会計へのデータ変換と連携
- **ワンクリック会計**: 「会計へ進む」ボタンにより、カルテ内の「実施済み」明細を `AccountingItem[]` 形式に変換し、ペット情報と共に会計新規作成画面へ引き継ぐ。
- **重複防止**: 既に会計が作成されている場合は「会計を確認」ボタンに切り替わり、二重請求を防止する。

### 4. 確定処理と読み取り専用
- **ステータス「確定済」**: カルテを確定すると、以後内容の編集は不可となり、印刷（診断書・処方箋等）のみが可能となる。

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 保存 | 「保存」ボタン | バリデーション → 保存（作成中ステータス） | 同画面 |
| 削除 | 「削除」ボタン | `ConfirmDialog` 後に削除 | `/medical-records` |
| タブ切替 | タブクリック | コンテンツ切替（遅延マウント・状態保持） | 同画面 |
| 診療区分選択 | PatientInfoCard ボタン | `MasterSelectModal` 開く | モーダル表示 |
| 担当医選択 | PatientInfoCard ボタン | `MasterSelectModal` 開く | モーダル表示 |
| バイタル入力 | PatientInfoCard「バイタル入力」ボタン | `VitalInputDialog` 開く | ダイアログ表示 |
| 治療項目追加 | 「+ 治療項目を追加」ボタン | `TreatmentSearchDialog` 開く | ダイアログ表示 |
| 治療完了 | 治療タブ「済」チェックボックス | `TreatmentMoveConfirmDialog` → 治療済みへ移動 | 同画面 |
| 会計へ進む | 会計タブ「会計へ進む」ボタン | 明細を `AccountingItem[]` に変換→会計画面へ | `/accounting/new?petId=xxx` |
| 会計を確認 | 会計タブ「会計を確認」ボタン | 既存会計詳細へ遷移 | `/accounting/:accountingId` |
| 印刷 | 「印刷」ボタン（確定済のみ） | `PrintPreviewDialog` 表示 → `window.print()` | ダイアログ表示 |
| 戻る | 「戻る」ボタン | カルテ一覧へ | `/medical-records` |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| ペット選択 | `/medical-records/new?petId=xxx` | ペット選択後 |
| カルテ一覧 | `/medical-records/:id` | 行クリック |
| 会計タブ「会計へ進む」 | `/accounting/new?petId=xxx` | 新規会計作成 |
| 会計タブ「会計を確認」 | `/accounting/:accountingId` | 既存会計参照 |

## バリデーション
- **担当医**: 未選択時にトースト警告＋スタッフボタンにエラー状態表示（`aria-describedby` → `FormFieldError`）
- **確定済みカルテ**: 読み取り専用表示（編集不可）、印刷ボタンが表示される
- `NavigationBlocker` により未保存状態での離脱時に確認ダイアログが表示される

## 状態管理

| 状態 | 型 | 説明 |
|------|-----|------|
| `activeTab` | `string` | アクティブタブ |
| `serviceType` | `string` | 診療区分 |
| `staffName` | `string` | 担当医名 |
| `chiefComplaint` | `string` | 主訴（Markdown テキスト） |
| `treatmentPlanItems` | `TreatmentItem[]` | 治療プランテーブルの項目 |
| `treatmentCompletedItems` | `TreatmentItem[]` | 治療済みテーブルの項目 |
| `examinationResults` | `ExaminationResultGroup[]` | 検査結果 |
| `vaccinationFormData` | `VaccinationFormData` | 予防接種フォームデータ |
| `vaccinationHistory` | `VaccinationHistoryItem[]` | 予防接種履歴 |
| `vitalHistory` | `VitalEntry[]` | バイタル履歴 |
| `checkupFormData` | `CheckupFormData` | 健診フォームデータ |
| `isDirty` | `boolean` | 未保存変更フラグ（`useUnsavedChanges`） |

## 印刷機能
- ヘッダーアクションに印刷ボタン表示（確定済み時のみ）
- `usePrint<MrDocumentType>` + `MR_DOCUMENT_TYPE_LABELS`（処方箋/診断書）で動的タイトル管理
- `PrintPreviewDialog` でプレビュー表示、`window.print()` で印刷実行
- 印刷エリア（`hidden print:block`、`data-print-area` 属性）に帳票を配置

## データ型
`MedicalRecord`, `TreatmentItem`, `TreatmentType`, `TreatmentStatus`, `VitalEntry`, `VitalFormData`, `DiagnosisFormData`, `ExaminationFormData`, `ExaminationResultGroup`, `VaccinationFormData`, `VaccinationHistoryItem`, `InterviewHistoryItem`, `MrDocumentType`, `CheckupFormData`

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/medical-records/:id` | カルテ詳細取得 | 実装済 |
| POST | `/api/v1/medical-records` | カルテ作成 | 実装済 |
| PATCH | `/api/v1/medical-records/:id` | カルテ更新 | 実装済 |
| DELETE | `/api/v1/medical-records/:id` | カルテ削除 | 実装済 |
| GET | `/api/v1/pets/:id` | 患者情報取得 | 実装済 |
| GET | `/api/v1/staffs` | スタッフマスタ取得 | 実装済 |

## 実装状況
- フロントエンド: 実装済（`features/medical-records/routes/MedicalRecordForm.tsx`）
- バックエンドAPI: 実装済（`handler/medical_record_handler.go` 他）

## 備考
- 旧仕様（SOAPS形式・6タブ）から大幅に変更。現在は9タブ構成で予防接種・検査・画像・会計確認が独立したタブとして存在
- 旧仕様の「処方タブ」は廃止され、薬剤はマスタ選択で治療テーブルに追加する方式に変更
- タブ順序: 問診→診察/治療プラン→治療→予防接種→**定期健診→検査**→画像→見積書→会計(医師確認)。定期健診が予防接種の直後・検査の前に位置する
- 治療テーブル（`TreatmentTable`）は診察/治療プラン・治療・見積書・会計の4タブで共用。表示カラムはコンテキストに応じて変化
- 旧仕様の `surgery_notes` カラム（手術・特記事項）は廃止。治療テーブルで管理
- `clinical_plans` テーブル（旧 `soap_notes`）が対応テーブル
