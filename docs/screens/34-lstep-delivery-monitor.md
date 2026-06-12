# Lステップ配信監視 仕様書 (L-Step Delivery Monitor)

## 概要
- **画面の目的**: システムが自動生成した Lステップ配信トリガーの実行状況、除外判定、および API 通信の成否をリアルタイムに監視する。
- **URLパターン**: `/lstep/delivery-monitor`
- **アクセス権限**: 外部連携管理権限が必要（`ResourceHospitalSettings`）

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
| **判定理由** | `excluded` の場合、なぜ配信がスキップされたか（例：死亡ガード、スタッフ除外、重複回避）。 |

### 1.2 高度な検索・フィルタ (`NotionFilter`)
- **ステータス別**: 失敗（failed）のみを抽出して再送検討等に使用。
- **オーナー名検索**: 特定の顧客にリマインドが正しくスケジュールされているか確認。

---

## 2. 主要な監視ロジック

### 2.1 自動除外ガード (Auto-Exclusion)
配信直前にバックエンドが以下のチェックを行い、不適切な送信を自動で阻止します。
1.  **死亡ガード**: ペットが亡くなっている場合。
2.  **オプトアウト**: 飼主が LINE 連携を解除、または配信停止設定にしている場合。
3.  **重複回避 (De-duplication)**: 同日に優先度の高い別の配信がある場合。

### 2.2 優先順位制御 (Priority Suppression)
`lstep_trigger_priorities` マスタの設定に基づき、低優先度のトリガーは `excluded` (理由: `suppressed_by_priority`) として処理されます。

---

## 3. 技術仕様

### 構成コンポーネント
- **`LstepDeliveryMonitorPage`**: メイン監視ページ。
- **`StatusBadge`**: 各ステータスを意味的なカラーで表示。

### API連携
| メソッド | エンドポイント | 用途 |
|:---|:---|:---|
| GET | `/api/v1/clinics/:clinic_id/lstep/delivery-monitor/summary` | 期間別配信サマリ（scheduled/fired/failed 等）を取得。 |
| GET | `/api/v1/clinics/:clinic_id/lstep/delivery-monitor/logs` | 配信ログ履歴の取得。 |
| GET | `/api/v1/clinics/:clinic_id/lstep/trigger-priorities` | 現在の優先順位設定の参照。 |

---
