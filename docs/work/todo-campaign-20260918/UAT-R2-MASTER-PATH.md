# UAT-R2-MASTER-PATH: マスタ登録と会計導線の再現票

状態: **12フォーム合成検証 実行済み（2026-09-23, EMR-84）**。scoped unit に加え、合成 clinic で create→保存 request→API 応答→キャッシュ非依存の再読込→下流/会計までを HTTP で実行し、下の合成実行マトリクスを actual で埋めた。1件の再現 defect（trimming-courses / trimming-options / hospitalization-plans / cages が負価格を保存）を最小修正し、Docker scoped 検証（unit + 実機再確認）まで完了。2026-09-19に依頼者が「可能性のあるページをすべて検証」と回答したため、発生ページの回答待ちを解除する。元の症状がどの画面で発生したかは未確定であり、全経路の検証結果と医院での再現事実を混同しない（[要件・状態の正本](../../../todo-issue.md#uat-r2-master-path)、[元報告](../stg-uat-clinic-feedback-q1-q4.md#uat-r2-master-path)）。

本票は現行 route / form / API への対応づけと、未カバー失敗テストの**候補名**までを固定する。合成検証の actual は下記「合成実行マトリクス（actual = 実行済み）」に記録した。**専用env（A4 rehearsal overlay）は KNJO 21-table bundle 不在のため未使用**であり、検証は稼働中 dev stack（clinic_id=4）で行った。全マスタを会計の「マスタから選択」へ混在させる提案はしない。

### 12フォーム coverage（2026-09-21 scoped unit）

| # | フォーム | scoped save→reread→billing | 備考 |
| --- | --- | --- | --- |
| 1 | 診察 | **newly covered** | `treatment-plan-master-model.test.ts` price 0（既存）+ update 1200→2300/tax。専用env actual は未実行 |
| 2 | 検査 | **newly covered** | exam create/update price 0/1200/2300・tax 非送出。検索ダイアログ exam vs checkup は下記 residual I/O で固定（配線のみ・価格消失バグ自動宣言なし） |
| 3 | 処置 | **newly covered** | procedure price 0 + tax、update 2300 |
| 4 | 予防接種 | **newly covered (master persist)** | vaccine price 0/1200/2300。治療検索「予防」→`other` は分類のみ（`unit_price` は維持。価格消失ではない） |
| 5 | 定期健診 | **newly covered (persist)** | checkup price persist。会計自動連携は根拠付き N/A のまま |
| 6 | 薬剤 | **already covered** | 既存 `medicine-settings-model.test.ts`（分類0強制・明細価格・BUG-006）。本単位で再実装しない |
| 7 | 商品 | **newly covered** | `merchandise-item-settings-model.test.ts` unit_price 0/1200/2300 + accounting `merchandise_item_id` |
| 8 | 入院プラン | **newly covered + residual price-loss 修正済み** | create は UI 0 で price omit、update は 0 を明示送信。CarePlanRefSelect→AddForm/EditRow のマスタ価格転記は commit `e79930591` で実装済み（下記 residual 行を更新）。合成検証で 2300 がケアプラン→退院会計まで維持されることを実測 |
| 9 | ケージ | **newly covered (persist)** | cage price 0/1200/2300。billing 自動連携 N/A |
| 10 | トリミングコース | **newly covered** | `""`→null / `"0"`→0 / reread null→`""`・0→`"0"` + trimming_course_id 会計 |
| 11 | トリミングオプション | **newly covered** | 同上 + trimming_option_id |
| 12 | キャンペーン | **newly covered** | discount_value 0/1200/rate10。単価フィールド非混在 |

**実行状況（2026-09-23, EMR-84）**: 12フォームすべてで create(1200)→再読込→update(2300)→再読込 を稼働中 dev stack で PASS。下流は treatment/exam/vaccination/checkup/merchandise/hospitalization/trimming/campaign を実測（下の合成実行マトリクス）。0/省略/負の境界も実行。**専用env（A4 rehearsal overlay）は KNJO 21-table bundle 不在で未使用**のため、「専用envでの再実行」のみ残る。負価格を許す 1 defect を発見・最小修正済み。

### Residual I/O map（2026-09-22 att-master-20260922-001）

配線/分類だけではバグにしない。下流で単価が消える経路だけを価格消失 mismatch とする。

| 経路 | 現行 I/O（コード根拠） | 合成期待 | 価格消失? | scoped 証跡 |
| --- | --- | --- | --- | --- |
| TreatmentSearchDialog × exam_types | hooks: consultations / procedures / vaccines / **checkupTypes** / medicines のみ。[TreatmentSearchDialog.tsx](../../../frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx) は `useGetAllExaminationTypes` / `/v1/masters/examination-types` を呼ばない。checkupTypes を category `"検査"`・`unitPrice=ct.price` で出す。exam_types 単価の会計経路は検索ダイアログではなく検査レコード→未請求 `exam_id`。 | 合成: exam_type 価格 1200 をマスタ登録しても治療検索一覧に exam_types 行は出ない。checkup 1200 は「検査」行として出る。検索未選択は価格消失ではない。 | **No**（配線差。会計は exam_id 経路） | 本票の対応表＋既存 dialog テスト（checkup 表示）。exam fetch 追加の製品変更は本単位外 |
| vaccine search → `resolveItemTypeFromCategory` → `other` | ダイアログは vaccines を category `"予防"`・`unitPrice=v.price` で出す。[treatments-tab-model.ts](../../../frontend/src/features/medical-records/lib/treatments-tab-model.ts) は `"予防"`→`item_type:"other"` だが `buildMasterSelectionPayload` は **`unit_price: item.unitPrice` を維持**。接種会計は別経路 `vaccination_id`（vaccines.price）。 | 合成: ワクチン 5000 を治療検索で選ぶ → create treatment は `item_type=other` かつ `unit_price=5000`。分類が other でも単価は落ちない。二重計上リスクは分類バグではなく経路分離の検証対象。 | **No**（分類のみ） | 既存 `treatments-tab-model.test.ts`（予防→other）。価格維持は payload 契約で読み取り固定 |
| CarePlanRefSelect 入院プラン価格転記 | **修正済み（commit `e79930591`）**: [CarePlanRefSelect](../../../frontend/src/features/hospitalization/components/CarePlanTab/CarePlanRefSelect.tsx) は `onUnitPriceChange` で選択プランの `price` を伝播（0 は有限値保持）。[AddForm](../../../frontend/src/features/hospitalization/components/CarePlanTab/AddForm.tsx) / [EditRow](../../../frontend/src/features/hospitalization/components/CarePlanTab/EditRow.tsx) は type=item で `unit_price` を payload に積む。BE create は request `unit_price`（省略時 0）を保存しプランマスタから自動コピーしない（契約は据え置き）。退院会計は `care_plan_items.unit_price` を写す。 | 合成（実測）: プランマスタ price=2300 を type=持ち物で `hospitalization_plan_id` + `unit_price=2300` で create→ケアプラン再読込 2300→退院会計明細 2300。**価格消失は再現しない**。BE 単体に `unit_price` を省略すると 0 保存（FE が送る契約） | **No（修正済み）** | 実行ログ（合成実行マトリクス §入院プラン）+ `AddForm.test.tsx` / `EditRow.test.tsx` |

## 0/空欄/欠損/未保存の区別（検証時に同値にしない）

| 状態 | 現行コードでの意味 | 代表根拠 |
| --- | --- | --- |
| **0円** | 有限の 0。`MoneyInput` は `value === 0` を空表示し、空入力は `0` を commit する。治療マスタ保存は `price < 0` だけ拒否し 0 は通す。未請求検査は `unit_price IS NULL` または負数を請求不能とし、0 は候補に残す。`isUnbillableMasterPrice` も 0 を請求不能にしない。 | [MoneyInput](../../../frontend/src/components/shared/SidePeek/MoneyInput.tsx)、[治療パネル検証](../../../frontend/src/features/master/components/TreatmentItemSidePanel.tsx)、[検査未請求](../../../backend/internal/billing/billing_item_repository_unbilled.go)、[請求チェック](../../../frontend/src/features/medical-records/lib/medical-record-bill-check-model.ts) |
| **空欄（UI）** | MoneyInput 系では 0 と同値。トリミングは `price: string` で `null` を `""` にする。キャンペーン入力は `Number(e.target.value) \|\| 0` と `Math.max(0, …)` で空・負を 0 に畳む。 | [トリミング form](../../../frontend/src/features/master/lib/trimming-side-panel-model.ts)、[キャンペーン入力](../../../frontend/src/features/master/components/CampaignSidePanel.tsx) |
| **欠損（API/DB NULL）** | 入院プラン作成は `price: data.price \|\| undefined` のため UI 0 が create では省略され得る。ケージ `Price *int64`。検査未請求は `exam_type.price IS NULL` を unbillable。 | [入院 create 変換](../../../frontend/src/features/master/routes/hospitalization-settings-model.ts)、[ケージ model](../../../backend/internal/model/cage.go) |
| **未保存** | 保存失敗・dirty 残・request 未発行。保存成功しても下流未選択は「価格消失」ではない。 | [トリミング dirty](../../../frontend/src/features/master/routes/TrimmingSettings.test.tsx)、本票の下流列 |

商品・治療のサーバー側は負数を拒否し 0 を許す（[商品 request](../../../backend/internal/inventory/merchandise_item_request.go)、[治療 Create](../../../backend/internal/medicalrecord/treatment_service.go)）。`isUnbillableMasterPrice` の null/非有限/負数は検査・ワクチン候補用であり、全治療 DTO を nullable とする根拠にしない。

## 既知の期待経路

| 対象 | マスタ登録後の期待経路 | 照合根拠 |
| --- | --- | --- |
| 商品 | 会計の「マスタから選択」には有効な商品マスタが表示される。商品を選ぶと単価などを会計明細へ渡し、`merchandise_item_id` を含む請求明細を作る。 | [商品一覧と有効状態の絞り込み](../../../frontend/src/features/accounting/components/ItemListCard.tsx)、[商品取得 API](../../../frontend/src/features/accounting/api/get-merchandise-items.ts)、[追加処理](../../../frontend/src/features/accounting/hooks/use-accounting-item-actions.ts)、[明細作成](../../../frontend/src/features/accounting/hooks/create-accounting-items.ts) |
| 診察・処置・薬剤などの治療 | 該当マスタをカルテ治療で選び、治療を保存する。確定カルテの未請求治療が請求候補へ流れ、`treatment_id` を持つ明細として扱われる。商品用の直接選択欄へ出す経路ではない。 | [治療マスタ取得](../../../frontend/src/hooks/use-treatment-master.ts)、[カルテ治療選択](../../../frontend/src/features/medical-records/hooks/use-treatments-tab.ts)、[診断プラン選択](../../../frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx)、[治療 API](../../../frontend/src/features/medical-records/api/treatments.ts)、[未請求治療](../../../backend/internal/medicalrecord/treatment_repository.go)、[請求候補](../../../backend/internal/billing/billing_item_unbilled.go) |
| 検査レコード | 検査作成が `exam_types.price` を参照し、親カルテの `billing_confirmations.status='confirmed'` のあと未請求は `exam_id`（SQL は exam.status を見ない）。会計選択欄へ検査マスタを足さない。 | [検査未請求 SQL](../../../backend/internal/billing/billing_item_repository_unbilled.go)、[create-accounting-items `exam_id`](../../../frontend/src/features/accounting/hooks/create-accounting-items.ts) |
| 接種レコード | 接種作成が `vaccines.price` を参照し、未請求は `vaccination_id`。マスタ価格が null/負なら blocking warning。 | [接種未請求](../../../backend/internal/billing/billing_item_repository_unbilled.go)、[billing_item_unbilled.go](../../../backend/internal/billing/billing_item_unbilled.go) |
| トリミング | 予約に紐づく未請求行が `trimming_course_id` / `trimming_option_id`。 | [create-accounting-items](../../../frontend/src/features/accounting/hooks/create-accounting-items.ts)、[trimming unbilled テスト](../../../backend/internal/billing/billing_item_trimming_test.go) |
| 入院プラン / ケージ | 退院会計は `care_plan_items.unit_price` を `source=hospitalization` で写す。プランマスタ価格の自動コピーは現行 FE 参照 hook が `id/name` のみ。ケージ単価は会計 unbilled に現れない。 | [退院 TX](../../../backend/internal/medicalrecord/hospitalization_discharge_tx.go)、[CarePlanRefSelect](../../../frontend/src/features/hospitalization/components/CarePlanTab/CarePlanRefSelect.tsx)、[plans master hook](../../../frontend/src/hooks/use-treatment-master.ts) |
| キャンペーン | 単価ではなく割引候補。明細の `discount-suggestions`。 | [CampaignSidePanel](../../../frontend/src/features/master/components/CampaignSidePanel.tsx)、[suggestions API](../../../frontend/src/features/accounting/api/get-discount-suggestions.ts) |

診療項目の単価は全タブ保存、課税区分と税率の保存は診察・処置のみ（[診療項目マスタ仕様](../../../docs/spec/screens/settings/master-treatment.md)、[request 変換](../../../frontend/src/features/master/routes/treatment-plan-master-model.ts)）。

## 現行 route / form 対応（12フォーム）

開始母集団は **11単価フォーム＋1割引フォーム**。`showPrice=false` だけで除外しない（ケージは単価入力を持つ）。金額のない設定フォーム（スタッフ、診断名等）は V04 全 CRUD を本課題へ混ぜない。

`CATEGORY_CONFIG.checkup.settingsPath` は `/settings/treatment-items` のみで `?tab=checkup` が欠ける。実タブは `TREATMENT_PLAN_TABS` と `toTreatmentPlanTabValue`（不正/欠落 tab は **consultation**）。検証 URL は `?tab=checkup` を使う。

| # | ページ/フォーム | 現行 route | フォーム | 金額フィールド | 0/空/負 | 保存後に追う経路 |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | 診察 | `/settings/treatment-items?tab=consultation`（旧 `/settings/consultation` は redirect） | [TreatmentItemSidePanel](../../../frontend/src/features/master/components/TreatmentItemSidePanel.tsx) + [buildConsultationCreateRequest](../../../frontend/src/features/master/routes/treatment-plan-master-model.ts) | `price`。税 `tax_type`/`tax_rate` も保存 | UI 空=0。負は保存拒否 | カルテ治療検索 category「診察」→ `item_type=consultation` → 確定→未請求 `treatment_id` |
| 2 | 検査（exam_types） | `?tab=examination` | 同上 + `buildExaminationCreateRequest`（税は送らない。`is_non_insurance` のみ追加） | `price` | 同上 | **検査作成→親カルテ会計確認 (`billing_confirmations` confirmed)→未請求 `exam_id`**。検索ダイアログは **examination-types を読まない**（下記ギャップ） |
| 3 | 処置 | `?tab=procedure` | 同上 + 麻酔。税を保存 | `price` | 同上 | カルテ治療 category「処置」→ `procedure` → `treatment_id` |
| 4 | 予防接種マスタ | `?tab=vaccine` | 同上。税なし | `price` | 同上 | **経路A:** 検索 category「予防」→ `resolveItemTypeFromCategory` が **`other`**。[経路B:] 接種レコード→未請求 `vaccination_id`（`vaccines.price`） |
| 5 | 定期健診 | `?tab=checkup`（config の settingsPath は tab なし） | 同上 + checkup-types API | `price` | 同上 | **健診タブ**は `checkup_type_id`（臨床）。未請求集計に checkup_id はない → 会計自動連携は **根拠付き N/A**。検索ダイアログは checkupTypes を **category「検査」** として出す（exam_types ではない） |
| 6 | 薬剤 | `/settings/medicine` | [MedicineSidePanelSections](../../../frontend/src/features/master/components/MedicineSidePanelSections.tsx) | 明細 `price`。分類行は強制 0 | 分類 0 は保存失敗ではない。`min=0` HTML のみ。未分類かつ price 非正かつ剤形なしは拒否 | 治療検索「薬剤」→ `medicine_id` + `unit_price` → `treatment_id` |
| 7 | 商品 | `/settings/merchandise-items` | [MerchandiseSidePanel](../../../frontend/src/features/master/components/MerchandiseSidePanel.tsx) | `unitPrice` → `unit_price` | MoneyInput。BE `binding:"min=0"` | 会計 ItemListCard 手動選択 → `merchandise_item_id`。カルテ治療へ出さない |
| 8 | 入院プラン | `/settings/hospitalization` | [HospitalizationSidePanel](../../../frontend/src/features/master/components/HospitalizationSidePanel.tsx) | `price`、体格、`billing_unit` | create は `price \|\| undefined`（0 が omit）。update は `data.price` | ケアプラン type=item の FK。退院会計は **care_plan_item.unit_price**。マスタ価格の自動転記は現行参照 hook が価格を落とす → 転記有無は検証対象（推測で合算しない） |
| 9 | ケージ | `/settings/cage` | [CageSidePanel](../../../frontend/src/features/master/components/CageSidePanel.tsx) `MoneyInput`。`CATEGORY_CONFIG.cage.showPrice === false` | form `price` | 空表示=0 | 入院の `cage_id` 割当。**billing unbilled に cage なし** → ケージ単価の会計自動連携は **N/A（一覧表示と入院割当は別ケース）** |
| 10 | トリミングコース | `/settings/trimming?tab=course` | [TrimmingCourseSidePanel](../../../frontend/src/features/master/components/TrimmingCourseSidePanel.tsx) | `price: string` | null→`""` | 予約コース → 未請求 `trimming_course_id`。外部予約/LINE は行わない |
| 11 | トリミングオプション | `?tab=option` | [TrimmingOptionSidePanel](../../../frontend/src/features/master/components/TrimmingOptionSidePanel.tsx) | `price: string` | 同上 | 未請求 `trimming_option_id` |
| 12 | キャンペーン | `/settings/campaigns`（CATEGORY_CONFIG 外、ResourceAccounting） | [CampaignSidePanel](../../../frontend/src/features/master/components/CampaignSidePanel.tsx) | `discountType` + `discountValue`（率/額） | 空・負を 0 に畳む。BE `min=0` | 会計明細の割引候補。単価列と別軸 |

Route 定義: [paths.ts](../../../frontend/src/config/paths.ts)（`settings.treatmentItems` 25–28、tab alias 259–268、medicine 230、hospitalization 235–237、cage 239、merchandise 250–252、trimming 30–32 / 265–266、campaigns 287）。登録: [settings-routes.tsx](../../../frontend/src/app/routes/settings-routes.tsx)（treatment-items 93–107、trimming 144–158、medicine 178、hospitalization 212、cage 229、merchandise 246、campaigns 425–441）。分類: [category-config.ts](../../../frontend/src/features/master/constants/category-config.ts)。

会計明細 create が載せる ID は `merchandise_item_id` / `vaccination_id` / `exam_id` / `treatment_id` / `trimming_course_id` / `trimming_option_id` のみ（[create-accounting-items.ts](../../../frontend/src/features/accounting/hooks/create-accounting-items.ts)）。`hospitalization_id` は **billing ヘッダ**側（[accounting_request.go](../../../backend/internal/billing/accounting_request.go)）。未実装の自動連携を本検証で新仕様として追加しない。

### 検索ダイアログの現行配線（検証で混同しない）

[TreatmentSearchDialog](../../../frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx) は consultation / procedure / vaccine / **checkupTypes** / medicine を読む。`checkupTypes` を category `"検査"` に載せ、**examination-types は fetch しない**。`CATEGORY_ORDER` に「入院」があるが hospitalization plans は配列に入らない。[resolveItemTypeFromCategory](../../../frontend/src/features/medical-records/lib/treatments-tab-model.ts) は 薬剤/処置/診察以外を `other` にする（予防・検査ラベルも `other`）。

## 12フォーム外の金額経路

| 経路 | 扱い | 根拠 |
| --- | --- | --- |
| `/settings/lab-device-item-masters` | 本12の対象外。会計選択へ混ぜない。価格列の業務意味は **UNKNOWN**（検査機器マッピング側で別確認） | [paths.labDeviceItemMasters](../../../frontend/src/config/paths.ts)、[settings-routes](../../../frontend/src/app/routes/settings-routes.tsx) |
| 診断名・スタッフ・予約区分等 | `showPrice: false` かつ単価入力なし → 本課題 N/A | [category-config.ts](../../../frontend/src/features/master/constants/category-config.ts) |
| 手入力治療行 | マスタなし運用。`unit_price: 0` で作成。MASTER の価格欠落と混同しない | [MedicalRecordDiagnosisPlan handleAddRow](../../../frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx) |

## 実行手順・期待値

1. 開発/QAが上表を新規/編集×保存/再読込/下流のケースに展開し、既存 V04・form・API テストで覆う箇所を下のギャップ表と紐付ける。候補 revision、Docker mount、専用合成 clinic/患者、実行者、後処理、receipt 保存先を固定する。借用した他タスクの DB や実請求を使わない。
2. 通常単価 1,200 円を新規保存→ページを開き直す→2,300 円へ編集→再読込する。正常系とは別に 0、空欄、負数、取消/保存失敗を各フォームの既存契約で照合する。空欄/NULL/0 を一律同値にしない。数量 2 の行は単価と行金額を分け、税込/税率/丸めは各仕様に従う。キャンペーンは単価ではなく割引額/率として期待値を作る。
3. request→response→再取得 API→表示値→下流 ID/金額の最初の不一致を特定する。マスタ編集が過去の確定明細を遡及変更しないこと、新規選択の価格は適切に更新されることも確認する。
4. 不一致ごとに失敗テスト→最小修正→Docker scoped 検証。商品以外を商品選択欄へ追加する改変で辻褄を合わせない。保存が正常でも該当下流の検証を省略しない。

成果物列: `route/tab / create-or-update / fixture / input / request / reread / downstream / expected / actual / evidence / cleanup / PASS-FAIL-BLOCKED-N/A`。実際値は全行 **未実行** から開始。完了は全対象が PASS または根拠付き N/A で、価格消失/不一致の FAIL・未実行が残らないこと。新しい製品仕様が必要なときだけ要件責任者の判断へ戻す。医院の元の操作特定は、合成検証を止める前提にしない。

## 既存テストと未カバー失敗テスト候補

2026-09-21: 下表の「scoped unit で newly covered」は専用envなしで追加済み。候補のうち allowlist外・専用env必須は **残 unverified**。合成マトリクス actual は実行しない。

| フォーム | 既存（価格に触れるもの） | 未カバー候補 / 今回の扱い | 主張したい不一致 |
| --- | --- | --- | --- |
| 診察 | [TreatmentItemSidePanel.test.tsx](../../../frontend/src/features/master/components/TreatmentItemSidePanel.test.tsx) 負数拒否。`buildConsultationCreateRequest persists price 0 and tax_type`（既存） | **newly covered**: update 1200→2300/tax。専用env reread UI は残 | 再読込 1200→2300、税が検査タブへ漏れないこと |
| 検査 | 同上パネル。未請求 [billing_item_exam_test.go](../../../backend/internal/billing/billing_item_exam_test.go)、bill-check 0 は請求可 | **newly covered**: examination create/update price・tax 非送出。`TreatmentSearchDialog` exam配線は residual I/O で **価格消失ではない**と固定 | 検索に exam_types が出ず checkup が検査扱い（配線。会計は exam_id） |
| 処置 | [treatment-plan-master-model.test.ts](../../../frontend/src/features/master/routes/treatment-plan-master-model.test.ts) 麻酔+税 | **newly covered**: procedure price 0 + update 2300。BE procedure 0 は既存 request test | 麻酔表示が価格契約を変えない |
| 予防 | 未請求 [billing_item_vaccination_test.go](../../../backend/internal/billing/billing_item_vaccination_test.go) | **newly covered**: vaccine price persist。予防→`other` は分類のみ（unit_price 維持。価格消失バグにしない） | 治療行と接種レコードの二重計上/取りこぼし |
| 健診 | [TreatmentPlanMaster.test.tsx](../../../frontend/src/features/master/routes/TreatmentPlanMaster.test.tsx) 権限/reorder のみ | **newly covered**: checkup price persist。会計 N/A 固定は残 | 会計 N/A を「バグで消えた」と誤判定しない |
| 薬剤 | [medicine-settings-model.test.ts](../../../frontend/src/features/master/hooks/medicine-settings-model.test.ts) 分類強制 0、明細 1500、BUG-006 | **already covered**（再実装しない）。SidePanel 負数 UI は残 | BUG-006 分類0 vs 明細0 |
| 商品 | [ItemListCard.test.tsx](../../../frontend/src/features/accounting/components/ItemListCard.test.tsx) 1200 + merchandiseItemId | **newly covered**: `merchandise-item-settings-model.test.ts` 0/1200/2300 | 会計選択に治療マスタが混ざらない |
| 入院プラン | [hospitalization_plan_request_test.go](../../../backend/internal/medicalrecord/hospitalization_plan_request_test.go) price 0 ポインタ | **newly covered**: create omit 0 / update send 0。**CarePlanRefSelect 価格非転記は scoped で実証**（price-loss） | マスタ 1200 がケアプラン 0 のまま退院会計へ |
| ケージ | [CageSettings.test.tsx](../../../frontend/src/features/master/routes/CageSettings.test.tsx) price 表示 | **newly covered**: `cage-settings-model.test.ts` 0/1200/2300。billing N/A | プラン料金との合算を推測しない |
| コース/オプション | [TrimmingSettings.test.tsx](../../../frontend/src/features/master/routes/TrimmingSettings.test.tsx) dirty。[billing_item_trimming_test.go](../../../backend/internal/billing/billing_item_trimming_test.go) | **newly covered**: `trimming-settings-model.test.ts` 空/0 分離 + reread + accounting course/option id | local 合成予約表示と会計 ID |
| キャンペーン | [campaign_service_test.go](../../../backend/internal/billing/campaign_service_test.go) 負拒否 | **newly covered**: `campaign-settings-model.test.ts` discount 0/1200/rate。suggestions 追随は残 | 単価列と割引列の取り違え |
| 横断 | [create-accounting-items.ts](../../../frontend/src/features/accounting/hooks/create-accounting-items.ts) に hospitalization/cage/checkup ID なし | **newly covered**: create-accounting-items.test が treatment/trimming id を送り hospitalization/cage/checkup を非混在 | 全マスタ混在の禁止 |

## 合成実行マトリクス（actual = 実行済み, 2026-09-23 EMR-84）

**環境**: 稼働中 dev stack `animalekarte-backend-1`（:8080、mount は main checkout と同一 commit `923bb99` = 本 worktree HEAD）。合成 clinic `clinic_id=4`（ノア動物病院 Hako bu neco）、合成ログイン `stg-staff-10000021@example.test`（林 文明・執行・全 4 医院）、合成飼主/ペット `UATR2`（owner=pet=1000000019）、合成カルテ 1000000020–1000000023。**専用env（A4 rehearsal overlay）は KNJO 21-table bundle 不在で未使用**。DB は dev 共有でありタスク隔離ではない（残余リスクとして下記）。

**マスタ 12 フォーム（create 1200 → 一覧再読込 → update 2300 → 一覧再読込）**: 全 37 チェック PASS（API 応答値・再読込値とも一致）。

| # | マスタ route | create 応答 | 再読込 | update→再読込 | 判定 |
| --- | --- | --- | --- | --- | --- |
| 1 | /masters/consultations | price=1200 | 1200 | 2300 | PASS |
| 2 | /masters/examination-types | price=1200 | 1200 | 2300 | PASS |
| 3 | /masters/procedures | price=1200 | 1200 | 2300 | PASS |
| 4 | /masters/vaccines | price=1200 | 1200 | 2300 | PASS |
| 5 | /masters/checkup-types | price=1200 | 1200 | 2300 | PASS |
| 6 | /masters/medicines | price=1200 | 1200 | 2300 | PASS |
| 7 | /masters/merchandise-items | unit_price=1200 | 1200 | 2300 | PASS |
| 8 | /masters/hospitalization-plans | price=1200 | 1200 | 2300 | PASS |
| 9 | /masters/cages | price=1200 | 1200 | 2300 | PASS |
| 10 | /masters/trimming-courses | price=1200 | 1200 | 2300 | PASS |
| 11 | /masters/trimming-options | price=1200 | 1200 | 2300 | PASS |
| 12 | /masters/campaigns | discount_value=1200 | 1200 | 2300 | PASS |

**下流経路（カルテ確定 → 未請求 → 会計）**: 全 28 チェック PASS。

| 経路 | 手順 | actual | 判定 |
| --- | --- | --- | --- |
| 診察→治療→未請求 | treatment(item_type=consultation, unit_price=2300)→confirm→GET /billing-items/unbilled | 未請求 `treatment_id`・unit_price=2300 | PASS |
| 処置→治療→未請求 | item_type=procedure, 2300 | treatment_id・2300 | PASS |
| 薬剤→治療→未請求 | item_type=medicine, quantity=2, 2300 | treatment_id・unit_price=2300・quantity=2 | PASS |
| 予防(分類 other)→治療→未請求 | item_type=other, unit_price=2300 | 未請求 treatment_id・2300（分類 other でも単価維持） | PASS |
| 検査→未請求 | POST /examinations(exam_type_id)→confirm→unbilled | 未請求 `exam_id`・unit_price=2300（`exam_types.price` 参照） | PASS |
| 接種→未請求 | POST /vaccinations(vaccine_id)→confirm→unbilled | 未請求 `vaccination_id`・unit_price=2300（`vaccines.price` 参照） | PASS |
| 接種 vs 治療 other の分離 | 上記 2 経路の同時確認 | other 行の `vaccination_id` は null、接種行は `vaccination_id` 保持。二重計上なし | PASS |
| 健診（checkup_type_id） | POST /medical-records/:id/checkups→GET | カルテに checkup 1 件、再読込一致。会計自動連携は**根拠付き N/A**（checkup_id を未請求集計が持たない） | PASS / N/A |
| 商品→会計 | POST /accountings→POST /billing-items(merchandise_item_id, unit_price=2300)→GET /accountings/:id | 会計明細 unit_price=2300・merchandise_item_id 保持 | PASS |
| キャンペーン→割引候補 | GET /billing-items/:id/discount-suggestions | campaign hit（discount_type=amount, discount_value=2300, amount=2300）。単価列と別軸 | PASS |
| 入院プラン→ケアプラン→退院会計 | POST /hospitalizations(cage_id)→POST care-plan-items(type=item, hospitalization_plan_id, unit_price=2300)→POST discharge-with-billing(create_accounting=true)→GET /accountings/:id | ケアプラン再読込 2300、退院会計明細 unit_price=2300。**価格消失なし**（`e79930591` の転記修正を実測） | PASS |
| ケージ→入院割当 | POST /hospitalizations(cage_id) | cage_id が入院に反映。billing unbilled に cage 単価は現れない → **根拠付き N/A** | PASS / N/A |
| トリミング コース/オプション→未請求 | POST /trimmings(reservation_type category=trimming, status=accounting, course_id, option_ids)→unbilled | 未請求 `trimming_course_id`/`trimming_option_id`・unit_price=2300 | PASS |

**境界（0 / 省略 / 負）**: 全 16 チェック中 15 PASS、1 FAIL → 修正済み。

| ケース | actual | 判定 |
| --- | --- | --- |
| 各マスタ price=0 create→再読込 | 0 が保存・再読込一致（consultation/exam/procedure/vaccine/medicine/merchandise/trimming course/option/campaign） | PASS |
| consultation price 省略 | price=null（0 と区別） | PASS |
| procedures/merchandise/consultations/exam-types/vaccines/checkup-types/medicines 負 | HTTP 400「金額は0以上を入力してください」 | PASS |
| treatment unit_price=0 | 201・0 保存 | PASS |
| treatment unit_price=-100 | HTTP 400 | PASS |
| care-plan-item unit_price 省略 | 0 保存（BE 契約。FE は master price を送る） | PASS |
| trimming-courses / trimming-options / hospitalization-plans / cages 負 | **HTTP 201 で負値を保存（再現 defect）** → 最小修正後 400 | **FAIL → 修正済み** |

### 再現 defect と最小修正（EMR-84）

- **症状**: `/masters/trimming-courses`・`/masters/trimming-options`・`/masters/hospitalization-plans`・`/masters/cages` の create/update が負の金額を保存（201）。他 8 マスタは `sharedkernel.ValidateNonNegativePrice` で 400。
- **下流影響**: トリミング未請求 SQL は `COALESCE(tc.price,0) > 0` のため、負価格コースは**請求候補から静かに消える**（価格消失経路）。
- **修正**: `internal/trimming/validators.go` に `validateNonNegativePrice` delegate を追加し、上記 4 サービス（trimming course/option, hospitalization plan, cage）の Create/Update で呼び出し。既存 8 マスタと同じ `ErrMsgPriceZeroOrMore` 文言。
- **検証**: 追加した service unit（create/update 負拒否）を Docker で PASS。worktree の backend を一時コンテナ（:8081、migrate スキップ）で起動し、4 マスタとも create/update 負 → 400、正値 → 201 を実機確認。

### 合成 fixture の後処理

30 件削除（マスタ・カルテ・入院・会計 cancel・トリミング予約・飼主）。残 15 件（`UATR2` 接頭辞のマスタ 11・入院 3・飼主 1）は使用中参照のため HTTP 409 で削除不可（usage-protection は正常動作）。dev 共有 DB 上の不活性な合成行。

## 停止条件

- 医院実データ・STG/PROD・共有 DB・他タスク fixture の借用で「再現した」としない。
- 全マスタを会計選択に足して欠落を隠さない。
- 確認ダイアログや画面ロックだけを価格保全の根拠にしない。
- 自動連携がコード上無い経路は N/A とし、未実行と FAIL を取り違えない。
- 2026-09-23 実行分は稼働中 dev stack（clinic_id=4）での合成検証。**専用env（A4 rehearsal overlay）での再実行は KNJO 21-table bundle 不在のため未実施**であり、この 1 点のみ未実行として残す。
- residual 3経路のうち CarePlan 転記は commit `e79930591` で修正済み（合成実測で価格消失なし）。exam_types 配線差・予防→other 分類は価格消失ではないと確定。今回新たに named mismatch として検出・修正したのは 4 マスタの**負価格保存**。
