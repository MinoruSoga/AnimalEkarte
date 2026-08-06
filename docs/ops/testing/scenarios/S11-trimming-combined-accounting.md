# S11: トリミング業務と診察併用精算

> **目的**: トリミング予約→トリミングカルテ記録→同日の診察カルテとの併用精算で、同一ペットの未請求明細（診察処置＋トリミングコース・オプション）が 1 つの会計に自動統合され合計金額が一致すること、を納品前に証明する。トリミングカルテ open は受付済のみ（AppointmentCard）。
> **所要目安**: 30分 / **深度**: 深い
> **仕様正本**: [screens/16-trimming-list.md](../../../spec/screens/16-trimming-list.md)・[screens/17-trimming-form.md](../../../spec/screens/17-trimming-form.md)・[screens/10-accounting-list.md](../../../spec/screens/10-accounting-list.md)・[screens/11-accounting-detail.md](../../../spec/screens/11-accounting-detail.md)・[reservation-to-record-flow.md §5.2-G/§8/§10](../../../spec/reservation-to-record-flow.md)

## 前提条件

- 環境: ローカル（seed 003_demo）。ログイン: `trimming` と `accounting` の view/create/edit 権限を持つロール（vet 等）。
- 対象ペット: ペット検索で「ステータス=生存」かつ当日の未精算会計・未請求明細がないペットを 1 頭選ぶ。同じ飼主に 2 頭目の生存ペットがいること（異常系 A2 で使用）。
- トリミングコースマスタ・オプションマスタに有効（active）で価格が 0 円でない項目が各 1 件以上あること。
- 依存シナリオ: なし（S01 で死亡登録したペットは使用しない）。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | トリミング一覧から新規登録 → 対象ペットを選択し、ステータス「予約」・コース 1 件・オプション 1 件を選択して保存する（[17 §1.1](../../../spec/screens/17-trimming-form.md)・#233） | 保存成功。コース・オプションにマスタ定義の価格が表示される（[17 §2.1](../../../spec/screens/17-trimming-form.md)）。トリミング一覧に「予約」で表示され、当日 appointment として受付カンバン「受付予約」列にも現れる（[16 §2](../../../spec/screens/16-trimming-list.md)・[flow §3.4/§3.7](../../../spec/reservation-to-record-flow.md)） |
| 2 | 受付カンバンで当該トリミングカードが「受付済」であることを確認し、そのカードからトリミングカルテを開いてスタイル要望・体重を追記保存する（「診療中」への遷移はカルテ作成の結果であり、open の前提ではない） | トリミングカードからは通常カルテではなくトリミングカルテへ遷移する（[flow §6 受付カンバン](../../../spec/reservation-to-record-flow.md)）。**受付済列からのみ open 可**（診療中 unrestricted open は実装に無い）。カルテ作成が診療中への契機 |
| 3 | 同日に同じペットの通常カルテを新規作成し、価格を持つ処置を 1 件以上入力して保存する | トリミングと診察は別 appointment として並存し、受付カンバンに 2 枚のカードが表示される（[flow §5.2-G](../../../spec/reservation-to-record-flow.md): 併用は appointments を 2 件作成）。トリミング記録の存在が診察カルテ入力を妨げない |
| 4 | 通常カルテの会計確認を確定し、受付カンバンでトリミングカードを「会計待ち」へ進める | 明示的な完了操作により双方の appointment が「会計待ち」になる（[flow §8 ステータス連動](../../../spec/reservation-to-record-flow.md)）。受付カンバンの「会計待ち」列に 2 カード |
| 5 | 会計一覧の「新規会計登録」から対象ペットを選択し、会計作成画面（`/accounting/new?petId=xxx`）を開く（[11 概要](../../../spec/screens/11-accounting-detail.md)） | **診察の処置明細とトリミングのコース・オプション明細が、未請求明細として 1 つの会計に自動で統合して取り込まれる**。新規会計 FE は `GET /api/v1/billing-items/unbilled-details?pet_id=`（BUG-013 envelope: `items` + typed `warnings`）を使用。BE は treatments / trimming（UNION ALL）/ vaccinations を pet 単位で pull 統合（`billing_item_service` + `FindUnbilledTrimmingItemsByPetID`）— #245 調査結果 (1) の証明 |
| 6 | 計算サマリを確認する | 税込合計 = コース価格 + オプション価格 + 処置明細合計（税込）と一致する（[11 §1.1/§3.1](../../../spec/screens/11-accounting-detail.md): 集中計算エンジン `calculateBillingTotals`） |
| 7 | 支払方法を選択して精算を完了する | ステータスが「精算済」になり、会計一覧に精算済・支払方法付きで表示される（`POST /accountings/complete`）。blocking unbilled warning がある pet は確定 disabled / BE Conflict（BUG-013） |
| 8 | 再度 `/accounting/new` を同じペットで開く | 未請求明細が空で、精算済みの処置・コース・オプションは再表示されない（[flow §10 補足](../../../spec/reservation-to-record-flow.md): `billing_items.treatment_id` / `trimming_course_id` / `trimming_option_id` 紐付けで再表示を防止）— 双方のカルテが請求済みであることの証明 |
| 9 | 受付カンバンとトリミング一覧を確認する | 会計完了により同日同一飼主・ペットの「会計待ち」appointment がまとめて「会計済」へ進む（[flow §10 補足](../../../spec/reservation-to-record-flow.md)）。【要実測】トリミング一覧のステータスバッジが「完了」になること（一覧バッジと appointment status の対応は仕様に明記なし）。**runtime 2026-08-01**: 一覧に『完了』バッジは観測（履歴）。会計完了→バッジ更新の end-to-end は DEFER |

