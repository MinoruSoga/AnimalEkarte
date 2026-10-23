# T02 / EMR-501 A・B 限定技術契約

| 管理項目 | 記載 |
|---|---|
| 文書ID・種別 | design.examinations.provenance-official-summary／限定技術契約 |
| 決定状態 | 2026-10-23：仕様整理ownerがレビュー済み原稿を技術案として選定。入力の未採択proposalは作成時点の履歴 |
| 実装状態 | 正式実行ID/claim・必要共有供給・実装接続・activationは未完。PR公開/main採用/実装完了は別 |
| 実機検証状態 | tests/build/codegen/DB/Docker/browser NOT_RUN。fixtureは期待値のみ |
| 文書確認日 | 2026-10-23（静的反映） |

入力原稿 `exam-provenance-summary-proposal-20261023.md` をMacで取得・可読確認。受領時SHA-256 `e31fbc254735e3fc4530a44f83326046cb6eb79d28eebdf44bde40ba08227b4e`。以下の調査時刻・source観察・Plane読取は原稿作成側の記録であり、本consumerの全コード/実DB再監査ではない。本反映でPlane操作・製品コード/API YAML/DDL/設定の変更は行わない。

<a id="exam-provenance-summary-scope"></a>
<a id="exam-provenance-summary-evidence"></a>
## 0. 根拠と優先範囲

