# EMR-308/468 次回予定の明示消化・同ID確定：採択技術契約

2026-10-22。root技術選定を文書化。照合mainは`a5b4a9553b3f003e90ea47660a109c5d526be59c`。技術選定済みとAPI機械収録・DDL/model・共有Q・domain実装・受入を区別し、後者は未完。今回はdocsのみ、tests/build/codegen/DB/Docker/browserはNOT_RUN。merge・レビュー依頼は行わない。

## 1. 正本・目的・所有

業務判断窓口はMinoru Soga。[明示消化FR/AC](../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-next-intent-explicit-consumption)、[会計完了後の追加承認](../../spec/workflows/reservation-to-record-flow.md#next-intent-after-accounting-consumption-20261022)、[予約技術§4.8](../../design/domains/reservations/technical-spec.md#reservation-visit-history-intent-contract)が業務正本。[480契約§5–10](emr-480-booked-duration-selected-contract-20261021.md)、[458 HTTP契約§6](emr-458-option-assignment-selected-contract-20261021.md)、[C06/Q](../../design/shared/operation-participation.md#verified-command-selected-20261021)を再利用し、同じ操作に第2key/台帳を増やさない。

目的は、スタッフが選んだ元予定と今回予約の関係を一度だけ残し、予約/記録の重複作成・無選択の自動吸収・再送による二重消化を除くこと。既決の明示確認以外の手入力/確認工程、通常カルテや会計の再確定を追加しない。計画指標は下記2操作と全callerの対象・認可・保存・復旧・担当対応があること、未説明差分0。実業務の所要時間・操作数改善は未測定で、品質保証/PASSへ換算しない。自動消化は採用しない。

468 W04/W05が予約writerの単一統合owner、308は明示消化6ACの追跡。API単一owner445が機械収録、446/466は必要保存/model/ERD、489はQ/ClaimMetadataWriter/ContextGate・registry/lookup allowlistの必要subset。480のtrimming participantは同commandに統合する。現在の実装sessionへ無断追加せず、実装着手時の単一claimでhunk/供給SHAを確定する。全domain/全489 Doneを待つgateは設けない。

## 2. 共通境界と操作A/B

以下は実prefix `/api/v1`込み。認証staffのみ、clinic/actorはtrusted context。A/Bともclosed JSONで未知/重複key、不正UTF-8、複数JSON値を拒否。operation_keyは非zero小文字canonical UUID。新ID/版文字列は`1..9223372036854775807`の正規十進文字列（符号/先頭0/空白/指数なし）。既存Bの任意fieldのID wireを一括変更しない。時刻は既存RFC3339の受理精度を保ってUTCへcanonical化し、無断丸めしない。全応答no-store。

| 操作 | method/path・binding intent | 現在認可・対象 |
|---|---|---|
| A 既存今回予約への履歴紐づけ | `POST /api/v1/reservations/{appointment_id}/consume-next-visit-intent`、`reservation.consume_next_visit_intent` | 両対象に現在`reservations:edit`。read/lookupは`reservations:view`、両対象の同clinic/同pet・現在scopeを検査 |
| B 元予定の同ID確定 | 既存`POST /api/v1/reservations/{intent_id}/confirm-tentative`、`reservation.confirm_tentative` | 現在`reservations:edit`と既存予約適格性/警告/患者保護。read/lookupは`reservations:view` |

owner_kind=`reservation`、actor_kind=`staff`。Aは記録/billing本文を変更/公開しないため、そのgrantや本文versionを一律追加しない。会計完了後の許可はAの専用履歴操作だけで、通常編集/取消制限を解除しない。B後の別操作であるcheck-inのparticipant・必要権限を維持する。A/B自体へ受付record writerを追加しない。

### 2.1 A request・状態検証

body exact、全key必須：

```json
{"operation_key":"<UUID>","expected_version":"<今回予約版>","next_visit_intent":{"id":"<元予定ID>","expected_version":"<元予定版>"},"confirmed":true}
```

pathは今回予約。`next_visit_intent`もexact2keys、confirmedはtrueだけ（false/null/省略拒否）。元予定と今回予約は別ID、同医院・同患者、削除なし。元予定はactive tentativeかつ未消化。今回statusは`pending/confirmed/checked_in/in_consultation/accounting/completed`だけ許可し、`tentative/cancelled/no_show`・削除・未知/nullは拒否する。pending/confirmedを来院済み・実施済みへ偽装せず、Aで状態を進めない。取消済み/未来院（no_show）の予約を再利用しない。

今回予約の業務field/status、カルテ/専用記録の本文・状態、請求/入金は不変。関係追加のため今回予約versionとeventだけを一回進める。元予定の終端と両対象の履歴が片方だけ成立してはならない。Aを診療/会計の再保存や本文view要求へ拡張しない。

### 2.2 B request・既存入力のpresence

必須keyは`operation_key,expected_version,start_time,end_time`。任意keyは既存`reservation_type_id,doctor_id,is_designated,visit_type,notes,warning_receipt`と480条件下の`trimming_selection`だけ。旧callerの任意`status`は新commandで拒否し、serverが既存の確定規則に従って決定する（現defaultはconfirmed）。新confirm-and-check-in routeは作らない。

- 任意fieldの省略は保存値保持。reservation_type_idは既存型/nullable契約を維持し、nullを勝手に別区分へ補完しない。doctor_idは省略保持、明示null解除、正ID指定は同医院/現在適格性を検査し、0を解除sentinelにしない。
- notesは空文字で消去、null拒否。is_designatedはbool、visit_typeはfirst/revisitの既存検証。省略visit_typeを540候補から補完しない。warning_receiptは既存opaque警告証明で、C06成功receiptと別。
- trimming_selectionは480のexact4keys `{course_id,option_ids,expected_booking_revision_id,expected_adoption_digest}`。previewで見せたend_timeを送信し、既存manual overrideを維持。選択非変更の省略保持と、480が要求する作成/選択変更時のpreviewを区別する。コース必須や新料金を追加しない。

元予定の同IDを本予約へ確定する操作であり、Aの別ID関係を作らない。`source_medical_record_id`と由来を保持する。休診/時間/能力/容量/重複/現在患者等は480と既存予約契約どおり、clientのstatusやopaque hashを許可の根拠にしない。

現confirm後のbest-effort通常カルテ作成は、[受付時の既存記録再利用/失敗復帰](../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-record-reuse)へ追従させる必要subset。B成功receiptをカルテ作成/受付成功へ読み替えない。Bの成功を確認してから、既存check-inを別の明示操作として行う。B unknown中は後続に進まず、receipt成功replayでbest-effort作成を再発火しない。

## 3. immutable command・Content・Preconditions

480のContent.Fields exact4 string keysを再利用：`{schema_version:"1",resource_kind:"reservation",target_ref:<path ID文字列>,payload_sha256:<lowerhex64>}`。digestはUTF-8 JCSのSHA256。path・wrapper.target_ref・command対象IDは一致必須。key/actor/clinic/CAS版/警告証明を業務payloadへ混ぜず、認証scope又はPreconditionsへ分ける。

| 操作 | immutable hash projection | Preconditions |
|---|---|---|
| A | exact `{schema_version:1,intent:"reservation.consume_next_visit_intent",appointment_id,next_visit_intent_id,confirmed:true}`。両IDは正規文字列 | 今回予約のexpected_versionと元予定のexpected_versionを別名・対象ID対応で保持 |
| B | 480 exact `{schema_version:1,intent:"reservation.confirm_tentative",target_ref,booking_changes,trimming_selection}`。booking_changesはBで許された明示業務入力だけ（start/endと任意設定）、status/actor/key/CAS等は含めない。trimming_selectionは省略時null、指定時Content3key `{course_id,option_ids,expected_adoption_digest}` | expected_version、expected_booking_revision_id、warning_receipt等の現在状態/警告前提。expected_adoption_digestはContentに残す |

省略と明示null/空/falseを区別し、DB現在値をhash前に埋めない。Bの数値ID入力は既存wireから検証後canonical ID文字列へ投影し、時刻は上記UTC、option_idsは480の数値順。projectionのpresenceを保った同じimmutable commandからApply時の保存値mergeを行う。

Q constructorがcanonical bytesをdeep-copyしpure projectorでContent全key一致を検証、claim/fingerprint前とApply直前に再検証する。Applyは`req.VerifiedCommand`だけから復元し、別DTO/pointer captureをしない。Qはrequest-memory、durable command表を新設しない。意味内容/対象/確認digestを変更して同key再送すればoperation-key-conflict。未成功で同じ意味内容のCAS前提だけを再確認した場合に限り同key再試行、成功後は元receipt固定。

## 4. 結果wire・HTTP・復旧

成功succeeded/replayedはHTTP200、exact required `{status,operation_key,result_refs,committed_versions,recorded_at}`。refs/versionsはstring mapでarrayを直置きせず、ID/版は正規十進文字列。recorded_atは元成功時刻。後日のcurrent状態や新版へ差し替えない。

| branch | result_refs exact | committed_versions exact |
|---|---|---|
| A | `appointment_id,appointment_change_event_id,next_visit_intent_id,next_visit_intent_change_event_id` | `appointment,next_visit_intent` |
| B 今回新T3を生成しない（非trimmingを含む） | `appointment_id,appointment_change_event_id` | `appointment` |
| B 今回のoperationで新T3を生成した場合のみ | `appointment_id,appointment_change_event_id,booking_revision_id` | `appointment,booking`（480同義版） |

既存trimming選択非変更で新T3を書かないBは予約refs/versionだけを返す。旧T3を今回operation/eventの成果として返さず、旧T3のoperation/eventを書き換えない。既存採用状態は別の現在view付きreadから取得する。時刻変更だけで新T3を必須とせず、480の既存採用条件に従い、無条件に新規採用しない。非成功exactは`{status,operation_key}`のみ。内部EffectSummaryをwireへ追加しない。現在本文は別の現在view付きGETで取得する。

[458§6](emr-458-option-assignment-selected-contract-20261021.md)の既存12statusを固定再利用：

| status | HTTP |
|---|---|
| succeeded / replayed / not-succeeded | 200 |
| validation | 400 |
| version-conflict / dependency-changed / edit-right-lost / operation-key-conflict | 409 |
| operation-in-progress | 202 |
| scope-not-found | 404 |
| not-authorized | 403 |
| result-unavailable | 503 |

strict入力/未認証/route scope拒否は既存ErrorResponseの400/401/403/404。独自status追加なし。unknown/busy/transport障害は同key照会、新key/別予約による救済なし。not-succeededは成功でも無条件再実行許可でもない。validation/版/依存拒否後は入力を保持して現在認可・対象・差分/CASを再確認する。Retry-Afterがあれば尊重する。

C06 Lookupの結果store障害は、現在認可後の既存generic error（例500）と、保存済みsticky unknownのresult-unavailableを区別する。503へ機械的一括変換しない。失敗/監査/event/receipt不整合は全rollbackし、予定だけ先に消化済みにしない。

## 5. 元target anchor・lookup

480既選定の`GET /api/v1/reservations/operations/{operationKey}?intent=...`へA/Bの完全intent名をallowlist追加する。staffだけ、認証clinic/同staff/intent/keyで検索、no-store。元anchorからだけ解決して現在`reservations:view`とscopeを再検査し、Aは両対象、終端元予定は履歴readerで検査する。request/URL/成功refsによる元対象の捏造fallbackなし。未知keyだけscope-not-found/404、元anchor欠落/identity不一致/認可不能は480/C06のnot-authorized/403。別actor/key/医院の存在を漏らさない。

`reservation_operation_targets`は480で採択した設計であり、このmainの実装済み表ではない。同じownerで必要差分を受領し第二表を作らない。PK `(clinic_id,binding_id)`、FK `(binding_id,clinic_id)` → `operation_bindings(id,clinic_id)`、immutable・clinic RLS、staffのline_subject_hashはNULLを維持。既存appointment targetのappointment_idにAの今回予約/Bの元予定を保存し、`next_visit_intent_id bigint NULL CHECK (>0)`を追加する。Aでは正ID必須かつ今回と異なるID、他intentではNULL。binding intentとこの有限shapeをclaim factory/整合guardで照合する。

claim winnerのみ`ClaimMetadataWriter`でbindingとanchorを同じ短TXに保存し、失敗は両rollback。loserは保存済みanchorを再利用し、incomingで補修しない。業務appointments FK/row lockを短claimへ持ち込まない。preflightとApplyで実targetのclinic/pet/identityを再検査する。成功replay/unknown/lookupは現在view、未成功write再試行はさらに現在edit/CASを検査。共通Q/resolver未受領のまま稼働可能としない。

## 6. Aの保存・event・最小DDL設計（適用しない）

新status/consume台帳は作らない。元予定は既存`cancelled`へ終端し、source側eventだけにtyped `consumed_by_appointment_id`を持たせる。通常Cancel/Clear関数は呼ばず、理由を消化として表示する。通常取消通知/復元UI/outbox/履歴readはtyped関係で区別し、次回予定通知候補と配送直前の再照合で消化済みを除外する。cancelledという文字列だけで通常取消/消化を判定しない。`source_medical_record_id`は出典用途のまま保持する。

Aは元予定のstatus/version終端、今回予約の関係追加に伴うversion/event、両event・必要監査・成功receiptを同最外TXでcommitする。今回予約の業務field/status、記録本文/状態、請求/入金は不変。版増分は既存guardと一回だけ対応させ、アプリ/triggerで二重加算しない。

必要保存差分は以下を446/466へ渡す。新migration番号・model物理file名はここで予約しない。

- `appointment_change_events.consumed_by_appointment_id bigint NULL`。FK `(consumed_by_appointment_id,clinic_id)` → `appointments(id,clinic_id) ON DELETE RESTRICT`、自己参照（appointment_idと同値）禁止。source eventの`intent='reservation.next_visit_intent_consumed'` iff relation非NULL、他eventはNULL。partial UNIQUE `(clinic_id,appointment_id) WHERE consumed_by_appointment_id IS NOT NULL`で元予定の二重消化を禁止。
- Aのbinding intent=`reservation.consume_next_visit_intent`。source eventは上記`reservation.next_visit_intent_consumed`、target eventは`reservation.consume_next_visit_intent`、同operation/clinic/actorで元/先のID・確定versionとA receipt refs/versionsへ一意対応する。actorは認証staffのみ、body由来にしない。
- 新consume deferred guardはこのtyped source/target eventを追加する時点で、binding actor/clinic/intent、原anchor両ID、source終端・target業務値/status不変、両eventのbefore/afterと確定版、成功receiptの4refs/2versionsを検査する。孤立source/target・不一致・receipt欠落をcommitできない。既存のreceipt存在guardだけではこの対応検査を満たさない。
- 追加操作時の対応を検査し、後日の別予約更新で過去event版=現在appointments.versionを恒久要求しない。旧eventを補完/backfillせずimmutable/RLS/元FKを保持する。
- consumed sourceだけはrestore/setによる再活性化を禁止。別の明示setは現在active予定を再取得/CASし、なければ新tentative IDを作る。過去consume/eventを削除しない。通常の誤取消復元まで禁止しない。

## 7. lock・全caller閉包・B切替

取得順は既存467のgraph gate（返されたctx伝播） → C06 binding NOWAIT/元結果再読取 → pet intent lock → 予約ID昇順row lockと既存owner規約。Aは両予約を一度ずつlockし、同clinic/pet・元tentative/未消化・今回status・期待版・現在認可を再検査。Bのtrimming採用は480の予約保護 → T2 → 区分/course/rule/options順と同TX participantを維持する。claim TXにこのlockを移さない。

| caller/writer | 受領すべき閉包 |
|---|---|
| Create/CreateBatch/admin/walk-in/LIFF/trimming内部createOrPromote/カルテhook | no-consume。通常作成だけで元tentativeを検索吸収しない |
| 明示confirm | Bだけへ統合。同ID昇格、旧status入力/無key/無版callerは切替対象 |
| 既存今回予約への明示紐づけ | Aだけへ統合。別ID関係と履歴を原子的に保存 |
| set/clear/cancel/restore/patient reassignment/直接repo | active予定/consumed sourceと両対象の版・scopeを再検査。別患者付替え・通常restoreで関係を壊さない |
| 通知候補read/outbox配送直前read/履歴UI | typed消化関係で除外/消化表示。旧cancelledだけの復元/通知へ流さない |

Aが照合する予約status/pet/clinic等を変更するwriterだけ予約版gateに参加させる。金銭だけのwriterへ予約版加算やbilling CASを新要求しない。全domainをgraph gateへ拡張しない。

APIと共有型を先に非活性で収録し、新client/key/CAS/presence/480 previewへ切替。旧key/版をserverが捏造して新command成功にしない。切替scopeの無保護writeは既存契約更新拒否/ErrorResponseで停止し、旧自動吸収へfallthroughしない。B旧ReservationAppointment responseは新C06成功shapeへ移す対象であり、旧callerの同時切替/GET再取得を受領条件にする。API/DDL/model・Q/anchor・全必要writer・通知閉包が揃わないまま有効化しない。

## 8. 固定fixtureと提出単位

実装票の最小単位はA保存/guard/lookup/caller閉包と、B同ID確定/480participant/受付後続境界。共有API/保存/Q hunkはownerへ集約し、独自stub/二重型/第二receiptで先回りしない。各PRはbase/head/採用SHA、供給file/symbol、consumer受領、AC→handler/request→service→repository/model→receipt/lookup→FEを記録する。

fixture作成と実行を分け、以下は全てNOT_RUN：

- A：許可6statusと拒否status/null/削除、pending/confirmed非来院表示、同ID/別pet/別院拒否、未選択無変更、会計後も今回状態/診療/請求/入金不変。
- A：両対象古版、患者付替え/取消/消化競合、同source別key二重消化、event片側/actor/refs/版不一致rollback、消化source復元拒否と通常誤取消復元保持、別setの新ID/active競合、通知候補と配送直前除外。
- B：省略保持/doctor null解除/notes空とnull、status入力拒否、visit_type保存値保持、trimming4key/digest/preview end/manual override、旧無key/無版拒否、後続check-in明示/unknown停止、通常カルテ作成失敗・再利用とreceiptの分離。
- 共通：strict decode/非正規ID/UUID、同key同内容replay・異内容conflict・CASだけ再確認、sticky unknown/busy/失権/他actor lookup、失敗binding anchor・claim loser・metadata失敗rollback、監査/receipt失敗、後日更新後の元receipt読取。

本稿は技術選定の公開で、API YAML/DDL/model/生成物/製品コード/実DBは未変更。実装・全writer受入・実機試験の完了を意味しない。

## 9. 固定mainの静的差分根拠

| 現在path | 観測と受領差分 |
|---|---|
| [reservation_intent_handler.go L33](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/reservation/reservation_intent_handler.go#L33)・同L172、[service L142](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/reservation/reservation_intent_service.go#L142) | 現requestはstatusを受け、operation_key/CASなし。confirmed既定とpost-commit通常カルテ作成があり、Bのstrict入力・receipt/受付正本への切替が必要 |
| [reservation_intent_repository.go L1079](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/reservation/reservation_intent_repository.go#L1079) | pet intent lock→予約row、tentative状態述語と由来保持は既存成果。旧createOrPromote/自動吸収の全caller閉包を代替しない |
| [storage_reservations.go L24](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/internal/model/storage_reservations.go#L24)・[001 L8924](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/migrations/001_init.sql#L8924) | eventに版/intent/operation/actorは存在、consumed_by列/対応guardは未収録。anchor表/A routeもこのmainに存在せず、設計だけで実装済みにしない |
| [api.yaml L13781](https://github.com/MinoruSoga/AnimalEkarte/blob/a5b4a9553b3f003e90ea47660a109c5d526be59c/backend/docs/api.yaml#L13781)・同L17777 | 旧Bはstart/end必須・任意statusとReservationAppointment成功応答。新A/B exact/C06 wireの機械収録とcaller切替は後続 |

これらは静的読解で、build失敗や実DB競合を実行観測した記録ではない。
