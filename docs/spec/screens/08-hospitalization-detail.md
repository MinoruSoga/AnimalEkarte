# 入院詳細・デイリーカルテ 仕様書 (Hospitalization Detail)

## 概要
- **画面の目的**: 入院患者のケアプラン（治療・給餌計画）の管理、および日々のバイタルや処置実施状況の時系列記録。
- **URLパターン**: `/hospitalization/:id`
- **アクセス権限**: `hospitalization` リソースの `view` 権限（`/hospitalization` 配下は `RequirePermission` でガード。操作権限は `usePermission` で制御）

---

## 画面構成

### 1. プランとログの二画面構成 (Expanded View)
デスクトップ環境では、計画と実績を一目で把握できる上下分割レイアウトを採用しています。
- **上段（ケアプラン）**: 入院期間中に行うべき投薬、給餌、処置等のスケジュールを管理。
- **下段（デイリーログ）**: 実際に実施したケアの内容やバイタル、スタッフ間の引き継ぎメモを日付単位で表示。

### 2. デイリーログのセクション構成
日次の記録は、日付ナビゲーションで対象日を切り替えつつ、「バイタル」「ケアログ」「スタッフメモ」の3セクションに分類して表示・入力されます。

---

## 主要機能

### 1. ケアプランの管理
- **タスク設定**: 獣医師が指示した投薬内容や給餌量を登録。
- **実施登録**: 看護スタッフが `DailyVitalsSection` / `DailyCareLogsSection` / `DailyStaffNotesSection` それぞれの追加ダイアログから、時刻・測定値（バイタル）または内容・メモ（ケアログ／スタッフメモ）を独立して記録します（ケアプランのタスクとの自動紐付けや実施者の記録項目はありません）。

### 2. 退院プロセスと会計連携
- **ステータス移行**: 「退院処理」を実行すると、患者のステータスが `discharged` に更新されます。
- **会計連携誘導**: 確認ダイアログの「退院後、そのまま会計画面へ進む」にチェックして実行すると、ケアプランの明細を取り込んだ会計が作成され、会計詳細画面（会計が作成されなかった場合は新規会計作成フォーム）へ遷移し、精算漏れを防止します。チェックしない場合はステータス更新のみ行い、入院一覧へ戻ります。

---

## 技術仕様

### 構成コンポーネント
- **`HospitalizationExpandedView`**: デスクトップ用レイアウトコンテナ。
- **`DailyRecordsTab`**: 時系列ログの表示・管理モジュール。
- **`DischargeAlertDialog`**: 安全な退院手続きのための最終確認ダイアログ。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/hospitalizations/:id` | 入院詳細情報の取得 | `hospitalization` | `view` |
| GET | `/api/v1/hospitalizations/:id/daily-records/:date` | 対象日の日次記録の取得 | `hospitalization` | `view` |
| POST | `/api/v1/hospitalizations/:id/daily-records` | 対象日の日次記録（空のデイリーカルテ）の作成 | `hospitalization` | `create` |
| POST | `/api/v1/hospitalizations/:id/daily-records/:date/vitals` | バイタル記録の追加 | `hospitalization` | `create` |
| POST | `/api/v1/hospitalizations/:id/daily-records/:date/care-logs` | ケアログの追加 | `hospitalization` | `create` |
| POST | `/api/v1/hospitalizations/:id/daily-records/:date/staff-notes` | スタッフメモの追加 | `hospitalization` | `create` |
| GET | `/api/v1/hospitalizations/:id/care-plan-items` | ケアプランアイテムの取得 | `hospitalization` | `view` |
| POST | `/api/v1/hospitalizations/:id/care-plan-items` | ケアプランアイテムの追加 | `hospitalization` | `create` |
| PATCH | `/api/v1/hospitalizations/:id/care-plan-items/:itemId` | ケアプランアイテムの更新 | `hospitalization` | `edit` |
| DELETE | `/api/v1/hospitalizations/:id/care-plan-items/:itemId` | ケアプランアイテムの削除 | `hospitalization` | `delete` |
| PATCH | `/api/v1/hospitalizations/:id` | 入院情報（ケージ等）の更新。会計連携なしの退院（ステータス・退院日更新）にも使用 | `hospitalization` | `edit` |
| POST | `/api/v1/hospitalizations/:id/discharge-with-billing` | 退院処理（会計連携含む） | `hospitalization` | `edit` |

---
