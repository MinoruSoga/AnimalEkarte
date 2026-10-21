# EMR-510 在庫台帳・実移動・履歴訂正 統合選定技術契約

2026-10-21 統合版 R2（独立レビュー補正反映）。設計担当選定稿、root レビュー用。外部write・runtime・DDL適用・試験なし。NOT_RUN。本文は既存選定/追加選定/対象外/実装依存を区別する。旧2稿に代わる読み切り契約であり、実装済み事実ではない。

## 0. 読取根拠と優先範囲

main を再取得して `0de244ba69f9287a0bacff6abeddabb50a712aac` を確認。Plane510の本文と全3コメント、484本文と公開採択文書、D13業務正本、finance-operations技術正本、001のmovement列/guard、treatment_service_tx.goを照合した。

先行snapshot/baseline/count/ledger/lookupの意味契約は維持。旧2稿の差分を本稿へ統合した。未公開のcount専用correction route/intentを重複追加せず、数量履歴訂正は§4の共通route/intentへ一本化する。最初の実装PRをcount限定にする場合も、同じDTO/Content/anchorを採用し、後続段階でroot型と後続count対応を広げる。旧稿のcount_event_id専用anchor列/専用success refsは今回の統合選定では追加しない。

最新Plane510本文/ACを05:48 UTCに再取得し、全3コメントと業務正本を再照合。要求は既存品目基準、入荷/出庫/廃棄、元出庫付き実返却、履歴訂正。『棚卸前の未登録実移動を棚卸後に初めて遡及登録』を必須とする記載は確認できない。独立した過去入力/訂正から減算・再減算しない記載はある。後日初回登録を新たな全体完了条件にせず、本選定の対象外とする。本人質問は出さない。

### 選定範囲と対象外

選定: 既存品目の一回限りの開始基準、新実査、明示的な実物入荷/出庫/廃棄/返却、既存台帳rootの数量誤記訂正、後続実査による現在庫効果の吸収、元出庫familyの累積返却制約、再送/照会/認可/監査。

明示対象外: 未登録の過去実移動の新規遡及登録、baselineの後日書換、実日時のみの訂正、医院/item/unit/原issue/臨床sourceの付替え、換算/小数丸め、ロット/期限順出庫/棚移動/医院間移管、自動発注、予約数やprepared/ready/プラン予定による減算、自動返金/税/会計/通知。これらを追加承認待ちとして本稿の完了を阻害しない。

数量以外の訂正は今回のclosed DTOに含めないという意味であり、既存業務を永続禁止する決定ではない。既存rootの数量訂正は後続countの有無にかかわらず§1の契約で閉じる。

### 先行基準・実査・readerの契約

GET /inventory/{id}/quantity-snapshot は {item_id,clinic_id,item_version,quantity,management_unit,quantity_base,last_restocked_at}。quantityは未照合ならnull、0に偽装しない。baseはnull又は{id,confirmed_quantity,management_unit,confirmed_at,confirmed_by,legacy_reference}。last_restocked_atは同itemの認可済み台帳から、現在有効数量q(root)>0のrestock rootのoccurred_at最大値をreaderで導出し、該当なしはnullとする。数量0へ訂正されたrootは原履歴に残すが、現在有効な実入荷日の候補から除く。訂正のrecorded_atを実入荷日へ混ぜず、旧未照合値を証拠として復活させない。GETはwriteしない。inventory_items.last_restockedやその代替列をrestock後のservice UPDATEで更新せず、quantity/versionの二重更新・二重版進行を禁止する。別projection triggerも新設しない。

POST /inventory/{id}/quantity-base のclosed DTOは {operation_key,expected_item_version,confirmed_quantity,management_unit,confirmed_at,legacy_reference}、legacy_reference nullable。既存baseなし/既存movementなしを同item lock下で検査、一度だけbase INSERT、既存base triggerがquantity/unit/base/versionを更新。数量確認を架空入荷にしない。新規理由必須を追加しない。別key二回目baseはdependency-changed。

POST /inventory/{id}/counts のclosed DTOは {operation_key,expected_item_version,expected_quantity_base_id,observed_quantity,management_unit,counted_at,reason}。current quantity=S、実査Qならdelta=Q-S、after=Q、version+1。delta0も新しい実査eventを保存する。新実査と、旧実査の誤記訂正no-opを区別する。古いCASを自動更新せず、出庫競合時は実査再確認する。

