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

item必須keys = {operation_key,intent,claimed_at,discovery_state,execution_state,receipt_digest,result}。receipt_digestはsucceeded時だけserver計算のimmutable原成功receipt canonical bytesのSHA256 lowerhex64、それ以外はnull。本文/targetをredactする場合もdigestから原receipt本文を返さず、既存認可範囲は変えない。
discovery_state = result | unavailable。
execution_state = succeeded | unresolved | closed_without_effect。
resultは当該domainの既存lookup response又はnullで、既存wireを混ぜない。成功receiptの権威ある読取のみsucceeded。closure証拠がある未成功keyのみclosed_without_effect。lookup不調/metadata不明はunavailable＋unresolved＋result=null。成功が確認できても現在本文を読めないLIFFは既存redacted200 resultのみ、target/answer/版/本文を追加しない。certificateは現在元target viewがなければ非開示でありredacted成功を新設しない。

順序(claimed_at DESC,operation_key DESC)、1page50。cursorはversion付きclosed payload {v:1,upper_claimed_at,upper_key,last_claimed_at,last_key} のbase64url表現で、scopeはrequest/authで毎回再検査。cursorを認可証明にせずPII/主体hashを含めない。不正400。改ざんで自分の過去行のページ区切りが変わっても権限は増えず、後段fenceはこのページ集合を信用しない。completeは当該上限の走査終了だけを表し、新claim不存在や次操作許可ではない。別actorのkey/件数は返さない。

LIFF selfは自分のreceiptだけを探し、未成功legacy metadata不明はunavailableとして残す。target不明を現在予約から推定しない。target queryに現れない旧不明keyがある場合、後述cutover gateが解除されるまで新操作を止める。

## 5. 明示次操作fence API

LIFF:
POST /api/liff/{clinicId}/revisit-intake/operation-recovery/fence
body {target_ref:{appointment_id,pet_id,target_generation},expected_recovery_generation,acknowledged_operations:[{operation_key,observed_execution_state,observed_receipt_digest}]}

証明書:
POST /api/v1/vaccination-certificate-operation-recovery/fence
body {target_ref:{vaccination_id,edition_id},intent:"issue"|"reprint",expected_recovery_generation,acknowledged_operations:[{operation_key,observed_execution_state,observed_receipt_digest}]}

acknowledged_operationsはoperation_key昇順・重複なし。observed_execution_stateはdiscoveryのexecution_state、observed_receipt_digestは同receipt_digestをそのままechoする。succeededはdigest必須、他stateはnull。serverはlock下でstate/digestを再取得し、観測値と違えば409 RECOVERY_CONFLICT、世代・closure・transitionを書かず再discoveryと明示確認を要求する。unresolvedを観測後に成功したkeyを、serverだけの判断でacknowledged_successへ昇格しない。対象generationにclaimされた全keyを、人が結果確認した集合として渡す。移行が完了した新経路では最大1key、controlがまだない場合は空集合。過去世代の全履歴をack対象へ混ぜない。

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
clinic_id,control_id,previous_generation,operation_key,observed_execution_state,observed_receipt_digest nullable, outcome enum acknowledged_success/closed_without_effect。observed state/digestのnull条件は§5と同じCHECK。
PK(clinic,control,previous_generation,operation_key)、transition複合FK。LIFFはauthor_customer_idも持ち同clinic/customer/key receipt FK、certificateはbinding_idを持ち同clinic/binding FK。control principal、元metadata、key/bindingの一致をguard/repositoryで検査。append-only。

空ackの世代進行もheader1行を必須保存し、items0件とする。header hashはSHA256(JCS({schema:"operation-recovery-ack/v2",acknowledged_operations:§5のkey昇順closed配列}))。scope/intent/generationはheader/親controlへ型付きで保存し、hashだけへ依存しない。履歴をstaff/systemのaudit_logsへLIFF customerとして偽装しない。この追加sinkが回復操作監査で、domainの元結果監査は変更しない。

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

