# EMR-503 / EMR-504：採択技術契約と Phase 境界

更新日：2026-10-21。Revision 4（root 採択済み）。R3 採択契約を保持し、PR #784 の9指摘に対する採択済み技術補正を公開用に固定する。設計採択であり、OpenAPI 反映・schema 適用・runtime 実装・試験合格・運用開始を意味しない。対象は EMR-503 現 Phase A/B、EMR-504 今回の技術整備A/B。本人の最新指示「Lステップ／DM連携側の機能はPhase1」に従い、LSTEP連携・DMをPhase1対象外とした旧記述を撤回する。504 A/BはPhase1内の今回の技術整備slice、Cは後続の技術研究としてA/Bの完了判定と分け、後続という理由でPhase1外とは扱わない。LINE予約・本人紐付け・再診問診のunlink/対象認可の必要受け手も維持する。通知再開の時刻境界は本人回答「ok」に基づき予約開始時刻 > authoritative server現在時刻へ確定したが、本書は今すぐの送信有効化・実送信を許可しない。外部送信・資格設定・鍵生成・DB 適用・API 実装・実装PRは本作業で行わず、この文書はそれらの実行許可を含まない。今回の公開は本書1文書だけのdocs-only PR。試験は NOT_RUN。

## 1. 結論と調査証拠

調査時点の「採用 wire 未固定」という判定は両票とも妥当であり、本書で現 Phase の技術契約を固定する。ただし業務・保存/配送原則まで未決ではない。次を再設計せず再利用する。

