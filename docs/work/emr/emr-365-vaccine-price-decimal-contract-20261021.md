# EMR-365 ワクチン小数単価：選定技術契約 R2

2026-10-21。Dots が技術選定した契約。本人が API の物理方式を直接承認したとの意味ではない。実装・runtime受入は未完、NOT_RUN。文書公開以外の外部更新、DB、実CSV、価格実データ、試験は扱わない。

## 1. 対象と根拠

本票は臨床用量ではなくワクチンマスタ単価の保存・輸送・計算・表示である。最新本文と全8コメント（ページ終端確認）を取得。旧コメントの正本pathは移動済みで、現在は docs/work/emr/emr-342-medicines-contract-design/DESIGN.md §9 と docs/work/emr/emr-364-365-contract-design/CURRENT-PATHS.md。旧提案の表示だけ変更案は採用しない。

固定main e71ddf03462a0ee7a9864ad2ce6854b243591770で確認した現在差分：
- vaccines.price は最終ERDでも nullable bigint。Vaccine.Price、create/update request、response、unbilled vaccination rowは *int64。
- PATCH は version 必須、price の省略/nullは既存ともnilで変更なし。新規は省略/nullで未設定、明示0は許容、負はserviceで拒否。
- frontend masterの共通TreatmentFormDataはprice:number、vaccine transformは欠落を0へ置換。MoneyInputはNumber変換が残る。ワクチンだけdecimalを通す別typed branchが必要で、全品目型変更の採用ではない。
- billing_items.unit_price/quantity/tax_rate/discount_rate は末尾DDLでは numeric。model/accounting.goとstorage_finance.goはExactNumeric。旧§9のbigint単価・numeric(10,1)数量という記述は現保存正本ではない。
- model/exact_numeric.goはpgtype.Numericの係数/指数/Validを保持する型。保存型は再利用可能だが有限値・範囲・API文字列表現・丸めを自動保証しない。
- 旧unbilled SQLは退役したvaccination/billing confirmation参照も残る。この関数のscan型だけ変更して新会計接続が完成したとは扱わない。484/500等の採択source reader・billing ownerの必要subsetへ接続する。

## 2. 選定方式と旧Aとの差

採択はA改訂：vaccines.price を numeric(23,4) nullable、Goは *ExactNumeric + vaccine専用有限十進validation。新しい臨床上限や薬剤量のscaleを選定しない。

旧A numeric(19,4) は整数部15桁で、既存bigintの全整数値を無損失で移せない。numeric(23,4) は整数部19桁でbigint既存値をそのまま保持し、小数4桁も持つ。この23/4は技術表現の選定であり、院内価格上限や本人承認済み値ではない。

新規/価格変更の許容値は有限・非負・小数有効桁4以下・整数部19桁以下。最大表現値9999999999999999999.9999。NaN/Infinity/-Infinity禁止。既存bigintの負値が実在するかは未調査。既存値を移行時に0/NULL/削除へ変換せず保持し、既存負値の現在の全経路での請求禁止は未証明であるため、移行後のreaderは不正旧値として分離し、既決非負の会計consumerへ渡さない。新しい全行非負CHECKを無検査で付けて移行不能や他項目編集不能にしない。新規/price変更の非負拒否は現在service規則の維持。

別案のunbounded numericは任意source scaleを保持できるが、既存§9のbounded契約を捨てて新たな資源制限/輸送範囲を定め直す。scaled bigintは倍率混同と乗算overflowの追加負担。numeric(19,2)は0.004を保持できない。よってA改訂を技術選定する。実装・実データ移行の完了や承認を表すものではない。

4桁を超える有効小数や範囲超過は、新APIでは400、CSVではpreflight拒否・producer needs_review。丸め/切捨てで収めない。0.40000等末尾0だけの超過はexact値として0.4000へ正規化可。未対応値を黙ってNULLにしてbundle受入しない。これは保存能力上の境界であり、医療上の適正値判定ではない。

## 3. Decimal wire・入力・表示

新serverのresponseは required price: JSON string|null。有限値を小数4桁固定で返す（例"0.4000"、"500.0000"）。NULLは明示null。保存済みlegacy負値が存在すれば"-1.0000"等のexact signed stringで返し、不正旧値として表示する。欠落や無料へ変換しない。create/明示価格変更入力のみ有限非負plain decimalとし、符号+/-、指数、桁区切り、空白、NaN等を拒否する。入力は整数文字列又は1〜4有効小数、末尾0は前節。空文字はnullにせず400。旧responseはprice,omitemptyのためNULL時はキー自体が欠落する。旧reader互換ではundefinedも未設定として保持し、0へ補完しない。

新createの省略/null=未設定、"0"=明示0。PATCH省略/null=変更なしという既存互換を維持する。未設定へ戻す新clear操作は本票では発明しない。非price PATCHではpriceを再送せず、legacy負値も既存のまま保持する。価格を編集した場合だけ非負validatorを適用する。これは現vaccine_service.go create/updateのvalidateNonNegativePrice(input.Price)がnilを許容し負を拒否する規則の維持である。価格以外の既存権限・医院scope・version CAS・保険設定null規則・actor/監査は変更しない。新C06 identityを本票だけで追加しない。