現在実行認可preflightの後、control不存在なら一意INSERTでgeneration1/activeNULLを作る（純readは作らない）。同じ業務gate→control→対象binding/receipt（複数移行対象がある場合key昇順）→既存対象・現在実行認可検査の順。controlの世代とtargets集合をserverで再照合し、ackのkey集合と各observed state/digestの完全一致を要求する。保存済みtransitionの同一request replayは保存ack hashで先に識別し、その後の現在state変化を理由に既存transitionを拒否しない。successを権威あるreceiptで確認し、LIFF claimedはreceipt lock下で成功answer/receipt不在が確定する場合のみ閉鎖可能。C06はsticky unknownなし、binding lock取得済み、成功receiptなし、DB-only effect/commit状況が確定している場合だけ閉鎖可能。commit不明やlookup不能をabsence扱いしない。unknownは従来解決ownerへ渡して停止。

同TXでtransition/header/items＋generation+1/activeNULLを保存。どれか失敗すればrollback。fence応答断はtransitionを現在認可付きで再取得し、同じexpected generationの再要求も同transitionへreplayed。遅延claimが先に成功すればfenceはそのkeyをack不足として拒否、遅延Applyが先に成功すればstate/digest不一致でfenceを拒否し、そのreceiptを再discoveryで確認してから明示fenceをやり直す。fenceが先なら旧世代claim/Applyが拒否される。

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

Slice Aのmachine contractはbackend/docs/api.yamlへdiscovery query2本だけ選定収録する。LIFFは既存Engine直下server、証明書は既存/api/v1 serverに対する相対pathを使い、既存lookup responseをdomain別に参照する。API固定・マージ後にmock FE PRへ進めるが、router/BEは未実装のままでありruntime公開や待機解除後の新key再開を意味しない。fence、旧submit/issueのrequired世代、DB・生成model・FE・公開フラグはこのSlice A収録に含めない。

Slice Aはdiscovery＋保留だけで独立実装可能。ただし未知/0件からの新key再開を許可しない。Slice Bはcontrol/closure/coordinator/全writer/FE wireを揃えた単位でのみ安全な明示再開を有効化する。471/472/473/474と486/487のownerを維持し、489は限定claim参加のみ。

現時点で新しい業務判断質問は不要。職員/飼主の新たな権限、外部送信、本文保存期間、失権回避を追加していない。原本PDF保存先・font選定や健診OCRとは別問題。legacyの実状態と既存writer閉鎖証跡は技術受入残であり、一般ユーザーへ根拠なく選択を丸投げしない。

受入は全てNOT_RUN: 遅延claim前後・Apply前後・fence応答断・同世代再送・ack違い・別target同key競合・別tab新key・別staff CAS・失権/再紐付け・sticky unknown・欠損metadata・空集合世代進行・int64上限・legacy bypass拒否・reprint+NULL拒否・issue+非NULL拒否・正reprint許容・controlなしlegacy成功の回復・fence後旧世代成功replay・同key別targetbodyの元認可。採択/文書公開とruntime保証の完成を区別する。

<a id="recovery-query-gap-selected-20261021"></a>
## 12. Slice A discovery 採択技術追補（API正本未収録・runtime未実装）

本節は2026-10-21採択の技術追補である。§4の該当queryの走査終端・cursor詳細は本節で補足し、詳細が重なる場合は本節を優先する。既存2queryのrequest/response wireとedition単位control/fenceは維持し、新readだけを追加選定する。基準mainは `c86c8dd2530bbea3d1810e07badb72f3ed1a07e1`（#803採用）。本人の既決の回復要件を満たすため、履歴別ページの本人reprint発見、走査終了の意味、cursor文法の3点だけを補う。業務要件の判断窓口はMinoru Soga、実装claimとは別。計画指標はこの3点のwire/現在認可/後方互換/残実装の対応が説明できることとし、新しい業務KPI・効果達成は設定しない。

