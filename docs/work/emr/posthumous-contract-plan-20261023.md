# 死亡後の生前事実入力：設計作業案と着手残条件（2026-10-23）

**計画整備。製品実装Ready・公開契約採択・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.posthumous-contract-plan-20261023 |
| 種別 | 未採択の設計作業案 |
| 決定状態 | 既決仕様の再利用と未選定候補を区別。新schema/route/lock合成案は未採択 |
| 実装状態 | 候補の供給・接続・全writer閉包は未完 |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23（資料受領・静的整形） |

[今回の索引・判定軸](emr-limited-audit-plans-index-20261023.md#limited-audit-status)を参照。以下の監査範囲・件数は受領した固定基準の報告であり、今回再度の全件照会を行ったという意味ではない。

本文の型名・exact key・enum・取得順に断定形を用いる箇所も、候補内の整合条件を述べる。既存正本が明示的に採択した範囲を除き、本記録の公開だけで採択・着手許可にはならない。候補を採択する場合はdesign正本へ配置し、必要供給と実装実行IDを別途確定する。

状態：**設計作業案。既決事項を再利用するが、以下の新しい接点・schema・総取得順は候補であり未採択。**

<a id="posthumous-plan-scope"></a>
## 1. 既決仕様の再利用と候補の境界

基準mainは `171b56575e7ee028e54358dad22ae28574853125`（PR843 merge、再読済み）。495最新決裁 `7b95529a-f045-49da-b5ae-aab5a315e2bb`（12:27 UTC）のP1を優先する。旧finalized拒否だけを除去し、既存の医院・有効親・ambient TX・正規lock順・CAS・監査・目的別保護を維持。494未参加はcaller別残差であり、全494先行・子writer全面停止へ戻さない。495 AはBE＋api.yaml同PR、codegen本人実行NOT_RUN、FEは549追跡。D2は最初のestimate_service変更unitへ同梱。

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

<a id="posthumous-plan-vital-packet"></a>
## 3. P1：カルテ内バイタルの適合packet（現在着手できる準備）

対象は既存medical_recordに紐づくvital。通常routeは POST `/medical-records/:id/vitals` と PATCH `/medical-records/:id/vitals/:vitalId`。今回、新しい公開route/intentは採択しない。API拡張の前に次の4差分を既存331 checklistへ固定する。

| 項目 | 実sourceで閉じたこと | 専用入力をReadyにする最小残条件 |
|---|---|---|
| CAS | `vital_repository.Update` はambient TX、version条件付きUPDATE＋version+1。`Version=nil`はCASを省略する既存互換動作 | 専用訂正adapterはversion必須としてnil分岐へ入れない。ID/版の範囲と既存rowのclinic/record/pet対応を再検査。createの重複防止は後述のoperation anchor設計を必要とし、record.versionを子集合版と偽装しない |
| payload/clear | 現 `vital_request.go` updateのポインタ型は省略とJSON nullを識別できず、JSON nullをclearとする選定根拠はない。`notes` createのnil→空文字は既存規則 | 専用wireのclearを勝手に新設しない。まず現owner updateと同じ非null値変更に限定した有限案か、presence-aware clear型の追加を明示選定する。旧値不明を補完しない。臨床上限/薬量値は追加しない |
| 監査/履歴 | `auditVitalTx`→`AuditTxLogger.LogEntryTx` は既存audit_logsへambient参加し、失敗時error。resource=vital、OldValue/NewValue、metadata medical_record_id。createはStaffIDをactorに使用 | 保存先は既存audit_logsを第一案とし、新revision表を当然の依存にしない。ただし現在の重要項目diffだけでは元保存版の完全性を保証しない。専用操作のbefore/after全許可field・元vital版・認証actor・確認医師/根拠・理由・生前時点・basic revision参照・binding/keyを同TXで追記するclosed audit shapeと認可付き履歴readを選定。create担当StaffIDを認証actorに流用しない |
| 生前確認/再送 | D03の理由・医師確認・原記録/以前保存版保持は既決。495のproposal/receipt方式自体は採択済みだが対象外資源 | vitalについて確認者・確認対象値/版を不変に結ぶ保存先/typed command/target anchor/lookupを選定する。確認者IDの束縛だけでは確認行為の証拠にならず、下記Ready条件2のserver適格性reader/権威ある確認証跡を必要とする。自由文やclient confirmedだけでは代替しない。495 C0/C1にvitalを追加しない。確定後だけchild変更＋履歴＋audit＋C06成功receiptを同TX、unknownは元key照会 |

**Ready前の残条件を4成果に閉じる：**
1. P0のlife snapshot出力と死亡writer/child総順序1行が供給SHA付きである。
2. vital専用「生前実施内容・元事実/元版・理由・現在確認医師」を束縛するclosed commandと耐久保存先が一つに選定され、source_ref/attestation等の未定義汎用objectがない。既存factあり訂正とchildなし後日入力は混ぜず、先行する方を明記する。 server側の現在医師資格/医院権限readerと、確認対象患者・元事実/版・実施時点・変更内容・理由へ束縛した認証済み確認操作の耐久receipt（又は既存正本に定義された同等の権威ある確認証跡）を必須供給として選定・受領する。入力actorと確認医師が別IDの場合も、client doctor_id/booleanだけで代行確認を成立させない。確認医師の認証された行為と現在適格性を検証できなければ保留する。新しい医師資格や代行権限は本案で付与しない。
3. その選定commandに対するexact method/path/body/response、494平坦fence、mandatory vital CAS又はcreate不存在条件、C06 scope/intent/元target lookupを1表にする。現時点は未選定であり『closed API』とは呼ばない。
4. 既存audit_logs再利用のclosed before/after＋metadataと履歴read、通常writer迂回排除一覧を揃える。履歴保証を既存sinkで満たせない具体項目が出た場合だけ、該当列/制約の限定DDL差分へ進む。

これは無期限の『各ownerと合意』ではなく、上記4つの欠落成果を331安全adapter責務で作成・既存供給ownerへ受領する条件。新規業務採否を本人へ聞き直す条件ではない。現時点で進められるのは既存型への適合表、current writer/caller一覧、監査差分、固定fixture設計。未選定の専用write実装や有効化はまだ開始可能と表示しない。

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

[331](https://app.plane.so/baritechllc/browse/EMR-331/)本文/全3コメント、[495](https://app.plane.so/baritechllc/browse/EMR-495/)本文/最新決裁を読取。固定mainの [494契約](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-494-edit-draft-contract-20261020.md) §§3–5、[495契約](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-495-consultation-correction-contract-20261021.md) §§0.2–0.3、[D03正本](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/spec/domains/medical-records/lifecycle.md#medical-record-deceased-retrospective)、[PR845](https://github.com/MinoruSoga/AnimalEkarte/pull/845)の固定head readerを再読。mainのvital_request/repository/audit/service_create/service_update、各route、model/ERDを限定静的照合した。

受入設計はclient doctor_id/booleanだけの確認拒否、別actorの確認証跡未供給/失権/対象版変更時の保留、認証済み確認操作と現在資格/医院権限の照合、死亡先行/child先行、half-dead/unknown、共通他院revision更新、死亡日時訂正、同日時点不明、元fact/child版変更、医師確認対象変更、失権/fence喪失、audit/receipt失敗、同key並行/異内容/応答断、通常writer迂回、担当者≠actor、旧値/履歴不変、副作用0を固定する。今回は全てNOT_RUN。全repo writer閉包、runtime Ready、臨床受入、331/495全体完了は認定しない。
