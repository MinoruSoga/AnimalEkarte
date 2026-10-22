# 品目マスタの全体割引除外・trimming旧属性解除：限定選定契約（A01 / A06）

2026-10-22技術選定済み。根拠の固定mainは `9eef73feac19c25571724372b0b19da160ba5cfb`。これは既決業務を実現する目標契約であり、採用済みコードへ実装されたという記録ではない。API正本への収録、必要DDL/model/ERD、製品実装・接続・受領・試験は未完。今回はdocsのみで、実DB・旧移行元の適用前棚卸し、codegen/build/test/Dockerは未実施（NOT_RUN）。migration番号は予約しない。

[coverage監査](emr-stage123-coverage-audit-20261022.md)でA01/A06を設計残とした事実を履歴として保持し、本稿がその技術選定を追加する。A02/A03の医院/LINE時間も後発の[限定選定契約](emr-clinic-line-time-selected-contract-20261022.md)で技術選定済みとなったが、実値確認・API/保存型/製品実装・受領・試験は別残件である。要件判断窓口はMinoru Soga、実装は各既存ownerの後続単一claimとし、現5セッションへ追加着手・再割当しない。本人へ既決業務を再質問せず、未照合の実医院値/移行元だけを適用時確認へ残す。

本稿AとBは同じ文書に置くが、保存型・wire・現在認可・監査・owner・受領条件を独立させる。Aの7resourceへのCAS追加をBへ広げず、Bのrequest-only clear flagをAのnullable保存方式に流用しない。計画指標は各操作の省略/明示値/拒否/再GET・原子性・旧writer境界が一意に追えること。新業務KPIや効果達成を宣言せず、別resource/第2writer/二重台帳を増やさない。

<a id="a-a01-全体割引除外の設定writer"></a>
## A. A01 全体割引除外の設定writer

### A.1 対象と単一identity

診療5種（診察/検査/処置/予防接種/健診）＋薬剤＋物販だけを対象に、既存POST/PATCH/GETを必要最小限拡張する。施術、測定field、保険商品、他の19resourceへ拡張しない。名称はDB/HTTP=`is_bill_discount_excluded`、Go=`IsBillDiscountExcluded *bool`、FE=`isBillDiscountExcluded`に統一する。

実prefixは `/api/v1`。下表の既存collectionにPOST/GET、`{id}`個別pathにPATCH/GETを用い、既存の他GET projectionでも当該品目設定を返す場合はA.3に従う。新resource/endpoint/権限/第2writerを設けない。

| 既存resource（保存表） | collection path（prefix除外） | grant（既存各action維持） | 版 | 設定writer受領 |
|---|---|---|---|---|
| consultations | /masters/consultations | master-medical | 既存version CAS | 447 |
| exam_types | /masters/examination-types | master-medical | 同方式version追加 | 447 |
| procedures | /masters/procedures | master-medical | 既存version CAS | 447 |
| vaccines | /masters/vaccines | master-medical | 既存version CAS | 447 |
| checkup_types | /masters/checkup-types | checkups | 既存version CAS | 447 |
| medicines | /masters/medicines | master-medical | 同方式version追加 | 447 |
| merchandise_items | /masters/merchandise-items | master-merchandise | 同方式version追加 | 449 |

新boolは品目別保険対象外・税・明細割引とは別。accounting grantからmaster編集を許可せず、POST/PATCH/GETのcreate/edit/view等、各既存routeのactionを維持する。trusted clinic/actorと元品目の同医院・現在権限をrequestおよび業務TXで検査し、client clinic/actorを採用しない。

### A.2 保存型とwrite presence

7表へnullable boolean列 `is_bill_discount_excluded` を追加し、DB defaultを置かない。既存NULLを一括false backfillせず、旧履歴や旧移行sourceの由来を推定しない。Go modelは*bool、HTTP responseはomitemptyなし。独立の設定表・新しい恒久履歴tableは作らない。API正本/DDL/model/ERDの実反映は未実施。

