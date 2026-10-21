# EMR-483 金額了承 before/after・結果照会：最小技術選定案

> 公開時の位置づけ（2026-10-21）：親レビューで選定した保存・wire・hash・税sourceの範囲を公開する。以下の提案／調査時点の記載はその経緯として保持し、runtime実装済み・全会計契約の実装着手可能を意味しない。現行基本運用は本人訂正により外税と確定。通常外税の閉じた経路は既決D10に従って先行可能で、一律停止しない。内税を含む値引／混在／有理税抜から整数内訳への算定subsetだけは未閉鎖で、選定まで当該subsetの有効化を禁止する。実装・試験・DB適用・マージはNOT_RUN／未実施。

2026-10-21、R4（統合全文）。親レビュー用の提案であり、採択・公開・実装済みではない。外部書込、API実装、実金銭操作、試験を行っていない（NOT_RUN）。

## 1. 結論・確定事項と残件

現 main `edf71515dbae11735a4184ff64c465e2ea51af88`、EMR-483 本文（更新 2026-10-20 06:32:48 UTC）と全コメント2件を確認。調査時点でA/Bは未選定であり、本書で保存契約を選定する。PR736 merge通知は runtime 接続の証拠ではない。

- A：既存了承履歴の request_hash と C06 binding fingerprint の二重再送正本を解消する binding 参照／互換制約。
- B：理由不要の金額了承専用 event intent。既存 correct/cancel/refund の理由必須は維持。
- 再質問しない業務：医師確認後の金額だけの変更は警告了承で可、医師再確認不要。薬剤処置追加削除・数量変更は別操作。確認前の準備は禁止しない。確定後訂正・入金・返金・精算・在庫・外部送信は本操作に含めない。
- `billing_price_acknowledgements.calculation_version` は AFTER snapshot の revision。会計 CAS は `billings.version`、行 CAS は実在する `billing_items.contract_version`。履歴の `line_version` と不存在の `billing_items.line_version` を混同しない。

## 2. 選定を勧める保存変更 A/B

### A：既存了承表を拡張し、別の了承・receipt 表は作らない

1. `billing_price_acknowledgements.operation_id bigint NULL` と `(operation_id,clinic_id) → operation_bindings(id,clinic_id) ON DELETE RESTRICT` を追加。request_hash を nullable にする。
2. shape CHECK は「旧行：operation_id IS NULL、request_hash は32byte」「新行：operation_id IS NOT NULL、request_hash IS NULL」の排他的2枝。新writerは新枝のみ。旧hashを再計算・ダミー投入・fingerprintコピーしない。旧行に架空bindingを割り当てない。
3. 旧 UNIQUE(clinic,billing,operation_key,item) を、旧枝だけの同列 partial UNIQUE と、新枝の UNIQUE(clinic,operation_id,billing_item_id) に分離。C06 scope は actor/intent を含むため、旧 key-only 制約で異なる正当scopeを誤衝突させない。
4. 新行では binding の clinic/owner=billing/intent=billing.price_acknowledgement/actor_kind=staff/actor_ref=acknowledged_by/operation_key の一致を insert guard で確認。履歴は既存 immutable trigger、clinic RLS、会計・行の複合 FK を保持。新枝は既存の deferred operation_success_required と同等の成功receipt必須検査の対象へ追加し、同TXの operation_results なしに了承行だけcommitできないようにする。短いclaimのanchorにはこの成功必須検査を付けない。
5. before_values/after_values は §9.1のclosed AckStoredValuesで、同会計の snapshot ID/revision、対象 billing_item_id、提示した単価・割引・配賦／税・合計の固定値を保存。行ごとの値は保存 snapshot/line の正確な値から作る。after snapshot ID は業務TX内採番の結果であり、事前Contentに入れない。snapshot参照と行帰属は insert guard で同clinic/billingに照合。billing_version/line_version は検査した事前 CAS、receipt は保存後版。

### B：専用 intent を既存 event 表へ追加

`billing_change_events.intent` CHECK に `price_acknowledgement` を追加。reason は本intentで NULL。correct/cancel/refund の非空理由制約を変更しない。偽finalize・correct＋ダミー理由にしない。event の source_calculation_id は BEFORE、after_values は AFTER snapshot ID/revision を明記。event revision は会計単位の既存 event revision を lock 下で次へ進め、calculation revision と同じ番号であることを要求しない。

新表は次の claim metadata 1表だけ。これは成功事実でなく、失敗／unknown binding の元対象を解決するための必須 anchor。

`billing_price_ack_operation_targets(clinic_id bigint, operation_id bigint, billing_id bigint)`：全列 NOT NULL、正bigint、PK(clinic_id,operation_id)、binding複合FK RESTRICT、immutable trigger、既存 apply_rls_policy(has_clinic_access)、scope付きrepo。billing_id に業務FKを付けず、短い claim が会計行lockを先取りしない。claim前の同clinic現在認可・対象存在確認、Applyでの再確認、lookupでの元対象認可は必須。削除済み／解決不能の会計を要求bodyの別IDで補修しない。owner/intent は binding から固定し、任意 metadata JSON は持たない。

489 の既選定 `ClaimMetadataWriter.CreateClaimMetadata(ctx,binding,request)` を再利用。winnerだけが canonical Content の billing_id から同claim TXでanchor保存。失敗はbindingもrollback、loserはwinner anchorを読む。独自 ClaimTarget 能力や第二 operation registry は作らない。既存 Runner に callback が runtime 実装済みとは扱わない。

## 3. 最小 API / 入力 DTO

prefix `/api/v1`。下記は追加案であり現在 route の説明ではない。

- POST `/accountings/{billing_id}/price-acknowledgements/preview`：純read、binding/了承履歴/receiptを作らない。
- POST `/accountings/{billing_id}/price-acknowledgements`：警告の明示了承後の一括保存。
- GET `/accountings/{billing_id}/price-acknowledgement-operations/{operation_key}`：固定intentの C06 lookup。queryなし、private/no-store。
- GET `/accountings/{billing_id}/price-acknowledgement-operations/{operation_key}/details`：成功receiptから固定snapshot/event/了承行を読む履歴。現在会計の値を元receiptへ代入しない。成功以外は詳細を返さない。

