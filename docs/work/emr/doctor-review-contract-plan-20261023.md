# 医師確認集合：内部reader・writer参加の設計作業案（2026-10-23）

**内部契約と既存billing Confirm/Returnの限定技術選定。製品runtime Ready・実装・公開API供給・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.doctor-review-contract-plan-20261023 |
| 種別 | 選定済みINTERNAL契約への接続と後続設計作業案 |
| 決定状態 | T06は承認済みidentity/数量/record集合の限定signature/digest/世代参加へ修正。確認時signature保存・共用医師reader方向を維持。既存billingのwire/保存差分/専用lock順を追加選定。実provider・DDL/API・共通gate・writer閉包の供給は未完、準備billing/Complete等は別packet |
| 実装状態 | 選定契約の供給・接続・全writer閉包は未完 |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-24 JST（限定契約の選定反映、実行NOT_RUN） |

[今回の索引・判定軸](emr-limited-audit-plans-index-20261023.md#limited-audit-status)を参照。以下の監査範囲・件数は受領した固定基準の報告であり、今回再度の全件照会を行ったという意味ではない。

2026-10-23のroot修正選定は既存technical-spec:450–456を優先し、広いclinical_content差分による失効と新current列の先行採用を撤回した。§3の非活性内部境界は[技術契約](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-scope)へ集約した。既存billingの公開wire・anchor・保存差分・専用取得順は[Confirm/Return契約](../../design/domains/accounting/doctor-review-confirm-return-contract.md#review-actions-scope)へ2026-10-24に限定選定した。以下のComplete・生前分C・退院D等は別packetに保持する。内容レビュー、文書公開、必要供給、consumer接続、全writer閉包、実行受入、activationを区別し、担当/実行ID/供給SHA受領済みとはしない。

<a id="doctor-review-plan-scope"></a>
## 1. 結論と確認済み根拠

先行の非活性内部typed reader/projectionと純粋世代参加契約を維持し、今回[既存billingの現在集合read・明示Confirm・理由付きReturn・不変履歴/C06 read・旧writer迂回排除](../../design/domains/accounting/doctor-review-confirm-return-contract.md#review-actions-scope)を技術選定した。準備billing作成/Complete gate有効化/既存Complete動作変更は含めず、全体機能sliceをReady/実装済みとはしない。

readerは一つの既存billingの完全な現在record集合（複数record可）とplan-backed treatmentを扱う。各linkを対応する同recordへ照合する内部範囲であって、製品の通常Completeをこの範囲に制限する業務判断ではない。billingなしComplete、新規作成、他sourceは今後の接続残ACとして保持する。生前分Cと退院takeover Dも別slice。選んだcohort内のwriter閉包を点検できる成果を先に作り、全source/全495/全489のDone待ちにしない。

当初の調査基準は `171b56575e7ee028e54358dad22ae28574853125`（PR843 merge）。先行内部選定は `b81abc26b3657bd662475a1f55b912672a24a265`、今回のConfirm/Return追加選定は `7dad9adc7f066fab5cc734dc8ff7e125a80d67d4` を基準とした入力を受領し、固定根拠を[内部契約](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-evidence)へ置いた。以下の既存171b565リンクは当初観測の固定参照として保持する。Planeは462本文と8コメント、481本文と3コメント、484本文と4コメント、495本文と15コメントを取得し、各responseの next_page_results=false を確認。過去の未merge/未選定記述を現在条件にしない。

- [462](https://app.plane.so/baritechllc/browse/EMR-462/): 06c5f90c-503d-44f0-9f91-977ff8b746a9（通常一source先行、C/D分離）、0c13fe63-b33b-42e1-a8f2-2da811b85ced（既存Complete 201/200・Billing wire、元target現在view）、6a7c73d4-f44d-4391-befe-85bb1f590c4b（nil死亡readerの残差）。
- [481](https://app.plane.so/baritechllc/browse/EMR-481/): f45a38ae-83c0-47ac-90ae-5ce96808b348（A/B/C/D維持、対象writer閉包）。
- [484](https://app.plane.so/baritechllc/browse/EMR-484/): 現世代prepared/performedと明示purpose。prepareと実引渡し/出庫は別。
- [495](https://app.plane.so/baritechllc/browse/EMR-495/): **最新本人決裁7b95529a-f045-49da-b5ae-aab5a315e2bbを優先**。P1共通helperのfinalized拒否のみ除去、医院/存在/有効親/ambient TX/正規lock順・各caller CAS/監査/目的別保護を保持。494未参加はcaller残差。AのBE+api.yaml同PR、codegen本人、D2は既存495契約の最初のestimate_service変更単位へ同梱する境界を参照。これを494全完了まで停止へ戻さない。

保存先・三軸分離・price-only再確認不要・臨床数量変更再確認は既決。新しい医師資格、終了条件、財務権限、棚卸/出庫副作用は追加しない。

<a id="doctor-review-plan-storage-evidence"></a>
## 2. mainで実在する保存形と実装残差

| 対象 | 確認した契約/コード | 今回の扱い |
|---|---|---|
| medical_records | clinical_set_generation bigint NULL、正値CHECK、clinical_set_digest bytea NULL/32byte。Goは *int64 / *[]byte | 現在record集合の状態。NULLを1/空集合/確認済みに補完しない |
| billings | 同じnullable世代/digest、doctor_review_set_id nullable、clinic/id込み複合FK | 会計側集合状態と確認参照。存在だけで有効としない |
| doctor_review_sets | billing_id必須、billing_clinical_generation正値、32byte digest、record_count、reviewed_by/at、nullable confirmation_note（後続ALTERで追加） | 既存の確認世代/digest等の保存形。会計側clinical_itemsの不変保存まで供給済みとは扱わない。第二確認表を今回追加しない |
| doctor_review_records | (clinic,review_set,record) PK、billing込みFK、正値世代/32byte digest、clinical_items JSON配列 | record集合は完全一致で照合。部分一致不可 |
| 001 L8160–8192 | opt-in metadata。parent-updating treatment triggerを作らず全writerをconsumer TXへ参加、と明記 | 共通DB triggerへの丸投げ/第二counterは禁止 |
| accounting_complete.go/tx.go | 旧completion_request_id/hash、旧status/totals caller、既存最外TX・unbilled版・takeover保護 | 削除列の復活ではなく462統合adapter。completeInTxにdoctor_review_sets照合がない |
| billing_confirmation_service.go | GetOrCreateが旧表をINSERT、Confirm/Returnが旧statusとLockDraftMedicalRecordを使用 | GETをwriterとして温存しない。互換reader/新確認adapterへ移行 |
| operation/contract.go, runner.go | VerifiedCommand Q、ClaimMetadataWriter、RunnerのGate(ctx) errorは実在 | 派生contextを引き継ぐContextGate拡張自体のruntime interface/hookは未実装。consumer接続だけの不足とはしない。Q等を未実装へ戻さず、consumer anchor/統合/実DB受入と分ける |

一次根拠:
- [confirmation_note追加ALTER](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/migrations/001_init.sql#L10964)と[保存model](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/model/storage_finance.go#L293)。CREATE TABLE直後だけで保存形を判断せず、物理名confirmation_noteをnoteという別列にしない。
- [最終DDL A1](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/migrations/001_init.sql#L8160-L8192)
- [doctor_review_sets ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/architecture/erd-detail/doctor_review_sets.md)、[doctor_review_records ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/architecture/erd-detail/doctor_review_records.md)
- [medical_records ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/architecture/erd-detail/medical_records.md)、[billings ERD](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/architecture/erd-detail/billings.md)
- [model/storage_finance.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/model/storage_finance.go#L270-L296)、[accounting.go](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/model/accounting.go#L92-L97)
- [Complete](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/accounting_complete.go)、[completeInTx](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/accounting_complete_tx.go)

<a id="doctor-review-plan-reader-candidate"></a>
## 3. 選定済みの非活性INTERNAL契約と未供給境界

[医師確認集合の内部技術契約](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-scope)を型・canonical projection/digest・純粋世代参加の正本とする。以下の旧anchorは参照互換として保持し、候補本文の二重管理はしない。

<a id="doctor-review-plan-completeness"></a>
### 3.1 対象と完全性

[対象集合](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-population)と[port/closed結果](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-ports)へ。複数record可、各linkの対応record一致、record全薬剤/処置とbilling全臨床行の独立完全性、混在unsupportedで部分knownを返さない条件を選定した。実DB current providerとplan/link供給は未完であり、合成typed値のknownをruntime供給認定にしない。

<a id="doctor-review-plan-projection-candidate"></a>
### 3.2 projection・digestと会計側履歴

[exact shape](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-shapes)、[canonical digest](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-canonical-digest)、[論理保存方向](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-storage-direction)へ。失効signatureは安定行/source・薬剤/処置identity・数量とrecord集合に限定する。current billing identityは481のactive binding＋保存済みtreatment_originalを、現在採用値との一致をadapterが証明できる行で再利用し、quantityはbilling_itemsから読む。treatment_originalは最後の取込snapshotで、manual override/keep_billing後の現在値を当然には証明しない。不明ならunresolvedを保持し現在treatmentで補修しない。新current projection列の自動採用は撤回し、既存doctor_review_setsへの確認時の限定signature全量不変保存の論理方向だけを維持する。物理保存の選定は[保存差分](../../design/domains/accounting/doctor-review-confirm-return-contract.md#review-actions-storage)へ進めたが、DDL/model・同TXprovider/readerは未供給で、既存列・完全履歴とはしない。doctor_review_recordsはrecord側であり、billing digestだけから過去会計内容を復元せず、現在source補修や481 v1無断拡張もしない。物理供給を待たず内部契約を選定したが、確認履歴機能Readyの不足は解消していない。

用法・経路・dose metadata/pattern revision、価格/sort・技術linkだけの変更は、対象identity/quantity/record集合が同じなら確認世代/digest不変。metadata-onlyをunsupportedにして実質的に会計停止する条件も作らない。薬量安全確認等の別目的gate・監査・履歴は保持する。対象signatureの評価自体が不明/失敗なら未評価として停止し、全dose writer受入済みとはしない。

<a id="doctor-review-plan-generation-writers"></a>
### 3.3 純粋世代参加と実writer供給

[before/final世代参加](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-generation)と[owner/AC台帳](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-owner-packets)へ。before/final双方の所属を含め、削除・移管前や同record別行を請求するbillingも落とさない。pure calculatorはIO/lock/write/commitをせず、実影響先列挙・phantom保護・期待CAS・監査・receiptの最外TX参加は未供給。NULL baselineをread/replayから初期化しない。

review_sets.billing_idは必須で、新準備workflow・新規billing Complete互換は別の未閉鎖packet。既存billingのConfirm wireは[追加契約](../../design/domains/accounting/doctor-review-confirm-return-contract.md#review-actions-confirm)で選定済みであり、runtime供給は未完。既存Completeのbillingなし新規作成を禁止したり、他billing確認を流用したりしない。本内部契約から製品Completeへ新禁止条件を追加しない。

<a id="doctor-review-plan-c06-candidate"></a>
## 4. 既存billingの選定契約と別packet

[Confirm/ReturnのC06・認可・response](../../design/domains/accounting/doctor-review-confirm-return-contract.md#review-actions-c06)を正本とする。確認した完全集合/世代/digest・理由/memoはimmutable command/Contentに、billing版/現在参照はPreconditionsへ分離する。同key成功/unknownは元target現在view認可後、CAS・医師資格・readinessより先に処理する。Q/ClaimMetadataWriterは実在を再利用するが、consumer anchor/供給は未完である。

<a id="doctor-review-plan-wire-dependencies"></a>
### 4.1 wire・保存・医師reader

[exact wire](../../design/domains/accounting/doctor-review-confirm-return-contract.md#review-actions-wire)、[set/Return/anchor保存差分](../../design/domains/accounting/doctor-review-confirm-return-contract.md#review-actions-storage)、[旧入口封鎖](../../design/domains/accounting/doctor-review-confirm-return-contract.md#review-actions-bypass)を選定した。doctor_review_setsのbilling_signature/operation_id、doctor_review_returns、billing_review_operation_targetsは必要供給であり現DDLに存在済みとはしない。旧GET GetOrCreateをpure readとして温存せず、旧APIから推定billing・架空確認へ変換しない。共有医師predicateのdesignation検証・ambient/NOWAIT adapterは未供給、Returnへ新医師限定なし。

既存Completeの初回201/再送200・Billing整数wire・元請求日は維持する。[準備billing/新規Complete](../../design/domains/accounting/doctor-review-confirm-return-contract.md#review-actions-complete-dependency)は別packetで、新billing_id必須や製品のbillingなしComplete禁止を追加しない。481/484/495/518の既存Content/CAS/receiptは各採択契約を維持する。

<a id="doctor-review-plan-lock-candidate"></a>
## 5. 専用lock順の選定と供給残

[専用TX/lock順](../../design/domains/accounting/doctor-review-confirm-return-contract.md#review-actions-locks)を正本とする。短claim終了→最外TXのreservation-graph:write:v1 try gate→binding NOWAIT→元target現在view/success/unknown→savepoint/Apply。必要clinic/staff/assignment保護→billing→Confirmだけrecords/plan/dependency/treatment/billing_items/active bindingsを取得し、未宣言上位lockが必要なら最外TXをやり直す。Returnはcurrent projection全量を要求しない。

481のbilling→必要close→record/sourceと484 source-onlyのeditrights→record/source、Complete close→billingは再定義せず、共通gate上位直列化と対象writer閉包の供給を必要とする。Confirm/Returnに不要なclose/fund/method/appointmentや506 participantを追加しない。ContextGate/try-gate、current provider、auth/staff、membership/writer参加は未供給で、gateだけで安全とはしない。現在Runner Gate(ctx)errorをContextGate実装済みとしない。

Confirm/Returnはclinical世代/digest不変。billing参照変更のversionは既存storage guardが一度増分し、手動version+1なし。mutation/履歴/監査/receipt失敗はcommit前だけ全rollback、commit後送信失敗は永続成功を反転させず元key lookupへ進む。

<a id="doctor-review-plan-owner-handoffs"></a>
## 6. 今回の参加台帳成果: writer閉包・共有owner・必要subset

| 入口/共有path | 既存機能owner | 今回必要なsubset/activation判定 |
|---|---|---|
| billing/accounting_complete.go, accounting_complete_tx.go | **462単一統合owner** | gate/集合/generation/digest/484再検査/483/505/506/489接点。各ownerが同ファイルを並行編集しない |
| billing_confirmation_service.go/repository.go、GET/confirm/return | 462、445 API | 旧GetOrCreate INSERT、旧table/status JOINを移行。GET純read、confirm/returnの対象billing明示と単一確認writer。旧互換routeが新gateを迂回しない |
| billing_item_service_create.go/update.go、unbilled/revision、billing_item_repository_unbilled.go | 481＋462受領、単一reader編集owner | 直接追加/削除/数量、source import/release/復元、manual/unknown、完全source列挙・医師確認projection。金額のみと臨床変更を分類 |
| medicalrecord/treatment_service.go Create/Update/Delete/BulkUpdateSortOrder、dose helper/各ambient writer | treatment domain / EMR-495の該当writer owner | source先取り禁止。承認済みidentity/quantity差とmetadata/技術link-only不変を対照し、薬量安全gate・監査は別責務。bulk/copy/correction/importも同cohortへ届く経路だけ列挙 |
| 518 plan/revision/link coordinator、516 link/転記 | 518/516 | charge_source/execution_sourceの別identity、両source link差替え、remove/restore/reparent、必要baseline。全516完成待ちなし |
| 484 fulfillment writer/reader | 484 | current source/context/世代/event-headの再検査。preparedのみでもrevisionを保持。readiness変更は医師確認集合を自動書換えしない |
| 495 B/C・通常guard、494 edit-right | 495/494 | P1既決を保持。対象の限定失効signatureを変える訂正/複写だけ462世代へ参加、終了eventだけは増分しない。医師確認を通常編集禁止へしない |
| billing totals/snapshot | 505 | 初回外税等、算式契約が閉じたsubsetのみ。税込未閉鎖を全体Done待ちにしない |
| acknowledgement | 483 | price-onlyのbefore/after・必要了承同TX。医師確認を失効させない。awaiting_price_acknowledgementの未適用停止を維持 |
| funds/payments/確定event | 506 | Completeが今使う限定ambient subset/必要lock/refだけ。一般506新wireは本案の採択対象外 |
| operation/{contract,runner,store,dto}.go | 489 | 実在Q/claim hookを再利用、元billing anchor resolverとconsumer参加・同TX成功だけ。共通再実装なし |
| routes.go/request/response/backend/docs/api.yaml | 445/API単一owner | 既存Complete wire維持、confirm/returnとlookup adapterの正式収録。生成は本人指示の扱いを維持 |
| 001/ERD、model/accounting.go/storage_finance.go | 446/466、既存model共有owner | 保存先再選定なし。選定済みsignature/Return/anchorの必要保存差分。DDL/model適合と供給SHAは未受領、DDL適用なし |

各行に「実行ID、現在claim、具体hunk、供給SHAまたは同PR合意、AC/残AC、旧writer拒否or移行」を記入してからimplementationへ渡す。現在4票のassigneeは空であり、機能owner表を個人割当済みと読み替えない。495の現claimは最新コメントのclaim/EMR-495を尊重し別claimしない。

全writerを全領域一括で待つのではなく、選択cohortの集合・世代を変え得る入口を閉じる。未対応adapterを候補集合から隠して「完全」としない。旧GET初期化・旧PATCHによる確定・旧confirm/return・CSV/job/importの直接writeをactivation漏れにしない。未知経路は調査残とし、存在しない実装を捏造しない。

<a id="doctor-review-plan-posthumous-discharge"></a>
## 7. 生前分C・退院Dは別の有限契約として残す

通常Aに死亡後client flagを足して全死亡gateを解除しない。生前Cの最初の対象も同じ一sourceで、正規readerが現在life state・死亡根拠・生前実施sourceの不変identity/版・実施日時根拠・元owner/record/billingを証明する。created_at/会計日/現在ownerを生前根拠にしない。nil life readerはinternal unavailable/500で停止し、既存491 ReadPetLifeState adapterの必要subsetへ統合する。不存在/不明は生存にしない。

Cでは現在life/death writerと同じgate・関連petの保護順を受領し、source/明細書込前に再検査する。時刻/出典が不明なら専用reconciliationへ返し、通常医師確認/readiness/締め後保護は免除しない。通常金融operation/元billing anchorを再利用し、死亡後専用key namespaceを作らない。既存請求回収へ新医師確認を一律追加しない。

Dは元stay/退院明細/再利用billing/必要405 source event/元owner/版の型付きread contractとambient participantを受領後に接続。Aの既存billing使用を「全退院takeover対応済み」としない。どちらも復活/新臨床行為/自動返金/出庫/予約再開なし。C/Dの公開wire/証明readerの具体契約はこのA案から既決とせず別成果にする。

<a id="doctor-review-plan-acceptance"></a>
## 8. reader限定1PRの完了条件と後続fixture

選定した先行の非活性内部契約は、(a) §3のclosed型・decimal/JCS入力/digest規則、(b) 完全集合の比較関数の入出力/エラー区分、(c) writerの効果分類と最外TXへの参加要求、(d) 未供給matrixを含む。今回追加のConfirm/Return契約も文書選定までで、製品への配線・API YAML・実DDLは達成対象にしない。fixtureコード作成とgolden bytes/hash生成は後段で、今回実物を作成・実行済みとはしない。

内部契約の限定技術選定と、文書の公開・実供給・後続機能契約採択を区別する。次の不足を「単に試験NOT_RUN」へ隠さない: 既存481 identityの現在採用値への適合providerと、確認時限定signature不変保存の物理供給、共用医師readerとdesignation受入根拠、選定済みconfirm/return wire・Return履歴・元billing anchorの実供給、準備billingと新規Complete互換、各cohortの必要lock資源宣言。したがって**full doctor-review/Complete機能sliceはReadyではない**。以下は後続機能接続で必要なfixture一覧。

すべてfixture仕様のみ、実行NOT_RUN。

1. 未確認/NULL世代/片方だけNULL/不正長digest/他院他billing/余剰欠落重複record/古いreviewは準備可・Complete不可。
2. record集合追加削除、数量1→2→1、臨床identity差替えは再確認。単価/割引/表示順/同値と用法・経路・dose metadata/pattern revision・技術link-onlyは、identity/quantity/集合が同じなら世代/digest不変。金額変更は必要な483了承を別判定。
3. 同期・直接明細変更・医師確認・Completeの各二先行順で誤通過なし。差異だけ保存は会計適用成功ではない。旧GETが確認を作らない。
4. 現世代take_home prepared+内容数量確認済みはclinical pending単独で拒否しない。deferred/cancelled/removed/unknown、revoked/旧世代readyは不可。請求/確認/prepareで在庫不変。
5. claim winner/loser、同key同内容CAS更新retry、同key別集合、成功応答喪失、sticky unknown、認可失効、元target欠損、別URL、lookup障害を分離。成功replayで現在master/CAS/readyを検査し直して失敗へ反転しない。
6. 確認後にbilling itemsが変わっても確認時の会計側限定signature不変内容を再読でき、record側内容と混同しない。digestだけ・現在sourceからの補修では履歴受入不可。世代/CAS/旧確認は[条件別fixture](#doctor-review-plan-generation-fixtures)で検証し、一律に両世代を増分しない。overflow、commit前の必須処理（監査・履歴・差異・receipt・同TXで作成を要求する読取projection等）の失敗だけを全rollbackの対象にする。projection digestのgolden bytesと空/null/unknown/decimal同値を固定。
7. billingなしCompleteの既存新規作成機能を無断禁止しない一方、架空確認/他billing確認を作らない。準備/新規互換adapterの閉鎖後に初回Complete=201、同成功再送=200、元請求日不変を検証。
8. Cは死亡登録対保存両順序、生前証明消失/変更/不明、主飼主変更、nil reader。Dはstay/source/再利用billing不一致。Aとは別受入。

<a id="doctor-review-plan-generation-fixtures"></a>
### 条件別の世代・CAS・確認fixture（提案、NOT_RUN）

| fixture条件（成功する最外TX） | record世代 | billing世代 | 期待CAS・旧確認の対照 |
|---|---|---|---|
| recordの承認済みsignatureに実差（追加/削除・identity・数量） | 該当recordだけ+1 | 影響するbillingだけ+1 | 変更対象の期待世代CASを検査。変更後の世代/digest不一致で該当する旧確認を失効投影し、保存済み確認履歴は保持 |
| billing-onlyの承認済みsignatureに実差 | 不変 | 該当billingだけ+1 | billingの期待世代CASを検査。recordを会計変更で増分せず、変更した会計集合の旧確認は失効 |
| billingのrecord集合だけを追加/削除 | 既存recordは不変 | 該当billingだけ+1 | billing期待世代CASと完全record集合を照合。所属集合が違う旧確認は無効、既存record内容/世代を変更しない |
| 同値/price-only/sort-only/metadata-only/technical-link-only（対象signature同一） | 不変 | 不変 | 既存writerのCAS等は省略せず、臨床集合の確認はこの変更だけでは失効しない。必要な483価格了承は別判定 |

metadata-only/technical-link-onlyは対象signatureが同じなら世代/digest不変で、unsupportedによる会計停止へ変換しない。対象identity/quantityの評価自体ができなければ未評価で止める。薬量安全確認等の別gate・監査は維持する。影響先はbefore/final双方の所属から列挙し、削除・移管前の依存billingも保持する。同じ最外TXに複数変更があっても各変更対象の増分は一回だけ。依存record世代を再読したreaderは再増分しない。期待CAS不一致なら業務値・世代・履歴・receiptはzero-writeで、旧確認を変更しない。成功時の失効と不変、失敗時の不変を別fixtureにする。

commit後のHTTP本文構築/送信失敗では、永続成功を失敗へ反転させず、rollbackしたとも報告しない。HTTPエラーが絶対に返らないとは約束しない。clientの結果不明は元key lookupへ進み、commit成否不明も新key再送で迂回せず既存unknown規則に従う。commit前の必須projection生成失敗とcommit後の応答喪失を別fixtureにし、選定契約を実行済みにしない。

完了は「技術契約選定」「正本公開」「必要subset供給」「consumer接続」「全対象writer閉包」「実試験」「activation」を別々に記録する。内部reader型/純粋参加計画は限定選定済みだが、必要供給以降の完了ではない。非活性reader/projectionの契約整備を先行し、通常Completeへの新しい禁止条件は採択しない。billingなし新規Complete/複数record/他source/C/Dと上記未供給を見えるまま保持する。

<a id="doctor-review-plan-references"></a>
## 9. 追加の固定参照

- [481採択全文](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-481-source-binding-contract-20261021.md)
- [484採択全文](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/work/emr/emr-484-fulfillment-writer-contract-20261021.md)
- [C06参加/最新Q実装状態](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/design/shared/operation-participation.md)
- [実operation contract](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/operation/contract.go)
- [医師確認gate正本](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/design/domains/medical-records/technical-spec.md#doctor-confirm-gate-revision-20261020)
- [会計の版と医師確認](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/design/domains/accounting/technical-spec.md#review-version-sync-proposal)
- [旧確認service移行接点](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/billing_confirmation_service.go)
- [実route/grant](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/routes.go#L100-L138)
