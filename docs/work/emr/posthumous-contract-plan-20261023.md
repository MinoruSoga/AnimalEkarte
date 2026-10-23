# 死亡後の生前事実入力：設計作業案と着手残条件（2026-10-23）

**計画整備。製品実装Ready・公開契約採択・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.posthumous-contract-plan-20261023 |
| 種別 | 未採択の設計作業案 |
| 決定状態 | 既存vital訂正先行の準備順とT06共用医師reader再利用方向を2026-10-23選定。新schema/route/intent/総lock合成は未採択 |
| 実装状態 | 候補の供給・接続・全writer閉包は未完 |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23（資料受領・静的整形） |

[今回の索引・判定軸](emr-limited-audit-plans-index-20261023.md#limited-audit-status)を参照。以下の監査範囲・件数は受領した固定基準の報告であり、今回再度の全件照会を行ったという意味ではない。

本文の新schema/route/intent/取得順の断定形は候補内の整合条件で、公開だけで採択・着手許可にはならない。今回選定した準備順と共用reader方向は以下に限定し、必要供給・実装実行ID・runtime Readyとは分ける。T06の内部選定範囲は[共用接点の技術契約](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-physician-dependency)へ接続する。

状態：**準備順の限定選定と設計作業案。新しい公開wire・保存schema・総取得順は未採択、実writeは未供給。**

<a id="posthumous-plan-scope"></a>
## 1. 既決仕様の再利用と候補の境界

今回の基準mainは `b81abc26b3657bd662475a1f55b912672a24a265`。当初171b565の固定参照は履歴として保持し、最新source入力とroot選定を以下へ反映する。495最新決裁 `7b95529a-f045-49da-b5ae-aab5a315e2bb`（12:27 UTC）のP1を優先する。旧finalized拒否だけを除去し、既存の医院・有効親・ambient TX・正規lock順・CAS・監査・目的別保護を維持。494未参加はcaller別残差であり、全494先行・子writer全面停止へ戻さない。495 AはBE＋api.yaml同PR、codegen本人実行NOT_RUN、FEは549追跡。D2は最初のestimate_service変更unitへ同梱。

331の最新コメント `7862ee05-9e27-42fe-820f-c4c7dd2accaf` にある技術残件を扱う。生前分の後日入力・誤記訂正の必要性は再質問しない。495 C0/C1/C2の採択本文は inquiry / clinical_plan / diagnosis の3資源に限定し、vital等を追加しない。今回の設計作業案はP0共通接点とP1バイタル適合packetまで。P2〜P4は末尾の未選定依存表に残す。

**既決の生前根拠：** D03 `medical-record-deceased-retrospective` と `AC-medical-records-retrospective-existing` は、実診療日・入力時刻・入力者・確認医師・元記録関係、理由・医師確認・履歴を要求する。既存childがない後日入力の業務許可はあるが、任意の自由文や未定義attestationだけで確認を代替する承認はない。495のsource-less documented_contentも3資源限定なのでvitalの根拠DTOに転用しない。対象患者・生前時点不明なら停止し、情報確認を行う。死亡日と同日のdateだけから前後を推定しない。

<a id="posthumous-plan-shared-dependencies"></a>
## 2. P0：既存供給と、足す必要がある正確な面

| 接点 | そのまま使う既存型／方式 | 限定差分・必要出力 |
|---|---|---|
| 死亡判定 | PR845（491 D6）head `23e944bd0ffad7c29b82b2b04e8771581deda2f4` の `PetLifeStateReader.ReadPetLifeState(ctx, clinicID, petID) (PetLifeState,error)`。alive/deceased、unknownはerror。ambient TX必須、immutable clinic binding解決→common subject FOR SHARE→lock後別statementのcurrent revision再解決 | reader再実装なし。同一ownerの既存解決処理から、後日入力用に下記snapshotを返す追加read面だけを供給。PR845自体は未mergeで、供給済みとはまだ書かない |
| 死亡snapshot（追加設計面） | local petsは医院利用identity、死亡正本はcommon_pet_basic_revisions。local pets.versionを死亡版へ代用しない | 内部 `ReadPetLifeSnapshot(ctx, clinicID, localPetID)` の出力候補：local_pet_id、binding_id、common_pet_id(UUID)、subject_version、basic_revision_id、basic_revision、life_state、deceased_at(nullable)。現行D6と同じlock/再解決/unknown規則。死亡理由・他院本文は不要。HTTP公開DTOではない |
| 編集fence | 494 §3の単一record bodyは平坦な `operation_key,edit_session_ref,edit_epoch,owner_expected_version`。新wrapper/header/bearerなし。session UUID、epoch/版は正規十進文字列。現在actor/grant＋held/holder/session/epoch一致 | vital owner版とrecord版を混同しない。494通常owner writeの `owner_expected_version` はvital訂正ではvital.versionに対応させる設計差分。495複数record用record_fencesと同じrecord版の意味へ流用しない。vital createの不存在/集合版は既存採択なしなので別途一行で選定が必要 |
| C06 | 短claim→必要Gate→binding NOWAIT→現在認可付きreplay/unknown→savepoint→Apply→成功result。ContentとCAS Preconditions分離、VerifiedCommand Q、既存12status | 新しいreceipt機構なし。vital専用intent/target metadata/lookup allowlistはまだ未選定。494 medical_record_operation_targetsはrecord anchorだけなので、vital target ID/元版/確認snapshotを保存できると誤認しない |

PR845はstateだけを返すため死亡時刻や版の証拠を返したことにはならない。一方D6のsubject SHAREとlock後再解決は既に具体化されており、この部分を抽象的な『合意待ち』へ戻さない。

**lock順の解決済み部分：** 494 §4.2/§5.1は必要Gate→binding→既存lifecycle保護→record_edit_rights→medical_records→owner/依存子安定順。claimに業務FK/row lockを持ち込まない。PR845のD6内部はcommon subject SHARE→別statement再解決。この両者は再設計不要。

**残る一点：** common subject SHAREを上記lifecycle保護のどの物理lockの前後へ挿すかは、死亡/取消の共通revision writerがlocal pet/recordを同時取得する順と合わせる必要がある。mainの `pet/repository.go RecordDeath/ClearDeath` と `lstep/lstep_lifecycle_service.go` は旧local状態参照が残るため、現sourceから架空の総順序を確定できない。P0の出力は「死亡・取消・common basic更新writerの実symbol＋全取得順」と「その同じ順でのvital参加位置」**各1行**。新しい全ownerロックや全予約gateは要求しない。供給済みwriterがなければ『死亡writer adapter未供給』と記録し、この限定実writeだけ保留する。491 D6 readerをもう一度作らせたり、331から別死亡writerを作ったりしない。

<a id="posthumous-plan-writer-order-facts"></a>
### b81固定sourceで確認したP0取得辺と未調停部分

| 既存surface | 観測したsymbol・取得順 | 供給/未確認の区別 |
|---|---|---|
| legacy死亡/取消 | lstep_lifecycle_service.HandlePetDeath / HandlePetRevival：WithTx→local pet UPDATE lock→IsPetDeceased→RecordDeath/ClearDeath→同TX audit→commit後tag処理 | 旧local列参照の移行残。common revision writerへ接続済みの証拠でなく、生前vitalへpostcommit tagを持ち込まない。Revivalの理由inputも未供給 |
| direct pet repo | RecordDeath/ClearDeathはDBOrTxで旧3 life列を条件UPDATE。Update→legacyLifecycleTransition等のhelperも残る | repo自体に必須ambient/common revision/CAS/C06/監査なし。helper存在から公開route到達を捏造しない |
| 通常pet更新/登録 | UpdateAndFind：local pet UPDATE→必要owner/insurance SHARE等→更新。登録はowner→master→pets INSERT | 通常基本情報/新患者の別owner。common初期revisionや死亡実装として採択しない |
| common revision保存 | 001:10816–10838のBEFORE INSERT guardがcommon subject UPDATEで版検証、deferred publishがoperation_resultsを要求しpointer更新 | DB側の直列化は実在するが、application death/undo/basic adapterの全取得順・供給SHAは未受領 |
| D6 reader（845固定headのみ） | binding非lock解決→common subject SHARE→別statementでlock後current revision再解決→状態分類 | readerの再実装不要。ただし時刻・binding/subject/revision版snapshotは未供給。open/unmergedでmain採用済みとはしない |

[D04正本の未来予約取消](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/docs/spec/domains/owners-pets/lifecycle.md#pet-death-future-reservations-cancel-20261022)は**必須**。死亡時の未来予約を理由/actor/time付きでcancelledにし、取消時も自動再開しない。subject直列化kernelだけを全死亡workflowと呼ばず、任意の手動取消へ降格しない。取消participant/orchestratorの必要reservation gate、lock順、原子性/完了receipt、unknown照合境界は未供給・未選定として残す。新通知種別は追加しない。

common_pet_basic_revisions INSERTはsubject、source clinic、species、recording staff、local pet/clinic、operation binding/clinicへの[FK](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/migrations/001_init.sql#L10774-L10785)参照lockも発生する。subject UPDATEだけの「単独lock」と扱わない。845のSwapCagesにあるlocal pet UPDATE→D6 subject SHAREと、subject UPDATE→local pet FK KEY SHAREの逆辺は、当該participantの順序付き先行取得又は適合で調停が必要。これはsourceからの取得辺分析で、deadlockを実行観測した報告ではない。全ownerへの新global lockは要求しない。

<a id="posthumous-plan-current-vital-callers"></a>
### 現vital callerと履歴の限界

| surface | 固定sourceの現状 | 専用訂正への不足 |
|---|---|---|
| record POST/PATCH | vitalService→WithTx→medical_records UPDATE親lock→owner SHARE/結合pet・owner SHARE→必要staff/assignment→repo→audit。PATCHの元rowはTX前読取 | 同一lock下のfresh before、必須vital CAS、life snapshot/fence、確認行為、C06/履歴が未接続。結合SHARE queryの行取得順をアプリが制御済みとしない |
| record DELETE | TX前row→親lock→clinic/id soft-delete→旧値audit、actor=nil、期待版なし | 初回訂正sliceへ遡及DELETEを含めない。source事実を新しい削除許可にしない |
| direct VitalRepository | Create/Update/DeleteはDBOrTx、親/fence/死亡/監査/ambient必須を自ら強制しない | 内部参加capability/入口guardと到達経路の限定台帳が必要。既存関数を安全participantと呼ばない |
| dose参照 | treatmentService.resolveDoseWeight→FindByMedicalRecordID | readonly consumer。新vital writerに数えず、訂正で旧dose snapshotを再計算しない |
| daily同表 | dailyRecordService.AddVitalRecord→hospitalization UPDATE→owner/pet検証→staff/assignment→FindOrCreateByDate→CreateVitalRecord→再読 | 親はdaily_records/hospitalizations。同表という理由でrecord edit fenceを適用しない。P4の別契約/activation境界 |
| copy/Undo | 指定route/source・索引検索で専用production symbolは未特定 | 全repo不存在とはしない。494 inquiryのcopy/Undo記載をvital API実在としない |
| import/maintenance | ApplyCutoverCommittingEachTable/importOneCutoverTable/copyCutoverTableはsession advisory/per-table TX・table lock・COPY/staging。旧Import/importWithBeginはdisposable target限定のdeleteDemoGraph/COPY | オンラインservice/C06を迂回する実maintenance経路。非並行運用又は別明示契約を受領し、オンライン対応済みとしない。synthetic_closing_fixtureのvital DELETEもsynthetic-onlyの別境界 |

既存audit diffの8項目はtemperature,heart_rate,respiration_rate,weight,weight_unit,recorded_at,staff_id,notes。版・完全before/after・認証確認者・source/life snapshot・理由・entry metadata・binding/keyを網羅しない。audit repositoryのCreateTx ambient必須とSQL失敗rollbackは再利用するが、optional JSON serializerのnil化では必須履歴の完全性を証明できない。現vital GETは現在値で、取得範囲では専用履歴reader/routeを未特定。既存sinkを不変履歴・認可reader供給済みにしない。

<a id="posthumous-plan-vital-packet"></a>
## 3. P1：カルテ内バイタルの適合packet（現在着手できる準備）

**選定した準備順：** 最初の有限sliceは、medical-record-ownedの既存vitalの訂正。必須の期待vital版、同患者/record、lock下の新鮮なbefore、信頼できる保存recorded_atが解決済みdeceased_atより厳密に前、認証済み医師の明示確認、完全不変履歴とC06 outcomeを必要とする。recorded_atは実施時点であり入力時刻と区別する。timestampが存在するだけで過去の正確さを証明せず、由来/精度不明・死亡時刻不明・dateだけの同日等は根拠確認の後続packetへ保留する。既存childなしの後日入力や正当な不確かな生前事例の業務許可を削除したのではない。遡及DELETE、会計/在庫/薬量の自動再計算は含めない。

対象は既存medical_recordに紐づくvital。通常routeは POST `/medical-records/:id/vitals` と PATCH `/medical-records/:id/vitals/:vitalId`。今回、新しい公開route/intentは採択しない。API拡張の前に次の4差分を既存331 checklistへ固定する。

| 項目 | 実sourceで閉じたこと | 専用入力をReadyにする最小残条件 |
|---|---|---|
| CAS | `vital_repository.Update` はDBOrTxを使い、version指定時だけCAS、version+1。repo自体はambient必須・親lock・死亡/fenceを強制しない。旧UpdateはTX前vitalをbeforeとして取得する | 専用訂正adapterはversion必須でlock下にrowを再読し、新鮮なbefore・clinic/record/petを照合。direct repoを安全な参加済み入口としない。create不存在/集合版は後続設計でrecord.versionを代用しない |
| payload/clear | 現 `vital_request.go` updateのポインタ型は省略とJSON nullを識別できず、JSON nullをclearとする選定根拠はない。`notes` createのnil→空文字は既存規則 | 専用wireのclearを勝手に新設しない。まず現owner updateと同じ非null値変更に限定した有限案か、presence-aware clear型の追加を明示選定する。旧値不明を補完しない。臨床上限/薬量値は追加しない |
| 監査/履歴 | `auditVitalTx`はresource=vital、OldValue/NewValue、metadata medical_record_idを同TX保存。create actorはStaffID。既存update diffは8臨床fieldだけで同値は監査なし。SQL失敗はerrorだが、optionalAuditJSON/MarshalJSONはserialization失敗をnilへ落とす | 既存audit_logsは候補sinkで完全履歴/不変保持/認可付き履歴reader供給の証拠ではない。必須typed payloadをmutation前にserialization検証し、不完全ならzero-write。全before/after・版・認証actor・確認行為・理由・source/life snapshot・binding/keyを同TX保持する型/readerは未選定。既存全監査のbest-effort規律を一括変更しない |
| 生前確認/再送 | D03の理由・医師確認・原記録/以前保存版保持は既決。495のproposal/receipt方式自体は採択済みだが対象外資源 | vitalについて確認者・確認対象値/版を不変に結ぶ保存先/typed command/target anchor/lookupを選定する。確認者IDの束縛だけでは確認行為の証拠にならず、下記Ready条件2のserver適格性reader/権威ある確認証跡を必要とする。自由文やclient confirmedだけでは代替しない。495 C0/C1にvitalを追加しない。確定後だけchild変更＋履歴＋audit＋C06成功receiptを同TX、unknownは元key照会 |

**Ready前の残条件を4成果に閉じる：**
1. P0のlife snapshot出力と死亡writer/child総順序1行が供給SHA付きである。
2. vital専用「生前実施内容・元事実/元版・理由・現在確認医師」を束縛するclosed commandと耐久保存先が一つに選定され、source_ref/attestation等の未定義汎用objectがない。既存factあり訂正とchildなし後日入力は混ぜず、先行する方を明記する。 server側の現在医師資格/医院権限readerと、確認対象患者・元事実/版・実施時点・変更内容・理由へ束縛した認証済み確認操作の耐久receipt（又は既存正本に定義された同等の権威ある確認証跡）を必須供給として選定・受領する。入力actorと確認医師が別IDの場合も、client doctor_id/booleanだけで代行確認を成立させない。確認医師の認証された行為と現在適格性を検証できなければ保留する。新しい医師資格や代行権限は本案で付与しない。
3. その選定commandに対するexact method/path/body/response、494平坦fence、mandatory vital CAS又はcreate不存在条件、C06 scope/intent/元target lookupを1表にする。現時点は未選定であり『closed API』とは呼ばない。
4. 既存audit_logs再利用のclosed before/after＋metadataと履歴read、通常writer迂回排除一覧を揃える。履歴保証を既存sinkで満たせない具体項目が出た場合だけ、該当列/制約の限定DDL差分へ進む。

これは無期限の『各ownerと合意』ではなく、上記4つの欠落成果を331安全adapter責務で作成・既存供給ownerへ受領する条件。新規業務採否を本人へ聞き直す条件ではない。現時点で進められるのは既存型への適合表、current writer/caller一覧、監査差分、固定fixture設計。未選定の専用write実装や有効化はまだ開始可能と表示しない。

医師readerは[T06と同じ既存predicateの共用方向](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-physician-dependency)を使う。AssertMedicalRecordDoctorInClinicはactive/nondeleted staff・staff_type=doctor・同院assignmentを照合するが、default doctorや任意license/occupationは承認証拠ではない。staff/reservation単一ownerからdesignation写像/legacy既定の検証根拠、ambient必須adapter、現在権限・保護順を受領する。適格性と認証済み確認行為の耐久証跡は別で、通常PATCHやclient doctor_id/booleanで確認済みにしない。T03独自の医師制度・grant・第二reader方式を作らない。

<a id="posthumous-plan-finite-supply"></a>
### 有限owner・供給物・合成AC

| owner | 必要出力（現在未供給/未選定） | AC・停止範囲 |
|---|---|---|
| 491/504＋死亡/取消統合owner | 実common writer symbol/取得順・FK辺・D04取消participant/必要gate/完了receipt、D6 snapshot追加面の固定SHA | 死亡/取消と訂正の両先行順、local-first/FK/SwapCages適合。欠ける専用writeだけ保留しP1通常編集や仕様整備は止めない |
| shared staff/reservation owner（T06共用） | 既存predicateの技術写像、designation対応/legacy根拠、ambient/現在適格性・保護順 | actor≠担当医、欠落/失権/未設定、client偽確認拒否。新資格付与なし |
| 331/494/489 | 既存vital訂正用exact method/body/result/intent・元target/確認anchor・lookup allowlist | mandatory vital版、source版とretry CASの区別、同key応答断/unknown。新routeやanchor型は今回未採択 |
| 331＋audit owner | mutation前必須serializer、完全不変履歴/同TX保存・認可付き履歴read | serialization/SQL/receipt失敗zero-write、担当者≠認証actor、元版再読。audit_logsは候補sink |
| vital/daily/import各owner | このsliceへの実caller参加とmaintenance非並行又は別契約、切替receipt | direct repo迂回、daily別親保持、import/COPY境界。全owner Done待ちでなく必要subsetだけ受領 |

上表は供給SHA・正式実行ID・実装claim受領済みではない。確認sourceや元targetを成功audit/receiptだけから解決せず、失敗bindingでも不変anchorと現在認可が必要。494のrecord-only anchorにvital/確認情報が既存保存済みとはしない。合成ACもコード/実行はNOT_RUNで、331全体Readyの判定ではない。

通常writer参加表は vital create/update/delete、直接repository caller、日次AddVitalRecord（同表別親）、copy/Undo/import経路、死亡/取消/common basic更新を列挙。実在symbol・親種・必要fence/CAS・audit/receipt・切替状態を各1行にし、同表だから日次へrecord fenceを掛けない。通常finalized解除は本人P1どおり進め、これを死亡gate解除に転用しない。

<a id="posthumous-plan-other-owners"></a>
## 4. P2〜P4は未選定依存一覧に留める

| 対象 | 既存方針／今回残す限定不足 |
|---|---|
| P2 健診 | 500本体/結果の段階成功・save_attempt/stage_result/result_revision、PDF独自保護を維持。専用生前command、checkup版/result head/package field集合、段階lookup対応が未選定。全体1TX化・親の架空自動生成・新route選定を今回しない |
| P3 接種 | 接種事実と次回予定/証明書editionを分離。nullable親の認可/fence、vaccination CAS、確認済み生前command、旧edition不変/訂正連携が未選定。未来予定・通知を自動起動しない |
| P4 日次 | 現routeはcontainer作成とvital/care/staff appendで子PATCHなし。daily_recordsにversion列なし。stable entry型・元版/訂正履歴・医院timezoneを持つ生前時点・append再送契約が未選定。存在しない版/PATCHを想定せず、入院/退院課金を混ぜない |

<a id="posthumous-plan-evidence"></a>
## 根拠・検証境界

今回b81の固定source追加根拠は[旧死亡/取消service](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/internal/lstep/lstep_lifecycle_service.go)、[pet repository](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/internal/pet/repository.go)、[common guard/publish](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/migrations/001_init.sql#L10816-L10838)、[取消の不足](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/docs/design/domains/owners-pets/technical-spec.md#pet-revival-target-gap)、[vital repository](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/internal/medicalrecord/vital_repository.go)、[既存audit serializer](https://github.com/MinoruSoga/AnimalEkarte/blob/b81abc26b3657bd662475a1f55b912672a24a265/backend/internal/audit/service.go)。報告は固定sourceの有限観測で、全repo探索・実DB・実行の独立再検証ではない。

[331](https://app.plane.so/baritechllc/browse/EMR-331/)本文/全3コメント、[495](https://app.plane.so/baritechllc/browse/EMR-495/)本文/最新決裁を読取。旧171b565時点の固定参照として [494契約](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-494-edit-draft-contract-20261020.md) §§3–5、[495契約](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-495-consultation-correction-contract-20261021.md) §§0.2–0.3、[D03正本](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/medical-records/lifecycle.md#medical-record-deceased-retrospective)、[PR845](https://github.com/MinoruSoga/AnimalEkarte/pull/845)の固定head readerを再読した当初報告を保持する。同旧基準のvital_request/repository/audit/service_create/service_update、各route、model/ERDを限定静的照合した。

受入設計はclient doctor_id/booleanだけの確認拒否、別actorの確認証跡未供給/失権/対象版変更時の保留、認証済み確認操作と現在資格/医院権限の照合、死亡先行/child先行、half-dead/unknown、共通他院revision更新、死亡日時訂正、同日時点不明、元fact/child版変更、医師確認対象変更、失権/fence喪失、audit/receipt失敗、同key並行/異内容/応答断、通常writer迂回、担当者≠actor、旧値/履歴不変、副作用0を固定する。今回は全てNOT_RUN。全repo writer閉包、runtime Ready、臨床受入、331/495全体完了は認定しない。
