# EMR-486 / EMR-487 証明書 API 選定契約（2026-10-20）

状態: **v1 selected（読取・描画の部分実装は末尾参照）／10-21追加要件は下記の未収録差分／試験NOT_RUN**。依頼者・採用判断者はMinoruSoga。目的は確認した発行内容を不変版へ固定し、再転記・二重発行・旧訂正版の有効出力を防ぐこと。操作数・所要時間の改善実測は実装受入へ残す。採用済み案をHTTP/DTOへ転記する文書で、業務要件の追加や実装完了認定ではない。

機械契約は [backend/docs/api.yaml](../../../backend/docs/api.yaml) の `Certificate*` schemas と `VaccinationCertificates` operations。全操作の `x-contract-status: selected` / `x-runtime-status: not-implemented` を固定する。なお、セッション5の読取・描画PR以降、`getVaccinationCertificateEdition` / `listVaccinationCertificateEditions` / `previewSavedVaccinationCertificate` の3操作は `x-runtime-status: implemented` とする（list の runtime path は gin のワイルドカード命名制約により `GET /vaccinations/{id}/certificate-editions`。文書側の `{vaccination_id}` パラメータ名は変更しない）。API公開revisionはこの文書とYAMLを含むPR headで識別し、rootレビュー後にFEへ引き渡す。旧 `emr486-certificate-design-proposal.md` の曖昧なHTTP/receipt案より、本採用契約を優先する。

Fable向けの入力・帳票表示・実装着手境界は [2026-10-21 Q&A](fable-qa-20261021.md) を参照。

## 基準・参照と照合

