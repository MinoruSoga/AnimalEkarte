# 死亡後の生前事実入力：設計作業案と着手残条件（2026-10-23）

**既存record所属vital訂正の限定技術契約と計画整備。製品runtime Ready・API供給・実装・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.posthumous-contract-plan-20261023 |
| 種別 | 選定済み既存vital訂正契約への接続と後続設計計画 |
| 決定状態 | 既存vital訂正先行/共用医師readerを維持し、2026-10-24に専用wire/intent/2不変表/lock順を限定選定。実供給と全死亡workflowは未完 |
| 実装状態 | 選定契約の供給・接続・全writer閉包は未完 |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-24 JST（限定技術契約の選定反映） |

[今回の索引・判定軸](emr-limited-audit-plans-index-20261023.md#limited-audit-status)を参照。以下の監査範囲・件数は受領した固定基準の報告であり、今回再度の全件照会を行ったという意味ではない。

既存record所属vitalの訂正について[選定技術契約](../../design/domains/medical-records/posthumous-vital-correction-contract.md#posthumous-vital-scope)を正本とする。専用wire/intent・最小anchor/完全成功履歴・必須reservation gateを含む専用順を選定したが、文書公開と必要供給/実装/実行ID/runtime Readyは別である。P2〜P4、childなし入力、別職員準備と医師確認、時点不明の根拠packetは未選定の後続として保持する。

<a id="posthumous-plan-scope"></a>
## 1. 既決仕様の再利用と候補の境界

今回の追加契約の基準mainは `7dad9adc7f066fab5cc734dc8ff7e125a80d67d4`。以下のb81 source表は先行観測として保持する。当初171b565の固定参照は履歴として保持し、最新source入力とroot選定を以下へ反映する。495最新決裁 `7b95529a-f045-49da-b5ae-aab5a315e2bb`（12:27 UTC）のP1を優先する。旧finalized拒否だけを除去し、既存の医院・有効親・ambient TX・正規lock順・CAS・監査・目的別保護を維持。494未参加はcaller別残差であり、全494先行・子writer全面停止へ戻さない。495 AはBE＋api.yaml同PR、codegen本人実行NOT_RUN、FEは549追跡。D2は最初のestimate_service変更unitへ同梱。

331の最新コメント `7862ee05-9e27-42fe-820f-c4c7dd2accaf` にある技術残件を扱う。生前分の後日入力・誤記訂正の必要性は再質問しない。495 C0/C1/C2の採択本文は inquiry / clinical_plan / diagnosis の3資源に限定し、vital等を追加しない。今回の設計作業案はP0共通接点とP1バイタル適合packetまで。P2〜P4は末尾の未選定依存表に残す。

**既決の生前根拠：** D03 `medical-record-deceased-retrospective` と `AC-medical-records-retrospective-existing` は、実診療日・入力時刻・入力者・確認医師・元記録関係、理由・医師確認・履歴を要求する。既存childがない後日入力の業務許可はあるが、任意の自由文や未定義attestationだけで確認を代替する承認はない。495のsource-less documented_contentも3資源限定なのでvitalの根拠DTOに転用しない。対象患者・生前時点不明なら停止し、情報確認を行う。死亡日と同日のdateだけから前後を推定しない。

<a id="posthumous-plan-shared-dependencies"></a>
## 2. P0：既存供給と、足す必要がある正確な面

| 接点 | そのまま使う既存型／方式 | 限定差分・必要出力 |
|---|---|---|
| 死亡判定 | PR845（491 D6）head `23e944bd0ffad7c29b82b2b04e8771581deda2f4` の `PetLifeStateReader.ReadPetLifeState(ctx, clinicID, petID) (PetLifeState,error)`。alive/deceased、unknownはerror。ambient TX必須、immutable clinic binding解決→common subject FOR SHARE→lock後別statementのcurrent revision再解決 | reader再実装なし。同一ownerの既存解決処理から、後日入力用に下記snapshotを返す追加read面だけを供給。PR845自体は未mergeで、供給済みとはまだ書かない |
| 死亡snapshot（追加設計面） | local petsは医院利用identity、死亡正本はcommon_pet_basic_revisions。local pets.versionを死亡版へ代用しない | 内部 `ReadPetLifeSnapshot(ctx, clinicID, localPetID)` の出力候補：local_pet_id、clinic_binding_id、common_pet_id(UUID)、subject_version、basic_revision_id、basic_revision、life_state、deceased_at(nullable)。現行D6と同じlock/再解決/unknown規則。死亡理由・他院本文は不要。HTTP公開DTOではない |
| 編集fence | 494 §3の単一record bodyは平坦な `operation_key,edit_session_ref,edit_epoch,owner_expected_version`。新wrapper/header/bearerなし。session UUID、epoch/版は正規十進文字列。現在actor/grant＋held/holder/session/epoch一致 | vital owner版とrecord版を混同しない。494通常owner writeの `owner_expected_version` はvital訂正ではvital.versionに対応させる設計差分。495複数record用record_fencesと同じrecord版の意味へ流用しない。vital createの不存在/集合版は既存採択なしなので別途一行で選定が必要 |
| C06 | 短claim→必要Gate→binding NOWAIT→現在認可付きreplay/unknown→savepoint→Apply→成功result。ContentとCAS Preconditions分離、VerifiedCommand Q、既存12status | 新しいreceipt機構なし。vital専用intent/target metadata/lookup allowlistは[追加契約](../../design/domains/medical-records/posthumous-vital-correction-contract.md#posthumous-vital-command)で選定、runtime供給は未完。494 medical_record_operation_targetsはrecord anchorだけなので、vital target ID/元版/確認snapshotを保存できると誤認しない |

PR845はstateだけを返すため死亡時刻や版の証拠を返したことにはならない。一方D6のsubject SHAREとlock後再解決は既に具体化されており、この部分を抽象的な『合意待ち』へ戻さない。

**lock順の解決済み部分：** 494 §4.2/§5.1は必要Gate→binding→既存lifecycle保護→record_edit_rights→medical_records→owner/依存子安定順。claimに業務FK/row lockを持ち込まない。PR845のD6内部はcommon subject SHARE→別statement再解決。この両者は再設計不要。

**専用参加位置は追加選定済み：** [必須予約gate→binding→local pet SHARE→common subject SHARE→fence→record→vital→owner/staff/FK NOWAIT](../../design/domains/medical-records/posthumous-vital-correction-contract.md#posthumous-vital-locks)。共通revision writerのsource-local KEY SHARE先行とdeath/undoの必要local UPDATE先行も[有限適合](../../design/domains/medical-records/posthumous-vital-correction-contract.md#posthumous-vital-common-writers)へ選定した。下表のFK/SwapCages辺は供給・交錯受入残として保持し、全死亡/取消writerの総順序がruntime供給済みとはしない。D04未来予約取消は必須で、subject-only kernelを全死亡workflowと呼ばない。

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
## 3. P1：選定した既存vital訂正と供給残

[専用訂正契約](../../design/domains/medical-records/posthumous-vital-correction-contract.md#posthumous-vital-scope)を正本とする。同院・同record/pet・daily親なしの既存未削除vitalだけ、同一の認証済み適格医師が明示confirm-and-correctを行う。元/訂正後の既知実測instantは現在死亡instantより厳密に前。default now()だけでは根拠にならず、医師の明示操作を元版/digest/差分/reason/life basisへ束縛する。別actorのdoctor_id/booleanで代行確認を作らず、新資格/権限なし。

選定範囲は[context/専用POST](../../design/domains/medical-records/posthumous-vital-correction-contract.md#posthumous-vital-wire)、[C06 scope/Content/Q/claim/lookup](../../design/domains/medical-records/posthumous-vital-correction-contract.md#posthumous-vital-command)、[2不変表・strict全履歴/audit・認可read](../../design/domains/medical-records/posthumous-vital-correction-contract.md#posthumous-vital-history)、[専用取得順](../../design/domains/medical-records/posthumous-vital-correction-contract.md#posthumous-vital-locks)である。通常wire全体は今回の成果境界外で、後方互換のための保存・将来改修禁止ではない。対象の既存vitalはデータ条件でありコード保存要件ではない。新wireの臨床数値はexact canonical文字列、全JSON null拒否・省略保持・notes空文字可、clearは後続。source版とretry CASを分離し、成功replayは新確認/新履歴を生成しない。

vital.versionはCAS WHERE＋手動version+1、billingのentity_version_guardとは別。lock下fresh before→完全serializer→mutation/after照合→不変history/必須audit/C06 receiptを同TX保存する。既存audit diff/optional nil serializerを完全履歴供給済みとせず、選定2表は現DDLには未供給。医師readerは[T06共用方向](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-physician-dependency)を維持し、ambient/NOWAIT・designation根拠と現在grant保護は単一owner供給待ち。

<a id="posthumous-plan-finite-supply"></a>
### 有限owner・供給物・合成AC

| owner | 選定済み契約に対する必要subset（未供給） | AC/停止範囲 |
|---|---|---|
| 491/504/D04統合 | D6 life/binding/版snapshot・NOWAIT primitive、source-local prelock・D04必須取消gate/全outcome/交錯receipt | SwapCages対common writer、通常parent-first対訂正、death取消対訂正の両先行順。全死亡workflow完成とはしない |
| shared staff/reservation（T06共用） | designation対応根拠、ambient/NOWAIT predicate、現在grant/assignment保護/source SHA | actor=確認者・担当医以外・失権/未設定・偽確認拒否。新医師制度なし |
| 331/494/489/464/467 | 選定専用participant/fence/Q/元target resolver/allowlist、共通ContextGate/try-gate | mandatory vital CAS、同key応答断/unknown、必須reservation gateはbinding前。既存Q等を未実装へ戻さない |
| 331＋audit/storage | 選定最小anchor/成功history、strict before/after/audit、deferred receipt/private reader | 必須serialization/SQL/result失敗zero-write、全過去値不変read。audit_logs候補選びを再開しない |
| vital/daily/import | 通常create/update/delete/direct caller死亡guard閉包、maintenance非並行又は別契約/切替receipt | daily別親へrecord fenceを誤適用しない。全owner Done待ちではなく必要subsetだけ |

正式実行ID・個人claim・供給SHA・API/DDL/model実装・実DB受入は未受領。Readyは計画/契約状態と実行前提の明示であり、実装完了ではない。現在できるのはこの有限実装単位の引渡し・caller/TX/fixture対応表の準備で、通常writer迂回を残したactivationは不可。全合成ACは[技術正本](../../design/domains/medical-records/posthumous-vital-correction-contract.md#posthumous-vital-acceptance)に置き二重管理しない。

childなし後日入力、別職員準備後の医師確認、時点不明の根拠packetは既決業務許可を保持した別の有限技術成果で、削除/再質問しない。遡及DELETE・会計/在庫/薬量の再計算、復活による救済、通知/予約自動操作は追加しない。495 P1通常編集を全面停止へ戻さない。

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