| 操作・入力 | 保存/拒否 |
|---|---|
| 新POSTのfield省略 | serviceがfalseを明示保存。DB default/Go zero値任せにしない |
| POSTのtrue/false | 指定booleanを保存 |
| POSTのnull/非boolean | 400、業務write0 |
| PATCHのfield省略 | 対象列をUPDATE mapへ入れず、既存NULL/false/trueをすべて保持 |
| PATCHのtrue/false | 対象列をmapへ明示SET。falseをzero値skipしない |
| PATCHのnull/非boolean | 400、同梱項目も含め全体無書込 |

PATCHでNULLへ戻す操作は設けない。encoding/jsonの*boolだけではnull/省略が合流するため、既存boolCreateField/boolUpdateFieldと同じpresence-aware decodeを再利用・限定調整する。保険用helperのcreate省略=trueは転用せず、本設定のservice既定falseを明示する。他属性の既存null/省略意味を一括変更しない。

### A.3 read証拠・既知未設定・unknown

GETは `is_bill_discount_excluded` keyを必ず返し、保存値true/false/nullをrawのまま表す。正常な元列読取で確認したDB NULLだけが「既知未設定」で、業務上の適用値はfalse。ただしrawNULLの由来を消さず、GETや無関係保存でfalseへ書換えない。FEもrawNULLと適用値falseを別に扱い、単に表示の既定falseを保存bodyへ混ぜない。

旧responseのfield欠落、列未接続、読取失敗、元品目不明はunknown。false/nullの成功応答を捏造せず、既存の失敗/非開示契約で停止する。`SELECT *`＋新struct=nilだけでは列未接続とDB NULLを判別できないので、当該列を明示取得するprojection等で列が正常取得された成功証拠を持つreaderを必須にする。存在しない列のエラーをnullable成功に変換しない。

④への変換は481§3.2の既選定shapeを再利用する。保存true/falseはknown/explicit、正常読取NULLはknown/missing/excluded=false、列/読取/元対象不明はunknown/unresolved/excluded=null。元master scope/版/digest・採用時点を保持し、旧履歴/snapshotへ現在設定を補完せず、旧会計・治療・見積・帳票を新field追加で更新しない。保険の既定値や非課税、明細割引禁止へ変換しない。

### A.4 版と互換

既存のconsultations/procedures/vaccines/checkup_typesのversion CASを再利用する。exam_types/medicines/merchandise_itemsの3表だけ、既存 `app_private.entity_version_guard()` と同じ方式の正bigint version（初期1）を追加する。初期1は今後のCAS起点で、過去版復元ではない。GET/保存結果に同名versionを既存4種と同じ整数/int64の限定adapterで返し、第二counter/updated_atを代用しない。

新設定を含むPATCHは編集開始version必須。新設定省略の旧PATCHは従来互換を維持するが、任意にversionを送った要求は必ず検証し無視しない。既存4種のversion必須条件を緩めない。新3種でも期待版はWHERE条件で照合し、版不一致は既存の競合response、validation不成立はzero-write。成功後の表示/編集版は実保存版を返す。

guardの同値no-op/実変更時だけ増分する規則を維持し、アプリがversionをSETして二重増分しない。create時のis_active=false補償UPDATE等で版が進む既存経路は、同TXで再読取し実保存版を返す。新boolだけ別TXにしたり、guardのために既存保存意味を変えたりしない。全26resourceの一律CAS/C06改修は対象外。

### A.5 監査・原子性・非所有writer

新boolのbefore/afterを既存audit sinkの同TXへ限定追加する。既存is_non_insurance helperは保険値しか監査しないため、保険とは別のmetadata/field識別で `is_bill_discount_excluded` のrawNULL/false/trueを分ける。beforeが存在しないcreateと、既存rowのrawNULLも混同しない。保険同時変更を含む全保存・必要監査は同じ最外TX、監査失敗は全rollback。trusted actor/clinic/resource IDを使い、新しい専用監査表やactor入力を作らない。

