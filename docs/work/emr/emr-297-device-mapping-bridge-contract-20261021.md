# EMR-297 日常機器受信と明示mapping sampleの接続採択契約 R4

2026-10-21。採択契約・runtime未実装・試験NOT_RUN。実装/試験/DB適用/実機/LLM呼出/患者データ外部転送なし。最新main e71ddf03462a0ee7a9864ad2ce6854b243591770、EMR-297本文＋全22コメント、268本文＋全16コメントを照合。指定関連269も本文＋全7コメントを確認したが日次会計帳票票であり、機器mappingに変更を加える対象ではない。

## 1. 採択業務と先行可能な範囲

D07の日常経路はHTTPS Chromeでページ表示中のNX600/AU10V受信、origin-wide排他、本日診療中カルテ選択、患者1件の待機、検体IDで患者推定しない、decode後raw非保存。PU-4010/IDEXXのdecoder存在は日常運用許可ではない。

EMR-297の明示sample経路は、結果1件の選択→原本DB保存→辞書の一意対応・未対応だけLLM下書き→人が送った対応だけ確定→以後は確定表による取込。原本とLLMを呼んだ行の送受信/失敗を保持し、未確定draftは臨床結果を書かない。新draftは古いdraft/確定表を勝手に変更しない。単位自動換算なし。既存gpt-6-luna/store:false/reasoning:none/toolsなしの方針は変更しないが、今回その実行・送信をしない。

この2経路の橋は『同じrawを常時保存する』ではなく、rawがある明示sampleのsource identityと、日常受信で既に保存したdecoded job identityを照合して関連付けること。日常rawを復元・生成してsample扱いしない。rawが既に破棄されたjobの未対応値は既存decoded itemから再解決できるため、mapping確定後の再処理を再採取待ちにはしない。

## 2. 保存済みframeの現在使えるsubset

過去のPlane証拠で確認できるもの（今回実行し直したPASSではない）:
- NX600: 628B、16コード、元slice [0,628)、SHA256 3dc2b8db25a879069cfd02e125fc574c96cab8f3f29b4ac56db1f64e0b82ab35。2026-10-11、旧source 8ca80c12のdecoder-only再生PASS。辞書draft/原本保存の合成DB試験は別source a759858d、臨床医院承認ではない。
- AU10V: 88B/1コード、PU-4010:449B/8コード。旧decoder-only原値/単位一致記録あり。PUは元高位bit付きrawと解析派生を分離する。
- IDEXX血球2040B/11項目、SNAP361B/FIV・FeLV：旧RED、359修正候補でGREENの別記録あり。現mainの運用適合や通信成功へ転記しない。

原本場所は既存old_db証跡/利用者Desktopの目録に記録されている。今回Macへアクセスせず、rawを新たに転送・Git/Planeへ掲載しない。既存frameでselector/DTO/辞書・未対応・保存境界のfixture設計は可能。合成fixtureや既存repository内の承認済みfixtureを使う実装準備に『新実機sample必須』を課さない。実医院ID・取得者役割・sample用途・species根拠が未束縛の原本を本番mapping確定やLLM送信へ使わない。

## 3. 現コードで埋める差分

- `mapping_draft.Create` はLLMを含むdraftItemsの後にInsertDraft。context cancel/store error等で原本を保存しない経路があり、『明示draftごと原本保持』を満たすため先に耐久保存へ分割する。
- `selectOneLabResult` は同fingerprint再送を畳み、異なる複数結果を拒否する。8KiB上限。RawFingerprintはdecoder正規化後指紋の場合があり、元bytes SHAと同じとしない。
- `ReceiveFrames` は既存jobの指紋hitを返し、rawをDBへ保存しない。新jobにはdecoded原値/単位/flag/順番を保存。保存後persistの別stageがある。重複受信はmapping更新や再persistの指示ではない。
- `LabDeviceWait` / 現serviceはpetだけをpersistへ渡す。カードにRecordIDがあっても、機器persistにMedicalRecordIDを引き渡していないため、選択したカルテとD03予定のidentityは別途閉じる必要がある。
- master/specはmutable upsertで版snapshotなし。speciesありはspecies specをgenericにoverlay、speciesなしはgenericだけ。受信時とpersist時に別解決するので一つの適用版を記録しないと結果由来を確定できない。
- persisted jobをpersisterは即noopにする。部分対応の残itemを『mappingを直したから』と再送で勝手に既存検査へ追加しない。
- species推定・実施日時欠損時のnow fallback等を新正本へ採用しない。D07の既決unknown保護を保つ。

