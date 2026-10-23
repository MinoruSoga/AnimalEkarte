# T01 / D4-JOB-01: 機器Detachの限定技術契約

| 管理項目 | 記載 |
|---|---|
| 文書ID・種別 | design.examinations.job-detach／限定技術契約 |
| 決定状態 | 2026-10-23：仕様整理ownerがレビュー済み原稿を技術案として選定。入力の未採択proposalは作成時点の履歴 |
| 実装状態 | 正式実行ID/claim・必要共有供給・実装接続・activationは未完。PR公開/main採用/実装完了は別 |
| 実機検証状態 | tests/build/codegen/DB/Docker/browser NOT_RUN。fixtureは期待値のみ |
| 文書確認日 | 2026-10-23（静的反映） |

入力原稿 `exam-detach-contract-proposal-20261023.md` をMacで取得・可読確認。受領時SHA-256 `53525eb5a683c8cd4ae2eadb02c56afc581d5d3b18e72644778fdd7de66773b1`。以下の調査時刻・source観察・Plane読取は原稿作成側の記録であり、本consumerの全コード/実DB再監査ではない。本反映でPlane操作・製品コード/API YAML/DDL/設定の変更は行わない。

<a id="exam-detach-scope"></a>
<a id="exam-detach-scope-detail"></a>
## 1. 結論と変更しない境界

既存 POST /api/v1/lab-imports/{job_id}/detach を、job ownerの単一C06最外operationへ移す。job専用のimmutable claim target、同TXの補償manifest、現在認可付きsame-key lookupを追加する。exam単体用examination_operation_targetsへ多件jobを押し込まない。各examでRunner.Executeを起動しない。

D4 coreはmain cff4debe2e4620584e3afc4faaabea370afc7d37の採択を継承する。result_state_version/current_revision_version/link-only digestの役割、501 C取りまとめ、既存C06を再選定しない。Detachだけを本契約の供給待ちとし、独立受信・閲覧・適合済みの他mutationを一律停止しない。

業務状態はdevice persisted→received、pet_id NULL、fingerprint保持、由来全検査の補償。fixture Revertのpersisted→revertedと別。使用receipt・manual mutation・official/confirmed・revision-less completed seal・現在の親/関連保護、plan owner linkによる拒否を保持する。自動unlink、確認ダイアログ、強制付替え、新しい解除権限、任意理由必須化、機器送信を追加しない。旧親finalized列を現在の診療終了eventへ無断置換しない。D4/最新D03の通常編集と補償保護の既存区別に従う。

<a id="exam-detach-evidence"></a>
## 2. 根拠・確認範囲

最新GitHub treeはmain cff4debe2e4620584e3afc4faaabea370afc7d37。以下をGitHubから再取得した。
- docs/design/domains/examinations/result-mutation-contract.md（blob 2089d243964649de86a43d98f41f244bef5fe3e7）§§4,6,7,11,13.1。
- docs/design/domains/examinations/technical-spec.md、docs/spec/domains/examinations/lifecycle.md。
- docs/design/shared/operation-participation.md、backend/internal/operation/{contract,dto}.go。
- docs/work/emr/emr-516-517-plan-owner-link-contract-20261021.md。
- docs/architecture/adr/007-lab-device-receive-and-commit.md。日常接続経路の古い記述は現在lifecycleの優先関係に従い採用しない。
- docs/architecture/erd-detail/lab_import_{jobs,revert_receipts,exam_retractions,exam_retraction_items}.md。
- docs/work/emr/emr-297-device-mapping-bridge-contract-20261021.md。
- backend/internal/medicalrecord/{routes_lab,lab_import_revert_service,lab_import_revert_tx}.goは移行差分のみに使用。未完成製品sourceを業務正本にしない。

Plane297の既存読取artifactも参照。今回Plane live readはcatalog changedで拒否され、最新コメントを再取得したとはしない。495/501/516の役割は上記最新採択文書とrootの既存調査に基づく。担当個人・現claim・供給SHAは未確認で、割当て済みとはしない。

<a id="exam-detach-target-identity"></a>
## 3. association_versionで空集合のcycleも固定する

「epoch不要」の旧設計を撤回する。必要DDLとしてlab_import_jobs.association_version BIGINT NOT NULL DEFAULT 1 CHECK(association_version>=1)を選定する。初期1は技術cutover基準で、過去のattach回数を復元した値ではない。updated_at/source_fingerprint/active exam ID集合だけでは同job/同pet/空集合のABAを区別できない。

association_versionはrefresh可能CASではなく、利用者が選択した対象cycle identity。job lock下でpet/status/active exam集合を変更する最外TXが一回だけ+1し、同TXの中間多段変更で重ねて増分しない。異なる最外TXの変更はそれぞれ進める。attach/persist/reprocess/確定失敗unlink/detach/revert/JobID変更/復元の全writerが参加し、dedup/no-op/read/replayは+0、失敗は全rollback。上限overflowを拒否し、巻戻し/再利用しない。

遅延した失敗unlink等も元cycleの期待association_versionをlock下で再照合し、現在版を無断再採用して新cycleを消さない。active examはD4既定の新IDで再作成するが、0件でもassociation_versionにより旧未成功keyは再attach後dependency-changed。旧成功keyは元receiptだけを返す。

このv2は新しい必要供給で、旧v1の稼働/公開済み大移行を仮定しない。新供給はv2へ統一し、旧draft keyが存在しても現在版を補完してApplyしない。結果不明を新keyで迂回しない。過去soft-deleted examの通常attach復活/上書きを認める新業務も追加しない。

<a id="exam-detach-command"></a>
## 4. Read context・closed command

### 4.1 元操作対象の読取

既存 GET /api/v1/lab-imports/{job_id} の認可済み応答にoptional detach_contextを加える。既存boardのカードにも同reader由来のcontextを加算可能とし、FEが表示値からdigestを合成しない。供給されない場合はGET jobを使う。context取得はclinical examination_detail usageを作らず、job操作対象を読むだけ。