- main根拠：`455ec2f4d84c5e08da568bf99fc6d825506b499d`。取得した現api.yaml blobは`9459cb23c968975417f4a44ea6f8d31ae701176f`。
- [EMR-503](https://app.plane.so/baritechllc/browse/EMR-503/)本文、全コメント1件（43382798-8a48-4513-966f-f5acc15acb7f、最終更新2026-10-20 12:45:17 UTC）を確認。
- [EMR-504](https://app.plane.so/baritechllc/browse/EMR-504/)本文、全コメント2件（ba4a2064-d8dc-4461-bc0f-81084dd772ba、c5b99cc5-2b28-424e-af74-2ee7abdf951e、最終更新12:45:50 UTC）を確認。listは各1ページでnext_page_results=false。
- [503選定済み保存/配送契約](https://github.com/MinoruSoga/AnimalEkarte/blob/455ec2f4d84c5e08da568bf99fc6d825506b499d/docs/design/domains/support/technical-spec.md#support-selected-delivery-contract)：同TXの院内報告＋確認snapshot＋intent＋receipt、宛先別claim、成功/unknown再送禁止、verified absent後の管理再試行、原本保持。
- [503業務FR/AC](https://github.com/MinoruSoga/AnimalEkarte/blob/455ec2f4d84c5e08da568bf99fc6d825506b499d/docs/spec/domains/support/README.md)：全院共有READ、mutationは報告医院、個人history別scope、スクショ個人情報包含可・強制マスクなし。
- [504業務FR/AC](https://github.com/MinoruSoga/AnimalEkarte/blob/455ec2f4d84c5e08da568bf99fc6d825506b499d/docs/spec/line/lstep-integration.md#requirement-fr-crm-line-unlink-future)：予約/参照保持、unlink停止、relink自動再開なし、選択再開は即時送信でない。followup・DM・unknownの各節も再利用。
- [504の旧Phase記述（固定SHAの調査履歴）](https://github.com/MinoruSoga/AnimalEkarte/blob/455ec2f4d84c5e08da568bf99fc6d825506b499d/docs/spec/line/architecture.md#future-phase-followup-adapter-20261019)：当時の記述は調査履歴として保持する。現在のPhase範囲は最新本人指示を優先し、LSTEP／DMをPhase1対象外とする根拠に使わない。IR-07/IR-D01/IR-D02・全17triggerを含む受領/実装確認は別に残る。
- [現API support](https://github.com/MinoruSoga/AnimalEkarte/blob/455ec2f4d84c5e08da568bf99fc6d825506b499d/backend/docs/api.yaml#L31996)：旧multipart create、best-effort自動Plane、旧plane_issue列、手動plane-ticketのみ。確認snapshot/operation/CAS/宛先別結果の新DTO・lookupはない。SupportBugReportはL13358。
- [現API LINE](https://github.com/MinoruSoga/AnimalEkarte/blob/455ec2f4d84c5e08da568bf99fc6d825506b499d/backend/docs/api.yaml#L15452)：旧link/unlink/confirm、L15801のsend/send-logs、L28825のlink-ownerのみ。選択再開・確認・dispatch結果照会wireはない。
- [実owner route](https://github.com/MinoruSoga/AnimalEkarte/blob/455ec2f4d84c5e08da568bf99fc6d825506b499d/backend/internal/owner/http_routes.go)：canonical `/owners/:id/*`。装飾的clinic-prefix PATCH aliasesは除去済み。本契約に古いFE/docsのclinic-prefixをコピーしない。
- [永続model](https://github.com/MinoruSoga/AnimalEkarte/blob/455ec2f4d84c5e08da568bf99fc6d825506b499d/backend/internal/model/storage_messaging.go)は`json:"-"`でありpublic DTOではない。[C06 contract](https://github.com/MinoruSoga/AnimalEkarte/blob/455ec2f4d84c5e08da568bf99fc6d825506b499d/backend/internal/operation/contract.go)はDB-only、外部効果は別outbox/照合責任と明記。

## 2. 共通 wire 契約

- prefix `/api/v1`。新DTOのbigint ID/revisionはcanonical正のsigned-64 decimal string、operation keyは既採用UUID。既存legacy DTOをこの文書だけで一括string化しない。
- write bodyはclosed object。認証actor/clinic、confirmed_by/at、外部結果・配送成功、doctor指定の事実をclient申告で作らない。確認する対象ID/内容/目的はContent、実行CAS/fenceはPreconditionsとして区別する。
- 大きい本文/配列はschema名付きcanonical JSONのSHA-256 commitmentをC06 Contentへ束縛する。typed immutable commandを保存しApply直前に同じprojectionから再計算する。別のunchecked payloadでwriteしない。確認対象の内容版はContent、単なる現在CASはPreconditions。
- 新規操作は非変更の現在認可/scope preflight後にclaim。claimと元targetの認可anchorを同じclaim TXで保存し、失敗bindingにもlookup可能な元対象を残す。既存keyでは保存済み元anchorから現在のlookup権限を検証してから内容競合/receipt開示。要求bodyのtargetへ付け替えない。
- 成功receiptは内部保存成功のみ。provider状態を成功receiptの一語で置換しない。GET/replayはprovider HTTPを起動しない。lookup/responseは明示DTOへ写像し、C06 EffectSummaryを丸ごと公開しない。
- C06到達後は既存Go `operation.ResultResponse` / OpenAPI `MedicalRecordPlanOperationResultResponse` の形を再利用する（package固有result_refs型は流用しない）。closed status/HTTP表は第7節に固定。新writeも200 succeeded、replay/lookupは200、busyは202。初案の201/処理中409/operation wrapperは撤回。

## 3. EMR-503 現 Phase A/B：操作・DTO 契約

既存一覧/status/chatは別機能として保持する。DELETE の route と報告医院 scope は保持するが、旧 handler の uploader.Delete は保持対象ではない。通常削除を論理削除だけに置換し、確認済み報告の添付原本・snapshot・intent・attempt/result・receipt と再送抑止 identity を残すことを A の利用開始条件とする。論理削除と必須 audit_logs（action=support_bug_report.delete、既存report resource/root、trusted actor/clinic/report IDと前後削除状態）を同じsupport-owner TXへ保存し、監査失敗では削除もrollbackする。現 handler の service.Delete commit後best-effort logAuditはこの必須監査へ流用しない。既存DELETEを本補正だけで新C06 operation APIへ移行せず、新route/DTO/roleや架空bindingを追加しない。実在するoperation参照がある場合だけ監査へ保持する。旧 create-error cleanup も commit 成否不明時や既存原本を物理削除しない。未参照 staging の扱いと別承認の保存期間/削除手続を混同しない。新規配送に旧auto-sendを残して迂回経路にしない。旧plane-ticketの「外部IDなしなら再送」は廃止対象。未移行旧clientには明確な契約不一致を返し、未確認送信を互換動作として続けない。切替順は確認対応client→旧送信経路遮断→新intent worker有効化で、A段階ではworker未起動。

| 操作 | request | response / 責務 |
|---|---|---|
| POST `/support/bug-reports/confirmation-preview` | 既存multipartのtitle/detail/page_url/route_path/user_agent/viewport/app_version/screenshot＋selected_destinations（provider＋設定上のdestination ref） | 200 `confirmation_payload, payload_digest, screenshot_digest`。純preview、院内報告/intent/外部送信なし |
| POST `/support/bug-reports/confirmed` | 同じmultipart＋operation_key＋confirmation_payload＋acknowledged_payload_digest | 200 C06 ResultResponse（refs: report_id/intent_id、version: report）。同じ内容を再検証し院内保存だけを成功として返す |
| GET `/support/operations/{operation_key}?intent=support.report_confirmed_create` | bodyなし | 同actor/clinic/C06元anchor認可後の保存結果。report未作成時のanchorは認証clinic＋作成actor＋固定operation family。別のshared reportへの変更権限にしない |
| GET `/support/bug-reports/{report_id}/delivery-intents` | bodyなし | 認可された報告のintentと宛先別result projection。既存の共有READを維持し、secret/provider資格/evidence原文は公開しない |
| POST `/support/bug-reports/{report_id}/delivery-intents/confirmation-preview` | `expected_report_version, selected_destinations` | 既存報告の現内容をserverで読んだconfirmation_payloadとdigest。既存報告の新しい宛先/内容は再確認、過去intentは不変 |
| POST `/support/bug-reports/{report_id}/delivery-intents` | `operation_key, expected_report_version, confirmation_payload, acknowledged_payload_digest` | 報告医院のみ。新確認intentとreceiptを原子保存。元intent成功/unknownの単純再送として新intentを使うことは禁止 |
| GET `/support/bug-reports/{report_id}/operations/{operation_key}?intent=support.delivery_confirm` | bodyなし | 保存された元report anchorで現在認可。URLとの不一致404 |

署名confirmation_tokenは既決の必須条件ではなく初案で追加した手段のため撤回する。新署名鍵・key rotation・expiryは導入しない。previewは純readのcanonical confirmation_payloadとSHA-256 digestを返す。FEは表示したpayloadをimmutableに保持し、明示確認時にその全文とacknowledged_payload_digest、同一添付bytesをsubmitする。serverは現在の許可された宛先設定と提出本文/添付bytesから同じprojectionを再構成し、全文・digest・acknowledgementを照合して保存する。client digest単独を認可や確認の証拠にしない。内容/宛先/添付/確認したreport版の変化は再preview・再確認。clientの任意URL/webhookを宛先にしない。

confirmation_payloadのclosed keysは `schema, report_ref, report_version, content, screenshot, destinations`。schema=`support-confirmation-v1`、新規reportはreport_ref/report_version=null。contentは `title,detail,page_url,route_path,user_agent,viewport,app_version` の7string。screenshotはnull又は `{sha256,media_type}`（bytesは別multipart）。destinationsは `{provider,destination_ref,configuration_digest}` の集合をprovider/destination_refの文字列順へ整列し重複を拒否。providerはdiscord/planeのみ、configuration_digestは資格を含めない送信先設定のcanonical digest。payload_digestはUTF-8 JCS SHA-256 lowercase64hex。credentials/URL署名値/表示期限/actor時刻はprojectionに含めず、clinic/actorは認証C06 scopeへ別途拘束する。serverの既存設定から正しい集合を再構成できない場合は停止。署名なしでも人が実際に画面を見たことの暗号学的保証はしないが、既決の確認UI・明示submit・確認内容の保存と変更時再確認を満たす。

旧request.goのtrim/入力検証をpreviewとsubmitで共通化し、確認画面には実際に保存/送信される正規化後の内容を表示する。既存8MiBとPNG/JPEG/WebPを維持し、文字数とbyte数を混同しない。ハッシュはHTTP multipart境界や一時URLでなく保存する本文/添付bytesの内容へ結び付ける。decoded NUL等保存不能textはclaim前に拒否。確認後の無断screening書換えは禁止、変更なら再確認。

添付object uploadはDB TXと原子的でない。previewでは永続uploadしない。submitでは検証済み同じbytesをstaging扱いで保持し、object 成功後に report・原本参照・確認 snapshot・intent・必須 audit_logs・receipt を同じ support owner DB TX へ保存する。監査欠落/書込失敗は report を含む全業務 write と成功 receipt を rollback する。旧 handler の保存後 best-effort logAudit をこの必須監査の実装に流用しない。DB失敗時に外部送信0、object孤立を報告成功にしない。保存済み原本を失敗cleanupで物理削除しない。一時物cleanupの具体方式は既存storage owner契約へ接続する。

宛先 projection：`destination_id, provider, destination_label, latest_attempt_id:null|string, latest_result_id:null|string, latest_result_revision:null|string, result_state:null|succeeded|definite_failure|unknown|verified_absent, external_id:null|string, external_url:null|string, dispatch_enabled:boolean`。attemptなしは「未試行」、attemptありresultなしは「未確定」として区別し、nullを未送信確定へ変換しない。設定なしとworker未有効を外部成功にしない。院内open/resolvedは独立。

### 異operation key間の同等intent抑止（採択済み・未実装）

第二request keyは追加しない。semantic identityは `(clinic_id, report_id, report_version, content_digest, destination_set_digest)` とする。別reportの同文を同じ事象へ自動統合しない。既存の confirmation 全体digestは `payload_digest` として保持し、以下の2digestで置換しない。

- `content_digest` は UTF-8 JCS の `{schema:"support-delivery-content-v1",content:{title,detail,page_url,route_path,user_agent,viewport,app_version},screenshot}` の SHA-256 raw32bytes。contentは確認済みの7string、screenshotはnull又は `{sha256,media_type}` のclosed shapeとし、添付bytesへの既存commitmentを維持する。
- `destination_set_digest` は UTF-8 JCS の `{schema:"support-delivery-destinations-v1",destinations:[{provider,destination_ref,configuration_digest},...]}` の SHA-256 raw32bytes。各要素はこの3fieldのみ、provider/destination_ref順へ整列し同組の重複を拒否する。既存のprovider/ref/configuration digestの意味・検証を再利用し、任意宛先や資格を含めない。
- 新規confirmationのreport_ref/report_version=nullは不変のまま保持し、生成後のreport ID/versionと上記2projectionを用いて既存report確認と同じsemantic identityを構成する。確認payload全体のdigestを直接同等性比較してnull refsと生成済refsの差を別intentへ逃がさない。

採択するschema差分は `support_delivery_intents.content_digest bytea` / `destination_set_digest bytea`。両方NULLを許すのは既存legacy行だけで、pairのCHECKは両NULL又は両非NULLかつ各 `octet_length=32`、新INSERTは両非NULL必須guardとする。UNIQUE `(clinic_id,report_id,report_version,content_digest,destination_set_digest)` を追加する設計であり、現DDLに存在するとは記載しない。既存の `UNIQUE(clinic_id,operation_id)`、RLS、immutable制約は維持する。共有446/466 ownerによる後続schema統合と、保存modelの `ConfirmedDestinations` 欠落・2digest field追従は必要な実装subsetであり、本PRでは実DDL/modelを変更しない。

短いmetadata claim終了後、業務TXのbinding/receipt処理→現在認可→report `FOR UPDATE`→現版・内容・確認・設定再検査→同reportの全intentのsemantic比較→同等なしの場合だけintent/destination/必須audit/receiptを保存する。同key成功replayは通常どおり優先する。全confirm writerは同じ親report lockへ参加し、workerのdestination lockをこの経路へ追加せず逆順を作らない。異keyで同等intentが既にあれば状態がsucceeded/unknown/未試行/失敗のいずれでも `dependency-changed`409・業務zero-writeとし、現在認可された既存delivery-intents GETへ戻す。別keyの成功receiptを作って新配送chainへ迂回せず、許可されたretryは元intentだけを使う。claim metadataが残ることと業務write成立は区別する。

既存両NULL行も同親lock下で保存 `confirmed_payload` / `confirmed_destinations` から上記closed projectionを厳密復元して全intentと比較する。復元不能・不整合は500 invariantで停止し送信0、同等なしと推測しない。移行前に既存重複を検査し、重複の自動統合・削除は行わず、解消/厳密比較が成立しなければ移行・有効化を停止する。runtimeでimmutable制約を解除したりlegacy行をbackfillしたりしない。これらは後続実装・移行条件の設計記述であり、今回のDB実行許可ではない。

### 宛先claim・照合・再試行

初回claim、結果記録、provider照合、管理再試行はworker/service境界であり、FEにraw claim/result登録APIを公開しない。管理UIが必要なら下記を別 operation として次の route を採用する。

- POST `/support/delivery-destinations/{destination_id}/reconciliations`：`operation_key, expected_attempt_id, expected_result_revision:null|string`。原report医院の既存mutation権限、現在元target照合。問い合わせだけで新規 send なし。expected_result_revision は required nullable とし、null は当該 attempt に result 行が存在しないことへの CAS、string は canonical 正 signed64 の最新結果版への CAS とする。欠落/null/0を同一視しない。destination lock 下で最新 attempt ID と結果の不存在又は最新版を照合し、競合は dependency-changed409・zero-write。provider 証拠を受け取る読取 HTTP は C06 Apply/TX の外に置き、server が同じ attempt に束縛した証拠で再確認し、result revision・audit_logs・receipt を同 TX に追記する。client が provider 証拠や成功を自己申告する DTO は追加しない。
- POST `/support/delivery-destinations/{destination_id}/retries`：`operation_key, expected_attempt_id, authorized_retry_result_id`。その宛先/attempt の最新 `verified_absent` を再照合し、同じ intent の同じ destination に属する全 attempt/result 履歴に succeeded が一度もないことを destination lock 下で検査する。歴史上 succeeded は配送済みとして terminal とし、外部側の後日削除/検索不在で verified_absent へ降格させたり、新しい retry を認可したりしない。照合 result writer と全 attempt writer も同じ destination lock・terminal 検査へ参加する。既存 support_attempt_guard は最新 result のみを見て definite_failure も許すため適合済みとはせず、446/466 の必要制約補正と503全 writer で上記条件を満たすまで B を有効化しない。現在の確認内容/宛先/認可が有効な場合だけ同intentの次attemptを作る。前attemptとauthorized result FKを保持。definite_failureだけやID欠落で不在証明を代用しない。
- 両操作の結果照会は同元reportのoperations routeへそれぞれの固定intentで接続。worker claim成功後は期限切れだけで二重claimしない。provider応答断・成功後DB失敗はblocking unknownで照合待ち。
- 初回/再試行の各外部送信直前に、現在 report の identity/version と本文・添付の content commitment を intent の report_id/report_version および保存済み確認 snapshot と再照合する。現在の本文・添付を同じ schema/正規化規則で再計算し、版又は commitment 不一致・確認不能なら外部送信0で再preview・再確認待ちとし、古い intent/snapshot を上書きしない。新規作成時の confirmation_payload の report_ref/report_version=null は確認時の原値として保持し、生成後の report identity/version は intent の保存列で照合する。照合した immutable 本文・添付だけを外部 call へ渡し、検査後の別 payload 再読に差し替えない。加えて、保存した destination_ref と非secret configuration_digest を現在設定から再計算して照合する。Aで保存しBで送る滞留 intent も例外にしない。不一致/設定不足なら外部送信0で再確認待ちとし、保存済み intent の宛先を上書きしない。照合済み同一 immutable 設定値をその呼出しへ渡し、検査後に別の現在設定を読み直して宛先をすり替えない。credential は digest/ログへ含めず、同じ宛先 identity に対応する設定として取得する。送信と設定変更がDB TXで原子的になる、又は開始済み外部効果を回収できるとは主張しない。
- provider側が「不存在」を証明できない場合はunknownのまま。新しいroot権限・管理者限定・証拠なし手動上書きを発明しない。署名済みURL失効や外部コピー消去は保証しない。

## 4. EMR-504：Phase1内の技術整備A/B と後続技術研究C

[旧Phase記述（固定SHAの履歴）](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/spec/line/lstep-integration.md#crm-delivery-phase-boundary-20261017)のLSTEP／DM対象外という扱いは、本人の最新指示「Lステップ／DM連携側の機能はPhase1」により撤回する。本節A/BはPhase1内の技術整備sliceとして整理し、A互換adapter・B契約整備をPhase1外へ除外しない。Cは後続技術研究でありA/Bの完了判定と分離するが、機能のPhase1範囲を狭める意味ではない。LINE予約・本人紐付け・再診問診の必要なunlink/対象認可も維持する。scope正本の修正は別担当の作業であり、本書だけを更新する。通知再開は本人の追加回答「ok」により開始時刻が現在より後の予約を対象とし、開始済みを除外する。任意horizonや別用途の時刻/状態述語で補完しない。送信有効化・実送信は本作業では行わない。

### A：既存互換adapterは新wire不要

LstepCsvImport.ErrorLog、snapshot Tags/ScenariosのNULL/空/JSON変換だけ。api.yaml既存CSV/analytics契約と対応し、dispatch設計待ちでこのsubsetを止めない。新DTOを発明せず元のNULL意味を保つ。

### B：binding writer再利用と明示選択再開

469のW07–11だけがlink/unlink/relinkを書き、491は共通本人情報、504は通知eligibilityの受け手。504専用の第二binding writerは作らない。既存PATCH `/owners/{id}/line-user-id` / alias `/line`、confirm、LINE customer link-ownerの公開wire更新は469/共有API ownerへ一箇所で集約。無条件旧writeの残存はCAS/停止を迂回するため、同じ469 service/gateへ接続する。

| 操作 | DTO / 結果 |
|---|---|
| GET `/owners/{owner_id}/line-notification-eligibility?cursor=...` | `binding_revision, binding_state, reservations:[{appointment_id, appointment_version, eligibility_state, eligibility_operation_id, can_resume, blocking_reasons}], next_cursor`。候補欠落を全件なしにせずページング。新期間の業務制限なし |
| POST `/owners/{owner_id}/line-notification-resumptions` | `operation_key, expected_binding_revision, selections:[{appointment_id, expected_appointment_version, expected_eligibility_operation_id:null\|string}]`。選択集合をContentに束縛。選択対象のみ同TXでconfirmedとreceiptを保存、通知即時送信なし。競合が1件でも全選択 zero-write |
| GET `/owners/{owner_id}/line-operations/{operation_key}?intent=line.notification_resume` | C06元owner anchor・現在owners:viewと医院/各予約閲覧scopeを確認。内部再開結果だけを返す |

eligibility GET の query allowlist は optional `cursor` のみ（省略=先頭、空文字/重複/未知 query は400）。page size は100件という技術上限で固定し、対象期間・選択可能総件数の業務制限にしない。appointment ID 数値昇順の keyset pagination。cursor は base64url-no-padding UTF-8 JSON の closed shape `{schema:"line-eligibility-cursor-v1",clinic_id,owner_id,binding_revision,after_appointment_id}` とし、各 ID/版は canonical 正 signed64 string。正規形・型・未知 key・URL owner/認証 clinic 不一致は400。各ページで現在 owners:view/reservations:view と行 scope を再検査し、cursor 自体を認可根拠にしない。現在 binding revision 不一致は409 ErrorResponseで先頭から再取得する。未終端は最後の返却 appointment ID の next_cursor、終端はnull。ページ横断の予約 snapshot は保証せず、表示候補の同時追加/変更は再取得でき、submit では選択全件の現在 CAS/認可を必ず再確認する。新しい署名鍵や永続 cursor state は追加しない。

通知再開候補の時刻predicateは `appointment.start_time > authoritative server現在時刻` とする。GET候補抽出で適用し、POSTでは選択appointmentのlock下の最終検査時点にauthoritative server現在時刻を再取得して全選択へ再適用する。開始時刻=現在時刻も対象外、当日開始済みも対象外とし、client clockやGET時点の古い時刻を保存時の判定根拠にしない。GET後に開始時刻へ到達した選択があればPOSTは再開せず全選択zero-writeのdependency-changed409とする。固定future horizon・対象期間の業務上限を追加せず、既存の状態・本人紐付け・現在権限・版・停止理由の検査は全て維持する。この境界は本人が「予約開始時刻が現在より後、開始済みは除外」に回答した「ok」による確定であり、既存の別用途readerから推測したものではない。

eligibilityには独立revisionが存在しないため架空のversion列をwire根拠にしない。既存operation_id＋state/binding_revision/appointment_versionを比較し、行欠落は明示nullの期待として区別する。APIにはoperation_idをcanonical stringでopaqueに返す。版競合後の再確認は同内容/同keyで現在Preconditionsを更新でき、選択集合変更は別の明示操作。

採択した権限は既存手動LINE操作のowners:editを再利用し、各予約の現在scope/閲覧可能性を別途検証する。新grantは追加しない。binding/対象がlinked以外、未確認、本人不一致、失権なら新しい再開不可。予約source変更、DM/独立停止、death由来、unknownは各理由を保持し、再開確認で解除しない。医療記録viewだけで送信権限を得ない。

unlinkとeligibility停止は469最外TX/graph参加に接続。relinkは停止を保持し、revisionを更新しても旧confirmedを再利用しない。死亡誤登録訂正も再開しない。具体的なgraph→C06 binding→owner/recipient/binding→pet/appointment/eligibilityの順は第7節に固定する。全予約一律lockを追加しない。保存直前にbinding revision・予約版・現在認可を再検証する。

### C：目的別確認・dispatch（後続技術研究、A/B完了判定と分離）

| 操作候補 | DTO / 結果 |
|---|---|
| GET `/owners/{owner_id}/followup-confirmation-context?medical_record_id=...&checkup_id=...&purpose=...` | source identity/version/digest、現在の実施根拠、医師指定の有無/根拠ref、binding revision、blocking理由。純read、checkup absentは省略。未取得を未実施/指定なしと断定しない |
| POST `/owners/{owner_id}/followup-confirmations` | `operation_key, medical_record_id, checkup_id:null\|string, purpose:general_guidance\|medical_followup, acknowledged_source_digest, expected_source_version, expected_binding_revision`。確認者/日時/医師指定はserverで取得。200 C06 ResultResponse（confirmation_id＋元source参照） |
| GET `/owners/{owner_id}/line-operations/{operation_key}?intent=line.followup_confirm` | 元owner/record/checkupの現在read認可、保存された確認結果を返す。送信成功を意味しない |
| GET `/owners/{owner_id}/line-dispatch-intents/{intent_id}` | immutable purpose/source/binding revision、確認ID、宛先の許可された表示、attempt/result別状態。現在owners:viewと元source read scope、secret/原evidence非公開 |

dispatch intent作成は既存の予約/案内送信callerが504 serviceへ渡す内部契約にし、任意payload/任意purposeを送れる新汎用public send APIを追加しない。既存POST `/owners/{id}/line/send`も同じgateへ接続する必要があるが、そのcaller全種の実装済みを本技術契約だけで宣言しない。明示送信preview/confirmationの必要subsetと未実装依存をPhase1範囲の中で固定する。

内部CreateDispatch入力は、元source typed ref、元事象identity、固定purpose、確認snapshot/payload digest、binding revision、必要ならfollowup_confirmation_id、元operation。DB purpose enumはreservation_confirmation/reservation_change/reservation_reminder/guidance/promotion/medical_followup。followup一般案内の`general_guidance`はdispatch `guidance`へ明示mappingし、同名だと仮定しない。

同元事象＋目的＋確認内容の重複防止はoperation keyだけでは足りない。新keyでも同一の元事象なら既存intentを照合するsource identity索引/claimが必要。現在DDLのUNIQUE(clinic_id,operation_id)だけでこの保証を満たすとは書かない。具体dedup projection/制約は共有storage ownerに要求する設計残（新時刻/配信回数ルールではない）。予約同日優先度等の既存claimを捨てない。

claim前と外部呼出直前で、現在source、確認内容、医師指定（medical）、binding、対象関係、DM、opt-out、停止、死亡由来、deploy/clinic gateと既存日次claimを検証する。claim後unlink時は未発信なら停止、開始済み外部作用の回収は保証しない。原子DB TXがHTTPとの競合を完全に解消したと主張しない。

結果状態は既存DDLのsucceeded/definite_failure/unknown/verified_absent。unknown/成功は新keyで救済しない。未送信確認後だけ現在条件を再判定し同intentの管理再試行。照合・retryのpublic操作が必要な場合も503と同様に元owner scopeへ閉じ、provider証拠をclient自己申告で上書きさせない。旧line_send_logs/owners.line_* / line_customers.owner_idへのfallbackは禁止。

## 5. 採択事項・実装依存・非実行境界

1. 503 現 Phase A/B：署名なし full immutable confirmation submit、新 confirmed path、純 preview、宛先別 read、管理 reconciliation/retry route を採用する。共有 READ・医院 mutation の既決 scope は変更しない。A は院内原子保存と配送未起動の表示、B は宛先別 claim・結果・照合・管理再試行を扱う。B の契約採択は外部送信実行の許可ではない。
2. 504 今回の技術整備A/B：A の既存 JSON adapter と B の選択集合 atomic 再開、eligibility operation_id CAS、owners:edit 入口を採用する。469 の binding writer と共通 lock 順・lookup anchor へ接続する。B の確認保存は即時通知を起動しない。
3. 504 C はPhase1範囲を狭めない後続技術研究として保持する。医師指定/実施 source の正本 ref・version/digest と同一事象 dedup projection は後続技術設計残であり、今回A/Bの完了判定とは分離する。doctor 列の存在だけを指定証拠にせず、新業務 workflow を捏造しない。
4. 503/504 の metadata table と domain repository は各 owner が所有する。schema 統合は446/466、共有 claim capability は489、binding writer closure は469、OpenAPI の統合は共有 API owner が担当する。採択済み能力の runtime 有無は別に確認し、共有 helper/router/OpenAPI を並行編集しない。
5. provider の不存在を証明できなければ unknown を保持し再試行不可。外部疎通・資格設定・キー生成・送信実行はこの契約で許可しない。モデル/DDL の存在を adapter や運用完成の証拠にしない。

## 6. 受入定義（未実行）

503 A：preview取消0 write、本文/宛先/添付差替え拒否、院内保存/監査失敗で外部0、同key同内容の1報告、別内容409、scope外拒否、共有readと医院mutation分離、保存成功/配送未起動の別表示。

503 B：宛先別片側成功/失敗、claim race1実行、外部成功後DB失敗unknown、検索不能は保留、verified_absentだけ次attempt、旧ID欠落再送禁止、削除後もreceipt/添付保持。

504 A：NULL/空/有効JSONを区別しroundtrip。504 B：unlink予約保持、relink/死亡訂正で自動再開0、選択対象だけ確認、未選択保持、binding/予約/eligibility競合zero-write、同key replayでも即送信0。

504 C（後続技術研究の受入であり今回A/Bの完了判定とは別）：一般/医療目的分離、医師指定なし拒否、記録createだけで送信0、claim後unlink/失権、DM例外と独立停止の分離、unknownの新key再送0、source同一別key重複0。

以上はfixture/testの定義であり、実行は全てNOT_RUN。503 現 A/B・504 Phase1内の今回技術整備A/B と後続技術研究 C の完了判定を分離する。採択 wire を OpenAPI へ公開する作業とconsumer実装・実機受入は別に判定する。

## 7. 閉じた C06 契約・claim metadata・B の実体接続

### C06 wire/HTTP

根拠は現 `backend/internal/operation/dto.go`、api.yaml L14145–14215、採用494/518。証明書consumer独自mappingとは混ぜない。503/504は同じouter shapeの専用成功refs schemaを定義する。

| status | HTTP | body |
|---|---|---|
| succeeded / replayed | 200 | status, operation_key, result_refs, committed_versions, recorded_at |
| not-succeeded | 200 | status, operation_keyのみ。成功ではない |
| operation-in-progress | 202 | status, operation_keyのみ。同key照会 |
| result-unavailable | 503 | status, operation_keyのみ。結果不明、再create禁止 |
| operation-key-conflict / version-conflict / dependency-changed / edit-right-lost | 409 | status, operation_keyのみ |
| validation | 400 | status, operation_keyのみ |
| scope-not-found | 404 | status, operation_keyのみ。binding照会不在専用 |
| not-authorized | 403 | status, operation_keyのみ |

この12 status以外を追加しない。503/504 Bに編集権fenceは要求しないためedit-right-lostを通常発生させないが共有schema enumを変更しない。成功参照/版/記録時刻は保存済みreceiptの原値で、現在配送状態は別GET。全operation lookup queryはintentだけ（unknown query400）、各routeの下記allowlist以外を400。全lookupはprivate,no-store。

専用success refsのclosed keys：503 create=`report_id,intent_id`／versions=`report`、503既存report確認=`report_id,intent_id`／versions=`report`、503照合=`report_id,destination_id,attempt_id,result_id`／versions=`result`、503 retry=`report_id,destination_id,attempt_id,authorized_retry_result_id`／versions=`attempt`（保存されたattempt_number）、504 B=`owner_id,resumption_operation_id`／versions=`binding`。可変の選択予約詳細はeligibility GETで取得しreceiptに新arrayを足さない。resumption_operation_idは保存済みC06 binding ID。履歴再現に必要な選択集合は不変owner operation metadataに保持する。

lookup intent allowlist：support全体routeは`support.report_confirmed_create`だけ、report配下は`support.delivery_confirm,support.delivery_reconcile,support.delivery_retry`だけ、504 B owner配下は`line.notification_resume`だけ。後続技術研究Cは今回A/Bのallowlistへ先取り登録しない。

pre-C06 errorは既存 `ErrorResponse` の `{error:string}` を採用し、C06 status/key/refsを混ぜない。HTTPは400=不正構文/型/UUID/未知field、401=未認証、403=現在認可拒否、404=対象不在/scope外/認可後URL不一致、409=確認内容/依存版変化、413=既存upload上限、415=既存media型拒否、422=owner意味検証、500=DB/invariant、503=サービス利用不可。新error.code enumを追加しない。query/submit応答喪失時は元key lookupへ。外部provider失敗はcommit済み内部操作のHTTP成功を巻戻さず、配送result DTOで表す。

### Metadata anchor の採択 schema

[#777 / EMR-502 の既採択契約](https://github.com/MinoruSoga/AnimalEkarte/blob/19855942f762f31f0403b5de0ddbc61b761fd94e/docs/work/emr/emr-502-manual-edit-contract-20261021.md#2-claim-と業務保存の順序)の optional capability `ClaimMetadataWriter` と `CreateClaimMetadata(ctx, binding, request) error` を再利用する。別名の ClaimTarget capability を新設しない。`NewManualEditExecution` は同文書第4節の manual 専用 factory 提案名であり、実装済みの共有 constructor ではない。503/504 もその単一 factory・immutable typed command・Content commitment 再計算という既採択パターンを使い、manual 型を流用したり未検証 payload を callback に捕捉したりしない。

494 の record anchor 所有と本契約の domain anchor 所有は分離する。record 専用表に support/owner を詰め込まず、owner 専有の `support_operation_targets` / `line_notification_operation_targets` の2表を採用し、446/466 の単一 schema writer に統合する。

- 共通列はすべて NOT NULL：`clinic_id bigint, operation_id bigint`（PK は `(clinic_id,operation_id)`）、`target_kind text, target_ref text, metadata jsonb`。両 bigint は正の signed bigint、CHECK > 0。生成 ID/default/version/actor/時刻を重複追加しない。即時・非遅延 FK `(operation_id,clinic_id) -> operation_bindings(id,clinic_id) ON DELETE RESTRICT` のみを置く。report/owner/appointment/staff の業務 FK は付けず、claim 前の業務 row KEY SHARE lock を先取りしない。
- table ごとの CHECK で target kind/shape を閉じる。support 表は `support_report_creation` 又は `support_report` だけ、LINE 表は `owner_line_notifications` だけ。creation target_ref は operation key と同じ canonical UUID、report/owner target_ref は canonical 正 signed64 decimal string（`[1-9][0-9]*`、最大9223372036854775807）。leading zero・符号・空白・範囲外を拒否する。creation UUID と binding key の一致、owner_kind/intent は binding 結合時にも検証する。
- metadata は JSON object の closed shape。support は schema 1 key のみ・値 `support-target-v1`。LINE は schema/appointment_ids の2 keys のみ・schema 値 `line-notification-target-v1`。appointment_ids は非空 JSON array、各要素は canonical 正 signed64 decimal string、numeric 昇順・重複なし。null、number 要素、空配列、未知 key、範囲外を拒否する。DB の table-local CHECK/検証関数と domain 検証の双方で閉じた shape/ID 配列を検証し、JSON の型だけで済ませない。callback は canonical Content に拘束した選択集合との完全一致を確認する。
- repository は INSERT/SELECT のみ。両表に表ローカル BEFORE UPDATE OR DELETE 拒否 trigger を設け、ordinary invoker とする。SECURITY DEFINER、新 role/権限、更新/補修/backfill/cleanup endpoint は追加しない。runtime に TRUNCATE/DDL/trigger 無効化を許さず、DB owner/superuser まで防ぐとは主張しない。
- 両表へ既存 `app_private.apply_rls_policy` と `app_private.has_clinic_access(clinic_id)` の FOR ALL USING / WITH CHECK を適用する。既存 ENABLE の境界を維持し FORCE/bypass 方針を新設しない。repository も trusted selected clinic + operation binding で絞り、binding の actor/owner_kind/intent と現在の domain scope/権限を別途検査する。RLS は actor/intent/予約ごとの認可を代替しない。
- support new create：target_kind=`support_report_creation`、target_refは同scopeのoperation UUID、metadataは `{schema: support-target-v1}`。actor/clinic/intentはbindingから取得。成功report IDは同TX成功receiptのrefsとreport clinicを検証して解決。失敗時は同actor/currentclinicのcreate権限だけを認可根拠にし、任意の共有reportを代用しない。
- support既存report：target_kind=`support_report`、target_refはcanonical report ID、metadataは `{schema: support-target-v1}`。管理reconcile/retryも元destination→intent→reportをpreflightで解決し、Contentへreport/destination/attemptを束縛する。claim後はincoming IDで元anchorを変更しない。
- 504 B：target_kind=`owner_line_notifications`、target_ref=canonical owner ID、metadataは `{schema: line-notification-target-v1, appointment_ids:[canonical ID...]}`。選択集合はContentと同じnumeric ID昇順・重複なし。不成功bindingにも保持。現在の同owner/各appointment read権限でlookupを認可する。
- INSERT ... DO NOTHING RETURNING の winner だけが ClaimMetadataWriter callback を ambient claim TX 内で実行（RETURNING/RowsAffected で判定）、anchor失敗でbindingもrollback、既存bindingはmetadataを修復/更新しない。callback は canonical req.Content と verified immutable command のみを使用し、nested TX・Execute 再入・業務 row 読取/lock・履歴/receipt 書込・外部作用を禁止する。競合 loser は原 binding/anchor を読み callback を再実行しない。claim はここで commit し業務 TX へ持ち越さない。anchor欠落/曖昧は403 not-authorizedで開示停止。DB読取故障で解決できない場合は503 result-unavailable。claim後のscope変更は新Applyで拒否しmetadataだけ保持。

### 504 B の認可と lock 順（採択契約と実装観察を区別）

現在実装：`persistence.AcquireReservationGraphWriteGate` / `AssertReservationGraphWriteGate` は現mainに存在する。古いrevisit-intake.mdの「共通実装未確認」は当時の履歴。gate namespaceは既存 `reservation-graph:write:v1`、最外TXの最初の業務read/lockより前。現 `operation.Runner.participate` は error-only Participant.Gate(ctx)→binding FOR UPDATE NOWAIT→保存 receipt/unknown 検査→新規なら savepoint→Apply の順であり、新規 binding の Apply 前に ResolveLookupTarget/AuthorizeLookup を自動実行しない。保存結果の disclose/replay 時の現在 read 認可と新規 write の認可を混同しない。503/504 の Apply 自身が immutable 元 anchor と trusted actor/clinic を解決し、同 TX 内の現在 write grant/read scope・対象関係を業務 write 前に必ず再検証する。入口 preflight だけでは代替不可。既採択 ClaimMetadataWriter capability の runtime 統合は489の責務であり、この採択を現 runner への実装済み宣言にしない。

現在の469対象W08 `ownerRepository.UpdateLineUserID` は owner FOR UPDATE→旧新recipientをsort/dedupe→`AcquireLineRecipientLock`→旧owner列write。W09 `lineCustomerRepository.UpdateOwnerLink` はbaseDB直接旧owner_id更新。これらは新binding/eligibility contractへ適合済みではない。469が同じgated ambient writerへ置換する依存を明示し、504が旧コードを再使用して正本を戻さない。

504 B 採択 lock 順：短いmetadata claim終了→最外業務TX開始直後のgraph gate→C06 binding NOWAIT→保存 receipt/unknown 検査（開示時は元 anchor 現在 read 認可）→新規は Apply savepoint→Apply 内の元 anchor/current write/read 認可→owner row FOR UPDATE→現在/旧recipient lockを文字列昇順（変更しないresumeは現在recipientだけ）→owner_line_binding_states FOR UPDATE→対象pet行ID昇順FOR SHARE→選択appointments ID昇順FOR UPDATE→reservation_notification_eligibilityを同appointment順FOR UPDATE→現在認可/版/停止再照合→selected eligibility＋必須 audit_logs＋receipt commit（監査失敗は全選択とreceiptをrollback）。eligibility行欠落のinsertは親appointment lockで直列化する。受理するsource/owner/LINE writerは全て上位graph gateへ参加し、下位取得後にgateを後付けしない。469 unlink/relinkもowner→recipient→binding→影響appointment→eligibilityの相対順を共有する。移管の既存NOWAIT保護は維持し、別業務の既存下位順を勝手に反転しない。相対順を共有できない既存writerは469のclosure残として停止/adapter化し、runtime cutover条件を明記する。

route grant 契約：eligibility GETとoperation lookupはowners:view＋認証医院owner scope＋選択予約のreservations:view/現在行scope。POST resumptionsはowners:edit＋同じ読取scope。既存LINE手動writeのowners:editを再利用し、reservations:editや新resourceを要求しない。GET router は read grant、POST router は既存 owners:edit middleware 必須とする。POST は成功 replay を含むすべての入口で現在 edit grant を要求し、その後も原 anchor の現在 read scope を再確認する。view-only 利用者には POST 成功 replay を許可せず、GET operation lookup で同 actor/clinic の保存済み原 receipt を取得する。GET は edit 不足を新書込権限へ昇格しない。既存469 PATCH `/owners/:id/line-user-id` / `/line` とline-id-confirmはowners:editを維持。既存DELETE独立経路を本契約だけで権限変更しない。503は専用grantなしの認証済スタッフ共有READ、mutationは報告医院のまま。

今回の技術整備の残りは上記新metadata schema/489 callback、469全対象writer参加と新bindingへの置換、採択済み public OpenAPI の反映・公開。504 Cのdoctor source/dedup/全triggerの研究残は今回B・Aの完了判定と分け、LSTEP／DMのPhase1範囲を除外しない。署名鍵の準備・expiry選定は残件から削除した。


## Revision 4 追加採択契約：context gate・必須監査

### 489 の nonbreaking ContextGate capability（単一方式）

実 `AcquireReservationGraphWriteGate(ctx)` は gate marker を持つ derived context を返す一方、現 Participant.Gate(ctx) は error-only、現 Runner は元 ctx を使い続ける。このままでは下位 AssertReservationGraphWriteGate が失敗する。既存 Participant interface は変えず、optional capability `ContextGate` の `GateContext(context.Context) (context.Context,error)` を489が追加する採択契約に固定する。Runner.participate は最外 WithTx callback の先頭で capability があれば GateContext だけ、なければ既存 Gate だけを1回実行し、両方を呼ばない。

504 participant は GateContext で既存 AcquireReservationGraphWriteGate に渡された tx ctx をそのまま渡し、返された ctx を返す。Runner は nil ctx と ambient TxFromContext の pointer 不一致を拒否し、元 tx と同一であることを検査する。以降の binding lock、receipt/unknown 読取、disclose、savepoint、Apply、監査、InsertResult、同 TX 内の unknown 記録まで returned ctx を一貫して使う。504 は returned ctx で AssertReservationGraphWriteGate を維持する。marker の自作/participant mutable field への ctx 捕捉/Assert 迂回/nested TX/DetachTx/下位 lock 後の再取得は禁止。実 graph gate が不要な503・既存 consumer は既存 Gate 挙動を維持し、新capabilityの副作用を及ぼさない。rollback 後の別 TX・read-only lookupへ終了済み gate ctx を流用しない。これは489の実装依存であり現在実装済みではない。

### audit_logs の同 TX event 契約

既存 audit_logs 列へ明示的に写像する。staff 起因は認証 clinic_id/actor_id、actor_type=staff、worker 起因は actor_type=system/actor_id=null とし、元 staff operation は metadata の参照に保存する。created_at はserver記録時刻。新しい IP/UA 収集必須要件を追加しない。本文/画像bytes/credential/provider証拠原文を監査へ複製しない。

- 503 create の action=`support_bug_report.create`、resource=`support_bug_report`、resource_id=report_id、old_value=null、new_value は report_id/intent_id/report_version/confirmation_digest。既存報告確認は action=`support.delivery_confirm`、同 resource/root、old/new に前後 report版と今回 intent/digest参照。
- 503 logical delete の action=`support_bug_report.delete`、resource=`support_bug_report`、resource_id=report_id。認証由来のtrusted actor/clinic/report IDと前後deleted_at/削除状態を必須とし、論理削除と同じsupport-owner TXでaudit_logsへ保存する。監査欠落/失敗は削除もrollback。実在operation参照があるときだけmetadataへ保持し、既存経路に架空C06 bindingを作らない。添付原本の物理削除を起動せず、receipt/intent/historyと再送抑止identityを保持する。
- 503 reconciliation/retry は action=`support.delivery_reconcile` / `support.delivery_retry`、同 report resource/root、old/new に対象 destination_id/attempt_id と前後 result_id/revision/state（不存在はnull）又は前後 attempt ID。metadata は schema=`support-audit-v1`、C06 operation_id/key/intent。worker claim/result の action=`support.delivery_claim` / `support.delivery_result`、同 report root、元 operation/intent/destination/attempt/result 参照を保持。claim と監査成立前に HTTP を開始せず、外部結果＋必要監査の保存失敗は外部効果を rollback と主張せず unknown 照合へ。
- 504 B は action=`line.notification_resume`、resource=`owner`、resource_id=owner_id。metadata は schema=`line-notification-resume-audit-v1`、C06 operation_id/key/intent、binding_revision、numeric昇順の選択 appointment_ids。old_value/new_value は同じ選択順で appointment_id、eligibility state/binding_revision/operation_id（行不存在はnull）を保持し、全選択 eligibility の変更と1 audit rowと成功 receipt を同一最外 TX で commitする。監査欠落/失敗は全選択 zero-write、receipt不成立。replay/GET は新しい監査行や送信を作らない。

追加 fixture 定義（実行 NOT_RUN）：GET/POST両方の開始時刻>server現在時刻、等時刻/当日開始済み除外、GET後開始到達で全選択zero-write、client clock変更無効・固定horizonなし、異key同semanticの同時確認/全配送状態で新intent0、同keyreplay優先、別report同文は非統合、新規null refsと生成後refsの同一identity、2digest正規化/宛先順序/重複拒否/new INSERT NULL拒否、legacy厳密比較/復元不能500送信0/既存重複で移行停止、初回/再試行の外部 call 前に report 版又は本文/添付が変われば送信0・再確認停止、新規作成 snapshot の null refs を保持したまま intent の生成後 report identity/version と照合、resultなし/null CAS の照合と競合、fresh Apply前失権、503/504監査失敗全rollback、eligibility複数ページ/不正cursor、ContextGate伝播・同TX assert・旧participant互換、過去成功後のprovider削除で再送不可、submit後設定変更で初回send0、通常DELETE後も添付原本保持、削除監査失敗で論理削除rollback、operation参照なしの既存DELETEでも架空bindingなしで必須監査成立。外部書込・コード/DB変更・試験・送信は本契約で行わない。
