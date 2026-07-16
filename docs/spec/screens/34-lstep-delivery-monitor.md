# Lステップ配信監視 仕様書 (L-Step Delivery Monitor)

## 概要
- **画面の目的**: システムが自動生成した Lステップ配信トリガーの実行状況、除外判定、および API 通信の成否をリアルタイムに監視する。
- **URLパターン**: `/lstep/delivery-monitor`
- **アクセス権限**: フロントエンド表示には外部連携管理権限（`ResourceHospitalSettings`）が必要。バックエンドAPIには Lステップ分析閲覧権限（`ResourceLstepAnalytics`）が必要。

---

## 1. 画面構成

### 1.1 配信実行ログテーブル (`DataTable`)
直近の配信（および配信予定）が時系列で表示されます。

| カラム | 説明 |
|:---|:---|
| **予定/実行日時** | シナリオが発火する、または発火した日時。 |
| **対象飼主** | メッセージ送信対象のオーナー名。 |
| **トリガー種別** | `first_visit_followup_3d`, `dormant_prevention_365d` 等のコード。 |
| **ステータス** | `scheduled`, `fired`, `excluded`, `failed` の 4 状態。 |
| **判定理由** | `excluded` の場合、なぜ配信がスキップされたか（実値: `delivery_excluded_flag`＝飼主の配信除外設定、`no_line_user_id`＝LINE 未連携、`excl_tag_delivery_stop`＝配信停止タグ）。 |

### 1.2 検索・フィルタ (`DeliveryMonitorFilters`)
- **期間**: 予定/実行日時の From-To 絞り込み。
- **トリガー種別**: `first_visit_followup_3d` 等のトリガーコードでの絞り込み。
- **ステータス**: `scheduled`/`fired`/`excluded`/`failed` による絞り込み（失敗のみ抽出して再送検討等に使用）。

---

## 2. 主要な監視ロジック

### 2.1 自動除外ガード (Auto-Exclusion)
配信直前にバックエンド（`checkExclusion`）が以下のチェックを行い、不適切な送信を `excluded` として自動で阻止します。
1.  **配信除外フラグ**: 飼主に配信除外設定がある場合（`delivery_excluded_flag`）。
2.  **LINE 未連携**: 飼主に LINE ユーザー ID が紐付いていない場合（`no_line_user_id`）。
3.  **配信停止タグ**: 飼主に配信停止タグが付与されている場合（`excl_tag_delivery_stop`）。

なお、ペットの死亡はこの直前チェックではなくライフサイクル処理（`HandlePetDeath` によるタグ除去・全頭死亡時の Lステップ全タグ削除）を通じて配信対象から外れる仕組みです。

### 2.2 優先順位制御 (Priority Suppression)
`lstep_trigger_priorities` マスタの設定に基づき、同日に優先度の高い別トリガーがある場合、低優先度のトリガーはログの `suppressed_by_priority` フラグ（+`suppression_reason`）が立てられ配信されません（`excluded_reason` とは別カラムで管理）。

---

## 3. 技術仕様

### 構成コンポーネント
- **`LstepDeliveryMonitorPage`**: メイン監視ページ。
- **`DeliveryMonitorFilters`**: 期間・トリガー種別・ステータスの検索フィルタ。
- **`LstepDeliveryMonitorLogsTable`**: 配信ログの一覧テーブル（ステータスは `BADGE` デザイントークンで色分け表示）。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/clinics/:clinic_id/lstep/delivery-monitor/summary` | 期間別配信サマリの取得 | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/delivery-monitor/logs` | 配信ログ履歴の取得 | `lstep-analytics` | `view` |

---
