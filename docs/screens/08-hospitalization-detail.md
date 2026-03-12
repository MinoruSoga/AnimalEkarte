# 入院詳細・デイリーカルテ 仕様書

## 概要
- **画面の目的**: 入院患者のケアプラン管理、デイリーログ（バイタル・ケアログ・スタッフメモ）の記録、入院サマリーの印刷
- **URLパターン**: `/hospitalization/:id`
- **コンポーネント**: `[R] HospitalizationDetail`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト（レスポンシブ: デスクトップ/モバイル分離）

### デスクトップレイアウト（`HospitalizationDesktopLayout`）
```
┌──────────────────────────────────────────────────────┐
│ ← 戻る  入院詳細・カルテ  [アクションバー]            │
├───────────────────────┬──────────────────────────────┤
│ 左カラム              │ 右カラム                      │
│ - 患者ヘッダー        │ - デイリーレコードセクション   │
│ - HospitalizationDetailActions │                    │
│ - ケアプランセクション │                              │
└───────────────────────┴──────────────────────────────┘
```

### モバイルレイアウト（`HospitalizationMobileLayout`）
```
シングルカラム: 患者ヘッダー → ケアプラン → デイリーレコード
```

## ヘッダーアクション（`HospitalizationDetailActions`）

| ボタン | 条件 | 動作 |
|--------|------|------|
| 印刷 | 常時 | `PrintPreviewDialog`を開く（入院サマリープレビュー） |
| 編集 | 常時 | `/hospitalization/:id/edit`へ遷移 |
| 会計を確認 | `linkedAccountingId` あり | 既存会計詳細へ遷移 |
| 会計へ進む | 退院済かつ`linkedAccountingId`なし | 治療プランを引き継いで会計新規作成へ遷移 |
| 退院 | ステータスが入院中 | `ConfirmDialog`を開く |

## 患者情報ヘッダー（`HospitalizationPatientHeader`）

| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| ペット名 | string | ペット名 | `pets.name` |
| 種/品種 | string | 動物種・品種 | `pets.species`, `pets.breed` |
| 飼主名 | string | 飼い主氏名（クリックで`OwnerQuickViewModal`） | `owners.name` |
| 入院No | string | 入院番号 | `hospitalizations.hospitalization_no` |
| 入院種別 | enum | 入院/ホテル | `hospitalizations.type` |
| ケージ名 | string | 割り当てケージ | `cages.name` |
| 入院開始日 | date | 入院開始日 | `hospitalizations.start_date` |
| 退院予定日 | date | 退院予定日 | `hospitalizations.end_date` |
| ステータス | enum | 入院中/退院済/予約 | `hospitalizations.status` |

## ケアプランセクション（`CarePlan/`）

| コンポーネント | 説明 |
|---|---|
| `CarePlanPreviewPopover` | ケアプラン概要ポップオーバー（ステータストグル付き: active ↔ completed 即時切り替え） |
| `CarePlanSection` | ケアプラン一覧表示 |
| `CarePlanItemRow` | ケアプラン項目行 |
| `CarePlanDialog` | ケアプラン追加/編集ダイアログ |

### CarePlanDialog フォーム項目

| フィールド | 入力部品 | 備考 |
|---|---|---|
| マスタ引用 | 「マスタ検索」ボタン → `TreatmentSearchDialog` | 処置・検査・薬をマスタから検索して自動入力 |
| 種類 | `Select` | `CARE_PLAN_TYPE_VALUES`: 食事/投薬/処置・検査/処置・指示/持ち物・その他 |
| 名称 | `Input` | 例: ロイヤルカナン消化器サポート |
| マスタ連動情報 | Badge 表示（単価(税込)、カテゴリ） | マスタ選択時のみ表示 |
| 詳細・指示量 | `Input` | 例: 30g / 1錠 / 左前肢 |
| タイミング | トグルボタン | `PLAN_TIMING_VALUES`: 朝/昼/夜（複数選択可） |
| メモ・特記事項 | `Textarea` | 例: ふやかして与える |
| ステータス | `Select` | `CARE_PLAN_STATUS_VALUES` |

### CarePlanItem フィールド

| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| 種類 | enum | 食事/投薬/処置・検査/処置・指示/持ち物・その他 | `care_plan_items.type` |
| 名称 | string | ケアプラン項目名 | `care_plan_items.name` |
| 詳細・指示量 | string | 投薬量や処置の詳細 | `care_plan_items.description` |
| タイミング | array | 実施タイミング（朝/昼/夜等） | `care_plan_items.timing` |
| ステータス | enum | active/completed/discontinued | `care_plan_items.status` |
| 単価(税込) | decimal | 単価 | `care_plan_items.unit_price` |
| カテゴリ | string | マスタ連動カテゴリ | `care_plan_items.category` |
| マスタID | string | マスタ参照ID（medicine_id / procedure_id / hospitalization_plan_id のいずれか） | `care_plan_items.medicine_id` / `care_plan_items.procedure_id` / `care_plan_items.hospitalization_plan_id` |

## デイリーレコードセクション（`DailyRecord/`）

| コンポーネント | 説明 |
|---|---|
| `DailyRecordSection` | デイリーレコードメインセクション |
| `DateNavigation` | 日付ナビゲーション（前後の日付へ移動） |
| `TimingSection` | 朝/昼/夜のタスク区分表示 |
| `TaskCompleteDialog` | タスク完了ダイアログ |
| `VitalDialog` | バイタル入力ダイアログ |
| `LogDialog` | ケアログ入力ダイアログ |
| `SimpleNoteForm` | スタッフメモ入力 |
| `Timeline` | 実施記録の時系列表示 |

### VitalDialog フォーム項目