7CRUDの既存map更新は、新fieldを受けた時だけその列をSETする。Save/Updates(struct)/upsert/import等の非所有更新がzero値falseで既存trueを上書きしないことを、入口ごとの受領表へ明示する。禁止対象は全API方式ではなく「新列を所有しない更新が新列をSETすること」。旧sourceの新INSERTを自動false化せず、実移行元の明示設定をmissingへ落とさない。削除再importによるresetを保持保証に数えない。

通常CRUD成功responseとC06保存成功receiptは別。454 CAS handoffに残るC06設計を本設定専用のroute/owner_kindで勝手に解消せず、既存操作参加の必要subsetは各ownerへ渡す。新bool監査追加だけで全C06対応済みとはしない。

### A.6 配備順・owner・受入

契約公開/API設計は先行可。稼働は必要schema適用とmodel/readerの整合確認→runtime有効化→FE接続の順。旧serverが新fieldを無視する状態で新FEを公開しない。適用前の実DB/旧移行元棚卸し・必要既存writer確認は実装受領条件であり、技術方式の未選定ではない。migration番号は実装時の既存規約で決め、今回は適用しない。

445が単一統合、446/466が必要DDL/model/ERD、527が本人正規生成結果の受領。447は臨床5種＋薬剤、449は物販、④481/483/505等は各既存採用subset。530 exactや現C/Eへ無断追加せず、独立した後続受領とする。

下記fixture集合Fを**7resourceそれぞれ**へ対応づける。fixture作成/実装と実行結果は別、すべてNOT_RUN。

F＝POST省略/true/false/null・非boolean、PATCH省略時NULL/false/true保持、明示true↔false、旧field欠落/列欠落/DB失敗/元不明unknown、編集開始版競合/無権限/別院、falseのmap保存、create補償後実版、同TX監査rollback、旧writer/import保持、過去snapshot不変。

| resource | fixture・受領 |
|---|---|
| consultations | F＋既存version CAS/品目保険同時変更の同TX維持、447 |
| exam_types | F＋追加version guard/field別resource非拡張、447 |
| procedures | F＋既存version CAS/品目保険監査維持、447 |
| vaccines | F＋既存version CAS/365価格契約との独立、447 |
| checkup_types | F＋checkups grant/既存version CAS、447 |
| medicines | F＋追加version guard/薬剤baseと用量pattern分離、447 |
| merchandise_items | F＋追加version guard/530 exact限定scopeへ混在なし、449 |

別resource案は7種のgrant/CAS/監査/原子性を二重化するため不採用。既存resourceの所有writerを限定拡張して、履歴を更新しない方式を選定した。

<a id="b-a06-trimming-course旧2属性の明示解除"></a>
## B. A06 trimming course旧2属性の明示解除

### B.1 PATCH-only wireと既存互換

現 `PATCH /api/v1/masters/trimming-courses/{id}` に更新専用 `clear_course_type_id?:boolean` / `clear_target_size?:boolean` の2flagだけを追加する。POST/GET/永続modelにflagを入れず、新DDL/modelは不要。値null単独の旧保持を変えない。局所tri-state案はその旧null送信の意味を変えるため今回は不採用。

各flagは対応するfieldだけへ作用する。以下の表をcourse_type_id/clear_course_type_id、target_size/clear_target_sizeの両組へ独立適用する。

| value | 対応flag | 意図・結果 |
|---|---|---|
| 省略 | 省略/false | 保持。UPDATE mapキーなし |
| null | 省略/false | 旧保持。UPDATE mapキーなし |
| 正常な非null値 | 省略/false | 値設定 |
| 省略/null | true | DB NULLへ解除。FE正規形はvalue省略＋flag true |
| 非null値（0/空文字を含む） | true | 矛盾400、同梱項目も全体無書込 |
| 任意 | null/非boolean | flag型不正400、全体無書込 |

flag decoderも省略とnullを区別して検証し、nullを省略扱いにしない。優先順位で値を捨てて解除せず、両field同時操作も一方不正なら全無変更。値設定のcourse_type_idは既存正ID/同医院存在検査、target_sizeは既存small/medium/large/cat。0/空文字/none enumを解除sentinelにしない。2fieldの連動解除、inactive参照の新しい禁止を追加しない。

