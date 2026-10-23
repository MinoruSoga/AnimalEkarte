# 医師確認集合：非活性の内部projection・digest・世代参加契約

| 項目 | 内容 |
|---|---|
| 文書ID | design.accounting.doctor-review-internal-projection |
| 種別 | T06の非活性INTERNAL技術契約 |
| 決定状態 | 2026-10-23 root技術選定。承認済みidentity/数量/record集合の限定signatureへ修正。広いmetadata差分・新current列の先行選定を撤回し、確認時signature保存と共用医師reader方向を維持 |
| 実装状態 | production provider・物理保存・全writer接続は未供給。記載するport名は設計名で実在runtime symbolではない |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23 |

<a id="review-internal-scope"></a>
## 選定範囲と採用基準

基準mainは `b81abc26b3657bd662475a1f55b912672a24a265`。root技術選定を候補facts資料に優先する。対象は一つの既存billingを入力とするtyped current projection、純粋canonicalization/digest、純粋な世代参加計画である。合成typed入力からの評価可能性を定義し、実DBでknownを返すadapter、HTTP、確認/Return/Completeのwriteや有効化、baseline自動生成を供給済みとはしない。[後続計画](../../../work/emr/doctor-review-contract-plan-20261023.md#doctor-review-plan-owner-handoffs)の責務・未完scopeを維持する。

既存[技術正本の確認対象](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/docs/design/domains/accounting/technical-spec.md#L450-L456)を優先し、失効signatureは安定行/source・薬剤/処置identity・数量とrecord集合だけに限定する。用法・経路・dose metadata等まで全projection差分で失効させる先の選定は撤回する。医院が認めた医師の代行可等の既決を再選定しない。内容レビュー、文書公開、必要subset供給、consumer接続、全対象writer閉包、実行受入、activationは別段階である。 2026-10-24の[既存billing Confirm/Return契約](doctor-review-confirm-return-contract.md#review-actions-scope)は本内部shapeを変えず、HTTP/保存差分/専用順を追加選定する後続正本である。以下の内部契約の選定を、462全体Ready・新しいComplete禁止条件・実装許可へ変換しない。

<a id="review-internal-population"></a>
## 対象集合と完全性

一つのbillingの現在record集合は**複数recordを許す**。「同record」とは各source/plan linkとcaptured source_record_idが、その行に対応するrecordへ一致することであり、billing内の全行を一つのrecord IDへ強制する意味ではない。

- 現在record集合はactive billing_source_bindings（claim_released_at IS NULL）のtreatment→medical_recordと、manual行の明示source_record_idから発見する。billings.medical_record_id単独や保存review集合を現在集合の索引にしない。
- 各対象recordの現在の薬剤/処置の全有効行と、そのbillingの全有効臨床行を独立に列挙する。一行だけ・未請求だけの取得を完全としない。treatmentの現非削除行を読む。is_selected、旧status、prepared/performed、plan lifecycleによる新しい読取除外を作らず、readinessは別判定とする。
- plan-backed treatmentを初期範囲とし、current linkはrevoked_at IS NULL、link_purpose=charge_source、owner_kind=treatment、exactly-one treatment_idとclinic/record/petを照合する。execution_sourceは別identity・別readinessで、charge_sourceを代替しない。plan/linkをtreatmentsの既存列として捏造しない。
- 非plan-backed/不明source、manual臨床行、非対応item kindが混在すれば全評価を失敗枝へ返す。非対応行を隠してknown部分集合を返さず、consultation/otherをmedicine/procedureへ変換しない。
- 完全な空集合を評価できても、請求可能・確認可能とは認定しない。上位の既存空請求拒否・前提条件はそのまま。未知baselineを1や空集合で補完せず、read/replayから初期化しない。

<a id="review-internal-ports"></a>
## 内部typed portと結果

全current portは同一transaction-consistent snapshot contextと認証済みclinic scopeを受ける。providerのscope/完全性/参照検証はportの責務であり、単一行lookupの寄せ集めや任意のcompleteフラグで証明しない。実装と供給SHAは未受領。

| 設計上のport | 入力・出力責務 | 今回の境界 |
|---|---|---|
| CurrentRecordProjectionPort | context・clinic・medical_record_idから、下記の完全record object又はclosed failure unionを返す | 対象recordの全薬剤/処置行。表示sanitize前の生typed master参照・同院所属を検査する |
| CurrentBillingProjectionPort | 同じcontext・clinic・billing_idから、現在の重複なしrecord IDs集合と全billing clinical itemsを返す | 保存reviewから独立に集合を発見。採用済みの現在billing identityを証明できなければprojection_inconsistent、adapter未配線はreader_unavailable |
| PureCanonicalProjection | 上記typed current値から下記canonical bytes/digestとclosed評価結果を算出する | IO・lock・writeなし。authority/readiness/grant/確認行為を認定しない |
| GenerationParticipationPlan | before/finalの完全projection、期待世代、完全な影響ID集合、writer効果の対応を受け、各対象一回の更新要求を算出する | 実DB update・独自TX・commit receiptを作らない。実CAS/監査/receiptは元ownerの最外TXの責務 |
| ReviewSnapshotPort | 後続でclinic/billing/review_setに束縛した不変billing preimageと全record snapshotを読む | 確認履歴の後続依存。current-only evaluatorは呼ばない。確認済対現在の純粋比較fixtureにはtyped値を渡せる |

評価結果のexact closed unionは以下。ここにHTTP status/wireを割り当てない。

| state | exact keys | reason |
|---|---|---|
| known | state,billing_id,billing_generation,billing_digest,records,billing_clinical_items | reason keyなし |
| unresolved | state,reason | missing_baseline / source_missing / scope_mismatch / projection_inconsistent / reader_unavailable |
| unsupported | state,reason | source_kind / item_kind / manual_clinical / clinical_content_shape |

参照欠落はsource_missing、院/record帰属不一致はscope_mismatch、重複・競合・保存digest不一致はprojection_inconsistent、NULL baseline/必要数量不明はmissing_baseline。IO障害をSQL NULLや成功空配列へ潰さない。失敗枝に部分成功payloadはない。recordsの各objectのexact keysはmedical_record_id,clinical_generation,clinical_digest,clinical_items。record_countを別保存する後続ではrecords長と一致させる。

<a id="review-internal-shapes"></a>
## v1の限定signatureと現在値の根拠

以下は表示/履歴metadata全体とは別の**確認失効signature**である。clinical_items / billing_clinical_itemsという既存の配列名も、この限定要素だけを意味する。全objectはunknown/duplicate keyを拒否し全key必須。ID/世代は正signed64のcanonical十進文字列 `[1-9][0-9]*`、最大9223372036854775807。JSON number、符号、先頭0、小数/指数/空白をIDに用いない。

| object | exact keys |
|---|---|
| record item | source_kind,treatment_id,item_kind,master_id,quantity |
| billing item | billing_item_id,source_kind,source_id,medical_record_id,item_kind,master_id,quantity |

source_kindはtreatment、item_kindはmedicine又はprocedure。record treatment_idは安定treatments.id、master_idは現在行のmedicine_id又はprocedure_id。対応しないtyped master IDの併存を生rowで検査し、表示sanitize後のnilをSQL NULLと推定しない。名称/category/現在priceからidentityを作らない。plan_item_id/charge_link_idは現sourceの帰属照合に用いる証拠であり、失効signatureへ含めない。正当な技術link ID交換だけでは再確認を発生させない。

record quantityはtreatments.quantity numeric(10,2)の非NULL exact値、元NULLはmissing_baseline。billing quantityは現在billing_items.quantityの非NULL exact値で、record数量と強制一致させない。billing source_idはactive bindingのtreatment_id、medical_record_idはcaptured source_record_idと対応treatment/planのrecordを照合する。

billing側item_kind/master_idは、481のactive bindingと保存済みimported_values.treatment_originalのtyped identityを優先再利用する。ただしtreatment_originalは**最後に取り込んだsource snapshot**であり、manual clinical override / keep_billing後の現在billing identityを当然には証明しない。adapterが当該行の現在の採用済みitem_kind/master_idを表すことを照合できる場合だけ再利用する。信頼できる現在identityを証明できなければunresolvedのまま（projection_inconsistent等の既存reasonへ写像）とし、現在treatmentから過去/会計identityを補修しない。実provider・行別適合は未供給で、新current列を自動必須化しない。

quantityはcanonical decimal stringで、float64 DTOを原値にしない。指数/先頭+なし、不要な整数先頭0・小数末尾0を除き、小数部が空なら小数点を除去し、負zeroは0とする。1と1.0はcanonical 1。元原値表現は元snapshotに保持し、新しい非負制約・臨床上限は追加しない。

このsignatureの全keyは非NULLで、nullは有効値ではない。元数量やbaselineがNULLならmissing_baseline、参照欠落はsource_missing等のclosed失敗枝とし、`"null"`や空文字で代用しない。省略も不正である。JSON number/bool、lone surrogate、重複keyを拒否し、文字列のtrim/Unicode正規化をしない。providerは全rowをこの限定objectへ投影し、rowにmetadata列があるだけでunknown key扱いにはしない。

clinical_content、content、admin_route、用法、dose量/単位/根拠・dose_pattern_revision_id、価格/割引/税/保険、sort、技術link ID等は失効signatureから除く。**対象identity/quantity/集合が同じなら、これらだけの変更で医師確認の世代・digestを変えず、metadata-onlyをunsupportedにして会計を止めることもしない。** 薬量安全確認・別目的gate・監査・履歴の既存責務は維持する。対象identity/quantityの評価自体が不明・失敗なら従来どおり未評価として停止する。既存reason名clinical_content_shapeは必要な限定signatureの非対応shapeを示し、除外metadataの欠落/変更を理由に使わない。

<a id="review-internal-canonical-digest"></a>
## 順序・canonicalization・digest

recordsはmedical_record_id、record clinical_itemsはtreatment_id、billing clinical_itemsはbilling_item_idの正確な整数昇順。float/辞書順で代用せず、同ID重複はprojection_inconsistent。表示順は別。UTF-8とJCSを用いる。以下のtextブロックは**論理shapeであり、引用符なしの文字列をそのままJCS入力とするものではない**。

```text
record:  {schema:"doctor-review-record-v1",clinic_id,medical_record_id,clinical_items}
billing: {schema:"doctor-review-billing-v1",clinic_id,billing_id,
          records:[{medical_record_id,clinical_generation,clinical_digest}],
          billing_clinical_items}
```

実際は下記の有効なJSON値へ写像してからJCSでserializeする。全property名・schema/source_kind/item_kind・ID/世代/quantity/digestはJSON string、records/clinical_items/billing_clinical_itemsはJSON array、各要素は上記exact objectである。空配列は完全性を証明した空集合に限り、未評価NULLの代用にしない。null・省略・未知keyはpreimageに入れず、前節の失敗へ写像する。schema literalとkeyは以下の通り固定し、表示用metadataや技術linkを足さない。

record preimageの合成JSON例（整形表示はJCS出力bytesそのものではない）：

```json
{
  "schema": "doctor-review-record-v1",
  "clinic_id": "1",
  "medical_record_id": "10",
  "clinical_items": [
    {"source_kind": "treatment", "treatment_id": "101", "item_kind": "medicine", "master_id": "201", "quantity": "1"}
  ]
}
```

billing preimageの合成JSON例。clinical_digestの`a`を64文字並べた値は型説明用のdummyであり、上例の実計算hashやgolden値ではない。実利用では対応recordの計算/保存検証済みdigestを用いる。

```json
{
  "schema": "doctor-review-billing-v1",
  "clinic_id": "1",
  "billing_id": "20",
  "records": [
    {"medical_record_id": "10", "clinical_generation": "2", "clinical_digest": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
  ],
  "billing_clinical_items": [
    {"billing_item_id": "301", "source_kind": "treatment", "source_id": "101", "medical_record_id": "10", "item_kind": "medicine", "master_id": "201", "quantity": "2"}
  ]
}
```

records要素のexact keysはmedical_record_id,clinical_generation,clinical_digest。各digestはSHA-256(JCS preimage)のraw32bytesを保存し、JSON内はlowercase64hexで表す。未マージv1の選定shapeをここで修正したもので、旧い広いshapeとの両用や既存digestの実行移行を主張しない。欠落/余剰/重複record、別院/別billingのrefを拒否し、record側とbilling側を無断逆同期しない。digestだけから確認時signatureを復元せず、現在sourceから補修しない。

<a id="review-internal-generation"></a>
## before/final世代参加計画

medical_records/billingsの既存clinical_set_generationを再利用し、第二counterや親更新triggerを作らない。元ownerの最外TXでbeforeと**最終**の限定signatureを比較する純粋な更新要求である。期待CAS不一致・overflowではcommit可能な更新計画を成立させず、実writerは最外TX全rollbackを保証する。純粋評価の成功をcommit済みと呼ばない。

影響record/billing集合は**beforeとfinal双方の所属の和集合**から求める。削除・移管・link解除前の依存先を落とさず、片側を変更後currentだけから再発見して完結させない。recordの承認済みsignature差では当該recordの別行を請求するbillingも依存先であり、変更treatmentを直接chargeするbillingだけでは足りない。影響先の実列挙・phantom保護・最終物理lock順は未供給で、pure calculatorはこれを代替しない。

| 最外成功TXの差分（supported shape内） | record世代 | billing世代 |
|---|---|---|
| recordの限定signature実差（追加/削除・identity・数量） | 該当recordだけ+1 | before/finalの当該record依存先billingを各+1 |
| billing-onlyの限定signature実差 | 不変 | 該当billingだけ+1 |
| billing record membership追加・削除・移管 | 既存record不変 | 影響billingを各+1 |
| 同値・価格/sort・metadata/技術linkだけの変更 | 不変 | 不変。ただし既存CAS・監査・価格了承等を免除しない |

同じ最外TX内の複数変更は各対象一回だけ増分する。A→B→Aの別成功TXは各回増分し旧確認は復活しない。readerが依存世代を再増分することはない。旧確認履歴は保存し、現在との世代/digest不一致による失効と、Returnの現在参照解除を別扱いにする。NULL baselineはunknownであり、この契約のread/replayが初期化しない。新規作成/reconciliationの実writer供給も今回含まない。

<a id="review-internal-storage-direction"></a>
## 論理保存方向の選定と物理未供給

先の「clinical_content不足を補う新current projection列」の自動採用は撤回する。current billing identityは前述の481保存済みtyped identityとactive bindingを、現在採用値との一致をadapterが証明できる行で優先再利用し、quantityはbilling_itemsから読む。臨床metadata全量の追加保存を限定確認機能の新しい必須条件にしない。既存snapshotが現在採用identityを証明できなければunresolvedであり、新列の自動選定や現在treatmentによる補修で救済しない。

確認時の**限定billing signature preimage全量**は既存doctor_review_setsへ不変保存する論理方向を維持する。digestだけでは当時のquantity/identity/record集合を復元できず、doctor_review_recordsのrecord側snapshotも代用しない。2026-10-24に[既存setのbilling_signature/operation_idと整合制約](doctor-review-confirm-return-contract.md#review-actions-set-storage)を必要保存差分として追加選定した。DDL/model・同TXprovider・保存readerは446/466＋481/462の未供給packetで、第二確認集合・第二臨床counterは作らない。Return履歴は確認集合と別の追記事実である。

確認成功の最外TXでclinic/billing/世代/digest・doctor_review_recordsと照合し保存する実装は未完。legacy欠落はNULL/unknown、空object/空配列defaultや現在treatmentによるbackfillは禁止。必要な限定signatureを現在/確認時とも証明できないまま確認有効化しない。表示/履歴metadataの既存保持責務は別であり、全metadata追加保存を本機能のactivation gateにしない。

<a id="review-internal-physician-dependency"></a>
## T03/T06共用の医師reader再利用方向

staff/reservation単一ownerが既存AssertMedicalRecordDoctorInClinic predicateを包むconsumer portを供給し、T03/T06で共用する方向を選定する。確認では認証済みactorを渡し、client doctor_id・record担当医・予約指名へ置換しない。新adapterはambient TX必須、現在active/nondeleted staff・同院assignment・doctor分類・既存操作/scope認可を別々に満たす。

既存predicateはrecord医師参照の検査であり、確認行為の証拠ではない。staff_typeのdefault doctorだけを承認証拠とせず、正本designation写像とlegacy/default/importの扱いを共有ownerが検証し、変更/取消writerとの保護順と供給SHAを受領する。新role/grant/資格table、免許番号heuristic、職種の自由名称からの推定は作らない。

reader欠落・設定対応未検証はunavailable/unconfiguredとして当該確認を止める。通常編集・readonly準備を一律停止しない。現在適格性と、患者・内容・版へ束縛した認証済み明示確認操作の耐久証跡は別で、後者も未供給。Returnへ新たな医師限定を推定せず、confirmation_noteをReturn理由で上書きしない。Returnの正式wire/履歴/action/receiptは[既存billingの追加契約](doctor-review-confirm-return-contract.md#review-actions-return)で限定選定したが、runtime供給は未完である。

<a id="review-internal-owner-packets"></a>
## 有限owner・必要供給・AC対応

下表は実行ID/claim取得・供給SHA受領済みとの主張ではない。必要subsetを単一ownerへ渡し、全親Done待ちにしない。

| 入口・owner | 必要供給物（現状未供給） | 合成AC・残AC |
|---|---|---|
| 462統合＋current port owner | 同snapshotの全record/会計行列挙、生typed参照、失敗union写像 | 複数record、全行、未請求以外、空と未知、混在unsupported、sanitize前不正参照 |
| 446/466＋481/462 | 481既存identityの現在採用値への適合provider、確認時の限定signature不変payloadの物理型・同TXproducer/readerの固定SHA | record側と会計側の限定signature独立差分、確認後の過去billing再読。実DB known/履歴受入は未完 |
| treatment domain / EMR-495の該当writer owner | service Create/Update/Delete/BulkUpdateSortOrder、createTreatmentInTx/updateTreatmentInTx、dose persist/clear、直接repo/copy/import到達経路の効果台帳 | 数量/identity差とmetadata/技術link-only不変の対照。対象identity/quantity不明は未評価。全direct writer探索完了ではない |
| 518/516 | charge link追加/解除/差替え、plan remove/restore/reparent、採用内容変更のbefore/final集合 | 削除・移管前の依存先保持。execution link/readinessだけを無条件clinical増分しない |
| 481＋billing item owner | import/apply_source/keep_billing/auto-sync/release/restore、直接Create/Update/Delete/DeleteAllByBillingID、Complete一括生成/置換の参加 | keepで会計値不変でもrecord依存差は反映。同record別billing伝播、membershipのみの対照 |
| 495/494 | cohortへ届くC転記/複写/訂正と目的別保護 | B終了eventだけは不変。P1通常編集をT06未供給で全面停止しない |
| 共有staff/reservation owner | 同predicate port、designation対応根拠、現在適格性/保護順/認証済み確認行為の証跡 | actor≠担当医、失権・不明設定・reader欠落、client ID/boolean偽装拒否。新医師制度なし |
| 462/API/audit/489/506等の各owner | confirm/Return/Complete wire、元target anchor、必要共通hook、総lock graph・phantom防止、receipt/監査 | 先行の内部投影選定では対象外。既存billingのconfirm/Return wire・元target anchor・専用順・旧入口封鎖は[追加契約](doctor-review-confirm-return-contract.md#review-actions-scope)で選定済み、実供給/迂回移行は未完。新規billing Complete互換、C/Dは別packet |

合成仕様（コード未作成・実行NOT_RUN）は、exact keys/null/missing/型、ID最大/範囲、decimal同値/文字列非正規化、重複・列挙順独立、schema差、失敗の部分payloadなし、before/final双方の所属、同TX一増分、別TX ABA、CAS/overflow zero-write、metadata-only/technical-link-only不変対identity/quantity/add/remove変化、対象signature評価不能の停止、Return参照解除対照を含む。golden bytes/hashは未生成。これを実DB provider・全writer閉包・実行PASSの代わりにしない。

| 対照fixture（仕様のみ、NOT_RUN） | 医師確認signature・期待結果 |
|---|---|
| admin_route/用法/dose metadata/pattern revisionだけ変更 | identity/quantity/集合同一ならrecord/billing世代・digest不変。既存薬量安全gate・監査/履歴は別に維持 |
| 正当なplan/link ID交換だけ | 現source照合は継続し、安定行/source・identity/quantity/集合同一なら世代・digest不変 |
| 薬剤/処置の追加・削除・master identity交換・数量変更 | 対応する限定signature差を検知し、対象世代を最外TX一回増分、digest変化、旧確認は失効。別TX ABAは旧確認復活なし |
| metadata読取/保存面は未供給だが限定signatureは完全に評価可 | metadata不足だけでunsupported/確認失効へ落とさない。全metadata保存を新gateにしない |
| 対象identity/quantityが不明又は取得失敗 | 失敗union、部分knownなし。metadata-onlyの同値判定とは区別 |
| manual override/keep_billing後に取込snapshotと現在採用identityの一致を証明できない | unresolved維持。現在treatmentで補修せず、新current列も自動採用しない |
| before/finalで所属が変わる・同recordの別行を他billingが請求 | 削除/移管前の依存先を落とさず、承認済みsignature/依存世代の差を各対象一回だけ反映 |

<a id="review-internal-evidence"></a>
## 固定source・関連正本

- [採用基準の計画本文](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/docs/work/emr/doctor-review-contract-plan-20261023.md)。旧候補のうち今回選定した範囲は本契約に集約する。
- [Treatment原値と型](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/internal/model/treatment.go#L28-L72)、[表示sanitizeを含む既存read](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/internal/medicalrecord/treatment_repository.go#L59-L220)。
- [binding/review保存model](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/internal/model/storage_finance.go)、[現review DDL](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/migrations/001_init.sql#L8160-L8192)。確認時の会計側限定signature全量保存は未供給。
- [481採択source契約](../../../work/emr/emr-481-source-binding-contract-20261021.md)、[518採択link契約](../../../work/emr/emr-518-storage-contract-20261020.md)、[484履行契約](../../../work/emr/emr-484-fulfillment-writer-contract-20261021.md)。既存のsource identity/readinessを変更しない。
- [既存doctor predicate](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/internal/reservation/reservation_intent_repository.go#L222-L239)、[staff default/任意職種](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/internal/staff/staff_service_core.go#L43-L106)、[未確認設定の停止規則](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/docs/design/domains/medical-records/technical-spec.md#L269-L279)。