本稿の文書採択・公開は、機能公開・API収録・①の実装完了を意味しない。#799/#800のSlice A採用は維持し、既存のfence availableは両domainともfalse、LIFF activation=false、owners scope=trueを変更しない。Slice B、新キー再開、証明書ルート接続は対象外。試験・生成・runtime実行はすべてNOT_RUN。

### 12.1 接種単位の本人reprint discovery（追加read）

既存§4のedition指定queryを後方互換のまま残し、別の純readを追加選定する。API正本への収録は①の未完了作業であり、今回実APIを追加した記録ではない。

- 実route: `POST /api/v1/vaccination-certificate-operation-recovery/reprints/query`。
- OpenAPI相対path: `/vaccination-certificate-operation-recovery/reprints/query`。既存 `/api/v1` serverに従いprefixを重複しない。
- request exact keys: `{target_ref:{vaccination_id},cursor:null|string}`。両keyとcursorの明示nullは必須。intentはserverで既存 `vaccinations.certificate.reprint` に固定し、edition/clinic/staff/actor/intentのclient指定を受けない。vaccination_idは既存証明書integer/int64、lossless decodeを維持する。
- response exact keys: `{items,next_cursor,complete}`。itemsは最大50件、各要素exact `{target_ref:{vaccination_id,edition_id},operation}`。両ID必須、edition_idは正int64。operationは§4の既存証明書itemと同じclosed shape（`CertificateOperationRecoveryItem`）でintent=`reprint`固定。既存domain lookup resultを再利用し、result_refs/committed_versionsのstring mapを独自の固定key集合へ縮小しない。

item.target_ref.vaccination_idはrequestのvaccination_idと一致、edition_idはimmutable元target metadataに由来し、operation.intentはreprintとする。成功result_refsのvaccination_id/edition_idも同じ原対象への一致を必須とする。その他の既存string map keyを禁止する意味ではない。欠落/不整合を現在値で補完せず、現在認可成立後の結果検証不能は既存unavailable/unresolvedの扱いに従う。現在の接種版と過去editionの保存版との一致は要求せず、履歴内部の整合と現在元target認可を別に検査する。

単一のrecovery_generation/active_operation_keyは返さない。各itemの原editionに対する既存edition指定queryでそのcontrolを読む。接種単位のcontrol/fence/lock/新永続表は追加せず、edition単位のcontrol identity・世代・排他を変更しない。旧queryのrequest/responseに新required fieldを追加しない。RV batch intentや別職員の操作へ対象を広げない。

各requestで認証済clinic/staffを確定し、現在の元接種への `vaccinations:view` を検査する。その本人scopeのC06 bindingとimmutable元target metadataから、成功receiptのない操作も列挙する。元metadataの接種・edition・元actor/intentと実editionの接種/医院関係を照合し、要求接種に属すると証明できたものだけを返す。最新edition限定・UI履歴の取得済page限定・有効PDF限定にしない。歴史editionの発見は旧PDF出力の許可ではなく、contentの現在有効性gateを維持する。成功台帳だけの検索、現要求のeditionへの付替え、現在履歴から未成功の原editionを推定する方式は採用しない。

原targetと現在認可が確定した操作のlookup読取不能は、既存の `discovery_state=unavailable / execution_state=unresolved / result=null / receipt_digest=null` として元edition付きで残す。成功が権威あるlookupで確認できた場合だけsucceededと成功digestを返す。元anchorが欠落/解決不能で要求接種への帰属も現在認可も証明できないlegacy操作は、推定edition/null editionのitem・key・件数を開示しない。その本人scopeの列挙で、そうした未分類操作を除外して完全な検索範囲と証明することもできない場合は、query全体を503 `RESULT_UNAVAILABLE` とし、空集合/complete=trueへ潰さない。これは既存§10のlegacy保留を純readで表現するもので、全editionの排他を新設しない。帰属を証明できた別接種の操作は検索対象外。元対象の現在認可失敗は既存403/404の非開示規約を優先し、certificate用redacted成功を作らない。

