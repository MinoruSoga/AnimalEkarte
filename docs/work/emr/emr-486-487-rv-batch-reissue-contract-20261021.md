# EMR-486 / 487 RV4文書セット・一括固定版・再発行日の追加選定契約

2026-10-21 R3（独立review 8点反映）。親採択済みの技術契約。runtime未実装・受入未完であり、本人が物理方式を直接指定した意味ではない。実装・API登録・DDL適用・試験・外部データ送信なし。試験NOT_RUN。既存v1を意味変更せず、新しい版付き契約の必要差分を明示する。

## 1. 根拠、採択範囲、未受領資料

読取main=e71ddf03462a0ee7a9864ad2ce6854b243591770。Plane486本文と全5コメント、487本文と全10コメントを確認。#790 head66b0b2ae85b5f94635b2382c31734f49065d727fはclosed/未merge、採択MTG文書は集約#784 head2ff5830297f85207f383e56b7421e0f15936c941で確認。#784は未merge。この固定値は読取時点であり最新作業完了を意味しない。

既決業務はRV4種類を一括出力できるmaster、再発行の表示発行日は当初値default＋変更可、接種日は不変、通常縦横は同edition、料金4欄optional/null/0・手入力・合計非自動・会計逆同期なし、診療packageと帳票セットは別、現在認可/旧訂正版有効出力拒否。

資料待ちは4種類の正式名称、各見本、既知単一RV見本との対応、印字欄/配置/紙寸法/ページ数/資料に依存するfield mapping。種類名、法定様式、医学値、新しい必須欄を推測しない。4種類を4ページとしない。既知rabies A4横を未受領の4種類へ一括適用しない。

既存読取3routeとrenderer/FE成果、#791の内部source version一致修正を再実装しない。EMR526の独立小修正をRV完成待ちへ戻さない。元ローカル調査稿の『FE小票保留/RV合流』は撤回済み。

## 2. identityの分離とセットmaster

kind=normal|rabiesは既存系統分類、layout=portrait|landscape|rabiesは既存v1の表示方式。新RVの4種類はこのenumを未知名称で拡張するのでなく、document_type_id(UUID)で識別する。position 1..4、名称、template ID、版番号とは別identity。

- document_type_id: 同医院内で安定した文書種別。名称変更/セット並替/テンプレート改訂で変えない。資料未受領の仮identityは名前nullで保持できるが本番出力不可。『資料待ち #1』のようなUI補助表示は正式名称や帳票本文へ保存しない。旧単一rabiesをどのtypeへ対応付けるかは資料照合前に決めない。
- set_id: 医院に一つのRV帳票セットidentity。診療package IDや接種種類IDを兼用しない。
- set_revision: 不変構成版。正整数、順序付き4member、document_type_id重複不可、positionは1..4全件。memberは型identityとtemplate_revision_idを結ぶ。draftはtemplate参照nullを許し、active化は全件解決済みだけ。
- template_revision_id: code-owned不変template registryのopaque UUID。document_type_id、正式表示名、schema/field-map版、template digest、renderer/font/asset manifest、各pageの幅高さ・向き・ページ構成を固定。自由なスクリプト/HTMLや外部URLを職員入力として実行しない。
- setのactive_revision_idはmasterの現在設定。保存済batchの構成は固定revision参照であり、master更新で書換え/失効/再発行しない。renderer資産が失われた場合は出力不能を示し別版へfallbackしない。

最小master表（新設する設計、実在ではない）:
1. vaccination_certificate_document_types: id uuid PK、clinic_id bigint、display_name text NULL、version bigint>0、created_at、updated_at。UNIQUE(id,clinic_id)。名称以外のsemantic type付替えは禁止。正式名称未確定はNULL。
2. vaccination_certificate_template_revisions: id uuid PK、clinic_id、document_type_id uuid、revision bigint>0、definition jsonb object、definition_sha256 bytea32、created_at。UNIQUE(id,clinic_id,document_type_id)、UNIQUE(clinic_id,document_type_id,revision)、type scoped FK RESTRICT。INSERT-only。definitionのclosed fieldsは{schema_version:1,document_type_id,display_name,snapshot_schema_version:2,field_map_version,template_sha256,renderer_version,font_asset_version,asset_manifest_sha256,pages:[{position,width_mm,height_mm,orientation}],display_issue_date_mapping}。pagesは1件以上/順序固定、幅高さは正の正確decimal string、orientation=portrait|landscape。display_issue_date_mappingはverified field-map内のfield key又はnull（欄なし）、恣意的な欄追加を許可する値ではない。資産参照は承認済みprivate registry内だけ。未知見本のdefinitionを合成してactiveにしない。
3. vaccination_certificate_sets: id bigint PK、clinic_id UNIQUE、version bigint>0、active_revision_id bigint NULL、created_at/updated_at。UNIQUE(id,clinic_id)。新構成revision append又はactive pointer変更でversion+1、no-op不変。type renameはtype.versionだけを進める。active FKは(active_revision_id,clinic_id,id)→set_revisions(id,clinic_id,set_id)、NULLは未active。
4. vaccination_certificate_set_revisions: id bigint PK、clinic_id、set_id、revision bigint>0、created_by、created_at。UNIQUE(id,clinic_id,set_id)、UNIQUE(id,clinic_id)、UNIQUE(clinic_id,set_id,revision)、set scoped FK RESTRICT。INSERT-only。
5. vaccination_certificate_set_members: clinic_id、set_revision_id、position smallint1..4、document_type_id uuid、template_revision_id uuid NULL。PK(clinic_id,set_revision_id,position)、UNIQUE(clinic_id,set_revision_id,document_type_id)、revision FK(set_revision_id,clinic_id)→set_revisions(id,clinic_id)、type FK(document_type_id,clinic_id)→document_types(id,clinic_id)、template FK(template_revision_id,clinic_id,document_type_id)→template_revisions(id,clinic_id,document_type_id)、全てRESTRICT。INSERT-only。deferred constraintでcommit時に4member/重複なしを保証。active pointerは全4template解決/同type対応/digest整合をguard検査。

全表tenant RLS＋明示clinic predicate、通常DELETE禁止。created_by/issued_by/performed_byはstaffs(id) FKで、trusted staffの現在clinic assignment/権限を検証する。staffs(id,clinic_id)複合FKを捏造し多院職員を拒否しない。時刻はserver、FK/unique/immutable guardと同時にmodel/ERDへ反映する設計。types/setsの可変列はactor/version必須の専用master serviceだけが更新し、immutable revision/memberは更新しない。任意の4種類以外をactiveにできる拡張機能は今回追加しない。