新DTOのID/版は canonical正signed64 decimal string、operation_key はC06の小文字ハイフンUUID・nil禁止。旧APIのnumeric IDを一括変更しない。金額／単価／率は exact decimal string、科学記数法/NaN/Infinity禁止、桁落ちさせない。単価・率は現modelのExactNumericへ写像。JPY確定結果は整数。共通保存termの率は非負exactで上限100を置かない。旧個別入力adapterのmax100はその既存入口だけで維持し、合算済み保存値の拒否へ転用しない。新たな丸め式・精度制限を発明しない。丸め・割引優先順はD10採用calculatorに従う。

preview request は closed `{changes,expected_billing_version,expected_line_versions}`。
changes は closed `{bill_discount_terms,lines}`。bill_discount_terms は required nullable な §8 DiscountTerms（null=現在条件を保持、空配列=明示割引なし）。lines は unique item_id 昇順の配列、各行 closed `{billing_item_id,unit_price,discount_terms}`。discount_terms は §8の既存複数率／定額を失わない明細scope条件の完全値を明示する。空linesは会計全体割引変更時だけ可。数量・名前・source・tax/insurance区分・owner/pet・確定状態・受領額は入力不可。全体割引は §8 snapshot conditions の会計割引入力に写像し、payment更新へ流さない。

expected_line_versions は closed `{billing_item_id,contract_version}` 配列で、サーバーが計算に使う全現在行を完全列挙する。影響行だけで他行の計算前提を無保護にしない。未知／重複／欠落／別会計行は拒否。

preview response は closed `{billing_id,changes,before,after,affected_item_ids,preconditions,calculation_details,has_changes}`。has_changesはboolean。
before は closed `{calculation_id,revision,rule_version,snapshot_digest}`、after は closed `{revision,rule_version,snapshot_digest}`。after revision は保存済最大revision+1を提示するが予約はしない。preconditions は closed `{billing_version,line_versions,doctor_review_set_id,billing_clinical_generation,billing_clinical_digest,records}`、records は `{medical_record_id,clinical_generation,clinical_digest}` の完全な昇順集合。calculation_details は closed `{before,after}` で、各値は既存 snapshot/line の全保存fieldを同じschema-tagged固定projectionとして返し、上記 snapshot_digest の対象とする。clientからそのJSONをwrite値として信用しない。既存calculatorの版付きJSON schemaは未定義だったため、§8で505所有の最小closed schemaを同時提案する。既存実装済みschemaの再利用とは主張しない。

submit は closed `{operation_key,changes,acknowledgement,preconditions}`。acknowledgement は closed `{schema_version,before,after,affected_item_ids,accepted,excess_warning_acceptances}`、schema_version=1、accepted=true。before/after は上記shape。clinic/actor/acknowledged_atは認証・server時刻のみ。通常金額了承にreasonを新必須化しない。

snapshot_digest は schema-tagged canonical JSON（JCS、UTF-8）の SHA-256 lowercase64hex。対象は保存される計算 header と全line内容、規則版・source_versions・conditions・tax_rate_totals・insurance_confirmation を含む。業務金額・率・数量は exact decimal string（schema_version/ordinal/計数は定義したJSON整数のまま）に統一し、lines は item ID順。server採番snapshot ID、operation_id、記録時刻、CASだけの版は除外。BE preview/submitで同じ固定projectionを使う。Applyが再計算した内容と提示 digest/revision/影響行集合が異なる場合は dependency-changed、保存しない。

Contentのexact fieldは§9.2の11キーのみ。rule_versionはbefore/after projection digest内に拘束し、別Content fieldを増やさない。CAS/確認集合の現在照合値は Preconditions。ただし計算snapshot内の業務根拠版を無差別に取り除かない。C06の文字列長制限に巨大JSONを直入れせず、256byte内の digest field と immutable typed command を使う。Apply直前に同じcommandからContentを再生成・照合し、未検証の別bodyを捕捉して書かない。

## 4. 現在認可・lock・462 接続

- preview/details/lookupは現在 accounting:view。submit route は既存 accounting:edit を維持し、Applyで現在grant再検証。実際の行／会計割引条件変更に限り既存 discount:edit を追加照合。同じ割引条件を保持した単価のみ変更や、それによる派生割引額の再計算だけには discount:edit を追加要求しない。通常会計権限だけで割引許可を拡張しない。旧 item handler が全て既にこの検査を持つとは主張しない。
- POSTはedit middlewareを要求するため、view-only再取得はGET lookup/details。同じPOSTをviewだけで成功replayできるとは併記しない。
- key lookupは認証clinic/staff/owner/intentから行う。既存bindingは永続anchorの元会計で現在viewを検査してからURL会計一致を照合。別targetでkey-conflictの開示認可を代用しない。
- 短いclaim TX完了 → 最外業務TX → `reservation-graph:write:v1` 共通gate → C06 binding lock/既存success確認 → billing row → 既存item更新ownerのclose境界 → 必要record ID昇順 → item ID昇順 → 下位snapshot/履歴/監査。既存itemの billing→close を close→billingへ反転しない。必要FK/policy/actor lockは既存owner順と同じambient TX。純previewはgate不要。新しい全domain gateを追加しない。
- 現在の保存済み BEFORE snapshot がない旧会計は、本操作で架空snapshotを補完せず、既存の会計準備／計算snapshot作成経路へ戻す。医師確認前の準備自体に本操作の了承を要求しない。
- Apply内でclinic/scope/grant、claim_state=draft、before current_calculation参照、会計・全行CAS、対象集合、現在doctor_review_sets/recordsを再検証。null clinical generation/digest/確認参照を「確認済み」に補完しない。
- 462の判定はレビュー集合の会計・全元record ID集合・record_count・両側clinical generation/digestと現在対象を完全一致で照合。金額変更ではこれらclinical世代・doctor_review_set_idを変更しない。別recordの履歴を同日／同petだけで代用しない。
- 現 main の462はgate runtime未接続。483初回PRはこの read-only判定participantと、対象clinical writerが同共通gateを守る必要subsetを受領するまで有効化しない。462全票Doneや全財務owner完了を前提にしないが、旧bool/statusを仮gateにして先行公開しない。
- 締め済み期間への既存post-close grant/理由条件を警告了承の無理由契約で消さない。最小初回sliceは締め済み期間を既存専用訂正へ戻す（このsliceでは処理しない）。通常price-onlyへの新理由必須化ではない。後続のpost-close対応は既存条件を保持した専用接続として扱う。

