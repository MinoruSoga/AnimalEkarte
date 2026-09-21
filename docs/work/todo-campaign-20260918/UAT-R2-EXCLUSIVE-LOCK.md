# UAT-R2-EXCLUSIVE-LOCK: 同時操作の判断票

状態: **競合ケース設計 READY／unit・mock 防御の再照合済み（2026-09-21）／治療・バイタル・処方・接種 Update の現防御・GAP・最小設計・テスト計画を追記（2026-09-22・製品 CAS は未 land）／DB apply 証明と 2 セッション実機は残件**。2026-09-19依頼者回答の「両方」により、目的はカルテの上書きと同じ会計の二重確定の防止に確定した。旧システムの画面占有を複製する要件ではない（[要件・状態の正本](../../../todo-issue.md#uat-r2-exclusive-lock)、[元報告](../stg-uat-clinic-feedback-q1-q4.md#uat-r2-exclusive-lock)）。画面占有ロックは未採用。

## 既存の防御と限界

| 合成データで確認する場面 | 現在の防御 | 防御の範囲 |
| --- | --- | --- |
| 2 セッションが同じカルテ所見・診断を読み、先に片方が保存した後、古い内容をもう片方が保存する | [保存処理](../../../frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts)は読み込んだ `existingClinicalPlanVersion` を保存要求へ渡す。[clinical_plan 更新](../../../backend/internal/medicalrecord/clinical_plan_repository.go)は `expectedVersion` を更新条件に含め、更新行がない場合に競合を返す。 | 古い保存の拒否。別端末が同じ画面を開くことは防がない。 |
| 2 セッションが同じ請求の明細を追加・変更する | [明細サービス](../../../backend/internal/billing/billing_item_service_create.go)は請求親を `LockAndFindByID` する。[会計リポジトリ](../../../backend/internal/billing/accounting_repository.go)の同メソッドは ambient transaction 内で請求行に `FOR UPDATE` 相当のロックを取る。[明細リポジトリ](../../../backend/internal/billing/billing_item_repository.go)の作成時参照検証も取引内で請求親行をロックする。 | この明細操作の書き込みを取引内で直列化する。すべての会計更新に同じロック経路があるという意味ではなく、医院の二重会計事故が解消するかは未検証。画面の占有もしない。 |
| 同じタブでカルテを編集し、保存せず画面を離れる | [ReadyPanels](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx)が dirty 状態を管理して [NavigationBlocker](../../../frontend/src/components/shared/NavigationBlocker/NavigationBlocker.tsx)へ渡す。[未保存変更フック](../../../frontend/src/hooks/use-unsaved-changes.ts)はタブ終了・ブラウザ遷移時の `beforeunload` 警告を登録する。 | 同じタブでの未保存離脱の警告。別端末・別タブの入室制御ではない。 |

いずれも**別端末からの画面表示を禁止する占有ロックではない**。既存の競合拒否、取引内の行ロック、離脱警告は目的と適用場面が異なる。旧システムの全面ロックをそのまま複製する判断はしない（[設計思想](../../product-philosophy.md)）。

## 実装担当が検証する事故と期待値

2つの独立したログインセッションA/B、同一医院の合成患者/カルテ/請求を用意する。各行を既存テストへ対応づけ、不足ケースをREDにしてから修正する。検証前に対象revision・専用環境/fixture・後処理を固定する。下表は未実行で、既存防御の存在だけでPASSにしない。

| ケース | 操作 | 受入条件 |
| --- | --- | --- |
| 古いカルテ保存 | A/Bが同版を読む→A保存→Bの古い版を保存。同時送信も確認 | 古い内容で上書きしない。競合表示、再読込/再入力手順があり、保存失敗を成功と見せない。Bの未保存入力を無警告で消さない |
| カルテ内の更新範囲 | 所見/診断だけでなく問診、治療、検査/接種等の更新経路を一覧化。同一行の編集/削除と確定が競合 | 行/資源ごとに守る境界を特定し、確定済みデータの上書き・消失を拒否。clinical_planのversionだけで全タブ保護済みとしない |
| 同一キーの会計再送 | 確定ボタン連打、応答喪失後に同一payload/keyを再送 | 同じ処理結果へ戻り、請求/支払/業務監査が増えない。異なるpayloadで同じkeyは拒否 |
| 異なるキーの二重会計 | A/Bが同じ既存会計を確定、または同一未請求治療/検査/接種を新規会計で同時確定。別々のkeyを使う | 同一業務対象の確定は1回のみ。敗者は既存結果か競合へ。別々の請求/支払として二重作成しない |
| 明細編集と確定 | Aが明細変更、Bが古い画面から確定 | 古い合計で確定せず、最新状態の再検証か競合拒否。部分的な明細/支払保存を残さない |
| 取消/通信断/再接続 | 取消・切断後に状態を再取得し再試行 | 不明な結果を新しいkeyで盲目的に再発行しない。取消後の状態を保持し、失敗は全rollback。永久ロックしない |
| 分離・独立操作 | 他医院で同一ID参照、同一医院の別カルテ/別会計を並行操作 | 他医院を読み書きしない。同一患者でも別の正当な会計まで一律禁止せず、衝突対象を定義する |

[会計complete](../../../backend/internal/billing/accounting_complete.go) は同一 `Idempotency-Key` のreplayと異なるpayloadの拒否を持ち、[transaction](../../../backend/internal/billing/accounting_complete_tx.go) が明細/支払等をまとめる。[frontend](../../../frontend/src/features/accounting/hooks/use-accounting-completion-action.ts) のkeyもmutation単位。この存在だけでは **別端末・別key** の二重会計の防止証拠にならない。

既存入口は [楽観ロックテスト](../../../backend/internal/medicalrecord/clinical_plan_repository_optimistic_lock_test.go)、[確定との競合](../../../backend/internal/medicalrecord/clinical_plan_finalize_concurrency_test.go)、[会計completeテスト](../../../backend/internal/billing/accounting_complete_test.go)。同一キーのmockテストと実DB並行トランザクションの検証を区別する。

## ケース→既存テスト対応

衝突キーは clinic＋対象行（カルテ/clinical_plan/明細/請求/未請求ソース）とする。下表は **既存テストの有無** であり、2セッション実機の実行結果ではない。実機列はすべて **UNKNOWN（未実行）**。医院事故の新規事実は記載しない。

| ケース | 既存テスト（ファイル / 関数） | 既存テストが証明すること | 未カバー（明示GAP） |
| --- | --- | --- | --- |
| 古いカルテ保存（所見・診断） | [clinical_plan_repository_optimistic_lock_test.go](../../../backend/internal/medicalrecord/clinical_plan_repository_optimistic_lock_test.go) `TestClinicalPlanRepository_Update_OptimisticLock`（stale `expectedVersion` は Conflict、一致時は version+1） | sequential な repo 更新で古い `expectedVersion` は上書きしない | 2 HTTP セッション同時送信は未カバー。`expectedVersion == nil` は照合スキップ（後方互換）。保存UIは [use-medical-record-save-action.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts) が version 未確定なら fail-closed |
| 確定 vs 所見編集/削除 | [clinical_plan_finalize_concurrency_test.go](../../../backend/internal/medicalrecord/clinical_plan_finalize_concurrency_test.go) `TestClinicalPlanFinalizeConcurrency`（update/delete が finalize より先に commit / finalize 先行なら child は Conflict） | clinical_plan 親行ロックと確定後の編集拒否 | 所見以外のタブ更新はこのテストの対象外 |
| 同一 Idempotency-Key の会計再送 | [accounting_complete_test.go](../../../backend/internal/billing/accounting_complete_test.go) `TestAccountingService_CompleteAccounting_IdempotentReplaySameDigest` / `…_IdempotentConflictDifferentDigest` / `…_AlreadyExistsResolvesToReplay`。FE 再利用は [complete-accounting.test.ts](../../../frontend/src/features/accounting/api/complete-accounting.test.ts) | 同一 key+同一 digest は create せず replay。同一 key+異なる digest は Conflict。Create UNIQUE→AlreadyExists を key 照会へ落とす | mock 経路が主。実DBで2接続が同一 key を同時 insert する証明は別途必要。`uq_billings_clinic_completion_request_id`（[001_init.sql](../../../backend/migrations/001_init.sql)）は同一 key 用 |
| 異なるキーの二重 complete | [accounting_complete_test.go](../../../backend/internal/billing/accounting_complete_test.go) `TestAccountingService_CompleteAccounting_DifferentKeySameMedicalRecordConflict`。Create の非 completion-key AlreadyExists → Conflict（[accounting_complete_tx.go](../../../backend/internal/billing/accounting_complete_tx.go) `createCompleteBillingHeader`、「このカルテには既に会計があります」）。schema: `idx_billings_medical_record_id_unique` | 別 Idempotency-Key でも同一 medical_record の第二 complete は Conflict（mock）。同一 key replay とは別経路 | 実DBで2接続が別 key を同時 insert する証明と 2 セッション実機は **残件**（migrate apply 後） |
| 未請求項目の二重請求 | 接種: [billing_item_vaccination_test.go](../../../backend/internal/billing/billing_item_vaccination_test.go) `TestBillingItemVaccinationProvenance_ConcurrentClaim` と `uq_billing_items_vaccination_lifetime`。治療: [004_billing_items_treatment_lifetime_unique.sql](../../../backend/migrations/004_billing_items_treatment_lifetime_unique.sql) + [billing_item_treatment_test.go](../../../backend/internal/billing/billing_item_treatment_test.go) `TestBillingItemService_CreateItem_DuplicateTreatmentProvenanceConflicts` / `TestBillingItemTreatmentProvenance_SecondClaimConflicts`（AlreadyExists→409）。検査: `uq_billing_items_exam_lifetime` + mock `TestBillingItemService_CreateItem_DuplicateExamProvenanceConflicts` | treatment/vaccination/exam の provenance UNIQUE 違反は Conflict | 治療 UNIQUE の **DB apply 証明**（`make migrate` はユーザー作業）と exam の実DB並行 claim は **残件**。complete 経由の二重作成は medical_record unique + provenance unique で抑止する設計 |
| カルテ API（clinical_plan 以外） | 親カルテ: [medical_record_repository_update_test.go](../../../backend/internal/medicalrecord/medical_record_repository_update_test.go) の `expectedVersion` 不一致 Conflict。検査改訂: [examination_revision_workflow_safety_test.go](../../../backend/internal/medicalrecord/examination_revision_workflow_safety_test.go) の stale version | 親レコードと検査 revision には CAS がある | **残GAP（詳細は下節）**: 治療 / バイタル / 処方 / 接種 Update に行 `version`/`expectedVersion` は無い（schema・model・request いずれも未装備）。clinical_plan version だけで全タブ保護済みとしない。本 unit は設計のみ・製品 CAS は land しない |
| 明細編集と確定 | 明細作成は親 `LockAndFindByID`（[billing_item_service_create.go](../../../backend/internal/billing/billing_item_service_create.go)）。実ロック: [accounting_repository_tx_atomicity_test.go](../../../backend/internal/billing/accounting_repository_tx_atomicity_test.go)。Complete は server totals 再計算 + payment 再検証: `TestAccountingService_CompleteAccounting_StaleScreenTotalRejected` / `…_PartialSplitRejected` / `…_OpenPeriodAtomicSuccess`（server total, not client）。確定後の遅延明細: `TestBillingItemService_CreateItem_LateItemOnCompletedConflicts`（+ testdb StatusGuard） | 古い合計での確定は InvalidInput。確定後の明細追加は Conflict | 実DB ambient 2tx と 2 ブラウザ実機の明細↔確定レースは **残件** |
| 取消/通信断後の再試行 | complete の N番目失敗 rollback: `TestAccountingService_CompleteAccounting_NthItemFailure_FullRollback` / `…_DBAtomicRollback`。FE は mutation 単位で key 再利用 | 同一 mutation の再送は増やさない設計 | 切断後に **新しい key** で再発行する2セッション手順のテストは **残GAP**。占有ロックの期限/解放は対象外（未採用） |
| 医院分離 | [accounting_repository_clinic_isolation_test.go](../../../backend/internal/billing/accounting_repository_clinic_isolation_test.go) が他院 FOR UPDATE を拒否 | clinic scope を破らない | 2セッション実機の分離確認は UNKNOWN |

## 最小サーバー側提案（占有ロックは採用しない）

[設計思想](../../product-philosophy.md) の順で、存在しない画面占有を最適化しない。確認ダイアログや全面ロックは安全性の成立根拠にしない。実装は **未カバー write に既存の expectedVersion / 行ロック / UNIQUE / Idempotency-Key を伸ばす** ことに限定する。製品コード・新規テストファイルはこの票の範囲外。

1. **expectedVersion（CAS）** — clinical_plan と親 medical_record に既にある。GAP の治療・バイタル・処方・接種など last-write-wins の更新へ、読み込み版を更新条件に含める。`expectedVersion == nil` のスキップ経路は新規 caller に広げない。
2. **行ロック（FOR UPDATE / ambient tx）** — 請求親と vaccination claim に既にある。complete と未請求ソース行の確定を同じ tx で固定し、古い画面の合計確定を commit 前に再評価する。ambient tx 不在は fail-closed。
3. **UNIQUE / 業務キー** — 同一 key は `uq_billings_clinic_completion_request_id`。異なる key の二重 complete は `idx_billings_medical_record_id_unique` の AlreadyExists→Conflict（landed）。治療 provenance は `uq_billing_items_treatment_lifetime`（migration 004、**未 apply のまま残件**）+ AlreadyExists→409（landed）。exam/vaccination lifetime UNIQUE は schema 済み。exam mock Conflict を追加済み。exam 実DB並行 claim は残件。
4. **Idempotency-Key** — 再送・連打専用。別端末が別 UUID を使うケースの防御には使わない。不明結果を新しい key で再発行しない契約を UI とサーバで揃える。

**採否（再掲）:** 画面占有ロックは未採用。開始/期限/解除/切断復旧を持つ occupancy は本票では提案しない。既存防御で足りる経路は再実装しない。



## 治療 / バイタル / 処方 / 接種 Update — 現防御・GAP・最小設計（2026-09-22）

対照パターン（既 land）: [clinical_plan Update](../../../backend/internal/medicalrecord/clinical_plan_repository.go) は `version` 列 + `expectedVersion` WHERE + RowsAffected==0 の Conflict 正規化。[親 medical_record Update](../../../backend/internal/medicalrecord/medical_record_repository.go) も同型。検査は revision workflow の version。会計は親 `FOR UPDATE` / UNIQUE / Idempotency-Key。**画面占有ロックは採用しない**（本節も提案しない）。製品 CAS・migration はこの票では land しない。

### 参照正本（コード）

| 経路 | Service Update | Repository Update | 子行ロック | 親 draft 直列化 |
| --- | --- | --- | --- | --- |
| 治療 | [treatment_service_tx.go](../../../backend/internal/medicalrecord/treatment_service_tx.go) `updateTreatmentInTx` | [treatment_repository.go](../../../backend/internal/medicalrecord/treatment_repository.go) `Update`（version 述語なし） | `LockByIDForUpdate`（ambient tx 必須・SEC-CS-F09） | `lockDraftMedicalRecord` |
| バイタル | [vital_service_update.go](../../../backend/internal/medicalrecord/vital_service_update.go) `updateVitalInTx` | [vital_repository.go](../../../backend/internal/medicalrecord/vital_repository.go) `Update`（version 述語なし） | **なし**（子行は Updates のみ） | `lockDraftParent` → 親 `LockByIDForUpdate` + finalized Conflict |
| 処方 | [prescription_service.go](../../../backend/internal/medicalrecord/prescription_service.go) `Update` | [prescription_repository.go](../../../backend/internal/medicalrecord/prescription_repository.go) `Update`（version 述語なし） | **なし** | `lockDraftMedicalRecord` |
| 接種 | [vaccination_service.go](../../../backend/internal/medicalrecord/vaccination_service.go) `Update` | [vaccination_repository.go](../../../backend/internal/medicalrecord/vaccination_repository.go) `Update`（version 述語なし） | `LockByIDForUpdate`（ambient tx 必須） | MR がある場合のみ `medicalRecords.LockByIDForUpdate`（**draft/finalized 判定は Update 経路に無し**） |

Schema / model: `clinical_plans.version` のみ。[treatments](../../../backend/migrations/001_init.sql) / `vital_records` / `prescriptions` / `vaccinations` と対応 model（`Treatment` / `VitalRecord` / `Prescription` / `Vaccination`）に **version 列は無い**。request/input にも `expectedVersion` は無い。

### 経路別 — 現防御 / GAP / 最小防御案

| 経路 | 現在の防御（証明できる範囲） | clinical_plan / 親 CAS・行ロック・UNIQUE との GAP | 提案する最小防御（製品 land は別 unit） |
| --- | --- | --- | --- |
| 治療 Update | (1) 親 draft `FOR UPDATE` で finalize と直列化。(2) 子 `FOR UPDATE` + discount 再検証（TOCTOU）。(3) clinic-scoped subquery Update。同一 tx 内の直列化と割引権限の失効は守る。 | **lost update**: A/B が同スナップショットを読み非割引フィールドを別々に保存すると後勝ち。`expectedVersion` / 子 `version` なし。billing provenance UNIQUE（`uq_billing_items_treatment_lifetime`）は **請求 claim** 用であり、治療行の内容上書き CAS の代替にならない。 | 子 `version INTEGER NOT NULL DEFAULT 1`（新 migration）+ Update WHERE `version = expectedVersion` + 成功時 +1。API/FE は読取版を必須送付（`nil` スキップを新規 caller に広げない）。既存子 `FOR UPDATE` と親 draft ロックは維持。UNIQUE 追加は不要（同一 PK 更新）。 |
| バイタル Update | (1) 親 draft ロック + finalized Conflict。(2) relation / staff 検証。(3) Update 後の同一 tx 再取得と audit fail-closed。 | 子行に `FOR UPDATE` も `version` も無い。親ロックは finalize レースを直列化するが、**同一 draft 親上の 2 セッション vital 内容競合は last-write-wins**。clinical_plan CAS と非対称。UNIQUE は適用場面なし。 | (優先) 子 `version` + `expectedVersion` CAS（clinical_plan 同型）。(補助) Update 前に vital 行 `LockByIDForUpdate`（ambient tx 必須・fail-closed）で同一 tx 内の他 writer と直列化。CAS 無しの行ロックだけでは「古い画面の上書き拒否」は成立しない点を明示。 |
| 処方 Update | (1) 親 draft `FOR UPDATE`。(2) medical_record_id 束縛。(3) Update 後同一 tx 再取得（MRC-01）。 | 子行ロック無し・version 無し。同一 draft 親上の処方フィールド競合は last-write-wins。clinical_plan CAS と非対称。 | 子 `version` + `expectedVersion` CAS。補助として prescription 行 `LockByIDForUpdate`（親ロック tx 内・FK/FOR KEY SHARE デッドロック注意は Create コメントと同型で DBOrTx 参加を維持）。UNIQUE 不要。 |
| 接種 Update | (1) 子 `FOR UPDATE` + 関係再検証（関係変更は Conflict）。(2) 並行 Update は行ロックで直列化（`TestVaccinationService_ConcurrentUpdatesSerialize`）。(3) MR 紐付け時は親行ロック。 | **内容の stale overwrite 拒否なし**（後勝ち）。**親 draft ガードなし**（治療/バイタル/処方と非対称。確定済みカルテ紐付け接種の編集可否は未契約）。`version`/`expectedVersion` なし。billing `uq_billing_items_vaccination_lifetime` は請求 claim 用。 | (1) 子 `version` + `expectedVersion` CAS。(2) `medical_record_id != nil` のとき `lockDraftMedicalRecord`（または status=draft 条件）を Update/Delete に揃え、確定後編集を Conflict。(3) 既存子 `FOR UPDATE` は維持。UNIQUE は Update CAS の代替にしない。 |

### 最小テスト計画（製品 CAS land 時・本 unit では実装しない）

命名は既存 clinical_plan / treatment TOCTOU / vaccination concurrency に揃える。mock と testdb を分離する。

| 経路 | 追加する最小テスト（案） | 証明すること | 既存で足りるもの（再実装しない） |
| --- | --- | --- | --- |
| 治療 | `TestTreatmentRepository_Update_OptimisticLock`（testdb）: stale `expectedVersion` → Conflict・書込なし、一致 → version+1。`TestTreatmentService_Update_StaleExpectedVersionConflict`（mock） | CAS 原子性とサービス Conflict 伝播 | `TestTreatmentService_Update_DiscountTOCTOU_*` / `TestTreatmentRepository_LockByIDForUpdate_RequiresAmbientTransaction`（割引・ロック） |
| バイタル | `TestVitalRepository_Update_OptimisticLock`；`TestVitalService_Update_StaleExpectedVersionConflict`；任意で `TestVitalRepository_LockByIDForUpdate_RequiresAmbientTransaction`（行ロックを入れる場合） | CAS と（採用時）ambient 必須 | `TestVitalService_Update_AuditFailureRollsBack` / refetch-in-tx（監査・再取得） |
| 処方 | `TestPrescriptionRepository_Update_OptimisticLock`；`TestPrescriptionService_Update_StaleExpectedVersionConflict` | CAS | `TestPrescriptionService_Update_FinalizedRejected`（親確定拒否） |
| 接種 | `TestVaccinationRepository_Update_OptimisticLock`；`TestVaccinationService_Update_StaleExpectedVersionConflict`；`TestVaccinationService_Update_FinalizedMedicalRecordRejected`（draft ガード追加時） | CAS + 親確定拒否 | `TestVaccinationService_ConcurrentUpdatesSerialize` / `TestVaccinationService_Update_RejectsConcurrentRelationChangeAfterValidation` / Lock ambient 必須 |

FE（別 unit）: 各タブの読取レスポンス `version` を保存リクエストへ必須で載せる。未確定なら fail-closed（clinical_plan の [use-medical-record-save-action.ts](../../../frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts) と同方針）。本票では FE 実装しない。

**採否再掲:** 占有ロック未採用。Idempotency-Key は再送専用で別端末 UUID の上書き防止には使わない。billing provenance UNIQUE は「同じ治療/接種を二重に請求しない」用であり、本節の **行内容 CAS** とは別レイヤ。

## 本 unit 残件（実行ゲート）

1. `make migrate` で `004_billing_items_treatment_lifetime_unique.sql` を適用し、testdb/実DB で treatment second-claim を再確認（エージェントは migrate しない）。
2. 候補 mount 済み専用環境で実トランザクション競合と 2 ブラウザセッション（古いカルテ保存・別 key complete・明細↔確定）を記録する。
3. 非 findings タブ（治療/バイタル/処方/接種）への `expectedVersion` 拡張は上節の最小設計を正本とし、**別 unit で製品 CAS を land**（本 unit は設計のみ。migration/製品コードは対象外）。

## 2 セッションの合成再現と採否ゲート

1. 上の対応表を衝突キー（clinic＋対象カルテ/明細/請求/未請求項目）の正本とする。GAP 行だけ RED→最小修正。防止目的を医院へ再質問しない。要件責任者の個人名と仕様変更の受入参照は変更前に実行票へ記録する。
2. scopedテストでRED→最小修正→GREENを実施し、候補mount済みの専用Docker/DBで実トランザクション競合と2ブラウザセッションの挙動を検証する。既存version/状態検証/一意性/行ロック/冪等性を優先し、確認ダイアログだけで安全性を成立させない。
3. ケース別にHTTP/競合表示、勝者の保存値、請求/支払/監査件数、rollback、再読込/再試行、医院分離、cleanupを記録する。外部通知がある場合はlocal stubを使い、実送信しない。すべての必須ケースの証拠が揃って完了。環境不足や未実行は残件にする。

**採否:** 全面的な画面占有ロックは未採用。既存防御で満たすケースは再実装しない。それでも占有が必要な不足ケースだけ、開始/期限/解除/切断復旧/権限/監査を別途裁定する。患者情報・秘密を本票に残さない。