master wire（/api/v1除外）:
- GET /masters/rv-certificate-set: 未設定なら{set:null}、設定済みなら{set:{id,version,active_revision_id,revisions:[{id,revision,members:[{position,document_type_id,display_name,template_revision_id,resolution:"awaiting_material"|"resolved"}]}]}}。revisionsはactiveとlatestだけ、同一IDは重複せずrevision降順、最大2件。全履歴を埋め込まない。read only。
- GET /masters/rv-certificate-set/revisions?before_revision=<positive decimal>&limit=<integer>: 同じmaster-medical:viewと現在医院scopeのread-only履歴。limitは省略50、1..100。before_revision省略は最新から、指定時はrevisionが厳密に小さい行をrevision降順で返す。response exact={set_id,items:[{id,revision,members:[{position,document_type_id,display_name,template_revision_id,resolution}]}],next_before_revision:null|"<positive decimal>"}。各member shape/順序は上記と同じ、4件固定。limit+1件で継続有無を確認し、続きがあれば最終返却revisionを文字列で返す。終端はnull、空itemsはnull。revisionは同医院set内で一意・不変なので次頁に重複しない。新appendは既読頁へ混入させず、最新へ戻る操作で再取得する。未知query、重複query、非整数/範囲外/空cursorは400。未設定setの履歴は404、認可失敗・DB障害を空履歴へ変換しない。旧revisionを削除/改訂する操作は追加しない。
- POST /masters/rv-certificate-set: {operation_key}。初期draft setと4個のopaque type、4memberのrevision1を一体作成。名前/templateはnull、active=NULL。既存setありはdependency-changed、既存同key成功はreplay。本番設定作成を今回実行する意味ではない。
- POST /masters/rv-certificate-set/revisions: {operation_key,expected_set_version,members:[{position,document_type_id,template_revision_id}]}。全4件、既存同医院typeだけ。新immutable revisionを作成しset.versionを1進め、active pointerを勝手に変えない。
- POST /masters/rv-certificate-set/activation: {operation_key,expected_set_version,set_revision_id}。全4resolved確認後activeを変更。既に同activeならno-op/audit/receiptのみ。
- PATCH /masters/rv-certificate-document-types/{type_id}: {operation_key,expected_version,display_name}。正式名が判明後に設定、trim非空、同値no-op。type identityは変更しない。

権限は既存master-medical:view、初期master作成はmaster-medical:create、構成/名称/active変更はmaster-medical:edit。新grantは作らない。template registry資産の採用はcode-owned更新のため、master画面から任意asset uploadやライセンス承諾をしない。名称はtemplate revision descriptorにも固定し、現在type名を過去PDFへ引き直さない。

### master C06と監査の追加選定

owner_kind=vaccinations。intentsは vaccinations.certificate.set.create / .set.revise / .set.activate / .type.rename。既存master権限で新key preflight、現在元target権限でlookupし、write時は当該create/editを再検査する。master lookupは GET /masters/rv-certificate-operations/{operation_key}?intent=set-create|set-revise|set-activate|type-rename、認証clinic/本人staff scopeのみ。

Content exact={schema_version:"1",master_target_ref,command_sha256}。create targetはclinic内固定"rv-set"、他操作はset又はtype IDを含むtyped target文字列。hashはclosed requestからoperation_key/expected_version/expected_set_versionを除いた意味値とURL targetのJCS、empty createも固定typed objectをhashする。first createは同clinicのset不存在を再検査しuniqueで二重作成拒否。同key同内容は元receipt、CASのみlive、別内容はkey conflict、無効4構成はvalidation、対象差/変更はdependency。

vaccination_certificate_master_operation_targets（新設案）: operation_binding_id bigint PK、clinic_id bigint、set_id bigint NULL、document_type_id uuid NULL、operation_kind text closed enum、canonical_command_bytes bytea NOT NULL、command_sha256 bytea32。createは両target=NULLで固定clinicの初期構成操作、set変更はset_id必須、renameはtype_id必須。binding同医院FKだけ、business FKなし、clinic RLS、immutable、winner-only ClaimMetadataWriter。現在target認可はmetadataから解決する。新revisionはset row、renameはtype row、active変更はset rowをFOR UPDATEし、Gate→binding→master rowの順。type/template依存は同医院固定refsを検査、cross-clinicを許さない。

同TX: master変更/新revision4member→audit_logs→C06 receipt。audit resource=vaccination_certificate_master、resource_id bigintはset操作で数値set ID、type.renameはNULL、actionはcertificate.set.create/revise/activate又はcertificate.type.rename、actorはtrusted staff。old/newは{set_id,set_version,active_revision_id,document_type_id,type_version,display_name}の操作該当値だけ（rename前後に同じUUIDを固定。UUIDを数値変換/hashしてresource_idへ入れない）、metadata={operation_id,set_revision_id:null又はID,document_type_id:rename時UUID又はnull,member_type_ids:4件又は空,no_op}。no-opは業務版不変・audit/receiptのみ。source接種/証明書発行/会計へwriteしない。

成功refsはcreate/revise/activate={set_id,set_revision_id}、rename={document_type_id}、versionsはset_version又はtype_version。操作の変更なしなら既存revision ID。共通C06 mappingは§6の新consumerと同じ。構成revision作成時はactiveを変えなくてもset version+1とし、revision一覧変化の競合を検出する。namespace既存registerにこのowner/intentを追加する設計で、既に登録済みとはしない。

## 3. edition/artifact/batchの保存

既存series/editions/artifacts/operationsをnormalとlegacy rabiesのために保持する。RV4種類用に別の全証明書DBを複製しない。

seriesへdocument_type_id uuid NULLを追加。normal/旧単一rabiesはNULL、新RV memberはkind=rabies＋非NULL。旧UNIQUE(clinic,vaccination,kind)はNULL用のpartial uniqueへ置換し、新RVにはUNIQUE(clinic,vaccination,document_type_id) WHERE document_type_id IS NOT NULL。同医院type FK RESTRICT。normal縦横のseriesは依然一つ。新旧系列を暗黙に統合しない。

editionsは既存edition/previous/source_version/snapshot/issued_by/issued_atを再利用し、origin_edition_id bigint NULLを追加（同医院同series FK）。初版（previous_edition_id=NULL）はNULL originをself-originとして読む。追加列がNULLの旧v1後続editionはprevious chainを同医院同seriesでrootまで辿ってoriginと初版display_issue_dateを解決し、欠落/分岐/不整合は停止する。NULLというだけで旧後続版自身を初版にしない。推測backfillや旧snapshot更新なし。再発行は最初のoriginを保持し、previousを直前版にする。snapshot schema v2はv1のsource/owner/pet/clinic/inputs/eligibilityを保持し、document_type_id(UUID又はnormalではNULL)、display_issue_date、original_display_issue_date、render_manifest_v2を持つ。original_display_issue_dateは初版の表示日で、連続再発行で直前日へ置換しない。未知欄追加は資料待ちであり、inputsは既存10欄のまま、全キー明示null/0を保存する。