## 5. 同TX書込・結果

現在値を固定して再計算し、対象item価格更新、手修正保護に必要な既存 billing_source_bindings.manual_override/差異処理、AFTER snapshot＋全lines、billings.current_calculation_id更新、影響行ごとの了承履歴、専用event、必須audit_logs、C06成功receiptを同じ最外TXでcommitする。全体割引で配賦結果が変わる行も影響行に含む。変更されない行を不要UPDATEして版を増やさない。

auditはstaff actor、会計target、操作key/binding、固定before/after参照、了承対象集合と保存前後版を保持。監査INSERT失敗も全rollback。IP/UAの新収集・別TX best-effortを成功条件に混ぜない。計算・履歴だけが残る部分成功を許さない。

C06 `ResultResponse` の直body。成功 result_refs は closed `{billing_id,before_calculation_id,after_calculation_id,change_event_id}`、committed_versions は closed `{billing,calculation,event}`、recorded_at必須。行ごとの保存後版・了承ID・actor・提示値はdetailsが保存済み履歴/snapshotから返す。receiptの内部 EffectSummary は金銭移動なし・在庫移動なしを記録し、内部mapをそのままwire公開しない。detailsは `financial_effect="none",inventory_effect="none"` を明示。

HTTP/status は現494/489採用表をそのまま使用：succeeded/replayed/not-succeeded=200、operation-in-progress=202、result-unavailable=503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、scope-not-found=404、not-authorized=403。非成功bodyはstatus/operation_keyのみ。証明書consumerのflat error/422/busy409を流用しない。構文・型・UUID・未知field等のpre-C06拒否は既存 ErrorResponseで、C06bodyと混在させない。

成功済みは現在read認可を検査して元receiptを返しApplyなし。旧CASだけの確定拒否は人の再確認後に同Content/keyで前提更新可。after revision/digest・価格・対象変更は別intent/keyだが、旧結果の確定を先に行う。unknown/busy/lookup失敗/未検出を新key救済にしない。unknownの自動再実行なし。late retryは既存会計/行CASでも保護するが、これをreload後key発見／新操作fenceの保証と称しない。

## 6. 最小実装PRの境界と残る受入

一つのprice-only操作 family に限定し、A/B DDL/model、上記metadata、489 callback接続、純preview/submit/lookup/details、必要462判定、監査・receiptを縦に通す。新しい了承registry、計算表、医師確認表、署名鍵は不要。446保存制約、466実DB統合、489共通能力、483domain API、462確認集合predicateの責任を維持。

旧 PATCH billing-items/accountings の同じ確認後価格変更が新了承を迂回しない切替条件も必要。通常準備や確定後専用訂正を一律廃止せず、対象となる確認後draft価格writeだけを新adapter参加または明示拒否にする。

静的設計から受入済みとはしない。実装時には実001/RLS/immutable/deferred成功制約、両先行順の同期・臨床変更・確認・価格変更、全体割引配賦、snapshot変化、権限失効、同key別内容、応答断replay、unknown、監査失敗rollbackを検証する。現作業は NOT_RUN。

## 7. 固定根拠

