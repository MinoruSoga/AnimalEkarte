# UAT-R2-EXCLUSIVE-LOCK: 同時操作の判断票

状態: **競合ケース設計・不足箇所の調査 READY／実行未実施**。2026-09-19依頼者回答の「両方」により、目的はカルテの上書きと同じ会計の二重確定の防止に確定した。旧システムの画面占有を複製する要件ではない（[要件・状態の正本](../../../todo-issue.md#uat-r2-exclusive-lock)、[元報告](../stg-uat-clinic-feedback-q1-q4.md#uat-r2-exclusive-lock)）。

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
| 異なるキーの二重 complete | **GAP** | なし。`Complete` は key 単位の replay のみ（[accounting_complete.go](../../../backend/internal/billing/accounting_complete.go) / [accounting_complete_tx.go](../../../backend/internal/billing/accounting_complete_tx.go) は `completion_request_id` 衝突だけ replay） | 別 UUID の2確定は冪等キーでは止まない。`idx_billings_medical_record_id_unique` は schema にあるが、この index 名を参照するテストは見つからない。2キー同時 complete の Conflict 契約は未固定 |
| 未請求項目の二重請求 | 接種: [billing_item_vaccination_test.go](../../../backend/internal/billing/billing_item_vaccination_test.go) `TestBillingItemVaccinationProvenance_ConcurrentClaim` と `uq_billing_items_vaccination_lifetime`。検査: `uq_billing_items_exam_lifetime`（schema） | 同一 vaccination_id の並行 claim を拒否する経路がある | **治療 `treatment_id` は UNIQUE ではなく通常 index**（`idx_billing_items_treatment_id`）。exam の lifetime unique に対応する並行テスト名は見つからない。未請求治療/検査の complete 経由二重作成は **GAP** |
| カルテ API（clinical_plan 以外） | 親カルテ: [medical_record_repository_update_test.go](../../../backend/internal/medicalrecord/medical_record_repository_update_test.go) の `expectedVersion` 不一致 Conflict。検査改訂: [examination_revision_workflow_safety_test.go](../../../backend/internal/medicalrecord/examination_revision_workflow_safety_test.go) の stale version | 親レコードと検査 revision には CAS がある | 治療 / バイタル / 処方 / 接種 Update に `expectedVersion` は見つからない。clinical_plan version だけで全タブ保護済みとしない |
| 明細編集と確定 | 明細作成は親 `LockAndFindByID`（[billing_item_service_create.go](../../../backend/internal/billing/billing_item_service_create.go)）。実ロック: [accounting_repository_tx_atomicity_test.go](../../../backend/internal/billing/accounting_repository_tx_atomicity_test.go) の FOR UPDATE 説明 | 同一請求行の明細 write を tx 内直列化 | complete が古い画面合計を再検証して拒否するかは complete テストに無い。**GAP** |
| 取消/通信断後の再試行 | complete の N番目失敗 rollback: `TestAccountingService_CompleteAccounting_NthItemFailure_FullRollback` / `…_DBAtomicRollback`。FE は mutation 単位で key 再利用 | 同一 mutation の再送は増やさない設計 | 切断後に **新しい key** で再発行する2セッション手順のテストは **GAP**。占有ロックの期限/解放は対象外 |
| 医院分離 | [accounting_repository_clinic_isolation_test.go](../../../backend/internal/billing/accounting_repository_clinic_isolation_test.go) が他院 FOR UPDATE を拒否 | clinic scope を破らない | 2セッション実機の分離確認は UNKNOWN |

## 最小サーバー側提案（占有ロックは採用しない）

[設計思想](../../product-philosophy.md) の順で、存在しない画面占有を最適化しない。確認ダイアログや全面ロックは安全性の成立根拠にしない。実装は **未カバー write に既存の expectedVersion / 行ロック / UNIQUE / Idempotency-Key を伸ばす** ことに限定する。製品コード・新規テストファイルはこの票の範囲外。

1. **expectedVersion（CAS）** — clinical_plan と親 medical_record に既にある。GAP の治療・バイタル・処方・接種など last-write-wins の更新へ、読み込み版を更新条件に含める。`expectedVersion == nil` のスキップ経路は新規 caller に広げない。
2. **行ロック（FOR UPDATE / ambient tx）** — 請求親と vaccination claim に既にある。complete と未請求ソース行の確定を同じ tx で固定し、古い画面の合計確定を commit 前に再評価する。ambient tx 不在は fail-closed。
3. **UNIQUE / 業務キー** — 同一 key は `uq_billings_clinic_completion_request_id`。異なる key の二重 complete は冪等キーでは不足。`idx_billings_medical_record_id_unique` / hospitalization unique の 23505 を Conflict へ写し、2キー同時 complete の RED を先に置く。治療 provenance は接種/検査と同様の lifetime UNIQUE（`treatment_id IS NOT NULL`）を検討し、exam の並行 claim テスト欠落も RED 対象。
4. **Idempotency-Key** — 再送・連打専用。別端末が別 UUID を使うケースの防御には使わない。不明結果を新しい key で再発行しない契約を UI とサーバで揃える。

**採否（再掲）:** 画面占有ロックは未採用。開始/期限/解除/切断復旧を持つ occupancy は本票では提案しない。既存防御で足りる経路は再実装しない。

## 2 セッションの合成再現と採否ゲート

1. 上の対応表を衝突キー（clinic＋対象カルテ/明細/請求/未請求項目）の正本とする。GAP 行だけ RED→最小修正。防止目的を医院へ再質問しない。要件責任者の個人名と仕様変更の受入参照は変更前に実行票へ記録する。
2. scopedテストでRED→最小修正→GREENを実施し、候補mount済みの専用Docker/DBで実トランザクション競合と2ブラウザセッションの挙動を検証する。既存version/状態検証/一意性/行ロック/冪等性を優先し、確認ダイアログだけで安全性を成立させない。
3. ケース別にHTTP/競合表示、勝者の保存値、請求/支払/監査件数、rollback、再読込/再試行、医院分離、cleanupを記録する。外部通知がある場合はlocal stubを使い、実送信しない。すべての必須ケースの証拠が揃って完了。環境不足や未実行は残件にする。

**採否:** 全面的な画面占有ロックは未採用。既存防御で満たすケースは再実装しない。それでも占有が必要な不足ケースだけ、開始/期限/解除/切断復旧/権限/監査を別途裁定する。患者情報・秘密を本票に残さない。