GET /inventory/{id}/ledger?cursor&limit は {item_id,snapshot_item_version,quantity_base,items,next_cursor}。初回head版を固定しcursorにheadと(resulting_version,id)を保持、後続ページもそのhead以下を読む。別item/不正cursorを拒否、失敗を空配列にしない。itemsは既存列に本稿の訂正意味列を追加した認可済みprojection。内部claim commandは公開しない。

GET /inventory/{id}/operations/{operation_key}?intent は本稿のclosed intent allowlistだけを許可する。元item現在認可後に結果を返す。baseline/countのintentは inventory.quantity_base.confirm / inventory.count.record。ContentはURL itemとDTOのoperation_key/expected_item_version以外の意味値をJCS hashへ固定。base IDは意味依存。成功refsはinventory_item_id/quantity_base_id、countのみmovement_event_idを追加、versionsはitem_version。base/event/projection/必須audit/receiptは同TX、serviceがtriggerのquantity更新を重ねない。

baseline audit action=inventory.quantity_base.confirm、count audit action=inventory.count.record、resource=inventory、resource_id=item ID、actor_type=staff、actor_id=trusted staff。old/new={quantity,item_version,quantity_base_id}、metadata={operation_id,quantity_base_id,movement_event_id:null又はID,management_unit,confirmed_at:null又は時刻,counted_at:null又は時刻,legacy_reference:null又は値,reason:null又は理由}。未設定/ゼロを区別し、同key replayはauditを重複させない。

既決: 実施/実引渡し時の職員明示出庫、準備/保存/請求だけでは減算しない、整数管理単位、負在庫拒否、原実績保持、理由付き訂正、過返却拒否、取消/返金だけの自動巻戻し禁止。

既存実装: inventory_movement_events.delta は現在庫の増減で after=before+delta。correctionのゼロdeltaは既に許容。return guardは固定の原issue.deltaと未補正return.delta合計のみで、訂正後の有効数量には未対応。treatment_service_tx.goのcreateはまだInventoryID指定時DecreaseStockを呼ぶ。これを新契約適合としない。

484は performed/prepared/revoked の請求適格性。preparedは引渡しではなく、performedのeventは全ての分割実移動を識別する保証がない。履歴generation変更/取消を在庫の巻戻しにしない。

## 1. 二つの数量と棚卸境界

数量訂正は記録の訂正であり、その時点の新たな実物移動ではない。

root実績 R の採用数量 q(R) は、未訂正なら count.after_quantity、restock/returnならdelta、issue/wasteなら-delta。訂正済みなら唯一の最新correction headに保存された corrected_quantity。すべて非負整数。0への訂正は記録誤りを0として保持するもので、root削除・返品・新入荷ではない。

実績rootの符号は restock/return=+1、issue/waste=-1、count=+1（当時実査数の差）。訂正の履歴数量差 dQ=q_new-q_old と、現在庫への効果 dS を分離する。

- R より後の登録版に count root がなければ dS=sign(R)×dQ。
- R より後の登録版に count root があるなら dS=0。最新の後続count IDを absorbed_by_count_event_id として保持する。後続実査で置き換え済みの過去差分を現在庫へもう一度足さない。
- 比較基準は同itemの immutable resulting_version。occurred_atの日時順で登録列を並べ替えない。correction自体を新countと扱わない。最新countの実査値訂正は、その後の通常入出庫を保持した差分反映となる。
- 棚卸をまたいでも issue/returnの有効数量と残未返却量は履歴から継続する。棚卸が返却可能残をリセットするわけではない。

例: count A=5、出庫2、count B=10、出庫1で現在9。Aを4へ訂正すると履歴差-1、現在庫効果0、現在9。Bを8へ訂正すると履歴差-2、現在庫7。A/B原本は不変。

既存台帳に既に登録されたrootの訂正を扱う。未登録の過去実移動の遡及新規登録は対象外。実日時だけで吸収済みと自動推定しない。

## 2. 出庫familyと累積訂正

familyは原 issue I と original_issue_id=I.id の全 return root。各rootの訂正chainは一直線で、過去の訂正は消さない。最新headをたどり q(I), q(return_j) を求める。

有効返却累計 R=sum(q(return_j))、残未返却量 U=q(I)-R。常に 0<=R<=q(I)。新しい実返却 quantity は正整数かつ quantity<=U。訂正後も全family最終値で同じ制約を満たす。原issueを2から1へ訂正して既返却2を放置することは拒否する。count吸収済みで現在庫効果0でもこの検査を省略しない。

