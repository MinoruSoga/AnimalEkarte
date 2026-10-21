# LIFF問診・接種証明書の再読込回復契約（統合採択技術契約）

2026-10-21。親採択済みの技術契約。runtime未実装・受入未完了であり、本人が物理方式を直接選定した意味ではない。R2を置換する全文。外部write、DB適用、provider操作、試験は行っていない。業務契約基準 edf71515dbae11735a4184ff64c465e2ea51af88。C06 runtime は 2026-10-21 05:38 UTC に最新 main 0de244ba69f9287a0bacff6abeddabb50a712aac の contract.go / runner.go / store.go を再取得確認。

## 1. 保証と対象

対象は本人のLIFF問診submitと職員本人の証明書issue/reprintだけ。ブラウザreload後も、未確定操作を発見して結果照会し、遅延して到着する旧送信を閉じてから次の明示操作へ進める。結果照会と安全な次操作許可は別である。

localStorage、sessionStorage、IndexedDB、追加URL/query/hash、Service Worker cacheへ本文・操作key・回復世代を永続化しない。メモリMapは再mount用補助でありreload回復保証ではない。reloadで失った未送信本文は復元保証外。serverに新しい本文draftを保存しない。既存lookup path中のoperation keyは既存契約どおりで、禁止対象の追加URL永続化とは区別する。

operation_keyが唯一の業務操作identity。回復世代は遅延要求の実行可否CASであり第二request IDではない。成功済み操作を取り消さず、履歴・原receiptを不変のまま保つ。PDF取得・紙の印刷成否を発行成功と混同しない。

## 2. 現存部分と必要追加

- C06 の上記再取得 HEAD の Participant はGate(ctx) error、Apply(ctx,binding,request)。Runnerは独立claim後に業務TXを開始しGate→binding NOWAIT→receipt/unknown→Apply。現store ClaimBindingはambient TXを拒否する。この HEAD には ContextGate / ClaimMetadataWriter / 原target同時claim機能がない。別の後発レビュー snapshot には winner-only ClaimMetadataWriter が存在するが、現在 main へ入った根拠ではなく、以下の必要差分をその別版にも照合する。
- 証明書のimmutable vaccination_certificate_operation_targetsとcanonical snapshot/manifestは486の既選定契約だがruntime未接続。本回復はそのtargetを再利用する。成功だけのvaccination_certificate_operationsを検索して未成功を落とさない。
- LIFFのrevisit_intake_submit_receiptsは(clinic_id,author_customer_id,operation_key) PK、author_subject_hash、request_hash、claimed/succeeded、result_answer_idを持つ。未成功にはtargetがないため、回復用target metadataを追加する必要がある。
- C06 sticky unknown、現在元target認可、LIFF本人scope＋失権時redacted成功は維持。C06不明をnot-succeededへ落とさず、LIFF claimedを実行中や失敗確定へ読み替えない。

以下のroute/schema/controlは新しい限定技術契約で、既存APIの利用だけでは完成しない。

## 3. Scopeとwire型

LIFF scope = trusted clinic＋認証customer＋認証subject hash＋appointment_id＋pet_id＋target_generation＋固定intent revisit-intake.submit。subject hashは既存の認証由来SHA256方式、token/sub原文を保存しない。target世代は既存の問診対象identityでありrecovery_generationとは別。

証明書scope = trusted clinic＋staff actor＋intent＋vaccination_id＋edition_id。intentはwire issue/reprintを既存C06 vaccinations.certificate.issue/reprintへ変換。issueはedition_id=null、reprintは正のedition_id。kind/layoutは既存Contentに含み、controlを別kind/layoutに分割しない。同じ本人・接種・intentの未確認発行を迂回しないためであり、normal/rabies seriesを統合する意味ではない。

IDsは各domainの既存wireを維持する。LIFFはcanonical positive decimal strings、証明書は現契約integer/int64（安全なdecode必須）。新recovery_generationは両方canonical positive decimal string（int64上限）。operation_keyは既存canonical UUID非nil。clientからclinic/customer/staff/subject hash/control_idを受理しない。JSONはclosed schema、未知field・重複key・不正型は400。レスポンスはno-store。