## 4. 経路・API・grant

実prefix `/api/v1`、下記は相対path。既存routeは維持し追加DTOとowner内部を変更。IDは既存wire互換を壊さず、新版contractでBIGINT正規十進文字列を採用する場合はcaller/serializerを一体更新する。UUID job/operationはcanonical nonnil。

### A 日常受信

POST /lab-device/frames、lab-import:create。既存payload_base64/device_hintを維持しdailyのみ。追加sample永続化・LLM自動呼出なし。serverは認証clinic、日常許可source、受信consumer/session、現待機を確認する。成功応答は既存Resultsを維持する。現実装は途中frameのerrorでHTTP失敗となっても先行frameは別TXでcommit済みの可能性があるため、batch全体atomic失敗と偽らない。同payload再送は各fingerprintで既存jobを再取得し二重結果を作らない。新しい部分成功envelopeをこの変更で追加しない。rawはmemoryでdecode後破棄、HTTP/body/APM/error logにも残さない。

### B 明示sampleとdraft

POST /lab-device/mapping-drafts、lab-import:create。新入力はoperation_key、payload_base64、device_hint、acquisition（sample_origin / captured_at nullable / device_id nullable / acquisition_evidence_ref nullable / usage_category unknown可）、linked_job_id nullable。clinic/患者はbodyで指定しない。sample_originはsaved_capture又はexplicit_sample_receive。後者は設定画面でsample用途を明示して開始した受信で、日常framesへの自動分岐ではない。sampleをこのAPIへ送っても待機消費・job作成・examination生成なし。

選択可能な元入力は8KiB以下。握手のみ/異なる複数結果/不正入力はfailed draftとして元入力を保持し臨床書込0。HTTPサイズ超過や認可失敗はDB保存せず拒否。適切な単独frameへ切り出す場合は元SHA・byte start/end・slice SHAをprovenanceへ記録し、選択slice1件だけを下書きの対象にする。PU bit7除去など変換はdecoder内部、元rawを上書きしない。

短いclaimでは元target anchorとsample commandのprivate rawだけを保存し、business effectを行わない。続くApplyで原本/draft・監査・C06成功receiptを同TX保存。これは『sample受領』成功でありLLM/下書き完成ではない。generation_stateを別に追加してpending→ready/failed/stopped/unknownを管理し、既存status unconfirmed/confirmed/failedと混同しない。既存GET /mapping-drafts/{id}（view）へgeneration_state・元SHA/selector版・linked_job_id・定義snapshot IDを追加し、raw/LLMbody/実値は現方針どおり公開DTOに出さない。

辞書のみまたはLLM未設定のblank候補はreadyまで先行できる。LLMを有効にする場合は各送信前に原requestをDBへ記録し、TX外で呼び、応答/安全なerrorを同itemへ記録。送信結果不明はunknownとして自動再呼出ししない。NULL/JSON null/不正bytesの既存JSON包みを維持、APIkeyを保存しない。既存generation stopは新生成を止め、確定表取込を止めない。実外部送信の範囲・承認・秘密情報の設定は今回の範囲外。

linked_job_idは同clinic/sourceとdecoder fingerprintに加え、保存decoded項目集合のdigestが一致した場合のみ記録する。jobの患者へsampleを付け替えたり、その患者の臨床結果とsample用途を同一視しない。rawが手元にない既存jobはこのAPIへ擬似rawを送らず、Dの再解決だけ利用する。

### C 人によるmapping確定

既存POST /lab-device/mapping-drafts/{id}/confirm、lab-import:edit。operation_key、期待draft版、期待mapping scope版、期待field定義版、既存itemsを受ける。新しい医師権限roleは作らない。人が送ったcode→field/value_shape/unitだけ確定。未提出code/blank/commentは未対応のまま。speciesはdraftの元値を維持し、患者のspeciesから無言で置換しない。

