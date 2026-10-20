# 共通operation receiptと参加表（C06）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | design.shared.operation-participation／共通helper契約と操作family単位のowner参加表 |
| 決定状態 | [C06選定契約](integration-contracts.md#detail-design-proposals-20261018)とEMR-489レビュー訂正R2を技術接続へ具体化。新しい業務判断を追加しない |
| 実装状態 | 共通helper・DTO・結果区分・取得順を公開（`backend/internal/operation`）。consumer（domain adapter・route・OpenAPI）は未接続 |
| 実機検証状態 | 未実施（DBを使うstore試験・API・UIは未実行。下記「未実施・残件」） |
| 文書確認日・基準 | 2026-10-20／実装base main `5c5219965ad44258b8ccd918ffe52179a6139ade`（PR #729 merge後）。業務内容の追補は#738 merge `5c812ee0e7b1f603fc0a58e70b9bf5ec08a88246`の引渡しを参照（本PRへのmergeはしない） |
| 対象 | 共通binding/result/attemptのhelper契約と、参加する操作familyの必要subset。各domainの実writer/reader/adapterは各consumer票が所有 |

本書は契約公開であり、結合実装・受入合格を表さない。全owner接続の完了を他票の着手条件にしない。

## 1. 共通契約と単一owner

| 実ファイル | 単一owner | 内容 |
|:---|:---|:---|
| [contract.go](../../../backend/internal/operation/contract.go) | EMR-489 | `Scope`／`Content`／`Preconditions`／`Request`／`Binding`／`Result`／`Status`／`AttemptOutcome`／`Rejection`／`IndeterminateEffect`／`Busy`／`Participant`／`LookupAuthorizer`／`Outcome` |
| [fingerprint.go](../../../backend/internal/operation/fingerprint.go) | EMR-489 | `NewContent`、HMAC-SHA-256鍵リング（`NewFingerprinter`、鍵versionで旧鍵照合） |
| [runner.go](../../../backend/internal/operation/runner.go) | EMR-489 | `Runner.Execute`／`Runner.Lookup`、取得順・最外transaction境界 |
| [store.go](../../../backend/internal/operation/store.go) | EMR-489 | `operation_bindings`／`operation_results`／`operation_attempt_outcomes`のGORM adapter。全queryに`clinic_id`条件 |
| [dto.go](../../../backend/internal/operation/dto.go) | EMR-489 | `ParseOperationKey`、`ResultResponse`（成功時だけresult_refs／committed_versions） |
| 3表のDDL・model | EMR-446／466（DDL）、EMR-445（model） | 本票で変更しない。`operation_results.state`は除去済みで、成功は行の存在 |

保存先DDLは[001_init.sql](../../../backend/migrations/001_init.sql)、詳細ERDは[operation_bindings](../../architecture/erd-detail/operation_bindings.md)・[operation_results](../../architecture/erd-detail/operation_results.md)・[operation_attempt_outcomes](../../architecture/erd-detail/operation_attempt_outcomes.md)。

### 1.1 束縛とfingerprint（R2）

**適用境界（今回の訂正が以下の旧外部効果例に優先）**: Participant は `Effects() == EffectsDatabaseOnly` を明示する。未指定・外部送信等は claim/Gate/Apply の前に拒否する。Apply は渡された同一DB transactionで完結する書込・必須監査・receiptだけを行う。outbox行INSERTは含められるが、実際の外部送信・ファイル書込・他DB等は対象外。送信は別ownerのoutbox配送/結果照合契約が必要で、本helperのunknown保存を外部送信の再実行fenceとして利用開始してはならない。以下に残るunknown fixtureはDBエラーの保守的停止であり、外部効果の安全性を実証しない。

失敗後のbusy/sticky unknown/結果読取エラー/unknown読取エラーもrollback後に現在認可を通す。unknown行記録後のcommitエラーには元のIndeterminateEffectを保持し、generic retryable_failureへ落とさない。永続停止を保存できない状態は引き続き未成立であり、外部consumer接続の許可にはならない。


- staff/customer の `actor_ref` は正の uint64 を先頭ゼロ・空白・符号なしの10進文字列で表す。Execute/Lookup は非正規入力を保存検索前に拒否し、黙って正規化しない。system は非空の命名 principal を使い数値化しない。
- 束縛scopeは認証済みの`clinic_id`＋`owner_kind`＋`actor_kind`/`actor_ref`＋`intent`＋`operation_key`（DDLの一意キーと同じ）。clientが指定したclinic・actorで行を選ばない。
- `Content`（業務内容）と`Preconditions`（CAS・fence等の実行前提）は別の型である。集合値（例: 対象行ID集合）はadapterが正規順（昇順）で1つの値にしてから`Content`へ入れる。順序違いの同じ内容を別fingerprintにしない。
- ownerがDB内効果の有無を判定できない場合は`Apply`が`*IndeterminateEffect`を返す。helperは`unknown`を、binding行lockを保持したままの同一transaction内でsavepointにより記録する。savepointは`Apply`の**前に**張る（書込後のsavepointではowner writeをundoできない）。lock解放前に停止状態がcommitされるため、rollback〜記録間に並行実行が`Apply`を再実行する窓はない。fingerprintの入力は`owner_kind`・`intent`・`Content`だけで、`Preconditions`は型の上でfingerprintへ入らない。`expected_versions`のように両者を一括する名前は使わない。
- `Apply`はfingerprint済みの`Request`そのものを受け取り、owner writeは`req.Content.Fields()`（読取専用の複製）から導出する。fingerprintされた内容と書き込まれた内容が乖離しreceiptを別内容で保存する経路を型とAPIで閉じる。
- gateや`Apply`が保持できないlockを取れなかった場合は`*Busy`を返す。helperは`operation-in-progress`として応答し、attemptを記録しない（binding行のlock失敗と同一契約）。
- fingerprintはサーバー鍵によるHMAC-SHA-256。鍵versionを保存し、旧鍵versionで再計算して照合する。鍵はconstructorへ注入し、環境変数名・設定配線は未選定（§5）。

### 1.1a 元対象の認可resolver（接続条件）

`LookupAuthorizer` はまず `ResolveLookupTarget(ctx, binding, storedResult)` で保存済み元対象を解決し、その `LookupTarget` だけを `AuthorizeLookup` へ渡す。helperはbinding ID/clinic一致と対象kind/refの存在を検査し、解決不能なら `not-authorized` で停止する。今回requestの別resourceによる認可代用は禁止（key-conflictでも同じ）。

consumerは成功receiptがないbindingも含め、binding ID＋医院から元対象を解決する永続owner metadata/readerを提供する必要がある。成功時はhelperがそのbindingから読んだreceiptのresult_refsを利用できるが、元対象の正本として十分な参照であり医院/identityを検査できる場合に限る。削除/欠損/曖昧で元対象が不明なら失敗扱いとし、今回のContentやURL IDへfallbackしない。現在schemaのfingerprintは対象を復元するデータではない。

実consumer用resolverと未成功bindingの元対象metadataは未接続。単純なpre-savepoint hookでは最外TXのrollbackを越えてmetadataを残せず、claimとの原子性・同key競合・crash回復の保存契約が必要である。このhook設計はconsumer利用開始条件であり未接続libraryのmerge前提へ戻さない。大規模storage変更をここへ追加せず各consumer/446の必要subsetとして残し、resolverがないconsumerの結果開示は利用開始BLOCKED。合成fixtureはbinding→元対象Aと、現在要求Bの権限を分けて検証する。実漏洩を確認した報告ではない。

### 1.2 取得順・最外transaction境界

| 段 | 処理 | 失敗時 |
|:---|:---|:---|
| 0 | ambient transaction内からの呼出しを拒否（最外であること） | 実行しない |
| 1 claim（短い独立tx） | bindingをINSERT（同scope・keyが既存なら何もしない）→保存行を読む。保存fingerprintとContentが不一致なら、現在認可を再検査してから終了（conflict応答自体が fingerprinted content の等値オラクルになるため、認可なきactorには`not-authorized`で隠す） | 認可あれば`operation-key-conflict`・認可なければ`not-authorized`。attemptは記録しない |
| 2 最外業務tx | ownerが渡すgate（例: 予約graph参加者だけ）→binding行`FOR UPDATE NOWAIT`→成功resultがあれば現在認可（`AuthorizeLookup`）を再検査して再実行せず返す→`unknown`が1件でもあれば現在認可を再検査して停止（認可なきactorには`not-authorized`で隠す。後の失敗行で解除しない）→`Apply`前にsavepoint作成→`Participant.Apply`（現在認可・scope・Preconditionsの再照合、`req.Content.Fields()`由来のowner write、必須監査）。`*IndeterminateEffect`時はsavepointへrollbackしてowner writeをundoし、同一tx内で`unknown`行を記録してcommit（lock解放前に停止状態を永続化）→`operation_results` INSERT | gateもlockも待たない。lock取得不可はbinding行・owner gateの双方で`operation-in-progress`（`*Busy`含む）とし、attemptを記録しない |
| 3 失敗後 | callback成功後のcommitエラーは`settledState`で再確認。receiptが見えなければbindingをNOWAITでlockしresult/unknownを再読取。保持中は`operation-in-progress`、再読取不能は`result-unavailable`、成功receiptありは現在認可を再検査して`succeeded`。lock下で未成功を確定できた場合のみfailure分類へ。callbackエラーによるrollbackは従来の失敗分類 | 読取失敗・`unknown`記録済みは`result-unavailable`（失敗行を重ねない）。それ以外は別txでattemptを記録（`*IndeterminateEffect`は通常tx内で記録済み。tx内記録に失敗した場合のみtx外でbest-effort記録し以後停止、`*Rejection`→`rejected`、`*Busy`・binding lock保持中→`operation-in-progress`で記録なし、その他→`retryable_failure`）。tx内記録とtx外fallbackの双方が失敗した場合は応答だけ停止し、永続unknown行は残らない。DB-onlyのcallbackエラーはowner writeをrollbackし、再Executeはbinding lock＋receipt再読取を通る。外部副作用を再実行してよいという意味ではなく、外部送信は本helperの非対象。commit結果未決着は上記lock下の再確認で区別する。attempt_keyは毎回サーバー生成、detailsは`{"status": ...}`だけ |

照会（`Runner.Lookup`）は、resultが見えなければbinding行をNOWAITでlockしてからresultと`unknown`を再読取する。lock取得前に他の試行がcommitした成功を`not-succeeded`と返さない。結果を返す全経路（照会・replay・失敗後の成功確認）で`AuthorizeLookup`を通す。`Participant`はこのため`LookupAuthorizer`を含む。

conflictの元対象認可には保存済み成功receiptを解決根拠として渡せるが、conflict応答のResult本文は常に空とする。成功前はbindingに対応するowner metadataで解決し、解決不能なら非開示。失敗後分類は成功receipt→既存unknown（読取不能も保守的停止）→Busyの順で、Gate競合によってunknown停止を隠さない。callbackで既存receiptをreplayした場合はcommit応答喪失後も再認可して`replayed`を維持し、新規Applyの成功は`succeeded`とする。Lookupでbinding発見後に結果を再読取できない場合は、現在認可の下で`result-unavailable`へ停止する。初回`FindBinding`自体の失敗はbindingの存在を推定しない。

共通helperはgateを自分で取らない。EMR-467の予約graph gateを全domainへ拡張しない。owner固有のlock順は`Apply`内で各ownerの既存規約に従う。

### 1.3 結果区分

| Status | 意味 | client／serverの扱い |
|:---|:---|:---|
| `succeeded` | owner writeと成功receiptがcommit済み | 初回結果を表示。照会でも同じ結果 |
| `replayed` | 同key同内容の成功済み | 保存結果を返し、owner writeを再実行しない |
| `not-succeeded` | bindingあり・成功なし（拒否・rollback後） | 人が最新前提を再確認し、同key同内容で再試行可 |
| `operation-key-conflict` | 同keyで別内容 | 再実行しない。内容を変えるなら旧結果の確定を確認した別key |
| `operation-in-progress` | 他の試行がbindingを保持中 | 同keyで再照会。新keyで作り直さない |
| `result-unavailable` | 成否不明（`unknown`記録、または再読取不能） | 自動補償・再create・外部再送を止める |
| `scope-not-found` | `FindBinding`で同scope・keyのbindingがない場合のみ。ownerの対象欠落には既存`validation`等を使い、このstatusを返さない | 他院・他actorのkeyの存在を開示しない |
| `not-authorized` | 現在の認可で結果を開示できない | 結果本文（result_refs等）を返さない |
| `validation`／`version-conflict`／`dependency-changed`／`edit-right-lost` | ownerの確定拒否（`*Rejection`） | writeなし。再取得・差分確認後に同key同内容で再試行可 |

HTTP status（403/404/409等）への対応は各ownerの既存非開示方針に従い、本helperでは決めない。

## 2. 参加表（操作family単位）

| 項目 | `billing.price_acknowledgement`（先行例） | EMR-491／493／502 |
|:---|:---|:---|
| domain所有票 | EMR-483（billing） | 各票 |
| 業務intent内容（Content、fingerprint対象） | `billing_id`、対象`billing_item_id`集合、了承したbefore計算版／after計算版、before／after金額 | 未判定（subset照合のみ・接続未着手） |
| 実行前提（Preconditions、fingerprint対象外） | `billing_version`、`line_version`（CAS） | 未判定 |
| 保存先 | `billing_price_acknowledgements`、`billing_change_events`、`operation_bindings`／`operation_results`／`operation_attempt_outcomes` | 未判定 |
| 監査 | 変更者・時刻・前後値の履歴（[会計の金額変更了承](../domains/accounting/technical-spec.md#review-version-sync-proposal)）。同tx | 未判定 |
| 成功結果 | `result_refs`: billing_id・了承行ID、`committed_versions`: 更新後billing_version／line_version | 未判定 |
| 拒否・unknown時の照会先 | `Runner.Lookup`（同scope・key、現在認可を再検査） | 共通 |
| 固定fixture | §3の6ケース | 未作成 |
| C06との優先関係・adapter変換 | C06共通規則に従う。金額のみ変更は警告了承で医師確認receiptを失効させない（AC-accounting-reviewed-price-only） | 未判定 |

483の判定履歴と保存契約の技術残件（確定した業務内容と未実装のDB subsetを分ける）:

- 判定履歴: 初稿では`billing_price_acknowledgements.calculation_version`の意味をDDL単独で判定せず保留した。#738の[2026-10-20引渡し](https://github.com/MinoruSoga/AnimalEkarte/blob/5c812ee0e7b1f603fc0a58e70b9bf5ec08a88246/docs/design/domains/accounting/technical-spec.md#price-acknowledgement-handoff-20261020)で、了承したbefore/after計算revision・対象行・提示金額／条件はContent、現在会計／行のCAS・fenceはPreconditionsと確定した。旧列名をこの合意だけで新しい保存契約へ読み替えず、必要なschema subsetは446／466へ渡す。
- 判定履歴: 初稿では`request_hash`と共通fingerprintの二重管理を483の判断待ちとした。最新の[446／466限定DB条件](https://github.com/MinoruSoga/AnimalEkarte/blob/5c812ee0e7b1f603fc0a58e70b9bf5ec08a88246/docs/work/db-contracts/unified-db-20261019.md#price-ack-db-handoff-20261020)では、二重再送正本の整理に必要な最小参照／制約、および理由不要price-only了承のevent契約が未実装。専用`price_acknowledgement` intentは第一候補であり採用DDLではない。保存契約は446、限定DB統合・受入は466、domain adapter／writer／readerは483、共通helper／DTO／参加表は489が所有する。
- 旧`billings.completion_request_id`／hashは除去済みで、復活させない。

## 3. 固定fixture（R2の6ケース）

テストは[fixtures_test.go](../../../backend/internal/operation/fixtures_test.go)の`TestC06Fixtures_R2`。DBを使わない合成fixture（in-memory storeと483を模したParticipant）。

| ケース（t.Run名） | Given／When | Then |
|:---|:---|:---|
| `same_content_cas_only_updated` | 古いCASで実行し拒否、同key・同Contentで前提だけ最新にして再実行 | 1回目`version-conflict`・成功receiptなし・attempt `rejected`。2回目`succeeded`。fingerprintとbinding行は同一 |
| `ack_before_after_version_changed` | 成功後、同keyで了承after計算版だけ変更 | `operation-key-conflict`、Apply追加なし。別keyなら新bindingで実行可 |
| `same_key_different_content` | 成功後、同keyで金額を変更 | `operation-key-conflict`、Apply追加なし、保存結果は不変 |
| `succeeded_result_resent` | 成功後、同key同内容を再送 | `replayed`、保存結果と一致、Apply追加なし |
| `unknown_result_lookup` | 別keyの実行でParticipantが`IndeterminateEffect`を返す | 初回は`result-unavailable`で`unknown`記録・成功receiptなし。同keyの再実行も照会も`result-unavailable`でApply呼ばれない。新key＋古いCASはownerのCASで`version-conflict`となり二重了承しない |
| `current_permission_revoked` | 成功後に認可を失効 | 照会は`not-authorized`、結果本文なし |

`current_permission_revoked`は照会に加えて`Execute`の同key再送でも`not-authorized`・Apply追加なしを確認する。

補助テスト: unknown後のgate拒否→回復でも停止が続く（`TestExecute_UnknownStaysStickyAfterGateFailure`）、初回読取とlockの間にcommitされた成功を照会が`succeeded`と返す（`TestLookup_SeesSuccessCommittedBeforeLock`）、gate失敗時に見つかった成功receiptも現在認可で隠す（`TestExecute_GateFailureWithCommittedReceiptRechecksGrant`）、gate競合は`*Busy`で`operation-in-progress`・attempt記録なし（`TestExecute_GateBusyRecordsNothingAndReportsInProgress`）、unknown停止も現在認可で開示（`TestExecute_UnknownDisclosureRequiresCurrentGrant`）、NOWAIT abort後はrollbackして`operation-in-progress`（`TestLookup_BusyBindingRollsBackAbortedTransaction`）、`unknown`はlock保持下の同一txでcommitされ後続実行がApplyに届かない（`TestExecute_UnknownCommitsUnderLockBeforeRelease`・savepointはApply前に作成されることを順序でpin）、unknown永続化の継続失敗時は応答のみ停止し、DB-only callback rollback後の再試行と外部副作用を区別（`TestExecute_UnknownPersistenceFailureKeepsResponseStopped`）、`Apply`はfingerprint済み`Request`を受け`Content.Fields()`経由で読む（`TestExecute_AppliesTheFingerprintedContent`）、key-conflict開示前にも現在認可を再検査（`TestExecute_KeyConflictHidesFromRevokedGrant`）、ambient tx内からの実行拒否、`retryable_failure`記録とrollback、lock保持中の`operation-in-progress`、binding不在の`scope-not-found`。

## 4. 非対象・維持する別契約

- EMR-467の予約graph gateを全domainへ拡張しない。`Participant.Gate`はownerが必要な場合だけ渡す。
- EMR-473のS1再診回答claimed/succeeded receiptは別契約のまま。C06へ統合しない。
- 旧列（`operation_results.state`、旧completion UUID／hash）を復活させない。migrationを追加しない。

## 5. libraryのmerge条件と利用開始条件

本PRは実consumer・route未接続のDB-only libraryである。libraryのmerge判断ではhelper契約の整合、限定test/vet、既知未検証の明記を確認する。全consumer接続・実RLS受入をlibrary mergeの前提へ混同しない。merge自体は別途判断し、本追補は承認や受入合格を意味しない。

利用開始には各consumerの現在認可・元対象metadata/resolver・CAS/fence・必須監査と成功receiptの同一TX接続、必要schema/fixture subset、実001/RLS/deferred/並行実行・結果照合の検証が必要である。下表は未実施のまま保持する。DB-only callbackがrollbackした場合のunknown記録失敗を、外部送信の二重効果の窓としてlibrary merge blockerにしない。外部送信・filesystem等は引き続き非対応で、別ownerのtransactional outbox／配信・照合契約が必要である。

### 5.1 未実施・残件

| 項目 | 状態 | 次の担当 |
|:---|:---|:---|
| 実DBのstore試験（claim一意性、ambient tx、NOWAIT、rollback、clinic scope、sticky unknown） | 未実施・未compile。`store_realdb_test.go`（build tag `operationrealdb`）は、main上で`internal/testdb`がPR #729のmodelに対しcompileできない（既存）ため、型検査も一度も通していない | worktree stack起動後に実行 |
| deferred trigger `operation_success_required`・RLSとの結合 | 未実施（test DBはAutoMigrate double） | 最初のconsumer接続時 |
| HMAC鍵のconfig配線・鍵rotation手順 | 未選定 | 最初のconsumer接続時 |
| 結果照会route・OpenAPI・codegen | 未着手。router／api.yamlは共有ファイルのため各ownerで追加 | EMR-483ほか |
| cancel済みcontextからの即時結果照合 | 後続の可用性改善。現在はfail-closedで停止し、新しい有効contextの同key Lookup/Executeで再照合する。キャンセル後の即時照合成功は保証しない | helper/consumerのcontext契約検討 |
| 491／493／502の参加行 | subset照合のみ・未判定 | 各票 |