## 4. Discovery API（純read）

LIFF実route:
POST /api/liff/{clinicId}/revisit-intake/operation-recovery/query
body oneOf:
- {scope:"target",target_ref:{appointment_id,pet_id,target_generation},cursor:null|string}
- {scope:"self",cursor:null|string}

証明書実route:
POST /api/v1/vaccination-certificate-operation-recovery/query
OpenAPI pathsは /vaccination-certificate-operation-recovery/query。
body {target_ref:{vaccination_id,edition_id},intent:"issue"|"reprint",cursor:null|string}。

response必須keys = {items,next_cursor,complete,recovery_generation,active_operation_key}。target queryはgenerationとactive keyを返し、selfはどちらもnull。control不存在は論理generation="1"、active=null、readによるrow作成なし。active keyを返す場合も元receipt scope/現在認可に従う。失権した対象のtarget queryは存在非開示、LIFF selfへ誘導可能。

item必須keys = {operation_key,intent,claimed_at,discovery_state,execution_state,result}。
discovery_state = result | unavailable。
execution_state = succeeded | unresolved | closed_without_effect。
resultは当該domainの既存lookup response又はnullで、既存wireを混ぜない。成功receiptの権威ある読取のみsucceeded。closure証拠がある未成功keyのみclosed_without_effect。lookup不調/metadata不明はunavailable＋unresolved＋result=null。成功が確認できても現在本文を読めないLIFFは既存redacted200 resultのみ、target/answer/版/本文を追加しない。certificateは現在元target viewがなければ非開示でありredacted成功を新設しない。

順序(claimed_at DESC,operation_key DESC)、1page50。cursorはversion付きclosed payload {v:1,upper_claimed_at,upper_key,last_claimed_at,last_key} のbase64url表現で、scopeはrequest/authで毎回再検査。cursorを認可証明にせずPII/主体hashを含めない。不正400。改ざんで自分の過去行のページ区切りが変わっても権限は増えず、後段fenceはこのページ集合を信用しない。completeは当該上限の走査終了だけを表し、新claim不存在や次操作許可ではない。別actorのkey/件数は返さない。

LIFF selfは自分のreceiptだけを探し、未成功legacy metadata不明はunavailableとして残す。target不明を現在予約から推定しない。target queryに現れない旧不明keyがある場合、後述cutover gateが解除されるまで新操作を止める。

## 5. 明示次操作fence API

LIFF:
POST /api/liff/{clinicId}/revisit-intake/operation-recovery/fence
body {target_ref:{appointment_id,pet_id,target_generation},expected_recovery_generation,acknowledged_operation_keys:[UUID...]}

証明書:
POST /api/v1/vaccination-certificate-operation-recovery/fence
body {target_ref:{vaccination_id,edition_id},intent:"issue"|"reprint",expected_recovery_generation,acknowledged_operation_keys:[UUID...]}

keysは昇順・重複なし。対象generationにclaimされた全keyを、人が結果確認した集合として渡す。移行が完了した新経路では最大1key、controlがまだない場合は空集合。過去世代の全履歴をack対象へ混ぜない。

200 response = {state:"advanced"|"replayed",previous_recovery_generation,recovery_generation,active_operation_key:null,acknowledged:[{operation_key,outcome:"acknowledged_success"|"closed_without_effect"}]}。
同じexpected generation＋同じtarget/intent＋同じcanonical ack集合の再要求は保存済みtransitionを返すだけで再増分しない。返すgeneration/active=nullは当該遷移直後の履歴であり、現在も空である保証ではない。UIはreplayed後にtarget discoveryで現在generation/activeを再取得してから次操作へ進む。expected generationのtransitionが存在してもack集合が異なれば409 RECOVERY_CONFLICT。世代をまたいだ盲目的再送は不可。operation_keyや第二idempotency keyをfenceへ追加しない。transitionの(control,previous_generation)一意性が応答断の再読取を可能にする。