context exact keys:
- schema: literal lab-job-detach-target-v2
- job_id: lowercase canonical nonnil UUID
- association_version: canonical positive signed-int64 decimal string
- target_sha256: lowercase 64hex

server canonical target bytesはUTF-8 canonical JSON、key順固定 schema,job_id,association_version,source_type,pet_id,examination_ids。schemaは上記literal。pet_idはcanonical positive decimal string又はJSON null。examination_idsはactiveな同clinic/job exam IDの数値昇順・重複無しstring配列、0件は[]。source_typeは保存enumそのもの。clientが新しいsourceを宣言できない。

job SHARE→active exams SHARE（ID昇順）で同snapshotを読む。親coordinatorが必要ならD4順に先取りする。全membership writerはjob UPDATEを先取りし、locked jobから集合を再読する。jobの途中状態を返さない。現在view許可がない場合context非開示。単なるreadでjob/association_version/baselineを初期化しない。non-persistedでもcontextを読めるが、現在stateがwrite可能との保証にはしない。

### 4.2 Mutation wire

POST /api/v1/lab-imports/{job_id}/detach
- Idempotency-Key必須。C06 ParseOperationKeyのcanonical lowercase nonnil UUID。
- JSON body exact: {"target":{"schema":"lab-job-detach-target-v2","job_id":"<UUID>","association_version":"<positive decimal>","target_sha256":"<64hex>"}}
- target/body/schema/association_version/digestのmissing/null、未知key、重複JSON key、非canonical UUID、URLとbody job不一致は400、zero-write。
- trusted clinic/actorは既存現在認証から取得。client clinic/actor/exam配列/測定値/理由/statusは受け付けない。
- examごとのexpected counter headerは要求しない。[D4 job引渡し](result-mutation-contract.md#exam-job-detach-handoff)どおりjob lock後の再読値をD4 participantへ渡す。
- 認可は現在lab-import:editと既存補償ownerのscope/保護。新role、通常exam DELETE grantの追加、独立jobへの一律D03 edit session要求はしない。
- 旧bodyなし呼出しはcontext自動補完せず400。API/FE同版切替。共有既存Idempotency-Key/CORSの受領を確認し、新header追加は不要。

### 4.3 C06 identity

Scope exact: ClinicID=trusted clinic; OwnerKind=lab-import; ActorKind=staff; ActorRef=current staff canonical ID; Intent=lab-import.device-detach。
Content exact keys: command_schema=lab-job-detach-command-v2, action=detach, job_id, association_version, target_sha256。
association_versionとこのdigestはrefresh可能CASではなく、明示選択した対象cycle/集合のidentity commitment。実行時exam counters、lifecycle/plan/fence世代をContentへ混ぜない。Content-only adapterで閉じ、別captured DTOをApplyで使わない。対象canonical bytesをclaim metadataへ保管してdigestと照合する。
PreconditionsはD4/coordinatorの型付き現在必要条件だけ。ユーザーの対象集合が変わる時は同内容retryではなく、元操作の成否を確認して新しい明示操作にする。

<a id="exam-detach-capacity-budget"></a>
### 初期利用開始の有限capacity

active exam 1000件、補償対象result item総数20000件、下記detach-capacity-v1（target＋全補償parent/item snapshot＋容量manifest投影）のcanonical UTF-8合計8 MiB（8×1024×1024 bytes）を上限とする。これは未実測の初期技術値であり、既存機器の業務上限/性能実測値ではない。8 KiB frame制限をjob全体件数の根拠にしない。

元targetの現在認可後、新規実行の既知超過はclaim前のlimit+1読取又は上限付きstream集計で検知し、部分成功なし・claimなしの413（ownerのdetach_capacity_exceeded）で返す。未認可の件数/本文は開示せず、共有C06 status enumを追加しない。保存済みsame-key成功replayは現在認可付き元結果の処理を先に行い、現在集合のcapacity再検査を成功再送の条件にしない。

claim後/lock後の再読でも同じ上限を検査し、競合増大時は全business変更をrollbackする。分割Detach/途中commit/切捨ては禁止。外側business TXの処理budgetは10秒（未実測の初期値）。超過/評価不能は既存unknown/元key結果照会に従って停止し、timeoutを不成功確定にしない。request cap、取得途中失敗、臨床利用保護の拒否は別に扱う。

大きい既存jobは明示未対応として保持し、上限を変更するなら別の性能根拠と契約改訂を必要とする。新規上限を満たしたことだけで臨床利用/版/認可の保護を免除しない。

容量定義追補：Library `libfile_48562e1048e88191a73fae7c8f4164ee` の `exam-detach-capacity-bytes-correction-20261023.md` を取得・可読確認した（受領時SHA-256 `cfaf22a4d2e84c9d1c76a4aecd028f687d88ed8fe1d21ad548e991b5e68d2f02`）。以下は容量専用closed projectionの選定で、既存保存/公開wireの変更ではない。

<a id="exam-detach-capacity-measure"></a>
### 容量投影と計量単位

容量はDB物理サイズ・メモリ使用量・HTTP応答総量ではなく、以下のdetach-capacity-v1投影のcanonical UTF-8 bytesで判定する。

B = len(T) + Σ len(Pe) + Σ len(Iei) + len(M)

- Tは§4.1のcanonical target全体。schema,job_id,association_version,source_type,pet_id,examination_idsの既定固定key順を維持する。
- Peは補償対象examごとの§7/7.1のexact parent snapshot全keys。
- Ieiはそのexamに属する各補償itemの下記exact保存scalar全keys。
- Mは下記容量専用manifest投影。実際のpaged GET envelope/cursorは加算しない。
- lenはJSONを下記規則でencodeした後のUTF-8 byte数。各成分を独立JSONとして計量し、成分間の改行・区切りは0 bytes。各JSON内部のbraces/brackets/quotes/colon/commaはすべて含める。同じ値が別componentに登場するときは各component分を計量する。
- B <= 8,388,608を許容し、B >= 8,388,609で超過。1000 exams/20000 itemsもinclusiveで、どれか一つでも超過すれば拒否する。

<a id="exam-detach-capacity-snapshot-projection"></a>
### Exact snapshot projection

Peのexact keys:

schema,id,clinic_id,medical_record_id,pet_id,exam_type_id,doctor_id,job_id,date,performed_date_precision,result_summary,machine,status,current_revision_version,result_state_version,created_at,updated_at,deleted_at

schema=lab-exam-retraction-parent-v1。provenance供給profileのv2はschema=lab-exam-retraction-parent-v2とし、result_source_kind,result_source_details,measured_atの3keysをrequired nullableで加える。profile間の選択・未収録とNULL・初official前の扱いは§7.1のまま。ここに列挙しない表示relation/contextを容量へ追加しない。

Ieiのexact keys:

id,exam_id,exam_type_field_id,name,inspection_value,normal_value,result,unit,reference_value,ref_min,ref_max,qualitative_min,qualitative_max,is_abnormal,status,sort_order,created_at,updated_at

これらは同headのD4 ExamResultD4から表示派生is_assessedだけを除いた保存scalar。独自のitem schema key、削除時刻、job ID等を追加しない。元item IDとsort_order/NULLを保持する。計量は各itemを一回ずつ行い、親間の走査順で総数は変わらない。実装の走査順はexam ID数値昇順→item ID数値昇順に固定する。

<a id="exam-detach-capacity-manifest-projection"></a>
### Exact capacity manifest projection

Mのexact keys:

schema,job_id,target_sha256,before_pet_id,after_pet_id,before_status,after_status,before_association_version,after_association_version,retracted_examination_count,examinations

- schema=lab-job-detach-capacity-manifest-v1。これは今回明示選択する**容量だけのprojection**であり、新規保存表/新公開DTOではない。
- job_id/target_sha256/before_pet_id/before_association_versionは元targetの実値。after_pet_id=null、before_status=persisted、after_status=received。
- after_association_version=before+1。retracted_examination_countは補償対象exam数のcanonical非負decimal string。
- examinationsはexam ID数値昇順、各要素のexact keysはexamination_id,medical_record_id,pet_id,before_result_state_version,after_result_state_version。元scope/版から作り、after_result_state_version=before+1。0件は[]。
- +1のint64上限拒否は投影前に行う。架空の飽和値を計量しない。
- 未採番binding/retraction ID、operation_key、recorded_at、clinic/actor等の運用metadata、audit/receipt envelope、DB overheadは容量投影から除外する。これらをNULLや推測IDで代入しない。この除外は保存manifest/audit/receiptからfieldを削除する変更ではない。
- この定義は可変長の臨床補償payloadと対象明細の論理上限である。固定運用fieldやDB/index/transaction overheadまで8 MiB以内になる保証ではない。別の既定件数上限・処理budgetも維持する。

<a id="exam-detach-capacity-serializer"></a>
### Canonical serializer

Tのobject key順は§4.1の既定順。それ以外の全objectはASCII key昇順。対象keysはすべてASCIIなのでlocale依存比較を使わない。配列順は上記固定順。

- whitespace・改行・BOMなし。区切りはcolon/commaだけ。空配列は[]、objectは規定keysを省略せず出力する。
- NULLはJSON null、booleanはtrue/false。missingとNULLを同一視せず、required key欠落はcontract failure。
- ID/版/件数はcanonical decimalのJSON string。UUIDとhashは既定canonical string。ExactNumericはD4で定義した正規decimalのJSON string又はnullで、floatを経由しない。
- sort_orderのみ符号付き整数のJSON number又はnull。10進表記、先頭+・指数・先頭0・-0は使用しない（0は0）。
- timestampは保存instantをUTCのYYYY-MM-DDTHH:mm:ss.fffffffffZ（小数9桁固定）へ無損失変換する。保存値がNULLならnull。dateはYYYY-MM-DD/nullであり、時刻への補完をしない。
- stringはtrim/Unicode正規化をしない。有効UnicodeをUTF-8で出す。double quoteは\"、backslashは\\。U+0000..U+001Fは小文字hexの\u00xxで必ず6 bytes、\n/\t等の短縮escapeは使わない。
- slash、HTML文字、U+2028/U+2029はescapeせずそのUTF-8 bytes。無効UTF-8・孤立surrogate・未定義型・不正timestampはcontract failureで停止し、replacement characterやゼロ値へ変換しない。

このserializerは容量投影専用のbyte確定規則であり、DB JSONBの内部表現や既存HTTP応答の空白/escapeを容量の根拠にしない。Tのdigest用canonical bytesにも同じprimitive文字列表現を適用し、既定のtarget key順を変えない。snapshot/manifest保存の意味・既存wire型は変更しない。

<a id="exam-detach-capacity-recheck"></a>
### Preclaimとlock後の一致

認可後preclaimとlock後の最終判定は同じschema profile・projection・serializerを使う。先行preclaimの取得後に値が変わった場合は、lock後の同snapshotから再生成して判定し、旧byte数を流用しない。既存same-key成功replayは元結果の処理を先に行い、現在容量を成功再送の条件にしない。

上限+1で止めるbounded counting sinkを使う。合計値を0..8,388,609で飽和させればoverflowを避けて判定できる。snapshot全件を巨大な連結bufferへmaterializeする必要はないが、書込対象を切り捨てて上限内に見せることは禁止。取得途中失敗・欠損・計量不能を0 bytesとして扱わない。

既知preclaim超過は既定413/no claim、lock後超過は既定の全business rollback。評価不能/timeout/commit結果不明は既定停止・元key照会に従い、容量判定を失敗確定の根拠へしない。部分Detach/途中commitなし。

<a id="exam-detach-capacity-byte-acceptance"></a>
### 追加受入期待（すべてNOT_RUN）

- B=8,388,607/8,388,608/8,388,609で、それぞれ許容/許容/超過。
- 日本語/補助平面文字、quote/backslash/control character、slash/HTML/U+2028/U+2029を含む同値payloadでbyte数が一致。
- NULLと空文字、missing拒否、ExactNumeric/ID大数、negative sort_order、同instantの異なるtimezone入力のcanonical byte数を確認。
- provenance snapshot-v1/v2の各exact key集合を計量し、未収録をNULL key追加へ変換しない。
- manifestのページ分割/取得順/未採番IDに総byte数が依存しない。
- preclaimからlock後のpayload増加で最終超過なら全rollback、元対象変更なら既定dependency-changed。

<a id="exam-detach-claim-target"></a>
## 5. Job専用claim metadata

選定追加表 lab_job_operation_targets:
- clinic_id bigint NOT NULL; binding_id bigint NOT NULL; PK(clinic_id,binding_id)。唯一FK(binding_id,clinic_id)→operation_bindings(id,clinic_id) RESTRICT。
- schema_version smallint NOT NULL CHECK=2。
- association_version bigint NOT NULL CHECK>=1。target_json/Contentの対象cycleと完全一致。
- action text NOT NULL CHECK='detach'; job_id uuid NOT NULL非nil。
- source_type text NOT NULL（現在device source allowlistをvalidatorで検証、日常利用範囲拡大の意味なし）。
- pet_id bigint NULL CHECK>0。
- target_sha256 char(64) NOT NULL lowercase64hex。
- authorization_targets jsonb NOT NULL：認可済みpreflightで読んだ元examの配列。要素exact {examination_id,medical_record_id,pet_id}、exam ID昇順、nullable関連は明示null、その他はcanonical正ID string。ID集合はtarget_json.examination_idsと完全一致。未成功lookupでも元record/petを現在examから差し替えず解決するためのimmutable anchorであり、命令本文/臨床値の保存先ではない。Applyでは各identityも一致必須、変化ならdependency-changed。
- target_json jsonb NOT NULL：§4.1 exact canonical targetのprojection。全fieldと列値一致、examination_idsの正ID/順序/uniqueをowner validatorで検査。hashはcanonical再serializationで照合し、jsonbの見かけ順へ依存しない。

全列immutable、UPDATE/upsert修復/通常DELETE禁止。既存tenant RLSに参加、新GRANT/roleなし。業務job/exam/petへのFKを持たず、claim TXで業務lockを取得しない。

新bindingだけ認可済みscoped preflightから対象を読む。read context digest一致を確認してclaim。binding＋metadataをClaimMetadataWriterで同一短TX INSERT。一方失敗なら両rollback。claim競合の敗者は勝者metadataを用いる。existing bindingは現在job active状態やwrite grantを先に要求せず、保存target/current viewを使うC06 replay/conflict/lookupへ進む。

<a id="exam-detach-manifest"></a>
## 6. 補償manifestと成功receipt

operation_resultsが唯一のC06成功判定。多件refをC06 string-mapへJSON文字列として埋め込まず、job ownerのimmutable manifestを参照する。新表はreceiptの競合正本でなく、成功効果の固定明細。

lab_job_detach_manifests:
- clinic_id,binding_id bigint NOT NULL、PK(clinic_id,binding_id)、複合FK→lab_job_operation_targets RESTRICT。
- job_id uuid NOT NULL、target_sha256 char(64) NOT NULL、schema_version smallint NOT NULL CHECK=2。
- before_association_version bigint NOT NULL CHECK>=1、after_association_version bigint NOT NULL CHECK=before+1。beforeはmetadata/Contentと一致。
- before_pet_id bigint NULL CHECK>0、before_status text NOT NULL CHECK='persisted'、after_status text NOT NULL CHECK='received'。NULLをCHECKだけで拒否したことにせず、成功manifest/API/replayの固定状態を保存層で保証する。
- recorded_at timestamptz NOT NULL（当TXの成功記録時刻）。

lab_job_detach_manifest_exams:
- clinic_id,binding_id,exam_id bigint NOT NULL、PK(clinic_id,binding_id,exam_id)、複合FK→manifest RESTRICT。
- retraction_id bigint NOT NULL、job_id uuid NOT NULL。
- before_result_state_version bigint NOT NULL CHECK>=1、after_result_state_version bigint NOT NULL CHECK=before+1（overflowを先に拒否）。
- medical_record_id,pet_id bigint nullable positive（元identityの現在認可anchor）。
- retraction FK(clinic_id,retraction_id,job_id,exam_id)→既存retractions(clinic_id,id,job_id,exam_id) RESTRICT。

manifest target/job/元pet/対象集合とmetadataを必ず照合し、manifestのexam集合はmetadataの全集合と完全一致。0件は明細0行。全列immutable。既存RLS・保存owner規約。新migrationの採番/実ファイルは446/466専有で、本案は適用しない。

C06 success/replay flat ResultResponse:
- result_refs exact {job_id,detach_manifest_binding_id,target_sha256}、全string。
- committed_versions exact={job_association:after_association_version}。canonical正十進文字列の実保存版を返す。各examの固定counterはmanifest明細へ保存し専用GETで返す。replayは元after版で、現在版に置換しない。
- recorded_atはmanifestと同じ成功記録時刻。replayで現在jobカードへ置き換えない。
- EffectSummaryは内部だけ。結果本文/原値/患者情報をC06 refsやログへコピーしない。

manifest GET /api/v1/lab-imports/{job_id}/detach-manifests/{binding_id}
- 成功C06 resultとmanifestの一致が前提。manifest単独の存在を成功扱いしない。
- exact response {schema,job_id,operation_key,target_sha256,before_pet_id,after_pet_id,before_status,after_status,before_association_version,after_association_version,retracted_examination_count,recorded_at}。
- schema=lab-job-detach-manifest-v2、after_pet_id=null、statesはpersisted/received、before/after_association_versionは保存済み正十進文字列。
- retracted_examination_countはcanonical nonnegative decimal string、nullable before_pet_idとUTC RFC3339Nano時刻。全件配列をheaderへ含めない。
- 明細は GET /api/v1/lab-imports/{job_id}/detach-manifests/{binding_id}/examinations?cursor=<cursor>。固定100行/ページ（今回選ぶ技術上限、既存業務上限とは主張しない）、exam ID昇順keyset。response exact {schema,job_id,operation_key,detach_manifest_binding_id,items,next_cursor}。schema=lab-job-detach-manifest-page-v1、items exact entry {examination_id,retraction_id,before_result_state_version,after_result_state_version}、ID/版string。0件items=[]/next_cursor=null。
- cursorはbase64url(no padding)のcanonical JSON {schema:"lab-job-detach-cursor-v1",binding_id:"...",after_examination_id:"..."}。クライアントの権限tokenではない。医院/actorはtrusted context、bindingはURLと一致、after IDは当manifestの実在明細であることを検査し、不一致/不正cursorは400。毎ページ同じ元target現在認可を再検査。署名や新secretを導入しない。immutable manifestをページ分割し、現在job全集合を再queryして混ぜない。
- 受信8KiB上限は存在するが、legacy/fixture/後続writerを含めたjob当たりexam総数の厳密上限は未立証。このためmanifest表示はページングとし、処理全体は[有限capacity](#exam-detach-capacity-budget)で別に制限する。内部補償単位は全件同TXを維持し、ページごとcommitにはしない。
- 本文snapshotはこのAPIへ露出しない。現在認可をlookupと同じ保存元targetから検査する。

<a id="exam-detach-atomicity"></a>
## 7. Lock・transaction・再検査

1. C06短claim後、単一Runnerの最外business TX。GateContext/binding NOWAIT/現在認可replay/unknown判定/savepointは既存契約どおり。
2. lifecycle→必要edit-right→record ID昇順→plan ID昇順→plan dependency→job UUID昇順→exam ID昇順→revision/children→usage/補償rowの安定順。D4/516の順を改変しない。
3. 非lock preflightで関連集合を計画し、lock後にjob/pet/record/active exam/active linkを再読。集合追加で下位lockを持ったまま上位を足さずrollback/replan。**利用者target自体が変化した場合はreplanで新集合を採用せずdependency-changed。**
4. locked jobのassociation_versionと全集合からcanonical targetを再生成しmetadata/Content cycle/digestと一致。現在cycleの自動補完なし。jobはdeviceかつpersisted。received/needs_review/reverted等を旧sourceのnil/no-op成功へ潰さず、未成功の新Applyはdependency-changed。同key成功replayは現在stateを再要求しない。
5. usage_tracking_startedを確認。usage missing marker/未知receipt/既存clinical-use/manual mutationがあれば拒否。各examの現在scope、official/confirmed/legacy seal、revision実在、親保護、image等依存、active plan owner linkを同TXで再照合。plan linkがあれば自動unlinkせず拒否。
6. job lock下で読み直したexam result_state_versionを期待値とし、各D4 ambient participantを一回だけfinish。元parent/itemsの補償snapshotを既存retractions/itemsに保存。保存snapshotはD4 nullable/date/ExactNumeric/版/counterを損なわないclosed projectionとし、旧sourceのDate.Formatや汎用model dumpを正本にしない。
7. soft-deleteとresult_state_version+1を一UPDATE/CAS。exam.job_id、過去revisionとitemsは保持し、FK/補償由来を切らない。全examのうち一つでも失敗なら全rollback。
8. jobをpersisted→received、pet=NULL、finished_at=NULL、persisted_count=0、association_version=before+1へ同TXの一回更新。fingerprint・decoded受信items・原受信時刻は保持。その他の既存count/属性を推測書換しない。
9. manifest＋全明細＋必要既存event/audit＋C06 operation_resultsを同TX。snapshot/audit/manifest/receiptの一つでも失敗ならjob/counter/soft-delete全部rollback。stdout logやnil-safe loggerを必須監査の代用にしない。postcommitのjob card読取失敗を補償失敗として再Applyしない。

補償snapshot exact parent projectionはD4 detailの保存scalarから {id,clinic_id,medical_record_id,pet_id,exam_type_id,doctor_id,job_id,date,performed_date_precision,result_summary,machine,status,current_revision_version,result_state_version,created_at,updated_at,deleted_at}。relation表示object/contextを含めない。item snapshotはD4 ExamResultD4保存scalar（is_assessed等の表示派生を除く）に固定し、元id/order/NULL/ExactNumericを保持。現immutable revisionは変更しない。新clinical officialを捏造しない。

### 7.1 補償snapshot schemaと501 Aとの狭い互換条件

parent_snapshotへ明示schema discriminatorを保存する。新規v1は `schema=lab-exam-retraction-parent-v1` と上記§7のexact保存scalarだけを持つ。501 Aでexamsのresult_source_kind/result_source_details/measured_atが供給されたschema profileでは、**新規補償は必ずv2**を使う。v2のexact keysはv1全keys＋result_source_kind/result_source_details/measured_at、schema=lab-exam-retraction-parent-v2。

v2の3fieldはrequired nullable。result_source_kindはin_clinic/external_manual/legacy_unknown/null、result_source_detailsは保存string/null（空文字とNULLを別に保持）、measured_atは保存timestamptzのUTC RFC3339Nano/null。日だけの値から00:00を作らず、NULLをlegacy_unknownへ補完しない。date/performed_date_precisionは元の保存値と独立に保持し、provenance/dateへ相互補完しない。timestampは元時点を損なわないlossless serializerを使う。

対象examをlockしてD4 counterを読む**同じTX/snapshot**でこの3属性を読み、retraction snapshot・manifest・counter CAS・audit・成功receiptと一緒にcommitする。初official前のunversioned examはcurrent_revision_version=NULLのまま保存する。revisionがないことを理由に属性を落とす、架空officialを作る、revisionへfallbackすることは禁止。revisionがあっても現在aggregateの補償snapshotを過去official値で置き換えない。

互換性:
- 501 Aの3列がまだ供給されていない旧schema profileではv1を使用可能。存在しない3列をNULLだったと主張しない。
- 3列供給後は全てNULLのrowでもv2とし、列を保持できないv1 writerを当該補償入口へ接続しない。起動/配線時に受領したschema profileとadapter capabilityを固定し、SQLのundefined-column捕捉でv1へ黙ってfallbackしない。
- readerは既存無schemaのlegacy snapshot、v1、v2を区別して読む。既存snapshotを書き換えず、legacy/v1で未記録の属性は「未収録」であり既知NULLではない。欠けたprovenanceを現在exam・job・revisionから後付け補完しない。未知schema又はv2の必須key欠落はcontract failureとして停止する。
- v2 snapshotを解釈する復旧/参照adapterも同版受領する。復元操作の新設・権限付与はしない。旧readerへv2を無言で渡して3fieldを落とさない。

必要依存は501 Aの**該当3列を実際に書くwriterとD4 aggregate reader、446/466のschema profile、当補償v2 snapshot serializer/reader**だけ。Aの外部出典UI全体、Bの感染症summary、mapping臨床承認、他院手入力機能全体をDetachの前提にしない。job-cycleはv2必須のまま、provenance3列未供給時は安全なv1補償snapshotを使うDetachが先行でき、3列を有効化する同releaseでv2 adapterを必須受領する。

追加AC-17（NOT_RUN）: unversioned・初official前・result_source_kind=external_manual、空文字details、NULL details、実保存measured_atとNULLの各fixtureでv2保存→再読が完全一致。v1/legacyは未収録を保持し、v2 reader欠落・未知schema・必須key欠落はfail closed。旧provenance profileのv1補償snapshot先行（job-cycleはv2）と3列供給後v2補償snapshot強制の切替を確認する。revision fallback/架空official生成0件。

<a id="exam-detach-audit"></a>
## 8. 必須監査

既存audit_logsを再利用。job aggregate auditを一件、必要既存exam補償auditを同TXで記録。
- actor_type=staff、actor_id=現在認証staff、clinic_id=trusted clinic。
- resource=既存lab_import、action=lab_import.detach（新consumerラベルとして明示）。mainのAuditLog.ResourceIDはnullable uint64でありUUID非対応。resource_id=NULLとし、job UUIDは下記metadata.job_idへ保存する。型拡張・数値変換・新audit表は不要。必須INSERTがerrorを返せるambient adapterを使う。
- metadata exact {schema_version,operation_binding_id,operation_key,job_id,target_sha256,detach_manifest_binding_id}。
- old_value/new_value exact {status,pet_id,persisted_count,finished_at}。理由は既存固定「検査受信の取り消し」を補償snapshotへ保存し、ユーザー入力理由と偽装しない。
- 全examとcounter/snapshotはmanifestで参照可能。raw bytes/検体ID/原値の一般log出力なし。

<a id="exam-detach-lookup"></a>
## 9. Lookup・unknown・失権

GET /api/v1/lab-imports/{job_id}/operations/{operation_key}?intent=lab-import.device-detach。
- intent allowlistはこの1値のみ。現在actor/clinicのC06 Scopeで検索。
- saved metadataから元jobを解決しURL一致を検査。URLやsuccess refsへfallbackしてtargetを生成しない。未成功bindingも同じmetadataを使う。
- 現在lab-import:viewと元job/元患者・元exam identityの既存scoped閲覧認可を再検査。元補償snapshot/history anchorでsoft-delete後も元対象を検査可能にする。現在新患者への再attachを過去対象への認可根拠にしない。旧患者view失権なら403。現在write grant喪失だけならviewが残る限り照会可。
- metadata欠落/不整合/元target解決不能は403非開示。lookup/replayでrepair/作成しない。
- timeout/parse failure/5xxは元key lookupへ。succeeded/replayedは元manifest参照、in-progressは同key再照会、result-unavailableは停止。not-succeededが確定してから人が対象と現在条件を再確認した同内容same-key retryのみ可。
- 元操作がunknownの間は新key・再Detach・自動attach・Revert・機器再送なし。jobの現在receivedだけで元key成功を推定しない。

既知capacity超過のowner 413 detach_capacity_exceededは[有限capacity](#exam-detach-capacity-budget)に従い、共有C06 enumを拡張しない。

HTTP: succeeded/replayed/not-succeeded=200、operation-in-progress=202、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、not-authorized=403、scope-not-found=404、result-unavailable=503。未認証401・不正JSON/body超過は既存envelope。lookupのstore読取errorはC06がerrorを返す既存経路と区別し、現在target認可後のgeneric 5xxをunknown扱いする。非成功にはrefs/versions/recorded_atを返さない。

<a id="exam-detach-revert"></a>
## 10. Revertとの関係

既存lab_import_revert_receiptsはfixture Revertの旧成功receipt。clinic+idempotency_key unique、job/reason/request_hash/result/retracted IDsを持つ。本案のDetachをそこへ書かず、result_status=receivedで流用しない。旧Revert responseをC06対応済みと呼ばない。

RevertもD4 counter/coordinator/mandatory audit/ambient全exam補償には参加させる。この有限変更は旧receiptのidentity/terminal stateを変えない。旧Revertのunknown回復を新Detach lookupで代行しない。既存keyは既存receipt readerが現在認可を検査して読む。過去receiptからC06 bindingをGET時に自動backfillしない。

RevertのC06同等unknown全体は**本Detach採択で完了としない**。現Revertがunknownを安全に処理できる受領証拠がない場合はその入口だけactivation外。今後統一するなら旧receipt保持/互換lookup/未成功targetを別の有限migrationとしてjob ownerが提出する。D4-JOB-01はDetach成果とRevert counter/lock参加を別受入行にし、Revert unknown残を隠さない。

<a id="exam-detach-writer-closure"></a>
## 11. 必須writer閉包とowner

- job owner（既存297/113/268周辺の実責任、個人claimは未確認）: detach handler/service、job target reader、metadata/manifest、lookup、既存Revert receipt適合調査。
- 501 C: D4 ambient result participant/nullable snapshot/CAS/historical view authorizer受領。D4を再設計しない。
- 516 S1/S1a/E-link: 現在関連集合/順序/coordinator/usage参加の必要subset。リンクを追加・取消してDetach許可を偽造しない。
- 446/466: 3限定追加表＋job association_version列のDDL/model/RLS/FK、既存retraction snapshot保存との整合。D4単体2DDLとは別scope。
- 445/API/FE owner: closed command/context/flat result/lookup/manifest read、旧無key caller閉鎖、generated成果の正規受領。新生成物を手編集しない。
- 489: C06既存Runner/claim/GateContextの採択subsetだけ。別Runnerの新設やC06全体再完了待ちは不要。

必要参加writerはjob→exam membershipを作るattach/persist/reprocess、削除/復元/JobID付替え、関連clinical read/manual mutation/plan-linkと補償。**association_versionの最外TX一回増分が必要**で、job lockと新exam-ID再作成・現集合再読のD4閉包を併せて受領する。遅延失敗unlinkは元cycle期待版を必須再照合する。receive重複skip・unlinked受信保存・mapping draft/LLM全体を待たない。mapping再解決は297の未persist限定を守り、補償済みjob再attachをmapping無制限再計算の抜け道にしない。

<a id="exam-detach-acceptance"></a>
## 12. 有限AC（期待値のみ、全NOT_RUN）

1. 単exam/multiple exam/0 examでjob独立targetが成立。N件は全N補償、manifest完全一致。空配列と取得失敗を区別。
2. 同key replayは元refs/time/manifest、counter+0、snapshot/audit追加0。
3. 同key別job/別target digestは409、元target現在認可がなければ403。
4. response loss→同keylookup、job現在状態だけで成功推定しない。unknownでは新key禁止。
5. detach→reattachで新exam IDs。旧成功keyは元結果だけ、旧未成功key/旧contextは新cycleを消さない。空集合→detach→同pet再attachの0件cycleも同じ。
6. 2人別key同target並行では先行だけ成功。後行はstate/集合差でzero-write。
7. clinical read/manual PUT/plan linkとDetach競合、job→examとcoordinator順で直列化。使用/関連あり拒否、auto-unlinkなし。
8. persisted以外/非device/unknown usage/保護版/不明scopeを旧nil成功へ潰さない。
9. N件中1件のCAS/snapshot/audit/manifest/result保存失敗で全rollback。counter/soft-delete/job状態の片落ちなし。
10. preflight後にrecord/job/active link集合変化。上位後付けlockなし、target変更を自動採用なし。
11. ID/ExactNumeric/NULL/date/履歴版をsnapshot保存→再読で保持。元itemsとofficial revision不変。
12. 原患者の失権・新患者だけview・別医院/別actor・metadata欠落でlookup/manifest非開示。soft-delete後の正当な歴史viewで元結果のみ開示。
13. no-key/no-context旧caller、direct HTTP、nested helper経路がunguarded成功しない。
14. received job再読/board/refetchはDetach成功receiptを生成しない。postcommit card失敗でもApply再実行なし。
15. fixture Revertがreceived化せずterminal reverted。Detachが旧revert receiptへ保存されず、旧Revertキーを新lookupで成功と偽装しない。
16. source_fingerprint/decoded items保持。未対応値を推測補完・再mapping・新臨床result生成しない。

<a id="exam-detach-dependencies"></a>
追加manifest AC（NOT_RUN）: before_status/after_statusのNULLをNOT NULLで拒否し、persisted/received以外もCHECKで拒否。正常保存の初回/replayは元manifestの非NULL固定状態を返し、欠損を成功応答へ補完しない。

追加cycle AC（NOT_RUN）: 0件のdetach/再attach、遅延失敗unlinkの旧版拒否、複数TXごとの増分、同TX多段変更の一増分、dedup/no-op/read/replay+0、途中失敗のversion/業務/manifest/receipt全rollback、旧draft keyの現在版補完禁止を確認する。

追加capacity AC（NOT_RUN）: exam999/1000/1001、item19999/20000/20001、UTF-8合計8 MiB直前/一致/超過、claim前既知超過の413/claim0、lock待ちの集合増大後全rollback、10秒budget超過と結果不明時の元key照会を区別する。受信frame上限から総数を推測せず、部分補償/切捨て/新key救済なし。

## 13. 真の残依存・最終レビューで確認する点

本案はルート/identity/保存/lock/receiptを具体化した選択案であり、本人への新業務質問は不要。追加3表のmanifest分離とRevert unknownを別有限残へ分ける方針は今回選定した。audit UUIDは既存nullable resource_id=NULL＋metadata.job_idへ固定し、未選定を残さない。

供給依存は対象writerのjob lock閉包、D4 counter/ambient participant、516必要coordinator、現在元target/historical view reader、DDL/API/FE正規成果、および501 Aの3列供給時だけ§7.1のv2 snapshot adapterとの同版受領。全501、全516、全297、実機sample/LLM/臨床mapping承認、全schema、D4 core再選定をブロッカーにしない。

実装着手には既存実行ID/claim・単一編集owner・path/hunk・採択SHA・供給SHA又は同PR合意が必要。文書採択と供給/実装/試験/activationを混同しない。未参加Detach/Revert入口は限定fail closed、独立receive/readは維持。

## 固定リンク
- https://github.com/MinoruSoga/AnimalEkarte/blob/cff4debe2e4620584e3afc4faaabea370afc7d37/docs/design/domains/examinations/result-mutation-contract.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/cff4debe2e4620584e3afc4faaabea370afc7d37/docs/spec/domains/examinations/lifecycle.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/cff4debe2e4620584e3afc4faaabea370afc7d37/docs/design/domains/examinations/technical-spec.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/cff4debe2e4620584e3afc4faaabea370afc7d37/docs/design/shared/operation-participation.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/cff4debe2e4620584e3afc4faaabea370afc7d37/docs/work/emr/emr-516-517-plan-owner-link-contract-20261021.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/cff4debe2e4620584e3afc4faaabea370afc7d37/docs/work/emr/emr-297-device-mapping-bridge-contract-20261021.md

## 追補A: 現writerの実pathと供給境界（08:35 UTC再照合）

以下はmain cff4debeのsource観察であり、必要lockが全て供給済みとの主張ではない。全行の閉包証拠をjob ownerが受領するまで該当Detachだけactivation不可。

| 実path / symbol | 現観察 | 必要供給 |
|---|---|---|
| lab_device_receive_service.go Attach → lab_device_receive_repository.go LockJobByID → AttachJobPetID → PersistLinkedJob | Attachはambient TXでjob FOR UPDATEを先取り | 現coordinator前段参加とlock後再検査を受領 |
| lab_device_receive_service.go receiveOne → persistLinkedOrUnlink → lab_device_exam_persist.go PersistLinkedJob/persistLinkedJob | receive保存後の独立persistはFindJobByIDで開始。全callerがjob lock済みとは言えない | PersistLinkedJob入口でambient job-lock capability必須、単独入口はcoordinator→job lock。membership追加前にcurrent再読 |
| lab_device_receive_service.go persistLinkedOrUnlinkの失敗後UpdateJobPetID(nil) | 別TXの直接更新。先行attach/detachの新状態を消さない保証未受領 | job lock＋元target/失敗対象の再検査、unknownならunlinkしない。受信自体と必要補償subsetを分離 |
| lab_import_examination_service.go PersistExam/persistExam/persistLabExamWrite/PersistBatch → lab_import_examination_write.go writeLabExamGraph | ambient参加はあるがjob lockを自前取得しない | JobID付きmembership writer全callerに同capability。fixture batchもjobを触る部分だけ参加、detach操作をfixtureへ広げない |
| lab_import_revert_service.go DetachDeviceJob/retractExam、lab_import_revert_tx.go revertLabImportInTx → lab_import_repository.go SoftDeleteExamByJob | job UPDATE→exam順は存在、親後取り/旧counterなし | 新coordinator/単一D4 finish、mandatory audit、全exam同TX |
| examination_repository.go Create/Update/Delete、examination_service_create.go、examination_service_update.go | 低層汎用writerは単体examだけで安全を証明できない。D4既存writer matrix対象 | job付きcreate/delete/identity変更をjob→exam参加へ。通常PATCHにJobID入力を新設しない |
| examination_revision_workflow_repository.go restoreMutableExaminationFromRevision/AdvanceRevisionCAS等 | mutable restore/counter対象。通常attachの復活経路ではない | D4 participantと既存history保護、job付きrestoreでmembershipを変えるなら同job lock。補償済みexamの無断復活禁止 |
| lab_device_receive_repository.go CreateJobWithItems/AttachJobPetID/UpdateJobPetID、lab_import_repository.go Update/CompareAndSetStatus | SQL update自体の行lockだけでは、先行read→exam insertの閉包にならない | 公開低層helperのambient capability受入、別TX/未guard呼出しを排除 |
| emr-297-device-mapping-bridge-contract-20261021.mdのreprocess participant | 契約は未persist限定、実装未供給 | その限定writerが既存jobのmembership/petを触る時のみ同capability参加。draft/LLM全体は依存外 |

追加callsite/運用import/restoreの全grep監査を今回完了とはしない。D4既存closure受入票へ実path・caller・供給SHAを対応させる。未確認を「既にlockあり」としない。

新ID前提の根拠: D4 §4.1は新規exam+itemsを新aggregate counter=1とし、§1.5は旧examを上書きしない。固定source lab_import_examination_service.go persistExamはID未設定のnew model.Examinationを作り、lab_import_examination_write.go writeLabExamGraphがexamRepo.Createへ渡す。既存ID復活/上書き処理ではない。duplicate skipは新examも既存exam書換も行わずD4で+0。これは設計根拠＋現在create実装の照合であり、全raw writerの閉包完了を意味しない。

## 追補B: 既存補償表との重複比較

既存retractionsはclinic/job/exam/retraction ID・actor/reason・parent snapshotを持ち、itemsはretraction IDへ紐づく。しかし操作binding/key・job全体target集合・全件成功・各D4 commit counterを結ぶ列はない。同jobにはdetach→reattach→detachの複数回があり、job_id/created_atによる推定groupingは不可。lab_import_revert_receiptsだけには旧fixture keyと全exam IDsがあるが、Detachを保存しておらずterminal revertedの責務である。

従って補償本文snapshotは既存2表を再利用し、新manifest明細にはretraction IDと版/identityだけ保存する。本文/itemsを二重コピーしない。3追加表は未成功target、成功aggregate header、成功exam-ref集合の異なる寿命/制約を分離する最小選択案。既存retractionsへのbinding_id付加＋zero-exam headerで2表へ減らす案もあるが、既存fixture補償表のnullable/legacy migration範囲を広げるため今回不採用。C06 operation_resultsだけへ可変長JSON stringを詰める案はtyped boundary/固定ページングを失うため不採用。