render_manifest_v2のclosed unionを次で固定する。Descriptor exact={template_revision_id,template_definition_sha256,renderer_version,font_asset_version,asset_manifest_sha256}（UUID、lowercase hex64、非空版文字列、非空版文字列、lowercase hex64）。normal branch exact={kind:"normal",layouts:{portrait:Descriptor,landscape:Descriptor}}、RV branch exact={kind:"rv_document",document_type_id,descriptor:Descriptor}。kind/branchの混在、normal片layout欠損、NULL descriptor/余分なlayoutキーを拒否する。normal Descriptorは既存code-owned normal template registryへUUID revisionとdefinition digestを付けて固定する限定拡張で、RV-only template_revisions表へ参照しない。normalのためのRV typeは作らない。既存normal portrait/landscapeの双方を同editionに固定し、縦横切替でeditionを増やさない。SnapshotV2 exact fieldsは{schema_version:2,kind,source,owner,pet,clinic,inputs,eligibility,document_type_id,display_issue_date,original_display_issue_date,render_manifest_v2}。source/owner/pet/clinic/inputs/eligibilityのnested fields/NULL制約は固定参照のv1 schemaをそのまま継承し、旧render_manifestはv2へ重複保存しない。normalはdocument_type_id=NULLかつnormal manifest、rabiesはUUIDかつ一致するRV manifest。旧v1 snapshotは変換/UPDATEしない。v2 readerを別schema dispatchで追加し、未知schemaは推測描画せず停止。

artifactsへlayoutにrv_documentを追加する新CHECKを選定。これは紙向きではなく『新RV memberの固定template出力』という技術discriminator。既存rabiesを再定義しない。既存unique(clinic,edition,layout)、object_key/hash/renderer版は維持。rv_documentのtemplate_versionはtemplate_revision_idのcanonical UUIDを保存し、editionの固定manifestと一致を検査。既存v1 readerへ新layoutを流さない。

新batch表:
- vaccination_certificate_batches: id bigint PK、clinic_id、vaccination_id、batch_revision bigint>0、source_version bigint>0、set_revision_id bigint、previous_batch_id bigint NULL、origin_batch_id bigint NULL、display_issue_date date、original_display_issue_date date、manifest jsonb object、manifest_sha256 bytea32、merge_renderer_version text、merged_object_key text、merged_content_sha256 bytea32、issued_by bigint、issued_at timestamptz、operation_binding_id bigint UNIQUE。UNIQUE(id,clinic_id)、UNIQUE(id,clinic_id,vaccination_id)、UNIQUE(clinic_id,vaccination_id,batch_revision)。set revision FK(set_revision_id,clinic_id)→set_revisions(id,clinic_id)。previous/origin FKは各(previous_batch_id又はorigin_batch_id,clinic_id,vaccination_id)→batches(id,clinic_id,vaccination_id)。vaccination/bindingは同医院scoped FK RESTRICT、actorはstaffs(id) FK。初batchはorigin=NULLをself-originと解釈。
- vaccination_certificate_batch_members: clinic_id、batch_id、position1..4、document_type_id uuid、edition_id bigint、artifact_id bigint。PK(clinic_id,batch_id,position)、UNIQUE(clinic_id,batch_id,document_type_id)、同医院batch/type/edition/artifact FK RESTRICT。artifact FKは既存(id,clinic_id,edition_id)へ対応。
- vaccination_certificate_batch_operations: clinic_id、operation_key uuid、operation_binding_id bigint PRIMARY KEY、operation_kind issue|reissue|reprint、batch_id bigint、performed_by、committed_at。同医院binding/batch FK RESTRICT。operation_keyはbindingのkeyとの一致検査用であり、(clinic_id,operation_key)のunique/PKを追加しない。認証actorとintentを含むC06 scopeの別bindingは同じUUIDでも別操作とする。旧単一certificate_operationsへ4つのedition/artifactをNULLや代表1件で詰め込まない。新rowは共通C06 bindingだけを再送正本とし、第二request hashで競合判定しない。

新表は全て成功確定後INSERT-only、tenant RLS、DELETE不可、success receipt必須deferred制約。batch/memberのdeferred guardは4type/順序/set revision一致、全editionとartifact同院同接種/元source版一致、内部snapshot.source.version=edition.source_version、template/hash一致、origin/previous系譜と連番を検査。batch manifestは {schema:1,vaccination_id,source_version,set_revision_id,display_issue_date,original_display_issue_date,members:[{position,document_type_id,edition_id,artifact_id,snapshot_sha256,content_sha256}],merge_renderer_version,merged_content_sha256}。自己hashを含めない。

batchは同接種のRV発行系譜を一つ持つ。master set revision変更があっても履歴batchを変更しない。新RV発行/reissueは4memberを同じTXで新しいeditionへ進める。変更が一部文書だけでも新batchは全4の固定組を持ち、他操作の『成功済み一部edition』を黙って流用しない。同key成功replayは新editionを一切増やさない。

## 4. 表示発行日、初回、再発行、再印刷

初回issueは既存規則のJST当日をpreviewで固定し、日跨ぎは再preview。初回の任意過去日入力を新設しない。

再発行は元の最新batch（normalは最新edition）を明示。defaultはorigin batch/editionのoriginal_display_issue_dateであり、元の直前版の表示日ではない。UIがdefaultを出した後も送信値は必須明示とし、表示日変更後は再preview・再確認。日付は実在するGregorian YYYY-MM-DD、年0001..9999の暦日として検証し（形式/型の境界であって医療・法的期間制限ではない）、日付を時刻や接種日へ変換しない。再発行に初回の『今日と一致』を要求せず、原日/過去日を拒否しない。法的期間、接種日以上、未来日禁止など未承認の業務日付境界を追加しない。これは表示日を操作日時へ偽装する許可ではない。

issued_at/recorded_atはserverの実際のcommit操作時刻。新batchとその4editionは同じtrusted timestampを使う。表示日を過去に変えても実timestampを過去値で上書きしない。接種日administered_onは変更不可。

再発行のsnapshot_basisを saved 又は current_source のclosed enumにする。savedは元固定snapshotの共通情報/inputsをコピーし、明示入力した表示日/既存入力欄だけを変更する。現在source版不一致なら拒否。接種訂正後など現在情報を使う場合はcurrent_sourceを明示して現resolverから新snapshotを作り、その全内容をpreviewで確認する。保存版の不足値を現在情報で埋めるfallbackはどちらにもない。current_sourceでもoriginal表示日とorigin系譜を保持する。基本情報更新だけで既存版を自動変更/失効しない。

