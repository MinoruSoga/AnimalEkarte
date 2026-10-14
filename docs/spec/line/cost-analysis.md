# LINE・Lステップ 運用コスト分析 (Cost Analysis)

> **目的**: LINE/Lステップ配信コストとROIの経済性分析を提供する。
> **読者**: 経営層・営業担当。
> **タイミング**: コスト試算・配信プラン検討時。

> **Animal Ekarte**: 自動マーケティング施策に伴う配信コストの予測と最適化
> **最新更新**: 2026-07-10

---

> **注記 (2026-08-01)**: 追記（2026-10-07 監査）— 配信ボリューム試算（旧 §2）は作業ノート [../../work/analysis/line-cost-analysis-roi-draft-20261007.md](../../work/analysis/line-cost-analysis-roi-draft-20261007.md) へ移設した。試算はフル稼働時の想定シミュレーションであり、稼働前提のモデルであって対象環境の実測ボリュームではない（原本の免責は移設先に保持）。
>
> Deploy gate（`LSTEP_WRITE_API_ENABLED`）/ Clinic gate（`is_sync_enabled` / API キー）の二重 gate 契約本文（`ErrWriteDisabled` とクライアント未構築の意図的スキップ `nil, nil` は別契約）の正本は [architecture.md](./architecture.md) 冒頭注記・§4。運用手順正本: [`docs/ops/deploy/runbooks/LSTEP_WRITE_API_PAUSE.md`](../../ops/deploy/runbooks/LSTEP_WRITE_API_PAUSE.md)（本 doc に手順を複製しない）。

---

## 1. 課金体系の概要

本システムの LINE 連携機能を利用するにあたり、以下の 2 つの外部コストが発生します。

1.  **LINE 公式アカウント (Messaging API)**: 月額固定費 ＋ 無料枠超過分の従量課金。
2.  **Lステップ (拡張ツール)**: プランに応じた月額固定費。

---

## 2. 配信ボリュームの予測

> 追記（2026-10-07 監査）: 配信ボリュームの想定シミュレーション（高優先度・必須配信 / 関係構築・CRM / 健診抽出・キャンペーンの 3 カテゴリ試算とチャート）は業務仕様ではなく作業ノートとして [../../work/analysis/line-cost-analysis-roi-draft-20261007.md](../../work/analysis/line-cost-analysis-roi-draft-20261007.md) へ移設した。本書は課金体系（§1）と実装済みの効果測定（§4）を仕様として保持する。

---

## 3. 現行の配信抑制とコストへの影響

- **opt-out の ownership**: カルテ側は `owners.lstep_opt_out` を所有する。webhook は follow / unfollow のみを処理し、Lステップ側 opt-out property とこの flag を同期する契約はない（2026-10-07 監査: 本書にしか記載のない契約のため保持）。
- 最終配信除外確認と同日の優先度抑制により**配信されなかった飼主分は、Messaging API の配信課金の対象にならない**。

追記（2026-10-07 監査）: 配信安全ガードの契約本文（最終除外確認の対象一覧・同日 claim と優先度抑制・既知の safety gap と是正契約）の正本は [lstep-integration.md §4](./lstep-integration.md)。本書では重複掲載をやめ、コスト上の含意のみを残す。

## 4. 実装済みの効果測定

`/lstep/analytics` が提供するのは trigger 別の delivery count / status と、配信後 30 日以内の `delivered_count`, `visited_count`, `visit_rate` である。売上額、診療費、message 単位 revenue、明示的な dormant-return KPI は計算しない。

### 計画中の ROI 指標（未実装）

> 追記（2026-10-07 監査）: revenue attribution・dormant return の acceptance criteria draft と「API、UI、test、計測定義が備わる前に実装済み指標として扱わない」注記は作業ノート [../../work/analysis/line-cost-analysis-roi-draft-20261007.md](../../work/analysis/line-cost-analysis-roi-draft-20261007.md) へ移設した。実装済みの効果測定の仕様は本節の上記本文のみ。

## 5. 料金 worksheet

> 追記（2026-10-07 監査）: 料金比較 worksheet の記録方法（公式 pricing URL・確認日・契約地域・通貨・税区分・登録数・月間配信数・超過単価の記録、特定 plan を推奨しない方針、simulation だけでの契約決定禁止）は作業ノート [../../work/analysis/line-cost-analysis-roi-draft-20261007.md](../../work/analysis/line-cost-analysis-roi-draft-20261007.md) へ移設した。課金体系の仕様（§1）は本書に留まる。