非2xxはこの新APIだけのclosed {error,code}：400 VALIDATION、401認証失敗、403 NOT_AUTHORIZED、404 SCOPE_NOT_FOUND、409 RECOVERY_CONFLICT（世代/ack差）、409 RECOVERY_BUSY（lock取得不能）、409 RECOVERY_REQUIRED（legacy未解決/対象不明）、503 RESULT_UNAVAILABLE（読取不明/sticky unknown/commit不明）。safe messageのみ。新codeを既存C06共通Statusへ追加しない。特定scopeの存在を漏らす前に現在認可する。

fenceの200は新業務操作成功ではなく、旧世代が以後実行不可という証拠。新しい本文・接種・PDFを生成しない。次操作は最新対象/版を取得し、必要previewと明示操作を経て新operation keyを生成する。成功確認だけで自動的に次writeや再印刷を送らない。

## 6. 認可

- certificate discovery/既知成功lookupは既存元targetへのvaccinations:view。issue fenceは既存issue実行認可vaccinations:view＋vaccinations:edit、reprint fenceは既存reprint実行認可vaccinations:view。viewで許可済みの限定再印刷controlを進めるだけで、接種edit/新issue/別staff閉鎖/一般DBwriteの権限を付与しない。
- LIFF target discoveryは現在target読取、selfは既存本人receipt scopeのみ。fence/新submitは現在本人target/editable条件を再検査。受付後など新回答不可なら旧成功確認だけ可能で、fenceから回答を再開しない。
- 新claim前に現在の実行認可をnonlocking preflightし、不許可requestでcontrol/binding/anchorを増殖させない。Apply/fenceの業務TX内でも再確認する。事前readはcommit権限ではない。
- 既存成功keyは元targetと現在lookup認可を先に解決し、現在CAS/世代/active不一致より先に成功receiptを返せる。同key異Contentは元認可後に既存conflict。新requestのURL/bodyを元target認可の代用にしない。POST middlewareがissue権限を要求する場合、権限を失った本人の成功回復はGET既存lookup/discoveryで行い、mutation入口を勝手にviewへ弱めない。

## 7. DB delta（具体選択案）

全表はtenant RLS＋明示clinic predicate、通常DELETE不可。以下で nullable と明示する列以外は全て NOT NULL（identity、世代、hash、監査時刻を含む）。臨床本文/送信tokenを含めない。schema ownerで001/ERD/modelを同期する。

### 7.1 LIFF targets

revisit_intake_operation_targets:
clinic_id bigint、author_customer_id bigint、operation_key uuid（既存receiptと同PK/FK RESTRICT）、appointment_id bigint>0、pet_id bigint>0、target_generation bigint>0、recovery_control_id bigint、recovery_generation bigint>0。subject identityはreceiptとcontrolで照合する。INSERTとreceipt claimは同じ短TX、UPDATE/DELETE拒否。business rowへのFKは置かずclaimで患者row lockを増やさない。原対象metadataであり現在対象へ上書きしない。

### 7.2 Controls（domain別2表）

revisit_intake_recovery_controls:
id bigserial PK、clinic_id bigint、author_customer_id bigint>0、author_subject_hash bytea length32、appointment_id bigint>0、pet_id bigint>0、target_generation bigint>0、generation bigint>0 default1、active_operation_key uuid nullable。
UNIQUE(clinic_id,id)、UNIQUE(clinic_id,author_customer_id,author_subject_hash,appointment_id,pet_id,target_generation)。active keyは同clinic/customer receiptへFK（null許容）、target metadataは(clinic,id) controlへFK。

vaccination_certificate_recovery_controls:
id bigserial PK、clinic_id bigint、staff_id bigint>0、intent text enum issue/reprint、vaccination_id bigint>0、edition_id bigint nullable、generation bigint>0 default1、active_binding_id bigint nullable。
intent は NOT NULL かつ CHECK (intent IN ('issue','reprint'))。NULL 安全な完全式は CHECK ((intent='issue' AND edition_id IS NULL) OR (intent='reprint' AND edition_id IS NOT NULL AND edition_id>0))。UNIQUE(clinic_id,id)。issueは(clinic,staff,intent,vaccination)のpartial UNIQUE、reprintは(clinic,staff,intent,vaccination,edition)のpartial UNIQUE。active bindingは同clinic bindingへFK。owner_kind/actor/intentの一致をrepository/guardで検査。target metadataへrecovery_control_id/recovery_generationを追加し(clinic,control) FK、不変。