- [EMR-483](https://app.plane.so/baritechllc/browse/EMR-483/)：本文と全2コメント。業務境界・調査時点のA/B未選定を確認。
- [D10了承接続](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/design/domains/accounting/technical-spec.md#price-acknowledgement-handoff-20261020)
- [446/466保存残](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/work/db-contracts/unified-db-20261019.md#price-ack-db-handoff-20261020)
- [了承実DDL](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/migrations/001_init.sql#L8227)、[snapshot/event実DDL](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/migrations/001_init.sql#L9397)、[確認集合DDL](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/migrations/001_init.sql#L8149)
- [現会計model](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/model/accounting.go)、[現了承model](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/model/storage_finance.go)
- [C06契約](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/operation/contract.go)、[wire](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/operation/dto.go)、[現OpenAPI](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/docs/api.yaml) MedicalRecordPlanOperationResultResponse。
- [489既選定callback](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/work/emr/emr-502-manual-edit-contract-20261021.md#2-claim-と業務保存の順序)
- [462接続](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/design/domains/medical-records/technical-spec.md#doctor-confirm-gate-and-price-ack-20261019)、[共通lock順](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/design/domains/reservations/technical-spec.md#reservation-global-lock-contract)
- [既存routes](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/billing/routes.go)、[旧価格DTO](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/billing/billing_item_request.go)、[post-close現処理](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/billing/billing_item_handler.go)、[既存割引認可能力](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/httpapi/discount_permission.go)

## 8. 505 計算 snapshot JSON v1 の最小同時提案

### 8.1 現状と算式の固定根拠

現 main の [billing_service.go](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/billing/billing_service.go) はfloat演算と除去済みmodel method参照。storage_finance.go はjson.RawMessage、DDLはobject/array型チェックのみである。EMR-505の現計画も「具体JSON keyはadapter DTOとしてversion管理」として未定義。既存の版付き共通schemaが実装済みとは扱わない。

一方、[D10計算正本](https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/spec/domains/accounting/lifecycle.md#accounting-calculation-examples) と同頁の accounting-input-gross-preservation / accounting-empty-rounding-master / accounting-insurance-window-scope は既決。以下はその保存形だけを選ぶ。行税抜積の円四捨五入、同scope率合算→率額四捨五入→定額、明細→全体、最大剰余配賦、税率ごとの税切捨て、税込元額保持、超過警告了承、保険スタッフ確認を変更しない。計算例を満たすexact calculatorは未実装。残余同順位は既存 technical-spec #accounting-fifth-design の「snapshotに凍結した明細連番（ordinal）昇順」という技術案を今回採択する提案。item ID昇順はJSON整列用途だけで、配賦優先順ではない。これは既決の最大剰余法を再現可能にする技術選定であり、本人が業務上の優先行を指定したとは扱わない。483はBEFOREの凍結ordinalをAFTERへ保持し、画面sort順で変えない。

全objectは additionalProperties:false。以下に列挙したkeyは全てrequired、nullableと明記したものだけnull可。OriginalProjectionの405原masterのversion?/updated_at?だけは明記したoptional presenceを保持する例外。schema_versionはJSON整数1、計数はJSON非負整数、ID/版は正signed64 canonical decimal string、JPYは非負signed64整数文字列、Dは非負exact decimal string。Dは整数部先行0除去・小数末尾0除去（0は"0"）、符号＋・指数・負0を禁止。これらは新しいchanges/applied_pricing/派生計算値の正規化。source_snapshot.original_projectionに405等から全copyした履歴decimal文字列だけは元字句（例"0.10"）を保持する例外とし、有限十進として厳密検証するが正規化書換えしない。snapshot digestはその保存字句を含むJCSへ計算し、現値との数値比較だけexact decimalとして行う。元字句の違いを数値の違いと誤認して価格変更を作らない。

### 8.2 共通部品と条件

- `DiscountTerm = {kind,value}`。kind=`rate|fixed`、rate value=D（非負、共通保存schemaに100上限なし）、fixed value=JPY。`DiscountTerms = DiscountTerm[]`。同scopeの複数条件は保持し、kind→exact value数値順にcanonical整列、同一条件の複数適用を勝手にdeduplicateしない。T4の合算済み150%はvalue="150"の単一rate termとして元値を保存でき、100+50等の架空の元条件へ分解しない。既存の個別入力入口だけにあるmax100を共通保存型へ持ち込まない。各scopeの超過警告・承認・適用上限は既決どおり別検証する。割引なしは[]。既存取得済み条件が欠ける場合は0/[]へ補完しない。
- `DiscountResult = {terms,base_amount,summed_rate,rate_discount_amount,fixed_discount_amount,requested_amount,applied_amount,excess_amount,excess_acknowledged}`。terms=DiscountTerms、base/rate/fixed/requested/applied/excessはJPY、summed_rate=D、excess_acknowledged=boolean。既決の超過了承が必要な時だけtrueを求める。new price acknowledgement は単なる数値コピーでこの既存条件を勝手に承認済みにしない。
- `TypedSource` は識別discriminator `kind` を持つ次のclosed union。treatment=`{kind,treatment_id}`、vaccination=`{kind,vaccination_id}`、examination=`{kind,examination_id}`、trimming=`{kind,trimming_source_line_id}`、hospitalization=`{kind,care_plan_item_id,hospitalization_id,event_id,event_revision}`、manual=`{kind}`。hospitalizationは405既選定のsource形をそのまま保持。各IDは正signed64、別kindの欄混在禁止。manualは実際の手入力行だけで、旧出所不明の救済kindにしない。
- `SourceVersion = {billing_item_id,source,binding_version,applied_source_version}`。source=TypedSource、binding_version/applied_source_versionは正signed64 nullable。manualのみ両方null、それ以外は481で確定した両値を必須。現時点の全dependency版をapplied_source_versionへ上書きしない。
- `TaxInput = {status,type,rate,origin,master,confirmed_by,confirmed_at}`。405の既選定enumを継承：status known/unknown、type included/excluded/exempt nullable、rate D nullable、origin master/staff_confirmed/legacy_unknown/unconfirmed、master nullable、actor ID/時刻 nullable。unknownはtype/rate/master/actor/time全null。masterはclosed `{kind,id,version,updated_at,tax_type,tax_rate}`、kindは `medicine|consultation|procedure|hospitalization_plan|merchandise_item|trimming_course|trimming_option` の7値（下記根拠へ固定）、id正、version正 nullable、updated_at RFC3339 nullable、tax_type/typeとtax_rate/rateの当時実値。正規master版がない時はnull、時刻から版を捏造しない。known時の由来・actor/timeは405の元snapshotの値を保持し、483が今確認したactorに置換しない。
- `PricingInput = {schema_version,input_unit_price,quantity,tax}`：schema_version=1、input_unit_price=D、quantity=D >0、tax=TaxInput。税込元単価はこのinput_unit_price＋tax.type=includedに一度だけ保存し、同義gross列を足さない。405 careのinput_unit_price整数字句という入力制約は維持する。これは計算用共通projection。405のnullable原入力・元字句・field presenceをこの非NULL計算型へ強制せず、下記original_projectionへ全保存する。

税master.kindの現存列根拠は `medicines / consultations / procedures / hospitalization_plans / merchandise_items` のtax_type/tax_rateで、405の入院sourceはmedicine/procedure/hospitalization_planの3種。追加のtrimming_course/trimming_optionは482が選定するnullable税2列と既存contract_versionの実装依存であり、現mainに税列があるとは扱わない。両kindはそれぞれtrimming_courses/trimming_optionsへ一対一対応し、482必要subsetが実装されるまで当該writerを有効化しない。vaccines/exam_types/checkup_typesを無根拠に追加しない。

staff_confirmedは各source ownerに既に採択された入力経路で記録された事実のみ読み取る。405の採択はfree-text treatment/manual other careという限定枝であり、他sourceの新writer許可へ転用しない。master-backed trimmingは482のmaster枝を使う。他source未選定の税確認経路は本案で新設せず、不明をスタッフ確認済みへ変換しない。

`ExactRatio = {numerator,denominator}`：非負整数文字列numerator、正整数文字列denominator、最大公約数1へ既約化、0は0/1。中間値なのでsigned64金額上限を分子へ流用せず、浮動小数へ変換しない。税込100円/1.1のような有限decimalで表せない税抜基準を正確に保存するための技術表現で、丸め規則変更ではない。最終JPY結果は既存bigint範囲を検査する。

### 8.3 既存JSON列への exact mapping

`snapshot.source_versions = {schema_version,lines}`。lines=SourceVersion[]、billing_item_id昇順、完全集合、重複なし。これは元適用根拠版であり、実行CASを混ぜない。

`snapshot.conditions = {schema_version,bill_discount,allocation,zero_total_basis}`。
- bill_discount=DiscountResult。
- allocation=closed `{method,tie_break,eligible_base_amount,lines}`、method=`largest_remainder`、tie_break=`frozen_ordinal_ascending`、eligible_base_amount=JPY、linesは `{billing_item_id,excluded,base_amount,allocated_amount}` のID昇順完全集合。excluded boolean、額JPY。除外の初期値／設定欠落時の既決適用結果を記録するが、既存不明値を勝手にfalseへbackfillしない。
- zero_total_basis=null または closed `{reason,basis_calculation_id,basis_item_ids}`。reasonは既決の0円理由の実入力（trim非空）、basis_calculation_id正nullable、basis_item_ids正ID昇順集合。0円の実明細請求で既決上必要な根拠を保持し、空請求を通さない。既存0円根拠は483で再入力を強制せず保持する。

`snapshot.tax_rate_totals = [{tax_category,tax_rate,taxable_net_amount,input_gross_amount,tax_amount,total_amount,rounding}]`。tax_category=`taxable|exempt`、tax_rate=D、taxable_net_amount=ExactRatio（下記）、他amount=JPY、rounding=`invoice_rate_floor`。区分→率数値順、一組一行。included/excludedは入力基準でありtax_categoryの別税グループにしない。同じ課税率の内税／外税行を二つの丸め単位に分割しない。input_gross_amountは元税込入力のある行の元額集約、該当行なしの場合null許可。included元額を税抜換算・再加算した別額に置換しない。

`line.source_snapshot = {schema_version,source,original_projection,applied_pricing,manual_override,original_calculation}`。source=TypedSource、original_projection=下記closed union、applied_pricing=今回計算に実際に使うPricingInput、manual_override=boolean nullable、original_calculation=null又はclosed `{billing_id,calculation_id,billing_item_id}`。再請求だけ元固定版を参照する。元projectionを共通計算型へ縮約して保存しない。

`OriginalProjection` はclosed `{kind,values}`のdiscriminated union。kind=`hospitalization_v1|trimming_v1|pricing_v1`。source.kindとの不一致は拒否する。

- hospitalization_v1：valuesはclosed `{schema_version,source,pricing}`。schema_version=1、sourceは405のhospitalization TypedSource全量。pricingはclosed `{schema_version,input_unit_price,quantity,tax}`、schema_version=1、input_unit_priceは整数円字句又はnull、quantityは元exact decimal字句又はnull。taxはclosed `{status,type,rate,origin,master,confirmed_by,confirmed_at}`で405のenum/null規則をそのまま使用。masterはnull又はclosed `{kind,id,tax_type,tax_rate,version?,updated_at?}`、kind=medicine/procedure/hospitalization_plan、version/updated_atの有無とnullは405の保存済みpresenceを保持する。unknownの全nullable値、元rate字句、確認actor/timeを全copyし、今のactor/timeを補わない。外側schema_versionはprojection版、内側pricing.schema_versionも省略しない。
- trimming_v1：valuesは482のimported_values全量でclosed `{schema_version,source,appointment_id,detail_id,master,adopted_master,input,tax_input,accounting_context}`。schema_version=1。source=`{kind,line_id,detail_revision}`（kind=trimming）、master=`{kind,id}`（kind=course|option）、adopted_master=`{version,unit_price}`（価格nullable整数円字句）。input=`{unit_price,quantity,discount_rate,discount_amount,discount_stage}`（単価nullable整数円、数量正exact、率非負exact、定額非負整数円、stage=line。旧unitは変換せず取込停止）。tax_input=`{schema_version,status,type,rate,master}`、unknownはtype/rate/master=null、knownのmaster=`{kind,id,version}`（kind=trimming_course|trimming_option）。accounting_context=`{tax_type,tax_rate,tax_category,is_insurance_applicable,provenance}`、provenance=`{kind,master,input_detail_revision}`（kind=master）、内側master=`{kind,id,version,tax_type,tax_rate}`（施術2kind）。tax enum/rate/nullable規則は482と同じ。原入力taxと会計採用taxの2根拠を一つへ潰さない。全ID/版の対応は元source/bindingへ検査する。
- pricing_v1：valuesはclosed `{schema_version,source,pricing}`。schema_version=1、sourceはhospitalization/trimming以外のTypedSource、pricingはclosed `{schema_version,input_unit_price,quantity,tax}`、単価/数量は元exact字句又はnull、tax=TaxInput（元字句保持）。当該source ownerが既に採択した原projectionがこのshapeに完全一致する経路だけで使う。他に元保存fieldがあるownerは黙って落とさず、そのsource専用closed枝の選定まで当該consumerを有効化しない。任意JSON escape hatchではなく、全source schema完成済みとの主張もしない。

original_projectionは原値保管、applied_pricingは非NULL/knownを検査した計算入力。原unknownを0へ変えず、unknownが残る行の新規取込/算定を停止する。保存済み旧原値の閲覧を禁止する意味ではない。

`line.discount_conditions = {schema_version,line_discount,bill_discount_excluded,allocation_base_amount}`。line_discount=DiscountResult、excluded boolean、allocation_base_amount=JPY。line.requested_discount_amount は line_discount.requested_amount のexact値、applied_discount_amount はline_discount.applied_amount、allocated_bill_discount はsnapshot.conditions.allocation内同item値。これらは独立の自由更新入力ではなく一致検査する。現在 billing_items.discount_rate は summed_rate、discount_amount は applied_amount のprojectionとし、後者を次回計算の固定額termとして再解釈しない。元rate/fixed termsはsnapshotの条件を正本として読み、単価変更後に同じtermsから再計算する。旧行の合計割引額から複数termsを逆推定する互換変換はしない。

`snapshot.insurance_confirmation = {schema_version,pets}`。petsはpet_id昇順で closed `{pet_id,contract_ref,verification_state,window_state,coverage_state,verified_benefit_amount,eligible_item_ids}`。pet_id正、contract_refは既存保険契約の正ID nullable、verification_state=`not_used|unconfirmed|unknown|confirmed`、window_state=`unknown|available|unavailable`、coverage_state=`unknown|eligible|ineligible`、verified_benefit_amount=JPY nullable、eligible_item_ids正ID昇順集合。not_usedはcontract_ref/benefit null・items[]。確認不明を0円にしない。confirmedでは利用契約と今回スタッフ確認額を記録するが、保険会社自動給付計算はしない。保険確認入力UIや確認者必須欄を追加せず、483は変更せず元snapshotから保持。保険の不明を保存できることと請求確定できることを混同せず、462/505既決確定gateで判定する。

#### 原projectionから計算用入力への明示変換

- hospitalization_v1：values.pricing全量を原保存。input_unit_price/quantityはexactに検証し値を変えず共通PricingInputへ、taxは元由来/actor/timeを保持して写す。計算用masterの不足optional keyをnullで構造化しても、原projectionのpresence/字句は変更しない。原NULL・unknownは計算値0へ変換しない。
- trimming_v1：values.input.unit_price/quantityを計算単価/数量へ、accounting_contextのtax_type/rateを計算taxへ写しorigin=masterとする。計算tax.masterは同context provenance.masterを写しupdated_at=null（既存contract_versionを使用）。confirmed_by/atはnullとし、元にない個別確認者を発明しない。実import actor/timeはoperation/監査に別保存。values.tax_inputは入力税基準の原根拠として保持し、会計税と同一化しない。input.discount_rate/amountは非負rate/fixedのline termsへ写し、合算率を架空の元条件へ分割しない。
- pricing_v1：原pricingから非NULL/known検査後の計算用projectionを作る。source ownerの未選定権限・税確認経路をこの変換で新設しない。
- line.unit_price/quantity/tax_rateはapplied_pricingの対応値と一致し、元projectionと異なる会計手修正を元値へ巻き戻さない。483はBEFORE applied_pricingの単価だけを変更し、数量/税/元projectionを保持する。current itemとの矛盾は整合確認へ戻す。
- 元projection全field、applied_pricing、割引条件、手修正区分、算定結果をdigestへ拘束する。source再適用と会計単価手修正は別操作。
- tax_rate_totals.input_gross_amountは今回applied_pricingの税込入力集約であり過去source価格の集約ではない。税込元額保持と整数配賦の未閉鎖境界は§8.5に従う。

### 8.4 SnapshotProjection v1 と digest

`SnapshotProjection = {schema_version,billing_id,revision,rule_version,source_versions,conditions,tax_rate_totals,insurance_confirmation,subtotal,tax_total,total_amount,lines}`。
lines各要素は closed `{billing_item_id,ordinal,unit_price,quantity,source_snapshot,discount_conditions,requested_discount_amount,applied_discount_amount,allocated_bill_discount,net_amount,tax_rate,tax_category,insurance_category}`。ordinalは正整数、unit_price/quantity/requested_discount_amount/tax_rateはD、quantity>0、applied/allocated/netはJPY、tax_categoryはtaxable/exempt（入力type=included/excludedはどちらもtaxable、exemptはexempt）、insurance_categoryは`eligible|ineligible|unknown`。insurance対象不明を対象外に補完しない。DB line_countはlines.lengthから導出し、実保存時はordinalも一意。

projectionのrule_versionは505が既決ルール実装を識別する固定文字列 `d10-calculation-20261015-v1`。これは新算式ではなく契約実装版。旧snapshotのrule_versionをこの値へ改名しない。旧形式から同一入力／結果を証明してprojectionできなければ「旧根拠不足」とし、現在masterで再計算して同等としない。

snapshot_digest はこのprojection全体のJCS UTF-8 SHA-256。before/afterは同じ定義。採番snapshot ID/operation_id/recorded_atはprojection外、billing/revisionはprojection内。sourceの過去版は内側、CASだけの現在billing/line版は外側。linesはID昇順で、ordinalも値として拘束する。DB保存JSON、read DTO、hash projection間の暗黙field除外をしない。署名tokenや新鍵管理なし。

503等のcontent commitment方式と同じく、サーバー再計算したprojectionをimmutable prepared commandに保持し、digestとの一致をApply直前にも検証する。previewが送った任意JSONをそのまま保存する方式ではない。

超過警告についてsubmit.acknowledgementに required `excess_warning_acceptances` を追加する。closed `{scope,billing_item_id,excess_amount}` 配列でscope=`line|bill`、line時item正、bill時null、額JPY>0。現在previewで提示された超過だけの完全集合を照合してacceptedへ写像する。超過なしは[]。これは既決超過警告のtransportで、新たな警告義務ではない。配列はscope→item ID順、重複不可。

### 8.5 未閉鎖の算定subsetと有効化境界

ExactRatioは有理値を失わず保存する型の選定であり、税込/税抜混在の計算対応を閉じた証拠ではない。税込100/税10%の税抜基準1000/11を、整数JPYであるDiscountResult.base_amount・allocation.base_amount・line.net_amount・snapshot.subtotalへどの段階でどう対応させるかは未閉鎖。現在列挙した整数fieldは当該混在subsetの完全schemaと断定しない。

505が既決の元税込額保持、行/率丸め、scope別割引、最大剰余、税率集約に沿う段階別の値/型/対応と合成fixtureを選定して引き渡すまで、内税を含む値引／混在／有理税抜から整数内訳への当該算定subsetのpreview/submit/importを有効化しない。通常外税の閉じた経路は既決D10に従って先行可能で、全会計・通常外税を一律停止しない。実装者が独自round/floorや新算式で埋めることは禁止。これは単なる試験NOT_RUNとは別の設計残であり、本人へ既決業務を再質問しない。

A/B保存、了承wire/Content、元target anchor、原projection保存、通常外税の閉じた計算経路の実装設計は分離して進められる。税込混在に依存するschema/算式のgateが閉じた後に当該consumerを追加有効化する。取消/再請求全体の残件を483 gateへ追加しない。

## 9. 483 の exact Content・監査・no-op

### 9.1 acknowledgement の型

before は closed `{calculation_id,revision,rule_version,snapshot_digest}`、ID/版=正signed64文字列、rule_version=保存版の非空文字列、digest=lowercase64hex。after は closed `{revision,rule_version,snapshot_digest}`で同型、calculation_id key自体を許さない。acknowledgement は closed `{schema_version,before,after,affected_item_ids,accepted,excess_warning_acceptances}`、schema_version=1、accepted=true、affected_item_ids=正ID昇順一意配列。変更対象の元値・結果値は両projectionから照合し、client表示値を独立正本にしない。

保存用 `AckStoredValues = {schema_version,calculation_id,calculation_revision,billing_item_id,snapshot_digest,line}`。schema_version=1、ID/版は正signed64 string、digestはlowercase64hex、lineは§8.4の対象line全shape。before/afterとも同schemaで、snapshot ID/版/item帰属とline内容を保存済みsnapshot/lineへ一致検査する。単なる部分価格mapや現在master値から復元しない。

### 9.2 Content.Fields の全key（全てstring）

`schema_version="1"`、`billing_id`、`before_calculation_id`、`before_revision`、`before_snapshot_sha256`、`after_revision`、`after_snapshot_sha256`、`changes_sha256`、`affected_item_ids_sha256`、`excess_warning_acceptances_sha256`、`accepted="true"`。

これ以外のfieldは拒否。ID/版は上記canonical decimal、hashはlowercase64hex。changes hashは§3のnullを保持したclosed changesのJCS、affected IDsと超過集合hashは各canonical配列のJCS。rule_versionはbefore/after projection digestに完全拘束し、256byte制限のContentへ長い規則名を別重複fieldとして入れない。clinic/actor/owner/intentは認証Scope、operation_keyはRequest.Key。preconditionsの現在CAS値はFieldsへ入れない。Applyは同じtyped commandから全11fieldsを再生成して比較する。nullの会計割引「保持」は、固定before snapshotと組にして解決するため、同key再送で別の現在条件へ意味変更しない。

### 9.3 audit_logs の exact 値

- actor_type=`staff`、actor_id=認証staff、clinic_id=認証clinic。
- action=`price_acknowledgement`、resource=`billings`、resource_id=対象billing bigint。
- old_value/new_value は closed `{calculation_id,calculation_revision,billing_version,line_versions}`。ID/版=decimal string、line_versionsは全計算行の `{billing_item_id,contract_version}` 昇順配列。oldは検査したbefore、newは保存後値。保存変更されなかった行の版は同値。
- metadata は closed `{schema_version,operation_id,operation_key,intent,acknowledgement_ids,doctor_review_set_id}`。schema_version=1、operation_id/各ack ID/review ID正decimal string、key=UUID、intent=`billing.price_acknowledgement`、acknowledgement_idsはitem ID順の `{billing_item_id,acknowledgement_id}`配列。余分な任意request JSONは記録しない。
- created_atはserver記録時刻。ip_address/user_agentは本契約ではNULL（新収集なし）。acknowledged_at/event.recorded_at/receipt.recorded_atには同じcommand記録時刻を使用し、clientが時刻を指定しない。実入金日／原請求日をこの時刻から作らない。

### 9.4 no-op と権限差分

成功bindingの再送は現在read認可後に元receiptへ先に収束し、現在値とのno-op判定をしない。新実行はbefore固定入力と正規化済み変更後入力を比較し、単価・明細割引terms・会計割引termsが全て同値ならno-op。previewは `has_changes:false` を追加して保存ボタンを無効化し、submitはC06 validation400（確定拒否）。snapshot/event/ack/audit/成功receiptは0件。claimが既に作られていればbinding/anchorとrejected attemptは残る。数字表記だけの違いで新履歴を作らない。

入力単価・割引条件が変わっていれば、丸め結果総額が偶然同額でもno-opではない。before/after内容と履歴を保存する。discount:editは正規化済み明細/会計termsが変わった時だけ検査。単価のみ変更でbaseが変わり、同じ率から派生する適用割引額が変わっただけでは要求しない。変更なしの割引をpayloadに含めたこと自体も追加認可の理由にしない。

## 10. 505 取消／再請求の次PR契約（今回483 sliceへ実装を混ぜない）

次PRは `billing.cancel` と `billing.reissue` の別family、別key/receipt。以下は後続境界の候補であり、Content全field、12statusへの全失敗mapping、監査closed DTO、資金参加・同時性を未閉鎖。今回483の完成条件・実装前提にしない。外部実行許可でもない。

- cancel candidate POST `/accountings/{billing_id}/cancellations`：closed `{operation_key,expected_billing_version,source_calculation_id,reason}`。reason trim非空、source_calculation_idは取消対象固定版でContent、現在billing版はCAS。現 routes.go の既存 `accounting-cancel:edit` を現在検査し、新たなroleを作らない。graph gate→binding→既存cancelのbilling/close等owner順で、取消state、cancel event、481 active source release（operation/reason/time付き）、必須audit、receiptを同TX。元snapshot/原請求日/確認歴を保持し、実返金0。既存資金の解除・復元が必要な対象は506の確認済み配分participantを同TXへ受領するまでこの次PRの対象にしない。単なる元残額や無入金推定で解除を自動選択しない。
- reissue candidate POST `/accountings/{billing_id}/reissues`：closed `{operation_key,expected_original_billing_version,original_calculation_id,selected_billing_item_ids,reason}`。元取消請求／固定版／選択集合／理由をContent、現在元会計版はCAS。既存会計createと元会計view grant、同clinic/owner/元pet/source帰属、取消済み、選択元sourceのactive claim不在を現在検証。新draft ID＋新番号、元固定価格/税/割引の選択行コピー、`billing_reissue_links`、新active source claim、必要履歴/監査/receiptを同TXにする。元全体合計や未選択行をコピーせず、選択行に対応した条件から既決算式で新projectionを作る。元の保険確認を新請求の有効確認として継承せず、再確認待ちを明示。
- 両familyのlookupは会計operations配下でintent allowlistを別に閉じ、失敗bindingにも元billing anchorを保持する。成功refsはcancel=`{billing_id,change_event_id}`、reissue=`{original_billing_id,new_billing_id,original_calculation_id,new_calculation_id}`、versionsは各保存後billing/event/calculation版。新draft作成は請求確定でなく、新原請求日は後続462確定成立時。元入金の自動コピー・元fundの自動移動・実返金・外部保険処理はなし。
- 取り消した元請求に残る入金の確認充当、戻した預り金の再充当、実返金は506の別操作。元/新金額を読んだだけで新資金を作らない。再請求後の価格修正は通常の新draft編集／必要な483了承へ、確定は462へ接続する。
- cancel grant は現 routes.go の `accounting-cancel:edit` と確定する。旧 `/accountings/:id/cancel` の互換adapter／切替拒否、資金参加と旧status更新の除去が成立する前に新取消routeを開放しない。これを理由に483 A/Bのschema/price-only設計まで未決へ戻さない。共有ファイルは462統合owner経由、505は独立算定adapterとこの次sliceのhunkを提出する。


## 11. 紙上照合（静的説明、試験 NOT_RUN）

根拠はD10 lifecycle #accounting-calculation-examples / #accounting-input-gross-preservation、technical-spec #accounting-fifth-design。以下は既存例の手計算によるschema表現確認であり、コード実行・自動試験・旧実機との同等性証明ではない。

1. 複数率と別scope：税抜10,000、明細terms rate10/rate20はsum30、rate額3,000、明細後7,000。全体rate10は700、最終税抜6,300。line_discountとbill_discountを別objectで保存し、40%へ合算しない。
2. 率額丸め→定額：税抜1,005、rate10→100.5を101、fixed50、requested/applied151、net854。税10%は85、総額939。条件termsとrate_discount_amount101を別保存し、元の率10を151円の定額へ変換しない。
3. 最大剰余：1,000/2,000に全体100、整数配賦33/66、残1は剰余が大きい後者へ。allocation33/67、net967/1,933。tieは発生していない。別例1/1/1に1円は凍結ordinal1/2/3の順で1/0/0。ordinalとitem IDが逆でもordinalを使い、表示sort変更では動かさない。
4. 税率集約：税抜5/5の10%行は同グループbase10→tax1→total11。各行の税0を合計しない。別例1,008ずつの10%/8%、全体100は各50、net958/958、tax95/76、total2,087。tax_rate_totalsを率別に保持する。
5. 小数数量：101×0.5→行額51、rate10の割引額5→net46、10%税4→total50。quantity=0.5を1や0へ丸めない。1×0.5が2行なら行別1/1で計2、積をまとめて1にしない。
6. 税込入力の保持：追加割引なしの税込100円・数量1を入力して再表示すると元入力100とincludedを保持し、請求総額100を内税換算だけで99へ減らさない。換算税抜基準が1000/11のような循環小数でもExactRatioで表現可能。invoice税の既決丸め結果と総額の保存関係を照合し、単純な税抜丸め→税再加算の結果を元税込額へ上書きしない。この例は元額保持の制約確認であり、混在税込／税抜・全割引組合せの実装適合済みとはしない。
7. 元値と適用値：import時単価100/数量1/税excluded、会計で適法に単価120へ変更したBEFOREならsource_snapshot.original_projection内の元単価は100、applied_pricingとline.unit_priceは120。483で130へ変更するAFTERはsource100を保持し、applied/lineだけ130。discount_terms不変ならdiscount:edit不要、数量と税を元sourceから再適用しない。

### 483採択と後続未閉鎖の区別

483 A/B、元target anchor、価格のみのwire/Content/CAS/監査/no-opと原projection保存は採択対象。505の税込有理基準と整数配賦/集計の対応は§8.5の設計未閉鎖subsetであり、共通算定schema全体が閉じたとは公開しない。§10取消／再請求は次sliceの方向・境界であり、family単位の完全contract未閉鎖を明記する。その未閉鎖を483完成gateに加えず、公開時にも505全体設計・取消／再請求の実装開始許可を得たとは書かない。

<a id="hospitalization-405-pricing-adapter"></a>
## 12. 405既存pricing参照と今回の原projection保存のadapter追補

[405 §5の既存mapping](emr-405-tax-input-contract-20261020.md#5-481--505-のexact-mapping)でいう `line.source_snapshot.pricing` は、入院eventからコピーした原料金入力を指す。今回§8.3の版付き505保存形では、この原値全量の格納先を `line.source_snapshot.original_projection.values.pricing` とし、`original_projection.kind="hospitalization_v1"` を照合する。valuesは§8.3で列挙した `{schema_version,source,pricing}` 全体を保持する。405文書や既保存snapshotのkeyを一括改名する指示ではない。

| 読取対象 | 原値の解決先 | 計算用入力との区別 |
|:---|:---|:---|
| 405既存形式 | 保存済み `source_snapshot.pricing` | 元字句・nullable・masterのoptional presence・由来・確認者／時刻をそのまま読む。現在masterから補完しない |
| 今回の505版付き形式 | `source_snapshot.original_projection.values.pricing`（hospitalization_v1） | 原値全量を保管し、計算に実際に使う `source_snapshot.applied_pricing` と分離する |

405 consumerを新形式へ接続するadapterは、既存形式か今回の版付きclosed形式かを保存schemaで識別し、上表の同じ原料金入力を返す。未知schemaや両形式の不整合は推測fallbackせず停止する。新形式へ同義の `source_snapshot.pricing` を重複保存して第二正本を作らない。旧形式を読む能力の存在だけで新形式への自動移行・履歴backfillを許可しない。

コピーではpricing.schema_version、input_unit_price／quantityのnullと原十進字句、taxの全field、masterのversion／updated_atの有無・null、origin・confirmed_by／confirmed_atを保持する。計算用applied_pricingだけで必要な非NULL／known検査と正規化を行い、原値を書き換えない。元保存fieldが§8.3のclosed枝で無損失に表現できない場合は黙って縮約せず、その専用枝の選定までconsumerを有効化しない。税込算定の未閉鎖境界は§8.5を維持する。

### 2026-10-21本人訂正：外税の基本運用と未閉鎖範囲

本人 Sentinel_9d2e828eb9308191be6b57221ddd8f02「間違えてました、外税ですね」により、現行の基本運用は外税と確定する。通常外税の閉じたpreview／submit／import経路は既決D10に従って先行可能であり、内税edgeの未閉鎖を全会計の停止条件へ広げない。

未閉鎖は内税を含む値引／混在／有理税抜から整数内訳への算定subsetに限定する。既存の「入力した税込額を維持する」要件は撤回しない。親確認では会計フロー本文に税／値引式はなく、旧動画では2行外税100%値引→0円だけを確認したため、この資料から内税edgeの式を選定したとは扱わない。新しい数式・丸め・税率は追加せず、段階別対応の選定責任と既決の有効化条件を維持する。