訂正対象が複数なら同familyを一つのcommandで修正可能とする。例えば issue2/return2 を issue1/return1へ訂正する場合、別TXでreturnを先に減らすと現在庫0から一時負数になり得る。一つのTXで最終有効量と現在庫合計効果を検査する。同じheadを対象にした別keyの再訂正はhead競合で拒否し、最新headで同じ目標量を再指定した場合はno-op。0へ直した記録を後に新しい根拠で別量へ訂正する場合は、最新headと理由を明示して、その時の採用数量との差だけを一度反映する。

別family・別itemを一括訂正するAPIはこの最小契約へ追加しない。restock/waste/countは単一root訂正。源item/unit/元issue/臨床source/実日時の付替えは数量訂正APIでは不可。元targetを変えることで補正制約を迂回しない。

## 3. 最小DB差分とguardの意味

movement_eventsへ追加する列:
- corrected_quantity bigint NULL CHECK >=0。新契約correctionでは必須、他kindではNULL。
- absorbed_by_count_event_id bigint NULL。同院同itemのmovement scoped FK、ON DELETE RESTRICT。非NULLなら対象はcount rootかつ訂正rootより後の登録版。
- physical_movement_key uuid NULL。新しいissue/return rootで必須、correction/countではNULL。§6の実移動identity。旧行はNULLのまま不明を表す。

correction_head_event_idとexpected_correction_head_event_idは未訂正時には原root ID、訂正済み時には唯一の最新correction IDを使用し、初回もNULLにしない。return DTOのexpected_issue_head_event_idも同じ規則。新correction.corrected_event_idはその直前headを参照する。UNIQUE(clinic_id,inventory_item_id,corrected_event_id) WHERE event_kind='correction' を追加。root型/chain所属/循環・分岐・欠落をguardで検証。absorbed_by_count_event_idは後続countの存在時にその最新IDと一致、存在しない時はNULLを要求し、clientが自由に省略/指定できない。新correctionはheadの採用数量とcorrected_quantityから履歴差を再算出し、吸収countの有無からdeltaを検査する。clientのdeltaを信用しない。

UNIQUE(clinic_id,physical_movement_key) WHERE physical_movement_key IS NOT NULL を追加。一つのphysical keyは一つのitemの一つのissue又はreturn root。複数品目を同時に扱う場合は同一出庫への無断拡大をせず、本段階では品目別確認を行う。

旧correctionのcorrected_quantity=NULLを一律backfillしない。既存chainのroot/順序/数量/吸収関係を証明できる限定adapterだけ旧deltaから有効量を復元する。証明できない行はreconciliation-requiredとして新訂正/返却を止める。旧データの書換・削除による解消は本稿の許可外。

新しい可変effective-quantity表や二重registryは作らず、同item lock下でimmutable root/chainから算出する。guardのreturn累計をこの同じ有効量関数に置き換える。correction batchの途中値だけで過返却を判定せず、影響したissue familyの最終量をDEFERRABLE INITIALLY DEFERRED constraint triggerで再検査する。単一returnも同最終制約を必ず通す。serviceだけの検査にしない。

既存item projection triggerを唯一のquantity/version writerとして保持する。correction挿入順は「delta>=0を先、delta<0を後、同群はroot ID昇順」。合計最終量が非負なら途中負数を作らない。各INSERTのbefore/after/previous/resultingはこの順で連続する。全予定行の中間quantityとv+変更root数も事前にint64範囲検査し、表現不能ならvalidationで全件ゼロwrite。overflowを丸めたり一部だけcommitしない。各変更rootに1event、ordinalもこの順。履歴だけ変更してdelta0でもeventを保存しitem versionを1進める。本当にqが同じならeventなし・version不変。

immutable/RLS/同院FK/operation成功receipt必須の既存保護を維持し、新列/新constraintを既存表へ統合する。新別triggerからprojectionを書いたり、一時的に制約を無効化したりしない。

## 4. 後段の公開DTOとC06

新共通入口（/api/v1除外）:
- GET /inventory/{id}/movements/{root_event_id}/correction-context
- POST /inventory/{id}/movements/{root_event_id}/corrections
- 既存提案operations lookupのintent allowlistへ inventory.movement.correct を追加。

