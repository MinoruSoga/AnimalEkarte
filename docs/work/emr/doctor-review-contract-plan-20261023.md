# 医師確認集合：内部reader・writer参加の設計作業案（2026-10-23）

**内部契約の限定技術選定と後続計画整備。製品実装Ready・公開HTTP契約採択・実行受入ではない。**

| 項目 | 内容 |
|---|---|
| 文書ID | work.emr.doctor-review-contract-plan-20261023 |
| 種別 | 選定済みINTERNAL契約への接続と後続設計作業案 |
| 決定状態 | T06内部projection/digest/純粋世代参加と論理保存・共用医師readerの方向は2026-10-23選定。物理保存/HTTP/総lock合成等は未供給・未選定 |
| 実装状態 | 候補の供給・接続・全writer閉包は未完 |
| 実機検証状態 | NOT_RUN（tests/build/codegen/DB/Docker/browser/設定変更なし） |
| 文書確認日 | 2026-10-23（資料受領・静的整形） |

[今回の索引・判定軸](emr-limited-audit-plans-index-20261023.md#limited-audit-status)を参照。以下の監査範囲・件数は受領した固定基準の報告であり、今回再度の全件照会を行ったという意味ではない。

2026-10-23のroot技術選定により、§3の非活性内部境界は[技術契約](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-scope)へ集約した。以下の§4–7の公開wire・anchor・総取得順等は引き続き未採択の後続案である。内部選定、文書公開、必要供給、consumer接続、全writer閉包、実行受入、activationを区別し、担当/実行ID/供給SHA受領済みとはしない。

<a id="doctor-review-plan-scope"></a>
## 1. 結論と確認済み根拠

今回選定した限定成果は、非活性の内部typed reader/projection契約と純粋世代参加計画・合成fixture仕様である。新HTTP route、医師確認write、準備billing作成、Complete gate有効化、anchor DDL、既存通常Completeの動作変更は含めない。全体機能sliceをReady/閉鎖済みとはしない。以下の§4–7は後続実装の依存/検討表であり、そのまま採択可能なclosed API/物理契約ではない。

readerは一つの既存billingの完全な現在record集合（複数record可）とplan-backed treatmentを扱う。各linkを対応する同recordへ照合する内部範囲であって、製品の通常Completeをこの範囲に制限する業務判断ではない。billingなしComplete、新規作成、他sourceは今後の接続残ACとして保持する。生前分Cと退院takeover Dも別slice。選んだcohort内のwriter閉包を点検できる成果を先に作り、全source/全495/全489のDone待ちにしない。

当初の調査基準は `171b56575e7ee028e54358dad22ae28574853125`（PR843 merge）。今回の選定は `b81abc26b3657bd662475a1f55b912672a24a265` を基準に再照合した入力を受領し、固定根拠を[内部契約](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-evidence)へ置いた。以下の既存171b565リンクは当初観測の固定参照として保持する。Planeは462本文と8コメント、481本文と3コメント、484本文と4コメント、495本文と15コメントを取得し、各responseの next_page_results=false を確認。過去の未merge/未選定記述を現在条件にしない。

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

[exact shape](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-shapes)、[canonical digest](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-canonical-digest)、[論理保存方向](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-storage-direction)へ。既存billing_itemsのschema-tagged current projectionと既存doctor_review_setsの不変billing preimageを論理方向として選んだ。列名・nullable/CHECK・migration/model/ERD・producer/read adapterは未供給で、既存列・完全履歴とはしない。doctor_review_recordsはrecord側であり、billing digestだけから過去会計内容を復元せず、現在source補修や481 v1無断拡張もしない。物理供給を待たず内部契約を選定したが、確認履歴機能Readyの不足は解消していない。

dose_weight/source・dose_param_snapshot等のv1外効果を無条件no-opにせず、完全に分類できないwriterはunsupported/clinical_content_shape。全dose writer適合や確認activationの証拠にしない。

<a id="doctor-review-plan-generation-writers"></a>
### 3.3 純粋世代参加と実writer供給

[before/final世代参加](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-generation)と[owner/AC台帳](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-owner-packets)へ。before/final双方の所属を含め、削除・移管前や同record別行を請求するbillingも落とさない。pure calculatorはIO/lock/write/commitをせず、実影響先列挙・phantom保護・期待CAS・監査・receiptの最外TX参加は未供給。NULL baselineをread/replayから初期化しない。

review_sets.billing_idは必須で、新準備workflow・新集合confirm wire・新規billing Complete互換は別の未閉鎖packet。既存Completeのbillingなし新規作成を禁止したり、他billing確認を流用したりしない。本内部契約から製品Completeへ新禁止条件を追加しない。

<a id="doctor-review-plan-c06-candidate"></a>
## 4. 後続依存: C06参加草案と公開互換（このPRで採択しない）

| 操作 | Content（業務内容） | Preconditions（実行前提） | 同TXの結果 |
|---|---|---|---|
| 医師確認（案 billing.doctor_review.confirm） | billing_id、確認対象billing/record世代とdigest、完全record集合、exact note commitment | billing CAS、現在record/source/行CAS、適用される現在fence | set/全records/現在参照/監査/receipt |
| 確認差戻し（案 billing.doctor_review.return） | billing_id、対象review_set_id、既存必要理由のcommitment | 現在参照、billing CAS、適用fence | 旧確認を不変保存し現在参照解除、監査/receipt |
| 既存Complete adapter | 既存Completeの業務payloadと対象billing/source identity、数量/金額/支払内容 | unbilled revision、現在billing/record/source CAS、現在有効なreview_set、現readiness/ack条件 | 既存Billing response相当の固定成功、確定event/必要fund/totals/監査/receipt |
| 481 import/resolution | 採択済みsource_stamp/decision/ack commitment | 採択済みbilling/item/binding/records等 | 481 refs/versionsを変更しない |
| 484 performed/prepared/revoke | 採択済みgenerationはContent、owner-link/kind/reason等 | plan/source/event-head/fence CAS | 採択済み484 refs/versions。医師確認を自動生成しない |
| 495/518/treatment通常writer | 各既存intentのimmutable command | 各既存CAS/fence | 元operation一つに462世代＋481同期/差異＋484 contextのambient参加 |

確認ではユーザーがこの集合を確認するため、確認対象generation/digestはContent。同keyで別集合へ差替え不可。Completeでは医師確認は業務payloadを変えない現在有効性のPreconditionで、再読後の同内容retryを許容する。別source/数量へ変えるのは別Content。既決481 source_stamp、484 source_generationをCAS扱いへ変えない。

scopeは認証医院・同actor、金融ownerはbilling。Completeの実際のintent識別子は現統合ownerの登録を再利用し、本案で第二namespaceを追加しない。新confirm/return案の名前はAPI単一ownerの登録差分として採択対象。Qは現operation.NewVerifiedCommandを利用し、Applyはreq.VerifiedCommandのみからdecodeする。別DTO closureとfingerprintを二系統にしない。

元targetはclaim winnerの同短TXで不変anchorとして保存する。案 `billing_review_operation_targets(clinic_id,operation_id,billing_id)`、PK(clinic,operation)、binding複合FK、immutable/RLS、業務billing FKなし。既存同等anchorの供給があれば再利用し二表にしない。claimには業務row lockを入れない。success前/拒否後/unknownもbinding+clinicから元billingを解決し、URL/requestをfallbackにしない。missing anchorは非開示。未確定bindingをGETで修復しない。

新規claim前は現在write/scope preflight、確定的なbinding不在だけclaim。lookup障害を不在扱いしない。既存bindingは先に元target現在accounting:viewを検査し、成功なら現在CAS/readinessに先行して保存結果を返す。

<a id="doctor-review-plan-wire-dependencies"></a>
### 4.1 現存wire/権限/理由の証拠と未閉鎖範囲

- routes.go L103–106の既存confirm/returnはいずれもaccounting:edit、GETはaccounting:view。handlerはExtractStaffIDをconfirmed_by/returned_byへ注入し、client指定の代理医師IDを受け付けない。担当医以外の資格ある医師が代行できる業務規則と、actorと確認doctorを別IDで保存する仕組みは別である。**後者は本案では採用しない**。
- 現存confirm bodyの許可keyはmemoのみ、string、TrimSpace後1000 Unicode rune以下、省略は空文字。return bodyの許可keyはreturn_reason/memoのみ、string。TrimSpace後return_reason必須・500 rune以下、memo1000 rune以下。両方8 KiB body上限、application/json、成功200の旧BillingConfirmation response。新billing/世代/digest/operation_keyはこのwireにない。これは現行実装の観測であり、そのまま新医師確認集合wireとして使えるという意味ではない。
- service Confirm/Returnは現在staffとclinic assignmentをrepository.LockActiveStaffAssignmentで検査し、staffs.deleted_at IS NULL、staffs.is_active=TRUE、staff_clinic_assignments.deleted_at IS NULLをSHARE lock下で照合する。確認した関数は職種/医院が認めた医師資格を検証していない。accounting:edit＋active assignmentを医師資格の代用にしない。
- 新集合confirmの医師readerは[T03/T06共用の既存predicate再利用方向](../../design/domains/accounting/doctor-review-internal-projection.md#review-internal-physician-dependency)を選定済み。実provider・designation対応根拠・保護順、既存return_reasonの新監査/履歴保存先と正規化後Content commitmentは**後続462/API/既存staff ownerの明示依存**。今回のreader仕様はこれらを閉じたとしない。旧表のreturn_reasonを復活させず、保存先未確定のままReturn機能Readyにしない。
- 既決の業務判断は再質問しない。正本が定める医院が認めた医師・担当医以外の代行可は維持し、技術readerの既存機構への対応付けを実装ownerへ渡す。

固定根拠: [request/理由上限](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/billing_confirmation_request.go)、[handler/actor注入](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/billing_confirmation_handler.go)、[repository/assignment検査](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/backend/internal/billing/billing_confirmation_repository.go)、[業務上の確認actor規則](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/design/domains/reservations/technical-spec.md#reservation-receiving-contracts)。

confirm成功refs案: billing_id,doctor_review_set_id。versions案: billing,clinical_set_generation。record別詳細は固定typed result detailに置き、C06 mapへJSON配列を押し込まない。Returnは元set refと保存後billing版。Completeは既存**初回201/成功再送200・Billing wire・整数wire**を保持。新確認adapterの400 validation、409 stale/dependency/key conflict、202 busy、503 recorded unknown、403現在非開示、404同scope bindingなしをC06から写像する案。現在再認可後のresult-store障害500と記録済みunknown503を分離。既存Complete全error envelopeの破壊変更はしない。

<a id="doctor-review-plan-lock-candidate"></a>
## 5. 後続依存: 総取得順の合成案（共有owner受領前・activation不可）

既決[予約§6.4](https://github.com/MinoruSoga/AnimalEkarte/blob/171b56575e7ee028e54358dad22ae28574853125/docs/design/domains/reservations/technical-spec.md#reservation-global-lock-contract)は共通gateで既存Complete close→billingとrefund/item billing→closeの逆辺を保護する。**全ownerに一つのFK順を強制する契約ではない**。481 §7と484 §5の部分順序を単純に互いのservice呼出しへネストさせない。

以下は未採択の総取得順案である。このsliceの実行plannerが業務lock前に必要全資源を列挙し、以下の列の順で取得する。各集合内は (clinic_id, typed kind, id) の昇順。同物理TX内で既取得の同じlockは再取得せず、ambient participantへ渡す。

1. 読取preflight/claim短TXを終了。既存scope/keyのbinding+元targetを固定。
2. 最外業務TX → 既存reservation-graph:write:v1 gate → C06 binding NOWAIT。ここまで業務row/policy/close/staff lock先取りなし。
3. 現在元target read認可 → 成功/unknownを先に返す。未成功だけsavepoint、現在write適格性と現在対象集合再発見。上位集合がpreflightと変わればwhole TX再開始、下位後に追加しない。
4. 予約状態を触るCompleteだけ必要staff-global→clinic-booking→pet-intent/lifecycle→appointment資源を既存順で先に確保。通常source writerへ不要な予約資源lockを要求しない。
5. 金融上位: **481 sync/import/confirmは関連billing ID昇順→必要close境界**。**既存Completeは既存close境界→対象billing**を維持。両者は共通gateにより同時に下位へ入らない。新fund操作/預り金lockを本sliceで設計し直さず506 ambient subsetの宣言を受領する。必要上位資源の未宣言はactivation不可。
6. 必要な494 edit-right行→medical_records ID昇順→既存plan rows ID昇順→plan_items dependency row→typed treatment ID昇順→billing_items ID昇順→active billing_source_bindingsの主キー順→履行event/referenceとreview set参照。
7. 現在集合/世代/digest、readiness、ack、税計算、現在grantを全部再検査し、writer/監査/receiptを同最外TXへ書く。commit前の必須処理失敗は全rollback。commit後のHTTP本文構築/送信失敗はこのrollback条件に含めず、永続成功を失敗へ書換えない。

上記は「入れ子484 participantに後から会計lockを取らせる」方式ではない。484/treatment/planの入口が、source lock前に481影響billingを発見して段5まで確保する。membership/link追加削除も同gateなので発見後phantomを許さない。全writerがgateを持たない段階ではこの議論は成立せず未activation。既存source単独484は金融資源集合が空のときだけ段5をskipする。

receipt replay/lookupは新業務lock/CAS/fenceを要求せず元target現在readのみ。495 P1の通常編集許可と本sliceの全writer適合判定は別であり、T06未接続を理由に全子writerを停止へ巻き戻さない。

<a id="doctor-review-plan-owner-handoffs"></a>
## 6. 今回の参加台帳成果: writer閉包・共有owner・必要subset

| 入口/共有path | 既存機能owner | 今回必要なsubset/activation判定 |
|---|---|---|
| billing/accounting_complete.go, accounting_complete_tx.go | **462単一統合owner** | gate/集合/generation/digest/484再検査/483/505/506/489接点。各ownerが同ファイルを並行編集しない |
| billing_confirmation_service.go/repository.go、GET/confirm/return | 462、445 API | 旧GetOrCreate INSERT、旧table/status JOINを移行。GET純read、confirm/returnの対象billing明示と単一確認writer。旧互換routeが新gateを迂回しない |
| billing_item_service_create.go/update.go、unbilled/revision、billing_item_repository_unbilled.go | 481＋462受領、単一reader編集owner | 直接追加/削除/数量、source import/release/復元、manual/unknown、完全source列挙・医師確認projection。金額のみと臨床変更を分類 |
| medicalrecord/treatment_service.go Create/Update/Delete/BulkUpdateSortOrder、dose helper/各ambient writer | 既存treatment/461/495 owner | source先取り禁止。quantity/identity/dose差分とsort/no-op対照。bulk/copy/correction/importも同cohortへ届く経路だけ列挙 |
| 518 plan/revision/link coordinator、516 link/転記 | 518/516 | charge_source/execution_sourceの別identity、両source link差替え、remove/restore/reparent、必要baseline。全516完成待ちなし |
| 484 fulfillment writer/reader | 484 | current source/context/世代/event-headの再検査。preparedのみでもrevisionを保持。readiness変更は医師確認集合を自動書換えしない |
| 495 B/C・通常guard、494 edit-right | 495/494 | P1既決を保持。対象clinical projectionを変える訂正/複写だけ462へ参加、終了eventだけは増分しない。医師確認を通常編集禁止へしない |
| billing totals/snapshot | 505 | 初回外税等、算式契約が閉じたsubsetのみ。税込未閉鎖を全体Done待ちにしない |
| acknowledgement | 483 | price-onlyのbefore/after・必要了承同TX。医師確認を失効させない。awaiting_price_acknowledgementの未適用停止を維持 |
| funds/payments/確定event | 506 | Completeが今使う限定ambient subset/必要lock/refだけ。一般506新wireは本案の採択対象外 |
| operation/{contract,runner,store,dto}.go | 489 | 実在Q/claim hookを再利用、元billing anchor resolverとconsumer参加・同TX成功だけ。共通再実装なし |
| routes.go/request/response/backend/docs/api.yaml | 445/API単一owner | 既存Complete wire維持、confirm/returnとlookup adapterの正式収録。生成は本人指示の扱いを維持 |
| 001/ERD、model/accounting.go/storage_finance.go | 446/466、既存model共有owner | 保存先再選定なし。必要anchor/制約差だけ。適合と供給SHAの受領、DDL適用なし |

各行に「実行ID、現在claim、具体hunk、供給SHAまたは同PR合意、AC/残AC、旧writer拒否or移行」を記入してからimplementationへ渡す。現在4票のassigneeは空であり、機能owner表を個人割当済みと読み替えない。495の現claimは最新コメントのclaim/EMR-495を尊重し別claimしない。

全writerを全領域一括で待つのではなく、選択cohortの集合・世代を変え得る入口を閉じる。未対応adapterを候補集合から隠して「完全」としない。旧GET初期化・旧PATCHによる確定・旧confirm/return・CSV/job/importの直接writeをactivation漏れにしない。未知経路は調査残とし、存在しない実装を捏造しない。

<a id="doctor-review-plan-posthumous-discharge"></a>
## 7. 生前分C・退院Dは別の有限契約として残す

通常Aに死亡後client flagを足して全死亡gateを解除しない。生前Cの最初の対象も同じ一sourceで、正規readerが現在life state・死亡根拠・生前実施sourceの不変identity/版・実施日時根拠・元owner/record/billingを証明する。created_at/会計日/現在ownerを生前根拠にしない。nil life readerはinternal unavailable/500で停止し、既存491 ReadPetLifeState adapterの必要subsetへ統合する。不存在/不明は生存にしない。

Cでは現在life/death writerと同じgate・関連petの保護順を受領し、source/明細書込前に再検査する。時刻/出典が不明なら専用reconciliationへ返し、通常医師確認/readiness/締め後保護は免除しない。通常金融operation/元billing anchorを再利用し、死亡後専用key namespaceを作らない。既存請求回収へ新医師確認を一律追加しない。

Dは元stay/退院明細/再利用billing/必要405 source event/元owner/版の型付きread contractとambient participantを受領後に接続。Aの既存billing使用を「全退院takeover対応済み」としない。どちらも復活/新臨床行為/自動返金/出庫/予約再開なし。C/Dの公開wire/証明readerの具体契約はこのA案から既決とせず別成果にする。

<a id="doctor-review-plan-acceptance"></a>
## 8. reader限定1PRの完了条件と後続fixture

選定した非活性内部契約は、(a) §3のclosed型・decimal/JCS入力/digest規則、(b) 完全集合の比較関数の入出力/エラー区分、(c) writerの効果分類と最外TXへの参加要求、(d) 未供給matrixを含む。製品への配線・API・DDLは同PRの達成対象にしない。fixtureコード作成とgolden bytes/hash生成は後段で、今回実物を作成・実行済みとはしない。

内部契約の限定技術選定と、文書の公開・実供給・後続機能契約採択を区別する。次の不足を「単に試験NOT_RUN」へ隠さない: 選定した会計側current/不変projectionの物理供給、共用医師readerとdesignation受入根拠、confirm/return exact wireとReturn理由保存先、準備billingと新規Complete互換、元billing anchor物理契約、506を含む全lock資源宣言。したがって**full doctor-review/Complete機能sliceはReadyではない**。以下は後続機能接続で必要なfixture一覧。

すべてfixture仕様のみ、実行NOT_RUN。

1. 未確認/NULL世代/片方だけNULL/不正長digest/他院他billing/余剰欠落重複record/古いreviewは準備可・Complete不可。
2. record集合追加削除、数量1→2→1、臨床identity差替えは再確認。単価/割引/表示順/同値は世代不変。金額変更は必要な483了承を別判定。
3. 同期・直接明細変更・医師確認・Completeの各二先行順で誤通過なし。差異だけ保存は会計適用成功ではない。旧GETが確認を作らない。
4. 現世代take_home prepared+内容数量確認済みはclinical pending単独で拒否しない。deferred/cancelled/removed/unknown、revoked/旧世代readyは不可。請求/確認/prepareで在庫不変。
5. claim winner/loser、同key同内容CAS更新retry、同key別集合、成功応答喪失、sticky unknown、認可失効、元target欠損、別URL、lookup障害を分離。成功replayで現在master/CAS/readyを検査し直して失敗へ反転しない。
6. 確認後にbilling itemsが変わっても確認時の会計側不変内容を再読でき、record側内容と混同しない。digestだけ・現在sourceからの補修では履歴受入不可。世代/CAS/旧確認は[条件別fixture](#doctor-review-plan-generation-fixtures)で検証し、一律に両世代を増分しない。overflow、commit前の必須処理（監査・履歴・差異・receipt・同TXで作成を要求する読取projection等）の失敗だけを全rollbackの対象にする。projection digestのgolden bytesと空/null/unknown/decimal同値を固定。
7. billingなしCompleteの既存新規作成機能を無断禁止しない一方、架空確認/他billing確認を作らない。準備/新規互換adapterの閉鎖後に初回Complete=201、同成功再送=200、元請求日不変を検証。
8. Cは死亡登録対保存両順序、生前証明消失/変更/不明、主飼主変更、nil reader。Dはstay/source/再利用billing不一致。Aとは別受入。

<a id="doctor-review-plan-generation-fixtures"></a>
### 条件別の世代・CAS・確認fixture（提案、NOT_RUN）

| fixture条件（成功する最外TX） | record世代 | billing世代 | 期待CAS・旧確認の対照 |
|---|---|---|---|
| record clinical projectionに実差 | 該当recordだけ+1 | 影響するbillingだけ+1 | 変更対象の期待世代CASを検査。変更後の世代/digest不一致で該当する旧確認を失効投影し、保存済み確認履歴は保持 |
| billing-only clinical projectionに実差 | 不変 | 該当billingだけ+1 | billingの期待世代CASを検査。recordを会計変更で増分せず、変更した会計集合の旧確認は失効 |
| billingのrecord集合だけを追加/削除 | 既存recordは不変 | 該当billingだけ+1 | billing期待世代CASと完全record集合を照合。所属集合が違う旧確認は無効、既存record内容/世代を変更しない |
| 同値/price-only/sort-only（supported shape内） | 不変 | 不変 | 既存writerのCAS等は省略せず、臨床集合の確認はこの変更だけでは失効しない。必要な483価格了承は別判定 |

v1外dose根拠など効果を分類できないwriterはunsupportedとし、上表のno-opへ押し込まない。影響先はbefore/final双方の所属から列挙し、削除・移管前の依存billingも保持する。同じ最外TXに複数変更があっても各変更対象の増分は一回だけ。依存record世代を再読したreaderは再増分しない。期待CAS不一致なら業務値・世代・履歴・receiptはzero-writeで、旧確認を変更しない。成功時の失効と不変、失敗時の不変を別fixtureにする。

commit後のHTTP本文構築/送信失敗では、永続成功を失敗へ反転させず、rollbackしたとも報告しない。HTTPエラーが絶対に返らないとは約束しない。clientの結果不明は元key lookupへ進み、commit成否不明も新key再送で迂回せず既存unknown規則に従う。commit前の必須projection生成失敗とcommit後の応答喪失を別fixtureにし、未採択案を実行済みにしない。

完了は「技術案採択」「正本公開」「必要subset供給」「consumer接続」「全対象writer閉包」「実試験」「activation」を別々に記録する。内部reader型/純粋参加計画は限定選定済みだが、必要供給以降の完了ではない。非活性reader/projectionの契約整備を先行し、通常Completeへの新しい禁止条件は採択しない。billingなし新規Complete/複数record/他source/C/Dと上記未供給を見えるまま保持する。

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