### B.2 保存・読取・現在認可

service update mapはflag true時だけ対象column:nil、値設定時column:value、保持時はキーなし。既存TX/現在scope/actor・画像audit等を維持する。既存master-trimmingのview/edit等を変えず、新grantや権限の読み替えをしない。PATCH検証後の片側書込・先commitを許さず、既存最外TXで全体保存する。

GETのnullable2fieldは現responseのomitempty省略を維持でき、専有FEが指定なしへ正規化する。Aのrequired nullable bool readをBへ持ち込まず、全GETのnullable表現を一括変更しない。course_type_idをAPIへ現実装どおりoptional nullable正IDとして収録し、PATCH-only2flagsと上表は更新専用schema等へ記載する。POST/GETへflagを漏らさない。API正本の編集は今回未実施。

#731採用のshort_name/image_object_keyは空文字、allowed_weekdaysは空配列が明示解除で、省略/nullは保持。この3属性は別契約で意味を変えず、#731がclear flag方式だったとは書かない。矛盾flag+value拒否の既存技術patternはreservation_type_service_core.goのclear_parent_id+parent_id拒否であり、#731に存在すると誤記しない。

### B.3 FE・owner・実装受領

FE update builderはCreate builderからtarget_size/course_type_idの2propertyを除き、dirty/初期値差分で未編集は省略、指定なしへの変更は対応flag trueだけ、値指定は値だけを送る。既存shortNameChanged/weekdaysChanged/imageChanged等は維持する。新規未指定NULLと更新解除を区別し、BE対応前のFE有効化を禁止する。再GETで保存確認する前に成功/指定なし反映済みと断定しない。

448の専有後続、①APIは445の単一ownerと446との必要契約引渡しで収録する。現Dのdescription対応とは別。request-only flagを生成modelへ手編集せず、全C06/CASや新表を本件の先行条件にしない。過去予約/施術/会計snapshotを変更しない。

fixtureは2field各々の値→指定なし→GET/再編集、解除後別属性更新、未編集保持/旧null保持、両flag同時解除/設定、矛盾・不正型・他院/無権限の全体無変更、#731の3属性回帰、画像履歴不変。全てNOT_RUN。実装・API収録・配備/FE接続・受領・実試験は未完であり、文書採択は機能公開ではない。

## 固定根拠（9eef73fe、採用済み実装と本稿の目標は区別）

- [業務：第六群/第十群](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/docs/spec/domains/accounting/lifecycle.md#L88)
- [既存技術対応](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/docs/design/domains/accounting/technical-spec.md#L170)
- [診療品目](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/docs/spec/screens/settings/master-treatment.md#L11)
- [薬剤](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/docs/spec/screens/settings/master-medicine.md#L13)
- [物販](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/docs/spec/screens/settings/master-merchandise.md#L13)
- [④既知未設定/unknown](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/docs/work/emr/emr-481-source-binding-contract-20261021.md#L41)
- [nullable/guard技術pattern（保険defaultは非転用）](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/migrations/004_master_insurance_versions.sql#L1)
- [presence-aware bool](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/medicalrecord/insurance_setting_request_field.go#L15)
- [保険だけの既存監査helper](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/medicalrecord/master_insurance_audit.go#L24)
- [既存route/grant](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/medicalrecord/routes_masters.go#L10)
- [物販route/grant](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/inventory/routes.go#L19)
- [旧属性の指定なし](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/docs/spec/screens/settings/master-trimming.md#L18)
- [旧PATCH DTO](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/trimming/trimming_course_request.go#L41)
- [旧update map](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/trimming/trimming_course_service.go#L83)
- [現GET省略形](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/trimming/trimming_course_response.go#L8)
- [旧FE builder](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/frontend/src/features/master/routes/trimming-settings-model.ts#L30)
- [矛盾flag拒否の技術pattern](https://github.com/MinoruSoga/AnimalEkarte/blob/9eef73feac19c25571724372b0b19da160ba5cfb/backend/internal/reservation/reservation_type_service_core.go#L103)