internal control idはAPIに出さず、業務request identityにしない。初期INSERT時だけgeneration1/activeNULL、その後は次の状態遷移だけ許可: claim時に同generation activeNULL→当該key、成功時はactive維持、fence時generation+1/activeNULL。同世代key差替え・手動generation上書き・成功時自動解除は禁止。世代int64上限は503で停止しwrapしない。

### 7.3 Transitions / closures（domain別各2表）

revisit_intake_recovery_transitions と vaccination_certificate_recovery_transitions:
clinic_id,control_id,previous_generation,new_generation,acknowledgement_sha256 bytea length32,recorded_at(server timestamptz)。PK(clinic,control,previous_generation)、UNIQUE(clinic,control,new_generation)、CHECK new=previous+1、control複合FK。append-only。

対応する *_recovery_transition_items:
clinic_id,control_id,previous_generation,operation_key, outcome enum acknowledged_success/closed_without_effect。
PK(clinic,control,previous_generation,operation_key)、transition複合FK。LIFFはauthor_customer_idも持ち同clinic/customer/key receipt FK、certificateはbinding_idを持ち同clinic/binding FK。control principal、元metadata、key/bindingの一致をguard/repositoryで検査。append-only。

空ackの世代進行もheader1行を必須保存し、items0件とする。header hashはSHA256(JCS({schema:"operation-recovery-ack/v1",operation_keys:昇順UUID配列}))。scope/intent/generationはheader/親controlへ型付きで保存し、hashだけへ依存しない。履歴をstaff/systemのaudit_logsへLIFF customerとして偽装しない。この追加sinkが回復操作監査で、domainの元結果監査は変更しない。

closed_without_effectは元receiptを書き換えない追加終結証拠。DB-onlyの効果なしを同lock下で確定したkeyだけを記録する。successreceiptがあるkeyにはacknowledged_successしか許可しない。元receipt/transition/metadataの矛盾は503で停止し、片側を削除して修復しない。

## 8. Claim / lock / Apply の閉鎖性

### 新claim

既存keyの検索→元target現在認可・Content一致確認を先に行う。新keyのみ実行認可preflight後、短いmetadata TXでcontrolを一意INSERT/SELECT FOR UPDATE→expected generation/activeNULL→binding又はLIFF receiptをclaim→winnerだけtargets INSERTとactive設定→commit。FK/insert失敗は全てrollback。別targetで同scope/keyを同時claimしたloserは自分のcontrolをactiveにせずTXを閉じ、勝者の元metadata/認可へ戻る。自分control lockを持ったまま別controlを取得しない。安全な短いlock待ち又はbusy処理とし、待ち時間を新lease/失敗確定に変換しない。

新requestへの配置は、LIFF既存submit JSONのトップレベル expected_recovery_generation（required）、証明書issue/reprintの既存preconditions内 expected_recovery_generation（required）とする。発行前previewは純readのため必須化せず、generationをContentやpreview fingerprintへ混ぜない。LIFFのhashは既存intent/target/form/7項目だけ、証明書Contentは既存7/4fieldsを維持。recovery世代はPreconditionでhashに入れない。ただしkeyの保存targets世代は不変のため、古い閉鎖keyへ新世代を送って再利用することはできない。

### 業務TX