順序は全対象editionを通じた `(claimed_at DESC,operation_key DESC)`、1page50、最初のpageで固定した上限tupleをcursorで維持する。pageごとに現在認可を再検査し、別職員のkey/件数は返さない。業務write・claim・修復・render・fenceをreadから開始しない。全応答no-store。queryエラーは既存closed `{error,code}` の `VALIDATION / NOT_AUTHORIZED / SCOPE_NOT_FOUND / RESULT_UNAVAILABLE` と既存HTTP400/401/403/404/503を維持し、fence用 `RECOVERY_*` を追加しない。

### 12.2 completeと走査上限（既存2queryへの明文化）

成功responseで `complete=true` と `next_cursor=null` は同値、すなわち `complete == (next_cursor == null)` とする。complete=falseでは非nullの有効cursorを返す。適用対象は既存LIFF target/self query、証明書edition指定query、および本節の新read。各scopeの固定上限以内の次行がないことをlimit+1等で確認し終えた場合だけ終端とし、空終端pageはitems=[]/next_cursor=null/complete=trueとする。失敗・認可不能・元target未分類を終端に置換せず、途中pageを部分成功として返さない。

上限より後の新claim、走査後の成功状態変化、遅延senderの不在は証明しない。0件/complete=trueは次write・新key・自動再送・fence成功を意味しない。UIの最大page数・時間制限・反復cursor検出で中断した場合も完了へ改変せず保留する。新readの全page取得が済んでも、各editionの既存control/fenceを省略できず、Slice Aでは結果確認と保留だけを維持する。

### 12.3 cursorの文字文法・厳密decode・上限

cursor=nullは初回だけ、文字列ならASCIIで1..2048文字（末尾paddingを含む）。serverはRFC4648のbase64url alphabet `A–Z a–z 0–9 - _` を使い、**必要な末尾 `=` paddingを付けた正規形**を出力する。改行/空白/`+`/`/`は禁止。入力は互換性のため正しいpadding付きとpaddingなしの両方を許容し、次の順で厳密に検査する。

1. decode前に型・ASCII・2048文字上限を検査する。`=` は末尾のみで0..2個。data部分の長さmod4=1は不正。padding付きなら全長mod4=0かつ必要数と完全一致し、不要なpaddingを拒否する。
2. paddingなし入力は内部で必要数だけ補ってstrict base64url decodeする。未使用pad bitsが0でない表現を拒否する。decode後に正規padding付きへ再encodeし、padding付き入力とは完全一致、paddingなし入力とは正規形の末尾 `=` を除いた値との完全一致を要求する。正規形へ直すだけで別bytesの表現を受理しない。
3. decoded bytesはstrict UTF-8 JSON object、重複key/未知key/不正Unicodeを拒否する。v1 exact fieldsは§4どおり `{v:1,upper_claimed_at,upper_key,last_claimed_at,last_key}`。vは整数1だけ、日時はUTC `Z` のcanonical RFC3339Nano、Gregorian年0001..9999・最大小数9桁、UUIDは既存canonical lowercase非nil UUID。last tupleはupper tuple以下。文字列の数値化や日時の省略補完をしない。
4. base64url文字列表現の厳密decodeとclosed typed JSONの検査を分ける。JSONの空白やobject key順だけでは拒否せず、JCS再canonical化やdecoded bytesとのbyte一致を要求しない。cursor内値をscope/認可証明にせず、requestと現在認証から毎回絞る。未知versionも400で推測decodeしない。

v1のproducerは、固定5keyを固定順で一度ずつ出力するcompact deterministic UTF-8 JSON（余分な空白なし、上記ASCII値を不要なUnicode escapeにしない）とする。timestampは各最大30 ASCII bytes、UUIDは各36 bytes、key/区切り/v=1を含む**producer出力**は最大211 bytes。padding付きbase64urlは `4×ceil(211/3)=284文字`（最大bytesの場合、paddingなし282文字）。これはproducer出力上限であり、入力decoded bytesの211byte制限ではない。入力全体はpaddingを含め2048文字、decoded bytesはtransport上限1536 bytesとclosed typed v1形・意味値で制限し、許容JSON空白/key順を211byte根拠で拒否しない。新しいJSON正規化ライブラリの導入を要求せず、将来versionの余白を未知version受理の根拠にしない。この算術は静的サイズ計算であり実試験PASSではない。

