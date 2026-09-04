# LINE・Lステップ 運用コスト分析 (Cost Analysis)

> **目的**: LINE/Lステップ配信コストとROIの経済性分析を提供する。
> **読者**: 経営層・営業担当。
> **タイミング**: コスト試算・配信プラン検討時。

> **Animal Ekarte**: 自動マーケティング施策に伴う配信コストの予測と最適化
> **最新更新**: 2026-07-10

---

> **注記 (2026-08-01)**: 以下の配信ボリューム試算はフル稼働時の想定シミュレーションである。試算値は稼働前提のモデルであり、対象環境の実測ボリュームではない。
>
> - **Deploy gate（`LSTEP_WRITE_API_ENABLED`）**: Write API（タグ付与・タグ解除・プロパティ更新）が OFF のとき、クライアントは **`ErrWriteDisabled` を返し HTTP を送らない（`nil` 成功にしない）**。したがって deploy gate が OFF の環境では Lステップ側への実 write 呼び出しは契約上発生しない（実測データの主張ではない。環境変数の実値は記載しない）。
> - **Clinic gate（`is_sync_enabled` / API キー）**: 同期無効または API キー未設定の clinic はクライアント未構築による意図的スキップ（`nil, nil`）であり、deploy gate とは**別契約**である。
> - enable / stop / rollback の運用手順正本: [`docs/ops/deploy/LSTEP_WRITE_API_PAUSE.md`](../../ops/deploy/LSTEP_WRITE_API_PAUSE.md)（本 doc に手順を複製しない）。

---

## 1. 課金体系の概要

本システムの LINE 連携機能を利用するにあたり、以下の 2 つの外部コストが発生します。

1.  **LINE 公式アカウント (Messaging API)**: 月額固定費 ＋ 無料枠超過分の従量課金。
2.  **Lステップ (拡張ツール)**: プランに応じた月額固定費。

---

## 2. 配信ボリュームの予測 (シミュレーション)

一ヶ月あたりの配信件数を以下の 3 つのカテゴリで試算します。

### 2.1 高優先度・必須配信 (再診・予防)
- **再診リマインド**: 月間来院数の約 40% に対して 1 通。
- **ワクチン期限案内**: 全登録患者の約 10% に対して 1 通。
- **試算**: 月間 500〜800 通程度。

### 2.2 関係構築・CRM (自動ステップ)
- **初診お礼**: 月間新規飼主数に対して 1 通。
- **誕生日祝い**: 全登録患者の 1/12 に対して 1 通。
- **試算**: 月間 200〜400 通程度。

### 2.3 健診抽出・キャンペーン (手動バルク)
- **定期健診案内**: `/lstep/checkup-sync` による抽出対象への一括付与。
- **試算**: 実施月のみ 1,000 通以上。

---

## 3. 現行の配信抑制と safety gap

- **同日重複抑制**: 同日に複数 trigger が競合する場合は priority と daily claim で 1 件を選ぶ。
- **opt-out の ownership**: カルテ側は `owners.lstep_opt_out` を所有する。webhook は follow / unfollow のみを処理し、Lステップ側 opt-out property とこの flag を同期する契約はない。
- **現行の最終除外確認**: owner opt-out、`delivery_excluded`、LINE ID、cached EXCL tag を確認する。

**既知の source gap**: 最終 claim / write 前の除外確認は pet death を直接読まない。全 pet 死亡時の tag cleanup は best-effort であり、durable exclusion flag を確立しない。一部候補 query も死亡 pet を除外しない。そのため死亡 pet 関連配信の絶対遮断や「物理同期」は主張しない。

必要な是正契約は、候補取得時と最終 claim / write transaction の両方で durable な owner/pet exclusion evidence を fail-closed に確認し、failure alert、reconciliation、手動停止 fallback を持つこと。この source defect は本 docs-only 変更では未修正。

## 4. 実装済みの効果測定

`/lstep/analytics` が提供するのは trigger 別の delivery count / status と、配信後 30 日以内の `delivered_count`, `visited_count`, `visit_rate` である。売上額、診療費、message 単位 revenue、明示的な dormant-return KPI は計算しない。

### 計画中の ROI 指標（未実装）

- **revenue attribution**: 対象 delivery と 30 日以内の completed accounting を clinic / owner 単位で重複なく結び、売上額と attribution rule を返す。timezone、refund、複数 delivery の帰属 rule を acceptance criteria に含める。
- **dormant return**: dormant trigger 配信時の stage と、その後の qualified visit を記録し、期間・重複排除 rule とともに復帰数を返す。

これらを API、UI、test、計測定義が備わる前に ROI の実装済み指標として扱わない。

## 5. 料金 worksheet

外部 provider の plan 名、価格、無料枠、超過単価は repository 外で変更され得るため、本仕様は特定 plan を推奨しない。検討時は LINE / Lステップの公式 pricing URL を運用記録に添付し、**確認日、契約地域、通貨、税区分、登録数、月間配信数、超過単価**を記録して比較する。provider の最新条件を確認せず、この simulation だけで契約を決めない。
