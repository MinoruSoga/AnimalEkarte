# 入院詳細・デイリーカルテ 仕様書

## 概要
- **画面の目的**: 特定の入院患者のケアプラン・日次記録（バイタル・ケアログ・スタッフメモ）を管理する
- **URLパターン**: `/hospitalization/:id`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト
```
┌──────────────────────────────────────────────────────┐
│ ← 戻る  [ペット情報ヘッダー]  [編集/退院ボタン]      │
│ (ペット名/飼主名/入院No/ケージ/ステータス)           │
├──────────────────────────────────────────────────────┤
│ タブ: [ケアプラン] [デイリーログ] [コスト]           │
│                                                      │
│ 各タブコンテンツ                                     │
│                                                      │
└──────────────────────────────────────────────────────┘
```

## 患者情報ヘッダー

| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| ペット名 | string | ペット名 | `pets.name` |
| 種/品種 | string | 動物種・品種 | `pets.species`, `pets.breed` |
| 飼主名 | string | 飼い主氏名 | `owners.name` |
| 入院No | string | 入院番号 | `hospitalizations.hospitalization_no` |
| 入院種別 | enum | 入院/ホテル | `hospitalizations.type` |
| ケージ名 | string | 割り当てケージ | `cages.name` |
| 入院開始日 | date | 入院開始日 | `hospitalizations.start_date` |
| 退院予定日 | date | 退院予定日 | `hospitalizations.end_date` |
| ステータス | enum | 入院中/退院済等 | `hospitalizations.status` |
| 飼い主要望 | text | 飼い主からの要望 | `hospitalizations.owner_request` |

## タブ構成

### ケアプランタブ
入院中の投薬・処置スケジュール管理。

| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| 種別 | enum | food/medicine/treatment/instruction/item | `care_plan_items.type` |
| 項目名 | string | ケアプラン項目名 | `care_plan_items.name` |
| 詳細・用量 | text | 投薬量や処置の詳細 | `care_plan_items.description` |
| タイミング | json | 実施タイミング（朝/昼/夜等） | `care_plan_items.timing` |
| ステータス | enum | active/completed/discontinued | `care_plan_items.status` |
| 単価 | decimal | 単価 | `care_plan_items.unit_price` |
| カテゴリ | string | カテゴリ | `care_plan_items.category` |

### デイリーログタブ
日次記録の管理（日付選択→その日のバイタル・ケアログ・スタッフメモ）。

#### バイタル (vitals)
| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| 記録時刻 | time | 計測時刻 | `vitals.recorded_time` |
| 体温 | decimal | 体温(℃) | `vitals.temperature` |
| 心拍数 | int | 心拍数(bpm) | `vitals.heart_rate` |
| 呼吸数 | int | 呼吸数(回/分) | `vitals.respiration_rate` |
| 体重 | decimal | 体重(kg) | `vitals.weight` |
| 備考 | text | メモ | `vitals.notes` |
| 記録者 | string | スタッフ名 | `staffs.name` |

#### ケアログ (care_logs)
| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| 記録時刻 | time | 実施時刻 | `care_logs.recorded_time` |
| 種別 | enum | food/excretion/medicine/treatment/other | `care_logs.type` |
| ステータス | enum | completed/partial/skipped | `care_logs.status` |
| 結果 | string | 実施結果 | `care_logs.value` |
| 備考 | text | メモ | `care_logs.notes` |

#### スタッフメモ (staff_notes)
| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| 記録時刻 | time | 記録時刻 | `staff_notes.recorded_time` |
| 内容 | text | メモ内容 | `staff_notes.content` |
| 記録者 | string | スタッフ名 | `staffs.name` |

### コストタブ
入院費用の自動計算。

| フィールド名 | 型 | 説明 |
|------------|-----|------|
| ケアプラン項目ごとの費用 | table | 単価×日数の費用一覧 |
| 小計 | calc | 税抜小計 |
| 消費税 | calc | 消費税額 |
| 合計 | calc | 税込合計 |

## UI コンポーネント
- **HospitalizationPatientHeader**: ペット情報・入院情報ヘッダー
- **HospitalizationBasicInfo**: 基本情報表示
- **CarePlanSection**: ケアプラン一覧
- **CarePlanItemRow**: ケアプランの1行
- **CarePlanDialog**: ケアプラン追加・編集ダイアログ
- **HospitalizationCostSummary**: コスト集計
- **DischargeAlertDialog**: 退院確認ダイアログ
- **HospitalizationMobileLayout**: モバイル対応レイアウト

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 編集 | 「編集」ボタン | 入院編集フォームへ | `/hospitalization/:id/edit` |
| 退院処理 | 「退院」ボタン | 確認ダイアログ→退院ステータスに変更 | 同画面（ステータス更新） |
| ケアプラン追加 | 「追加」ボタン | CarePlanDialog開く | ダイアログ |
| ケアプラン編集 | 行の編集ボタン | CarePlanDialog開く（編集モード） | ダイアログ |
| ケアプラン削除 | 行の削除ボタン | 確認後削除 | 同画面 |
| バイタル記録 | 「バイタル追加」ボタン | バイタル入力フォーム | ダイアログ/インライン |
| ケアログ記録 | 各ケアプラン項目のチェックボックス | 実施記録を保存 | 同画面 |
| スタッフメモ追加 | 「メモ追加」ボタン | テキスト入力して保存 | 同画面 |
| 戻る | 「戻る」ボタン | 入院一覧へ | `/hospitalization` |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| 入院一覧 | `/hospitalization/:id` | 行/カードクリック |
| 「編集」ボタン | `/hospitalization/:id/edit` | ボタンクリック |
| 退院完了 | `/hospitalization` | 退院処理後 |

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/hospitalizations/:id` | 入院詳細取得 | 📋 未実装 |
| POST | `/api/v1/hospitalizations/:id/discharge` | 退院処理 | 📋 未実装 |
| GET | `/api/v1/hospitalizations/:id/care-plans` | ケアプラン一覧 | 📋 未実装 |
| POST | `/api/v1/hospitalizations/:id/care-plans` | ケアプラン追加 | 📋 未実装 |
| PUT | `/api/v1/care-plans/:id` | ケアプラン更新 | 📋 未実装 |
| DELETE | `/api/v1/care-plans/:id` | ケアプラン削除 | 📋 未実装 |
| GET | `/api/v1/hospitalizations/:id/daily-records` | デイリーレコード一覧 | 📋 未実装 |
| POST | `/api/v1/daily-records/:id/vitals` | バイタル追加 | 📋 未実装 |
| POST | `/api/v1/daily-records/:id/care-logs` | ケアログ追加 | 📋 未実装 |
| POST | `/api/v1/daily-records/:id/notes` | スタッフメモ追加 | 📋 未実装 |

## 実装状況
- フロントエンド(ui-sample): 実装済（モックデータ使用）
- バックエンドAPI: 未実装
