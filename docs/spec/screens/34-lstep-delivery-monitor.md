# Lステップ配信監視 仕様書 (L-Step Delivery Monitor)

## 概要
- **画面の目的**: システムが自動生成した Lステップ **配信トリガー** の実行状況、除外判定、および API 通信の成否をリアルタイムに監視する。
- **観測範囲**: `lstep_delivery_trigger_log` のみ。会計確定後の CPM 同期など **ordinary タグ同期（request-local secondary）は本画面の対象外**であり、当該経路は trigger log に書かない。
- **URLパターン**: `/lstep/delivery-monitor`
- **アクセス権限**: FE/BE とも `ResourceLstepAnalytics:view`。親 `/lstep` は権限を加算しない。
- **Deploy gate（`LSTEP_WRITE_API_ENABLED`）**: Write 系が無効のとき HTTP を送らず **`ErrWriteDisabled` を返す（`nil` 成功にしない）**。判定・除外・ログ行の作成と監視 UI は継続する（[`LSTEP_WRITE_API_PAUSE.md`](../../ops/deploy/LSTEP_WRITE_API_PAUSE.md)。enable / stop / rollback 手順は同 runbook が正本）。
- **Clinic gate（`is_sync_enabled` / API キー）**: 同期無効または API キー未設定の clinic はクライアント未構築による意図的スキップ（`nil, nil`）。deploy gate とは**別契約**である。

---

## 1. 画面構成

### 1.1 配信実行ログテーブル (`DeliveryLogsTable`)
指定期間（既定は当日）の配信（および配信予定）が予定日時の降順で表示されます（サーバ側ページネーション付き）。

| カラム | 説明 |
|:---|:---|
| **種別** | `first_visit_followup_3d`, `dormant_prevention_365d` 等のトリガーコード（日本語ラベルで表示）。 |
| **飼い主** | メッセージ送信対象のオーナー名。 |
| **予定日時** | シナリオが発火する予定の日時（`scheduled_at`）。 |
| **ステータス** | `scheduled`, `fired`, `excluded`, `failed` の 4 状態。 |
| **送信日時** | 実際に配信された日時（`fired_at`。未送信は「—」）。 |
| **除外理由** | `excluded` の場合、なぜ配信がスキップされたか（実値: `delivery_excluded_flag`＝飼主の配信除外設定、`no_line_user_id`＝LINE 未連携、`excl_tag_delivery_stop`＝配信停止タグ）。 |

### 1.2 検索・フィルタ (`DeliveryMonitorFilters`)
- **期間**: 予定日時（`scheduled_at`）の From-To 絞り込み。
- **トリガー種別**: `first_visit_followup_3d` 等のトリガーコードでの絞り込み。
- **ステータス**: `scheduled`/`fired`/`excluded`/`failed` による絞り込み（失敗のみ抽出して再送検討等に使用）。

### 1.3 サマリ表示 (`DeliverySummaryCards`)
- **ステータス別サマリカード**: 予定（`scheduled`）・送信済（`fired`）・除外（`excluded`）・失敗（`failed`）・優先度抑制（`suppressed_by_priority`）の 5 枚の件数カード。
- **失敗警告バナー (`DeliveryFailedWarning`)**: 期間内に `failed` が 1 件以上ある場合のみ表示される警告。
- **除外理由内訳 (`DeliveryExcludedReasonBreakdown`)**: `excluded` がある場合に理由別件数をバッジ表示。
- ヘッダの「更新」ボタンでサマリ・ログの両方を再取得する。

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

### 2.3 バッチ失敗契約と durable 計上
定時の配信トリガーバッチは **scheduled multi-resource best effort** である。

- multi-clinic / multi-owner / multi-trigger で 1 件失敗後も他対象は続行する。
- 1 飼主処理の失敗は **single-owner propagation** で上位へ伝播し、`BatchRunResult`（`Processed = Succeeded + Failed`）と監査 metadata の `processed_count`/`error_count` に **必ず計上**する（silent swallow 禁止）。
- 必須 dependency 欠落（settings 未構成・clinic 一覧取得失敗等）は fail-closed。
- 画面上の `failed` 行・失敗サマリは、上記 durable 計上のうち **配信トリガーログに落ちた owner 単位の結果**を観測する UI である。バッチ全体の `BatchRunResult` は scheduler / 監査側の観測点であり、本画面 API のレスポンス envelope ではない。
- 候補 owner に対する owner / 当日 claim / 抑制 / tag-cache 読みは clinic スコープ bulk-read を必須とし、owner 数線形の N+1 を置かない（opt-out・suppression・daily-claim 意味論と bounded memory は維持）。

Deploy gate（`LSTEP_WRITE_API_ENABLED`）OFF 時は外部タグ write が HTTP 未送信かつ `ErrWriteDisabled` でも、除外・抑制・ログ作成と本監視 UI は動作する。Clinic gate（`is_sync_enabled=false` 等）のクライアント未構築スキップとは別契約である。再有効化の enable / stop / rollback は [`LSTEP_WRITE_API_PAUSE.md`](../../ops/deploy/LSTEP_WRITE_API_PAUSE.md) を正とし、本 spec に手順を複製しない。

---

## 3. 技術仕様

### 構成コンポーネント
- **`LstepDeliveryMonitorPage`**: メイン監視ページ。
- **`DeliveryMonitorFilters`**: 期間・トリガー種別・ステータスの検索フィルタ。
- **`LstepDeliveryMonitorLogsTable`**: 配信ログの一覧テーブル `DeliveryLogsTable`（ステータスは `BADGE` デザイントークンで色分け表示。`Pagination` によるページ送り）。
- **`LstepDeliveryMonitorPageParts`**: サマリカード（`DeliverySummaryCards`）・失敗警告（`DeliveryFailedWarning`）・除外理由内訳（`DeliveryExcludedReasonBreakdown`）の部品群。

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/clinics/:clinic_id/lstep/delivery-monitor/summary` | 期間別配信サマリの取得 | `lstep-analytics` | `view` |
| GET | `/api/v1/clinics/:clinic_id/lstep/delivery-monitor/logs` | 配信ログ履歴の取得 | `lstep-analytics` | `view` |

---
