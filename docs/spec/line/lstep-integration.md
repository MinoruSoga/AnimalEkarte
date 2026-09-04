# Lステップ連携 戦略仕様書 (L-Step CRM Integration)

> **目的**: CPM判定・全15種配信トリガーのLステップ戦略を定義する。
> **読者**: マーケ担当・実装者。
> **タイミング**: CPM判定・配信トリガー仕様の確認時。

> **Animal Ekarte**: カルテデータに基づいた自動マーケティングの実現
> **最新更新**: 2026-08-31 | **ステータス**: 実装済みだが既知の安全・性能gapあり。deployment / credential / external scenario / monitoring readiness は [release readiness runbook](../../ops/deploy/runbooks/STG_PRE_DEPLOY_READINESS_CHECK.md) 等のrelease evidenceで別管理

---

> **注記 (2026-07-10 / 更新 2026-07-31)**: Lステップへの Write API（タグ付与・タグ解除・プロパティ更新）は **`LSTEP_WRITE_API_ENABLED`（既定 OFF）+ clinic `is_sync_enabled` の二重 gate** で抑止。gate OFF 時は外部 HTTP 0 + `ErrWriteDisabled`（silent noop 成功ではない）。その後のローカル DB / cache / audit / logging は呼び出し元ごとの失敗契約に従い、一律には継続しない。詳細: [`docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md`](../../ops/deploy/LSTEP_WRITE_API_PAUSE.md)

---

## 1. 連携のコンセプト

本システムと Lステップの連携は、**「カルテが判断し、Lステップが届ける」**という役割分担に基づき、臨床データに裏打ちされた 1to1 マーケティングを実現します。

### 役割の定義
- **カルテシステム (判定の脳)**: 来院頻度、累計売上、診療内容から「顧客の状態」をリアルタイムに解析。
- **Lステップ (配信の腕)**: カルテが付与した「タグ」をトリガーに、最適なタイミングでメッセージを自動送信。

---

## 2. 顧客分類ロジック (CPM分析)

### 2.1 CPM V2 (来院回数ベース)
日常的な運用の中心となる判定基準。累計来院回数（医院ごとに閾値変更可、既定値は以下）で 5 段階を判定する。
- **出会い**: 累計 0〜1 回。
- **これから**: 累計 2〜3 回。
- **いいかんじ**: 累計 4〜7 回。
- **ファミリー**: 累計 8〜12 回。
- **ノア**: 累計 13 回以上。

LTV 上位 20% は CPM ステージとは独立した `LTV_上位20` タグとして別途付与される。

### 2.1a CPM V1（金額＋期間）と残余
タグ同期の既定は医院設定の CPM バージョンに従う。V1 は Encounter / Growing / Core / Spot / Noah / Dormant の 6 区分。該当しない飼主は `cpm_unclassified`（配信対象外。`allCPMStages` に含めない）。顧客集計ダッシュボードの人数チップは 6 区分 + Unclassified の 7 つ（画面正本: [36-aggregation-dashboard.md](../screens/36-aggregation-dashboard.md)）。

### 2.2 休眠予備軍 (VISITタグ)
最終来院日からの経過日数に応じ、以下のタグを自動付与・解除します。
- `VISIT_120日超` / `VISIT_180日超` / `VISIT_220日超` / `VISIT_240日超`

---

## 3. 自動配信トリガーの一覧 (全 15 種)

カルテ上のイベントに基づき、Lステップ側のシナリオが発火します。これらは**「衝突回避優先順位」**（数値が小さいほど優先）に従って制御されます。