2026-10-23 08:28–08:30 UTCにGitHub mainの正本文書とPlane501本文・全4コメントをread-only再取得した。
- [D07 lifecycle](https://github.com/MinoruSoga/AnimalEkarte/blob/cff4debe2e4620584e3afc4faaabea370afc7d37/docs/spec/domains/examinations/lifecycle.md): 取得blob dc8e815eb0795244fccbb6aa8d69f897aa34e33b。
- [D07 technical spec](https://github.com/MinoruSoga/AnimalEkarte/blob/cff4debe2e4620584e3afc4faaabea370afc7d37/docs/design/domains/examinations/technical-spec.md): blob 79cf72a02470169a25bb6eda921390cbba899fb9。
- [D4正本](https://github.com/MinoruSoga/AnimalEkarte/blob/cff4debe2e4620584e3afc4faaabea370afc7d37/docs/design/domains/examinations/result-mutation-contract.md): blob 2089d243964649de86a43d98f41f244bef5fe3e7。
- [Plane501](https://app.plane.so/baritechllc/browse/EMR-501/): 4e0266a1-43be-440d-afa9-2e11de99d42cのA/B未選定、af76270e-0b25-43ab-8b76-abcddb1af639のC採択済みを区別する。個人assigneeは空、機能ownerが無いという意味ではない。
- ERD main: exams blob d300cf82ee166bd0f9c27e4a6b014230aad56910、examination_revisions b52bf1509230903a1d04fd0c1a4b32f5ee78ddb3、examination_field_semantic_mappings bb1b0123a1790ad320c6e2b76c78f14fd61411d3、examination_revision_items 4c60cb197965cf245191d64252c98a13ccef7999。ERDは静的DDL定義で稼働DBの証拠ではない。

業務の再質問は不要。臨床mapping具体値・種対応・出典承認は47の既存承認bundle依存であり、この案は値を生成しない。

重要な差: Plane501 Bは有効official限定。lifecycleは保存済みnonofficial/workingの表示も要求する。Bの正式な名前を「official結果要約」とし、その値候補にはnonofficialを混ぜない。既存detail/履歴とD03/D04のnonofficial consumerに残るACを末尾で明示する。Bを全感染症表示完成とせず、確認操作を新しい保存必須条件にしない。

<a id="exam-provenance-summary-provenance-wire"></a>
## 1. A: wireの選定

新しい別臨床write ownerは作らない。既存 POST /api/v1/examinations と PATCH /api/v1/examinations/{id} に、以下の追加subsetをAPI ownerが同版で供給する。PATCHのD4 v1の9field意味は変更せず、新規provenance経路を exam-result-command-v2 に分離する。旧v1のdate:null=不変更を勝手にclearへ変えない。

### 1.1 closed provenance object

provenance は presentなら全4key必須、未知/重複key拒否、object自体のnull拒否。

- schema: literal "exam-provenance-v1"
- source_kind: "in_clinic" | "external_manual" | "legacy_unknown"
- source_details: string | null
- performed_date: exact {date: YYYY-MM-DD | null, precision: "day" | "unknown"}

新規external createは source_kind="external_manual" 必須。source_details=nullは出典詳細不明、非nullはtrim後非空。既存正本に長さ上限は確認できないため、本案の有限な技術選択としてtrim後1〜2000 Unicode code pointかつUTF-8最大8000 bytesを採る（臨床要件の既決値ではない）。超過は400、切捨てなし。既存read値はこの上限で切捨てず保存原値を保持する。病院名必須・証拠添付必須にしない。空白だけは400、nullへ自動変換しない。dateはGregorian有効日、時刻/zone禁止。dayは非null、unknownはnullの対だけ。month/year/timestampをdayへ丸めない。日付省略と不明を区別するため、新規external createではprovenance必須。通常院内createの既存経路へexternal専用入力必須を拡張しない。

PATCH provenance omitted=不変更、present=4key全置換。performed_date omitted/null objectは400。既存top-level dateとprovenanceを同送すると400（同値でも拒否）。provenance内date:null+unknownだけが明示日付解除。source_kindをlegacy_unknownへ新規入力・他値から変更することは禁止する。ただし現在保存値がlegacy_unknownで、同じ値を保持してdate/detailsを訂正するPATCHは許可し、lock後のbefore値との一致を検証する。legacy_unknownを理由にその行の他属性訂正を一律禁止しない。旧NULLはNULLのままread保持し、単なる読戻しでlegacy_unknownへ正規化しない。院内jobの来歴を手入力でexternalへ偽装する変更は禁止。元来歴が明示済みのsource_kind変更は通常PATCHに載せずD07既存訂正の対象へ返す。

PATCH v2 top-level allowlist = D4 v1の9field + provenance + change_reason。provenanceがある場合v2を選ぶ。change_reasonはstring、trim後非空・最大500 rune（既存訂正理由と同じ境界）。既存provenance/dateの訂正で必須。理由だけの更新は400。既存変更に理由を要求するが、新規入力に新しい医師承認は要求しない。再confirmは既存どおりstatusのみ、provenance/項目訂正と同送不可。

前official参照はclient入力ではない。previous_official_revision_id、revision id/version、actor、recorded_at、measured_at、job_idはこのprovenance requestに許可しない。doctor解除は今回Aの出所/date範囲には含めず、D4 v1 doctor:nullの意味を変更しない。

### 1.2 read DTO

D4 detailのatomic snapshotへrequired `provenance` と `revision_context` を追加する。items-only応答に不必要な出所情報を重複追加しない。detailのschema別cacheはv2に切り替える。

provenance exact keys = schema, source_kind, source_details, date, performed_date_precision, measured_at。
- source_kind = in_clinic/external_manual/legacy_unknown/null。NULLは未収録の旧値でありin_clinicへ補完しない。
- source_details = 保存string/null。既存空stringはそのまま返し、表示上は不明として扱う。
- date = YYYY-MM-DD/null、performed_date_precision = day/unknown/null。旧null精度はunknownへ書換えず未収録として保持。
- measured_at = 保存UTC RFC3339Nano/null。external手入力は本sliceでnull。日だけの資料へ00:00を作らない。

revision_context exact keys = current_revision_id, current_revision_version, kind, previous_official_revision_id, recorded_by, recorded_at, change_reason, recording_evidence。
- ID/版はcanonical positive decimal string/null。kindはworking/official/null。
- revision未作成は前三項・previous・理由をnull。recording_evidenceはexact {state, operation_binding_id, result_state_version}。state=verified_operation|legacy_unrecorded。ID/版はcanonical string/null。
- 新A保存はD4の同TX必須監査にbinding_id、clinic_id、examination_id、after_result_state_versionを記録する。read snapshotの(exam clinic/id, result_state_version)に完全一致する成功監査とC06 success resultをbinding_idで結合し、actorと成功RecordedAtをrecorded_by/atへ返す。最大created_at/最新ID/近似時刻検索はしない。1世代に複数binding、失敗attempt、success receiptとのtarget/version不一致はsource_integrity_error。
- 導入前rowやA証拠未導入writerの既存rowで対応監査が無い場合はrecorded_by/at=null、recording_evidence={state:legacy_unrecorded,operation_binding_id:null,result_state_version:null}として原値readを許す。来歴未収録を通信/構造破損と混同せず、doctor/現在viewer/nowで埋めない。新A writerがsuccessを返すには上記証拠作成必須、欠落は保存rollback。読取側だけでは旧rowと実装欠陥の欠落を断定しない。
- revisionありは固定revisionの保存値を返す。previousはserverで検証した参照のみ。recording_evidenceはrevisionのactor/created_atを置換しない補助情報で、当該固定revisionを作ったsuccess bindingが立証できればverified_operation、無ければlegacy_unrecordedとし、旧revision自体を拒否しない。
- detailのtop-level date/precisionとprovenanceは同じ保存値から作り、差があれば契約不整合として成功応答にしない。

### 1.3 前officialと保存

現examsはdate/precisionを持つがsource_kind/details/measured_atを持たない。revisionだけへprovenanceを保存すると初回confirm前の手入力が失われる。これを架空working/official作成で回避しない。

選定する最小追加DDL（446専有/466統合、現migration規約による）:
- exams.result_source_kind TEXT NULL CHECK IN('in_clinic','external_manual','legacy_unknown')
- exams.result_source_details TEXT NULL
- exams.measured_at TIMESTAMPTZ NULL
既存row default/backfillなし。exams date/precision、revisionの同名5属性、previous_official_revision_idは既存を再利用。新しい台帳・患者列・別結果本文表なし。D4 counter追加とは別に明示受領する。

新external保存ではexamsとitemsと監査へ保存し、初回confirmを強制しない。current_revision_version=nullを維持できる。既存working/official append時は同TXのaggregateからprovenance/date/measured_atを固定revisionへコピーする。既存revisionはimmutable、現在master/現在examから過去版を補完しない。unconfirmは解除元officialをpreviousへ、working更新は同じ元officialを引継ぎ、再confirmは直前系列のofficialをpreviousへ保持する。初回officialはprevious=null。

previous FKは現DDLだと同医院しか保証しないため、writerは同clinic・同exam・kind=official・過去version・整合する系列を検証し、自己参照/cycle/別exam/未来版を拒否する。legacy不整合を最大version検索で修復しない。復元も過去snapshotの値を使うが、appendのpreviousは現在訂正系列から決め、過去のpointerをそのまま戻さない。

revision_itemsは現ERD上のnonnull判定列とD4 nullable current itemsに差がある。501 Aが勝手にnormal/false/0へ埋めない。446/466と501 Cのnullable snapshot adapter受領を必要subsetにする。既存official旧値は変更しない。

<a id="exam-provenance-summary-write-safety"></a>
## 2. A: 操作安全性と失敗

PATCHはD4のIdempotency-KeyとX-Exam-Result-State-Version必須。Scope/action/update、claim target、結果counter、before/after identity、C06 receipt/lookup、Q、lock順を再設計しない。v2 canonical commandにprovenanceとreasonを含め、Content.command_schema/hashをv2に固定。既存v1成功receiptのreplayを引き続き解決できるようdual decoderを残す。同key v1/v2は別Contentでkey-conflict。Applyが別DTO closureを読むことは禁止。

provenance/date・items・新revision・counter(+1)・監査・usage・成功receiptは同一最外TX。新external createはD4規定どおり存在しないexamのexpected counterを要求せずcounter=1。create ownerの既存C06契約とcurrent patient authorizationを受領して接続する。createのkey/固定target/未成功lookupが未供給ならexternal新規writeだけactivation停止し、A readと既存exam更新の準備は先行可能。ここはD4採択からcreateが完成したと推測しない明示dependencyである。

HTTP/C06失敗はD4継承。invalid date/presence/source=400 validation、無認証401、現在grant不足403、scope不在404、版/依存/operation key競合409、unknown503/照会中202。before/after anchorの非開示を維持。監査保存失敗は全rollback。timeoutは元key lookupし新key再createしない。成功後refresh失敗は保存失敗に変更しない。

<a id="exam-provenance-summary-summary-wire"></a>
## 3. B: official-only endpointとclosed DTO

選定route: GET /api/v1/examinations/infection-summary?pet_id={id}&surface=patient_top|medical_record&medical_record_id={id}
- pet_id/surface必須。medical_record surfaceだけmedical_record_id必須、patient_topでは禁止。未知query/重複keyを400。
- clinicはtrusted選択医院、actorは現在認証。query clinic/actor指定なし。新たなfilter/期間/limit/pageなし。全履歴UIではない。
- patient_top: 現在D04対象患者閲覧認可 AND examinations:view AND 対象行scope。
- medical_record: 現在medical-records:viewと指定medical_record_id/pet_id整合 AND examinations:view AND 対象行scope。
- D04既存患者scope resolverを使用し、新しいgrant名を作らない。owners:view単独で開放しない。D20/LIFFへの接続なし。各source refを辿る詳細も再認可する。

200 exact top-level keys:
`schema, clinic_id, pet_id, animal_species_id, surface, medical_record_id, evaluated_at, projection_digest, completeness, mapping_state, entries, notices`
- schema="exam-official-infection-summary-v1"。
- IDはcanonical decimal string、species/recordはnullable。evaluated_atはsnapshot評価時刻で臨床実施日ではない。
- completeness="complete"のみが200成功。件数0と部分取得を同型空応答で混ぜない。
- top-level mapping_state=verified|not_configured|unverified|species_unknown|not_applicable|mixed。全entryが同じmapping_stateならその値、異なる場合はmixed。承認済みspecies/mapping根拠なしにverifiedを生成しない。
- entriesはFIV→FeLV→mf→HWの固定順で各meaningを一件ずつ返す。各entryの必須mapping_stateはverified|not_configured|unverified|species_unknown|not_applicable。混在してもverified entryを抑止しない。未設定/未検証/種不明のentryは全candidate配列空、selected_source_ref=null、selection=mapping_unknown。明示not_applicableはselection=not_applicableで全候補空/selected_source_ref=null。raw valueや臨床結果を推測しない。
- noticesは閉じたenum配列: nonofficial_records_present, correction_in_progress, unknown_date_present, unmapped_records_present, source_metadata_unknown。存在情報も同snapshot/同じ行認可からだけ導出し、非開示行数・ID・存在を漏らさない。

entry exact keys: meaning, mapping_state, selection, selected_source_ref, candidates, unknown_date_candidates, ancillary_candidates。
- meaning=FIV|FeLV|mf|HW。selection=unique_latest|ambiguous|unknown_date_only|no_official_value|mapping_unknown|not_applicable。
- selected_source_refはunique_latest時だけSourceRef、他はnull。
- candidatesは保存実施日/信頼できる実施日時と精度による最大候補集合（同順位・比較不能を含む）すべて。unknown_date_candidatesは日付不明候補。ancillary_candidatesはvalue_stateがpending/not_performed/unknownの非置換候補。unknownは下記定義の原値空かつ明示availability証拠なし等に限り、非空原文が臨床的に未解釈という理由でvalue_presentから除外しない。

Candidate exact keys: source_ref, raw_value, unit, value_state, date, performed_date_precision, measured_at, source_kind, source_details, correction_state, previous_official_revision_id, mapping_ref。
- raw_value/unitは保存文字列。value_state=value_present|pending|not_performed|unknown。value_stateは保存原値の有無/明示availabilityであり、陽性/陰性/正常/実施確認/意味理解の判定ではない。以下のclosed adapterを使う。
  1. raw_valueは固定revision_items.inspection_valueをそのまま使う。result/normal_value/status/is_assessed/is_abnormalは代替原値にしない。
  2. 認可済み固定revision/itemに結び付いた明示availability証拠がない通常official行は、inspection_valueのUTF-8 byte長が1以上ならvalue_present、0ならunknown。trim/truthiness/数値化なし。「0」「陰性」「-」「不明」や空白原文も非空の保存表現として原文保持し、意味が判らないことだけでancillaryへ追い出さない。value_presentは臨床的に有効な検査値と認定する意味ではなく、lifecycleの「原値文字列・欠測表現を保持／0等を空へ変換しない」に基づく表示上の存在区分。
  3. pending/not_performedは明示保存証拠が当該clinic/exam/revision/itemと一意に結び付き、相互排他的状態をownerが供給した場合だけ。現revision_itemsのstatus等から導かない。この証拠adapterは未供給依存であり、現表だけの行では両状態を生成しない。証拠が存在するが対象/版/状態に矛盾する場合はsource_integrity_error。
  4. value_presentは通常の順位候補に含め、解釈不能な原値でも同じ時間順位を適用する。unknown（原値空・availability未収録）、pending、not_performedだけancillaryへ分ける。空の新しい行で以前の非空保存原値を消さない。名称/自由文から医学的意味を付与しない。
- correction_state=current_official|correction_in_progress。workingそのものの原値をCandidateへ返さない。
- SourceRef exact = {owner:"examinations", examination_id, revision_id, revision_version, revision_item_id}。全ID required canonical string。exam_resultsの可変item IDや配列位置で代用しない。
- mapping_ref exact={mapping_id, revision}。同clinic、固定field ID、承認済みspecies、meaningへ明示対応する既存mapping行。名称/LLM/任意固定IDによるmappingなし。
- raw enum外/参照切れ/不整合は部分候補を200へ隠さず失敗。原値の臨床的未解釈はvalue_state=unknownを意味しない。原値の存在状態と通信/参照破損は別に扱う。

<a id="exam-provenance-summary-completeness"></a>
## 4. B: 完全性・系列・source identity

一つのDB整合snapshotで現在認可、患者/種対応、active exam集合、各headと固定official系列、revision items、mappingを読む。exams/revisions/itemsだけ別時点を混合しない。候補取得に既存履歴20件APIを使わない。

各exam系列について:
1. current headがofficialならそれだけを現候補にする。過去officialを別検査として二重算入しない。
2. headがworkingでprevious_official参照が有効なら、元officialをcorrection_in_progressとして返す。workingの新しい原値/dateをofficialへ混ぜない。
3. official未作成/旧sealed revisionなしは値候補外。認可済み存在だけnoticeへ。架空officialを生成しない。
4. soft-delete/適法無効化された系列は有効候補外。別examへ移された/固定petと不整合なsourceは勝手に現在petへ帰属させない。
5. head、previous、item、保存種/患者、mappingの系列/医院整合を検査。修復や最新masterによる履歴再構成なし。

rankingは保存された実施日/実施日時と、立証できる精度だけを使う。異なる既知日は日付で比較する。同日候補の双方に信頼できる保存measured_atがあり、当該sourceの実施日時として精度・timezone/offset・由来が明示され、同じ時間軸で順序を立証できる場合に限り、その実施日時で比較する。DBのtimestamptz型や値の非nullだけでは信頼性/精度の証明としない。reader adapterは元job/sourceの保存証拠を同snapshotで検証し、表示用入力時刻から生成された値を採用しない。一方でも日単位のみ/時刻精度不足/zone由来不明ならその同日候補間の順序は不明。同じ実施日時も同順位。信頼できる日時が一部だけの集合では、証明できない比較を日時あり優先へ変換しない。
一意な直近候補は、他の全既知日候補より後であることを立証できた候補だけ。候補間の時間区間が重なる/比較不能なら最大候補集合を返しambiguousとする。日だけの候補はその保存実施日の全日区間であり、架空00:00へ点化しない。入力/受信/更新/確認時刻、exam ID、院内外優先は順位根拠にしない。日付不明は別集合で既知日を上書きしない。日付不明しかなければ全不明候補の原値と検査日不明を返す。

mappingは既存verified_by/verified_at/evidence/operation_idだけを形上満たす架空行では不可。47から対象bundleとmapping行の採用revisionを受領する。現tableにeffective/revoked statusが無いのでmax(revision)を無条件activeとする選択はしない。承認bundleが指定する行revision集合をread構成に固定し、別bundleへ切替時は全再投影。mapping writer UI/承認workflow/臨床値seedは本slice外。旧field IDが明示mappingへ解決できなければmapping_unknown/noticeとし、no_recordへ偽装しない。

completeは同医院・患者・要求view scopeの全関連sourceを走査し、かつ全検査に対する当該viewの完全性を認可resolverが保証できるときだけ。部分行認可しか証明できなければ全患者最新を名乗らず403。正常0候補は200 no_official_valueであり未検査ではない。DB timeout/中断/上限到達/完全性に必須なreader不足・取得失敗/系列破損は503にしてpartialを返さない。既知の未供給availability証拠adapterだけならpending/not_performedの詳細区分を供給せず、非空official原文はvalue_present、空原文はunknownの既定fallbackで先行可能。提供済readerの実通信/整合失敗を未供給fallbackで隠さない。本案の有限技術予算は未実測の初期値であり、1応答全体でCandidate合計1000件、JSON UTF-8 body 1 MiB、source走査100000 revision-item行、実行10秒（現在認可の再解決も含む）と選定する。これは既存の業務件数制限ではない。サーバ内部stream/集計で全候補を評価し、同日ties・unknown_date_candidates・ancillary_candidatesを含む返却集合を構成するが、予算を一つでも超える場合は503 summary_incompleteを返し、本文候補/件数/途中digestなし。先頭1000件成功や同日候補の一部だけ返却は禁止。本sliceは候補ページング/持続snapshot tokenを追加しない。正常completeのdigestは走査完了後にだけ確定する。予算を増減する後続変更も明示契約改版とし、UIの全患者最新という意味を変更しない。

HTTP errorは既存API error envelopeを使いcodeは400 invalid_summary_request、401 unauthenticated、403 summary_not_authorized、404 scope_not_found、503 summary_unavailable|summary_incomplete|source_integrity_error。error bodyに患者原値/source refs/候補数を含めない。詳細診断は権限ある内部ログへ。readは結果保存/確認/機器受信を起動しない。必要clinical-use receiptは既存D07 reader participantを受領し、failure時は503、独自no-opで迂回しない。

<a id="exam-provenance-summary-freshness"></a>
## 5. projection freshnessとcache

全candidate配列は完全SourceRefのowner→examination_id→revision_id→revision_version→revision_item_id順へ安定化する。ID/版は正確な整数比較を使いNumber化しない。meaningは上記固定順、noticeは宣言済みenumの固定順、mapping ref/input集合は登録固定順又はID/版tuple順へ正規化してからresponse/digestを生成する。重複SourceRefは整合エラーであり隠してdedupeしない。これは表示/serialization順で、同日候補の臨床順位をIDで決めるものではない。

サーバは毎GETを正本snapshotから計算し、初回sliceに永続summary表・eventual materialized view・共有結果cacheを導入しない。HTTP Cache-Control: private, no-store。304でcurrent access再検査を省略しない。

projection_digest="sha256:"+lowerhex。canonical objectはschema/clinic_id/pet_id/animal_species_id/surface/medical_record_id、適用mapping bundleの固定ID/revision集合、全走査対象のexam/result_state_version/head/revision/item identityと本文/出典/date、採用/除外区分、各候補value_stateと適用availability証拠のstable identity/版（証拠なしは明示null）、noticeを含む。availability adapterのexact証拠ref/profileが未供給の間はその枝を利用開始せず、任意JSON/現在時刻をhash代替にしない。非空raw valueの既定value_present fallbackは維持する。evaluated_atはhashから除外。単なるupdated_at最大値や表示候補だけのhashは不可。digestはCAS権限tokenではない。

FEはメモリcacheだけ、key=[schema,clinic_id,actor_kind,actor_id,session_generation,pet_id,animal_species_id,surface,medical_record_id,auth_generation]。actor identity/session_generation/auth_generationは認証・session管理コンポーネントが管理する内部contextからのみ受領し、ユーザー入力・URL・保存された古い世代・検査APIの任意値を採用しない。login/logout、actor/医院切替、認識したrole/scope変更・失権では、世代更新と旧本文cacheの非表示・破棄を新contextの取得/表示より先に行う。別loginで同じ初期世代を使って旧cacheを再利用せず、session境界でも分離する。

各要求のclient request_generationは応答順序だけを管理する。現在のactor/session/clinic/auth世代と要求世代が全て一致する最新応答のみ反映し、遅延応答で破棄済みcacheを復活させない。serverは毎requestの現在認可を別途検査し、これらFE世代を認可根拠にしない。未通知の遠隔失権の即時反映SLAや新認証APIは設けず、403又は認証再確認で判明したcontext変更時にも上記破棄を行う。Cache-Control:no-storeだけでメモリcacheが消えるとは扱わない。

患者/種変更でも旧値を即非表示・破棄。background refetchの途中で古い「現在確定」バッジを維持しない。D4 detail versionをsummaryへ合成しない。

同clientのcreate/update/unconfirm/reconfirm/delete/許可済み無効化成功receiptで該当clinic/petのsummaryを即stale化してGET。同じrecord headerも同cacheを使う。mapping bundle変更・画面再entry・window focus・再接続でもGET。表示中は60秒ごとの再確認を本案の技術選択として選定し（既存正本の指定値ではない）、タイマー停止/非表示復帰で再取得。60秒以内の反映をSLA/業務保証にせず、network・非表示・認可による停止を含むbest-effort読取とする。TTLは臨床有効期限ではない。再確認中/失敗は「更新未確認・前回取得時点の記録」と明示し、古い値を現在値として保証しない。別actorの変更まで瞬時更新を保証する設計ではない。

保存成功後GET失敗は保存receipt成功を保持し、再保存せず読取のみretry。401/403/404又はscope/auth変更時は旧本文を破棄し非表示、503/network時だけ同scopeの過去snapshotを更新未確認として保持可。遅延応答が新しい要求世代や保存成功後のinvalidate世代を上書きしない。

<a id="exam-provenance-summary-handoff"></a>
## 6. owner/必要subset/停止範囲

| subset | 既存機能owner | 受領する物 | 未受領時の停止範囲 |
|---|---|---|---|
| A wire/DTO/current provenance | 501 A、445 API | v2 strict schema、adapter、生成物、API/FE同版SHA | 新A write/read shape activationのみ |
| A preofficial storage | 446専有/466統合 | exams 3列、既存revision属性との対応、nullable snapshot必要差分 | A provenance保存のみ。D4 coreを未選定へ戻さない |
| A transaction | 501 C、489、516必要subset | D4 ambient participant、v2 Q/content、audit/receipt/lookup、全writerコピー | 当該write activation。Detachは別契約 |
| 新external create | 501 create owner/489 | 既存create C06のexact key/target/成功・未成功lookup/生成ID receipt、採用SHA | 新規external createだけ。既存exam read/update準備は可 |
| B projector | 501 B | 同snapshot完全reader、fixed source/series解決、official-only DTO | B endpoint activation |
| B permission | 234/既存D03・D04 read owner/501 | current permission resolverと全対象scope証明、usage reader参加 | 該当surface B公開 |
| B availability証拠 | 501 Bと既存D07 source/項目状態owner | 固定revision/itemへ結び付く明示実施・結果待ち状態の保存証拠とreader adapter、exact証拠ref/profile・stable identity/版（現revision_itemsだけでは不足） | pending/not_performedの詳細区分だけ。通常非空official原値はvalue_presentとして先行可、空値はunknownを保持 |
| B mapping | 47/既存master owner | 臨床承認bundle、正確species/field/mapping row revision・出典 | 対象meaning値投影だけ。合成fixture設計は可 |
| B FE refresh | 501 BとD03/D04 consumer | schema別cache、invalidate対象、失敗表示、旧request破棄 | 該当surface表示 |

個人担当・claim/hunk・正式実行IDは本案では割り当てない。実装着手packetで path/symbol、同PR/別PR、単一編集owner、供給/受領SHA、activation対象、残ACを記録する。既存主配置は backend/internal/medicalrecord/examination_{request,response,service_create,service_update}.go、examination_revision_{service,repository,workflow_repository}.go とFE検査手入力/患者summary。新projectorファイル名/route登録は445現claimと合意してから編集。生成物手編集なし、001への無条件追記なし。

<a id="exam-provenance-summary-acceptance"></a>
## 7. 固定AC（期待値のみ、全件NOT_RUN）

A01 external createで日付不明/null、出典不明/nullを保存し再読。actorは認証値、doctor未選択保持、架空日付/working/officialを生成しない。
A02 omitted/null/day/unknown、未知精度、invalid日、日時string、duplicate/unknown key、top date併用を区別。拒否はzero-write。details 2000 code point/8000 bytes境界と超過、legacy_unknown保持訂正対新規採用、旧監査未収録read対新A監査保存失敗を別caseとして確認する。
A03 unversioned→初official→unconfirm→working変更→再confirmでprovenance snapshotと前official系列が整合。旧official不変。
A04 別院/別exam/未来/workingをpreviousへ偽指定は入力拒否。保存内参照不整合はfail closed。
A05 v1 date:null不変更を保持、v2 unknownだけ明示clear。旧v1成功keyとv2を混ぜない。
A06 同時provenance/items更新はD4 CAS一方のみ成功。required audit/revision失敗は全rollback。応答断lookupで二重保存なし。
B01 全候補が20件超、古い日が後日入力。全体の実施日で選定、入力時刻を使わない。
B02 同日でも双方の保存実施日時/精度/zone/由来が立証できる場合は実施時刻で比較する。一方日単位・時刻精度不足・zone不明なら候補を併記し、同時刻も曖昧。同値/不一致/院内外を固定優先にしない。信頼できる時刻が一部だけの混合集合で日時あり優先にしない。日付不明は既知直近と別、unknownのみも原値保持。
B03 working訂正中は元officialだけ訂正中表示、working原値をofficialへ混入させない。再confirm後は新official、無効化後は旧値をcurrentとして残さない。
B04 未承認/未設定/旧field対応不明/species不明をno_recordと区別。FIV verified/FeLV unverified等のmixedは各entry状態とtop-level mixedで返しverified候補を保持。明示not_applicableもmapping_unknownと区別。合成mapping検査を47臨床承認の証拠にしない。
B05 非空inspection_valueの0/陰性/「-」/意味不明原文はvalue_presentとして保持し順位候補へ。value_presentを陰性/正常/実施確認へ読み替えない。空stringは明示availability証拠なしならunknown。result/status/assessmentで原値又はavailabilityを補完しない。固定item証拠があるpending/not_performedのみ区分し、これらの新しい行は既存値を消さない。
B06 権限不足、部分scope、跨院、別pet/record、認可失効で元値/存在noticeを漏らさない。D20/LIFF経路なし。
B07 snapshot途中のconfirm/mapping切替/患者変更でも世代混合なし。broken series、timeout、truncationは503で空成功にしない。Candidate合計1000/1001、body 1 MiB境界、走査100000/100001、10秒予算で超過は503かつpartial原値なし。同日tiesだけで予算を超えても一部成功へ落とさない。
B08 保存成功後refresh失敗でも保存成功保持、同scope旧値は更新未確認。401/403、login/logout、actor/医院切替、認識したrole/scope変更・失権では旧本文cacheを先に破棄し、別session/同じ初期世代で再利用しない。遅い成功応答を新しいactor/session/auth/request世代へ適用しない。未通知遠隔失権の即時反映を保証しない。
B09 polling/focus/reconnect/書込invalidateで再評価。証拠だけの変更でもvalue_state/availability identity版を束縛したdigestが変化。DB集合順が変わっても正規順のresponse/digestは一致し、重複SourceRefは失敗。projection digest変更を臨床状態変更の診断としない。

既存AC対応: external-source/write-safetyはA、date-order/ambiguous/mapping-unknown/no-record-failure/access-boundaryはB、revision-refreshはA+B。infection-auto-savedとinfection-result-statesのnonofficial原値表示はBのDone対象外で、D03/D04の既存detail/履歴接続consumerへ明示残す。selected-print、機器mapping297、尿86、死亡後331、Detach、臨床47、全D07実機ACは閉じない。

## 結論

この案でAの出所/date presence/前official/初回確定前保存先と、Bのofficial専用route/DTO/認可/完全性/鮮度/失敗を一つの有限な選択として提示する。既存Cの採択を再議論しない。供給SHAとcreate C06必要subset、臨床mapping47は受領依存として残す。本書で技術選定を固定するが、コード/DB/試験/activation済みとは記載しない。

## 原稿の採択前レビュー履歴
候補応答/走査予算、legacy_unknown保持訂正、監査の固定binding+counter identityと旧来歴不明read、details有限上限、60秒pollの技術選択性を追補。外部投稿なし。

独立レビュー追補: §4の同日候補を一律曖昧にする誤りを修正。正本に従い、双方の信頼できる保存実施日時・精度・時間軸を立証できる比較だけ順位へ使用する。入力/受信/記録時刻による代替は引き続き禁止。採択前の調整記録。

独立レビュー追補2: Bのvalue_stateを固定revision_items.inspection_valueの原文存在と明示availabilityに分離したclosed adapterへ具体化。非空原文は意味不明でも順位候補、空原文はunknown。pending/not_performed証拠は限定owner依存として残し、新enum保存列・臨床意味は追加しない。