LIFF: reservation-graph:write:v1 → control FOR UPDATE → receipt FOR UPDATE →既存appointment/target/current authorization locks→answer/receipt commit。
証明書: 既存consumerで必要なowner Gate（不要ならno-op）→control FOR UPDATE→C06 binding NOWAIT→既存D08 parent/child/series順→domainwrite/receipt commit。
controlはcommitまで保持。Gate は既存 immutable 元metadataから選んだ control の lock と必要な owner gate だけを担い、現在 request target から control を選ばず、generation/active/closure/現在 editability を理由に拒否しない。Runner の binding lock→成功receipt再読→sticky unknown の順を維持し、未成功 Apply の直前だけ現在認可・世代・active・closure を検査する。成功replayでは旧世代でも元成功を返す。未成功Applyはtargetsの保存control/genが現在値に一致しactive keyが自分であることを検査しなければwrite不可。closureが存在する旧keyは同Content・新CASでも実行不可。C06では既存409 dependency-changed、LIFFは新consumer code409 OPERATION_CLOSEDとして返しlookup/discoveryへ誘導する。C06共通enumは変更しない。

### fence

現在実行認可preflightの後、control不存在なら一意INSERTでgeneration1/activeNULLを作る（純readは作らない）。同じ業務gate→control→対象binding/receipt（複数移行対象がある場合key昇順）→既存対象・現在実行認可検査の順。controlの世代とtargets集合をserverで再照合しack完全一致を要求する。successを権威あるreceiptで確認し、LIFF claimedはreceipt lock下で成功answer/receipt不在が確定する場合のみ閉鎖可能。C06はsticky unknownなし、binding lock取得済み、成功receiptなし、DB-only effect/commit状況が確定している場合だけ閉鎖可能。commit不明やlookup不能をabsence扱いしない。unknownは従来解決ownerへ渡して停止。

同TXでtransition/header/items＋generation+1/activeNULLを保存。どれか失敗すればrollback。fence応答断はtransitionを現在認可付きで再取得し、同じexpected generationの再要求も同transitionへreplayed。遅延claimが先に成功すればfenceはそのkeyをack不足として拒否、遅延Applyが先に成功すればそのreceiptを回復。fenceが先なら旧世代claim/Applyが拒否される。

外部render/uploadは業務TX外のまま。旧世代準備処理がfence後にblobを作っても、業務発行commitへ入れず、未参照候補として既存cleanup設計へ渡す。PDF準備成功を業務成功にしない。

## 9. 現C06で不足するextension

現ClaimBindingを呼んだ後に別TXでtargets/controlを追加する実装は禁止。crashで未発見keyが残るためである。489は既存APIを破壊せず、限定consumer用のclaim transaction coordinatorを追加する必要がある。

coordinatorの契約: 既存scope/key/content fingerprint計算→短TX開始→consumer BeforeClaimでtyped controlをlock/検査→同TX内binding INSERT/競合winner読取→新winnerのみimmutable targets/activeを保存→commit。既存keyは保存元targetを返し、現在requestのtargetでanchorを作り直さない。既存ClaimBindingのambient禁止を削除して全callerへ開放するのでなく、内部TX用store primitiveと限定coordinatorに分離する。metadata失敗ではbindingもrollback。副作用/外部呼出禁止。

### 既存bindingの専用dispatchと成功優先

consumer は claim/Execute の前に同scope/keyの既存bindingと immutable 元metadataを preload する。元認可を現在 request body へ fallback しない。既知成功が見える場合は現在元target lookup認可と既存 fingerprint 一致判定を経て元結果を返す（異Contentは元認可後 operation-key-conflict、本文非開示）。この専用成功経路は新control存在・現在世代・active・業務CAS・現在editableを要求せず、一般 C06 の全consumer順序を変更しない。mutation middleware の既存 edit 要件は維持するため、edit失権後の成功回復は GET lookup / discovery に限る。

成功未確認で control metadata がある場合だけ、元controlを注入した Participant で通常業務TXへ進む。Gate では lock のみ、binding lock 下で成功receiptを再読した後の未成功 Apply だけを世代/closureで拒否する。preload後に成功がcommitされた場合もこの再読で回復する。元metadataがない legacy は既存 Lookup の権威ある receipt/unknown 照合だけを使用し、成功なら元receiptによる現在元target認可で返す。未成功又は対象を確定できない場合は RECOVERY_REQUIRED / 非開示 / RESULT_UNAVAILABLE の該当結果で停止し、Applyへは進めず新controlの推定作成もしない。旧keyを新request targetへ再bindしない。

