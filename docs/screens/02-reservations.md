# 予約管理 仕様書

## 概要
- **画面の目的**: 診療・トリミング・手術等の予約をカレンダー形式で一元管理する。
- **URLパターン**: `/reservations`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト
```
┌──────────────────────────────────────────────────────────────┐
│ 予約管理  [新規予約ボタン]                                    │
├──────────────────────────────────────────────────────────────┤
│ [← 今日 →]  [yyyy年 M月]  [担当医フィルタ Select]          │
│ [予約種別カラー凡例]              [月表示 / 週表示 Select]   │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  カレンダービュー（月表示 or 週表示）                         │
│  - 週表示: タイムライン形式（時間軸 × 曜日、15分スナップ）   │
│  - 月表示: グリッド形式（日付 × 予約一覧、最大4件+他N件）   │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

## ツールバー項目

| 項目 | 説明 |
|-----|------|
| 日付ナビゲーション | 前/次（月単位 or 週単位）、今日ボタン |
| 年月表示 | `yyyy年 M月` 形式 |
| 予約種別カラー凡例 | 動的マスタ連動（`ServiceTypeMaster.color`）。診療=#dbeafe(bg)/#5b8def(text) / 検診=#dcfce7/#16a34a / 手術=#ffe2e2/#f87171 / ワクチン=#f3e8ff/#a855f7 / 入院=#cefafe/#0891b2 / トリミング=#ffedd4/#f97316 |
| 担当医フィルタ | Stethoscope アイコン付き Select。「すべての医師」＋全予約データから動的抽出した医師名一覧 |
| 表示切替 | 月表示 / 週表示 Select（`CalendarView`） |

## 表示項目（MonthView 予約カード）

| フィールド名 | 説明 |
|------------|------|
| 時刻 | `H:mm` 形式（太字） |
| 初診/再診 | visitType バッジ（初=赤背景 / 再=青背景） |
| ペット名 | `petName` |
| 飼い主名 / 担当医 | `ownerName` + ` / ` + `doctor`（2行目、低opacity） |
| 背景色 | 予約種別に対応したパステルカラー |
| 件数制限 | 1日最大4件表示・超過分は「他 N 件」表示 |

## 表示項目（WeekView 予約カード）

| フィールド名 | 説明 |
|------------|------|
| 時刻 | `H:mm` 形式（太字） |
| 初診バッジ | visitType が first の場合のみ「初」バッジ（赤背景） |
| ペット名 | `petName`（太字） |
| 飼い主名 | 高さ36px超で表示（低opacity） |
| 担当医 | 高さ52px超で表示（低opacity） |
| 予約種別名 | 高さ68px超で表示 |
| ステータスドット | checked_in=blue / in_consultation=purple / accounting=orange / completed=gray / cancelled=red（右上に丸ドット） |
| 現在時刻線 | 当日列に赤い水平線で現在時刻を表示 |
| 重複処理 | 同時間帯の予約は横並び分割表示 |

## ReservationFormModal ウィザード構成

- **Step 1: ペット検索・選択**（StepIndicator 1=ペット選択 / 2=予約情報）
  - `PatientSearch`（飼主名・ペット名テキスト検索）
  - `PatientSelectionTable`（検索結果テーブル、行クリックでペット選択）
  - 選択済みペットは `SelectedPetChip`（PawPrint アイコン付き、種バッジ、×ボタン）で表示
  - 編集モード時はStep 1をスキップ
- **Step 2: 予約情報入力**（`ReservationFormFields`）

## ReservationFormFields フォーム項目

| フィールド | 入力部品 | 備考 |
|----------|---------|------|
| 日付 | `Popover` + `Calendar` | 日付選択 |
| 時間帯（開始） | `Select`（30分刻み 0:00〜23:30） | Clock アイコン付き |
| 時間帯（終了） | `Select`（30分刻み 0:00〜23:30） | ArrowRight で接続 |
| 予約区分 | `Select`（serviceType マスタ連動） | `MasterLink` 付き |
| 初診/再診 | `RadioGroup`（first / revisit） | カスタムラベルUI（カラードット付き） |
| 担当者 | `Select`（staff マスタ連動、active のみ） | `MasterLink` 付き |
| メモ | `Textarea` | |

## 表示項目（ReservationDetailModal）

| セクション | 項目 |
|-----------|------|
| アクセントバー | visitType に応じた色帯（初診=赤 / 再診=青） |
| ヘッダー | 初診/再診バッジ（丸ドット付き）、予約種別名 |
| ステータスセレクター | 現在ステータスの色帯＋6段階ドロップダウン（予約確定/受付済/診療中/会計待ち/完了/キャンセル、各色ドット付き） |
| 日時カード | 日付（yyyy年 M月 d日 (E)）、時間帯（H:mm – H:mm）、Calendar/Clock アイコン |
| 患者情報 | ペット名（太字）、飼い主名 |
| 診療詳細 | 担当医 + 指名バッジ（amber 背景）、予約区分（Tag アイコン付き） |
| メモ | notes（amber 背景カード、FileText アイコン付き、条件表示） |
| フッターアクション | 削除ボタン（ゴミ箱）、編集ボタン、種別別レコード作成ボタン（カルテ作成 / トリミング記録作成 / 入院・ホテル登録） |

## UI コンポーネント

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `ReservationManagement` | `[R]` | メインページ（UI層のみ） |
| `useReservationManagement` | `[H]` | 予約CRUD・モーダル・バリデーションロジック |
| `MonthView` | `[C]` | 月表示カレンダー |
| `WeekView` | `[C]` | 週表示カレンダー（`motion/react` アニメーション、D&D） |
| `ReservationFormModal` | `[C][M]` | 予約作成/編集モーダル（2ステップウィザード） |
| `ReservationFormFields` | `[C]` | フォームフィールド群 |
| `ReservationDetailModal` | `[C][M]` | 予約詳細モーダル |
| `PatientSearch` | `[C]` | 患者検索コンポーネント |
| `PatientSelectionTable` | `[C]` | 患者選択テーブル |
| `ConfirmDialog` | `[S][M]` | キャンセル確認 |

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 新規予約作成 | 「新規予約」ボタン | `ReservationFormModal` を開く | モーダル表示 |
| タイムスロットクリック | 週表示の空き時間クリック | 選択時間で `ReservationFormModal` を開く | モーダル表示 |
| 予約詳細表示 | 予約カードクリック | `ReservationDetailModal` を開く | モーダル表示 |
| 予約編集 | 詳細モーダル「編集」ボタン | `ReservationFormModal` を開く（編集モード） | モーダル表示 |
| 予約削除 | 詳細モーダル「削除」ボタン | 確認ダイアログ後、削除 | 同画面 |
| ステータス変更 | 詳細モーダルのステータスドロップダウン | ステータス更新 | 同画面 |
| カルテ作成 | 詳細モーダル「カルテ作成」ボタン | カルテ作成画面へ遷移 | `/medical-records/new?petId=xxx` |
| 予約移動 | 週表示でカードD&D | 時間変更（15分単位スナップ、`motion/react` 使用） | 同画面 |
| 月/週表示切替 | ドロップダウン選択 | ビュー切替 | 同画面 |
| 前後移動 | ← → ボタン | 前/次の週または月へ | 同画面 |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| ダッシュボード「新規予約」 | `/reservations` | ボタンクリック |
| 詳細モーダル「カルテ作成」 | `/medical-records/new?petId=xxx` | サービス種別:診療系 |
| 詳細モーダル「トリミング記録作成」 | `/trimming/new?petId=xxx` | サービス種別:トリミング |
| 詳細モーダル「入院・ホテル登録」 | `/hospitalization/new?petId=xxx` | サービス種別:入院/ホテル |

## バリデーション
- 重複チェックエラー時: 構造化トースト（`toast.warning` + description）
- 新規時: 日付セルクリック日を初期値、デフォルト時間 10:00-11:00
- `ReservationFormFields` の日付・予約種別・担当医の各 `SelectTrigger` に `aria-describedby` で `FormFieldError` 接続

## 状態管理

| 状態 | 型 | 説明 |
|------|-----|------|
| `currentView` | `CalendarView` | 月/週表示切替 |
| `currentDate` | `Date` | 表示中の日付 |
| `selectedDoctorFilter` | `string` | 担当医フィルタ |
| `isFormModalOpen` | `boolean` | 予約フォームモーダル表示フラグ |
| `isDetailModalOpen` | `boolean` | 詳細モーダル表示フラグ |

## データ型
`ReservationAppointment`, `ReservationFormData`, `CalendarView`, `ReservationStatus`, `VisitType`

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/reservations` | 予約一覧取得 | 未実装 |
| GET | `/api/v1/reservations/:id` | 予約詳細取得 | 未実装 |
| POST | `/api/v1/reservations` | 予約作成 | 未実装 |
| PUT | `/api/v1/reservations/:id` | 予約更新 | 未実装 |
| DELETE | `/api/v1/reservations/:id` | 予約削除 | 未実装 |
| POST | `/api/v1/reservations/:id/check-in` | 受付処理 | 未実装 |
| POST | `/api/v1/reservations/:id/cancel` | 予約キャンセル | 未実装 |

## 実装状況
- フロントエンド(ui-sample): 実装済（モックデータ使用）
- バックエンドAPI: 未実装

## 備考
- 予約種別カラーはマスタ（`service_types.color`）と連動し、動的に凡例を生成する
- WeekView のドラッグ&ドロップは `motion/react` を使用（Y軸ドラッグで時間変更、15分単位スナップ）
- 旧仕様では検索コンポーネントが `PatientSearch` のみだったが、Step1 に `PatientSelectionTable`（`SelectedPetChip` 表示付き）が追加されている
