# EMR-405 入院税入力と EMR-505 固定計算版への受渡し契約

採用日: 2026-10-20。状態: **技術設計採用済み・未実装・検証 NOT_RUN**。識別: `EMR-405-TAX-DESIGN-20261020`。

Dotsが採用した保存設計と13:15 UTCの補正を本文に統合した。旧「提案／親採用前」「Q1の業務再承認待ち」は現在の状態を表さない。税込入力額維持、当時根拠保存、旧不明の非推定という既決業務を実現する契約であり、新しい業務ルールを追加しない。

調査基準mainは `da27989fa812c1319b62a1c3b4684c419b110fe1`、公開作業の基準mainは `86c9caa34d8498133eb46d4f019c0b08549db554`。両者間で本書の根拠となる入院・会計コード／schema／業務正本に変更はない。以下のrepoリンクは公開基準SHAへ固定する。Plane本文（更新13:19 UTC）と最新コメント（13:20 UTC）も再確認した。

## 1. 採用範囲と変えない境界

実際の退院会計元は **CarePlanItem** である。`treatment_plans` の概算は別であり、これを追加請求しない。[D05の退院契約](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/spec/domains/hospitalization/lifecycle.md#fr-hospitalization-discharge-and-billing)と[退院TXの現コード](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/medicalrecord/hospitalization_discharge_tx.go#L112-L140)が一致する。治療概算とケア明細の行IDを混同しない。

既存 `stay_plan_change_events` のJSONを税入力履歴の正本として再利用し、会計固定版へコピーする。`treatment_plans` に税列を2本追加するだけの方式は採用しない。現 `billing_source_bindings` は4種のsourceだけで入院careを型付き保存できないため、481と446/466へ限定binding拡張を渡す。税APIの追加だけで505入院sliceの接続完了にはならない。

| 既決事項 | 本書で守ること |
|---|---|
| D10: 当時の税区分・税率、入力額、割引scope、規則版の保持 | 請求書の税率単位税額と行表示用税込額を分ける |
| D10: 税抜入力は単価×数量後に円丸め | 数量自体を整数化しない。元数量0.25／0.49を保持する |
| D10: 税込元額保持、追加済み明細へのmaster自動反映なし | 税抜換算・丸め・再加算で元税込額を変えない。明示再適用だけ新履歴にする |
| D05: 追加・変更・中止の元行と変更経過 | 保存を実施／会計確定へ変換せず、概算と請求を分ける |
| D05: source=hospitalization、careの現数量1 | 日数／per_day／per_nightへ自動拡張せず、概算・ケージ料金を重ねない |
| 旧税情報不明の非推定 | 税込／外税／0%／10%へ自動補完しない。STG方針を削除／reset許可と解釈しない |

根拠: [D10税込入力額維持](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/spec/domains/accounting/lifecycle.md#accounting-input-gross-preservation)、[D10丸め・master再適用](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/spec/domains/accounting/lifecycle.md#accounting-empty-rounding-master)、[D05プランと記録](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/spec/domains/hospitalization/lifecycle.md#fr-hospitalization-distinct-plans-and-records)。小額値引きは手動という既決を守り、混在率・税込換算・割引複合の不足を独断の式変更で埋めない。

本採用で保存場所・出所・unknownの意味の設計gateは閉じる。499 writer、481 binding、505算定の実装やlegacy完全一致の検証を完了扱いにしない。

## 2. 現storageと必要差分

現定義はmigrationを静的に読んだ状態であり、稼働DBの確認結果ではない。

| 保存対象 | 現在の保存形 | 差分／再利用上の注意 |
|---|---|---|
| `hospitalization_plans` | `tax_type NOT NULL DEFAULT excluded`、`tax_rate numeric NOT NULL DEFAULT 0.10` | 新規選択時の明示master実値を読める。過去適用税の証明にはならない |
| `medicines` / `procedures` | 同税列。proceduresにversionあり | 読取実値をsnapshot。medicines／hospitalization_plansへ同名版を推定しない |
| `treatment_plans` | `unit_price bigint NULL DEFAULT 0`、`quantity numeric(10,1) NULL DEFAULT 1`、`discount_rate numeric(5,2) NULL DEFAULT 0`、`discount_amount bigint NULL DEFAULT 0`。税列／master FKなし | 名称からmasterを逆引きしない。quantityは0.25／0.49保持のためnumericへ追従 |
| `care_plan_items` | `unit_price bigint NULL DEFAULT 0`、medicine／procedure／hospitalization_planのnullable FK。税／版列なし | `CarePlanItem.UnitPrice int64` はDB NULLを表現できない。nullable model／readerへ追従し旧NULLを0にしない |
| `stay_plan_change_events` | care XOR treatment、正のrevision、objectのbefore／after、actor／time／operation／前event。行ごとrevision unique | 金額・税履歴の正本に再利用 |
| `billing_items` | unit_price／quantity numeric NOT NULL、tax_type NOT NULL DEFAULT excluded、tax_rate numeric NULL DEFAULT 0.10 | 取込時に値を明示。unknownをdefaultで埋めない |
| `billing_source_bindings` | 正のapplied_source_version、object imported_values、nullable manual_override、contract_version。入院sourceなし | 当時値はimported_values。現在masterから再生成しない |
| `billing_calculation_lines` | numericのunit_price／quantity／tax_rate NOT NULL、tax_category NOT NULL、object source_snapshot、net_amount bigint | 税未確認の最終行をfake tax=0で保存しない |
| `billing_calculation_snapshots` | source_versions／conditions object、tax_rate_totals array、rule_version、整数JPY totals、revision／operation_id | 結果・入力根拠・元版を同一TXで固定 |

根拠: [詳細ERD一覧](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/architecture/erd-detail/index.md)、[入院model](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/model/hospitalization.go#L104-L125)、[固定版model](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/model/storage_finance.go#L12-L50)。

## 3. 入院入力契約 v1

税専用表や自由更新の税キャッシュ列は増やさない。499所有の `stay_plan_change_events.after_values` に以下の `pricing` を含め、`before_values` は直前snapshotの全値を保持する。現在行の単価等はprojection、税正本はevent。eventのない旧行はmissing eventとしてunknownを返し、過去eventを捏造しない。

以下のJSON名は採用設計であり、まだ既存wireではない。共有API ownerが後続PRで `backend/docs/api.yaml` へ正式反映し、その公開SHAをconsumerへ渡す。

| JSON path | 型／null | 意味・由来 |
|---|---|---|
| `pricing.schema_version` | integer `1`、必須 | DTO契約版。source revision／billing CASと別 |
| `pricing.input_unit_price` | canonical整数円文字列 またはnull | 元入力単価。既存bigint契約を維持し、小数円を新規受付しない。既存NULLはnull、新規料金明細は必須。税込なら税込単価のまま |
| `pricing.quantity` | decimal文字列 またはnull | 元数量。care現取込は明示`"1"`。treatment概算は入力数量そのもの |
| `pricing.tax.status` | `known` / `unknown` | unknownをfalse／0%へ変換しない |
| `pricing.tax.type` | `included` / `excluded` / `exempt` またはnull | 既存TaxType enumに一致。known時必須、unknown時null |
| `pricing.tax.rate` | decimal文字列 またはnull | 比率。10%=`"0.10"`。exemptは明示`"0"`で課税0%とtypeで区別。unknown時null |
| `pricing.tax.origin` | `master` / `staff_confirmed` / `legacy_unknown` | free text名から推定しない |
| `pricing.tax.master` | object またはnull | master適用時のkind／id／tax_type／tax_rateの実値。正規versionがある場合だけversion、ない場合はnullと実updated_atを別保持 |
| `pricing.tax.confirmed_by` / `confirmed_at` | server actor ID／timestamp またはnull | 明示選択・確認を成立させたserver actor／time。client任意値は信用せず、unknownはnull |

`pricing.input_unit_price` と `tax.type=included` が元税込単価を一意に表す。同義の `original_gross_unit_price` を重複保存せず、税抜換算後額へ差し替えない。行積／丸め結果は505の派生結果として別名・規則版で保持する。before／afterは税だけでなく元金額・数量・割引を識別できる既存料金snapshotのshapeを含む。

### 入力・変更・旧行確認

- 新規master-backed care: actorが選択した同医院・有効masterの税区分／率をserverで読みeventに固定する。client指定IDを無検証で信用しない。金額手修正で税情報を消さず、master変更時の明示再適用は前後を新eventへ残す。
- free text treatment／manual other care: 既存手入力理由を維持し、スタッフの明示確認から税区分／率を保存する。架空master IDは作らない。unknownで概算保存できるが確定税額／確定税込総額と表示しない。food／instructionの0円だけで非課税と推定しない。
- 旧行の税確認: 通常更新可能ならchange event、完了後なら既存 `post_completion_correction` と理由必須を継承。「今確認した事実」であり作成当時から既知と記さない。最新masterの明示再適用も新eventにする。
- PATCH: 未指定は保持。tax指定時はstatus／type／rate／originを一体検証し、typeだけの部分mergeをしない。known→unknownの通常clearは不可、誤記訂正の正式操作でのみ履歴付きにする。旧unknownでmemoだけ更新する場合はunknownを保持。0は実値、NULLは不明。

## 4. 最小DB変更要求と原子保存

本節は446/466等schema ownerへの未実装要求であり、DDL適用を行った記録ではない。

### 4.1 元care・入院・医院・event版をDBで拘束する

1. `billing_source_bindings.source_kind` に `hospitalization` を追加し、nullable bigintの `care_plan_item_id` / `hospitalization_id` / `stay_plan_event_id` を追加する。hospitalization時は3列すべて非NULLかつ旧4source FK全NULL、旧4kindでは新3列全NULLのCHECKとする。複合FKのNULLによる検査skipを防ぐ。
2. eventへ通常 `UNIQUE(id,clinic_id,hospitalization_id,care_plan_item_id,revision)` を追加する（設計名 `uq_stay_plan_event_care_binding`）。既存PKと論理的に重複するが複合FKの候補キーを提供する。既存部分uniqueはFK参照先へ流用できない。
3. bindingへ `FOREIGN KEY(stay_plan_event_id,clinic_id,hospitalization_id,care_plan_item_id,applied_source_version) REFERENCES stay_plan_change_events(id,clinic_id,hospitalization_id,care_plan_item_id,revision) ON DELETE RESTRICT` を追加する（設計名 `fk_billing_source_stay_event`）。異care／異版eventはDB自身が拒否する。
4. eventの既存 `(hospitalization_id,clinic_id)→hospitalizations(id,clinic_id)` と `(care_plan_item_id,hospitalization_id)→care_plan_items(id,hospitalization_id)` が元行・医院を推移的に拘束する。bindingに同じcare／入院の独立FKを重ねない。
5. eventの既存immutable UPDATE／DELETE trigger、成功receipt必須のdeferred constraint trigger、親削除のRESTRICTは保持する。この関係一致には1 UNIQUE＋1複合FK＋CHECK拡張で足り、関係用の新constraint triggerは不要。serviceも409／422へ変換するがserviceだけの保証にはしない。

active uniqueは `(clinic_id,care_plan_item_id) WHERE care_plan_item_id IS NOT NULL AND claim_released_at IS NULL`。event IDだけのuniqueでは同じcareの異版を二重請求できるため不可。`applied_source_version=event.revision` としevent IDを版の代用にしない。release／reissueは505の既存理由・operation・元履歴を継承する。

根拠: [event現定義](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/migrations/001_init.sql#L9336-L9361)、[binding最終定義](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/architecture/erd-detail/billing_source_bindings.md)。

### 4.2 current projection・eventの同一TX／CAS／deferred検査

- 入院rowを既存共通lock順で取得し、care／treatment rowと最新event revisionを照合する。expected revision（event未作成の旧行は明示0）でCASし、revision+1のeventと現在行の単価／数量／割引投影を**同一TX**に保存する。後から独立UPDATEで修復しない。
- current readerは最新eventから税を返し、現在行の単価／数量がevent入力値と一致することを確認する。eventなし旧行は現在の元値＋tax unknownで読める。eventありで不一致なら `DATA_INTEGRITY_CONFLICT` として算定／取込を停止し、現在行やmasterを優先して黙って補正しない。
- 新規pricing schema v1 eventのINSERTとcare／treatment対象料金列変更を監視する **DEFERRABLE INITIALLY DEFERRED constraint trigger** を追加する。commit時に最新eventの入力額／数量／割引snapshotと現在列をNULL-safeに照合する。料金変更にはその操作が追加した後継eventが必要で、旧eventの再利用による書換えを拒否する。event INSERT側も投影更新欠落を拒否する。
- triggerは再算定しない。v1入力単価の整数字句・bigint範囲と、数量／率のexact numericを検査してから比較する。関係一致の複合FKと、この投影一致triggerは別の保証である。
- 導入時に旧データをUPDATEしてtriggerを起動しない。旧shape eventは読取でunknown／契約旧版と分類し、次の明示料金変更でv1後継eventを作る。memo変更を税確認・一括変換の口実にしない。beforeは変更前行／直前eventの実読取値とし、旧unknownや過去masterの不明を保持する。
- `treatment_plans.quantity numeric(10,1)` は無制限numericへ変更し、ExactNumeric／decimal文字列へ統一する。現在値を丸め直さない。DB NULLを潰すcare単価model／readerもnullableへ追従する。current投影とimmutable履歴は二つの自由更新正本にはしない。

### 4.3 未知と既存請求

required taxが揃うまで `billing_items` のdefault excluded／0.10でplaceholderを作らない。未知行は候補readerに残し、取込を `422 TAX_BASIS_UNKNOWN` で停止する。退院のみ（`CreateAccounting=false`）はこの税gateと独立して継続できる。退院＋会計の一commandが未知で失敗したら全体rollbackし、部分退院成功にしない。会計作成と現金実収受は別。

care IDのない既存入院請求は名称／金額から逆推定しない。今後の新規取込に契約を強制し、既存再請求は元保存根拠を使う経路と旧根拠不足を明示分類する。一括backfillは範囲外。

## 5. 481 → 505 のexact mapping

| 元 | binding／draft | 最終固定版 |
|---|---|---|
| care.id／入院id／event.id＋revision | source_kind=hospitalization、3参照列、applied_source_version | `line.source_snapshot.source={kind,care_plan_item_id,hospitalization_id,event_id,event_revision}` |
| event.after_values.pricing | `imported_values.pricing` 全copy。billing_itemsの単価／数量／tax_type／tax_rateを明示設定 | line単価／数量／率と `source_snapshot.pricing` 全copy |
| known tax.type | billing_items.tax_type | tax_categoryとの対応は505 adapter内で明示。included／excludedは入力基準、exemptは非課税 |
| master実値・スタッフ確認actor／time | imported_valuesに保持 | source_snapshotに保持。現在masterから再生成しない |
| event.revision | applied_source_version | snapshot.source_versionsの行参照。billing.Version／snapshot.revisionと別 |
| 手修正 | manual_overrideと前後値。未判定NULLをfalseにしない | 元入力と実適用値の両方、行／全体割引scopeをdiscount_conditionsへ |
| 算定結果 | draft見積と区別 | line.net_amount、snapshot.tax_rate_totals／subtotal／tax_total／total_amount、rule_version |

imported_valuesは取込時の不変根拠。会計修正で入院eventを改変せず会計側履歴を残す。取消再請求の初期値は元snapshotとし現在masterを自動適用しない。source同期の数量変更等は既存医師再確認gateへ接続し、405独自の臨床権限は追加しない。

## 6. 実装受渡し単位

既存endpointは共有API ownerが `api.yaml` とroute登録を照合して正式公開する。下記は未実装の担当差分であり、未確認routeを新設済みと主張しない。

| owner | 対象と接続内容 |
|---|---|
| 499 入院writer | `backend/internal/medicalrecord/care_plan_item_{request,service,response,repository}.go`、`backend/internal/model/hospitalization.go` のnullable reader、`storage_reservations.go` の既存event model。親登録からの一括書込も参加 |
| 499 治療概算 | `treatment_plan_{request,service,response}.go`、`hospitalization_service_create.go`。数量／率をfloat64からdecimal字句保持へ、単価は整数円字句。既存金額履歴＋税JSONを保存し、別請求を作らない |
| 481 入院→会計 | `hospitalization_discharge_tx.go` の固定excluded／DefaultTaxRate／float税を採用契約へ差替え、binding writerと505 calculatorへ接続。退院TX統合の単一ownerを499と固定 |
| 505／462 | `backend/internal/billing/billing_service.go` のCalculateBillingTotalsと固定計算writer／reader。`accounting_complete.go`／`accounting_complete_tx.go` の統合窓口は462。既存float式を目標正本へ昇格しない |
| FE | `frontend/src/features/hospitalization/api/{types.ts,transforms.ts,care-plan-items.ts,treatment-plans-write.ts}`、`hooks/use-hospitalization-form-model.ts`。unknownは「税情報未確認」、総額を確定税込と表示しない。入院専用adapterで欠落をexcluded／0.1にする共通計算既定値を使わず、共通算定を二重実装しない |
| 446／466・共有API owner | 最終migration／model／概要ERD・生成詳細ERDの限定差分、`backend/docs/api.yaml`正式契約。共有ファイルを505／499が並行上書きしない |

HTTP契約: enum／decimal／presence不正は400、権限403（存在秘匿は既存404方針）、税unknown／不正は422 `TAX_BASIS_UNKNOWN`／`TAX_BASIS_INVALID`、元版不一致／既active請求は409。DB／取得失敗は5xxとしてunknownへ変換しない。同operationの成功replayは保存結果を返し、最新masterで再計算しない。処理中・結果不明は489の同key照会契約を継承する。

## 7. 後続受入fixture（記述のみ・全件NOT_RUN）

1. included `"1100"`、qty `"1"`、rate `"0.10"`: 元税込1100、税100、総額1100。保存／再読込／転記で同じ根拠。
2. excluded `"1000"`、qty `"1"`、rate `"0.10"`: 税100、総額1100。
3. exempt `"1100"`、qty `"1"`、rate `"0"`: 税0、総額1100。課税0%と区分別保存。
4. 旧NULL単価、税欠落、event欠落は不明保持。現在master変更で過去根拠を補完しない。
5. 同careの版1／版2を二請求へactive取込は拒否。取消release後の選択再請求は元版関係を保持。
6. 旧行の明示確認はbefore unknown／after known、新actor／time。過去当時の確認へ偽装しない。
7. 別医院master／event、異care／入院／版の複合FK拒否、CAS競合、同key別内容、途中失敗、eventだけ／投影だけのcommit拒否と全体rollback。
8. treatment数量 `"0.25"`／`"0.49"` が保存／APIを丸めなしで往復。care数量1から日数課金へ拡張しない。
9. 税込割引・混在税率は元額保持の不変条件を確認し、実数期待値は505既決規則に従う。未確認legacy内税算式を新採用しない。

405保存契約について本人へ再質問する業務項目はない。446/466／481／499への契約反映は実装受渡しである。旧システムとの内税逆算・小数数量・混在税率・割引複合の完全一致証明はEMR-367／505の別残件であり、本設計採用で完了しない。公開不要な原資料・原画像・Driveリンク・ローカル作業パスは転載しない。

## 8. 根拠と実施状態

- [Plane EMR-405](https://app.plane.so/baritechllc/browse/EMR-405/): 採用本文、コメント `14ec114f-70b1-4a47-bd77-d057059970bf`（2026-10-20 13:20 UTC）。旧質問待ちより採用記録を優先。
- [Plane EMR-505](https://app.plane.so/baritechllc/browse/EMR-505/): 固定算定・legacy parityの後続責務。
- [D05技術設計](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/design/domains/hospitalization/technical-spec.md#L54-L70)、[care writer現コード](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/medicalrecord/care_plan_item_service.go#L323-L405)、[care request](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/backend/internal/medicalrecord/care_plan_item_request.go)。
- [stay event ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/architecture/erd-detail/stay_plan_change_events.md)、[care ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/architecture/erd-detail/care_plan_items.md)、[treatment概算ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/86c9caa34d8498133eb46d4f019c0b08549db554/docs/architecture/erd-detail/treatment_plans.md)。

この公開PRはdocsのみ。runtime／API／migration／生成物は変更せず、試験、Docker、隔離環境、鍵作成、KMS設定、DB適用を実行していない。runtime検証はSKIP、全fixtureはNOT_RUNであり、設計採用を機能完成・試験PASSと扱わない。
