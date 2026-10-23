# 診療録所属の既存バイタル：死亡後の生前事実訂正契約

| 項目 | 内容 |
|---|---|
| 文書ID | design.medical-records.posthumous-vital-correction |
| 技術選定 | 2026-10-24 JST / 2026-10-23 UTC、root・独立レビューを経た既存vital訂正の限定契約 |
| 基準main | `7dad9adc7f066fab5cc734dc8ff7e125a80d67d4` |
| 状態 | 技術契約選定済み。共通gate・writer適合・保存/API供給・実装・activationは未完。製品runtime Readyではない |
| 検証 | tests/build/codegen/実DB/Docker/browserはNOT_RUN。以下は実装に渡す契約と期待値で、実行証拠ではない |

[D03業務正本](../../../spec/domains/medical-records/lifecycle.md#medical-record-deceased-retrospective)と[作業計画](../../../work/emr/posthumous-contract-plan-20261023.md#posthumous-plan-vital-packet)に接続する。既存FR/AC・医師資格・一般編集権限は拡張しない。新しい業務概念は追加せず、既存事実・確認・訂正履歴・operationの技術接続を定める。

<a id="posthumous-vital-scope"></a>
## 1. 有限範囲

「既存vital」は対象データ条件であり、既存コード/API/DTOを保存する要件ではない。確定仕様を優先して必要な実装とcallerを置換する。以下の既存GET/通常writer不変更は今回の専用訂正という成果境界を指し、旧wire維持や将来の改修禁止ではない。C06・現在認可・履歴/不明値保全等の選定済み安全契約は別要件として維持する。

現在医院の既存・未削除medical recordに所属する既存・未削除vitalだけを対象とする。`medical_record_id`はpathと一致し、`daily_record_id=NULL`。患者・recordの付替えはしない。最初の供給は**同一の認証済み適格医師が入力actorと確認者を兼ねる明示confirm-and-correct操作**である。他職員の準備と後の医師確認は別の有限packetに残し、通常編集を医師限定にするものではない。

元の保存`recorded_at`と訂正後`recorded_at`の両方が、現在解決した`deceased_at`より厳密に前の既知の実測時点でなければならない。NOT NULL/default now()だけで正確な実測時点とはしない。認証済み医師の明示操作はsource版/digestへ束縛し、元・訂正後時点が当該既存事実の実測時点であることを確認する。確認できなければ送信せず、serverも確認欠落・順序不成立を拒否する。由来/時刻が不確かな旧事実やchildなし後日入力を永久禁止したのではなく、後続の根拠packetへ保留する。

既存の数値/値検証だけを使い、臨床閾値・安全値・薬量判断を追加しない。以前の値はNULLを含め保持する。会計・在庫・薬量再計算、受診事実生成、診療開始/終了、予約取消、通知、死亡状態変更は行わない。通常PATCHのmagic flagではなく専用actionを使い、495 C0/C1/C2の3資源へvitalを追加しない。

専用の最小不変claim anchorと成功時の完全不変訂正履歴の2表を選定する。必須audit_logsとC06を再利用し、第二operation結果機構や第三の確認receiptを作らない。短claimに臨床request/command本文を保存しない。

<a id="posthumous-vital-wire"></a>
## 2. wireと型

以下の相対routeは既存`/api/v1`配下。既存HttpOnly Cookie認証と`X-Requested-With: XMLHttpRequest`を維持する。clinic/actorはtrusted server contextだけから得る。新bearer、Idempotency-Key header、wrapper、client actor/doctor IDは追加しない。

ID/版は正のcanonical十進文字列、schema_versionは文字列`"1"`。UUIDはlowercase hyphenatedの非zero、digestはlowercase64hex SHA-256。時刻は明示offsetを持つRFC3339 instantをUTC・PostgreSQL microsecond精度へ正規化する。より細かな非zero桁は黙って切捨てず拒否し、同じ対応instantの異なるoffset表記は同一化する。duplicate/unknown JSON key、不正Unicode、埋込NULを拒否する。

この新endpointの臨床数値は末端までcanonical **文字列**。temperature/weightはplain decimalで、負号は非zeroだけ、先頭0・指数・plus・末尾小数0・不要小数点を禁止する。zeroは`"0"`だけでnegative zeroを拒否する。heart/respirationは同じ符号/先頭0規則の整数文字列で小数不可。JSON数値tokenや非canonical文字列は黙って正規化せず拒否する。float64を介さずexactにtargetへ変換し、非有限/表現不能値を拒否する。temperature/weightは専用adapterからmodel.ExactNumeric、heart/respirationは既存target整数型の範囲。既存weight正値検証と`Kg|g`を維持し、新閾値なし。既存GET/通常writer全体の数値wire変更は今回の限定成果には含めない。後方互換のために変更を禁止するものではなく、必要な新仕様への置換はcallerを含む完結単位で行う。textはUTF-8検証後そのまま保持し、reasonは非空白文字を要求するが書換えない。

<a id="posthumous-vital-context"></a>
### 2.1 context read

`GET /medical-records/{record_id}/vitals/{vital_id}/retrospective-correction-context`

200のexact keysは次の論理shape（型記法はJSON実例ではない）。

```text
{schema_version,medical_record_id,vital_id,pet_id,source_version,
 source_snapshot_sha256,source_snapshot:SNAPSHOT,
 death_basis:{basic_revision_id,subject_version,deceased_at},
 eligibility:"confirmation-required"}
```

SNAPSHOTは全key必須の次のshape。

```text
{id,clinic_id,medical_record_id,daily_record_id,pet_id,version,recorded_at,
 staff_id,temperature,heart_rate,respiration_rate,weight,weight_unit,notes,
 created_at,updated_at,deleted_at}
```

ID/版は十進文字列。daily_record_id/deleted_atはこのsubsetではnull。staff_id/temperature/heart_rate/respiration_rate/weight/weight_unitは実際のNULLを保持し、数値は上記canonical文字列又はnull。notesはstring、時刻はcanonical UTC instant。現在staff/pet/masterのnested projectionは入れない。このcontext/historyの選定表現であり、通常vital GET全体は今回範囲外とする。旧GET維持や将来の改修禁止を要求しない。

source_snapshot_sha256はSNAPSHOTのUTF-8 JCS bytesのSHA256でidentity/版も含む。現在record/vital view認可を使い、短いread TXでlocal-pet/common/record/vitalの順に整合readし、record/vitalはSHARE lockとする。edit権取得、版変更、proposal/receipt保存はしない。hashは比較値であり認可・独立署名ではない。contextは記録時刻の史実性を保証せず、eligibilityはconfirmation-requiredのまま。

本有限契約はエラーDTO `{error:string}`を選定し（既存コード互換を理由とする固定ではない）、不存在/他院/target不一致/削除は404、現在view拒否403、既存record subset外又は既知死亡instantなし409、必要life未解決/read失敗503。不明を生存/適格/空snapshotへ落とさない。

<a id="posthumous-vital-correct"></a>
### 2.2 confirm and correct

`POST /medical-records/{record_id}/vitals/{vital_id}/retrospective-corrections`

必須top-level exact keys:
`operation_key,edit_session_ref,edit_epoch,owner_expected_version,source_version,source_snapshot_sha256,source_recorded_at,death_basic_revision_id,death_subject_version,death_deceased_at,reason,confirmation`。

任意の臨床変更keyは`recorded_at,staff_id,temperature,heart_rate,respiration_rate,weight,weight_unit,notes`だけ。
confirmationはliteral `"confirm-existing-prior-life-fact-and-correction-v1"`。これは行う明示操作の名前で、単独proof/bearerではない。confirmed boolean、自由attestation、doctor_id等で検査を迂回しない。認証された現在適格医師がUIに示された正確なsource/差分/reason/death basisを送信し、serverは成功した認証済み行為を効果と同TX保存する。通常PATCH/autosaveはこの操作を生成しない。

臨床変更fieldは少なくとも1つあり実差を生むこと。同値はvalidationで、確認履歴/版を増やさない。省略はlock済み元値（NULLも含む）を保持する。**全body fieldで明示JSON nullを拒否**し、初回subsetにclear操作はない。notes=`""`は空文字として有効でNULLではない。nullable clearは後続presence-aware拡張とする。

owner_expected_versionはvital.version、source_versionは実際に確認した元vital版であり、子集合/record版ではない。このin-place訂正では両者とlock下current vital.versionの一致が必須。元版/digest変更は新Contentだが、同sourceのedit fence更新だけはPrecondition変更である。

source_recorded_atは保存元snapshotと一致し、訂正後は指定recorded_at又は元値。両方がdeath_deceased_atより厳密に前。同時刻/不明date-only/捏造時点は不可。現在lifeはdeceasedで、nonnull死亡instant・確認したbasic revision/subject版に完全一致すること。half-dead OR分類でも死亡instantなしではこのsliceの順序を証明できない。根拠不足は後続へ保留し、死亡変更/復活で救済しない。

staff_idは歴史上の記録/測定staffでactor/確認者ではない。変更するnonnull IDだけ既存active参照/同院検証を行い、未変更の歴史ID/NULLを保持する。旧staffの現在activeを一律要求したり現在actorへ置換しない。

<a id="posthumous-vital-results"></a>
### 2.3 成功・失敗

採択済みC06のoperation.ResultResponseを直接返す。これは共通operation契約の再利用であり、旧vital DTOとの互換維持ではない。
`{status,operation_key,result_refs?,committed_versions?,recorded_at?}`。
成功/replayのrefs exactは`{medical_record_id,vital_id,correction_history_id}`、versions exactは`{vital}`、map値は全てstring。recorded_atはserver成功履歴時刻。臨床本文・資格情報・fingerprint・EffectSummary・commandはreceiptへ返さない。

| 結果 | HTTP |
|---|---|
| succeeded / replayed | 200 |
| validation | 400 |
| version-conflict / dependency-changed / edit-right-lost / operation-key-conflict | 409 |
| operation-in-progress | 202 |
| scope-not-found / not-authorized | 404 / 403 |
| result-unavailable | 503 |
| not-succeeded | 200（成功ではない） |

C06前の不正JSON/認証/route target欠落は上記選定エラーDTOと400/401/403/404。元版/digest不一致はversion-conflict、確認life/binding変更はdependency-changed、holder/session/epoch喪失はedit-right-lost、現在医師適格性/edit拒否はnot-authorized、不明lifeはresult-unavailableで成功なし。同keyでreason/source/変更値が変わればkey conflict。serialization/保存失敗は業務TX rollbackと既存C06 retryable/unknown規則を使い、新statusを追加しない。

<a id="posthumous-vital-command"></a>
## 3. command・claimと元target照会

scopeはowner_kind=`medical_record`、intent=`medical_record.vital_retrospective_correct`、actor_kind=`staff`、actor_ref=認証staff十進ID、clinic_id=trusted clinic。bodyのoperation_keyはcanonical UUIDで、495 correction_confirm intentを使わない。

Canonical REQUEST exact keys:
`{schema_version,medical_record_id,vital_id,source_version,source_snapshot_sha256,source_recorded_at,death_basic_revision_id,death_subject_version,death_deceased_at,reason,confirmation,changes}`。
record/vital IDはpathから得る。changesは指定臨床fieldだけで、省略値を挿入しない。JCS順で上記数値/時刻/text/unitを保持する。operation key/actor/edit session/epoch/owner_expected_versionはREQUESTに入れずscope/Preconditionsへ分離する。

Canonical COMMAND exact keys:
`{schema_version,request,pet_id,life_identity,confirmed_by}`。
confirmed_byは認証staff。life_identity exactは`{local_pet_id,clinic_binding_id,common_pet_id}`で初回targetと不変bindingから解決する。common UUIDはprivateで、新public identity DTOにしない。request.source_snapshot_sha256に束縛されない可変before値をcaptureしない。full snapshotはApplyのlock下で再読/digest照合する。

REQUEST/COMMANDをfloat損失なしのUTF-8 JCS化し各SHA256を取る。これらowner-private比較hashは第二fingerprint/namespaceではない。業務意味の束縛は既存C06 Contentのみ。Content.Fields exactは`schema_version="1",command_sha256,target_kind="record",target_ref=record ID,intent="medical_record.vital_retrospective_correct"`。NewContentとNewVerifiedCommand Qを使い、ApplyはJCS/hash/scope/target/intent/actor/元target一致を検査したQだけからtyped changesを復元する。別capture payloadを読まない。

Preconditions exactは`edit_session_ref,edit_epoch,owner_expected_version`。現在CASで確認済み元版/death basisを書換えない。epochだけの更新はretry前提だがsource意味又はlife subject版変更は新明示確認/Contentを要し、元outcome確定後にだけ新key。unknown/in-progressの新key迂回なし。

<a id="posthumous-vital-claim-storage"></a>
### 3.1 medical_record_vital_correction_targets（選定保存差分、未供給）

| 列/制約 | 選定内容 |
|---|---|
| clinic_id, binding_id | bigint NOT NULL、PK(clinic_id,binding_id) |
| medical_record_id, vital_id, pet_id | bigint NOT NULL、各CHECK >0 |
| source_version / source_snapshot_sha256 | integer NOT NULL CHECK >0 / char(64) NOT NULL lowercase64hex |
| clinic_binding_id / common_pet_id | bigint NOT NULL CHECK >0 / uuid NOT NULL非zero。local_pet_idはpet_idと同じで重複列なし |
| schema_version | smallint NOT NULL CHECK=1 |
| request_sha256, command_sha256 | char(64) NOT NULL lowercase64hex |
| recorded_at | timestamptz NOT NULL DEFAULT clock_timestamp() |
| 唯一のFK | (binding_id,clinic_id)→operation_bindings(id,clinic_id) ON DELETE RESTRICT |
| index | (clinic_id,medical_record_id,vital_id,binding_id) |

record/vital/pet/clinic/staff/subjectへのbusiness FKは置かず、Gate前の短claimが業務rowを先取りしない。既存has_clinic_accessのtenant RLS、新grantなし。INSERT後immutableで、既存contract-history方式でUPDATE/DELETEを拒否。private writerと毎readでschema/target/hash/commandとの完全一致を検査し、破損時に現在値から元identityを再構成しない。

この操作は遅延/background Applyがなく、lookupは元target現在認可とC06 receiptだけを必要とするため、495のfull-command耐久保存より狭いanchorを選ぶ。immutable binding Content fingerprintとrequest/command hashがpayloadを束縛し、scalar identityが元targetを保持する。臨床変更/reason/full beforeは短claimに入れず、成功full historyまでは当該request-memory Qにだけ存在する。第二key/outcomeなし。

初回は非lock認可/identity preflight→request/command準備→Runner短claimでbinding INSERTとClaimMetadataWriterによる本表INSERT。同fresh winnerだけが書き、失敗はbindingもrollbackする。claim metadata成功は医師確認成功ではなく臨床履歴も作らない。

既存keyはincoming比較より先に保存元targetの現在scope認可とanchor/Content整合を検査する。incoming REQUESTを正規化しrequest hash/元fieldを照合、一致request＋保存pet/life identity＋scope actorだけでCOMMANDを再構成する。command hash/binding fingerprint検証後にQ化し、現在life/sourceで元意味を差替えない。anchor欠落/破損はresult-unavailable、異なる有効意味はkey conflict、修復なし。別recordで同keyを送ってもanchor上書きなし。bodyなしlookupはCOMMAND再構成/Applyをしない。

`GET /medical-records/{record_id}/operations/{operation_key}?intent=medical_record.vital_retrospective_correct`

既存allowlistへこのintentだけを追加。binding+clinic indexの本表からResolveLookupTargetし、URL record不一致404。認証済み元actor/scopeと元targetの現在view（同院/pet関係含む）を要求し、URL/current requestを元target証拠にしない。成功replay/readはedit権再取得/再確認/再Applyを要せず、成功後fence喪失でfresh writeへ変えない。削除後に既存history-viewで認可できなければfail closedで、新削除閲覧grantを捏造しない。

<a id="posthumous-vital-history"></a>
## 4. 成功full history・監査・read

audit_logsのgeneric nullable JSON・diff-only・marshal失敗省略・per-binding臨床履歴uniqueなし・権威ある専用read未特定という不足に対し、専用不変履歴を選定する。audit_logsは必須監査として保持し、一般audit consumerを変更しない。成功history自身が認証済み確認証跡で、operation成功receiptはC06である。

<a id="posthumous-vital-history-storage"></a>
### 4.1 medical_record_vital_correction_history（選定保存差分、未供給）

| 列/制約 | 選定内容 |
|---|---|
| id | bigint generated PK |
| clinic_id,binding_id | bigint NOT NULL |
| medical_record_id,vital_id,pet_id | bigint NOT NULL CHECK >0 |
| source_version,resulting_version | integer NOT NULL、source >0、resulting=source+1 |
| actor_staff_id,confirmed_by | bigint NOT NULL、actor >0、confirmed_by=actor_staff_id |
| recorded_at / schema_version | timestamptz NOT NULL / smallint NOT NULL CHECK=1 |
| canonical_history / history_sha256 | text NOT NULL / char(64) NOT NULL lowercase64hex |
| UNIQUE | (clinic_id,binding_id)、(clinic_id,vital_id,resulting_version)、(id,clinic_id) |
| 唯一の複合FK | (binding_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICT |
| index | (clinic_id,medical_record_id,vital_id,id DESC) |

identity参照は業務lock下に検証し、cascade可能なbusiness FKを付けない。immutable triggerでUPDATE/DELETE拒否。deferred success-required triggerは同clinic/bindingのoperation_resultsを検査し、participant側も成功refsが同history rowを指すことを返却前に検証する。tenant RLS＋既存record view、新たな広範staff閲覧なし。過去revisionを捏造/backfillせず、最初のeventは訂正時に観測したsource_versionの事実だけを表す。

Canonical HISTORY exact keys:
`{schema_version,medical_record_id,vital_id,pet_id,operation_binding_id,operation_key,command_sha256,source_version,resulting_version,actor_staff_id,confirmed_by,confirmation,reason,recorded_at,life_snapshot,before,after,changed_fields}`。
before/afterは完全SNAPSHOT。changed_fieldsは実差のあった8臨床field名の辞書順配列でversion/updated_atを含めない。
life_snapshot exactは`{local_pet_id,clinic_binding_id,common_pet_id,subject_version,basic_revision_id,basic_revision,life_state,deceased_at}`。identity/版は文字列、life_stateはdeceased、deceased_atはnonnull UTC。basic_revisionは実際のCommonPetBasicRevision.Revision系列番号で、row ID/subject版/架空counterではない。basic_revision_idは別のrow ID。これはprivate履歴で、public readではcommon UUID/bindingを除き確認版/時刻を保持する。現在master名・免許番号・auth/session token・edit_session_refは入れない。

typed serializerはSQL NULL/空/zeroを保持し、unknown/incomplete key、不正有限数表現、identity/版不一致を拒否する。mutation前にJCS serialize→exact shape再parse/検証し、検証済みbytesを保存する。必須履歴にbest-effort MarshalJSONを使わない。DB guardは非空parseable JSON・基本typed shape/必須値を検査し、SQLにJCS関数があるとは仮定しない。canonical bytes/hash検証はapplication write/readが所有する。SQL predicateはIS TRUE/coalescingでUNKNOWN通過を防ぐ。read時decode/hash/schema不一致は503で部分履歴を返さない。

検証後にserver transaction時刻Tをclock_timestampから一度取り、history.recorded_at/after.updated_at/C06 result.RecordedAtへ共用する。before時刻は不変。vitalはCAS WHERE `clinic/id/medical_record_id/pet_id/deleted_at IS NULL/version=source_version`、指定field＋**手動version+1**＋updated_at=Tで更新する。返却/再読した実afterと完全な期待snapshotを照合し、DB値が違えば確認済み意味や履歴を修正せずrollback。固定001ではvital_recordsにentity_version_guardはなく、billingsと混同しない。historyの事前serializationは意図したafterで行えるが、実保存値一致が必須である。

<a id="posthumous-vital-audit"></a>
### 4.2 必須audit_logs

action=`vital.retrospective_correct`、resource=`vital`、resource_id=vital ID、actor_type=staff、actor_id=認証staff、clinic_id=trusted。old_value/new_valueはhistory before/afterと同じ8臨床field全量（明示NULLを含む）をmutation前にstrict serializeする。
metadata exactは`{schema_version:"1",medical_record_id,vital_id,pet_id,operation_binding_id,operation_key,correction_history_id,source_version,resulting_version,confirmed_by,history_sha256}`でID/版は十進文字列。reason/bodyをmetadataへ重複保存せず権威ある完全履歴へ保持する。必須JSON/SQL失敗は全rollback。strict RawMessage adapterで既存ambient audit保存へ渡し、一般audit規律を一括変更しない。

<a id="posthumous-vital-history-read"></a>
### 4.3 history read

`GET /medical-records/{record_id}/vitals/{vital_id}/retrospective-corrections?before_id={id}&limit={n}`

limitは省略default20、整数1..100（pagination値で臨床規則ではない）。before_idは任意正十進文字列。history.id降順、`id < before_id`のkeysetでoffsetなし。exact200は`{items:[HISTORY_VIEW],next_before_id:string|null}`。追加1行で続きがあると確認できたときだけ最後の返却IDをnextにし、それ以外null。

HISTORY_VIEW exact:
`{id,schema_version,medical_record_id,vital_id,pet_id,source_version,resulting_version,actor_staff_id,confirmed_by,confirmation,reason,recorded_at,death_basis,before,after,changed_fields}`。
death_basisはcontextと同じpublic shape、ID/版は文字列、before/afterは完全SNAPSHOT。common UUID/operation binding/command/fingerprintは返さない。idはhistory row ID。

`GET /medical-records/{record_id}/vitals/{vital_id}/retrospective-corrections/{history_id}`は同HISTORY_VIEW、元target不一致404。

両readとも既存medical_records:viewの現在record/vital history-view・同院/pet関係/row可視性を検査し、actor所有だけで認可しない。既存accessを拡張せず、history hash/full schemaを検証する。空一覧は新方式eventなしを意味し、以前の編集不存在を証明しない。通常測定履歴/current GETは別である。

<a id="posthumous-vital-locks"></a>
## 5. TXと並行性

### 5.1 専用consumerの選定順

採択済み[予約global lock契約](../reservations/technical-spec.md#reservation-global-lock-contract)はrecord子入力を含むため、予約を変更しなくても**reservation-graph:write:v1は必須**。operation binding及び全local-pet/common/reference lockより前に取得する。464の選定optional ContextGateと467 try-gateを共用する。optionalはRunner拡張を指し、本操作の参加任意を意味しない。第二hook/namespace/遅いgate取得なし。共通runtime供給は未完である。

短claim完了→最外Runner TXでreservation-graph try Gate→binding NOWAIT→元target現在認可/replay/unknown→savepoint→Apply。

Applyの明示lock順:
1. clinic-scoped local pet **SHARE NOWAIT**。安定local identity下でbindingを非lock再解決し、保存anchor/Qのpet/record関係を照合。
2. common_pet_subjects **SHARE NOWAIT**→別statementでcurrent revision再解決・同subject life snapshot。D6内部を再利用しsnapshot出力/NOWAIT adapterを追加、第二classifierなし。
3. record_edit_rights **UPDATE NOWAIT**、actor/session/epoch/current holder検証。
4. medical_records **UPDATE NOWAIT**、current clinic/有効親/pet/view/edit検証。旧finalized拒否なし。
5. vital_records **UPDATE NOWAIT**、target/parent/pet/daily-null/source版/full digest再読。TX前rowをbeforeにしない。
6. 必要local owner **SHARE NOWAIT**。明示先行取得の代わりに隠れたblocking JOINを呼ばない。lock済recordに対するowner identity変更は拒否。
7. 認証医師と変更先staffを重複除去・数値ID順で **SHARE NOWAIT**、必要同院assignmentを(staff_id,clinic_id)順で **SHARE NOWAIT**。共用designation検証predicateを適用し、auth/staff ownerの現在grant保護をこの位置で保持。後から先行rowへ強いlockを取りに戻らない。
8. 実INSERT/UPDATEが必要とする残りFK参照（例auditのclinic）を未取得なら **KEY SHARE NOWAIT**。全ownerをlockしない。bindingは取得済み、history/claimにbusiness FKなし、record/vital/変更staffも先行取得済み。
9. source/death/time/confirmation/current認可検証→strict history/audit bytes→vital CAS→実after検証→history/audit INSERT→C06 refs/version。Runner resultとdeferred検証まで一体commit。

明示clinical/fence/reference lockの競合はoperation.Busy→savepoint rollback→operation-in-progress。元keyを維持しspin/新key/部分効果なし。abort statement後に成功扱いしない。grant検証errorは既存分類を保つ。

これはlegacy通常vitalのparent-first順そのものではない。逆順再取得/独立TXを隠すhelperを呼ばない。既存parent-first writerのlocal pet SHAREは新SHAREと互換で、既にparentを保持していれば新parent NOWAITが即失敗する。local UPDATE保持者は新local prelockをcommon取得前に止める。late owner/staff辺はNOWAITで有界化し、全global順監査済みとはしない。

<a id="posthumous-vital-common-writers"></a>
### 5.2 共通writerの有限適合

将来common-basic revision participantはsource local pet **KEY SHARE NOWAIT→common subject UPDATE NOWAIT**とし、後のrevision FKが新たなlocal待ちを発生させない。source local検証は保持し、guard-trigger subject版/pointerとdeferred publishを二重加算しない。

正当なidentity/lifecycle保護でlocal pet UPDATEを要するdeath/undo adapterはそれを最初に取り、common保持後のlocal upgradeをしない。必要複数localは数値ID順、複数common subjectはUUID順。同commonを共有するだけで全local petをlockしない。後から別localを逆順に取得せず、取消orchestratorが追加資源を要すればsubject前に宣言するかNOWAIT rollbackとする。

PR845固定headのSwapCagesはlocal pet UPDATE→common reader SHAREで、common UPDATE→source local FK KEY SHAREは逆辺となる。source local prelockでこの辺を除く。実deadlock観測/全repo無deadlock保証ではない。[D04未来予約取消](../../../spec/domains/owners-pets/lifecycle.md#pet-death-future-reservations-cancel-20261022)は必須で任意手作業へ降格しない。必要reservation gateはbinding/local/commonより前、取消TX/outcomeはD04 orchestratorへ保持し、全取得順供給済みとはしない。取消予約の自動復活なし。

activation receiptは、(a) local-first SwapCages対common revision、(b) parent-first vital writer対本訂正、(c) death取消orchestrator対本訂正の3交錯を示すこと。現writer適合不足はnamed supplier blockerであり、この専用契約の選定順を未定義へ戻さない。全D04/491/D15 writer完成とはしない。

<a id="posthumous-vital-confirmation"></a>
## 6. 確認とUI

actor/confirmed_byはserver導出。T06と同じ[共用医師predicate](../accounting/doctor-review-internal-projection.md#review-internal-physician-dependency)をmandatory ambient adapterと単一staff/reservation ownerの検証済みdesignation写像で供給する。staff_type default doctor、license/occupation text、通常assignmentだけで新鮮な権威を代用しない。欠落/未検証は本actionを止め、新role/grant/第二資格方式を作らない。

UIは正確な元vital/版、全変更差分、元・訂正後測定時刻、死亡instant/版、reasonを提示し明示confirm-and-correctを送信する。意味入力変更は表示済確認を無効化し、元outcome確定後だけ新key。取消はwriteなし。autosaveでdraftを保持しても本endpointへ自動送信しない。同actorがedit fenceを保持する。既存正本で許された担当医以外の医師も確認可で、担当医/測定staffを書換えない。

成功履歴にactor=confirmer、元/訂正値、reason、確認life basisを保持する。他actorのdoctor_id/booleanで再利用可能な確認権にしない。失敗claimは成功確認証拠ではない。

<a id="posthumous-vital-acceptance"></a>
## 7. 合成ACと有限供給

全て期待仕様のみ、NOT_RUN。

| 条件 | 期待結果 |
|---|---|
| 元版/digest又は未変更fieldを他writerが変更 | 409/zero-write。CAS更新で古いsource確認を救済しない |
| 明示再取得でsession/epochだけ更新、source/death同一 | 既知拒否後の同Content/key可、自動確認なし |
| 自身の成功後sourceが変わった同key再送 | 元target現在認可付き旧receipt、再適格性/fence/source mutation検証・履歴追加なし |
| 同key別field/reason/time/target | 元target認可を先に行いkey conflict |
| unknown/応答断 | 元key lookupだけ、新key迂回なし |
| death/undo/revisionと訂正の両先行順 | common lock後別statement再解決、古い確認lifeはdependency-changed |
| NULL status＋nonnull死亡時刻 / 不明life・死亡時刻なし | deceased OR分類を維持 / 本slice順序証明不可 |
| 元/訂正時刻の同時・後・精度損失・不明時点 | 拒否。default now()の存在を根拠にしない |
| 確認欠落/偽医師ID/失権/actor≠editor | zero-write |
| 未変更歴史staff inactive / staff変更 | 前者保持、後者現在active参照検証 |
| NULL/空/zero、数値文字列表現 | 区別保持。指数/plus/非正規/JSON number拒否、float衝突/未選定clearなし |
| lock busy/serializer/update/history/audit/result/deferred失敗 | 全効果rollback、claimが残っても成功確認なし |
| 後日の元値変更、破損履歴 | full before/afterを不変再読、現在master補修なし、破損は部分成功にしない |
| 通常create/update/delete/direct repo・daily/import | 死亡変更の専用guard迂回を閉じる。495 P1全子停止へ戻さずdaily別親/maintenance境界保持 |
| 訂正成功 | visit/charge/inventory/dose/通知/予約/死亡変更の副作用0 |

| owner | 必要subset/receipt（未供給） |
|---|---|
| 491/D6＋504/D04統合 | life snapshot・NOWAIT shared primitive固定SHA、source-local先行と必須予約gate・取消全outcome/交錯適合。PR845を採用済みとしない |
| T06共用staff/reservation | designation検証、現在grant/assignment保護、ambient predicate/source SHA |
| 331/494 | 専用participant/fence/target view-edit、通常medical-record vital create/update/delete/direct callerの死亡拒否/dispatch閉包 |
| 489＋464/467 | 実在Q/claim metadata/lookup/replay/errorを再利用。ContextGate/try-gateで必須reservation gateをbinding前へ供給。新status/hook/receiptなし |
| audit/storage | 選定2不変表、strict serializer/private reader、必須audit/deferred receipt guard。DDL適用なし |
| daily/import | maintenance非並行又は別契約、同record代替mutationの迂回なし。全親Done待ちにしない |

この同actor・既存vital訂正subsetに新しい未回答業務判断はない。childなし後日create・別職員準備/医師確認・時点不明は既決ACを維持した別の有限技術packetであり、削除/再質問しない。個人claim/実行ID/供給SHA・実装/受入は未完。Readyは計画/契約状態と実行前提を区別する。

<a id="posthumous-vital-evidence"></a>
## 8. 固定根拠と確認限界

2026-10-24のroot選定入力（Library version0）を日本語正本へ整理した。入力のrevision3は原稿内改訂番号でLibrary version0と別。PR845のopen Draft/head23e944bdの記述は原稿の22:23 UTC観測で、現在状態の再保証ではない。

固定main `7dad9adc7f066fab5cc734dc8ff7e125a80d67d4`:

- [既存vital版のCAS/手動加算前提](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/migrations/001_init.sql#L6338-L6350)、[storage guard](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/migrations/001_init.sql#L10236-L10243)、[billingだけのguard接続](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/migrations/001_init.sql#L10904)。実DB未確認。
- [494 fence/claim/順序](../../../work/emr/emr-494-edit-draft-contract-20261020.md)、[495別資源の既存訂正契約](../../../work/emr/emr-495-consultation-correction-contract-20261021.md)、[C06 DTO](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/internal/operation/dto.go)。
- [vital model](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/internal/model/vital.go)、[exact numeric](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/internal/model/exact_numeric.go)、[現request](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/internal/medicalrecord/vital_request.go)、[既存validation](https://github.com/MinoruSoga/AnimalEkarte/blob/7dad9adc7f066fab5cc734dc8ff7e125a80d67d4/backend/internal/medicalrecord/vital_validation.go)。
- [audit既存保存形](../../../architecture/erd-detail/audit_logs.md)、[共用預り初回受領hook](../accounting/owner-deposit-initial-receipt.md)。旧b81 source/FK/SwapCages/daily/import観測は[作業計画](../../../work/emr/posthumous-contract-plan-20261023.md#posthumous-plan-writer-order-facts)へ保持する。

文書の静的整合と固定source確認だけで、実DB・全repo writer閉包・runtime供給・臨床受入を証明しない。