root_event_idは原実績。correction IDならrootを勝手に付替えず拒否。issueを指定した場合はfamily全体を返す。returnを単独で訂正する場合もfamily情報を返し、commandはそのrootだけを変更する。複数root訂正はissue URLを使用する。

GET closed: {item_id,item_version,quantity_base_id,management_unit,current_quantity,root_event_id,latest_count_event_id,roots,family}。
roots[] closed: {root_event_id,event_kind,original_quantity,effective_quantity,correction_head_event_id,absorbed_by_count_event_id,occurred_at}。
family=null 又は {issue_event_id,effective_issued_quantity,effective_returned_quantity,remaining_returnable_quantity,return_roots}。return_rootsは同じroots形式で完全取得。ページ途中のfamilyを完全値としない。

POST closed: {operation_key,expected_item_version,expected_quantity_base_id,expected_latest_count_event_id,management_unit,reason,changes}。
changesは非空配列、各要素 {root_event_id,expected_correction_head_event_id,corrected_quantity}。重複root拒否、正規化時root ID昇順。単一count/restock/waste又は同issue familyだけ。expected_latest_count_event_idはnullable。比較は登録版の最新count。ID/版はcanonical positive decimal string、数量はcanonical nonnegative decimal string、signed bigint上限以下。wire number、小数、指数、+、先頭ゼロ、未知field、重複JSONkey、不正Unicodeを拒否。reason/unitはtrimして非空、それ以外のUnicode正規化なし。operation/physical keyはcanonical UUID。offset付きRFC3339を検証しUTC instantへ正規化。hashはRFC8785 JCS UTF-8のSHA256 lower hex64。新専用DTOの規則であり、既存numeric API全体をstringへ移行しない。算術は広い整数で中間値を計算し、保存値/版はint64範囲を検査する。

owner_kind=inventory、intent=inventory.movement.correct、schema_version:"1"。Content={schema_version,item_ref,root_event_ref,quantity_base_ref,command_sha256}。hash対象はURL item/rootとPOSTのoperation_key/expected_item_versionを除いた全command。changesのhead、latestcount、baseを意味的依存として固定。item CASだけlive precondition。immutable typed commandをclaim metadataとApplyで同じcanonical bytesから再構成する。