原draft、採用fieldと親検査種別の同医院/現在有効性、mapping value_shapeの既存許容値、generation ready、currentgrantとCASを同TX検査。master/spec upsert、draft confirmed、immutable mapping revision、監査、C06receiptを同TX。採用値/単位は人確認内容を保持し、異単位を値換算しない。失敗draftのconfirm禁止、同成功key再送は元結果。新規draft作成やconfirmから既存job自動再処理は起動しない。

### D 未保存jobの明示再解決

新POST /lab-imports/{job_id}/resolve-mapping、lab-import:edit。operation_key、expected_job_version、expected_generic_mapping_revision、expected_species_mapping_revision、expected_record_versionを受ける。species版は未知species時だけnull（baseline未導入をnullに見せない。対象scopeのbaseline未導入はactivation停止）、record版は未紐付け時だけnullとし、現在jobの対象に対するnull条件もCAS検査する。payload/値/任意petは受けず、保存decoded原値だけを使用する。初回対象はreceived/needs_reviewで未persistのjob。linked患者/recordがある場合は同じidentityを保存時に再検査し、なければ未紐付けの解決結果だけ保存して臨床結果0。

0対応ならneeds_reviewを維持し未対応件数/行を同TX更新。対応がある場合は既存D07の部分対応許可を維持し、対応行のみexam type別に保存。残原値/未対応行を消さない。job items・counts・適用mapping snapshot・N検査/結果・監査・usage marker・receiptを同TX。どれか失敗で全rollback。receiptだけ成功して結果が失われる形にしない。

persisted jobへの再解決は依存変更として拒否し、既存原結果/official保護を持つD07訂正へ渡す。再送hitは元jobを返すだけ。部分persist後の未対応行を追加実施する新機能はこの橋で勝手に作らない。detach再attachによる保護回避もしない。

## 5. clinic / patient / record / order同一性

日常待機は医院に1件、同患者持続/期限切れ未紐付けの既決を維持。選択カードのmedical_record_idをwaitへ追加してpetと一緒に保存する最小deltaを選ぶ。PutWaitはrecord/petを同医院・本日診療中・生存scopeで確認。古いpetのみ待機行はrecordを同患者候補から推測backfillせず、次の明示カード選択で更新する。受信時wait_id/record_id/pet_id/選択版をjobへsnapshotし、後続患者切替で元jobの対象を変えない。

保存時は元recordとpet、現在認可、現在recordの許容状態を再確認。無効化/死亡/対象変更なら未紐付け/要確認として臨床writeを止め、別患者や別recordにfallbackしない。既存persist失敗後の自動unlinkは、失敗したTXに結果write0と確認できる場合の補償だけに限定し、既persist・結果不明・D03関連ありで無条件患者NULLへしない。

order=共通plan item identity。存在しない機器order番号・specimen IDをorderとみなさない。別操作である既存plan owner link（516等）が確定したIDを明示選択した場合のみ、その516操作が同clinic/record/pet/検査種別/有効予定と結果ownerを照合してリンクを同TX保存する。297の受信/resolve TXに新規リンク作成を混ぜない。候補が複数ならpatient一致で選ばない。plan未指定の既存日常受信を不可能にする新必須orderは追加しない。結果保存とorder表示は同じexam ID/版を参照し、plan連動のため別Examinationを作らない。機器へのorder送信機能を追加しない。

## 6. mapping版の最小保存設計

既存lab_device_item_masters / lab_device_mapping_specs / draft / itemsを維持。新しいcounterはraw受信順・検査版とは共用しない。

追加 `lab_device_mapping_heads`：PK(clinic_id,source_type,species_scope)、revision>0。species_scope=''がgeneric、canine/felineが既存正規化scope。追加 `lab_device_mapping_revisions`：同scope+revision UK、snapshot_json、snapshot_digest、confirmed_draft_id nullable、operation_id、actor/time。snapshotはcode昇順のactive mapping行（code、field_id、exam_type_id、採用field定義snapshot identity、unit、value_shape、source master/spec ID、採択由来）。未対応codeを架空mappingに含めない。immutableで過去版削除不可。

genericの既存有効手動対応を失効させない。導入baselineは現設定の静的snapshotとして保存し、過去の臨床承認者/日時を捏造しない。由来unknownを明記する。EnsureDefaultsのfield未設定行はmapping確定と数えない。draftconfirmだけでなく既存master PATCH/configuration/spec writerすべて同head増分・snapshot監査を通す。これが未完なら新適用版readerをactivateしない。