## 確認観点

- 未請求明細の取得はペット単位。**現行 FE の新規会計は `GET /api/v1/billing-items/unbilled-details?pet_id=`**（legacy 生配列 `GET /billing-items/unbilled` は非移行 caller 用に残存）。飼主単位ではない（[16 §2](../../../spec/screens/16-trimming-list.md)）。施術完了を契機に会計側へ push する仕組みではなく、会計作成時に pull で取り込まれる。
- **BUG-013 underbilling 防止**: unbilled-details の `warnings` に `blocking: true` がある場合、または details 取得失敗時は新規会計の確定が無効化される（`AccountingDetail` / `AssertNoBlockingUnbilled`）。silent partial success はしない。
- トリミング明細が未請求候補になるのは appointment が「会計待ち（accounting）」に進んだ後、診察処置は会計確認の確定後 — 手順 4 を飛ばすと手順 5 で明細が現れないのは仕様どおり（[flow §8/§10 補足](../../../spec/reservation-to-record-flow.md)）。
- 金額計算は臨床・会計・印刷帳票で同一ロジック `calculateBillingTotals` を共有し、レイヤー間で 1 円の誤差も生じない（[11 §3.1](../../../spec/screens/11-accounting-detail.md)）。
- clinic_id 隔離: 明細・会計・appointment がすべて同一 clinic に属すること。会計確定が `audit_logs` に記録されること — DB 参照は USER 実施。
- ルート: トリミング一覧 `/trimming`、会計新規 `/accounting/new?petId=`。

## 異常系

| # | 操作 | 期待結果 |
|:--|:--|:--|
| A1 | トリミングのみ「会計待ち」（診察カルテは会計確認前）の状態で会計を作成し精算する。その後、診察側の会計確認を確定して再度会計を作成する | 1 回目はトリミング明細のみで単独精算でき、2 回目は処置明細のみが未請求として現れる — 診察だけ先・トリミングだけ後の分割会計が引き続き可能（[flow §5.2-G-6](../../../spec/reservation-to-record-flow.md): 分割会計を許可） |
| A2 | 同じ飼主の別ペットに同日の未請求明細（処置またはトリミング）を作った上で、対象ペットの会計新規作成を開く | 別ペットの明細は混入しない（統合対象は「同日同一飼主・ペット」単位 — [flow §10-6](../../../spec/reservation-to-record-flow.md)。未請求明細 API が pet_id 単位のため）。【要実測】仕様文書に明記された否定条件ではないため初回実行で確認し期待結果へ昇格する。**runtime 2026-08-01 BLOCKED**: 同日 multi-pet 未請求 fixture 未作成 |
| A3 | 手順 5 の統合された会計から一方の明細だけを残して精算する | 【要実測】会計詳細画面での明細削除による分割精算の UI 経路と、削除した明細が未請求へ戻ること（[flow §5.2-G-6](../../../spec/reservation-to-record-flow.md) は分割会計を許可するが UI 導線は仕様に明記なし）。**runtime 2026-08-01 BLOCKED**: 統合 multi-detail 会計 fixture 未作成 |

## 実装突合
- 突合日: 2026-08-07
- HEAD: 844e43f69
- 変更:
  - 新規会計 FE の未請求取得を `unbilled-details`（BUG-013 items+warnings）に更新
  - blocking unbilled で確定拒否・AssertNoBlockingUnbilled を確認観点に追加
  - 精算を `POST /accountings/complete` に合わせて記載
  - trimming UNION ALL 取得・pet_id 単位 pull は現行 main で再確認

- runtime 2026-08-07: **BLOCKED** — authenticated UI requires E2E_LOGIN_* or non-empty DEV_ADMIN_* in host `.env.local` (currently empty). Stack healthy :3003/:8080.