再発行で確認内容が元版と全く同じ（日時originや実issued_atを比較対象にしない）なら、previewはunchanged=trueと元batch/editionを示し、issue/reissue POSTはvalidationで新しい版を作らず同版再印刷導線へ案内する。日付以外の明示入力変更は既存訂正の新版保存。normal縦横を替えただけもunchanged。

RV4memberのdisplay_issue_dateは同じ一つの確認日をsnapshotへ保存する。4種類別々の日付選択は今回追加しない。各見本に発行日欄があるか/どう表示するかは後日mappingで確定し、既知RVへ独立発行日欄を捏造しない。

## 5. RV v2 wireとnormal再発行

全pathは/api/v1を除く。既存wireのID/source版/edition版はinteger/int64を維持。新UUID identityのみUUID、C06 refs/versionsはdecimal strings。FEはunsafe integerを丸めずlossless decode/境界拒否。hash projectionだけ全integerをcanonical decimal stringへ変換する既存JCS方式を維持。全DTO closed、未知property/重複key/不正UTF8拒否。date/UUID/hashの型規則を固定し、opaque IDから名前/法的種別を推定しない。

全preview（RV batch・normal v2・保存版preview）のJSON `pdf_base64` と、batch/normal v2のcontent GETは[既存HTTP節のPDF wire規則](emr-486-certificate-api-contract-20261020.md#http-9操作)を継承する。`pdf_base64` は標準base64（必要な末尾paddingあり、改行なし）、content呼出側は `Accept: application/pdf` を指定し、200は `application/pdf` binary bodyとして読む。非成功は既存error契約を維持し、新たな406判定やwrapperを追加しない。BEでの実対応は担当実装の残件とする。

- GET /vaccinations/{id}/rv-certificate-context
  {vaccination_id,source_version,eligibility,can_issue,set:null|{id,version,active_revision_id,members},latest_batch:Batch|null,legacy_single:null|{latest_edition_id,source_version,validity},initial_display_issue_date,blocking_reasons}。setはrequired nullable。membersはactive_revision_idがあればその固定revision、active=NULLならrevision最大のdraftを返す（設定済setにrevisionなしは整合エラー）。members exact={position,document_type_id,template_revision_id,resolution}、resolution=awaiting_material|resolved。未設定set=NULLはblocking_reasons=["set-not-configured"]、設定済みactiveなしは["set-not-active"]、未解決active参照は["template-material-pending"]とし全てcan_issue=false。その他のclosed reasonsは"source-ineligible","legacy-mapping-pending"で、複数理由はこの列挙順に返す。DB不整合/DB障害/認可失敗をset=NULLや資料待ち200へ潰さず既存ErrorResponseで停止。eligibility/validityは既存閉enum構造を継承する。GETでsetやclaimを作らない。legacy_singleは同医院旧document_type_id=NULLのrabies系列最新版metadataだけで、なければNULL。4memberへ補完しない。
- POST /vaccinations/{id}/rv-certificate-preview
  {mode:"issue"|"reissue",parent_batch_id:null|ID,set_revision_id,snapshot_basis:"current_source"|"saved",display_issue_date:null|date,inputs:CertificateInputs,preconditions:{source_version,latest_batch_id:null|ID,set_version}}。
  issueはparent=NULL、basis=current_source、date=NULLを要求、最新batchなし。reissueはparent=最新batch、date必須、basis明示。source更新や別set revisionへの確認済み置換もreissueに含める。通常の再発行は親set revisionをdefaultにし、master最新revへ黙って置換しない。別rev選択は4type集合が一致する時だけこの再発行APIで扱い、type集合変更/legacy単一→4の対応は後述の資料mapping停止条件へ。
  response {draft:BatchDraft,preview_fingerprint,unchanged,preview:{media_type:"application/pdf",pdf_base64,member_page_ranges:[{position,first_page,last_page}]}}。BatchDraftは§6のexact型と同一。page番号は1始まりinclusive、全4rangeはposition順で連続・重複なし。全4renderと結合が完了した場合だけ200。page rangesは実PDFから算出。previewにprint buttonを置かず、部分成功PDFを最終出力としない。previewはDB業務write/正本uploadなし。
- POST /vaccinations/{id}/rv-certificate-batches
  preview request全fields＋{operation_key,preview_fingerprint}。modeでissue/reissue intentを選択。client snapshotは受けない。期待source/latest/set版と確認済み内容を照合。成功はC06 ResultResponseのみ。
- GET /rv-certificate-batches/{batch_id}: Batch={id,vaccination_id,batch_revision,set_revision_id,previous_batch_id,origin_batch_id,display_issue_date,original_display_issue_date,source_version,issued_at,issued_by,validity,members:[{position,document_type_id,edition_id,edition,artifact_id,snapshot_sha256,content_sha256,edition_detail}],manifest_sha256}。各memberへedition_detail={id,series_id,vaccination_id,kind:"rabies",document_type_id,edition,source_version,previous_edition_id,origin_edition_id,issued_at,issued_by,snapshot:SnapshotV2,snapshot_sha256}を含める。member.edition_id=edition_detail.id、member.edition=edition_detail.edition、member.snapshot_sha256=edition_detail.snapshot_sha256を必須とし、FE/BEとも各edition.source_version===snapshot.source.versionを検査する。historyの内部一致を現在context.source_versionとの一致へ置換しない。originは初batch/初editionでもself IDに解決して返す。history閲覧可、旧版へ有効content pathを配らない。
- GET /vaccinations/{id}/rv-certificate-batches?page&limit:既存page/limit envelope、batch_revision降順。
- POST /rv-certificate-batches/{batch_id}/reprints: {operation_key,preconditions:{source_version,latest_batch_id}}。現在有効batchの既存5artifactをdigest検証し、版不変で再印刷receiptだけ作る。欠損なら503、現在templateで再生成しない。
- GET /rv-certificate-batches/{batch_id}/content: 結合PDFだけapplication/pdf、private/no-store。全4member現在有効性と結合digestを返却直前にも検査。
- GET /vaccination-certificate-operations/{operation_key}?intent=rv-batch-issue|rv-batch-reissue|rv-batch-reprint:既存lookupへ新allowlist。元anchor現在認可後のみ返す。

通常証明書の表示日変更は POST /vaccination-certificate-editions/{edition_id}/reissue-preview と /reissues を追加。preview request={snapshot_basis,display_issue_date,inputs,layout,preconditions:{source_version,latest_edition_id}}、reissue requestはoperation_key/preview_fingerprintを追加。normalのみ、このrouteへRV memberを渡して1文書だけ再発行しない。response exact={snapshot:SnapshotV2,snapshot_sha256,render_manifest_sha256,preview_fingerprint,unchanged,source_version,latest_edition_id,layout,pdf_base64,media_type:"application/pdf",template_version,renderer_version}、success refsは既存vaccination_id/series_id/edition_id/artifact_id、versionsはedition/source_version。旧同版reprint routeと意味を混ぜない。normal両layoutのmanifest固定/同editionは維持する。

### normal v2 reader/FE互換の明示切替
既存pathのquery contract_version=2を新consumer選択子として追加する。省略又は1は旧v1、その他の値/重複指定は400 validationのpre-C06 ErrorResponse。新selector対応範囲は GET /vaccinations/{id}/certificate-context、GET /vaccination-certificate-editions/{edition_id}、GET /vaccinations/{id}/certificate-editions?kind=normal、POST /vaccination-certificate-editions/{edition_id}/preview、POST /vaccination-certificate-editions/{edition_id}/reprints、GET /vaccination-certificate-artifacts/{artifact_id}/content、GET /vaccination-certificate-operations/{operation_key}?intent=issue|reissue|reprint。API schema/operationIdは既存pathを重複登録せずquery branchを明記する。
- selector=2のnormal read/context.latest/historyはSnapshotV1|SnapshotV2をschema_versionでdispatchし、edition envelopeの既存fieldsを維持する。v2ではorigin_edition_idを解決済みself/root IDとして追加。旧v1はwireのまま返しstored snapshotをv2へ変換しない。全件で内部source版validatorを再利用する。未知schemaは409 dependency-changed。context.latest.rabiesは旧NULL-type系列のみで新RV memberを混ぜない。RV4は専用contextから取得する。
- selector省略/1は既存v1 responseを維持。v2が必要なedition/context.latest/history pageを読もうとした場合はresponse全体を既存409 dependency-changedで止め、v2をv1に丸める/隠して古い版をlatestと返す/部分page成功はしない。旧FEへv2 JSONを黙って送らない。
- selector=2の保存版preview/content/reprintは同normal editionのschema分岐で現在有効性と固定manifestを検査する。既存artifactはdigest検証read。未生成のnormal別layoutは同editionの固定Descriptorで候補を準備し、reprintでartifact＋operation＋audit＋receiptを同TX保存する。previewは純描画のみ。新editionなし、現template fallbackなし。contentは保存済artifactだけで再生成しない。成功refsは既存4refs、versionsはedition/source_version。reprint Contentのexact4keysは旧版と同じedition_id/layout/snapshot_sha256/render_manifest_sha256、v2 hashは新manifest branchを使用。reprint intentも既存vaccinations.certificate.reprintを維持しselectorはContentへ混ぜない。
- 新normal再発行2routeは常にv2。旧initial issue/編集previewはv1を維持し、対象系列latestがv2になった後の旧writer呼出は409 dependency-changedで新再発行導線へ案内する。既知key成功receiptの照会/replayをこの制限で未成功へ戻さない。新FEは初回v1発行結果もselector=2 readerで読める。
- 配備は先にselector reader/parser/sharedvalidatorと両schema描画/reprint/content/lookupを揃え、最後にnormal reissue UI/writerを有効化する。同医院で旧FE利用中はwriter activationを保留し更新案内を出す。古い開いたtabの要求は上記409で安全停止する。未準備の読取先へv2保存を先行させない。RV資料の待機とは独立した必要subsetであり、既存#791修正は再利用する。

新RV memberの単独有効content/reprint/issue公開routeは追加しない。既存v1でdocument_type_id非NULLのseriesを扱う要求は拒否する。4文書の構成確認・履歴metadataは見られるが、新RVの最終出力はbatch全4のみ。v1のnormal入口は維持。

## 6. C06、durable payload、認可、競合

新RV intentsは vaccinations.certificate.rv-batch.issue / .reissue / .reprint。normal追加intentは vaccinations.certificate.reissue。owner_kind=vaccinations、trusted clinic/staff actorを使用しclient scopeを受けない。

query intentからScope.Intentへの固定対応（略名はURL queryだけ、C06 scope/enumとして流用しない）: set-create→vaccinations.certificate.set.create、set-revise→vaccinations.certificate.set.revise、set-activate→vaccinations.certificate.set.activate、type-rename→vaccinations.certificate.type.rename。rv-batch-issue→vaccinations.certificate.rv-batch.issue、rv-batch-reissue→vaccinations.certificate.rv-batch.reissue、rv-batch-reprint→vaccinations.certificate.rv-batch.reprint。既存issue→vaccinations.certificate.issue、reprint→vaccinations.certificate.reprint、新normal reissue→vaccinations.certificate.reissue。未知queryは400 validation、別namespaceのintentを転用しない。progressはrv-batch-issue/reissueだけ。全Scope.OwnerKindはvaccinations。

Content(batch issue/reissue) exact fields={schema_version:"2",vaccination_id,parent_batch_id:"none"又はID,set_revision_id,source_content_version,batch_draft_sha256,preview_fingerprint}。Content(batch reprint)={schema_version:"2",batch_id,manifest_sha256,merged_content_sha256}。normal reissue={schema_version:"2",vaccination_id,parent_edition_id,layout,source_content_version,snapshot_sha256,render_manifest_sha256,preview_fingerprint}。いずれもC06の64fields/各256bytes以内。短いhashだけを使いPII全文をContentへ入れない。

hash対象はtyped draft/snapshotの全nested integer projection→RFC8785 JCS UTF8。BatchDraft exact={schema:2,mode,vaccination_id,source_content_version,set_revision_id,parent_batch_id,origin_batch_id,snapshot_basis,display_issue_date,original_display_issue_date,inputs,members,merge_renderer_version}。mode=issue|reissue、snapshot_basis=current_source|saved、parent/originは初回NULL/再発行は親/root ID、inputsは既存全10keysのCertificateStoredInputs。membersはposition昇順のexact4件、member exact={position,document_type_id,snapshot:SnapshotV2,snapshot_sha256,render_manifest_sha256}。snapshot内source.vaccination_id/version、日付、inputsとdraft共通値の一致を必須とし、4type順/manifest template refsはset_revisionへ一致させる。mode=issueではsnapshot_basis=current_sourceを要求する（field名はsnapshot_basisのみ）。将来edition/artifact ID、operation key、actor、実行timestamp、CASはdraftへ含めない。batch_draft_sha256=SHA256(JCS(integerProjection(BatchDraft)))、preview_fingerprint=SHA256(JCS(integerProjection({schema:2,draft:BatchDraft,output:"rv_batch_pdf"})))。merge版はdraft内の一箇所だけ。
private canonical_command_bytesのbatch write型 exact={command_schema:1,operation_kind:"issue"|"reissue",draft:BatchDraft,preview_fingerprint}、batch reprint型 exact={command_schema:1,operation_kind:"reprint",batch_id,manifest,manifest_sha256,merged_content_sha256}。manifestは§3 exact型。private commandから公開draftへはwrite時command.draftだけを一方向projectionし、private command全体をpreview responseへ出さない。command_sha256はそのtyped command全体のJCS(integerProjection) digest。pure projectorはwrite時draft各値とcomputed draft hash/fingerprint、reprint時batch ID/manifest hash/merged hashから上記Content exact fieldsを生成する。二重保持hashは再計算一致を必須とする。
normal reissue preview_fingerprint=SHA256(JCS(integerProjection({schema:2,parent_edition_id,snapshot_basis,layout,snapshot})))。snapshot/manifest hashはそれぞれSnapshotV2/render_manifest_v2全体。normal targetsは既存canonical_snapshot_bytes/canonical_render_manifest_bytesに加えreissue専用canonical_reissue_context_bytes bytea NULLを追加し、reissueだけ{schema:1,parent_edition_id,snapshot_basis,layout,preview_fingerprint}をimmutable保存する（旧issue/reprintはNULL）。pure projectorはこのcontextと固定snapshot/manifestからContentを再構成し、parent/layout/source/hash/fingerprintを全照合する。batch manifestのhashはcommit後の固定ID対応を別に証明する。自己参照hash/実行nowなし。

preconditions source/latest/set版と現在認可fenceはlive CAS。snapshotの確認source_content_version、set_revision_id、parent、日付/入力/manifestはimmutable内容。CASだけの再確認は同key、内容変更は旧結果を解決してから新preview/新key。samekey別内容は元target現在認可後operation-key-conflict。現在依存不一致はdependency-changed、CASはversion-conflict。成功済み同内容は現在source/master変更で未実行扱いへ戻さずreceiptを返し、新PDF取得は別の現在有効性gateへ。

新private owner metadata vaccination_certificate_batch_operation_targets を選定:
operation_binding_id bigint PK、clinic_id bigint、vaccination_id bigint>0、batch_id bigint NULL、operation_kind issue|reissue|reprint、canonical_command_bytes bytea NOT NULL、command_sha256 bytea32。唯一のFKは(binding,clinic)→operation_bindings(id,clinic)、RLS＋全列immutable。新bindingとClaimMetadataWriterで同じ短claim TXに保存、business row FK/read/lockなし。batch_idはissue=NULL、reissueは元batch、reprintは対象batch。snapshot/manifestはtyped command内に固定しhash照合、client captured DTOでApplyしない。successなしでも元target lookup可。

normal reissueは既存vaccination_certificate_operation_targetsのoperation_kindにreissueを追加、edition_id=元版必須、既存canonical snapshot/manifest bytesへv2を保存する。成功保存先vaccination_certificate_operationsもoperation_kind CHECKをissue|reprint|reissueへ限定拡張し、reissueは新edition/選択layoutの新artifact必須。既存binding非NULL/request_hash NULLの新writer guard、legacy request_hash保持、scope/key PK衝突拒否、audit＋receipt同TXの写像をそのままreissueへ適用する。新台帳/第二idempotencyを作らない。guardはtargetsの元editionと保存operationsの新editionを同一IDと誤比較せず、新previous_edition_id=元edition、同series/clinic/source/manifest、Content layoutとartifact layout、成功refs/versionsの新edition一致を検査する。normal reprintは既存edition維持で同guardの別branch。共通VerifiedCommandはrequest-memoryのみの採択追加案で、durable storageではない。本private command保存は本consumerの明示追加設計。495用owner-private commandが全ownerで利用可能とは扱わない。master/batch/normal-v2の当該commandも、現在元target認可後に検証した保存bytesを[採択済みQ](../../design/shared/operation-participation.md#verified-command-selected-20261021)へ載せ、Applyはreq.VerifiedCommandからだけtyped入力を再構成する。既存v1とContent-only consumerは変更しない。claim/Apply前にstored bytes→typed decode→canonicalize一致→pure Content projector全key一致を確認し、認可anchorが欠落/不整合/解決不能なら現Runner同様403 not-authorizedで非開示。元targetと現在開示認可が確定した後のprivate command/結果データの読取失敗・検証不能だけ503 result-unavailableで停止。

新issue/reissue/編集previewは既存vaccinations:view＋edit。batch read/history/content/reprint/lookupはview。元接種/患者/医院の現在認可を毎要求検査し、新grantなし。根拠は固定参照486 v1契約の権限表で、保存版preview/reprint/content/lookupはvaccinations:view、発行前編集preview/issueはview+editと既決。新batch reprintとnormal selector=2 reprintはこのviewを継承し、新発行/reissueにeditを要求する。既存他mutation routeのmiddlewareを一括変更しない。edit喪失後の発行結果回復はGET lookup/readに限定する。masterの権限は§2の既存master権限で、発行権限へ兼用しない。

C06 mappingは本新consumerで succeeded/replayed/not-succeeded=200、operation-in-progress=202、result-unavailable=503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、scope-not-found=404、not-authorized=403。非成功はstatus/operation_keyのみ、pre-C06は既存ErrorResponse。旧v1 selector省略/1の422/409 operation-in-progress/403 edit-right-lostは無断変更せず、新route及びselector=2 consumerのmappingと分ける。RV success refs={vaccination_id,batch_id}、versions={batch_revision,source_version}。4member refsや内部EffectSummaryをResultResponseへ追加せずbatch readerで取得。reprintも同refs/versions、履歴的な成功であり紙出力成功ではない。

初回同sourceに別keyで二重batch:最新batchなしCAS＋接種row lock＋batch revision uniqueで片方だけ成功。再発行もparent/latest CAS＋意味的unchanged検査で不要な二重editionを防止。batch operationのidentityはoperation_binding_idだけで、同UUIDを持つ別actor/intentのbindingを衝突させない。lookup/replayは元actor/clinic/owner_kind/intent/keyのC06 scopeと現在元target認可を通し、別scopeの結果を返さない。同接種の実際の競合は上述の現在batch/source CASで判定する。

## 7. 候補準備・部分失敗・全4確定

方式は『候補準備は部分再開、業務発行は全4一括commit』。結合PDFは旧#790で未選定だった方式を今回Dotsが技術選定するもので、本人がPDF結合形式を指定したという記録にはしない。個別4artifactとimmutable batch manifestを保持し、結合順はset revisionのpositionで固定する。4個別操作を順番にcommitして最後にbatch成功にする設計は不採用。1つのC06 operation/bindingに4member＋結合PDFを対応させる。

A. 元target現在認可→既存binding優先lookup。新keyのみ実行権限preflight、preview再構成/一致確認後claim+immutable command。既存bindingは保存内容で比較し、現在情報でcommandを作り直さない。
B. stored commandのhash/Contentを再検証、業務TX外で4memberの決定的PDFと結合PDFを準備する。候補object locatorはv2/clinic/binding/command_hash/member/{document_type_id}.pdf、結合は同prefix/batch/{merge_version_digest}.pdf。同locatorのexpected content digestを下記準備descriptorへ保存し、成功候補の発見にobject listingや毎回全件再renderを要求しない。これはprivate store内相対key、患者名/自由入力/資格情報/公開URLを含めない。既存approved certificate artifact store adapterへ渡し、bucket/資格情報作成・権限拡張は今回実行しない。
準備descriptorの追加表 vaccination_certificate_prepared_artifacts: operation_binding_id bigint、clinic_id bigint、ordinal smallint0..4（0=結合、1..4=set position）、document_type_id uuid NULL（結合のみNULL）、command_sha256 bytea32、object_key text、content_sha256 bytea32、byte_length bigint>0、verified_at timestamptz。PK(binding,ordinal)、binding同院FKのみ、RLS、INSERT-only。新snapshot/editionではなくprivate準備metadata。Putの前にdescriptorを書かず、bytes照合完了後の別短TXでINSERTする。既存descriptorと異なる再INSERTは停止し上書きしない。object書込成功→descriptor保存前の断では当該一件だけ同固定commandで再render/照合してdescriptorを補う。他の検証済descriptorは再read/digest照合して再利用する。descriptor保存失敗は発行成功にしない。

C. PutIfAbsent→read/hash照合で候補保存確認。同locator既存bytes一致は再利用、相違は上書きせず503。timeoutはHEADだけを信用せずstored size/digestまたはreadで照合。未確認で別keyへfallbackしない。決定的renderer/font/merge manifestが同じでもbytes不一致なら停止。PDF実行now/乱数/host fontなし。
D. 一部成功/一部失敗は『準備済み候補』『未準備』『状態不明』として区別、C06成功なし/editionなし。認可付きoperation progress読取はmemberのposition/typeと preparation_state=prepared|missing|unknown、batch_state=prepared|missing|unknownを返せるが、発行成功と表示しない。再試行は同binding/commandで既存候補digestを照合し不足だけ準備。成功済み業務editionの再発行ではない。
E. 全5artifact照合後、owner Gate→C06 binding NOWAIT→D08既存parent/child→vaccination→必要series ID昇順/セット参照の順でlock。set/全4type/元batch/latest/source/認可/候補digest再確認。初series作成は同vaccination lock下でunique。immutable 4edition＋4artifact行＋batch＋4member＋batch operation＋必須audit＋C06 receiptを同TXで保存。新editionのissued_atは同server instant。外部render/uploadをTX内へ入れない。
F. commit応答断は同key lookup、succeededならbatch readerへ。unknown/busyは新keyも自動再Applyも禁止。definitive not-succeededのみ現在認可/依存を再照合し、同内容同keyで候補から再開。発行途中に接種訂正がcommitすれば旧source候補を確定せず停止。旧訂正版を復活させない。

progress routeは GET /vaccination-certificate-operations/{operation_key}/preparation?intent=rv-batch-issue|rv-batch-reissue、現在元targetのview必須。response={operation_key,preparation_only:true,members:[{position,document_type_id,state}],batch_state}。成功/失敗の正本は別C06 lookup。preparedはdescriptorとobjectが一致、missingはdescriptorなし＋object不在を確定できた場合、その他はunknown。object locator/blob URL/PIIを出さず、provider不明をmissingに落とさない。readでrender/upload/claimを開始しない。

## 8. storage寿命・孤児回収・read有効性

候補の正本locatorはclaim commandから決定する。候補成功の有無だけをDB発行成功へ昇格しない。artifact storeはprivate Get/PutIfAbsent/StatDigest/DeleteCandidateのadapter。既存環境のnil storeや未採用fontのまま本番利用可能とは言わない。provider実結合・適合埋込font/ライセンス資産の導入は既存runtime依存であり、4様式資料待ちと別に明記する。

本稿は保存済edition/artifact/batchを削除しない。候補cleanup ownerは486のcertificate artifact maintenance component。自動TTLだけでは削除しない。C06 bindingの現在状態が成功参照ありなら永久候補扱い禁止、unknown/busy/実行可能な未成功なら保持。明示recovery fenceでclosed_without_effectになったbindingに限り、全batch/member/artifactの参照不存在と未実行閉鎖を再確認して候補を回収可能にする。最大保持日数を勝手に定めない。閉鎖機能未実装なら候補を保持し自動回収しない。読取失敗・provider不明を参照なしと扱わない。

reconciliationはbinding→candidate keys→DB batch/member/artifact参照→実blob digestの順に照合。DB成功/欠損blobはartifact_unavailable、再描画fallbackやedition再発行で埋めない。blobあり/DB不明は保留、closed確定/DB参照なしだけ孤児候補として扱う。削除自体は今回実行せず、運用上のデータ削除承認を迂回する根拠にしない。

batch有効性は接種未削除・現在source版一致・同接種最新batch・4seriesそれぞれ最新edition・4内部version整合。優先順はsource_deleted > source_changed > superseded。master現在active revision不一致だけで保存済batchを失効させない。validity検査失敗をPDF再構成で救済しない。既に渡した紙/bytesの回収保証はしない。

audit_logsのactionはvaccinations.certificate.rv_batch.issue / .reissue / .reprint、resource=vaccination_certificate_batch、resource_id=batch ID。actor_type=staff、actor_id=trusted staff。old/new={batch_id,batch_revision,source_version,display_issue_date,original_display_issue_date}、metadata={operation_id,parent_batch_id,set_revision_id,member_edition_ids,member_artifact_ids,manifest_sha256,merged_content_sha256}。日付未変更を医療内容変更と偽らない。normal reissueは既存certificate auditの専用reissue action、元/新edition/元日/表示日/operation IDを同TX保存する。audit失敗は全DB rollback、候補は未参照のまま保持。

## 9. reload回復と既存共通契約への必要subset

既知keyを保持した再試行/部分準備再開は本稿で閉じる。新RV batchが旧issue/reprint用recovery controlに自動適合しているとはしない。

root採択済みだが読取時点未公開のrecovery契約へ新consumerを追加する設計: certificate target_refをtagged unionへ版付き拡張し、batch issue/reissueは{kind:"rv_batch_write",vaccination_id}、batch reprintは{kind:"rv_batch_reprint",vaccination_id,batch_id}。既存target_ref形式は別versionで維持。write controlはissue/reissueを同じrv_batch_write familyで直列化し、異なるintentへの切替で未解決操作を迂回しない。実C06 intentは保存metadataで区別する。normal新reissueも既存normal issueと同接種のwrite control familyへ統合する必要subsetとして扱う。

query/fenceの現在本人/元target認可、recovery_generation CAS、active operationの遅延旧送信閉鎖、success不変/unknown停止を維持する。新private batch target metadataをdiscovery対象へ足す。successなしもdiscoverできるようにし、成功batchだけを検索しない。control generationは実行preconditionでContentへ混ぜない。既存controlからの移行時に未解決旧操作がある場合はcutover停止し、新controlを空に作って抜けない。

実装前に必要なのはこの対象union/control adapter/metadata discovery参加であって、共通C06全文を新設することではない。未接続の段階はreloadでkey/対象を特定できなければ停止し、新keyで救済しない。localStorage等へ本文/keyを勝手に永続保存しない。未送信本文復元は保証しない。

## 10. 互換、資料mapping、着手subset

既存v1 normal/rabies/API snapshot v1/legacy rowsを保持。新RV document_type非NULLを旧v1 parserへ渡さず、v2 readerへ分岐。新RVが有効化された医院の新規RV発行/再発行/有効再印刷入口はbatchだけ。旧v1 rabiesのwrite/reprint/contentを迂回口として残さず、その医院ではdependency-changedでbatch導線へ案内する。normalのselector省略/1はv1 shapeを維持し、v2保存後の旧reader拒否/新selector移行は§5に従う。cutover前後の対応は医院単位のcapabilityとして明示し、未知資料の行だけ単一RVへfallbackしない。旧単一RVを4種類完了と表示しない。既知旧版は履歴readerに残す。旧単一RVから4種類へ移る際は、正式4typeと既知見本の対応が判明してから明示mappingを確認するまで出力migrationを停止し、4コピーや同版番号の捏造をしない。

FE487先行subset: type/set/member/batchのschema、closed DTO mock、全4準備状態の表示、日付default/変更/repreview、同版reprint判定、内部source版共通validator、null/0/fees維持、型不明/資料待ち表示、C06同key照会。保存済み接種詳細の共通証明書panelを拡張し、独立画面/FE PDF rendererを新設しない。資料なしでも合成type UUID・任意ラベルなしのmockで構造fixtureを作れるが本番見本完成としない。

BE486先行subset: master revisions/四件guard、v2 domain/reader/writer、C06 targetとpure command projector、候補準備orchestrator、全4原子commit、normal reissue/date lineage、現在認可、有効性、lookup/progress、既存source訂正writerとのlock閉包。446/466へ限定DDL/model/ERD、489へ既採択claim/VerifiedCommand/recovery adapterの必要subsetだけを依頼する。新grant・全DB完成待ち・別個C06実装を作らない。

487 server renderer subset: descriptor-based4member renderer interface、merge順/実page range、全4成功判定、digest/byte決定性の合成fixture。実見本mapping・各paper spec・field mapping・golden/実印刷だけを資料待ちにする。provider/fontの既存接続依存は実装依存として別管理し、『資料さえ来れば本番ready』と偽らない。

採択docsへの最小同期: D08 lifecycleの業務/9ACは変えず本稿へ技術参照。technical-specと486旧wire文書の『master/batch/date未選定』を本稿の選定範囲だけ更新し、既存v1限定契約は保持。api.yamlのv2新route/schema/operationIdは共有ownerがこのclosed contractから別実装PRで公開し、runtime not-implementedと区別する。今回#784に追加するのは文書のみ。

## 11. fixture（すべてNOT_RUN）

- 4欠損/重複/違うtype-template/異医院/未解決descriptorはactive/preview/issue拒否。4文書が6ページでも順序/範囲一致。
- 1memberが失敗したとき残り3候補を有効発行済みとしない。1〜3member候補だけ成功、merge失敗、PUT応答断、同key再開で既存一致bytes再利用、candidate mismatch拒否。DB発行は0又は全4、receipt1。descriptor直前/直後の断も当該候補へ収束。
- 全4commit後応答断、同key replay同batch、別key古いlatest CAS拒否、同内容reissueはunchanged、unknown新key禁止。
- 初日A→表示日Bへ再発行→次defaultA。original日/接種日/旧PDF不変、実issued_atは現在。日付変更後古いfingerprint拒否。
- normal portrait/landscape同edition、layoutだけ変更は再発行しない。normal10入力nonnull拒否、RV全10optional/null/0、合計非同期、料金逆同期/明細生成なし。
- master改訂で旧batch不変、保存済定義/資産不明は停止。接種訂正成功後旧4member/結合PDF拒否、訂正rollbackで有効性維持。
- normal v1/v2混在history、selector省略でv2遭遇409、新selectorの両schema成功、旧tab停止、別layout reprintでも同edition、operations reissue CHECK/guard/receipt整合、type UUID監査NULL＋metadata、set版append/activate/no-op、scoped FKをfixture化。
- source内部版不一致、member refs/digest混在、legacy単一を4複製、unknown schemaをv1に丸める動作を拒否。
- audit/member/receipt/deferred guard失敗全rollback、candidate回収は閉鎖証拠/参照不存在が必要。権限喪失/別actor lookup/元target不一致で情報非開示。
- recovery新union未接続時は未提供を表示し、未確定key喪失で新key救済ゼロ。ready/prepared/診療package applyから発行・接種・請求・在庫・通知ゼロ。

## 固定参照

- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/work/emr/emr-486-certificate-api-contract-20261020.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/internal/model/storage_clinical.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/backend/migrations/001_init.sql#L8315-L8400
- https://github.com/MinoruSoga/AnimalEkarte/blob/2ff5830297f85207f383e56b7421e0f15936c941/docs/spec/domains/vaccinations/lifecycle.md#fr-vaccinations-rv-batch-20261021
- https://github.com/MinoruSoga/AnimalEkarte/blob/2ff5830297f85207f383e56b7421e0f15936c941/docs/work/emr/emr-486-certificate-api-contract-20261020.md#certificate-mtg-contract-gap
- https://app.plane.so/baritechllc/browse/EMR-486/
- https://app.plane.so/baritechllc/browse/EMR-487/

共通VerifiedCommand/request-memory、495のprivate durable command、recovery controlのroot採択草稿は追加依存として照合したが、今回読取時点で全owner runtimeへ提供済み/公開済みとは認定していない。
