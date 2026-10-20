# 共通operation receiptと参加表（C06）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | design.shared.operation-participation／共通helper契約と操作family単位のowner参加表 |
| 決定状態 | [C06選定契約](integration-contracts.md#detail-design-proposals-20261018)とEMR-489レビュー訂正R2を技術接続へ具体化。新しい業務判断を追加しない |
| 実装状態 | 共通helper・DTO・結果区分・取得順を公開（`backend/internal/operation`）。consumer（domain adapter・route・OpenAPI）は未接続 |
| 実機検証状態 | 未実施（DBを使うstore試験・API・UIは未実行。下記「未実施・残件」） |
| 文書確認日・基準 | 2026-10-20／main `5c5219965ad44258b8ccd918ffe52179a6139ade`（PR #729 merge後） |
| 対象 | 共通binding/result/attemptのhelper契約と、参加する操作familyの必要subset。各domainの実writer/reader/adapterは各consumer票が所有 |

本書は契約公開であり、結合実装・受入合格を表さない。全owner接続の完了を他票の着手条件にしない。

## 1. 共通契約と単一owner

| 実ファイル | 単一owner | 内容 |
|:---|:---|:---|
| [contract.go](../../../backend/internal/operation/contract.go) | EMR-489 | `Scope`／`Content`／`Preconditions`／`Request`／`Binding`／`Result`／`Status`／`AttemptOutcome`／`Rejection`／`Participant`／`LookupAuthorizer`／`Outcome` |
| [fingerprint.go](../../../backend/internal/operation/fingerprint.go) | EMR-489 | `NewContent`、HMAC-SHA-256鍵リング（`NewFingerprinter`、鍵versionで旧鍵照合） |
| [runner.go](../../../backend/internal/operation/runner.go) | EMR-489 | `Runner.Execute`／`Runner.Lookup`、取得順・最外transaction境界 |
| [store.go](../../../backend/internal/operation/store.go) | EMR-489 | `operation_bindings`／`operation_results`／`operation_attempt_outcomes`のGORM adapter。全queryに`clinic_id`条件 |
| [dto.go](../../../backend/internal/operation/dto.go) | EMR-489 | `ParseOperationKey`、`ResultResponse`（成功時だけresult_refs／committed_versions） |
| 3表のDDL・model | EMR-446／466（DDL）、EMR-445（model） | 本票で変更しない。`operation_results.state`は除去済みで、成功は行の存在 |

保存先DDLは[001_init.sql](../../../backend/migrations/001_init.sql)、詳細ERDは[operation_bindings](../../architecture/erd-detail/operation_bindings.md)・[operation_results](../../architecture/erd-detail/operation_results.md)・[operation_attempt_outcomes](../../architecture/erd-detail/operation_attempt_outcomes.md)。

### 1.1 束縛とfingerprint（R2）

- 束縛scopeは認証済みの`clinic_id`＋`owner_kind`＋`actor_kind`/`actor_ref`＋`intent`＋`operation_key`（DDLの一意キーと同じ）。clientが指定したclinic・actorで行を選ばない。
- `Content`（業務内容）と`Preconditions`（CAS・fence等の実行前提）は別の型である。fingerprintの入力は`owner_kind`・`intent`・`Content`だけで、`Preconditions`は型の上でfingerprintへ入らない。`expected_versions`のように両者を一括する名前は使わない。
- fingerprintはサーバー鍵によるHMAC-SHA-256。鍵versionを保存し、旧鍵versionで再計算して照合する。鍵はconstructorへ注入し、環境変数名・設定配線は未選定（§5）。

### 1.2 取得順・最外transaction境界

| 段 | 処理 | 失敗時 |
|:---|:---|:---|
| 0 | ambient transaction内からの呼出しを拒否（最外であること） | 実行しない |
| 1 claim（短い独立tx） | bindingをINSERT（同scope・keyが既存なら何もしない）→保存行を読む。保存fingerprintとContentが不一致なら終了 | `operation-key-conflict`。attemptは記録しない |
| 2 最外業務tx | ownerが渡すgate（例: 予約graph参加者だけ）→binding行`FOR UPDATE NOWAIT`→成功resultがあれば再実行せず返す→最新attemptが`unknown`なら停止→`Participant.Apply`（現在認可・scope・Preconditionsの再照合、owner write、必須監査）→`operation_results` INSERT | gateもlockも待たない。lock取得不可は`operation-in-progress` |
| 3 失敗後 | 新しい読取でresultの有無を再確認。あれば成功（commit済みを失敗へ反転しない） | 読取失敗は`result-unavailable`。resultなしは別txでattemptを記録（`*Rejection`→`rejected`、その他→`retryable_failure`）。attempt_keyは毎回サーバー生成、detailsは`{"status": ...}`だけ |

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
| `scope-not-found` | 同scope・keyのbindingなし | 他院・他actorのkeyの存在を開示しない |
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

483の未判定・技術残件（EMR-483が確定する。本票で推測しない）:

- `billing_price_acknowledgements.calculation_version`がContent（了承したafter計算版）かPreconditions（CAS）か、DDL単独では判定できない。BLOCKED。
- `billing_price_acknowledgements.request_hash`と`operation_bindings.fingerprint`は同じ要求を二箇所で識別する二重管理の候補。同値コピーとは定義しない。解消方針は483が決める。
- 旧`billings.completion_request_id`／hashは除去済みで、復活させない。

## 3. 固定fixture（R2の6ケース）

テストは[fixtures_test.go](../../../backend/internal/operation/fixtures_test.go)の`TestC06Fixtures_R2`。DBを使わない合成fixture（in-memory storeと483を模したParticipant）。

| ケース（t.Run名） | Given／When | Then |
|:---|:---|:---|
| `same_content_cas_only_updated` | 古いCASで実行し拒否、同key・同Contentで前提だけ最新にして再実行 | 1回目`version-conflict`・成功receiptなし・attempt `rejected`。2回目`succeeded`。fingerprintとbinding行は同一 |
| `ack_before_after_version_changed` | 成功後、同keyで了承after計算版だけ変更 | `operation-key-conflict`、Apply追加なし。別keyなら新bindingで実行可 |
| `same_key_different_content` | 成功後、同keyで金額を変更 | `operation-key-conflict`、Apply追加なし、保存結果は不変 |
| `succeeded_result_resent` | 成功後、同key同内容を再送 | `replayed`、保存結果と一致、Apply追加なし |
| `unknown_result_lookup` | bindingに`unknown`を記録 | 実行も照会も`result-unavailable`、Apply呼ばれない。新key＋古いCASはownerのCASで`version-conflict`となり二重了承しない |
| `current_permission_revoked` | 成功後に認可を失効 | 照会は`not-authorized`、結果本文なし |

補助テスト: ambient tx内からの実行拒否、`retryable_failure`記録とrollback、lock保持中の`operation-in-progress`、binding不在の`scope-not-found`。

## 4. 非対象・維持する別契約

- EMR-467の予約graph gateを全domainへ拡張しない。`Participant.Gate`はownerが必要な場合だけ渡す。
- EMR-473のS1再診回答claimed/succeeded receiptは別契約のまま。C06へ統合しない。
- 旧列（`operation_results.state`、旧completion UUID／hash）を復活させない。migrationを追加しない。

## 5. 未実施・残件

| 項目 | 状態 | 次の担当 |
|:---|:---|:---|
| 実DBのstore試験（claim一意性、ambient tx、NOWAIT、rollback、clinic scope） | 未実施。試験は`store_realdb_test.go`（build tag `operationrealdb`）。`internal/testdb`がPR #729のmodelに対しcompileできない状態はmain上で既存 | worktree stack起動後に実行 |
| deferred trigger `operation_success_required`・RLSとの結合 | 未実施（test DBはAutoMigrate double） | 最初のconsumer接続時 |
| HMAC鍵のconfig配線・鍵rotation手順 | 未選定 | 最初のconsumer接続時 |
| 結果照会route・OpenAPI・codegen | 未着手。router／api.yamlは共有ファイルのため各ownerで追加 | EMR-483ほか |
| 491／493／502の参加行 | subset照合のみ・未判定 | 各票 |