| 優先度 | トリガーコード | 内容 |
|:---:|:---|:---|
| **2** | `dormant_prevention_365d` | 最終来院から 1 年経過。 |
| **3** | `checkup_followup` | 健診作成後に起動する follow-up。現行 call に健診結果や精密検査要否の predicate はない。 |
| **4** | `filaria_alert` / `flea_tick_alert` | フィラリア・ノミダニシーズン開始。 |
| **5** | `dormant_prevention_240d` | 休眠防止ステップ（240日経過）。 |
| **6** | `dormant_prevention_210d` | 休眠防止ステップ（210日経過）。 |
| **7** | `dormant_prevention_180d` | 休眠防止ステップ（180日経過）。 |
| **8** | `vaccine_deadline_60d` / `vaccine_deadline_30d` | ワクチン期限の 60日/30日前通知。 |
| **10** | `food_refill_reminder` | 療法食の購入から約 30日後の案内。 |
| **11** | `next_visit_reminder` | カルテ指定の「次回来院推奨日」の当日。 |
| **12** | `birthday_message` | ペットの誕生日の当日祝い。 |
| **13** | `first_visit_followup_3d` / `first_visit_followup_7d` | 初診の 3日後、7日後の調子伺い。 |
| **14** | `first_visit_welcome` | 初回 medical record 作成後（post-commit）の挨拶。会計完了起点ではない。 |

---

## 4. 配信安全ガード (Safety Controls)

現行の最終配信除外確認は owner opt-out、`delivery_excluded`、LINE ID、cached EXCL tag を確認する。同日 claim と優先度抑制により、同日に競合する trigger は高優先度を選ぶ。

**既知の safety gap**: `PetStatus` は `alive` / `deceased` であり、`transferred` status はこの契約にない。最終 claim / write 前の確認は pet status を直接読まない。全 pet 死亡時の tag cleanup は best-effort で、durable exclusion flag を確立しない。また一部候補 query は死亡 pet を除外しない。したがって死亡・転院関連配信の絶対遮断は保証しない。

必要な是正契約は、候補取得時と最終 claim / 外部 write 直前の両方で durable な owner/pet 除外証跡を fail-closed に確認し、cache / external write failure の alert、reconciliation、手動停止 fallback を持つことである。この source defect は本 docs-only 変更では修正しない。

---

## 5. 失敗契約（経路別・混同禁止）

LSTEP の失敗時挙動は経路ごとに 1 契約だけを持つ。詳細アーキテクチャは [architecture.md](./architecture.md) §4。

| 経路 | 契約 | 要点 |
|:---|:---|:---|
| 定時バッチ（配信トリガー・休眠・LTV・健診予防など multi-clinic / multi-owner） | **scheduled multi-resource best effort** | 1 件失敗後も続行。`BatchRunResult`（`Processed = Succeeded + Failed`）と `processed_count`/`error_count` 監査で **durable な部分結果計上を必須**。必須 dependency 欠落は fail-closed。silent swallow は新規禁止 |
| 1 飼主分のタグ同期本体（バッチ内 1 owner 処理を含む） | **single-owner propagation** | 望ましいタグ Add/Remove 失敗は呼び出し元へ伝播し、上位の Failed 計上に載せる |
| 会計確定後の CPM 同期、手動 LINE 送信後の purpose タグなど副次処理 | **request-local nonfatal secondary notification** | 本処理（会計・送信本体）は成功のまま。副次失敗はログ必須で本処理を反転させない。**`lstep_delivery_trigger_log` は使わない** |

### 5.1 書き込み停止と運用停止

- **Write dual-gate**: タグ付与・解除・プロパティ更新の外部 HTTP は deploy/clinic gate で抑止され `ErrWriteDisabled` が返る。その後のローカル DB / cache / audit / logging は呼び出し元ごとの失敗契約に従い、一律には継続しない。[`LSTEP_WRITE_API_PAUSE.md`](../../ops/deploy/LSTEP_WRITE_API_PAUSE.md)
- **clinic 停止**: `is_sync_enabled=false` の医院は再有効化後もサービス層で同期対象外。

### 5.2 定時 delivery バッチの読み取り要件

通常経路は clinic スコープの bulk-read を使う。ただし production には bulk failure 後の per-owner day-log / owner / tag-cache read fallback があり、owner 数線形となる既知の degraded-mode gap がある。実行上限・metrics を備えた bounded fallback にするか、source fallback を除去するまで unconditional な no-N+1 invariant は主張しない。opt-out・suppression・daily-claim の意味論と bounded memory は維持する。

---

## 6. 運用と効果測定

- **配信監視 (`/lstep/delivery-monitor`)**: **自動配信トリガー**の実行ログ、除外理由、API 失敗を監視（ordinary タグ同期の request-local 経路は対象外）。
- **来院転換分析**: メッセージ配信から 30 日以内の来院率を自動集計。

---
