# EMR-308/468 次回予定の明示消化・同ID確定：採択技術契約

2026-10-22。root技術選定を文書化。初回照合mainは`a5b4a9553b3f003e90ea47660a109c5d526be59c`。2026-10-23のpreview/版/カウンタ/タグ回復追補は`b73e7a920be2a93cf1996f0ff9c6fe3825e22e67`を追加根拠とし、初回基準と分ける。到着済日付/projection/消化source soft-delete追補は`03932bbbdabcbc37306373cad0997638b88c9800`を追加根拠とする。技術選定済みとAPI機械収録・DDL/model・共有Q・domain実装・受入を区別し、後者は未完。今回はdocsのみ、tests/build/codegen/DB/Docker/browserはNOT_RUN。merge・レビュー依頼は行わない。

## 1. 正本・目的・所有

業務判断窓口はMinoru Soga。[明示消化FR/AC](../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-next-intent-explicit-consumption)、[会計完了後の追加承認](../../spec/workflows/reservation-to-record-flow.md#next-intent-after-accounting-consumption-20261022)、[予約技術§4.8](../../design/domains/reservations/technical-spec.md#reservation-visit-history-intent-contract)が業務正本。[480契約§5–10](emr-480-booked-duration-selected-contract-20261021.md)、[458 HTTP契約§6](emr-458-option-assignment-selected-contract-20261021.md)、[C06/Q](../../design/shared/operation-participation.md#verified-command-selected-20261021)を再利用し、同じ操作に第2key/台帳を増やさない。

目的は、スタッフが選んだ元予定と今回予約の関係を一度だけ残し、予約/記録の重複作成・無選択の自動吸収・再送による二重消化を除くこと。既決の明示確認以外の手入力/確認工程、通常カルテや会計の再確定を追加しない。計画指標は下記2操作と全callerの対象・認可・保存・復旧・担当対応があること、未説明差分0。実業務の所要時間・操作数改善は未測定で、品質保証/PASSへ換算しない。自動消化は採用しない。

468 W04/W05が予約writerの単一統合owner、308は明示消化6ACの追跡。API単一owner445が機械収録、446/466は必要保存/model/ERD、489はQ/ClaimMetadataWriter/ContextGate・registry/lookup allowlistの必要subset。480のtrimming participantは同commandに統合する。現在の実装sessionへ無断追加せず、実装着手時の単一claimでhunk/供給SHAを確定する。全domain/全489 Doneを待つgateは設けない。

## 2. 共通境界と操作A/B

以下は実prefix `/api/v1`込み。認証staffのみ、clinic/actorはtrusted context。A/Bともclosed JSONで未知/重複key、不正UTF-8、複数JSON値を拒否。operation_keyは非zero小文字canonical UUID。新ID/版文字列は`1..9223372036854775807`の正規十進文字列（符号/先頭0/空白/指数なし）。既存Bの任意fieldのID wireを一括変更しない。時刻は既存RFC3339の受理精度を保ってUTCへcanonical化し、無断丸めしない。全応答no-store。

| 操作 | method/path・binding intent | 現在認可・対象 |
|---|---|---|
| A 既存今回予約への履歴紐づけ | `POST /api/v1/reservations/{appointment_id}/consume-next-visit-intent`、`reservation.consume_next_visit_intent` | 両対象に現在`reservations:edit`。read/lookupは`reservations:view`、両対象の同clinic/同pet・現在scopeを検査 |
| B 元予定の同ID確定 | 既存`POST /api/v1/reservations/{intent_id}/confirm-tentative`、`reservation.confirm_tentative` | 現在`reservations:edit`と既存予約適格性/警告/患者保護。read/lookupは`reservations:view` |

owner_kind=`reservation`、actor_kind=`staff`。Aは記録/billing本文を変更/公開しないため、そのgrantや本文versionを一律追加しない。会計完了後の許可はAの専用履歴操作だけで、通常編集/取消制限を解除しない。B後の別操作であるcheck-inのparticipant・必要権限を維持する。A/B自体へ受付record writerを追加しない。

### 2.1 A request・状態検証

bodyは下記5key必須、追加optional keyは`unknown_date_confirmation`だけ：

```json
{"operation_key":"<UUID>","expected_version":"<今回予約版>","next_visit_intent":{"id":"<元予定ID>","expected_version":"<元予定版>"},"confirmed":true,"chronology_digest":"<preview lowerhex64>"}
```

pathは今回予約。`next_visit_intent`もexact2keys、confirmedはtrueだけ（false/null/省略拒否）。元予定と今回予約は別ID、同医院・同患者、削除なし。元予定はactive tentativeかつ未消化。今回statusは`pending/confirmed/checked_in/in_consultation/accounting/completed`だけ許可し、`tentative/cancelled/no_show`・削除・未知/nullは拒否する。pending/confirmedを来院済み・実施済みへ偽装せず、Aで状態を進めない。取消済み/未来院（no_show）の予約を再利用しない。

今回予約の業務field/status、カルテ/専用記録の本文・状態、請求/入金は不変。関係追加のため§6の内部next_visit_consumption_revisionだけを+1し、既存storage triggerが今回予約versionを一回増分、eventを保存する。元予定の終端と両対象の履歴が片方だけ成立してはならない。Aを診療/会計の再保存や本文view要求へ拡張しない。

### 2.1a 元診療との時系列・日付不明確認（2026-10-23採択）

元`source_medical_record_id`で示される診療の業務日と、今回予約の来院業務日を既存正本に従うtyped readerで比較する。日付は厳密Gregorian `YYYY-MM-DD`のDate意味値とし、sourceは元診療の業務日、targetは下記の予定日/実来院日branchのresolverを使う。created_at/intended_dateやレジ日界を代用しない。両日既知でtarget < sourceなら拒否、同日は拒否しない。必要なsource identity/clinic/petも再検査し、不正な他患者出典を日付不明として救済しない。元出典なし/正常に確認した欠損日付はunknown、読取障害/破損/未結線はunknown確認で迂回せず保存停止とする。

Aの追加optional wireは`unknown_date_confirmation:{confirmed:true}`（object exact1key）。このkeyの省略は未確認、指定する場合は非nullのobjectとliteral trueだけを許し、null/false/他key/不正型を拒否する。top-level既存`confirmed:true`は対象の消化選択であり、日付不明の了承を兼ねない。日付不明をスタッフへ表示し、別の明示操作で了承した時だけこのobjectを設定する。対象確認だけから自動生成せず、未確認trueを捏造しない。

#### 純read previewのclosed wire

`GET /api/v1/reservations/{appointment_id}/next-visit-intent-consumption-preview?next_visit_intent_id={id}`を選定。staff認証、認証clinicと現在`reservations:view`で両対象scopeを検査し、source診療日は下記の目的限定projection adapterで同医院/同患者を確認する。汎用医療GETのread認可は変更しない。未知/重複queryを拒否、no-store、write/claim/receiptなし。認可失敗・読取障害・未接続・FK不整合・別患者をunknownへ落とさず、既存ErrorResponseで非開示/失敗を返す。200は完全に評価できた次のexact形だけ：

```json
{"clinic_id":"<ID>","appointment":{"id":"<ID>","version":"<版>"},"next_visit_intent":{"id":"<ID>","version":"<版>"},"chronology":{"source_business_date":null,"target_business_date":"2026-10-23","state":"unknown","unknown_reason":"source-missing"},"chronology_digest":"<lowerhex64>"}
```

全key必須、各object closed。ID/版は正規十進文字列、日付は厳密Gregorian YYYY-MM-DD|null。source業務日は`medical_records.date`。targetがpending/confirmedなら`appointments.start_time`の医院JST予定日、checked_in/in_consultation/accounting/completedなら対象予約の受付event→actual_visit_factsを511契約のsource検証/訂正閉包まで解決した`visit_date_jst`を採る。到着済targetでstart_timeを実来院日にしない。読取で新しい実受診履歴を生成しない。created_at/intended_date/現在日/レジ日界への代用なし。

到着済targetの内部readerは、明示appointment IDに属する受付eventからfact sourceを辿り、clinic/pet・source_version・binding/receipt等の511 source verificationと全訂正graphを照合し、一意な有効terminalの日付を採る。petだけで他来院を拾わず、MAX(id)/最新recorded_at/checked_in_atで推測しない。正当なlegacy証拠欠損・一意証明不能はtarget_business_date=null/unknownとし、既存unknown_reasonのtarget-date-missing（sourceも欠損ならboth-dates-missing）へ写像する。取得障害/未結線/scope破損は停止。予約statusだけで来院根拠を捏造せず、訂正済旧factの元日付を残存採用しない。根拠は[511 reader契約§6](emr-511-actual-visit-reader-contract-20261021.md)のsource verification/closureと[actual_visit_facts DDL](https://github.com/MinoruSoga/AnimalEkarte/blob/03932bbbdabcbc37306373cad0997638b88c9800/backend/migrations/001_init.sql#L10058)。

目的限定の予約消化日付projection内部adapterを選定する。GETは`reservations:view`、Applyは`reservations:edit`と両予約の現在clinic/scope、lookup/replayは§5の現在viewを使う。任意record IDをclientから受けず、認可済みsource intentの保存済みsource_medical_record_idのみserverが辿る。同clinic/petを検査し、譲渡時の歴史ownerを改変せず旧owner差だけで出典を排除しない。公開は既選定日付/判定/理由/digestだけで、本文・診断・処置・金銭・メモ・汎用record DTO・fact証拠は返さない。

このadapterは汎用readerの認可無効化ではなく、511のconsumer既存grantによる最小投影と同じく目的を限定した別projection。医療汎用GETのmedical-records:viewは維持し、予約消化へ医療read grantを一律追加しない。新grant/既存roleの権限付与は行わない。adapter/source検証/同患者・訂正閉包と必要依存保護は未実装の受領subsetで、scope不正をunknownへ変換しない。

chronology.stateは`not-before|before|unknown`。両日既知はtarget<sourceだけbefore、同日はnot-before、unknown_reason=null。片方以上が正当に欠損した場合だけunknownとし、理由は`source-missing|source-date-missing|target-date-missing|both-dates-missing`。両日欠損を最優先、次いでsource参照なしならsource-missing、source参照は正しいが日付欠損ならsource-date-missing、targetだけ欠損ならtarget-date-missing。存在するFKの未解決/不整合をsource-missingへ隠さない。本文や氏名を投影せず日付・判定だけ返す。

`chronology_digest = lowerhex(SHA256(UTF8(JCS(projection))))`。projection exactは`{schema_version:"next-visit-consumption-chronology/v1",clinic_id,appointment_id,next_visit_intent_id,source_medical_record_id,source_business_date,target_business_date,state,unknown_reason,target_date_evidence}`。IDは正規文字列、source identityは保存元record ID又は正当に出典なしのnull、日付・判定・理由は上記値。医院/両対象/出典と確認内容を一つに束縛し、新credential/署名鍵は作らない。digestは認可証明ではない。

target_date_evidenceは内部closed `{kind,appointment_status,checkin_source_set_digest,fact_closure_digest,source_verification_digest,adopted_fact_id,adopted_fact_source_version}`。kind=scheduled/actual/unknown、statusは採用した現在status。scheduledでは証拠5fieldをnull、actualでは採用fact ID/source版は正規文字列、集合/訂正閉包/検証結果の3digestはlowerhex64。到着済unknownでは採用ID/版null、正常取得した候補/閉包/検証結果（欠損や一意証明不能も含む）のcanonical digestを持つ。内部digestは511 typed根拠のcanonical投影から生成し、日付同値でも採用fact/訂正閉包/source版検証の変化を検出する。fact/source生証拠を公開responseへ追加しない。

A requestのchronology_digestは必須lowerhex64、unknown_date_confirmationは別の明示了承。両方をContentへ含める。POST Applyは既存lock順とsource ownerの必要日付依存保護（到着済targetの受付fact/訂正ownerを含む）の下で現在認可・同患者・同じprojectionを再取得し、digest不一致はdependency-changedで無write。beforeは了承があっても拒否、unknownは別確認なしなら拒否、not-beforeは別確認不要。確認内容/digestを変えて同key再送すればoperation-key-conflict。GETを跨ぐserver内snapshot保持を仮定せず、client日付を正本にしない。CASは両予約expected_versionで別照合する。

予約versionだけではfact訂正を検知できない。Applyは採用factだけでなく候補source/全訂正閉包を再取得し、訂正writerと共有するownerの依存保護/serializationで新しい訂正edgeのphantomも防ぐ。既存有限row集合のlockだけで閉包安定と扱わない。必要なsource検証・訂正writer参加/依存保護が未受領なら有効化を止め、予約版CASだけで代替しない。

source eventの既存監査JSONへ`chronology_check:{source_business_date,target_business_date,state,unknown_reason,chronology_digest,unknown_date_confirmed:boolean}`を記録し、実評価・Content・明示確認を対応づける。actorはtrusted、本文複写なし。元成功replayを現在日付で再評価/書換えしない。

目的限定日付projection、source同患者照合、到着済targetの511 source検証/訂正閉包・訂正writerのphantom保護subset、Date adapterとFE preview接続は**未受領の利用開始前提**。閉じたread wireを選定したことを実在route/API/adapter完成としない。未結線をunknownで代替しない。

### 2.2 B request・既存入力のpresence

必須keyは`operation_key,expected_version,start_time,end_time`。任意keyは既存`reservation_type_id,doctor_id,is_designated,status,visit_type,notes,warning_receipt`と480条件下の`trimming_selection`だけ。`status`は既存の閉じたenum `pending|confirmed`を保持し、省略時は既存どおりconfirmedとする。他status/nullは拒否し、明示値をContentへ含める。新confirm-and-check-in routeは作らない。

- 任意fieldの省略は保存値保持。reservation_type_idだけは明示nullを省略相当（保存値保持）の旧互換とする。presence-aware decodeでnullを識別した後、Q/hash前にnull→省略へcanonical化する。正IDだけ変更、0/負/不正型拒否、default補完/clear禁止。doctor null解除/notes null拒否とは別で、API説明もこのfield固有例外を収録する。doctor_idは省略保持、明示null解除、正ID指定は同医院/現在適格性を検査し、0を解除sentinelにしない。
- notesは空文字で消去、null拒否。is_designatedはbool、visit_typeはfirst/revisitの既存検証。省略visit_typeを540候補から補完しない。warning_receiptは既存opaque警告証明で、C06成功receiptと別。
- trimming_selectionは480のexact4keys `{course_id,option_ids,expected_booking_revision_id,expected_adoption_digest}`。previewで見せたend_timeを送信し、既存manual overrideを維持。選択非変更の省略保持と、480が要求する作成/選択変更時のpreviewを区別する。コース必須や新料金を追加しない。

元予定の同IDを本予約へ確定する操作であり、Aの別ID関係を作らない。譲渡済tentativeではpet intent/予約row lock下で現在pet ownerを同医院のpet SHARE lock付きreaderから解決し、その導出owner_idを確定予約へ保存する。旧予定ownerへ固定せず、解決失敗は停止する。`source_medical_record_id`と元診療出典は保持し、このserver導出ownerをclientの任意入力やhash前の補完値にしない。休診/時間/能力/容量/重複/現在患者等は480と既存予約契約どおり、clientのstatusやopaque hashを許可の根拠にしない。

Bはstatus=confirmedでも通常カルテ作成hookを呼ばない。現confirm後のbest-effort通常カルテ作成hookは切替時に除去する対象で、Bの初回成功/replay/失敗/unknownのいずれもrecord作成0とする。カルテ新規作成・既存再利用・作成失敗時の原子性は、[受付時の既存記録再利用/失敗復帰](../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-record-reuse)に従う別check-in操作だけの責務。B成功receiptをカルテ作成/受付成功へ読み替えず、B成功を確認してから既存check-inを別の明示操作として行う。B unknown中は後続に進まない。

## 3. immutable command・Content・Preconditions

480のContent.Fields exact4 string keysを再利用：`{schema_version:"1",resource_kind:"reservation",target_ref:<path ID文字列>,payload_sha256:<lowerhex64>}`。digestはUTF-8 JCSのSHA256。path・wrapper.target_ref・command対象IDは一致必須。key/actor/clinic/CAS版/警告証明を業務payloadへ混ぜず、認証scope又はPreconditionsへ分ける。

| 操作 | immutable hash projection | Preconditions |
|---|---|---|
| A | exact `{schema_version:1,intent:"reservation.consume_next_visit_intent",appointment_id,next_visit_intent_id,confirmed:true,chronology_digest,unknown_date_confirmation}`。chronology_digestは必須の確認内容commitmentとしてContentへ含める。追加確認は省略時null/指定時§2.1a object、両IDは正規文字列 | 今回予約のexpected_versionと元予定のexpected_versionを別名・対象ID対応で保持。§2.1aのsource日付依存も既存ownerのlock下で再評価 |
| B | 480 exact `{schema_version:1,intent:"reservation.confirm_tentative",target_ref,booking_changes,trimming_selection}`。booking_changesはBで許された明示業務入力だけ（start/endと任意設定）、明示status（pending/confirmed）を含め、actor/key/CAS等は含めない。trimming_selectionは省略時null、指定時Content3key `{course_id,option_ids,expected_adoption_digest}` | expected_version、expected_booking_revision_id、warning_receipt等の現在状態/警告前提。expected_adoption_digestはContentに残す |

Bのreservation_type_id:nullだけは§2.2の省略canonical化を先に行う。それ以外の省略と明示null/空/falseを区別し、DB現在値をhash前に埋めない。Bの数値ID入力は既存wireから検証後canonical ID文字列へ投影し、時刻は上記UTC、option_idsは480の数値順。projectionのpresenceを保った同じimmutable commandからApply時の保存値mergeを行う。

Q constructorがcanonical bytesをdeep-copyしpure projectorでContent全key一致を検証、claim/fingerprint前とApply直前に再検証する。Applyは`req.VerifiedCommand`だけから復元し、別DTO/pointer captureをしない。Qはrequest-memory、durable command表を新設しない。意味内容/対象/確認digestを変更して同key再送すればoperation-key-conflict。未成功で同じ意味内容のCAS前提だけを再確認した場合に限り同key再試行、成功後は元receipt固定。

## 4. 結果wire・HTTP・復旧

成功succeeded/replayedはHTTP200、exact required `{status,operation_key,result_refs,committed_versions,recorded_at}`。refs/versionsはstring mapでarrayを直置きせず、ID/版は正規十進文字列。recorded_atは元成功時刻。後日のcurrent状態や新版へ差し替えない。

| branch | result_refs exact | committed_versions exact |
|---|---|---|
| A | `appointment_id,appointment_change_event_id,next_visit_intent_id,next_visit_intent_change_event_id` | `appointment,next_visit_intent` |
| B 今回新T3を生成しない（非trimmingを含む） | `appointment_id,appointment_change_event_id` | `appointment` |
| B 今回のoperationで新T3を生成した場合のみ | `appointment_id,appointment_change_event_id,booking_revision_id` | `appointment,booking`（480同義版） |

既存trimming選択非変更で新T3を書かないBは予約refs/versionだけを返す。旧T3を今回operation/eventの成果として返さず、旧T3のoperation/eventを書き換えない。既存採用状態は別の現在view付きreadから取得する。時刻変更だけで新T3を必須とせず、480の既存採用条件に従い、無条件に新規採用しない。非成功exactは`{status,operation_key}`のみ。内部EffectSummaryをwireへ追加しない。現在本文は別の現在view付きGETで取得する。

[458§6](emr-458-option-assignment-selected-contract-20261021.md)の既存12statusを固定再利用：

| status | HTTP |
|---|---|
| succeeded / replayed / not-succeeded | 200 |
| validation | 400 |
| version-conflict / dependency-changed / edit-right-lost / operation-key-conflict | 409 |
| operation-in-progress | 202 |
| scope-not-found | 404 |
| not-authorized | 403 |
| result-unavailable | 503 |

strict入力/未認証/route scope拒否は既存ErrorResponseの400/401/403/404。独自status追加なし。unknown/busy/transport障害は同key照会、新key/別予約による救済なし。not-succeededは成功でも無条件再実行許可でもない。validation/版/依存拒否後は入力を保持して現在認可・対象・差分/CASを再確認する。Retry-Afterがあれば尊重する。

C06 Lookupの結果store障害は、現在認可後の既存generic error（例500）と、保存済みsticky unknownのresult-unavailableを区別する。503へ機械的一括変換しない。失敗/監査/event/receipt不整合は全rollbackし、予定だけ先に消化済みにしない。

## 5. 元target anchor・lookup

480既選定の`GET /api/v1/reservations/operations/{operationKey}?intent=...`へA/Bの完全intent名をallowlist追加する。staffだけ、認証clinic/同staff/intent/keyで検索、no-store。元anchorからだけ解決して現在`reservations:view`とscopeを再検査し、Aは両対象を検査する。成功replay/lookupはactive FindByIDへ依存せず、stored anchor/receiptと認可付き限定unscoped両対象readで再検査し、消化sourceのsoft-deleteだけで404にしない。typed履歴readのunscopedもclinic/対象ID/現在認可を必須にし、scopeなし全件readや復元を許可しない。request/URL/成功refsによる元対象の捏造fallbackなし。未知keyだけscope-not-found/404、元anchor欠落/identity不一致/認可不能は480/C06のnot-authorized/403。別actor/key/医院の存在を漏らさない。

`reservation_operation_targets`は480で採択した設計であり、このmainの実装済み表ではない。同じownerで必要差分を受領し第二表を作らない。PK `(clinic_id,binding_id)`、FK `(binding_id,clinic_id)` → `operation_bindings(id,clinic_id)`、immutable・clinic RLS、staffのline_subject_hashはNULLを維持。既存appointment targetのappointment_idにAの今回予約/Bの元予定を保存し、`next_visit_intent_id bigint NULL CHECK (>0)`を追加する。Aでは正ID必須かつ今回と異なるID、他intentではNULL。binding intentとこの有限shapeをclaim factory/整合guardで照合する。

claim winnerのみ`ClaimMetadataWriter`でbindingとanchorを同じ短TXに保存し、失敗は両rollback。loserは保存済みanchorを再利用し、incomingで補修しない。業務appointments FK/row lockを短claimへ持ち込まない。preflightとApplyで実targetのclinic/pet/identityを再検査する。成功replay/unknown/lookupは現在view、未成功write再試行はさらに現在edit/CASを検査。共通Q/resolver未受領のまま稼働可能としない。

## 6. Aの保存・event・最小DDL設計（適用しない）

新status/consume台帳は作らない。元予定は消化専用の単一UPDATEで既存`cancelled`へ終端し同時にsoft-deleteし、source側eventだけにtyped `consumed_by_appointment_id`を持たせる。通常Cancel/Clear関数は呼ばず、理由を消化として表示する。通常取消通知/復元UI/outbox/履歴readはtyped関係で区別し、次回予定通知候補と配送直前の再照合で消化済みを除外する。cancelledという文字列だけで通常取消/消化を判定しない。`source_medical_record_id`は出典用途のまま保持する。

Aは元予定のstatus/deleted_at/version終端、今回予約の内部カウンタ変更に伴うstorage version増分/event、両event・必要監査・成功receiptを同最外TXでcommitする。今回予約の業務field/status、記録本文/状態、請求/入金は不変。版増分は既存guardと一回だけ対応させ、アプリ/triggerで二重加算しない。

必要保存差分は以下を446/466へ渡す。新migration番号・model物理file名はここで予約しない。

- sourceは期待version・status=tentative・deleted_at IS NULL付きの消化専用**単一UPDATE**でstatus=cancelledとdeleted_atを同時設定する。共有storage guardがversionを一度だけ増分し、RETURNING版をsource event/receiptへ固定、target counter/eventと原子的に保存する。既存Clearのstatus UPDATE→Delete二段処理を流用しない。ID/intended_date/source_medical_record_idを保持、物理DELETEなし。通常active一覧・preview・新consumeは削除済sourceを除外する。
- `appointments.next_visit_consumption_revision bigint NOT NULL DEFAULT 0 CHECK (next_visit_consumption_revision>=0)`を追加する必要DDL設計。A targetでは期待appointments.version CASの下でこの内部counterだけ+1する。既存entity_version_guardが変更を検知してappointments.versionを一度増分するため、アプリからversionを加算しない。sourceはstatus/deleted_atの単一変更で既存triggerが一度だけ増分する。FEへ新必須CASを増やさない。overflowは無write失敗。
- target eventのbefore/afterへ内部counterを記録。専用deferred guardをcounter更新/対応event追加へ適用し、old+1、同operationのsource消化関係・target event・binding/actor/receipt refs/版との対応を検査する。単独counter更新・飛び値・減少・無関係操作を拒否し、targetのそれ以外の業務列/statusは不変。共有entity_version_guardの例外を追加せず、後日の現在版と過去eventを恒久一致させない。
- `appointment_change_events.consumed_by_appointment_id bigint NULL`。FK `(consumed_by_appointment_id,clinic_id)` → `appointments(id,clinic_id) ON DELETE RESTRICT`、自己参照（appointment_idと同値）禁止。source eventの`intent='reservation.next_visit_intent_consumed'` iff relation非NULL、他eventはNULL。partial UNIQUE `(clinic_id,appointment_id) WHERE consumed_by_appointment_id IS NOT NULL`で元予定の二重消化を禁止。
- 複数のsourceを同じtargetへ紐づけることは許可する。target側`consumed_by_appointment_id`へのUNIQUEは追加しない。各sourceごとに別の明示確認・独立operation/receipt・source/target eventを保存し、target期待版を毎回CASして関係追加版を一回増分、過去関係を全件保持する。同sourceの同key再送は元receipt固定でtarget版を再増分しない。既存single-active次回予定制約を維持し、複数active作成や混合batchへ拡張しない。
- Aのbinding intent=`reservation.consume_next_visit_intent`。source eventは上記`reservation.next_visit_intent_consumed`、target eventは`reservation.consume_next_visit_intent`、同operation/clinic/actorで元/先のID・確定versionとA receipt refs/versionsへ一意対応する。actorは認証staffのみ、body由来にしない。
- 新consume deferred guardはこのtyped source/target eventを追加する時点で、binding actor/clinic/intent、原anchor両ID、追加時点のsource cancelled/deleted_at非NULL・target業務値/status不変、§2.1aの日付判定/確認とContent・監査の対応、両eventのbefore/afterと確定版、成功receiptの4refs/2versionsを検査する。孤立source/target・不一致・receipt欠落をcommitできない。既存のreceipt存在guardだけではこの対応検査を満たさない。
- 追加操作時の対応を検査し、後日の別予約更新で過去event版=現在appointments.versionを恒久要求しない。旧eventを補完/backfillせずimmutable/RLS/元FKを保持する。
- consumed sourceだけはrestore/setによる再活性化を禁止。soft-deleteを理由に原履歴/receiptを消さず、§5の認可付き限定unscoped readで参照する。別の明示setは現在active予定を再取得/CASし、なければ新tentative IDを作る。過去consume/eventを削除しない。通常の誤取消復元まで禁止しない。

## 7. lock・全caller閉包・B切替

取得順は既存467のgraph gate（返されたctx伝播） → C06 binding NOWAIT/元結果再読取 → 必要staff-global ID昇順 → clinic-booking ID昇順 → pet-intent → lifecycle/予約ID昇順row lockと正本§6.4の既存owner規約。Bの時刻/区分/担当等の予約条件変更はstaff-global → clinic-booking → pet-intentを省略/逆転せず取得し、その後に条件・容量・適格性を再検査する。Aは両予約を一度ずつlockし、同clinic/pet・元tentative/未消化・今回status・期待版・現在認可を再検査。Bのtrimming採用は480の予約保護 → T2 → 区分/course/rule/options順と同TX participantを維持する。claim TXにこのlockを移さない。

| caller/writer | 受領すべき閉包 |
|---|---|
| Create/CreateBatch/admin/walk-in/LIFF/trimming内部createOrPromote/カルテhook | no-consume。通常作成だけで元tentativeを検索吸収しない |
| 明示confirm | Bだけへ統合。同ID昇格、閉じたstatus入力を保持し無key/無版callerは切替対象 |
| 既存今回予約への明示紐づけ | Aだけへ統合。別ID関係と履歴を原子的に保存 |
| set/clear/cancel/restore/patient reassignment/直接repo | active予定/consumed sourceと両対象の版・scopeを再検査。別患者付替え・通常restoreで関係を壊さない |
| 通知候補read/outbox配送直前read/履歴UI | typed消化関係で除外/消化表示。旧cancelledだけの復元/通知へ流さない |

Aの消化/Bの同ID昇格が初回成功commitした後にだけ、影響する旧/現在ownerの`next_visit_*` tagを既存`syncNextVisitTagsBestEffort`と同責務で再計算する。C06 replayから再実行せず、外部同期を業務TXへ持ち込まない。tag同期失敗をcommit済みreceiptの失敗や再消化条件へ変換しない。

タグ同期の既存回復根拠（main b73e7a9）：[上位next_visit_tag_sync.go](https://github.com/MinoruSoga/AnimalEkarte/blob/b73e7a920be2a93cf1996f0ff9c6fe3825e22e67/backend/internal/reservation/next_visit_tag_sync.go)は全owner batchに30秒のbest-effort期限、重複owner除外、失敗warnを適用する。[下位SyncNextVisitTag L91–119](https://github.com/MinoruSoga/AnimalEkarte/blob/b73e7a920be2a93cf1996f0ff9c6fe3825e22e67/backend/internal/lstep/lstep_tag_sync_visit_next.go#L91)はprovider副作用前にdurable EnqueueNextVisitSync、成功時は同revisionをackする。[expiry batch](https://github.com/MinoruSoga/AnimalEkarte/blob/b73e7a920be2a93cf1996f0ff9c6fe3825e22e67/backend/internal/lstep/lstep_batch_expiry.go)は各run内でcursorを進めるが、失敗をackせずpending queue等の候補に残し、次回再走査を先頭から行って失敗ownerを恒久的に飛ばさない。既存queue到達後の回復と、予約commit→hook到達/queue投入までのcrash又はqueue投入失敗を分け、後者を完全回復保証しない。未記録配送intentは既存マーケティング回復subsetの残件であり、C06 replayから再送/consume再実行しない。新outbox実装/外部exactly-once保証を本稿で発明しない。

Aが照合する予約status/pet/clinic等を変更するwriterだけ予約版gateに参加させる。金銭だけのwriterへ予約版加算やbilling CASを新要求しない。全domainをgraph gateへ拡張しない。

認可付き既存予約list/detail/tentative選択DTOへ`version:string`を供給する必要subsetを含める。model.Reservation.Versionをcanonical decimalへ投影し、updated_at推定/FE自作版は禁止。他のID型は全面変更しない。Aはpreviewの両対象版を保持、B modalは現在detail再取得のid/versionを保持して送る。API ReservationAppointment・BE response・FE型/cache/変換へ同時伝搬し、必要版がreadできる前にwrite cutoverしない。現response/modelの存在だけで版供給済みとしない。

APIと共有型を先に非活性で収録し、新client/key/CAS/presence/480 previewへ切替。旧key/版をserverが捏造して新command成功にしない。切替scopeの無保護writeは既存契約更新拒否/ErrorResponseで停止し、旧自動吸収へfallthroughしない。B旧ReservationAppointment responseは新C06成功shapeへ移す対象であり、旧callerの同時切替/GET再取得を受領条件にする。API/DDL/model・Q/anchor・全必要writer・通知閉包が揃わないまま有効化しない。

## 8. 固定fixtureと提出単位

実装票の最小単位はA保存/guard/lookup/caller閉包と、B同ID確定/480participant/受付後続境界。共有API/保存/Q hunkはownerへ集約し、独自stub/二重型/第二receiptで先回りしない。各PRはbase/head/採用SHA、供給file/symbol、consumer受領、AC→handler/request→service→repository/model→receipt/lookup→FEを記録する。

fixture作成と実行を分け、以下は全てNOT_RUN：

- A：許可6statusと拒否status/null/削除、pending/confirmed非来院表示、同ID/別pet/別院拒否、未選択無変更、会計後も今回状態/診療/請求/入金不変。
- A：元診療より前/同日/後、純read preview正常/同日/before/各unknown_reasonと認可・障害非unknown、必須digestの対象/日付変化拒否、source又はtarget日付不明（未確認拒否/明示確認可）、到着済targetの予定日とactual日差・recheck-in/訂正閉包・同pet別来院非採用・一意不明/legacy欠損unknown・phantom訂正のdigest不一致、目的限定grantの予約view/edit・汎用医療view未付与でも許可範囲だけ投影・異患者/医院非unknown・歴史owner保持、created_at/intended_date代用禁止、reader障害/未結線停止、日付根拠変更再確認、同来院への複数source独立履歴/target逐次CAS・counter+1/storage版一回増分・単独/飛び値counter拒否・同source replay非増分、source単一UPDATEの版一回増分/active除外・soft-delete後replay/lookup保持・scopeなしunscoped拒否、両対象古版、患者付替え/取消/消化競合、同source別key二重消化、event片側/actor/refs/版不一致rollback、消化source復元拒否と通常誤取消復元保持、別setの新ID/active競合、通知候補と配送直前除外。
- B：reservation_type_id null→省略canonical同Content/正ID変更/0等拒否、read versionのlist/detail/tentative→FE往復・古版拒否、省略保持/doctor null解除/notes空とnull、status省略confirmed/明示pending/confirmed保持・他status/null拒否とContent一致、譲渡後pet owner解決/旧診療出典保持、visit_type保存値保持、trimming4key/digest/preview end/manual override、旧無key/無版拒否、後続check-in明示/unknown停止、Bの正常/再送/失敗/unknownすべてでrecord作成0。カルテ作成失敗/再利用のfixtureは別check-in操作へ属し、Bに保持しない。
- 共通：strict decode/非正規ID/UUID、同key同内容replay・異内容conflict・CASだけ再確認、sticky unknown/busy/失権/他actor lookup、失敗binding anchor・claim loser・metadata失敗rollback、監査/receipt失敗、後日更新後の元receipt読取、staff-global/clinic-booking/pet-intent順、初回commit後だけ旧/現owner tag再計算・replay非再実行、durable enqueue前失敗/同revision ack/expiry失敗owner再走査と保証外crash区間。

本稿は技術選定の公開で、API YAML/DDL/model/生成物/製品コード/実DBは未変更。実装・全writer受入・実機試験の完了を意味しない。

## 9. 固定mainの静的差分根拠

| 現在path | 観測と受領差分 |
|---|---|
| [reservation_intent_handler.go L33](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/reservation/reservation_intent_handler.go#L33)・同L172、[service L142](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/reservation/reservation_intent_service.go#L142) | 現requestはstatusを受け、operation_key/CASなし。confirmed既定とpost-commit通常カルテ作成がある。Bではこの作成hookを除去し、strict入力・receiptへ切替。record作成/再利用は別check-inの同TX責務 |
| [reservation_intent_repository.go L1079](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/reservation/reservation_intent_repository.go#L1079) | pet intent lock→予約row、tentative状態述語と由来保持は既存成果。旧createOrPromote/自動吸収の全caller閉包を代替しない |
| [storage_reservations.go L24](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/model/storage_reservations.go#L24)・[001 L8924](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/migrations/001_init.sql#L8924) | eventに版/intent/operation/actorは存在、consumed_by列/対応guardは未収録。anchor表/A routeもこのmainに存在せず、設計だけで実装済みにしない |
| [api.yaml L13781](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/docs/api.yaml#L13781)・同L17777 | 旧Bはstart/end必須・任意statusとReservationAppointment成功応答。新A/B exact/C06 wireの機械収録とcaller切替は後続 |

追補根拠main b73e7a9：[entity_version_guard L10222–10228](https://github.com/MinoruSoga/AnimalEkarte/blob/b73e7a920be2a93cf1996f0ff9c6fe3825e22e67/backend/migrations/001_init.sql#L10222)と[appointments trigger L10578](https://github.com/MinoruSoga/AnimalEkarte/blob/b73e7a920be2a93cf1996f0ff9c6fe3825e22e67/backend/migrations/001_init.sql#L10578)はversion-only更新を許可しない。[reservation_response.go](https://github.com/MinoruSoga/AnimalEkarte/blob/b73e7a920be2a93cf1996f0ff9c6fe3825e22e67/backend/internal/reservation/reservation_response.go)とAPI ReservationAppointmentへのversion供給、preview route/digest、内部counter/guardは今回選定した後続受領で、実装していない。[expiry候補query](https://github.com/MinoruSoga/AnimalEkarte/blob/b73e7a920be2a93cf1996f0ff9c6fe3825e22e67/backend/internal/reservation/reservation_tentative_expiry_query.go)は期限切れcacheとdurable retry候補を合流する既存成果。

03932bbb追補の根拠：[通常予約query](https://github.com/MinoruSoga/AnimalEkarte/blob/03932bbbdabcbc37306373cad0997638b88c9800/backend/internal/reservation/reservation_repository_queries.go#L83)はsoft-delete除外、[旧Clear L1052–1073](https://github.com/MinoruSoga/AnimalEkarte/blob/03932bbbdabcbc37306373cad0997638b88c9800/backend/internal/reservation/reservation_intent_repository.go#L1052)はstatus更新とDeleteの二段。本契約のsource消化はこれを流用せず単一UPDATE/RETURNING版とする。511根拠projection/訂正閉包・目的限定adapter・限定unscoped履歴readは必要実装subsetで、ここでは実装していない。

これらは静的読解で、build失敗や実DB競合を実行観測した記録ではない。