### 既存hookの利用範囲と追加差分

最新確認 HEAD は hook 未実装である。一方、別の後発 ClaimMetadataWriter 版を実装baselineに採る場合は、その winner-only 同claim-TX callback と失敗時 rollback を再利用し、同機能を二重追加しない。ただしその callback は binding INSERT 後に呼ばれるため、それだけでは本案の control→binding lock 順も BeforeClaim の世代/active検査も満たさない。追加必須は binding INSERT より前の限定 BeforeClaim control lock と、winner-only callbackでの targets/active 原子的保存、既存binding専用dispatchである。binding→control の逆順で代用しない。

error-only Gate でも単なる control lock は可能だが、共有graph gateが派生ctxを返す場合は既採択 ContextGate subset（又は同等の派生ctxを以後の binding/Apply 全部へ渡す限定adapter）を実装依存として明示する。error-only Gate 内で派生ctxを捨てる実装は不成立。この最新 HEAD に ContextGate 実装済みとはしない。既存Runnerのunknown保存・commit不明reconcileは維持し、consumer参加計画で元target認可→成功replay優先→未成功guard→Applyを固定する。

LIFFはC06を無理に流用せず既存receipt/answer triggerを維持し、同じ制御順の専用claim adapterを追加する。control/target世代を検査せずanswerを作れる既存入口を残さない。

## 10. Legacy / 利用開始

generationを送らない旧clientの新writeは新経路では400 VALIDATION、既知成功lookupは引き続き可能。既存未成功targetsに世代を推定backfillしない。古いsenderがなお実行可能ならcontrol導入だけで安全とはいえない。

cutover前に新対象全writerがguard対応し、旧入口が新claim/Apply不能であることを確認する。legacy未成功key/target不明が残るprincipalはRECOVERY_REQUIREDで停止し、元receipt/効果の権威ある照合と旧writer停止を完了するまで新targetへも自動再開しない。TTL・0件・一定時間経過で解除しない。移行の実操作は本設計では実行しない。旧成功は元wireで閲覧でき、新controlへの移行可否は成功と旧未確定の全範囲を確認して判断する。

## 11. UI・分割・残判断

reload直後は未送信bodyを再構成せずdiscovery。unknown/unavailable/legacy未対応は保留。結果確認後、利用者が次の操作を明示した時だけfence→最新対象取得→入力/preview→新key。成功画面へ戻っただけでfenceや新writeを送らない。fence後に別タブが先にclaimすれば新claimを拒否して回復へ戻る。別staffの同接種発行は既存series/source CASが担当し、本人controlを全staff排他へ拡張しない。

Slice Aはdiscovery＋保留だけで独立実装可能。ただし未知/0件からの新key再開を許可しない。Slice Bはcontrol/closure/coordinator/全writer/FE wireを揃えた単位でのみ安全な明示再開を有効化する。471/472/473/474と486/487のownerを維持し、489は限定claim参加のみ。

現時点で新しい業務判断質問は不要。職員/飼主の新たな権限、外部送信、本文保存期間、失権回避を追加していない。原本PDF保存先・font選定や健診OCRとは別問題。legacyの実状態と既存writer閉鎖証跡は技術受入残であり、一般ユーザーへ根拠なく選択を丸投げしない。

受入は全てNOT_RUN: 遅延claim前後・Apply前後・fence応答断・同世代再送・ack違い・別target同key競合・別tab新key・別staff CAS・失権/再紐付け・sticky unknown・欠損metadata・空集合世代進行・int64上限・legacy bypass拒否・reprint+NULL拒否・issue+非NULL拒否・正reprint許容・controlなしlegacy成功の回復・fence後旧世代成功replay・同key別targetbodyの元認可。採択/文書公開とruntime保証の完成を区別する。

## 固定根拠

- https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/operation/contract.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/operation/runner.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/operation/store.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/work/db-contracts/revisit-intake.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/work/emr/emr-486-certificate-api-contract-20261020.md

## Runtime 再確認根拠（2026-10-21 05:38 UTC）

- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/operation/contract.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/operation/runner.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/operation/store.go
