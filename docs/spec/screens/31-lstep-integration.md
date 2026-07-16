# Lステップ連携・管理 仕様書 (L-Step Integration)

## 概要
- **画面の目的**: Lステップ（LINE公式アカウント拡張ツール）との高度な連携設定、自動配信タグの管理、およびマーケティング施策の効果分析。
- **URLパターン**: 
  - 連携設定: `/settings/integrations/lstep`
  - タグ管理: `/settings/lstep/tags`
  - 健診対象者抽出: `/lstep/checkup-sync`
  - 分析レポート: `/lstep/analytics`
- **アクセス権限**: 
  - 連携設定・タグ管理・健診対象者抽出: 外部連携管理権限が必要（`ResourceHospitalSettings`）
  - 分析レポート: Lステップ分析閲覧権限が必要（`ResourceLstepAnalytics`）

---

## 1. 連携設定 (Integration Settings)

### 1.1 API 設定
- **Channel Access Token**: Messaging API 通信用の長期トークン。
- **LステップベースURL**: Lステップ側 API の接続先ベースURL。

### 1.2 判定閾値設定 (CPM/LTV)
医院の運営方針に合わせ、以下の判定基準をカスタマイズ可能です。
- **CPM バージョン**: V1（金額＋回数）または V2（回数重視）を選択。
- **ステージ境界値**: 「コア」「ノア」等と判定するための累計売上額や来院回数の閾値。

---

## 2. タグ管理と自動同期

### 2.1 自動管理タグ
システムが自動的に付与・剥がしを行うタグ群。
- **CPM ステージ (V2)**: `CPM_01_出会い`, `CPM_02_これから`, `CPM_03_いいかんじ`, `CPM_04_ファミリー`, `CPM_05_ノア`（V1 選択時は `cpm_encounter` / `cpm_growing` / `cpm_core` / `cpm_spot` / `cpm_noah` / `cpm_dormant` の英数字タグ）。
- **来院間隔**: `VISIT_120日超`, `VISIT_180日超` 等。
- **属性**: `PET_犬あり`, `PET_猫あり`, `LTV_上位20`, `LTV_フード購入あり`。

### 2.2 健診対象者一括同期 (`CheckupSyncPage`)
- **抽出ロジック**: 健診種別・動物種・最終来院日・年齢・慢性疾患有無・CPM ステージ・累計診療費・年間来院回数・最終健診実施日を任意に組み合わせて絞り込むフォーム方式（固定条件ではない）。
- **アクション**: 抽出リストから送信可能な対象者を選択し、任意のタグ名（例: `checkup_annual_2026`）を指定して一括付与、Lステップ側のセグメント配信へ繋げます。

---

## 3. 技術仕様

### 3.1 同期エンジン
- **リアルタイム同期**: 会計完了、ペット登録、死亡記録などのイベントをトリガーに即座にタグを更新。
- **バッチ同期**: LTV 上位 20% の再計算や休眠判定を深夜バッチで実行。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/clinics/:clinic_id/lstep-settings` | 連携設定と判定閾値の取得 | `hospital-settings` | `view` |
| PATCH | `/api/v1/clinics/:clinic_id/lstep-settings` | 連携設定の更新 | `hospital-settings` | `edit` |
| DELETE | `/api/v1/clinics/:clinic_id/lstep-settings` | 連携設定の削除 | `hospital-settings` | `delete` |
| POST | `/api/v1/clinics/:clinic_id/lstep-settings/test-connection` | 接続テスト | `hospital-settings` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/trigger-priorities` | 配信トリガー優先順位の参照 | `hospital-settings` | `view` |
| PATCH | `/api/v1/clinics/:clinic_id/lstep/trigger-priorities` | 配信トリガー優先順位の更新 | `hospital-settings` | `edit` |
| GET | `/api/v1/clinics/:clinic_id/lstep/tag-summary` | 現在のタグ保有者数の統計取得 | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/owners` | 指定タグの飼主一覧取得（タグ管理） | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/checkup-sync/preview` | 健診対象者抽出条件のプレビュー取得 | `owners` | `view` |
| POST | `/api/v1/clinics/:clinic_id/lstep/checkup-sync` | 指定条件の対象者へのタグ一括付与 | `owners` | `edit` |
| GET | `/api/v1/clinics/:clinic_id/lstep-tag-code-mappings` | タグ名ごとの外部コード紐付け取得 | `hospital-settings` | `view` |
| PUT | `/api/v1/clinics/:clinic_id/lstep-tag-code-mappings/:tag_name` | タグ別コード紐付け更新 | `hospital-settings` | `edit` |
| GET | `/api/v1/lstep-tag-config/auto-managed-prefixes` | 自動管理プレフィックス一覧取得 | `hospital-settings` | `view` |
| POST | `/api/v1/lstep-tag-config/auto-managed-prefixes` | 自動管理プレフィックス追加 | `hospital-settings` | `create` |
| DELETE | `/api/v1/lstep-tag-config/auto-managed-prefixes/:id` | 自動管理プレフィックス削除 | `hospital-settings` | `delete` |
| GET | `/api/v1/lstep-tag-config/condition-tag-mappings` | 条件別タグマッピング一覧取得 | `hospital-settings` | `view` |
| POST | `/api/v1/lstep-tag-config/condition-tag-mappings` | 条件別タグマッピング追加 | `hospital-settings` | `create` |
| DELETE | `/api/v1/lstep-tag-config/condition-tag-mappings/:id` | 条件別タグマッピング削除 | `hospital-settings` | `delete` |
| GET | `/api/v1/lstep-tag-config/send-purpose-tag-prefixes` | 送信目的別タグプレフィックス一覧取得 | `hospital-settings` | `view` |
| POST | `/api/v1/lstep-tag-config/send-purpose-tag-prefixes` | 送信目的別タグプレフィックス追加 | `hospital-settings` | `create` |
| DELETE | `/api/v1/lstep-tag-config/send-purpose-tag-prefixes/:id` | 送信目的別タグプレフィックス削除 | `hospital-settings` | `delete` |
| GET | `/api/v1/clinics/:clinic_id/lstep/analytics/delivery-stats` | 月次配信統計の取得 | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/analytics/visit-conversion` | 来院転換データの集計 | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/csv-imports` | 友だち属性 CSV インポート履歴の取得 | `lstep-csv-import` | `view` |
| POST | `/api/v1/clinics/:clinic_id/lstep/csv-imports/friend-attributes` | 友だち属性 CSV のアップロード | `lstep-csv-import` | `edit` |

---