Go入力はJSON文字列を正確にparseし、整数係数と指数へ変換。SQL Scanner/Valuerを通すがfloat64経由を禁止。新型にpgtypeのNaN/Infinityが来た場合も入口で拒否。金額演算はbig.Int係数/scale又は同等exact decimalで行い、丸め境界でのみ整数円へ落とす。

表示は保存値の正確な十進表記から末尾0だけ除去する。500.0000→500、0.4000→0.4、0.0400→0.04、0.0040→0.004。1桁への強制丸め/「正・小額」markerより情報落ちがなく、小数第1位も表示できる。保存値と表示値に差を作らない。NULLは未設定、0は0円、正の微小値を無料/ハイフン表示にしない。編集draftもstringで保持しNumber/parseInt/valueAsNumberへ通さない。

## 4. API互換とactivation

ID/versionの既存数値wireを本票で一括文字列化しない。変えるのはpriceのみ。以下を一つのcutover計画とし、同PRというだけで同時deployと見なさない。
1. FE全ワクチン受け手を string|null、旧undefined（priceキー欠落）、旧safe整数numberの互換readerへ先行移行。旧numberはNumber.isSafeIntegerを満たすときだけ即十進string化。安全整数の旧負値もlegacy readとして正確に識別するが、新writeの負値許可にはしない。キー欠落は未設定のまま扱う。小数/unsafe numberの丸め救済禁止。旧number送信は移行中の既存server向けにのみ維持する。
2. 新serverはprice stringを正本として受け、互換期間に旧JSON整数tokenのみ受理する専用decoderを持つ。tokenはUseNumber/RawMessageから整数としてparseしfloat64を通さない。ただし旧JS callerで既に失われた数値は復元できないためsafe整数範囲を超える旧numberは拒否する。decimal numberは常時拒否。
3. newserverが必ず文字列を返す前に旧number-only FE/接種履歴/治療選択/会計候補/帳票・export callerの更新・古い開いたclientの退役又は再読込を確認する。未更新callerを残したまま小数master入力を有効化しない。backend codeが独立deployされるなら価格機能の有効化はこのclosure後とする。
4. API仕様/生成型/fixtureを同revisionへ固定し、新producerの小数が一つでも保存される前に下流全体がexact値を受ける。旧serverへの小数送信禁止。旧caller不在確認後legacy integer入力互換を撤去できる。

新endpointや権限を作る案ではない。全受け手確認の実績は未取得、実装票で必要subsetを具体追跡する。

## 5. 計算は既決D10に従う

単価保存時・会計候補転記時に円へ丸めない。元単価×元数量の積をexactで求め、税抜入力の行金額確定時に1円単位四捨五入（非負値half-up）する。その後、同scopeの割合を合算し割合値引額を円四捨五入、定額値引、会計全体割引の最大剰余配賦、請求書/税率別合計から税額を一度切捨てる。元単価・数量・行積・各段階・採用master版を保存し、master改定で既存請求を書き換えない。

これは新しい丸め業務案ではなく accounting/lifecycle.md第五群・第十群・367追加節の既決。旧§9の「丸め順すべて未決」は訂正対象。0.4×3=1.2を途中保持し、税抜行金額は1円。0.5×1=0.5→行1円。0.04×1→行0円でも元単価0.04を0へ書き換えない。保険無し/割引無しで1円の10%税率groupなら税0円、請求1円。これは合成入力上の例でありワクチン全てを10%と認定しない。

税込入力がある場合の入力額保持は維持。505の税込混在/割引/整数内訳の必要subsetを参照し、365独自の内税逆算式を足さない。旧unbilledの0.10固定を正本とせず採用税区分/税率を保持する。表示税込額と請求書税額の算定基準を混同しない。

数量は元のexact値を維持する既決367に従い、0.25/0.49等を丸めない。臨床単位・在庫換算・投与量上限・数量既定値をここで新設しない。billing_items.unit_price/quantity、estimate_items.unit_price/quantity、billing_calculation_lines.unit_priceは最終DDLでunbounded numericであり、numeric(23,4)から狭いscaleへの暗黙roundは不要。他領域のtrimming bigint単価まで拡張しない。billing_calculation_snapshotsのsubtotal/tax_total/total_amountとcalculation_linesのnet_amount/applied_discount_amount/allocated_bill_discountは非負bigintで、上限9223372036854775807。各確定結果と集計をexact演算後、int64変換前に0以上かつこの上限以下と検査し、超過時は確定TX開始前又は全rollbackで拒否する。DB暗黙cast/Go overflowに委ねず、単価のtruncateで回避しない。保存可能単価最大と請求可能整数円上限は異なる。途中decimal積に単価上限や整数円上限を適用し、新しい価格業務上限を作らない。

## 6. CSV・MIG-31 handoff

旧producerのsource ruleは変えない：YOB優先/YSS-only補完、商品PK一意リンク、旧税blank/1→旧単価そのまま、0/負/NULL/空/リンク無し/未知税/2/3→bare空CSV→NULL。÷1.1を復活しない。正の小数は4桁以内でexactのまま保持し、表示行金額0からsource単価0判定をしない。

