# UAT-R2-MASTER-PATH: マスタ登録と会計導線の再現票

状態: **医院の操作情報待ち（UNKNOWN）**。これは既知の実装経路を照合するための票であり、医院での再現結果や修正完了を示さない。元の課題は「マスタ入力後に金額が空、会計画面に出ない」だが、登録画面・マスタ種別・単価・操作順はまだ特定されていない（[元 TODO](../../../todo-issue.md#uat-r2-master-path)、[UAT フィードバック](../stg-uat-clinic-feedback-q1-q4.md#uat-r2-master-path)）。

## 既知の期待経路

| 対象 | マスタ登録後の期待経路 | 照合根拠 |
| --- | --- | --- |
| 商品 | 会計の「マスタから選択」には有効な商品マスタが表示される。商品を選ぶと単価などを会計明細へ渡し、`merchandise_item_id` を含む請求明細を作る。 | [商品一覧と有効状態の絞り込み](../../../frontend/src/features/accounting/components/ItemListCard.tsx)、[商品取得 API](../../../frontend/src/features/accounting/api/get-merchandise-items.ts)、[追加処理](../../../frontend/src/features/accounting/hooks/use-accounting-item-actions.ts)、[請求明細 API](../../../frontend/src/features/accounting/api/create-billing-item.ts) |
| 診察・処置・薬剤などの治療 | 該当マスタをカルテ治療で選び、治療を保存する。確定カルテの未請求治療が請求候補へ流れ、`treatment_id` を持つ明細として扱われる。商品用の直接選択欄へ出す経路ではない。 | [治療マスタ取得](../../../frontend/src/hooks/use-treatment-master.ts)、[カルテ治療選択](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts)、[治療 API](../../../frontend/src/features/medical-records/api/treatments.ts)、[未請求治療の抽出条件](../../../backend/internal/medicalrecord/treatment_repository.go)、[請求候補への変換](../../../backend/internal/billing/billing_item_unbilled.go) |

商品マスタとカルテ治療の単価は負数を拒否し、0 を許す（[商品入力と保存](../../../backend/internal/inventory/merchandise_item_request.go)、[商品サービス](../../../backend/internal/inventory/merchandise_item_service.go)、[治療サービス](../../../backend/internal/medicalrecord/treatment_service.go)）。ただし、請求チェックの検査・ワクチン候補で `null`・非有限・負数を「価格未設定」とする判定は、全治療 DTO の単価が nullable という意味ではない（[候補判定](../../../frontend/src/features/medical-records/lib/medical-record-bill-check-model.ts)、[元 TODO](../../../todo-issue.md#uat-r2-master-path)）。診療項目の単価は全タブ保存、課税区分と税率の保存は診察・処置のみという仕様も、対象タブを特定してから照合する（[診療項目マスタ仕様](../../../docs/spec/screens/settings/master-treatment.md)）。

## 医院回答を入れて照合する表

| ケース | 期待値（既知の仕様から） | 医院での実際値 | 判定に必要な入力 |
| --- | --- | --- | --- |
| 商品マスタを登録・再読込 | 保存した商品単価が再読込後も維持され、有効な商品なら会計の直接選択に現れる。 | **UNKNOWN** | 登録画面、商品マスタ ID、入力単価、保存 request/結果、再読込後の単価と有効状態、会計での操作順と表示結果 |
| 診察・処置・薬剤などを登録・再読込 | 対象マスタの単価が保存・再読込され、カルテ治療に選択した後、確定カルテの未請求候補を経て請求明細になる。 | **UNKNOWN** | 登録画面とマスタ種別/ID、入力単価、保存 request/結果、再読込後の単価、カルテ治療/確定/会計の操作順と表示結果 |
| 会計の「マスタから選択」で商品以外を探す | この欄は商品マスタ経路であり、治療マスタはカルテ経由で確認する。 | **UNKNOWN** | 実際に開いた会計画面、検索したマスタ種別/ID、操作順、医院が期待した結果 |

この票に共通して未確認の項目: **医院が使った登録画面 UNKNOWN、マスタ種別 UNKNOWN、入力額 UNKNOWN、保存後・再読込後の額 UNKNOWN、会計操作順 UNKNOWN、実際の表示・請求結果 UNKNOWN、医院の期待結果 UNKNOWN、要件責任者（個人名）UNKNOWN、業務目的 UNKNOWN**。患者・診療データをこの票へ転記せず、同一 ID を用いた診断に必要な操作と金額の事実だけを収集する。

## 次の判定ゲート

1. 医院から上記の画面、マスタ種別/ID、入力額、保存 request と再読込後の額、カルテ・会計の操作順、期待結果、要件責任者の個人名と業務目的を受け取る。現時点の原因は **UNKNOWN**。
2. 同一マスタ ID で保存 request → 保存結果 → 再読込した単価を照合する。ここで不一致なら **単価保存・再読込の失敗** として切り分け、該当経路の失敗テストと最小修正を検討する。
3. 単価が一致する場合は、対象が商品か治療かを確認し、上表の正しい会計経路と医院の操作を照合する。カルテ経由が正しく、会計の商品選択だけを探していたなら **操作経路の案内** を検討する。カルテ確定後の未請求または明細で不一致なら、その同一 ID の連携を調べる。
4. 個人名を持つ要件責任者と業務目的が確定するまでは機能変更を決めない。全マスタを会計の直接選択に混在させる案は、現行の経路と二重管理防止の方針に合わない（[UAT フィードバック](../stg-uat-clinic-feedback-q1-q4.md#uat-r2-master-path)、[設計思想](../../product-philosophy.md)）。

完了条件: 同一 ID の対象経路・期待値・実際値・原因・最小修正または案内を根拠付きで記録すること。医院の入力がそろうまでは製品修正と医院 UAT 判定を **BLOCKED** とする（[元 TODO](../../../todo-issue.md#uat-r2-master-path)）。