本節のimmutable typed commandは[採択済みQの共通契約](../../design/shared/operation-participation.md#verified-command-selected-20261021)に従い、Applyがreq.VerifiedCommandだけから再構成する。Qはrequest-memoryのみでowner-private耐久保管とは別、CAS/fenceはPreconditions、Contentは既定の短いstring mapを維持する。489のQ必要subsetは未実装依存であり、別command/DTO/pointerのcaptureで迂回せず、既存Content-only consumerは変更しない。

owner-private canonical_command_bytesは既定の業務意味値だけを保持する。既存keyでは現在元target認可後に保存bytesのdecode・再canonical化・hashを検証して同じQへ載せ、現在item CASを保存業務commandへ混入しない。Qを第二の耐久保管先にしない。


新keyは現在inventory:editでpreflightしてからclaim。binding lookup errorを不在と扱わず、view-only新keyはbinding/anchor/attemptの永続化ゼロ。claim race loserは勝者bindingの元anchorを再認可しincoming URLで補完しない。既存bindingは元anchorの現在inventory:viewでlookup/replay。Applyはeditを再検査。元target現在認可後URL不一致は既存404。既存mutation routeのedit middlewareを勝手にviewへ緩めない。POST入口がedit必須の配置ではview-onlyの保存済結果回復はGET lookupを使用し、内部replayのview認可規則だけでPOSTの入口grantを迂回しない。binding初回callbackは選定中のinventory_operation_targetsへ保存する。この表は既実装ではない。列はclinic_id bigint NOT NULL、binding_id bigint NOT NULL、inventory_item_id bigint NOT NULL CHECK >0、root_event_id bigint NULL CHECK >0、canonical_command_bytes bytea NOT NULL。PK(clinic_id,binding_id)、唯一のFKは(binding_id,clinic_id)→operation_bindings(id,clinic_id) ON DELETE RESTRICT。base/count新規記録はroot=NULL、訂正はURL root、returnは元issueをrootとする。UUIDのphysical keyと検証済みimmutable source参照/期待stampはcanonical commandに保持し、重複claim registryを追加しない。claim callback内でsource proofの読取・lockを行わず、Applyで保存参照/期待stampからsource proofを再構成して比較する。

489採択ClaimMetadataWriter.CreateClaimMetadata(ctx,binding,request)が新bindingと同じ短いclaim TXでINSERT。INSERT … DO NOTHING RETURNINGで新bindingを得た作成者だけcallbackを呼ぶ。業務FK/row read/lock/nested TX/receipt/外部作用はclaimへ持ち込まない。clinic RLS、INSERT-only immutable guardを持ち、既存bindingのanchorは付替えない。成功なしもlookup可能。

Content不一致=operation-key-conflict。Content一致後item CAS不一致=version-conflict。base/head/latestcount/単位/依存の変化=dependency-changed。無効数量/過返却/不足=validation。保存済成功は現在head/CASで再実行せず元receipt。unknown/busy中は同key照会のみ、別key開始禁止。

公開HTTP/statusは succeeded/replayed/not-succeeded=200、operation-in-progress=202、result-unavailable=503、operation-key-conflict/version-conflict/dependency-changed/edit-right-lost=409、validation=400、scope-not-found=404、not-authorized=403。pre-C06入口だけ既存{error:string}。既存finance他routeは変更しない。本consumerで通常発生しない状態も共有12値のenum/HTTP対応を維持し、新しいedit fenceを導入する意味ではない。成功refs={inventory_item_id,quantity_base_id,root_event_id,correction_operation_id}、versions={item_version}。correction_operation_idはbinding IDであり、同operation_idのevent列をledgerから読める。配列やEffectSummaryをResultResponseへ追加しない。非成功refsなし。no-opも同形式、event0件でaudit/receiptだけを保持する。

## 5. 原子性・監査・旧writer閉包

業務TXは既存owner Gate（不要ならno-op）→C06 binding lock（NOWAIT）→必要なsource owner下位locks→inventory item FOR UPDATE→root/chain/count/family読取→最終command再検査→全訂正event→必須audit→receiptの順。認可の非locking read preflightと排他Gateを区別し、Applyで現在grantを再検査する。sourceを変更しない訂正ではsource下位locksは不要で、inventory lock後に臨床sourceを後取りしない。全stock writerは同item lockを共有し、直接quantity更新を拒否する。familyの新returnとissue量訂正もこの境界で直列化する。

audit_logs: action=inventory.movement.correct、resource=inventory、resource_id=item ID、actor_type=staff、actor_id=trusted staff。old_value/new_value={quantity,item_version,roots:[{root_event_id,effective_quantity,head_event_id}]}。metadata={operation_id,quantity_base_id,latest_count_event_id,reason,no_op,event_ids}。保存内部summaryと公開receiptは分ける。actor/recorded_atはserver、訂正event.occurred_atは原rootを保持。audit失敗/最終family制約/receipt失敗は全event/projectionをrollback。

UIは『当時の記録数量の変更』『今回の現在庫への効果』『後続実査で吸収済み』を区別。ledgerは変更前後・原root・head・吸収countを表示。履歴補正を実返品や入荷へ分類せず、last_restockedをcorrection日時から作らない。

legacy PATCH.quantity/last_restocked、treatment createのDecreaseStock、薬剤master同期、ready/prepared、診療プラン予定、会計確定/取消/返金から本intentを発火しない。旧treatment減算を撤去/明示intentへ切り分ける最小subsetを510 activation条件とする。税/請求/売上/返金/臨床state/通知へのwriteはゼロ。

## 6. 実移動identityとS3の具体的依存

physical_movement_keyは『同じ実物の一回の動き』を指すUUID。HTTPの再試行/CAS/訂正操作を指すoperation_keyとは意味が異なる。単にすべてのcommandへ二つ目のrequest IDを足すのではない。issue/returnだけが必要で、count/数量訂正には付けない。

最初の明示実移動確認時に一度生成し、当該確認操作の間は同じphysical keyを保持する。元keyを保持できている通信再送/CAS再確認では新keyを生成しない。送信前の耐久draft保存先・未確定claimのdiscovery・reload復元は本稿では提供/選定しておらず、claim metadataがあるだけで自動回復できるとは保証しない。応答不明なら新physical keyも新operation keyも生成しない。reload等で元物理移動・元operation key・結果を特定できず、未送信か送信済みかも判別できない場合は操作を停止する。新key発行や数量/名称一致の推測で救済しない。原移動成功後の数量修正はcorrectionsであり、新physical keyによるissue/returnではない。意図的な次の部分実施/次の実引渡し/次の実返品だけ、新しい明示確認とphysical keyを持つ。同じtreatment IDを一意にして部分実施を潰さない。

別operation keyで同じphysical keyが既に成功rootに使われていた場合、現在元itemの認可を確認し、新規Applyはdependency-changedで副作用ゼロ。別keyを他の成功receiptへ偽装replayしない。inventory:viewの専用 GET /inventory/{id}/physical-movements/{physical_movement_key} は認可済みroot IDと元operation keyを返して既存lookupへ案内する。このGETは成功rootだけを対象とし、未成功claimの発見APIではない。失敗してrootがない場合、元keyを保持しているときに限り元key lookupを使う。元key不明なら上記停止条件へ戻す。成功なしから再claimする場合もphysical keyは維持する。

このkeyは現実の物を自動認識する仕組みではない。利用者が同じ現実の動きに新しいidentityを意図的に作った場合までDBが検出できるとは主張しない。画面は元sourceの登録済移動/実返却と未解決操作を提示し、追加実移動の確認と記録訂正を分ける。別端末での無関係なkey発行を名前/数量一致だけで勝手に同一視しない。

S3 request最小差分: issue/returnの既存提案DTOへphysical_movement_keyを追加しhash/claim metadata/成功rootへ固定。issueのactual_kindは in_clinic_administration / take_home_handover / merchandise_handover のclosed enum、actual_confirmed=trueを必須とする。prepared/performedをactual_confirmedへ自動変換しない。数量は職員が管理単位で明示し、臨床小数から換算しない。

source adapterの引渡し契約は {clinic_id,source_kind,source_id,source_stamp,record_id_nullable,pet_id_nullable,inventory_item_id,management_unit,actual_kind} のimmutable proof。source_stampは {kind:"version",value:<canonical decimal string>} 又は {kind:"sha256",value:<hex64>}。source_kindは treatment / billing_item / merchandise_item のclosed unionとし、前二者は現行source_treatment_id/source_billing_item_id、後者は下記追加列を使う。sourceの所属/現在view/typed itemと在庫ID対応を既存ownerが検証し、無対応・曖昧ならdependency-changedで停止。名称一致、MedicineID=InventoryID、preparedからの推定は禁止。billing_item未作成の物販はmerchandise_item proofを用い、架空の請求行を作成しない。

484の全部完成待ちにはしない。510が必要なのは source read adapterと明示出庫導線、安定identity、元sourceの現在認可/医院照合、管理単位明示、旧create減算撤去。484 ready保存と在庫commitは別結果で、readyが成功しても在庫未確定を表示する。実施eventを生成する必要がある複合UXでも、preparedのみからstock呼出ししない。source付き新実移動は共有owner Gateが返すcontextをC06 binding以下へ引き継ぎ、Gate→binding NOWAIT→plan→dependency→typed source→inventory itemの順とする。Gateと下位locksを混同せず、inventoryを先にlockする逆順writerを残さない。C06は最外TX、nested transactionは作らない。


### S3追加wireの固定

GET /inventory/{id}/issue-context?source_kind=...&source_id=... はinventory:viewと当該sourceの既存viewで純読取し、{item_id,item_version,quantity_base_id,current_quantity,management_unit,source:{kind,id,stamp},supported_actual_kinds,existing_movement_refs}を返す。supported_actual_kindsはtyped sourceの明示分類から返し、不明を名称から補完しない。existing_movement_refsは同item/sourceのroot一覧を完全取得又は明示ページ情報付きで返し、部分一覧を全件としない。GETでphysical key/claimを永続化しない。treatment/billing_itemはsource ownerの既存versionをstampへ使用し、versionが採択済み実装へ接続していなければ偽の値を返さず停止する。merchandiseは§7の閉じたfingerprintを使う。


POST /inventory/{id}/issues のclosed DTOは {operation_key,physical_movement_key,expected_item_version,expected_quantity_base_id,quantity,management_unit,occurred_at,reason,actual_kind,actual_confirmed,source}。source={kind,id,expected_stamp}、上記3kindのunion。item/clinic/actorはURL又はtrusted context。source proofはserverで再構成し、clientのrecord/pet/clinicを信頼しない。performed/preparedのevent IDを必須の実物identityにしない。既存source viewを持たない場合はその情報を漏らさず拒否、新しいgrantは作らない。

POST /inventory/{id}/movements/{issue_event_id}/returns のclosed DTOは {operation_key,physical_movement_key,expected_item_version,expected_quantity_base_id,expected_issue_head_event_id,quantity,management_unit,occurred_at,reason,actual_confirmed}。元issue IDはURL、returnには別sourceを受けず原issueを保持。expected issue headは意味依存としてhashへ含める。最新return累計はitem lock下で算出。別returnが増えた場合item CASで再確認し、旧量を黙って減らして通さない。

両route quantityは正整数string、actual_confirmedはliteral true。operation_key/expected_item_version以外の全DTO＋URL参照をJCS hashに固定。source_stampもContentへcommitする。intents=inventory.issue.record / inventory.return.record。claim anchorは§4の選定inventory_operation_targetsを使用する。itemをinventory_item_id、元issueをroot_event_id(nullable)、physical keyとsource参照/期待stampをcanonical_command_bytes内へ保存する。source proofはApplyで再構成・照合する。physical keyを新しいclaim tableへ二重保存しない。hash不一致/key conflict、CAS/version conflict、source/head/base不一致/dependency changedを分ける。

成功refs={inventory_item_id,quantity_base_id,movement_event_id}、returnではoriginal_issue_idを追加。versions={item_version}。audit action=inventory.issue.record又はinventory.return.record、resource=inventory、old/new={quantity,item_version}、metadata={operation_id,movement_event_id,physical_movement_key,original_issue_id:null又はID,source_kind:null又はkind,source_id:null又はID,quantity,management_unit,occurred_at,reason}。actor_type=staff、actor_id=認証済trusted staff IDを記録する。serverはactorを認証contextから決定する主体であり、system actorへ置換する意味ではない。body actorを受けず、必須staff欠落時は既存認証拒否でwriteゼロ。created_at等の記録時刻はserver、audit/event/receipt同TX。claimに残るimmutable source参照/期待stampと、Applyで再構成・照合した出所の監査が対応を保持し、公開ResultResponseには載せない。

POST /inventory/{id}/restocks と /wastes はclosed DTO {operation_key,expected_item_version,expected_quantity_base_id,quantity,management_unit,occurred_at,reason} とし、intent=inventory.restock.record / inventory.waste.record、Contentはoperation_key/expected_item_version以外とURL item、同じ監査形でsource/physical keyはNULL。数量の符号はrouteが固定。restock/wasteも未知結果中は新key禁止。本稿は当該実物移動の明示記録を扱い、過去未登録移動の遡及登録導線は追加しない。実日時をserverの記録時刻と混ぜない。

## 7. 実装セッションが完結させる順序と境界

A. 510第一PR: §0のsnapshot/baseline/count/ledger/lookup。独立して実装可能。
B. 510の数量訂正最小PR: §4の共通APIをcount root/後続countなしに限って接続可能。独自count専用route/intent/anchorを追加しない。Aと489claim participantの必要subset。全510完了ではない。
C. 510後段履歴PR: 本稿§1–5の列/有効量/absorbed-by-count/同family batch/guard/DTO/ledger。既存issue/return rootの訂正はここで閉じる。sourceが過去に消えていても原rootを付替えずinventoryの現在認可と保存参照を使う。
D. 510実移動PR: restock/waste/issue/return、physical key、実移動確認、effective return cap、source adapter、旧treatment直接減算撤去。DはCの有効量関数を使用し、旧sum(delta)へ戻さない。元keyを保持する確認操作と同key照会までが選定範囲で、耐久draft/reload自動回復は未提供。特定不能時のfail-closed停止を実装し、回復済みと表示しない。
E. 全入口activation: 446/466が001/model/ERD・RLS/immutable/constraints、489がclaim metadataとreceipt参加、510がwriter/reader/画面、484/481が必要source読取/明示導線を接続。共有router/OpenAPI/codegenは合意ownerへ渡し、別系統を作らない。

billing_item未作成の物販sourceを閉じる選定subset: 510と物販ownerが、価格名ではなく現存merchandise item ID＋明示inventory item ID＋現在版によるsource proofを追加する。movementへsource_merchandise_item_id bigint NULL、FK(source_merchandise_item_id,clinic_id) REFERENCES merchandise_items(id,clinic_id) ON DELETE RESTRICT、参照先のUNIQUE(id,clinic_id)を必要なら追加、既存source列とのnum_nonnulls(source_treatment_id,source_billing_item_id,source_merchandise_item_id)<=1、immutable保存、現在ResourceMasterMerchandise:viewの検査を同一PRに含める。自動マッピング表や価格マスタ変更の権限は新設しない。紐付け未確認時は停止。これは新ビジネスルールでなく、既決の物販実引渡しを名称推測なしで表現する追加technical subsetである。現物model MerchandiseItemはID/ClinicID uint64(bigint)、version列なし、inventory参照なしを確認済み。架空のsource_versionを要求せずsource_stamp=sha256(JCS({clinic_id,id,category,is_active,deleted:<bool>}))を使用する。名前/価格/税変更だけでは実移動identityを作り直さず、医院/対象/有効性をApplyで再検査。masterはread-onlyで、当該移動に明示選択したinventory IDとunitをcommandへ固定する。物販全体への永続マッピングを勝手に作らない。

将来Dのreload回復まで提供する場合の限定技術依存は、元item現在認可付き未確定操作discovery、送信前後の保存先/寿命、遅延旧送信の閉鎖、新操作開始の明示境界。現回復契約（LIFF/証明書）は510を含まないため流用完了扱いにしない。localStorage/URL/未送信本文永続化を勝手に選定しない。この未提供機能はA〜CやDの台帳guard/有効返却制約を止める条件ではなく、安全reloadを完成と主張する場合に別途閉じる依存である。

各subsetはコード/生成物/fixtureを作る実装責任を明記し、文書公開をruntime有効化としない。実行はNOT_RUN。全DBや全484を一括ブロッカーにしない一方、source閉包未接続の入口を適合済みとしてonにしない。

## 8. 必須fixture（作成計画、NOT_RUN）

1. 古いcount訂正は履歴差のみ、最新count訂正は現在庫差分。複数countを越えても最後のcountを変えない。
2. issue5→return2→issue量4へ訂正:有効残2。returnを1へ訂正:有効残3。さらにissue量3へ訂正:有効残2。旧値へ毎回差を再適用しない。
3. issue2/return2、現在庫0を同familyで1/1へ訂正:最終量0、event2、receipt1。単独issue1は過返却拒否。単独return1は負在庫なら拒否。
4. count吸収済みissueを縮小して過返却になる訂正も拒否。返却済みrootを0へ訂正後、新実返却は最新有効残の範囲だけ。
5. 同head別key競合、同key内容相違、item CAS、latestcount競合、history-only版増分、本物no-op版不変、audit/receipt/deferred検査失敗全rollback。
6. 同physical key別operation key再出庫ゼロ、部分実施は別明示key、元keyを保持するunknownは旧key照会、reloadで元key/送信有無不明なら停止し新key救済ゼロ、ready/保存/請求/予定のみ在庫不変。
7. 単位不明/小数/overflow/欠落chain/旧補正不明は停止。別院・失権・元target付替え・原issue付替え拒否。
8. 原sourceのgeneration変更/ready取消/請求取消でも過去stock eventを消さず自動返却しない。

## 9. 根拠URL

- https://app.plane.so/baritechllc/browse/EMR-510/
- https://app.plane.so/baritechllc/browse/EMR-484/
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/docs/spec/domains/inventory/business-rules.md
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/docs/design/domains/finance-operations/technical-spec.md#inventory-selected-contract
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/migrations/001_init.sql#L9752-L9789
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/migrations/001_init.sql#L10300-L10313
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/model/merchandise_item.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/inventory/routes.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/backend/internal/medicalrecord/treatment_service_tx.go
- https://github.com/MinoruSoga/AnimalEkarte/blob/0de244ba69f9287a0bacff6abeddabb50a712aac/docs/work/emr/emr-484-fulfillment-writer-contract-20261021.md

## 10. 整備完了と実装完了の境界

今回の既存履歴訂正・台帳実移動の選定範囲について、追加の本人業務質問はなし。未登録実移動の遡及新規登録を今回必須にする根拠は確認できず、対象外として保持する。これを勝手な要件追加や新しい承認待ちにしない。

root採択後に設計整備の完了を判断できるが、実装/API登録/DDL適用/activation/実試験の完了とは別。A〜Eの必要subsetと全writerの閉包、残る旧データ不明の停止表示を実装セッションが引き受ける。試験コード/fixtureの作成と実行NOT_RUNを区別し、試験成功を捏造しない。実装/merge所有は実装sessionのまま。

R2レビュー補正: Gate→binding順へ全文統一、未提供の耐久draft/reload保証撤回と特定不能停止、last_restocked_atの有効restock reader導出・二重UPDATE禁止、初回head=root IDを反映。