| フィールド | 入力部品 | 備考 |
|---|---|---|
| 記録時刻 | `Input`（type=time） | 現在時刻で初期化 |
| 体温 (℃) | `Input`（number, step=0.1） | |
| 体重 (kg) | `Input`（number, step=0.01） | |
| 心拍数 (/min) | `Input`（number） | |
| 呼吸数 (/min) | `Input`（number） | |
| メモ | `Textarea` | |

### LogDialog フォーム項目（入院ケアログ用）

ログ種別に応じてタイトル・説明・プレースホルダーが変化:
- `food`: 「食事記録」（完食、1/2など）
- `excretion`: 「排泄記録」（良便、軟便など）
- `medicine` / `other`: 「活動・メモ」（内容）

| フィールド | 入力部品 | 備考 |
|---|---|---|
| 記録時刻 | `Input`（type=time） | 現在時刻で初期化 |
| 内容・量 | `Input` | 種別依存のプレースホルダー |
| 詳細メモ | `Textarea` | |

### TaskCompleteDialog フォーム項目

| フィールド | 入力部品 | 備考 |
|---|---|---|
| タスク情報 | 読み取り専用表示 | タスク名 + 詳細（背景カード） |
| 実施時刻 | `Input`（type=time） | 現在時刻で初期化 |
| 実施メモ (任意) | `Textarea` | |

## 印刷機能

- ヘッダーアクションに印刷ボタン表示
- `PrintPreviewDialog`（`[S][M]`）でプレビュー表示
- 印刷コンテンツ: `HospitalizationSummaryDocument`（入院日数自動計算・1日あたり費用表示）
- `usePrint<HospDocumentType>`フックで印刷状態管理
- `window.print()`で印刷実行
- ドキュメント種別: `summary`（「入院サマリー」）

## 退院処理

- `ConfirmDialog`（タイトル「退院処理を行いますか？」）を表示
- 確認時: ステータスを「退院済」に変更、退院日を本日に設定
- チェックボックス「退院後、そのまま会計画面へ進む」（デフォルトON）
  - ON時: 「退院して会計へ」ボタン、退院後に会計新規作成へ遷移（治療プランを`AccountingItem[]`に変換して`state`経由で渡す）
  - OFF時: 「退院処理を実行」ボタン、一覧画面へ遷移

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 編集 | 「編集」ボタン | 入院編集フォームへ | `/hospitalization/:id/edit` |
| 退院処理 | 「退院」ボタン | ConfirmDialog → 退院ステータスへ変更 | 一覧 or 会計画面 |
| 会計へ進む | 「会計へ進む」ボタン | 治療プラン変換 → 会計新規作成へ | `/accounting/new?petId=xxx` |
| 会計を確認 | 「会計を確認」ボタン | 既存会計詳細へ | `/accounting/:id` |
| 印刷 | 印刷ボタン | `PrintPreviewDialog`表示 | ダイアログ |
| ケアプラン追加 | 「追加」ボタン | CarePlanDialog開く | ダイアログ |
| ケアプラン編集 | 行の編集ボタン | CarePlanDialog開く（編集モード） | ダイアログ |
| ケアプラン削除 | 行の削除ボタン | 確認後削除 | 同画面 |
| ケアプランステータストグル | CarePlanPreviewPopover | active ↔ completed 即時切替 | 同画面 |
| タスク完了 | タイミングセクションのチェック | TaskCompleteDialog開く | ダイアログ |
| バイタル記録 | 「バイタル追加」ボタン | VitalDialog開く | ダイアログ |
| ケアログ記録 | 「ケアログ追加」ボタン | LogDialog開く | ダイアログ |
| スタッフメモ追加 | SimpleNoteForm | テキスト入力して保存 | 同画面 |
| タイムライン確認 | - | 実施記録を時系列で参照 | 同画面 |
| 飼主情報確認 | 飼主名クリック | OwnerQuickViewModal表示 | モーダル |
| 戻る | 「戻る」ボタン | `location.state.from`または入院一覧へ | `/hospitalization` |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| 入院一覧 | `/hospitalization/:id` | 行/カードクリック |
| 「編集」ボタン | `/hospitalization/:id/edit` | ボタンクリック |
| 退院（会計へ進む） | `/accounting/new?petId=xxx` | チェックON + 退院確認 |
| 退院（一覧へ） | `/hospitalization` | チェックOFF + 退院確認 |

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/hospitalizations/:id` | 入院詳細取得 | 未実装 |
| POST | `/api/v1/hospitalizations/:id/discharge` | 退院処理 | 未実装 |
| GET | `/api/v1/hospitalizations/:id/care-plans` | ケアプラン一覧 | 未実装 |
| POST | `/api/v1/hospitalizations/:id/care-plans` | ケアプラン追加 | 未実装 |
| PUT | `/api/v1/care-plans/:id` | ケアプラン更新 | 未実装 |
| DELETE | `/api/v1/care-plans/:id` | ケアプラン削除 | 未実装 |
| GET | `/api/v1/hospitalizations/:id/daily-records` | デイリーレコード一覧 | 未実装 |
| POST | `/api/v1/daily-records/:id/vitals` | バイタル追加 | 未実装 |
| POST | `/api/v1/daily-records/:id/care-logs` | ケアログ追加 | 未実装 |
| POST | `/api/v1/daily-records/:id/notes` | スタッフメモ追加 | 未実装 |

## データ型

`Hospitalization`, `CarePlanItem`, `DailyRecord`, `VitalRecord`, `CareLogRecord`, `StaffNoteRecord`, `Task`, `TimelineItem`, `HospDocumentType`

## 実装状況
- フロントエンド(ui-sample): 実装済（LocalStorageによるデータ永続化）
- バックエンドAPI: 未実装