新価格profileのpriceセルはplain decimal4桁、bare空だけNULL。列名/順11列を価格だけで変更しない。値の意味/型が変わるためcompiled version/hashを更新し、producer/consumer同一artifact SHA receiptを要求する。旧legacy-21/medicines-22-candidate及び364のtax-only candidateは変更しない。将来、税率と小数単価を併用するprofileは独立flagの寄せ集めにせず、明示した単一profile/version/hashへ束縛する。tax-only candidateの先行契約は365待ちにせず、そのreceiptだけで小数単価を受入可能にしない。具体pinは本案中に仮SHAを捏造しない。

新profile preflightでprice語彙/有限性/範囲/scaleをCOPY前に検査。applyも同byte SHAとtarget型を確認しDB暗黙roundを防ぐ。verifyは件数/clinicだけでなく価格NULL数とexact値の正規化digest照合を追加。old integer bundleは旧profileとして旧契約のまま読めることと、新decimalprofileとして黙って受理しないことを分離する。

製品側選定・文書mergeだけではMIG-31のproducer再開receiptではない。producer source全file SHA/CSVschema/version・consumer pin/型受領・合成fixture定義・双方同版合意を37/41/349へ返してから、別権限の実装/生成/適用へ進む。実データや旧DBにこの調査から触れない。

## 7. 最小PR単位と必要subset

A 製品exact単価: vaccines DDL/model/ERD、vaccine request/service/response、master専用decimal FE branch、api.yaml/生成型、既存CAS/認可維持、合成fixture。
B 会計接続: 新source readerから採用master価格版snapshot、ExactNumeric明細・505算定・単価を受ける全callerのstring、未設定warning/明示0保持。計算後整数円の会計DTO/FE/exportも実際のexact reader契約又は丸めなし境界拒否を確認する。既存number wireがunsafe整数を運べると仮定せず、365から全金額wire文字列化を勝手に追加しない。484等を丸ごとgateにせずそのsource identity/適格性subsetだけ。
C CSV profile: provider31価格写像＋consumerpreflight/COPY/verify、37pin/41生成/349受領。A/B未受領でpriceNULL保留を解除しない。

AはコードPRを独立完結できても小数入力/decimal CSV適用のruntime activationはBと古いcaller退役を待つ。Bの未設計算定全体を365の追加臨床判断へ置換しない。スキップされた実行試験はNOT_RUNでありfixture記述と区別する。

## 8. 固定fixture案（全てNOT_RUN）

0.4→0.4000/表示0.4、0.04→0.0400/0.04、0.5→0.5000/0.5、500→500.0000/500、API0→0.0000/0円、source0→NULL、API負→400、旧保存負→signed exact read/非price PATCH保持/非負consumerへ渡さない、旧priceキー欠落→未設定、createNULL→未設定、PATCHNULL→不変更、空文字→400。
0.4×3=1.2→税抜行1、0.5×1→行1、0.04×1→行0でも原単価保持。0.00001→拒否、0.40000→0.4000、9223372036854775807→exact移行、9999999999999999999.9999→保存可/請求演算overflow別判定、10000000000000000000→拒否。NaN/Infinity/指数/unsafe legacy number拒否。
CSV roundtrip/price NULL差、同master改定CAS競合、legacy profile誤装、新profile中に超過1行ならCOPY前停止、旧number-only caller残存でactivation停止、原価格版の過去請求不変。確定JPY結果9223372036854775807はexact保存境界、9223372036854775808はint64変換前拒否、複数行集計での超過も全rollback。既存unsafe number金額wireを暗黙通過しない。

## 9. 選定範囲と残る受領

Dotsが採択した技術選定は4点：A改訂numeric(23,4)、price文字列wire＋段階互換、表示は丸めず末尾0除去、新decimal CSV profileの同版receipt方式。既決丸め・MIG-31 source規則・非負/NULL/0を再質問しない。旧§9のPROPOSEDはこの採択範囲に限って置き換える。producer同版受領・runtime/試験・DB適用は別の未完条件。

新しい臨床上限、数量の既定、単位換算を決める必要はない。本票を理由にそれらの本人質問を追加しない。

## 固定ソースURL

以下は全て e71ddf03462a0ee7a9864ad2ce6854b243591770 配下。
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/work/emr/emr-342-medicines-contract-design/DESIGN.md#9-proposed--unapprovedワクチン単価の小数精度emr-365-連携案
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/spec/domains/accounting/lifecycle.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/architecture/erd-detail/vaccines.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/architecture/erd-detail/billing_items.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/model/exact_numeric.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/model/vaccine.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/medicalrecord/vaccine_request.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/medicalrecord/vaccine_service.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/billing/billing_item_repository_unbilled.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/frontend/src/features/master/routes/treatment-plan-master-model.ts
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/frontend/src/lib/transforms/treatment.ts

独立レビューで確認した追加固定根拠：
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/medicalrecord/vaccine_response.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/sharedkernel/validators.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/migrations/001_init.sql#L9381-L9438
