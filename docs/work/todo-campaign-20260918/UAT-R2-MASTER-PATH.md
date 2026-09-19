# UAT-R2-MASTER-PATH: マスタ登録と会計導線の再現票

状態: **全経路の検証設計 READY／テスト未実行**。2026-09-19に依頼者が「可能性のあるページをすべて検証」と回答したため、発生ページの回答待ちを解除する。金額を入力する全マスタの新規/編集と、利用先・会計までを検証する。元の症状がどの画面で発生したかは未確定であり、全経路の検証結果と医院での再現事実を混同しない（[要件・状態の正本](../../../todo-issue.md#uat-r2-master-path)、[元報告](../stg-uat-clinic-feedback-q1-q4.md#uat-r2-master-path)）。

## 既知の期待経路

| 対象 | マスタ登録後の期待経路 | 照合根拠 |
| --- | --- | --- |
| 商品 | 会計の「マスタから選択」には有効な商品マスタが表示される。商品を選ぶと単価などを会計明細へ渡し、`merchandise_item_id` を含む請求明細を作る。 | [商品一覧と有効状態の絞り込み](../../../frontend/src/features/accounting/components/ItemListCard.tsx)、[商品取得 API](../../../frontend/src/features/accounting/api/get-merchandise-items.ts)、[追加処理](../../../frontend/src/features/accounting/hooks/use-accounting-item-actions.ts)、[請求明細 API](../../../frontend/src/features/accounting/api/create-billing-item.ts) |
| 診察・処置・薬剤などの治療 | 該当マスタをカルテ治療で選び、治療を保存する。確定カルテの未請求治療が請求候補へ流れ、`treatment_id` を持つ明細として扱われる。商品用の直接選択欄へ出す経路ではない。 | [治療マスタ取得](../../../frontend/src/hooks/use-treatment-master.ts)、[カルテ治療選択](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts)、[治療 API](../../../frontend/src/features/medical-records/api/treatments.ts)、[未請求治療の抽出条件](../../../backend/internal/medicalrecord/treatment_repository.go)、[請求候補への変換](../../../backend/internal/billing/billing_item_unbilled.go) |

商品マスタとカルテ治療の単価は負数を拒否し、0 を許す（[商品入力と保存](../../../backend/internal/inventory/merchandise_item_request.go)、[商品サービス](../../../backend/internal/inventory/merchandise_item_service.go)、[治療サービス](../../../backend/internal/medicalrecord/treatment_service.go)）。ただし、請求チェックの検査・ワクチン候補で `null`・非有限・負数を「価格未設定」とする判定は、全治療 DTO の単価が nullable という意味ではない（[候補判定](../../../frontend/src/features/medical-records/lib/medical-record-bill-check-model.ts)、[元 TODO](../../../todo-issue.md#uat-r2-master-path)）。診療項目の単価は全タブ保存、課税区分と税率の保存は診察・処置のみという仕様も、対象タブを特定してから照合する（[診療項目マスタ仕様](../../../docs/spec/screens/settings/master-treatment.md)）。

## 全ページの対象一覧

開始時点の母集団は **11単価フォーム＋1割引フォーム**。下表を [route定義](../../../frontend/src/config/paths.ts)・[分類設定](../../../frontend/src/features/master/constants/category-config.ts)・実フォームと突合し、追加の金額経路があれば同じ票へ足す。`showPrice=false` だけで除外しない（ケージは単価入力を持つ）。金額のないフォームは理由付きN/Aとし、V04全CRUDをこの課題の範囲へ混ぜない。

| ページ/フォーム | 入力根拠 | 保存後に追う経路 |
| --- | --- | --- |
| `/settings/treatment-items?tab=consultation`（診察） | [診療項目request変換](../../../frontend/src/features/master/routes/treatment-plan-master-model.ts) | カルテ治療→医師確認→確定→未請求→会計 |
| 同 `tab=examination`（検査） | 同上 | 治療として選ぶ経路と、検査作成→医師確認の検査候補→会計を別ケース |
| 同 `tab=procedure`（処置） | 同上 | カルテ治療→医師確認→確定→未請求→会計 |
| 同 `tab=vaccine`（予防接種） | 同上 | 治療として選ぶ経路と、接種作成→医師確認の接種候補→会計を別ケース |
| 同 `tab=checkup`（定期健診） | 同上 | 治療候補/健診作成の実際の価格参照を追跡。会計連携の有無はsourceで確認 |
| `/settings/medicine`（薬剤・分類/明細を区別） | [薬剤単価欄](../../../frontend/src/features/master/components/MedicineSidePanelSections.tsx) | 治療検索→カルテ治療→会計。分類行の価格0を保存失敗と誤判定しない |
| `/settings/merchandise-items`（商品） | [商品フォーム](../../../frontend/src/features/master/components/MerchandiseSidePanel.tsx) | 会計の商品選択→明細→保存/再読込 |
| `/settings/hospitalization`（入院プラン） | [入院フォーム](../../../frontend/src/features/master/components/HospitalizationSidePanel.tsx) | 入院のプラン/料金単位/日数→会計参照の有無を追跡 |
| `/settings/cage`（ケージ） | [ケージ単価欄](../../../frontend/src/features/master/components/CageSidePanel.tsx) | 入院のケージ料金参照とプラン料金の関係を追跡。合算を推測しない |
| `/settings/trimming?tab=course`（コース） | [コースフォーム](../../../frontend/src/features/master/components/TrimmingCourseSidePanel.tsx) | トリミング選択→会計。予約側の価格表示はlocal合成データで確認 |
| `/settings/trimming?tab=option`（オプション） | [オプションフォーム](../../../frontend/src/features/master/components/TrimmingOptionSidePanel.tsx) | トリミング追加料金→会計。外部予約/LINE送信は行わない |
| `/settings/campaigns`（割引額/率） | [割引フォーム](../../../frontend/src/features/master/components/CampaignSidePanel.tsx) | 保存/再読込と適用先の割引額。単価と別軸で照合 |

下流はカルテ治療・検査・予防接種・定期健診・医師確認・見積、入院、トリミング、会計新規/詳細を対象にする。会計連携の参照型は [明細作成](../../../frontend/src/features/accounting/hooks/create-accounting-items.ts) にある `treatment_id` / `exam_id` / `vaccination_id` / `trimming_course_id` / `trimming_option_id` 等を確認する。未実装の自動連携を本検証で新仕様として追加しない。

## 実行手順・期待値

1. 開発/QAが上表を新規/編集×保存/再読込/下流のケースに展開し、既存V04・form・APIテストで覆う箇所を紐付ける。候補revision、Docker mount、専用合成clinic/患者、実行者、後処理、receipt保存先を固定する。借用した他タスクのDBや実請求を使わない。
2. 通常単価1,200円を新規保存→ページを開き直す→2,300円へ編集→再読込する。正常系とは別に0、空欄、負数、取消/保存失敗を各フォームの既存契約で照合する。空欄/NULL/0を一律同値にしない。数量2の行は単価と行金額を分け、税込/税率/丸めは各仕様に従う。キャンペーンは単価ではなく割引額/率として期待値を作る。
3. request→response→再取得API→表示値→下流ID/金額の最初の不一致を特定する。マスタ編集が過去の確定明細を遡及変更しないこと、新規選択の価格は適切に更新されることも確認する。
4. 不一致ごとに失敗テスト→最小修正→Docker scoped検証。商品以外を商品選択欄へ追加する改変で辻褄を合わせない。保存が正常でも該当下流の検証を省略しない。

成果物列: `route/tab / create-or-update / fixture / input / request / reread / downstream / expected / actual / evidence / cleanup / PASS-FAIL-BLOCKED-N/A`。実際値は全行 **未実行** から開始。完了は全対象がPASSまたは根拠付きN/Aで、価格消失/不一致のFAIL・未実行が残らないこと。新しい製品仕様が必要なときだけ要件責任者の判断へ戻す。医院の元の操作特定は、合成検証を止める前提にしない。
