# UAT-R2-MASTER-PATH: マスタ登録と会計導線の再現票

状態: **全経路の検証設計 READY／テスト未実行**。2026-09-19に依頼者が「可能性のあるページをすべて検証」と回答したため、発生ページの回答待ちを解除する。金額を入力する全マスタの新規/編集と、利用先・会計までを検証する。元の症状がどの画面で発生したかは未確定であり、全経路の検証結果と医院での再現事実を混同しない（[要件・状態の正本](../../../todo-issue.md#uat-r2-master-path)、[元報告](../stg-uat-clinic-feedback-q1-q4.md#uat-r2-master-path)）。

本票は現行 route / form / API への対応づけと、未カバー失敗テストの**候補名**までを固定する。合成検証の actual は全行 **未実行 / UNKNOWN**。全マスタを会計の「マスタから選択」へ混在させる提案はしない。

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

## 既存テストと未カバー失敗テスト候補（ファイルは追加しない）

actual 実行はしない。候補は RED 用の **ファイルパス + テスト名**。既存が価格 persist まで見ていないものはギャップとする。

| フォーム | 既存（価格に触れるもの） | 未カバー候補（提案のみ） | 主張したい不一致 |
| --- | --- | --- | --- |
| 診察 | [TreatmentItemSidePanel.test.tsx](../../../frontend/src/features/master/components/TreatmentItemSidePanel.test.tsx) `shows Japanese field error and does not call onSave when price is negative` | `treatment-plan-master-model.test.ts` `buildConsultationCreateRequest persists price 0 and tax_type` | 再読込 1200→2300、税が検査タブへ漏れないこと |
| 検査 | 同上パネル。未請求 [billing_item_exam_test.go](../../../backend/internal/billing/billing_item_exam_test.go)、[medical-record-bill-check-model.test.ts](../../../frontend/src/features/medical-records/components/medical-record-bill-check-model.test.ts) `空・負の価格を請求不能と判定する`（0 は false） | `TreatmentSearchDialog.test.tsx` `loads examination types as 検査 and does not label checkup types as 検査` | 検索に exam_types が出ず checkup が検査扱い |
| 処置 | [treatment-plan-master-model.test.ts](../../../frontend/src/features/master/routes/treatment-plan-master-model.test.ts) price 500 は麻酔ケースの副次 | `TreatmentItemSidePanel` create reread price; BE procedure 0 allowed | 麻酔表示が価格契約を変えない |
| 予防 | 未請求 [billing_item_vaccination_test.go](../../../backend/internal/billing/billing_item_vaccination_test.go) | `treatments-tab-model.test.ts` `vaccine search category 予防 maps to item_type other not a vaccination_id` | 治療行と接種レコードの二重計上/取りこぼし |
| 健診 | [TreatmentPlanMaster.test.tsx](../../../frontend/src/features/master/routes/TreatmentPlanMaster.test.tsx) 権限/reorder のみ | FE `CheckupsTab` が price を会計へ送らないことの固定 | 会計 N/A を「バグで消えた」と誤判定しない |
| 薬剤 | [medicine-settings-model.test.ts](../../../frontend/src/features/master/hooks/medicine-settings-model.test.ts) 分類強制 0、明細 1500 | `MedicineSidePanel` 負数 / 空欄 Number("")=0 / 分類0を保存失敗にしない | BUG-006 分類0 vs 明細0 |
| 商品 | [ItemListCard.test.tsx](../../../frontend/src/features/accounting/components/ItemListCard.test.tsx) 追加時 price 1200 + merchandiseItemId。[merchandise_item_handler_test.go](../../../backend/internal/inventory/merchandise_item_handler_test.go) Create/Update | FE `MerchandiseSidePanel` 0 persist / 負数 BE 400 のフォームテスト | 会計選択に治療マスタが混ざらない |
| 入院プラン | [hospitalization_plan_request_test.go](../../../backend/internal/medicalrecord/hospitalization_plan_request_test.go) price 0 ポインタ保持。[hospitalization_plan_service_test.go](../../../backend/internal/medicalrecord/hospitalization_plan_service_test.go) | `hospitalization-settings-model.test.ts` `create omits price when UI 0` vs update sends 0。CarePlanRefSelect が plan.price を unit_price にしない | マスタ 1200 がケアプラン 0 のまま退院会計へ |
| ケージ | [CageSettings.test.tsx](../../../frontend/src/features/master/routes/CageSettings.test.tsx) 一覧/CRUD に price 表示。[cage_repository_test.go](../../../backend/internal/medicalrecord/cage_repository_test.go) Price 2000 | 入院割当が cage.price を使わない / billing に cage 明細が無い | プラン料金との合算を推測しない |
| コース/オプション | [TrimmingSettings.test.tsx](../../../frontend/src/features/master/routes/TrimmingSettings.test.tsx) dirty。fixture price 3000/1000。[billing_item_trimming_test.go](../../../backend/internal/billing/billing_item_trimming_test.go) course/option ID | 空文字 price の create omit vs 0。[trimming-side-panel-model] reread `""` vs `"0"` | local 合成予約表示と会計 ID |
| キャンペーン | [campaign_service_test.go](../../../backend/internal/billing/campaign_service_test.go) DiscountValue 負拒否。[campaign_request.go](../../../backend/internal/billing/campaign_request.go) `min=0`。[master-settings-index-model.test.ts](../../../frontend/src/features/master/routes/master-settings-index-model.test.ts) カード経路 | `CampaignSidePanel` 空入力が 0 になる契約。suggestions がマスタ更新後に追随するか | 単価列と割引列の取り違え |
| 横断 | [create-accounting-items.ts](../../../frontend/src/features/accounting/hooks/create-accounting-items.ts) に hospitalization/cage/checkup ID なし | 会計 ItemListCard が consultations を fetch しない回帰 | 全マスタ混在の禁止 |

## 合成実行マトリクス（actual = 未実行）

固定: 合成 clinic、金額 1200→2300、0、空、負、保存失敗。実行者・mount・receipt は実行時に埋める。

| route/tab | create-or-update | fixture | input | request | reread | downstream | expected | actual | evidence | cleanup | 判定 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| treatment-items?tab=consultation | create | UNKNOWN | 1200 | UNKNOWN | UNKNOWN | カルテ治療 treatment_id | 再読込 1200、未請求同額 | 未実行 | — | — | UNKNOWN |
| treatment-items?tab=consultation | update | UNKNOWN | 2300 | UNKNOWN | UNKNOWN | 新規選択は新単価、旧明細は非遡及 | 同上 | 未実行 | — | — | UNKNOWN |
| treatment-items?tab=examination | create/update | UNKNOWN | 1200/0/負 | UNKNOWN | UNKNOWN | exam_id 未請求。検索ダイアログは別ケース | 0 は未請求可、NULL/負は unbillable | 未実行 | — | — | UNKNOWN |
| treatment-items?tab=procedure | create/update | UNKNOWN | 1200 | UNKNOWN | UNKNOWN | treatment_id | 税も保存 | 未実行 | — | — | UNKNOWN |
| treatment-items?tab=vaccine | create/update | UNKNOWN | 1200 | UNKNOWN | UNKNOWN | 接種 vaccination_id と治療 other を分離 | 二重計上しない | 未実行 | — | — | UNKNOWN |
| treatment-items?tab=checkup | create/update | UNKNOWN | 1200 | UNKNOWN | UNKNOWN | CheckupsTab。会計自動は N/A | 保存/再読込のみ必須。会計 N/A | 未実行 | — | — | UNKNOWN |
| /settings/medicine | create/update | UNKNOWN | 明細1200 / 分類0 | UNKNOWN | UNKNOWN | 治療 medicine_id | 分類0≠保存失敗 | 未実行 | — | — | UNKNOWN |
| /settings/merchandise-items | create/update | UNKNOWN | unit_price 1200 | UNKNOWN | UNKNOWN | ItemListCard merchandise_item_id | 治療マスタ非表示 | 未実行 | — | — | UNKNOWN |
| /settings/hospitalization | create/update | UNKNOWN | 1200 と 0 omit | UNKNOWN | UNKNOWN | ケアプラン/退院 | マスタ価格転記の有無を実測 | 未実行 | — | — | UNKNOWN |
| /settings/cage | create/update | UNKNOWN | 1200 | UNKNOWN | UNKNOWN | 入院 cage_id。会計自動 N/A | 一覧再読込。合算しない | 未実行 | — | — | UNKNOWN |
| trimming?tab=course | create/update | UNKNOWN | "1200" / "" | UNKNOWN | UNKNOWN | trimming_course_id | 空と 0 を分離 | 未実行 | — | — | UNKNOWN |
| trimming?tab=option | create/update | UNKNOWN | "1200" | UNKNOWN | UNKNOWN | trimming_option_id | 同上 | 未実行 | — | — | UNKNOWN |
| /settings/campaigns | create/update | UNKNOWN | amount 1200 / rate 10 | UNKNOWN | UNKNOWN | discount-suggestions | 単価ではない | 未実行 | — | — | UNKNOWN |

## 停止条件

- 医院実データ・STG/PROD・共有 DB・他タスク fixture の借用で「再現した」としない。
- 全マスタを会計選択に足して欠落を隠さない。
- 確認ダイアログや画面ロックだけを価格保全の根拠にしない。
- 自動連携がコード上無い経路は N/A とし、未実行と FAIL を取り違えない。
- 本単位ではテストファイルを追加しない。失敗テスト実装は別タスク。
