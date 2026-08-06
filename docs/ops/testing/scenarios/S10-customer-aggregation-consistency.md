# S10: 顧客集計ダッシュボード整合

> **目的**: 顧客集計ダッシュボードの主要指標（LTV・来院頻度・CPM 分布）を会計実データと突合し、`medical_record_id` のない完了済み手動会計が LTV に算入され、来院回数には影響しないことと、見積が売上に混入しないことを証明する。
> **所要目安**: 15分 / **深度**: 薄い
> **仕様正本**: [顧客分析・集計機能](../../../spec/customer-aggregation.md) / [顧客集計ダッシュボード](../../../spec/screens/36-aggregation-dashboard.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。ログイン: 権限グループ「執行」のスタッフ（owners の view を保有）。
- **依存シナリオ**: S08 の 1 件目の全額精算まで実行済みであること。S08 精算後の飼主を「飼主 A」として使用する。S08 の新規会計作成 payload は `medical_record_id` を送らない（手動会計）。
- **seed の実データ（2026-08 突合）**:
  - `medical_records.csv` は **大量の履歴あり**（主に八王子 clinic_id=1。ヘッダのみではない）。
  - `billings.csv` は pending のみ（completed 0 件・`completed_at` なし）→ **LTV 金額は S08 で作った完了会計が必須**。
  - `estimates.csv` は八王子に draft 見積が多数（承認済みでも LTV 非算入であることの対照は、飼主 B で見積を 1 件用意して確認）。
- **飼主 B（見積のみ・データ準備要）**: 過去 365 日に completed 会計がない飼主を 1 名選び、見積を 1 件作成（または seed draft 見積の飼主を使う）。会計一覧で完了会計が無いことを事前確認。
- 「1年以上」検証: 八王子では seed 来院履歴があるため最終来院タブの over_1y が埋まることがある。城東中心で実施する場合はデータ準備または「来院なし」中心で確認。
- CSV 出力（#10）はダウンロード発生の確認までとし、ファイル内容の全列検証は行わない（数式インジェクション中和は SEC 修正済み — 内容監査は任意）。
- **期間既定は売上タブ当年度（`year: CURRENT_YEAR` = JST 年）**。#5 は同じ期間軸に揃える。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:---|:---|
| 1 | 顧客集計ダッシュボード `/aggregation` を開く | 売上ランキングタブが既定表示され、CPM セグメント 6 段階（Encounter/Growing/Core/Spot/Noah/Dormant）の人数チップが表示される。長時間ハングせず、最悪 25s 前後でエラー表示に落ちること（BUG-012） |
| 2 | CPM チップのいずれかをクリック | 一覧が該当セグメントの飼主に絞り込まれる |
| 3 | 「来院回数」「最終来院」タブへ順に切替 | タブに応じて列構成が変わる（来院回数系の列／最終来院日・経過日数の列） |
| 4 | 売上ランキングタブで飼主 A を探し、年間診療費と来院回数を確認する | S08 の `medical_record_id` なし完了会計が年間診療費に算入される。手動会計自体は来院として数えられず、来院回数は `medical_records` に記録された診療日の件数から変わらない |
| 5 | 会計一覧 `/accounting` で飼主 A・診療日=対象年（または過去 365 日）・ステータス「精算済」でフィルタし、請求金額を合算 | 合算額が #4 の年間診療費と一致する（`medical_record_id` の有無を問わず、期間内の完了会計を算入）。UI ラベルは「精算済」（completed） |
| 6 | 売上ランキングタブのフィルタパネルで金額基準を「売上総額」→「入金額」→「返金控除後」と切替 | 基準は `gross_total_amount` / `paid_amount` / `net_paid_amount`。`medical_record_id` なし会計も各実データに応じて算入される |
| 7 | 売上ランキングタブで飼主 B を確認 | 見積の金額は年間診療費に算入されず、売上ゼロ（またはランキング対象外）のまま（見積≠会計） |
| 8 | 最終来院タブのフィルタで区分「1年以上」を選択 | 一覧が「1年以上」（`over_1y`）の飼主のみに絞り込まれる |
| 9 | 絞り込んだ一覧の任意の飼主数名の経過日数を確認 | いずれも 365 日以上であり、最終来院日と矛盾しない |
| 10 | 一覧で飼主 A の行をチェックし CSV 出力を実行 | 表示中のリスト内容が CSV としてダウンロードされる（仕様正本 36 §1.3） |
| 11 | 一覧の飼主名（飼主 A）をクリック | 飼主詳細 `/owners/:id` へ遷移し、会計履歴セクションで #5 と同じ会計が確認できる（ドリルダウン） |

## 確認観点

- **LTV と来院回数の対象分離**: LTV は医院・飼主が一致する完了（completed）会計を `billings` から集計し、`medical_record_id` のない手動会計も含む（`ltv_repository.go` の ba サブクエリ）。期間判定は `COALESCE(bmr.date, b.scheduled_date)`。payments 集計は **billings.clinic_id で scope**（BUG-012）。来院回数・最終来院日は論理削除されていない `medical_records` の診療日だけを数え、手動会計では増えない。見積（estimates）は売上に算入されない。
- **タイムアウト境界（BUG-012）**: BE `ListOwnerAggregation` は `aggregationQueryTimeout = 20s`（`backend/internal/lstep/aggregation_service.go`）。FE axios は一覧・CPM 人数とも `timeout: 25_000`（`get-aggregations.ts` / `get-cpm-stage-counts.ts`）。無限 loading は不具合、タイムアウト後のエラー表示は正常。
- **最終来院区分の境界**: 90/180/365 日は `ltv_repository.go` の SQL 固定値。区分は within_3m / over_3m / over_6m / over_1y / no_visit（「3ヶ月未満／3ヶ月以上／6ヶ月以上／1年以上／来院なし」）。
- **並行ロード**: CPM チップ（6 並列 total クエリ）と一覧テーブルは別クエリで非同期取得される。片方が先に表示される瞬間があっても不具合ではない。
- **CPM「Dormant」チップとの区別**: CPM 休眠は `cpm_v1_dormant_days`（既定 240 日）の別軸。#8 の「1年以上」（365 日固定）と件数不一致は不具合ではない。
- **CPM チップの固定**: チップは CPM V1 の 6 区分に固定で、医院設定の CPM バージョンが V2 でも切り替わらない（仕様正本 36 §1.2）。
- **来院なしの扱い**: 最終来院タブの「来院なしを含む」と `no_visit` は「1年以上」とは別分類。#8 の絞り込みで混入しないこと。
- **clinic_id 隔離**: 集計 API は `GET /api/v1/clinics/:clinic_id/owners/aggregations`（lstep ドメイン）。クリニック切替で他院の飼主が混入しないこと。

## 実装突合
- 突合日: 2026-08-07
- HEAD: 844e43f69
- 変更:
  - seed「臨床ヘッダのみ／見積なし」記述を撤回（medical_records 大量・estimates draft 多数・billings は pending のみ）
  - BUG-012: BE 20s / FE 25s timeout と payments clinic scope を確認観点に追加
  - 金額基準 API 値（gross/paid/net）と UI ラベル「精算済」を明記
  - LTV 期間軸 `COALESCE(bmr.date, b.scheduled_date)` を実装どおり記載