適用はgeneric版＋species版（未知speciesはnull）と採用field版の組を固定した `lab-mapping-application/v1`。species spec優先overlayの既存意味を維持し、異種specへfallbackしない。各itemはその組から一度だけ解決する。受信時の仮解決とpersist時の解決を混ぜず、実persist TXで確定した組を記録する。

新 `lab_import_mapping_applications`：id、clinic/job、application_no、generic_revision、species_revision nullable、snapshot_digest、resolved_items_json、unmapped_count、operation_id nullable（旧受信owner処理は元job transaction参照）、actor/time。同job/application_no UK。resolved_items_jsonはjob item ID→mapping採用ref/field/型/単位/needs_review理由のsnapshotで、原値本文は既存itemに保持し重複保存不要。日常初回と明示再解決どちらもこの記録を同TX生成する。jobにversion>0とcurrent_application_idを追加してCAS。itemsの原code/value_raw/unit/flag/sort_orderは不変、解決pointerだけ更新。既persist結果のsnapshotは新masterで再構成しない。

sample draftには raw_sha256（元payload）、selected_frame_sha256、selected_start/end、decoder_version、selected_source_fingerprint、decoded_digest、acquisition_json、linked_job_id nullable、generation_state、version、create_operation_idを追加。現RawFingerprintと元hashを混同しない。linked_job_idはclinic複合FK。sample再生成は新draftで旧rawを上書きしない。C06同keyretryは同draftを返すが、明示新keyは同rawでも別下書きにできる。

原対象回復用 `lab_mapping_operation_targets`：operation_id PK、clinic、kind、draft_id/job_id nullable、content_digest、不変元scope、原actor。claimと同TX追加。canonical command JSON/digestとsample専用private_raw_payloadの保存列・NULL条件・8KiB CHECK・Apply検証は§7aの既存定義に従う。targetsはbindingとの同医院FKだけを持ち、まだ存在しないdraftや変動するfield/master/recordへのFKをclaimで取得しない。draft/job/record/pet/planの同医院と存在・状態はApplyで検査し、汎用metadataだけで認可しない。

## 7. 同TX / lock / duplicate / C06

新command限定で現C06flat12statusを再利用し既存frames応答は変更しない。owner_kind=lab_mapping、intent=lab_mapping.sample_register / lab_mapping.confirm / lab_mapping.resolve。GET /lab-device/mapping-operations/{operation_key}?intent= は元actor本人かつ（現在lab-import:view又は当該元intentに必要な現在create/edit grant）でID/版だけの限定結果照会。別actorは同医院viewだけで他人のoperationを探査できない。raw/臨床値を返す権限は増やさない。

C06 Contentは短いscalar stringだけのflat mapとし、schema-tag JCSのkind・元対象・raw_sha256/size・人のmapping内容等はowner-private commandへ分離する。expected_* CASはPreconditionsだけ。同key内容変更=operation-key-conflict。ResultResponseはstatus/operation_key/result_refs/committed_versions/recorded_at、details追加なし、EffectSummary内部のみ。sample refs=draft_id、confirm refs=draft_id/mapping_revision_id、resolve refs=job_id/application_id（exam IDsは既存report readerへ）。committed_versionsはdraft_version、mapping_revision、job_versionの該当分。有限対応は succeeded/replayed/not-succeeded=200、operation-in-progress=202、result-unavailable=503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、scope-not-found=404、not-authorized=403。scope-not-foundはbinding不存在であり単なる業務対象不一致を勝手に同statusへ変換しない。型・認可・サイズ等のpre-C06失敗は既存HTTP errorであってreceiptではない。失敗binding元targetを現在認可してから照会し、既成功は現CASより先に原結果を返す。

全新writerのlock順を同一化：current actor/assignment gate→C06 binding→clinic受信station（待機を使う操作だけ）→record/pet owner gate（必要時）→job/draft→mapping headsを(scope文字列昇順)→field行ID昇順→examination/plan owner graph。これと逆にjobからrecordを取る既存attach/revert writerは同owner helperへ統合し、部分採用でdeadlockを残さない。C06 gateとbindingの実コード差は採択済みContextGate/ClaimMetadataWriter必要subsetへ接続、未実装を実在扱いしない。