- 最新main基準: `05da02f30cc5320c929095c6c9d6f96a4ac3f065`（PR748再診契約を含む）。PR746由来 `/me` 契約も保持。
- [Plane EMR-486](https://app.plane.so/baritechllc/browse/EMR-486/) 本文 updated `2026-10-20T14:19:59.834978Z` と1コメント、[EMR-487](https://app.plane.so/baritechllc/browse/EMR-487/) 本文 updated `2026-10-20T14:20:16.001698Z` と1コメントを読取。4料金欄optional手修正、snapshot/fingerprint固定、会計逆同期なし、合計自動計算なし、518独立を適用。
- [D08業務正本](../../spec/domains/vaccinations/lifecycle.md#certificate-editions-20261017)、[通常縦横共有版](../../spec/domains/vaccinations/lifecycle.md#certificate-shared-layout-edition-20261018)、[D08技術案](../../design/domains/vaccinations/technical-spec.md#certificate-edition-proposal)を照合。本節は10-20の照合履歴。10-21追補ではD08正本と技術文書も更新し、RV4種類・再発行日を追加する。
- [001既存4表とcurrent view](../../../backend/migrations/001_init.sql#L8316): series unique(clinic,vaccination,kind)、edition単調増分/previous trigger、artifact unique(clinic,edition,layout)、issue/reprint成功operationのedition/artifact非null。current viewは接種未削除・現在source版一致・同series最新版を判定し、現在認可は別途必要。
- [C06参加契約](../../design/shared/operation-participation.md)、[contract.go](../../../backend/internal/operation/contract.go)、[dto.go](../../../backend/internal/operation/dto.go)を照合。Content/CAS分離、DB-only Apply、string値refsを維持。
- [fingerprint.go](../../../backend/internal/operation/fingerprint.go#L25)の最大64 fields/各256 bytes、[runner.go](../../../backend/internal/operation/runner.go#L117)のContent照合とowner rejection許容status、[dto.go](../../../backend/internal/operation/dto.go)の成功時限定string refsを読取照合。既存PKは[001_init.sql](../../../backend/migrations/001_init.sql#L8358)の `(clinic_id,operation_key)` を維持。
- [vaccine model](../../../backend/internal/model/vaccine.go)にRV識別列なし。名称・価格・speciesから推定しない。

開始時にGitHubの486/487契約PR、remote branch、共有repositoryのclaim/worktreeを照合し、同じ証明書契約の既存作業を確認できなかった。共有YAMLの編集ownerは本作業のみ。EMR-518のpackage仕様・Fable branchは別作業である。

## 件数・表現上の不足（業務変更せず報告）

1. 採用案の「9 paths / 9 operations」は列挙と一致しない。発行POSTと履歴GETは同じ `/vaccinations/{vaccination_id}/certificate-editions` であり、OpenAPIでは **8 unique paths / 9 operations**。9経路というユースケース数は保持し、架空の9番目のpathを作らない。
2. `ArtifactDescriptor` のshapeは採用されているが、9操作のresponse内に配置が指定されていない。standalone schemaとして定義し、editionやC06 ResultResponseへ独断で追加しない。FEはC06 `artifact_id` から既定のcontent routeを構築できる。descriptor返却配置は後続確認事項。
3. rabies専用入力集合は `certificate_number / injection_tag_number / dog_registration_tag_number / body_size / distinguishing_features / injection_amount_text / registration_fee / injection_fee / tag_fee / total` の10欄（CertificateInputsの全欄）と確定する。[D08の既知見本対応表](../../spec/domains/vaccinations/lifecycle.md#certificate-editions-20261017)・[手動料金契約](../../spec/domains/vaccinations/lifecycle.md#L174)・[技術設計の見本対応](../../design/domains/vaccinations/technical-spec.md#L164)による分類であり、新しい必須項目・番号採番・医学的既定・料金計算を追加しない。normalでは各欄の省略又はnullだけ許可し、空文字・空白文字列・明示0を含め1欄でも非nullなら422 validation。inputs={}は許可し、保存時は全10欄を明示nullへ展開する。rabiesでは全欄optional/nullableと型検証を維持し必須化しない。
4. schemaにないproperty拒否、normal/rabiesとlayout/manifest必須キーの組合せ、保存時全inputsキー、現在認可/CASはserver検証も必要。OAS 3.0.3には `if/then` がないため、requestのkind/layoutとsnapshotのkind/manifest必須キーは`anyOf`で表現し、descriptionも併記する。snapshotのall top-level requiredに加え、列挙されたnested fieldsも全キー保持、nullable値だけ欠落を表す。
5. C06文書§5.1はHMAC config配線/rotationを「未選定」と記載している。採用案の「既存C06配線を使用」は将来の再利用方針として保持し、利用可能と認定しない。実装時に489との必要subset照合が必要。実鍵作成・閲覧・保存は本作業で行わない。

これらを理由に業務ルールやrouteを改変しない。9操作の契約mock準備と、未指定点・内部実装の受入を分離する。

## 責任と現在認可

486: 専用HTTP/DTO、元情報resolver、edition DB writer/reader、C06 participant/lookup authorizer、RV属性保存、およびDB writerから487 rendererを呼ぶ準備orchestrator。487: FE、現行v1の既知3見本のversioned server renderer/template/font/assetとPDF preview。RV4種類一括出力は追加差分で、3見本対応だけでは完成しない。共有api.yamlの統合は486一人。

| 操作 | 必須権限 |
| --- | --- |
| context/history/edition/read、保存版preview、reprint/content/lookup | `vaccinations:view`（lookupは保存元対象とintentの現在認可） |
| 発行前編集preview、issue | `vaccinations:view` + `vaccinations:edit` |
| master適格属性編集／候補読取 | 既存 `master-medical:edit`／既存master view |

患者・医院・接種・元カルテ関係を毎要求現在認可で照合。actor/clinicは認証context由来。閲覧だけで発行不可。親カルテfinalizedを追加の一律禁止条件にしない。死亡後の既存記録読取・訂正の既決を維持し、新資格や法定必須入力を推測追加しない。

## DTOの固定

OAS 3.0.3。ID/source versionはinteger/int64、minimum 1（既存wire準拠）。editionは正のint64。operation_keyはlowercase hyphenated UUID、nil UUID拒否。hashはlowercase hex64。C06 `result_refs` / `committed_versions` はstring値。wire int64をJS binary64で丸めないため、FEの既存MoneyInputの表現限界を明示し、BEはint64を検証する。

`CertificateKind=normal|rabies`。`CertificateLayout=portrait|landscape|rabies`。normalはportrait/landscapeだけ、rabiesはrabiesだけ。組合せの意味不一致は422 validation。

### CertificateInputs

全10欄はrabies専用。kind=rabiesでは全property optional/nullableと型検証を維持し、必須化しない。kind=normalでは各欄の省略又はnullだけ許可し、空文字・空白文字列・明示0も非nullとして422 validation。inputs={}は許可する。未送信とnullは新発行時に同じ空欄へcanonicalizeする。保存snapshotでは全キーを明示null含め保持。明示0や番号文字列の先頭0・空白を黙って数字変換しない。

| field | 型・意味 |
| --- | --- |
| certificate_number / injection_tag_number / dog_registration_tag_number | string/null。内部IDと別の印字値 |
| body_size | small / medium / large / null |
| distinguishing_features / injection_amount_text | string/null。注射量を自動既定化・医学換算しない |
| registration_fee / injection_fee / tag_fee / total | integer/int64/null、JPY円。新minimum・非null・合計一致制約なし |

4料金欄は発行前に職員が手動修正できる。totalは独立入力で自動sumしない。新規料金初期値は4欄ともnull、現在会計/masterから転記しない。料金修正後は再preview。normalへこの10欄の非null値は422 validation。形式だけの任意欄はnullで保持。通常の飼主・ペット・接種・医院情報は既存snapshot.source/owner/pet/clinicから取得し、CertificateInputsを基本情報編集APIにしない。

### CertificateSnapshot v1

全top-level key必須。未取得値はnullableのまま、勝手に0/空文字へ補完しない。取得障害はnull欠落と区別し503。

| key | 固定shape |
| --- | --- |
| schema_version / kind | 1 / CertificateKind |
| source | vaccination_id, version, vaccine_id, vaccine_name:string/null, administered_on:date, next_scheduled_on:date/null, lots:array[string/null]固定4件, doctor_name:string/null |
| owner | name,name_kana,postal_code,address,phone: 各string/null |
| pet | name,species,breed,sex,coat_color: 各string/null、birth_date:date/null、age_text:string/null |
| clinic | name,director_name,postal_code,address,phone,fax,email,url,veterinarian_association,designated_veterinarian,veterinarian_name: 各string/null |
| display_issue_date | date。normal独立発行日用。rabies帳票へ独立発行日欄を追加しない |
| inputs | 全キー保持のCertificateStoredInputs |
| eligibility | rabies_eligible:boolean/null, vaccine_id:ID, vaccine_version:int64/null, provenance:master_snapshot/legacy_unresolved |
| render_manifest | renderer_version,font_asset_version,templates:{portrait,landscape,rabies:string}。対象外キー省略可、normal両layout必須 |

`snapshot_sha256`はsnapshot本体外のresponse/metadata。自己参照hashにしない。医院独自欄の未取得値はnullで表示空欄、サンプル固定値を入れない。

CertificateEditionは `id,series_id,vaccination_id,kind,edition,source_version,previous_edition_id:ID/null,issued_at:date-time,issued_by:ID,validity,snapshot,snapshot_sha256,available_layouts:CertificateLayout[]`。validityは保存statusではなく現在reader projectionで `current|superseded|source_changed|source_deleted`。複数理由の優先順はsource_deleted > source_changed > superseded。

ArtifactDescriptorは `artifact_id,edition_id,layout,template_version,renderer_version,content_sha256,content_path`。content_pathは同origin保護API相対path。object key/署名blob公開URLを返さない。返却配置は未指定。

## HTTP 9操作

serverは `http://localhost:8080/api/v1`、以下は相対path。JSON request/responseは `Content-Type: application/json`、PDF contentだけ `application/pdf` binary。単一DTOは直接返し、一般 `{data:...}` envelopeを新設しない。履歴collectionだけ既存pagination envelope。認証は既存Cookie/JWT、POSTは既存 `X-Requested-With: XMLHttpRequest`。

| # | method / path | request → 成功response |
| --- | --- | --- |
| 1 | GET /vaccinations/{vaccination_id}/certificate-context | 200 `{vaccination_id,source_version,eligibility:{state,allowed_kinds,provenance},can_issue,latest:{normal:Edition/null,rabies:Edition/null},initial_inputs}` |
| 2 | POST /vaccinations/{vaccination_id}/certificate-preview | `{kind,layout,inputs,preconditions:{source_version,latest_edition_id:ID/null}}` →200 `{snapshot,snapshot_sha256,preview_fingerprint,source_version,latest_edition_id,pdf_base64,media_type:"application/pdf",template_version,renderer_version}` |
| 3 | POST /vaccinations/{vaccination_id}/certificate-editions | `{operation_key,kind,layout,inputs,preview_fingerprint,preview_display_issue_date:date,preconditions:{source_version,latest_edition_id:ID/null}}` →201 succeeded / 200 replayed、C06 ResultResponseだけ |
| 4 | GET /vaccination-certificate-editions/{edition_id} | 200 CertificateEdition |
| 5 | GET /vaccinations/{vaccination_id}/certificate-editions?kind=normal&page=1&limit=20 | 200 `{data:CertificateEdition[],total,page,limit}`。kind必須、page既定1、limit既定20/1..100（既存境界）、edition降順 |
| 6 | POST /vaccination-certificate-editions/{edition_id}/preview | `{layout}` →200 `{edition_id,layout,pdf_base64,media_type,template_version,renderer_version,content_sha256}` |
| 7 | POST /vaccination-certificate-editions/{edition_id}/reprints | `{operation_key,layout,preconditions:{source_version,latest_edition_id}}` →200 C06 ResultResponse（succeeded/replayed） |
| 8 | GET /vaccination-certificate-artifacts/{artifact_id}/content | 200 application/pdf binary、必須 `Cache-Control: private, no-store` |
| 9 | GET /vaccination-certificate-operations/{operation_key}?intent=issue\|reprint | 200 C06 ResultResponse（成功又はnot-succeeded）、失敗は下表 |

contextのlatestは無効版もvalidity付きで示す。フォームへ既発行値をcopyするのはFE明示操作で、自動更新・自動発行ではない。

発行前previewは、認可済み保存済み原情報と入力から同server rendererで純計算。series/edition/operation receipt/業務監査/接種autosaveを作らず、正本artifact uploadなし。通常access logは業務writeと別。FEは印刷ボタンのない確認previewとして表示する。端末が取得済みPDFをコピーできないとは保証しない。

初回issueの現行v1契約はpreviewのdisplay_issue_dateへ固定してserverが再構成し、preview生成可能なJST日付との整合を確認する。初回は任意過去日不可、日跨ぎはdependency-changedで再previewとし、初回発行日の新たな自由度は採用しない。再発行にはこの制約を一律適用せず、10-21採用の「当初の発行日を初期値として変更可能」を別途満たす必要がある。現行v1 DTOは再発行の表示日選択を表現できず、下記の契約拡張が必要。serverで版固定manifestを解決しhash一致を検査、client supplied snapshotを信用しない。前版があれば同series edition+1/previous IDの訂正版を保存し履歴を上書きしない。normal layout切替だけでissueしない。成功後はC06 refsからedition reader→content readerへ取得。

edition/historyは履歴参照可、旧版に有効content_pathを含めない。保存版previewは現在有効edition限定。既存artifactはdigest検証read、未生成normal別layoutは同snapshot・固定manifestで純描画し、そのpreviewではartifact行/receiptを作らない。欠けた旧snapshotを現在情報で埋めない。

reprintは表示日変更を伴わない同版再印刷で、現在edition限定。表示日変更を伴う再発行は新版作成であり、このreprint DTOでは表現しない。既存artifact再利用、未生成normal別layoutは同snapshot・固定manifestから同editionへ追加し、版増分なし。reprint receiptは出力準備済みを示し紙の印刷成功ではない。初回issue後最初の印刷はissue artifact refsを利用でき、別reprint receiptを自動作成しない。

contentは返却直前も現在grant・医院患者・current view検査。digest不一致/正本欠損は503、現在template再描画で黙って成功させない。旧版は409 dependency-changed。FE印刷はこのPDFだけを使用。既に渡した出力の回収は保証外。

lookupのscopeは認証context＋owner_kind=vaccinations＋actor_kind=staff＋canonical staff ID＋intent=vaccinations.certificate.issue/reprint。client指定scope不可。URLへ別対象IDを渡して認可をすり替えない。保存元対象metadataからresolverが現在認可する。

## C06 Content・再送・失敗契約

Contentは既存C06 NewContentの制限（最大64 fields、field名・値それぞれUTF-8 256 bytes以下）を維持し、次の短いcanonical文字列だけで構成する。snapshot JSON、manifest JSON、料金・氏名等の自由入力文字列をContent fieldへ直接入れない。

Content(issue)のfield名は厳密に `vaccination_id`, `kind`, `layout`, `source_content_version`, `snapshot_sha256`, `render_manifest_sha256`, `preview_fingerprint` の7個とする。vaccination_id/source_content_versionは正のint64のcanonical decimal string（1–19 ASCII bytes）、kindはnormal/rabies（6 bytes）、layoutはportrait/landscape/rabies（最大9 bytes）、3つのhashはlowercase hex64（64 ASCII bytes）。source_content_versionは確認済みsnapshot.source.versionであり、今回の実行CASをコピーする欄ではない。

Content(reprint)のfield名は厳密に `edition_id`, `layout`, `snapshot_sha256`, `render_manifest_sha256` の4個とする。edition_idは正のint64のcanonical decimal string（1–19 ASCII bytes）、layoutとhashはissueと同じ規則。issueだけのpreview_fingerprint/source_content_versionを空文字で追加しない。全field名は256 bytes以内であり、任意fieldを追加しない。

Preconditionsのみ: 現在照合用source_version CAS、latest_edition_id CAS、現在認可fence。確認済みsource版と現在CASを分離する。CASはContentへ混ぜない。内容が変わった場合は旧結果を確定してから再preview・別keyとする。

Fingerprintの唯一の再送正本は上記Contentを既存NewContentへ渡したC06 HMAC binding。snapshot_sha256/render_manifest_sha256/preview_fingerprintは内容commitmentであり、独立した再送receiptでも認証tokenでもない。scopeは認証context由来。owner writeはreq.Content.Fields()の短いcommitmentと、それに一致すると検証済みのimmutable snapshot/manifestだけから導出する。別途captureしたcommand、未検証の現在resolver出力、client supplied snapshotをwriterの入力にしない。秘密鍵/新資格情報をpreview用に作らない。

issue/reprintではまず同scope/keyの既存bindingとimmutable metadataを検索し、保存済み元対象を現在認可する。既存bindingがあるときは、現在sourceを再構成する前に、metadataのcanonical bytesとC06 commitmentの整合、および今回requestの元の確認内容との一致を検査する。issueはURL vaccination_id、kind、layout、未送信をnullへ展開した全inputs、preview_display_issue_date、preview_fingerprintを保存済みsnapshot/Contentと比較する。source_content_versionは保存snapshot.source.version由来であり、今回のCAS値で置換しない。reprintはURL edition_idとlayoutを保存済みtarget/Contentと比較する。CASだけの変更はContent相違と判定しない。

requestが保存済み内容と異なる場合は、元対象の現在認可後に409 operation-key-conflictで停止し、成功receiptや別対象の情報を返さない。保存済みContentを無条件で再利用して異なるrequestを同内容扱いにしない。一致するときだけ保存済みcommitmentから同じC06 Contentを構成し、既存C06 HMAC bindingとの照合を行う。metadata欠損・canonical bytes不正・digest不一致は内容一致と推定せず、現在認可の下で503 result-unavailableとして停止する。

その後、保存済み成功/unknown/busyをC06の結果照合規則で扱う。成功済み同内容ならsource変更・日跨ぎ・renderer現在版変更・現在CAS不一致を理由に過去成功を未実行扱いへ戻さず、歴史的成功receiptをreplayedで返す。新たなPDF取得は別途現在有効性gateを通す。unknown/busyは共通停止/同key照会規則を維持する。確定not-succeededで再試行するときだけ、現在認可・source/CAS・preview内容・日付・最新edition等の実行前提を改めて照合する。旧requestの再送照合のために現在情報から新snapshotを作り、保存済みContentを置換しない。

GET operation lookupの200 not-succeededは既存C06 ResultResponse `{status,operation_key}` とし、error/code envelopeを付けない。非2xx errorだけを既定のflat error shapeにする。成功時のresult_refs/committed_versionsはstring値で維持し、非成功へそれらを出さない。既存PK衝突の422 validationと、同scope/key別Contentの409 operation-key-conflictを混同しない。

非2xxの失敗はflat `{error:string,code:string,status?:C06Status,operation_key?:uuid}`。非成功へresult_refs/committed_versions/edition/artifact内容を出さない。既存401 `{error}`、500 generic errorも許容。共通C06 DTOのomitempty自体は変更せず、選定certificate成功schemaではrefs/versions/recorded_atを必須とする。

| HTTP | code / C06 status | FE/readerの意味 |
| --- | --- | --- |
| 400 | validation | malformed JSON/ID/key/enum |
| 422 | validation | kind非適格/unknown、kind-layout意味不一致、normalの専用欄非null、schema_version!=1、既存証明書operation PK衝突（他scope情報なし） |
| 403 | not-authorized / edit-right-lost | 現在権限なし。未知scope存在非開示。保存元対象不明も403 |
| 404 | scope-not-found | lookupでbinding不在だけ |
| 404 | not_found（C06 statusなし） | 通常resource不在 |
| 409 | version-conflict | 現在source/latest CAS不一致 |
| 409 | dependency-changed | preview snapshot/renderer/date不一致、接種訂正済/削除/旧版出力、旧snapshotで別layout再現不能。再取得・再preview、旧版print不可 |
| 409 | operation-key-conflict | 同key別Content、現在元対象認可後のみ。旧結果非開示 |
| 409 | operation-in-progress | 同key照会、別key禁止 |
| 503 | result-unavailable | sticky unknown/結果読取失敗。自動再create/補償停止 |
| 200 lookup | not-succeeded | bindingあり確定成功なし。source再取得・人の確認後、同key同Content再試行 |
| 503 | render_failed / artifact_unavailable（C06 statusを捏造しない） | 業務TX前描画/正本blob障害等。成功/未実行を推定せず、発行済receiptの可能性はLookup照合 |
| 500 | 既存generic error | SQL/path/stack非公開 |

C06 statusはhyphen形式。旧案のstale_source等へ改名しない。error本文はローカライズ可、機械判定はcode/status。

## 内部準備の必要subset（全て未実装）

FE mock契約の固定は、次の内部DB・claim adapter・renderer実装を待たない。既存4表の存在を専用writer/reader完成と扱わない。

- `vaccination_certificate_operation_targets`: operation_binding_id PK、clinic_id、vaccination_id（正整数CHECK）、edition_id nullable（非null時正整数CHECK。issueはnull、reprintは対象edition ID必須）、operation_kind issue|reprint、canonical_snapshot_bytes bytea NOT NULL、canonical_render_manifest_bytes bytea NOT NULL。従来案の曖昧なcanonical_content_bytesはこの2列へ置換する。canonical_snapshot_bytesは後述schema v1のhash用integer projectionを適用したsnapshotのJCS UTF-8 bytes、canonical_render_manifest_bytesは同snapshot.render_manifestのJCS UTF-8 bytes。wire DTOそのものをstring型へ変更する意味ではない。読み戻しはschema v1を理解するdecoderで行い、integer projectionのdecimal stringを該当integer欄だけ正確にint64へ復元する。印字番号等の本来string欄は変換しない。再canonicalizeが保存bytesと完全一致することを検査する。
- FKはbinding(id,clinic)だけ。vaccination/editionへの業務FKを付けず、claim時に業務rowのKEY SHARE lockを先行取得しない。元ID・医院・関係・存在は後段の正規gate/業務lock取得後に照合する。IDはimmutable lookup anchorであり、存在・権限の証明ではない。全列immutable、医院RLS。共通bindingを新規作成する短いclaim txと同じtxでこのmetadataをINSERTするC06 claim adapterを489へ要求する。並行claimで既存bindingとなった場合は既存metadataを読み、更新・置換しない。成功receiptがないbindingも元対象認可可能とする。issueでは初回の非locking source解決・preview照合後のsnapshotを、reprintでは保存済みeditionのsnapshotを固定する。現在のC06 helperにはこのatomic claim hookがなく、利用開始の実装依存である。

- 既存 `vaccination_certificate_operations` を保持。request_hashはlegacy履歴を消さずNOT NULLだけ緩和、32bytes CHECKは非null旧値に維持。operation_binding_id FK/UNIQUEを追加。新writerはbinding非null/request_hash NULL、再送判定にrequest_hashを使わない。移行後新規INSERTはtriggerでもこの組合せを強制し、既存行の更新でlegacy値を消さない。旧hashからbindingを復元せず、NULL bindingのlegacy履歴として保持。既存(clinic,operation_key) PK衝突は他scope結果を開示せずreject。実既存行調査/移行検証は別工程で、初期DDLだけで履歴なしと仮定しない。
- RV: `vaccines.rabies_certificate_eligible boolean nullable default NULL`、`vaccinations.certificate_eligibility_snapshot jsonb nullable`（vaccine id/version/flag/provenance）。新保存時server copy、過去NULLはunknown。名称/price/species推定・一括falseなし。master編集で過去snapshot更新なし。過去行解決は認可/期待版/監査付き既存訂正adapterへ。trueはnormal/rabies、false/unknownはnormalだけ、unknownのrabies要求は422 validation。


既存PRIMARY KEY(clinic_id,operation_key)は維持するため、C06上で別actor/intent/scopeに属するbindingでも同一医院・同じoperation_keyの証明書operationは共存できない。この既存制約による衝突は、当該request自身の現在認可を通したうえで、Participantから `Rejection{Status: StatusValidation}` を返す確定拒否（HTTP 422、code/status=validation、一般化した説明、result_refs/committed_versionsなし）とする。別scopeのbinding/receipt、対象ID、actor、intent、発行内容は返さない。C06同一scope/keyのContent不一致を表すoperation-key-conflictへ変換しない。Runner.isRejectionStatusはStatusKeyConflictをowner rejectionとして受け付けない。

成功operation INSERTの競合でも同じ扱いとし、対象PK違反だけを識別してvalidationへ変換する。他のunique/FK/DBエラーを一括してvalidationにしない。並行競合でINSERTが失敗したときもowner transactionをrollbackし、edition/artifact/audit/共通success receiptの部分commitを残さない。既存operationへupsert・更新・流用しない。

### 準備とDB-only commitの順序

A. 同scope/keyの既存bindingを先にread。存在すれば再送優先順へ。初回だけ非locking現在認可/source read＋preview再構成確認→C06 claimと元対象metadataを原子的に作成。claimに業務row FK/lockなし。並行claimで既存bindingなら保存Content照合と再送優先順へ戻る。未成功でもresolver可能。

B. PDFを準備する前にも、binding/clinicに対応するimmutable metadataを読み、snapshot/manifestのcanonical bytesと短いContent commitmentの一致を検証する。digestの一致を検査していないmetadataやcaptureしたresolver出力からPDFを生成しない。業務TX外でprivate候補PDFを生成/upload。候補keyはclinic/binding/snapshot-hash/layout/renderer-versionに決定的に対応し同key同bytesへ収束。PDF CreationDate/ModDateはsnapshotの固定display_issue_date由来、実行nowなし。document IDはsnapshot hash＋layout＋manifestから決定的に導出。random UUID/乱数、不定順object列挙、ホスト依存fontなし。renderer versionはPDF library/圧縮設定/font/asset manifestを固定し、同version byte決定性をgoldenで検証する（NOT_RUN）。再試行でbytesが違えば候補上書きせずartifact_unavailableで停止。ここで接種/edition/請求成功なし。upload結果不明は同key HEAD/hash照合、未確認で別blob fallbackなし。

C. Runner.ExecuteのDB-only Participant。共通取得順は必要なowner gate→binding NOWAIT→Apply内のD08既存parent/child→series。targets metadataのanchorは正規順で業務lockを取得した後に実在・医院・関連を検査する。Applyはbinding ID＋clinicでimmutable metadataを再読取し、schema v1 canonical bytesの妥当性、SHA256(canonical_snapshot_bytes)==req.Content.snapshot_sha256、SHA256(canonical_render_manifest_bytes)==req.Content.render_manifest_sha256、保存snapshot.render_manifestのcanonical bytes==canonical_render_manifest_bytesを検査する。issueではsnapshot.source.vaccination_id/version・kindがContent/targetと一致し、layoutがkindに適合し、hash用projection({layout,snapshot})のSHA256がreq.Content.preview_fingerprintと一致することも検査する。reprintではedition_id/layoutがContent/targetと一致し、そのeditionの保存snapshot/manifestとmetadataが一致することを検査する。どれか不一致なら書込を開始しない。

現在認可・source/CAS・preview snapshot・最新editionも正規lock下で再照合する。現在情報は実行可否の検証に用い、確認済みsnapshotの値を黙って更新するために使わない。初回seriesの作成はsource lock下でuniqueを守る。検証済み固定snapshotからedition/必要artifact/証明書operation/必要audit/共通success receiptを同一TXで保存する。uploadはApplyへ入れない。整合性検査が失敗した場合は成功receiptを作らず安全側に停止し、失敗後も保存済み成功の有無を共通照合する。

D. commit後は同key Lookup照合。render/upload失敗へ成功receiptを作らず同key同内容で準備再試行。unknownは共通契約どおり自動再実行停止、旧版復帰なし。未参照候補blobは正本と区別して既存保持方針で後処理。

C06 runnerには現在atomic target claim hookがない。489/446限定subsetが利用開始依存であり、既に接続可能とは記載しない。HMAC鍵は既存C06機構を再利用する方針だがconfig配線の成立は別確認。実鍵・外部tokenをcanonical_snapshot_bytes/canonical_render_manifest_bytesへ保存しない。

## canonical JSON / hash（schema v1）

RFC8785 JCSのobject key順序/string escapeを使用。ただしwire int64をbinary64へ丸めないよう、hash用projectionでは全integerをまずcanonical decimal stringへ変換し、この変換済JSONをJCSでUTF-8化する。snapshot JSONB/公開DTOの整数をstringへ変える意味ではない。fraction/exponent/NaN/Infinityはv1 integerとして拒否、-0は整数0へ正規化。

- キー順はUTF-16 code unit昇順、property入力順は不問。配列は仕様順、lots4件をソートしない。重複JSON keyはparse時拒否。
- inputs未送信は全キー明示nullへ展開後hash化。他のschema外propertyは拒否し、unknownを黙って落とさない。
- Unicode code pointを維持しNFC/NFD変換なし。lone surrogate/不正UTF-8拒否。実日本語をUTF-8化し、quote/backslash/control等必要escapeのみJCS適用。slash不要escapeなし。末尾newline/BOM/空白整形なし。
- dateはYYYY-MM-DD、date-timeはUTC RFC3339正規形。固定snapshotに実行nowなし。整数projectionは符号付き10進・先頭0/plus/指数なし。印字文字列"0001"は変換対象外。
- snapshot_sha256はhash用projection(snapshot)のJCS UTF-8 bytes、render_manifest_sha256はhash用projection(snapshot.render_manifest)のJCS UTF-8 bytes、preview_fingerprintはhash用projection({layout,snapshot})のJCS UTF-8 bytesをそれぞれSHA-256したlowercase hex64。snapshotとmanifestの同じcanonical bytesをimmutable metadataへ保存する。C06 HMACはこのhash値を含む短いContent fieldのみを既存NewContent規則でfingerprintする。full snapshot/manifest JSONをNewContentの単一fieldへ入れない。hash用途を混同しない。

- schema_version!=1はwriter/previewで422 validation。保存未知版はhistory readerも409 dependency-changed、preview/content/reprintも同409で停止。内部snapshotをerrorへ漏らさない。metadataだけの救済readerは範囲外。推測描画/旧版変換なし。

期待mini-vectors（仕様上の期待bytes。試験結果ではない）:

| 入力 | hash projectionのJCS bytes |
| --- | --- |
| `{"b":2,"a":null}` | `{"a":null,"b":"2"}` |
| `{"amount":0,"number":"0001","name":"狂犬病"}` | `{"amount":"0","name":"狂犬病","number":"0001"}` |
| `{"amount":9007199254740993}` | `{"amount":"9007199254740993"}`。9007199254740992への丸め不可 |
| inputs.registration_fee省略 / 明示null | 展開後は同bytes。明示0は`"0"`で別bytes |
| NFC `é` / NFD `e`＋combining acute | 別bytes/hash。正規化しない |

totalと各料金欄の合計が違っても値を変更しない。full snapshotの合成fixture例を末尾に掲載する。renderer/PDF goldenの実行証拠ではない。

## 合成request / response例

実患者・実tokenではない。hash例はshape用の64桁hex placeholderで、実PDFやhash照合合格の証拠ではない。

```json
{"kind":"rabies","layout":"rabies","inputs":{"certificate_number":"0001","registration_fee":null,"injection_fee":3000,"tag_fee":0,"total":3000},"preconditions":{"source_version":7,"latest_edition_id":null}}
```

issue requestは同kind/layout/inputs/preconditionsに次を追加する:

```json
{"operation_key":"11111111-1111-4111-8111-111111111111","preview_fingerprint":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","preview_display_issue_date":"2026-10-20"}
```

issue成功（直接C06 ResultResponse）:

```json
{"status":"succeeded","operation_key":"11111111-1111-4111-8111-111111111111","result_refs":{"vaccination_id":"101","series_id":"201","edition_id":"301","artifact_id":"401"},"committed_versions":{"edition":"1","source_version":"7"},"recorded_at":"2026-10-20T14:30:00Z"}
```

次のreprint例はnormal editionだけに成立。rabies版なら422 validation:

```json
{"operation_key":"22222222-2222-4222-8222-222222222222","layout":"landscape","preconditions":{"source_version":7,"latest_edition_id":301}}
```

拒否例:

```json
{"error":"The certificate source has changed. Refresh the preview.","code":"dependency-changed","status":"dependency-changed","operation_key":"11111111-1111-4111-8111-111111111111"}
```

## FE487のmock PR出口と実装残件

rootレビュー後、公開PR headを現行v1の限定契約としてmock明示adapterで全9操作shapeを固定できる。RV4種類一括出力・再発行日の追加要件完成とはしない。既知3見本、null料金/明示0/合計不一致を無断修正しない、料金変更後再preview、issue成功→edition読取→PDF取得、同key replay、busy/unknown、失権、接種訂正後旧版拒否、normal縦横同edition、render失敗を架空fixtureにする。

FE専有PRは画面/adapter/mock/契約fixtureまで完結可。backendなしを成功へfallbackせずmock/実接続を明示分離。生成models.ts手編集なし。server rendererは487所有の別componentとして契約準備可能だが、FE-only PRをBE完成・実印刷合格とは扱わない。内部DB/claim adapter/元対象resolver/reader/writer/認可gate/renderer/golden/実印刷/実API接続は未実装又は未検証。

会計/masterへ逆同期・自動転記・自動合計なし。preview/issue/reprintからpackage適用/明細生成/再請求/接種createを起動しない。医院別packageの設計はEMR-518の独立工程で、初期料金転記元/fee role mappingは本mockの必須依存にしない。

## 作業・検証境界

10-20の変更対象は本Markdownとbackend/docs/api.yamlのみ（当時の作業履歴）。10-21追補は指定の6 Markdownのみを変更し、api.yamlや実装コードを変更しない。D08技術文書は10-21追補の対象に含める。コード/runtime/DDL/model/試験ソース/Docker/DB適用/実鍵は引き続き対象外。runner・製品試験・描画・実印刷・DB適用は **NOT_RUN**。通常hookやpush後の自動CIが発生しても、未実装の実証と混同しない。Plane更新とmergeはroot担当。

10-20当時の静的確認結果（今回の再実行ではない）: 既存Ruby YAMLで解析成功、新規duplicate key 0、全local refs 3,017件を解決、既存347 pathsと既存components全件保持、新規8 paths/9 operations/36 schemas、全527 operationIdの重複なし。C06 exact field集合と64 fields/各256 bytes制限、lookup not-succeededのstatus/operation_keyだけのshapeを静的照合。PreviewRequest/IssueRequestと保存snapshotのnormal branchはCertificateNormalInputsの10欄null-only制約を適用し、保存時はCertificateStoredInputsの全キー必須制約も維持。rabiesの共通型を変更していないことを差分確認。Markdownのlocal参照先とJSON例6件の構文、指定2ファイルだけのdiffとgit diff --checkも確認済み。これらはruntime検証ではない。API文書の公開と内部利用開始条件、FE mock PR完成、製品受入を分けて報告する。

### full snapshot v1の合成fixture（未実行）

全nullキーと4 LOT、fixed manifestを展開したmock authoring用。実医院/患者/renderer版を示さない。display_issue_dateは合成preview日付で、rabies帳票へ独立欄を増やさない。

```json
{
  "schema_version": 1,
  "kind": "rabies",
  "source": {"vaccination_id":101,"version":7,"vaccine_id":11,"vaccine_name":"合成ワクチン","administered_on":"2026-10-20","next_scheduled_on":null,"lots":[null,null,null,null],"doctor_name":null},
  "owner": {"name":null,"name_kana":null,"postal_code":null,"address":null,"phone":null},
  "pet": {"name":null,"species":null,"breed":null,"sex":null,"coat_color":null,"birth_date":null,"age_text":null},
  "clinic": {"name":null,"director_name":null,"postal_code":null,"address":null,"phone":null,"fax":null,"email":null,"url":null,"veterinarian_association":null,"designated_veterinarian":null,"veterinarian_name":null},
  "display_issue_date": "2026-10-20",
  "inputs": {"certificate_number":"0001","injection_tag_number":null,"dog_registration_tag_number":null,"body_size":null,"distinguishing_features":null,"injection_amount_text":null,"registration_fee":null,"injection_fee":3000,"tag_fee":0,"total":3000},
  "eligibility": {"rabies_eligible":true,"vaccine_id":11,"vaccine_version":3,"provenance":"master_snapshot"},
  "render_manifest": {"renderer_version":"mock-renderer-v1","font_asset_version":"mock-font-v1","templates":{"rabies":"mock-rabies-v1"}}
}
```

## 2026-10-21 読取・描画部分の実装（セッション5）

- 実装スコープは読取3操作のみ（`getVaccinationCertificateEdition` / `listVaccinationCertificateEditions` / `previewSavedVaccinationCertificate`）。context・issue preview・issue・reprint・content route(#8)・operation lookup・writer・artifact upload/insert・C06・FE は引き続き対象外。
- artifact 読取アダプタは fail-safe（nil store・open/read失敗・サイズ超過・sha不一致・非PDF先頭 → `artifact_unavailable`）。production の store は未結合（nil）のため、既存 artifact 行があっても writer セッションが storage を結合するまでは 503 `artifact_unavailable` を返す。
- レンダラーは stdlib の最小 PDF 1.7 実装（固定オブジェクト順・xref・`/ID`・日付は snapshot 由来で実行時刻不使用）。日本語フォントの非埋め込み Adobe-Japan1 `HeiseiKakuGo-W5`（UniJIS-UTF16-H）は契約の固定 font/asset 規則を満たさないため、production の純描画は `certificateRendererProductionEnabled=false` で無効化し 503 `render_failed` で停止する（適合する埋め込みフォント asset 採用まで）。保存版 preview は production で現状 503 を返す（artifact store 未結合 + 純描画停止）。
- バージョン定数: renderer `ekarte-certificate-pdf/1`、font asset `adobe-japan1-heisei-kakugo-w5-unembedded/1`、templates `normal-portrait/1` / `normal-landscape/1` / `rabies-a4-landscape/1`。manifest 不一致は `dependency-changed` で失敗する。
- 用紙（2026-10-21 依頼者確定、PR #772 comment 6305181050）: normal portrait = A4 縦 210x297mm、normal landscape = A4 横 297x210mm、既知のrabies見本 = A4 横 297x210mm。未知のRV4種類全体への適用ではない。
- 検証: compile/test/render/print はいずれも NOT_RUN（依頼による実行禁止）。

<a id="certificate-mtg-contract-gap"></a>
## 2026-10-21採択追加要件と現行v1の境界

採用判断者MinoruSoga。[RV4種類一括出力と再発行日の業務正本・9AC](../../spec/domains/vaccinations/lifecycle.md#fr-vaccinations-rv-batch-20261021)へ接続する。現行v1のnormal/rabies、portrait/landscape/rabies、9操作とschema v1を今回書き換えたものではない。「通常縦横＋単一RV＝3様式で全要件完成」という前提は撤回し、v1は既知見本の限定契約とする。10-20の照合結果・実装予定と末尾の読取3操作の部分実装を区別し、追加要件を実装済みとしない。

| 採択済み要件 | v1に未収録の必要差分／未決 |
|:---|:---|
| RV4種類のmaster | kind=rabies／layout=rabiesだけでは4種類を識別できない。帳票種別identity・帳票セット構成版とtemplate／snapshot／artifact対応が必要。4名称・見本はクライアントから後日共有予定、種類名・様式・紙寸法は未確定。既知A4横を全種類に適用しない |
| 一括出力wire／batch | 単一issue／reprintと単一edition/artifact refsだけでは4種類の一括結果を表現できない。構成の欠損／重複検出、種類ごとの成功／失敗／unknown・版とartifact対応、同一操作の照合・再開が必要。batch receipt・commit単位・route／DTOは未選定。4種類を4ページや特定の結合PDF方式と同一視しない |
| 再発行の表示発行日 | 当初の発行日を追跡して初期値にし、変更した日を確認した新版snapshot／preview／fingerprintへ固定する入力が必要。現行preview_display_issue_dateはv1初回previewとの照合値であり、任意の再発行日選択fieldではない。現行reprintは固定版のみ。初回の任意過去日不可・日跨ぎ再previewを再発行へ一律適用する旧読替えは採用しない |
| 時刻／版／PDF storage | 表示日と実issued_at／recorded_atを分離し、後者はサーバー操作時刻。表示日変更は新版を保存し旧版・artifactを保持。同版再印刷は版増分なし。種類・構成版・edition・固定描画版・digestとstorageを結合する必要がある。既存store未結合の部分実装で充足とはせず、DDL／storage方式・wire拡張は未選定 |

再発行日の詳細validationは後続設計とし、初回発行日の未回答の自由度、接種日変更、未知様式の独立発行日欄を追加しない。通常縦横は同版の表示方式を維持する。診療項目packageと帳票セットを分離し、package applyから自動発行しない。現在認可・患者対応・有効性gate・旧訂正版保護と結果不明時の照合を維持し、一部失敗を全成功にせず成功分を無条件再発行しない。

受入は上記正本の欠損・重複・一部失敗・再発行日・履歴不変の9ACへ対応付ける。試験はNOT_RUN、静的差分・参照確認のみ。api.yaml・実装・DB・外部OCR／実患者PDF処理・実印刷は今回の対象外。Lステップ／DMはこのPRへ含めない。
