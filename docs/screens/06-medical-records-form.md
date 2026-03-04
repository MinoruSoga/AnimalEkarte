# カルテ入力/編集 仕様書

## 概要
- **画面の目的**: SOAPS形式（Subjective, Objective, Assessment, Plan, Surgery/Special）に基づく診療記録の作成・編集
- **URLパターン**:
  - 新規作成: `/medical-records/new?petId=xxx`
  - 編集: `/medical-records/:id`
- **アクセス権限**: 認証済ユーザー全員

## 画面レイアウト
```
┌──────────────────────────────────────────────────────┐
│ ← 戻る  [ペット情報ヘッダー]  [保存/確定ボタン]      │
│ (ペット名/飼主名/患者番号/担当医/来院日)             │
├──────────────────────────────────────────────────────┤
│                                                      │
│ タブ: [問診] [身体検査] [診断・処置] [処方] [画像]   │
│       [見積]                                         │
│                                                      │
│ 各タブのコンテンツエリア                             │
│                                                      │
└──────────────────────────────────────────────────────┘
```

## タブ構成

### 問診タブ (Interview)
| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| 来院種別 | enum | 初診/再診 | `visit_type` |
| 主訴 | textarea | 来院理由 | `chief_complaint` |
| 現病歴・経過 | textarea | 病歴の経過 | `subjective` |
| 既往歴 | 参照表示 | 過去の診療記録 | 別テーブル参照 |
| ワクチン歴 | 参照表示 | ワクチン接種記録 | `vaccinations` |

### 身体検査タブ (Physical Exam)
| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| 体重 | decimal | 体重(kg) | `objective` (JSON内) |
| 体温 | decimal | 体温(℃) | `objective` (JSON内) |
| 心拍数 | int | 心拍数(bpm) | `objective` (JSON内) |
| 呼吸数 | int | 呼吸数(回/分) | `objective` (JSON内) |
| 身体検査所見 | textarea | 身体検査の所見 | `objective` |

### 診断・処置タブ (Diagnosis & Plan)
| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| 診断名 | string | 診断名称 | `assessment` |
| 処置内容 | textarea | 処置の詳細 | `treatment` |
| 計画・方針 | textarea | 治療計画 | `plan` |
| 手術・特記事項 | textarea | 手術内容や特記事項 | `surgery_notes` |
| 処置一覧 | テーブル | 使用した処置・薬品マスタ選択 | `accounting_items`連携 |

### 処方タブ (Prescription)
| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| 処方内容 | textarea | 処方箋の内容 | `prescription` |
| 処方マスタ選択 | 検索 | マスタから薬剤選択（未実装） | `master_items` |

### 画像タブ (Images)
| フィールド名 | 型 | 説明 |
|------------|-----|------|
| 画像ギャラリー | gallery | アップロード済み画像一覧（未実装） |
| 画像アップロード | file | 画像のアップロード（未実装） |

### 見積タブ (Estimate)
| フィールド名 | 型 | 説明 | DB列 |
|------------|-----|------|------|
| 見積明細 | table | 診療項目・薬品の見積一覧 | `accounting_items`候補 |
| 合計金額 | calc | 自動計算 | - |
| 会計作成 | button | 会計データを作成 | `accountings` |

## UI コンポーネント
- **DiagnosisHeader**: ペット情報・バイタル表示ヘッダー（固定表示）
- **MedicalRecordInterview**: 問診タブ
- **MedicalRecordExamination**: 身体検査タブ
- **MedicalRecordDiagnosisPlan**: 診断・処置タブ
- **MedicalRecordTreatment**: 処置一覧・マスタ選択
- **TreatmentSearchDialog**: 処置・薬品のマスタ検索ダイアログ
- **MedicalRecordEstimate**: 見積タブ
- **EstimateForm**: 見積フォーム
- **VaccinationHistory / InterviewHistory**: 既往歴・ワクチン歴参照コンポーネント
- **MedicalRecordImage**: 画像管理タブ
- **MedicalRecordVaccination**: ワクチン接種タブ

## ユーザーアクション

| アクション | トリガー | 処理内容 | 遷移先 |
|-----------|---------|---------|--------|
| 一時保存 | 「保存」ボタン | ステータス「作成中」で保存 | 同画面 |
| カルテ確定 | 「確定」ボタン | ステータス「確定済」に変更 | カルテ一覧 or 同画面 |
| 会計作成 | 見積タブ「会計作成」 | 会計データ生成→会計画面へ | `/accounting/new?petId=xxx` |
| タブ切替 | タブクリック | コンテンツ切替 | 同画面 |
| 処置追加 | 「処置追加」ボタン | TreatmentSearchDialog開く | ダイアログ |
| 戻る | 「戻る」ボタン | カルテ一覧へ | `/medical-records` |

## 画面遷移

| 遷移元 | 遷移先 | 条件 |
|--------|--------|------|
| ペット選択 | `/medical-records/new?petId=xxx` | ペット選択後 |
| カルテ一覧 | `/medical-records/:id` | 行クリック |
| 見積タブ「会計作成」 | `/accounting/new?petId=xxx` | 会計作成ボタン |

## バリデーション
- **主訴**: 確定時は必須推奨（警告のみ）
- **確定後**: 編集不可（読み取り専用表示）

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/medical-records/:id` | カルテ詳細取得 | 未実装 |
| POST | `/api/v1/medical-records` | カルテ作成 | 未実装 |
| PUT | `/api/v1/medical-records/:id` | カルテ更新 | 未実装 |
| POST | `/api/v1/medical-records/:id/finalize` | カルテ確定 | 未実装 |
| GET | `/api/v1/pets/:petId/medical-records` | ペットのカルテ一覧 | 未実装 |
| GET | `/api/v1/master/items` | マスタ項目一覧（処置・薬品） | 未実装 |

## 実装状況
- フロントエンド(ui-sample): 実装済（モックデータ使用、画像・処方機能は未実装）
- バックエンドAPI: 未実装