確定mapping writeはdraft→heads→fields、臨床persistはjob→heads→fields。互いの既存rowを後から逆取得しない。profile/LLM外部処理中にDB lockを保持しない。監査失敗はmapping更新/結果保存/receiptもrollback。受信fingerprint uniqueは現(clinic,source_type,source_fingerprint)を維持し新しいdevice_idを勝手に重複namespaceへ加えない。

raw元hash、正規化measurement fingerprint、decoded digest、operation key、mapping revisionを別概念として保持。payloadの同一再送を畳むselectorと、臨床jobのunique、mappingdraftの明示生成履歴は別。unknown項目・部分対応件数を同jobitemsと同TXで更新し、EMR-360成果を再実装しない。

### 7a. Hashと境界の閉じた定義

以下は新しい技術contractであり現実装済みschemaではない。nested object/arrayの投影はowner-private canonical commandでありC06 Contentそのものではない。SHA256は小文字hex、UTF-8 JCSをhashし、未知keyを拒否する。ID/版/byte長/ordinalは正規十進文字列で投影、日時はUTC RFC3339、nullは省略しない。値/単位/flagはtrim・数値変換・Unicode正規化をせず保存文字列そのまま。配列順を勝手に整列して原順を消さない。

- decoded_digestの投影: {schema:"lab-decoded/v1",source_type,species,measured_at,items:[{ordinal,device_item_code,value_raw,unit_raw,flag_raw}]}。ordinalは元decode順で重複codeも保持。保存jobにないheader・pet名・検体IDを混ぜない。旧jobがこの全fieldを保存していない場合はdigest一致を捏造せずlinked_job_idを拒否するが、既存保存fieldを用いた再解決は妨げない。source_fingerprintは別照合でありdigestに循環包含しない。
- sample owner-private command: {schema:"lab-mapping-sample/v1",clinic_id,payload_sha256,payload_size,device_hint,acquisition:{sample_origin,captured_at,device_id,acquisition_evidence_ref,usage_category},linked_job_id}。payload_sha/sizeはserverが受信bytesから計算する。client指定acquisitionは申告情報であって実機receipt認証済みの証拠ではない。切出し元SHA等はserverが元bytesを保持して検証できる場合のみverified_provenanceへ記録する。callerがsliceだけ送った元fileのSHA/offsetはclaimed_provenanceで区別し検証済みとしない。
- confirm owner-private command: {schema:"lab-mapping-confirm/v1",clinic_id,draft_id,items:[{device_item_code,exam_type_field_id,value_shape,chart_unit}]}。同code複数提出は400、itemsはcodeのUTF-8順へ正規化する。期待draft/head/field版はすべて別PreconditionsでありContentへ混ぜない。value_shape/chart_unitは既存ConfirmLabMappingItemの意味を維持する。
- resolve owner-private command: {schema:"lab-mapping-resolve/v1",clinic_id,job_id}。原decoded内容は不変job anchorで束縛しApplyがdigestを再検査する。期待job/generic/species/record版はPreconditionsのみ。APIから原値・source・patient・任意mappingsnapshotを差し替えられない。
- mapping snapshot: {schema:"lab-mapping-revision/v1",clinic_id,source_type,species_scope,rows:[{device_item_code,exam_type_field_id,exam_type_id,field_definition_revision,field_definition_digest,chart_unit,value_shape,source_table,source_row_id,confirmed_draft_id,provenance_kind}]}。rowsはcode順、nullable confirmed_draft_idとfield_definition_revisionはnullを残す。field_definition_digestは採用した既存field定義のimmutable snapshot hashを必須とし、既存field版がない場合もその定義snapshotを別保存し、存在しない版番号を生成しない。provenance_kind=legacy_baseline/manual_master/draft_confirm。source_tableは既存master/specの2値限定。同scope/revisionのimmutable行が本体であり、snapshot内部へ自己digestを埋め込まない。

sample選択不能時はselected_*とdecoded_digestをnullにし、元payload_sha/sizeと安全なerror codeだけが確定する。bytesを保持したfailed draftの登録はsample受領成功、HTTP不正認可/過大payloadは未受領。通信中断がdurable commit前なら保存済みと報告せず同key照会、保存後なら原receiptを返す。「どんな中断でも絶対保存」とはしない。