既存FEの `/^[A-Za-z0-9_-]+={0,2}$/` とmax2048はtransportの一次検査であり、padding数・pad bits・strict UTF-8・closed typed v1意味値の厳密decode仕様ではない。既存opaque cursor契約を入力JSONの空白/key順によって狭めず、producer出力規則と入力許容範囲を区別する。今回FE/BEのparserは変更しない。

### 12.4 API正本への最小反映範囲と①の受領条件

実ファイル `backend/docs/api.yaml` は未変更。445の単一ownerが今後反映する未完了作業は次だけとする。

| 対象 | 最小反映 | 保持する境界 |
|---|---|---|
| 新path | 上記相対pathへPOST、operationId候補 `queryVaccinationCertificateReprintRecovery` | 既存queryのoperationId/path/required shapeを変更しない。収録時に一意性を確認 |
| 新closed schemas | `CertificateReprintRecoveryQueryRequest`、`CertificateReprintRecoveryTargetRef`、`CertificateReprintRecoveryEntry`、`CertificateReprintRecoveryQueryResponse` | request targetはvaccinationだけ、entry targetはvaccination+edition。IDは既存int64。entry.operationは既存CertificateOperationRecoveryItem参照＋intent=reprint制約。required nullable cursor/next_cursor、items最大50 |
| 既存共通cursor | `OperationRecoveryCursor` にmaxLength2048、transport alphabet/padding patternと厳密decode説明を追補 | nullable/minLength1維持。patternだけで厳密decode済みとしない。v1payloadを公開の自由入力DTOにしない |
| 完了/エラー | 全query responseのcomplete説明に§12.2の同値条件、全queryにno-store。既存query error schemaとHTTPを参照 | 新status/独自ResultResponse/fenceエラーを足さず、certificate resultの既存string mapを維持 |

受領条件は技術追補採択済み、API正本への反映未完、domain列挙/元target認可/serializerの接続未完、FE read adapter未受領を別欄で管理すること。文書採択を①反映完了や②③完成としない。fixture条件（すべてNOT_RUN）は、履歴別pageの本人未成功reprint・複数edition順序・別staff非開示・元target不明legacy・認可成立後のlookup不調・成功digest条件・途中claim/固定上限・空終端・client走査中断・padding有無/誤数/非zero pad bits・不正UTF-8/重複key/未知version・JSON空白/key順許容・producer最大v1サイズ・入力decoded211bytes超の許容・2048/2049境界とする。fixtureのための実行・新キー再開・Slice Bを先行させない。

## 固定根拠

- https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/operation/contract.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/operation/runner.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/backend/internal/operation/store.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/work/db-contracts/revisit-intake.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/edf71515dbae11735a4184ff64c465e2ea51af88/docs/work/emr/emr-486-certificate-api-contract-20261020.md

## Runtime 再確認根拠（2026-10-21 05:38 UTC、当時の記録）

- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/operation/contract.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/operation/runner.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/operation/store.go

## 後発採用の限定確認（#778）

上記§2/§9の「hook未実装」「最新HEAD」は05:38 UTC snapshotの履歴である。#778は2026-10-21 11:48:30 UTC、`eb137881c013f9bd7b0335c01c62ede31429cb23` でmain採用済みであり、winner-only ClaimMetadataWriterの同claim-TX hookは再実装しない。この限定採用だけで回復用BeforeClaim/control/closure/ContextGate/consumer参加が完成したとはしない。生成同期EMR-527の本人実行・成果受領も別残件であり、#778マージを生成完了/全C06対応と読み替えない。