generation第一sliceは耐久保存後の辞書/blank計算のみで完結可能。generation pendingは同draftに対する内部CAS（versionとpending状態）でready/failedへ一度適用し、process再起動時にpendingだけ再取得できる。LLMを呼ぶ拡張は送信intent durable記録とunknown判定を実装するまでactivateしない。既存外部LLMが動いていると仮定して二重workerを作らない。stopped/unknownから暗黙再送はなく、同rawの別draftを明示作成する際も外部送信の新実行承認を別途満たす。

C06へ渡すContentのexact flat mapは {schema:"lab-mapping-operation/v1",intent:"lab_mapping.sample_register"|"lab_mapping.confirm"|"lab_mapping.resolve",clinic_id:"<decimal>",target_kind:"sample_registration"|"draft"|"job",target_id:"<clinic_id for new sample; draft_id or job_id otherwise>",command_sha256:"<SHA256(owner-private JCS)>"}。値はすべて短いstring、nested/配列/raw本文を載せない。expected_*はC06の既存Preconditionsへだけ投影する。command_sha256にCASを混ぜない。

VerifiedCommandはメモリ内の型付き検証済み値であり、再起動後の回復用DBではない。sample registerはclaimと同TXのoperation_targetsにowner-private command JSON/digestとsample専用private_raw_payload byteaを耐久保存する。これは不変command入力でありdraft/business行はclaimで作らない。private_raw_payloadはsample intentのみ必須、他intentはNULL、最大8KiBをDB CHECKする。Applyはprivate_raw_payloadの実byte長とSHA256が不変commandのpayload_size/payload_sha256に一致することを検査し、不一致なら原本の置換をせず失敗とする。検査した同bytesを既存draft byteaへコピーしdraft/監査/receiptを同TX commitする。途中停止は受領結果不明であり成功とは返さない。confirm/resolveも同targetsにcanonical command JSONを保存し、receipt/履歴へrawや臨床本文を漏らさない。Applyは型付きcommandからJCS/digestを再計算しC06のcommand_sha256と一致確認し、pure projectorが再生成したflat map全6キーが元Contentと完全一致することを確認する。workerがmemory commandを復元できると仮定しない。失敗bindingでも元target/current認可が可能なanchorを保持する。rawはC06 Content/owner command JSONへbase64保存しない。claim commandのprivate raw保管はreceiptだけでは回復できない失敗・未確定入力を失わないための限定保存であり、日常framesには存在しない。削除TTLを勝手に追加しない。

### 7b. Preconditions reader と待機DTO

新 GET /lab-imports/{job_id}/mapping-context（lab-import:view）は {job_id,job_version,status,clinical_persisted,record_id,record_version,pet_id,generic_scope:{source_type,species_scope:"",revision},species_scope:null|{source_type,species_scope,revision},field_definitions:[{field_id,revision:null|"<decimal>",definition_digest}],decoded_digest} を返す。record_id/record_version/pet_idは未紐付けならすべてnull、古いpetのみjobはpet_idのみ既存値でrecord_id/record_version=nullとしclinical_persisted=falseでも臨床保存可能とはみなさない。読み取り時点のcurrent grant/clinic/jobを確認し、Apply時はすべて再確認する。実値rawはこのreaderに含めない。recordに紐付かないresolveはmapping解決だけで終える。

既存 GET /lab-device/mapping-drafts/{id} の新context部分は {draft_version,mapping_scope:{source_type,species_scope,revision},field_definitions:[{field_id,revision:null|"<decimal>",definition_digest}]}。field定義版が実在しない場合、confirm requestのexpected_field_definitionsは {field_id,expected_revision:null,expected_definition_digest} で固定し、Applyで選定fieldの現定義を同じcanonical snapshotから再hashする。版がある場合もdigestを併用し、未存在版を捏造しない。field digestのexact投影は {schema:"lab-field-definition/v1",clinic_id,field_id,exam_type_id,name,inspection_value,normal_value,unit,sort_order,type_is_active,type_deleted}。これは現model/examination_type.goのExamTypeFieldと親ExaminationTypeの既存列だけで、ID/sort_orderは正規十進string（sort_order null維持）、type_is_active/type_deletedはbool、名称/原値/単位はstring。現在fieldにvalue type/choices/revision列はないためそれらを架空の属性にしない。value_shapeは既存mapping側のenumでありfield型と混同しない。type_deletedは親DeletedAt.Validだけを投影し、timestamp自体は含めない。

PUT /lab-device/wait のDTOにmedical_record_idを追加し、既存pet_idと両方必須の新caller契約へ更新する。既存 GET /lab-device/board のtoday visit cardにrecord_versionを追加する。既存cardのmedical_record_id（LabDeviceTodayVisit.RecordID）/pet_idとこのrecord_versionを選択元としてPUTへ渡し、job用mapping-contextから新規待機recordを推測しない。expected_record_versionはこのboard cardの版を渡し、旧pet-only callerが勝手に別recordへ紐付かないようactivationをcaller/BE一体で行う。成功応答は既存wait情報へmedical_record_id/record_versionを追加する。古い待機は新規明示選択まで臨床自動persist不可で、受信jobは未紐付け/要確認とする。新plan ID入力はframes/wait/resolveに追加しない。D03共通planとのリンクは既存516 owner-link操作が担当し、この297 sliceはrecord identity保存とexam ID/版を引き渡すところまで。既存リンクがある結果はその保護を照会し、取り外し・revertの迂回を禁止する。

lookup GETは§7の元actor限定view-or-originalgrant方式に固定し、新mutationは既存create/edit middlewareを維持する。viewだけの利用者によるPOST新claimを許可しない。全12statusは§7の有限表だけとし独自unknown statusは増やさず、不明結果はresult-unavailableで表す。

## 8. 不足を限定した着手単位

A 297: 先行原本保存/selected provenance/C06/JSON nullable caller、辞書・blankのdraft、確定表snapshotと確定監査。保存済みframeの仕様を使い、実機receipt待ちにしない。
B 297＋268受信owner: wait record identity、decoded jobのmapping application、明示再解決、全writerlock/CAS閉鎖。D03 plan linkは516採択identityの必要subsetのみ。
C 268/402: 現医院・機器個体・接続/操作者・許可build/操作範囲付き実機receipt、日常NX/AUの実運用受入。A/Bの実装設計完了と別であり現在実施しない。
D 359/360: 既存parser/未対応件数修正の成果を参照。過去GREENを最新main実行済みと報告しない。IDEXX/PUの運用可否をこの変更で拡張しない。

必要な本人への新業務質問は現時点なし。既存のsample用途/clinic/master/実機許可は運用受入資料として未確認を維持する。269の金額判断や新LLM採用をこの設計へ混ぜない。LLM外部処理の実行・資格情報投入・臨床mapping確定・DB適用は別承認。隔離環境も試験も作成/実行しない。

記述するfixture/AC：日常rawDB0とsample raw保持、LLM前中断でも原本存在、複数結果/握手failed、元hash/変換後hash分離、SQLNULL/JSONnull/invalidbody保持、未知species、同code別医院、generic/spec版競合、明示blank/未対応、confirm監査rollback、同job再送/resolve idempotency、部分persist拒否、旧wait petのみ移行、record/pet/plan不一致、失権、原値単位不等式保持。すべてNOT_RUN。

## 固定根拠

- https://app.plane.so/baritechllc/browse/EMR-297/ （最新本文・22コメント、2026-10-21取得）
- https://app.plane.so/baritechllc/browse/EMR-268/ （最新本文・16コメント）
- https://github.com/MinoruSoga/AnimalEkarte/blob/e71ddf03462a0ee7a9864ad2ce6854b243591770/docs/spec/domains/examinations/lifecycle.md
- 同commit backend/internal/medicalrecord/{routes_lab,lab_device_receive_service,lab_device_mapping_draft,lab_device_mapping_select,lab_device_mapping_store,lab_device_item_master_service,lab_device_exam_persist,lab_device_decode}.go
- 同commit backend/internal/model/{lab_device_mapping_draft,lab_device_item_master,lab_import}.go
- 保存済み旧証拠の入口：old_db/docs/lab-go/hospital-field-pack/README.md、evidence/capture-inventory-20261011.json、nx600-decoder-replay-20261011.json、emr297-nx600-draft-db-20261011.json。今回はPlaneの固定hash付き過去受領記録を根拠とし、元Macファイルや旧DBを再実行・再照合したと主張しない。
