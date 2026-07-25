# AnimalEkarte — TODO

> 更新: 2026-07-24（BE9 implementation完了後reconciliation）

## 運用

- 本書は、エージェントが直ちに着手できる未完了タスクの台帳とする。
- タスクは「個別タスク詳細」節に `### <タスクID>: <タイトル>` 形式で追加する。
- 対応済みセクションは削除し、完了記録はgit履歴と各実装testを正本とする。
- GitHub Issueと対応するタスクはIssueのstateを実測し、Issue一覧を本書へ重複掲出しない。
- release/運用gateは実装タスクと混在させず、[`q&a.html` OPS-13〜17](q&a.html#ops)と該当runbookで追跡する。

## 正本の境界

| 内容 | 正本 |
|------|------|
| 着手可能な実装タスク | 本書の「個別タスク詳細」 |
| GitHub Issueのstate・一覧 | GitHub Issues |
| BE9構造移行・進捗・release gate | BE9は2026-07-24にcode complete（release pending）。境界は[ADR-006](docs/architecture/adr/006-backend-domain-package-boundaries.md) / [boundary map](docs/architecture/be9-2a-boundary-map.md)、release gateは[`q&a.html` OPS-13〜17](q&a.html#ops) |
| FEデザイン準拠・リファクタリング計画 | [`FE-refactor.md`](FE-refactor.md) |
| BE10 backend規約適合（フォルダ構成）リファクタ計画 | [`BE-refactor.md`](BE-refactor.md) |
| 今フェーズで着手しない事項 | [`phase2.html`](phase2.html) |
| 着手保留・任意検証のBE技術債 | [`BE-pending.md`](BE-pending.md) |
| PO判断・USER実操作・P0ブロッカー | [`q&a.html`](q&a.html) |
| Issueを3セッションで着手するためのガイドview（正本=各Issue・受け入れ条件を複製しない） | [`3-session-agent.html`](3-session-agent.html)（削除・退役の対象にしない） |

## 個別タスク詳細

### TASK-ADR003: 予約⇔会計の支払方法二重保持解消（ADR-003 案1B TRIGGER）

- PO-006 裁定済み。DEC-9（2026-07-25 q&a.html）で GitHub Issue 起票を待たず本書追跡へ変更。**着手時期 = 納品前（2026-07-26〜27）** — 2026-07-25 の納品日 7/27 延期（理由=残作業の全対応）に伴い「納品後」から前倒し。
- 内容の正本 = q&a.html PO-006／DEC-9。USER が Issue 起票したら本エントリを Issue へ移設し二重掲出しない。

### TASK-251: 締め集計 category contract 確定実装（#251・8→12分類）

- 業務決裁確定（q&a.html DEC-21・**USER 本人裁定** 2026-07-25）。**着手時期 = 納品前（2026-07-26〜27）** — 2026-07-25 の納品日 7/27 延期（理由=残作業の全対応）に伴い S3 送りから前倒し。contract 正本 = DEC-21・#251 本文（本エントリは実装スコープと着手時期の入口であり決裁の「なぜ」は複製しない）。
- Phase 0 棚卸し（外部エージェント調査・Fable spot-verify）で確定した実装スコープ:
  - ① 正式カテゴリ = 12分類（enum 現状追認）。#251 タイトル「8分類」→「12分類」修正は Issue 本文転記（USER 承認後）に含める。
  - ② hospitalization 退院会計の other 固定を撤廃し CarePlanItem.Type／Procedure.IsSurgery→category resolver（`backend/internal/medicalrecord/hospitalization_service.go:431`）。treatment 経路（`backend/internal/billing/billing_item_service.go:405,462`）と共通化＝category contract 単一ソース化。
  - ③ vaccination を接種記録（`Vaccination.VaccineID`→Vaccine）から会計明細自動生成。`BillingItem` へ VaccineID provenance 列追加の migration が必要。自動化は停止／失敗通知／監査／idempotency（原則⑤）。
  - ④ hotel source=`HospitalizationTypeHotel`（②連動）、training は新規 source 設計。両カテゴリ維持。
  - 含意(a) category authority を BE resolver に一本化し FE/client は保持しない。
  - 含意(b) 締め集計の未知値 fail-closed = 生カラム無制限 GROUP BY（`backend/internal/billing/accounting_repository_reports_close.go:44`・`cash_register_service.go:265`）を12値 allowlist 経由にし typo/legacy を締め表へ黙って通さない（受け入れ条件「unknown/legacy を黙って変換しない」）。
  - 含意(d) 全書込経路（treatment/hospitalization/vaccination/trimming/merchandise/manual）を同一 typed category source に集約。
- #247（月次統合表）は本 TASK の contract 完了後に着手。
- Issue #251 本文への決裁転記（タイトル「8分類」→「12分類」修正含む）と着手時期の前倒し反映は、いずれも 2026-07-25 に USER 承認のうえ完了済み（live read-back で実測確認）。todo.md / q&a.html DEC-21 / #251 本文の3者は同期済み。以後 contract の参照先は #251 本文と DEC-21 とし、本エントリは着手時期と実装スコープのみを持つ。
- 出典: #251 Phase 0 棚卸し Completion Report（2026-07-25・DEC-21）。

### SEC-SWEEP-01: 単一pet_id FKを持つread経路の親pets clinic相関 全数掃引

- read-only調査。BUG-429（`acb3e4929`で修正済）と同型の防御ギャップが他packageに残っていないかを確定する。過去のクロステナントread IDOR監査（13 repo修正）はBE-012（慢性疾患）より前に実施されており、BUG-429はその監査漏れだった。以後に追加された子テーブルreadに同じ漏れがある可能性は実在する。
- 検出対象: 子テーブルが `pet_id` 単一FK（clinic複合FKでない）を持ち、read述語が `clinic_id = ? AND pet_id = ?` 相当のみで親petsへの相関JOIN/EXISTSを欠くもの。修正済みの参照実装は `backend/internal/pet/chronic_condition_repository.go:37,52`。
- 成果物: 該当package/メソッドの file:line 一覧と、各々の実害経路判定（service層で親所有検証が先行するか＝防御されているか、BUG-429の`List`のように素通しか）。修正は本タスクに含めず、件数確定後に別途起票する。
- 相関にpets側の `deleted_at` / `deceased_at` を含めないこと（含めるとsoft-delete済・死亡ペットの履歴が黙って消える挙動回帰になる）。これは掃引結果を修正へ展開する際の必須制約。
- 出典: BUG-429対応時に判明したNew Work（2026-07-25）。
- **セッション分類 = S3（#239 の先行条件）**。台帳は BUG-429 を「#239実装前のsecurity blocker」と分類しており、本掃引はその同一欠陥クラス（`pet_id` 単一FKで親pets clinic相関を欠くread）の全数である。加えて #239「医院別レコードを残す同一owner/petリンクと所属院内の統合履歴」は**意図された cross-clinic read を新設する**機能であり、意図しない漏洩を残したまま実装すると両者を区別できなくなる。よって残9件の修正は #239 着手前に完了させる。


#### SEC-SWEEP-01 実行結果（2026-07-25）

- **判定: AUDIT COMPLETE / attribution gate BLOCKED / live exposure 9件**。schema-first + raw/backtick SQL補完 + clinic-isolation独立passで state (B)=3、state (C)=9 を再現。行データを返す7経路は **CRITICAL**、件数のみ2経路は **HIGH**。コード修正・test追加は本unit外のため未実施。
- schema universeは `grep -cE "REFERENCES +pets" backend/migrations/*.sql` で `001_init.sql:10`、他migrationは0、`grep -nE "FOREIGN KEY.*pet"` は出力なし。10表は `001_init.sql:1144,1308,1336,1400,1422,1463,1487,1505,1731,1876`。
- current model mappingは10件すべて明示的: `pet_chronic_conditions`→`model.PetChronicCondition` (`backend/internal/model/pet_chronic_condition.go:27`)、`appointments`→`model.Reservation` (`model/reservation.go:78`)、`hospitalizations`→`model.Hospitalization` (`model/hospitalization.go:55`)、`medical_records`→`model.MedicalRecord` (`model/medical_record.go:53`)、`prescriptions`→`model.Prescription` (`model/prescription.go:28`)、`vaccinations`→`model.Vaccination` (`model/vaccination_record.go:45`)、`checkups`→`model.Checkup` (`model/checkup_record.go:30`)、`exams`→`model.Examination` (`model/examination_record.go:54`)、`vital_records`→`model.VitalRecord` (`model/vital.go:40-42`)、`billings`→`model.Billing` (`model/accounting.go:94`)。生成時の「9 explicit + 1 implicit」はstale。
- consumer discoveryはprompt指定のmodel型grep + quoted-table grepに加え、raw/backtick/JOIN identifierのschema-first passを実施。指定grepだけでは `backend/internal/lstep/checkup_sync_repository.go:116,168,170` 等を落とすため補完した。
- scoped baselineは空だったが、最終照合時のsnapshotには本unitの `M todo.md` に加え、別sessionの `M backend/docs/api.yaml`、`M backend/internal/reservation/nested_summary_response.go`、`M backend/internal/reservation/reservation_response_test.go` が出現した。本unitの書込操作は `todo.md` のみだが、prompt指定のadded-set gateはbackend pathを含むため **BLOCKED**。調査中に先行していたservice/audit系driftは外部sessionにより消え、上記reservation/API driftへ変化した。いずれも変更・復元・stageしていない。

##### table summary

- A/B/C/Dはglobal unique methodのstateを各consumer-table associationへ投影した件数。同一methodは後段inventoryで1回だけ分類し、複数tableを読む場合は`tables`列にまとめるため、table間の単純合計はglobal totalと一致しない。

| table | production read files | A/B/C/D | census rows |
|---|---|---:|---:|
| `pet_chronic_conditions` | `backend/internal/pet/chronic_condition_repository.go`<br>`backend/internal/lstep/checkup_sync_repository.go` | 3/0/0/4 | 7 |
| `appointments` | `backend/internal/billing/billing_item_repository.go`<br>`backend/internal/clinic/clinic_repository.go`<br>`backend/internal/csvimport/cutover_import.go`<br>`backend/internal/reservation/appointment_admin_repository.go`<br>`backend/internal/reservation/reservation_intent_repository.go`<br>`backend/internal/reservation/reservation_repository.go`<br>`backend/internal/reservation/reservation_type_repository.go`<br>`backend/internal/trimming/trimming_course_repository.go`<br>`backend/internal/trimming/trimming_option_repository.go` | 2/0/2/87 | 91 |
| `hospitalizations` | `backend/internal/clinic/clinic_repository.go`<br>`backend/internal/staff/staff_repository.go`<br>`backend/internal/medicalrecord/cage_repository.go`<br>`backend/internal/medicalrecord/care_plan_item_repository.go`<br>`backend/internal/medicalrecord/hospitalization_repository.go`<br>`backend/internal/medicalrecord/medicine_repository.go`<br>`backend/internal/medicalrecord/procedure_repository.go` | 0/0/1/65 | 66 |
| `medical_records` | `backend/internal/billing/accounting_repository_ltv.go`<br>`backend/internal/billing/billing_confirmation_repository.go`<br>`backend/internal/billing/billing_item_repository.go`<br>`backend/internal/clinic/clinic_repository.go`<br>`backend/internal/csvimport/cutover_import.go`<br>`backend/internal/lstep/checkup_sync_repository.go`<br>`backend/internal/lstep/lstep_delivery_trigger_log_repository.go`<br>`backend/internal/medicalrecord/checkup_repository.go`<br>`backend/internal/medicalrecord/clinical_plan_repository.go`<br>`backend/internal/medicalrecord/examination_repository.go`<br>`backend/internal/medicalrecord/inquiry_repository.go`<br>`backend/internal/medicalrecord/medical_record_image_repository.go`<br>`backend/internal/medicalrecord/medical_record_owner_visit_repository.go`<br>`backend/internal/medicalrecord/medical_record_repository.go`<br>`backend/internal/medicalrecord/treatment_repository.go`<br>`backend/internal/medicalrecord/vaccination_repository.go`<br>`backend/internal/reservation/reservation_intent_repository.go`<br>`backend/internal/reservation/reservation_repository.go`<br>`backend/internal/staff/staff_repository.go`<br>`backend/internal/owner/ltv_repository.go`<br>`backend/internal/inventory/repository.go`<br>`backend/internal/medicalrecord/chief_complaint_repository.go`<br>`backend/internal/medicalrecord/diagnosis_name_repository.go`<br>`backend/internal/medicalrecord/consultation_repository.go`<br>`backend/internal/persistence/scope.go` (helper-only, census 0) | 5/2/5/180 | 192 |
| `prescriptions` | `backend/internal/medicalrecord/prescription_repository.go` | 0/0/0/6 | 6 |
| `vaccinations` | `backend/internal/clinic/clinic_repository.go`<br>`backend/internal/csvimport/cutover_import.go`<br>`backend/internal/medicalrecord/vaccination_repository.go`<br>`backend/internal/medicalrecord/vaccine_repository.go` | 1/0/0/28 | 29 |
| `checkups` | `backend/internal/clinic/clinic_repository.go`<br>`backend/internal/lstep/checkup_sync_repository.go`<br>`backend/internal/medicalrecord/checkup_field_repository.go`<br>`backend/internal/medicalrecord/checkup_repository.go`<br>`backend/internal/medicalrecord/checkup_type_repository.go` | 0/0/1/32 | 33 |
| `exams` | `backend/internal/clinic/clinic_repository.go`<br>`backend/internal/staff/staff_repository.go`<br>`backend/internal/csvimport/cutover_import.go`<br>`backend/internal/medicalrecord/exam_type_repository.go`<br>`backend/internal/medicalrecord/examination_repository.go`<br>`backend/internal/medicalrecord/lab_import_repository.go`<br>`backend/internal/medicalrecord/medical_record_image_repository.go` | 1/1/0/55 | 57 |
| `vital_records` | `backend/internal/csvimport/cutover_import.go`<br>`backend/internal/medicalrecord/daily_record_repository.go`<br>`backend/internal/medicalrecord/medical_record_repository.go`<br>`backend/internal/medicalrecord/vital_repository.go`<br>`backend/internal/staff/staff_repository.go` | 1/2/0/37 | 40 |
| `billings` | `backend/internal/billing/accounting_repository.go`<br>`backend/internal/billing/accounting_repository_ltv.go`<br>`backend/internal/billing/accounting_repository_reports_close.go`<br>`backend/internal/billing/accounting_repository_reports_daily.go`<br>`backend/internal/billing/accounting_repository_reports_monthly.go`<br>`backend/internal/billing/accounting_repository_unpaid.go`<br>`backend/internal/billing/billing_item_repository.go`<br>`backend/internal/billing/payment_method_master_repository.go`<br>`backend/internal/billing/refund_repository.go`<br>`backend/internal/clinic/clinic_repository.go`<br>`backend/internal/inventory/merchandise_item_repository.go`<br>`backend/internal/lstep/checkup_sync_repository.go`<br>`backend/internal/medicalrecord/medical_record_repository.go`<br>`backend/internal/medicalrecord/treatment_repository.go`<br>`backend/internal/owner/ltv_repository.go`<br>`backend/internal/staff/staff_repository.go`<br>`backend/internal/billing/accounting_repository_actor.go` (helper-only, census 0)<br>`backend/internal/billing/accounting_repository_reports_shared.go` (helper-only, census 0)<br>`backend/internal/csvimport/cutover_payment_target.go` (helper-only, census 0) | 1/2/7/90 | 100 |

##### state (C) — exposed

- **CRITICAL** `billingItemRepository.FindUnbilledTrimmingItemsByPetID`: `backend/internal/billing/billing_item_repository.go:342-346,369-374`。`GET /billing-items/unbilled` は `billing/routes.go:109` → `billing_item_handler.go:106-112` → `billing_item_service.go:417` で親pet検証なし。提案: `EXISTS (SELECT 1 FROM pets p WHERE p.id = a.pet_id AND p.clinic_id = a.clinic_id)`。
- **HIGH** `billingItemRepository.CountNonAccountingTrimmingByPetAndDate`: `backend/internal/billing/billing_item_repository.go:424-426`。`GET /billing-items/ungrouped-same-day` は `billing/routes.go:110` → `billing_item_handler.go:160-166` → `billing_item_service.go:435` で親pet検証なし。同じappointments親相関を提案。
- **CRITICAL** `hospitalizationRepository.FindAll`: `backend/internal/medicalrecord/hospitalization_repository.go:47-50`。`GET /hospitalizations` は `medicalrecord/routes.go:314-315` → `hospitalization_handler.go:36-53` → `hospitalization_service.go:246-248` で親pet検証なし。提案: `EXISTS (SELECT 1 FROM pets p WHERE p.id = hospitalizations.pet_id AND p.clinic_id = hospitalizations.clinic_id)`。
- **CRITICAL** `treatmentRepository.FindUnbilledByPetID`: `backend/internal/medicalrecord/treatment_repository.go:82-92`。`GET /billing-items/unbilled` は `billing/routes.go:109` → `billing_item_handler.go:100-112` → `billing_item_service.go:405-407` で親pet検証なし。提案: `EXISTS (SELECT 1 FROM pets p WHERE p.id = mr.pet_id AND p.clinic_id = mr.clinic_id)`。
- **CRITICAL** `treatmentRepository.FindHistoryByPetID`: `backend/internal/medicalrecord/treatment_repository.go:100-105`。`GET /pets/:id/treatment-history` は `medicalrecord/routes.go:267` → `treatment_handler.go:78-103` → `treatment_service.go:206-212` で親pet検証なし。提案: `EXISTS (SELECT 1 FROM pets p WHERE p.id = medical_records.pet_id AND p.clinic_id = medical_records.clinic_id)`。
- **HIGH** `treatmentRepository.CountFinalizedUnconfirmedByPetAndDate`: `backend/internal/medicalrecord/treatment_repository.go:194-203`。`GET /billing-items/ungrouped-same-day` は `billing/routes.go:110` → `billing_item_handler.go:155-166` → `billing_item_service.go:427-428` で親pet検証なし。提案: `EXISTS (SELECT 1 FROM pets p WHERE p.id = medical_records.pet_id AND p.clinic_id = medical_records.clinic_id)`。
- **CRITICAL** `checkupFieldResultRepository.FindByPetID`: `backend/internal/medicalrecord/checkup_field_repository.go:88-96`。`GET /checkups/field-results` は `medicalrecord/routes.go:218` → `checkup_field_handler.go:88-96` → `checkup_field_result_service.go:104-105` で親pet検証なし。提案: `EXISTS (SELECT 1 FROM pets p WHERE p.id = checkups.pet_id AND p.clinic_id = checkups.clinic_id)`。
- **CRITICAL** `accountingRepository.FindAll`: `backend/internal/billing/accounting_repository.go:192-194,205-210`。`GET /accountings` は `billing/routes.go:117-118` → `accounting_handler.go:36-68` → `accounting_service_core.go:12-13` で親pet検証なし。
- **CRITICAL** `accountingRepository.FindAllForClinics`: `backend/internal/billing/accounting_repository.go:197-210`。同entryのmulti-clinic分岐 `accounting_handler.go:69-80` → `accounting_service_core.go:21-22` で親pet検証なし。
- nullableなbillingsは `billings.pet_id IS NULL OR EXISTS (SELECT 1 FROM pets p WHERE p.id = billings.pet_id AND p.clinic_id = billings.clinic_id)` とし、非pet請求を保持する。全提案で親側 `deleted_at` / `deceased_at` を含めない。

##### state (B) — guarded by caller

- `medicalRecordRepository.CountByPetID`（述語 `backend/internal/medicalrecord/medical_record_repository.go:460-466`）: caller 1は `backend/internal/pet/service.go:336` のclinic-scoped `FindByID` が `:340` より先行。caller 2は `medical_record_handler.go:62-72` のrelation-scoped lookupが `:78-82` より先行し、親pets相関は `medical_record_repository.go:230-242`。
- `medicalRecordRepository.FindFirstVisitDateByPetID`（述語 `backend/internal/medicalrecord/medical_record_repository.go:476-484`）: `backend/internal/pet/service.go:218-224` でclinic-scoped `petRepo.FindByID` が先行。
- `labImportRepository.IsDuplicate`（述語 `backend/internal/medicalrecord/lab_import_repository.go:202-211`）: non-nil petでは `backend/internal/medicalrecord/lab_import_examination_service.go:165` の `petRepo.FindByID` が重複read `:185` より先行。nil petはpet-key readなし。

##### calibration / follow-up

- calibration PASS: `chronicConditionRepository.FindByPetID`=A (`backend/internal/pet/chronic_condition_repository.go:37`)、`FindByID`=A (`:52`)、`FindActiveConditionCodesByOwner`=A (`:102-103`)。
- 修正は別unitで9件を同じ親clinic相関shapeへ収束させる。raw SQL/GORM双方を検出するstatic lintを既存6 lintの隣へ追加する価値が高い。
- `daily_records` / `care_logs` / `exam_results` / `billing_items` / `medical_record_images` / `medical_record_addenda` 等のgrandchild相関は本unit外。別classとしてschema-firstで起票・掃引する。
- runtime suite / coverage / lintはGo source不変更のため不実施。generated/export artifactなし、stageなし、tracked/ignored probeは非該当。

##### verification / failure signature

- saved prompt validator: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-sec-sweep-01-pet-parent-clinic-correlation.md` → `Prompt Craft Harness Validation: PASS`, exit 0。
- census reconciliation: 54 production files（helper-only 4を含む）、`A=8 B=3 C=9 D=334 total=354`、actual census mismatch 0、duplicate classification key 0。`grep -c "^### " todo.md` はbefore/afterとも6、`grep -c "^### SEC-SWEEP-01" todo.md` は1、`git diff --check -- todo.md` は出力なし。
- Failure Signature 1: consumer wrapper attempt 1はzshでtable/type pairを分割せずoverbroad model searchになったため破棄。`table:type` parsingへ変更して成功し、失敗出力をevidenceに使用していない。
- Failure Signature 2: 初回write-back生成はMarkdown backtick未escapeによりJS `SyntaxError: Unexpected identifier 'grep'`。writeなし。backtickを文字列として安全に構築して成功。
- Failure Signature 3: completeness addendum attempt 1はpatch行prefix欠落で`Invalid patch hunk`、attempt 2は古いsummary contextで不一致。いずれもwriteなし。current exact contextを再読して適用成功。
- Failure Signature 4: unique-inventory置換 attempt 1はBUG-430境界文字列の仮定違いで`BUG-430 boundary not found`、writeなし。実見出しを再読してattempt 2で成功。

##### full classification inventory (global unique methods)

- 同一fileが複数tableを読む場合も exported method は全体で1回だけ分類する。`tables` はconsumer associationであり、stateはmethod全体のpet-key readについて解決済み。global totals: A=8 / B=3 / C=9 / D=334 / total=354。

| tables | file:line | method | state | evidence |
|---|---|---|---|---|
| `billings`<br>`medical_records` | `backend/internal/billing/accounting_repository_ltv.go:29` | `SumPaidByOwner` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records` | `backend/internal/billing/accounting_repository_ltv.go:45` | `MaxSingleVisitAmountByOwner` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records` | `backend/internal/billing/accounting_repository_ltv.go:67` | `FindOwnersByAnnualRevenue` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository_reports_close.go:14` | `GetCloseAggregate` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository_reports_daily.go:13` | `GetDailySummary` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository_reports_monthly.go:14` | `GetMonthlyReport` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository_reports_monthly.go:21` | `GetMonthlyReportByPeriod` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository_unpaid.go:87` | `FindUnpaidByBilling` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository_unpaid.go:111` | `FindUnpaidByOwner` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository_unpaid.go:151` | `SumUnpaidByOwner` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository_unpaid.go:170` | `FindMonthlyUnpaidCarryover` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository.go:192` | `FindAll` | C | 述語 `:192-194,205-210`; entry `accounting_handler.go:36-68` |
| `billings` | `backend/internal/billing/accounting_repository.go:197` | `FindAllForClinics` | C | 述語 `:197-210`; entry `accounting_handler.go:69-80` |
| `billings` | `backend/internal/billing/accounting_repository.go:279` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository.go:283` | `FindByIDForClinics` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository.go:320` | `LockAndFindByID` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository.go:350` | `Create` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository.go:366` | `Update` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository.go:386` | `SavePayment` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/accounting_repository.go:469` | `SavePaymentSplits` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/billing/billing_confirmation_repository.go:33` | `FindByMedicalRecordID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/billing/billing_confirmation_repository.go:45` | `Create` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/billing/billing_confirmation_repository.go:54` | `Update` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/billing/billing_confirmation_repository.go:73` | `LockActiveStaffAssignment` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go:50` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go:62` | `FindByBillingID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go:108` | `ValidateCreateReferences` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go:230` | `Create` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go:240` | `Update` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go:256` | `Delete` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go:269` | `UpdateBillingTotals` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go:288` | `HasItemByOwnerSince` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go:304` | `HasFoodPurchaseByOwnerSince` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go:321` | `FindUnbilledTrimmingItemsByPetID` | C | 述語 `:342-346,369-374`; entry `billing_item_handler.go:112` |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go:421` | `CountNonAccountingTrimmingByPetAndDate` | C | 述語 `:424-426`; entry `billing_item_handler.go:166` |
| `billings` | `backend/internal/billing/payment_method_master_repository.go:31` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/payment_method_master_repository.go:43` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/payment_method_master_repository.go:54` | `Create` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/payment_method_master_repository.go:61` | `Update` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/payment_method_master_repository.go:68` | `Delete` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/payment_method_master_repository.go:74` | `CountUsageByPaymentMethodID` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/payment_method_master_repository.go:87` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/refund_repository.go:32` | `Create` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/refund_repository.go:68` | `FindByBillingID` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/refund_repository.go:105` | `SumByBillingID` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/billing/refund_repository.go:125` | `SumByBillingIDAndPaymentMethod` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:23` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:34` | `FindByStaffID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:50` | `LockActiveByID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:68` | `LockByIDForUpdate` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:83` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:92` | `FindCompany` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:105` | `Create` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:116` | `Update` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:130` | `Delete` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:141` | `CountOwnersByClinicID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:153` | `CountStaffByClinicID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go:166` | `CountBlockingReferencesByClinicID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`exams`<br>`medical_records`<br>`vaccinations`<br>`vital_records` | `backend/internal/csvimport/cutover_import.go:88` | `CopyFrom` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/inventory/merchandise_item_repository.go:32` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/inventory/merchandise_item_repository.go:44` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/inventory/merchandise_item_repository.go:48` | `Create` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/inventory/merchandise_item_repository.go:55` | `Update` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/inventory/merchandise_item_repository.go:62` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/inventory/merchandise_item_repository.go:68` | `CountUsageByMerchandiseItemID` | D | pet_id read keyなし、またはwrite |
| `billings` | `backend/internal/inventory/merchandise_item_repository.go:90` | `Delete` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/inventory/repository.go:61` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/inventory/repository.go:101` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/inventory/repository.go:105` | `Create` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/inventory/repository.go:117` | `Update` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/inventory/repository.go:124` | `Delete` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/inventory/repository.go:133` | `DecreaseStock` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/inventory/repository.go:150` | `DeleteByNameAndMedicineCategory` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/inventory/repository.go:163` | `UpdateNameByMedicineCategory` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/inventory/repository.go:177` | `CountUsageByInventoryID` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`checkups`<br>`medical_records`<br>`pet_chronic_conditions` | `backend/internal/lstep/checkup_sync_repository.go:77` | `FindCheckupSyncPreview` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go:67` | `Create` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go:74` | `ExistsTodayByOwnerAndType` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go:88` | `UpdateStatus` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go:107` | `CountByStatusAndDateRange` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go:132` | `CountExcludedReasonByDateRange` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go:157` | `CountSuppressedByPriorityDateRange` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go:172` | `FindByDateRangeWithFilters` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go:213` | `CountByTypeAndStatus` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go:229` | `CountVisitConversionsByType` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go:267` | `FindByOwnerAndDate` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go:281` | `UpdateSuppressed` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/cage_repository.go:35` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/cage_repository.go:47` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/cage_repository.go:51` | `Create` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/cage_repository.go:58` | `Update` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/cage_repository.go:65` | `Delete` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/cage_repository.go:69` | `CountUsageByCageID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/cage_repository.go:81` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/care_plan_item_repository.go:37` | `FindByHospitalizationID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/care_plan_item_repository.go:52` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/care_plan_item_repository.go:66` | `Create` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/care_plan_item_repository.go:74` | `Update` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/care_plan_item_repository.go:91` | `Delete` | D | pet_id read keyなし、またはwrite |
| `checkups` | `backend/internal/medicalrecord/checkup_field_repository.go:50` | `FindByCheckupTypeID` | D | pet_id read keyなし、またはwrite |
| `checkups` | `backend/internal/medicalrecord/checkup_field_repository.go:71` | `FindByCheckupID` | D | pet_id read keyなし、またはwrite |
| `checkups` | `backend/internal/medicalrecord/checkup_field_repository.go:88` | `FindByPetID` | C | 述語 `:93-96`; entry `checkup_field_handler.go:88-96` |
| `checkups` | `backend/internal/medicalrecord/checkup_field_repository.go:114` | `ReplaceForCheckup` | D | pet_id read keyなし、またはwrite |
| `checkups`<br>`medical_records` | `backend/internal/medicalrecord/checkup_repository.go:54` | `FindByClinicID` | D | pet_id read keyなし、またはwrite |
| `checkups`<br>`medical_records` | `backend/internal/medicalrecord/checkup_repository.go:95` | `FindByOwnerID` | D | pet_id read keyなし、またはwrite |
| `checkups`<br>`medical_records` | `backend/internal/medicalrecord/checkup_repository.go:112` | `FindByMedicalRecordID` | D | pet_id read keyなし、またはwrite |
| `checkups`<br>`medical_records` | `backend/internal/medicalrecord/checkup_repository.go:130` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `checkups`<br>`medical_records` | `backend/internal/medicalrecord/checkup_repository.go:148` | `LockByIDForUpdate` | D | pet_id read keyなし、またはwrite |
| `checkups`<br>`medical_records` | `backend/internal/medicalrecord/checkup_repository.go:164` | `Create` | D | pet_id read keyなし、またはwrite |
| `checkups`<br>`medical_records` | `backend/internal/medicalrecord/checkup_repository.go:172` | `Update` | D | pet_id read keyなし、またはwrite |
| `checkups`<br>`medical_records` | `backend/internal/medicalrecord/checkup_repository.go:176` | `Delete` | D | pet_id read keyなし、またはwrite |
| `checkups` | `backend/internal/medicalrecord/checkup_type_repository.go:36` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `checkups` | `backend/internal/medicalrecord/checkup_type_repository.go:45` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `checkups` | `backend/internal/medicalrecord/checkup_type_repository.go:49` | `Create` | D | pet_id read keyなし、またはwrite |
| `checkups` | `backend/internal/medicalrecord/checkup_type_repository.go:57` | `Update` | D | pet_id read keyなし、またはwrite |
| `checkups` | `backend/internal/medicalrecord/checkup_type_repository.go:64` | `Delete` | D | pet_id read keyなし、またはwrite |
| `checkups` | `backend/internal/medicalrecord/checkup_type_repository.go:68` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `checkups` | `backend/internal/medicalrecord/checkup_type_repository.go:74` | `CountUsageByCheckupTypeID` | D | pet_id read keyなし、またはwrite |
| `checkups` | `backend/internal/medicalrecord/checkup_type_repository.go:87` | `CountChildrenByParentID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/chief_complaint_repository.go:33` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/chief_complaint_repository.go:45` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/chief_complaint_repository.go:49` | `Create` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/chief_complaint_repository.go:57` | `Update` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/chief_complaint_repository.go:64` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/chief_complaint_repository.go:70` | `CountUsageByChiefComplaintTypeID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/chief_complaint_repository.go:82` | `Delete` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/clinical_plan_repository.go:33` | `FindByMedicalRecordID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/clinical_plan_repository.go:49` | `Create` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/clinical_plan_repository.go:98` | `Update` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/clinical_plan_repository.go:131` | `Delete` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/consultation_repository.go:37` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/consultation_repository.go:46` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/consultation_repository.go:50` | `Create` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/consultation_repository.go:58` | `Update` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/consultation_repository.go:65` | `Delete` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/consultation_repository.go:69` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/consultation_repository.go:75` | `CountChildrenByParentID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/consultation_repository.go:89` | `CountUsageByConsultationID` | D | pet_id read keyなし、またはwrite |
| `vital_records` | `backend/internal/medicalrecord/daily_record_repository.go:39` | `FindByHospitalizationID` | D | pet_id read keyなし、またはwrite |
| `vital_records` | `backend/internal/medicalrecord/daily_record_repository.go:54` | `FindByHospitalizationIDAndDate` | D | pet_id read keyなし、またはwrite |
| `vital_records` | `backend/internal/medicalrecord/daily_record_repository.go:68` | `FindOrCreateByDate` | D | pet_id read keyなし、またはwrite |
| `vital_records` | `backend/internal/medicalrecord/daily_record_repository.go:83` | `CreateVitalRecord` | D | pet_id read keyなし、またはwrite |
| `vital_records` | `backend/internal/medicalrecord/daily_record_repository.go:91` | `CreateCareLog` | D | pet_id read keyなし、またはwrite |
| `vital_records` | `backend/internal/medicalrecord/daily_record_repository.go:99` | `CreateStaffNote` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/diagnosis_name_repository.go:36` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/diagnosis_name_repository.go:54` | `FindAllByCategoryID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/diagnosis_name_repository.go:77` | `FindAllByFilter` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/diagnosis_name_repository.go:90` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/diagnosis_name_repository.go:94` | `Create` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/diagnosis_name_repository.go:102` | `Update` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/diagnosis_name_repository.go:109` | `Delete` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/diagnosis_name_repository.go:125` | `CountUsageByDiagnosisNameID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/diagnosis_name_repository.go:138` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/exam_type_repository.go:35` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/exam_type_repository.go:44` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/exam_type_repository.go:53` | `Create` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/exam_type_repository.go:61` | `Update` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/exam_type_repository.go:68` | `Delete` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/exam_type_repository.go:72` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/exam_type_repository.go:77` | `CountUsageByExamTypeID` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/exam_type_repository.go:90` | `CountChildrenByParentID` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/examination_repository.go:44` | `FindAll` | A | pet filter `:50`、親相関 `:193-201` |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/examination_repository.go:84` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/examination_repository.go:101` | `LockByIDForUpdate` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/examination_repository.go:120` | `FindByJobID` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/examination_repository.go:139` | `Create` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/examination_repository.go:148` | `Update` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/examination_repository.go:156` | `Delete` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/examination_repository.go:257` | `CountItemsByExamID` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/examination_repository.go:274` | `FindAllItemsByExamID` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/examination_repository.go:297` | `ReplaceItemsByExamID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/hospitalization_repository.go:43` | `FindAll` | C | 述語 `:47-50`; entry `hospitalization_handler.go:43-53` |
| `hospitalizations` | `backend/internal/medicalrecord/hospitalization_repository.go:74` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/hospitalization_repository.go:99` | `LockByIDForUpdate` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/hospitalization_repository.go:111` | `Create` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/hospitalization_repository.go:122` | `Update` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/hospitalization_repository.go:136` | `UpdateIfNotDischarged` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/hospitalization_repository.go:151` | `Delete` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/hospitalization_repository.go:162` | `CountCarePlanItemsByHospitalizationID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/hospitalization_repository.go:175` | `CountDailyRecordsByHospitalizationID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/hospitalization_repository.go:188` | `CountTreatmentPlansByHospitalizationID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/inquiry_repository.go:47` | `SaveByMedicalRecordID` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/lab_import_repository.go:51` | `Create` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/lab_import_repository.go:58` | `Update` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/lab_import_repository.go:91` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/lab_import_repository.go:113` | `Create` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/lab_import_repository.go:120` | `FindByJob` | D | pet_id read keyなし、またはwrite |
| `exams` | `backend/internal/medicalrecord/lab_import_repository.go:202` | `IsDuplicate` | B | guard `backend/internal/medicalrecord/lab_import_examination_service.go:165`→`:185` |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/medical_record_image_repository.go:35` | `FindByMedicalRecordID` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/medical_record_image_repository.go:52` | `Create` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/medical_record_image_repository.go:60` | `Delete` | D | pet_id read keyなし、またはwrite |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/medical_record_image_repository.go:76` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/medical_record_owner_visit_repository.go:32` | `FindLatestByOwner` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/medical_record_owner_visit_repository.go:51` | `FindOwnerVisitSummary` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/medical_record_owner_visit_repository.go:81` | `FindOwnersByFirstVisitDate` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/medical_record_owner_visit_repository.go:104` | `FindOwnersByLastVisitDays` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/medical_record_owner_visit_repository.go:129` | `FindOwnersByNextVisitRecommended` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/medical_record_owner_visit_repository.go:161` | `FindDormantOwnerEntries` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/medical_record_owner_visit_repository.go:193` | `FindDormantOwnerEntriesCursor` | D | pet_id read keyなし、またはwrite |
| `medical_records` | `backend/internal/medicalrecord/medical_record_owner_visit_repository.go:206` | `FindDormantOwnerEntriesCursorAt` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:121` | `FindAll` | A | pet filter `:154-155`、親相関 `:230-242` |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:334` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:338` | `FindByAppointmentID` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:356` | `FindByIDForClinics` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:382` | `Create` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:398` | `Update` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:428` | `Delete` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:460` | `CountByPetID` | B | guard `backend/internal/pet/service.go:336`→`:340`; `medical_record_handler.go:72`→`:81` |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:476` | `FindFirstVisitDateByPetID` | B | guard `backend/internal/pet/service.go:221`→`:224` |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:493` | `CountByOwnerID` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:511` | `LockByIDForUpdate` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go:526` | `CountEstimatesByMedicalRecordID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/medicine_repository.go:34` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/medicine_repository.go:54` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/medicine_repository.go:60` | `CountUsageByMedicineID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/medicine_repository.go:79` | `CountChildrenByParentID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/medicine_repository.go:91` | `Create` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/medicine_repository.go:99` | `Update` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/medicine_repository.go:106` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/medicine_repository.go:110` | `Delete` | D | pet_id read keyなし、またはwrite |
| `prescriptions` | `backend/internal/medicalrecord/prescription_repository.go:37` | `FindByMedicalRecordID` | D | pet_id read keyなし、またはwrite |
| `prescriptions` | `backend/internal/medicalrecord/prescription_repository.go:50` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `prescriptions` | `backend/internal/medicalrecord/prescription_repository.go:63` | `FindActiveByOwner` | D | pet_id read keyなし、またはwrite |
| `prescriptions` | `backend/internal/medicalrecord/prescription_repository.go:78` | `Create` | D | pet_id read keyなし、またはwrite |
| `prescriptions` | `backend/internal/medicalrecord/prescription_repository.go:87` | `Update` | D | pet_id read keyなし、またはwrite |
| `prescriptions` | `backend/internal/medicalrecord/prescription_repository.go:103` | `Delete` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/procedure_repository.go:37` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/procedure_repository.go:45` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/procedure_repository.go:49` | `Create` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/procedure_repository.go:56` | `Update` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/procedure_repository.go:63` | `Delete` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/procedure_repository.go:69` | `CountUsageByProcedureID` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/procedure_repository.go:88` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `hospitalizations` | `backend/internal/medicalrecord/procedure_repository.go:93` | `CountChildrenByParentID` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records` | `backend/internal/medicalrecord/treatment_repository.go:54` | `FindByMedicalRecordID` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records` | `backend/internal/medicalrecord/treatment_repository.go:68` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records` | `backend/internal/medicalrecord/treatment_repository.go:82` | `FindUnbilledByPetID` | C | 述語 `:85-89`; entry `billing_item_handler.go:112` |
| `billings`<br>`medical_records` | `backend/internal/medicalrecord/treatment_repository.go:100` | `FindHistoryByPetID` | C | 述語 `:100-105`; entry `treatment_handler.go:78-103` |
| `billings`<br>`medical_records` | `backend/internal/medicalrecord/treatment_repository.go:194` | `CountFinalizedUnconfirmedByPetAndDate` | C | 述語 `:194-203`; entry `billing_item_handler.go:166` |
| `billings`<br>`medical_records` | `backend/internal/medicalrecord/treatment_repository.go:214` | `Create` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records` | `backend/internal/medicalrecord/treatment_repository.go:222` | `Update` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records` | `backend/internal/medicalrecord/treatment_repository.go:240` | `Delete` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records` | `backend/internal/medicalrecord/treatment_repository.go:254` | `BulkUpdateSortOrder` | D | pet_id read keyなし、またはwrite |
| `medical_records`<br>`vaccinations` | `backend/internal/medicalrecord/vaccination_repository.go:45` | `FindAll` | A | pet filter `:54`、親相関 `:51,169-178` |
| `medical_records`<br>`vaccinations` | `backend/internal/medicalrecord/vaccination_repository.go:85` | `FindByOwner` | D | pet_id read keyなし、またはwrite |
| `medical_records`<br>`vaccinations` | `backend/internal/medicalrecord/vaccination_repository.go:101` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `medical_records`<br>`vaccinations` | `backend/internal/medicalrecord/vaccination_repository.go:113` | `LockByIDForUpdate` | D | pet_id read keyなし、またはwrite |
| `medical_records`<br>`vaccinations` | `backend/internal/medicalrecord/vaccination_repository.go:128` | `Create` | D | pet_id read keyなし、またはwrite |
| `medical_records`<br>`vaccinations` | `backend/internal/medicalrecord/vaccination_repository.go:135` | `Update` | D | pet_id read keyなし、またはwrite |
| `medical_records`<br>`vaccinations` | `backend/internal/medicalrecord/vaccination_repository.go:142` | `Delete` | D | pet_id read keyなし、またはwrite |
| `medical_records`<br>`vaccinations` | `backend/internal/medicalrecord/vaccination_repository.go:216` | `FindOwnersByVaccineDeadline` | D | pet_id read keyなし、またはwrite |
| `vaccinations` | `backend/internal/medicalrecord/vaccine_repository.go:38` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `vaccinations` | `backend/internal/medicalrecord/vaccine_repository.go:50` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `vaccinations` | `backend/internal/medicalrecord/vaccine_repository.go:62` | `Create` | D | pet_id read keyなし、またはwrite |
| `vaccinations` | `backend/internal/medicalrecord/vaccine_repository.go:69` | `Update` | D | pet_id read keyなし、またはwrite |
| `vaccinations` | `backend/internal/medicalrecord/vaccine_repository.go:76` | `Delete` | D | pet_id read keyなし、またはwrite |
| `vaccinations` | `backend/internal/medicalrecord/vaccine_repository.go:80` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `vaccinations` | `backend/internal/medicalrecord/vaccine_repository.go:85` | `CountUsageByVaccineID` | D | pet_id read keyなし、またはwrite |
| `vaccinations` | `backend/internal/medicalrecord/vaccine_repository.go:98` | `CountChildrenByParentID` | D | pet_id read keyなし、またはwrite |
| `vital_records` | `backend/internal/medicalrecord/vital_repository.go:40` | `FindByMedicalRecordID` | D | pet_id read keyなし、またはwrite |
| `vital_records` | `backend/internal/medicalrecord/vital_repository.go:54` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `vital_records` | `backend/internal/medicalrecord/vital_repository.go:70` | `Create` | D | pet_id read keyなし、またはwrite |
| `vital_records` | `backend/internal/medicalrecord/vital_repository.go:78` | `Update` | D | pet_id read keyなし、またはwrite |
| `vital_records` | `backend/internal/medicalrecord/vital_repository.go:94` | `Delete` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`medical_records` | `backend/internal/owner/ltv_repository.go:88` | `FindOwnerLTV` | D | pet_id read keyなし、またはwrite |
| `pet_chronic_conditions` | `backend/internal/pet/chronic_condition_repository.go:33` | `FindByPetID` | A | 相関 `backend/internal/pet/chronic_condition_repository.go:37` |
| `pet_chronic_conditions` | `backend/internal/pet/chronic_condition_repository.go:45` | `FindByID` | A | 相関 `backend/internal/pet/chronic_condition_repository.go:52` |
| `pet_chronic_conditions` | `backend/internal/pet/chronic_condition_repository.go:59` | `Create` | D | pet_id read keyなし、またはwrite |
| `pet_chronic_conditions` | `backend/internal/pet/chronic_condition_repository.go:66` | `Update` | D | pet_id read keyなし、またはwrite |
| `pet_chronic_conditions` | `backend/internal/pet/chronic_condition_repository.go:84` | `Delete` | D | pet_id read keyなし、またはwrite |
| `pet_chronic_conditions` | `backend/internal/pet/chronic_condition_repository.go:97` | `FindActiveConditionCodesByOwner` | A | calibration相関 `backend/internal/pet/chronic_condition_repository.go:102-103` |
| `appointments` | `backend/internal/reservation/appointment_admin_repository.go:37` | `FindAllByMonth` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/appointment_admin_repository.go:56` | `FindAllByDay` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/appointment_admin_repository.go:81` | `FindTimeRangesByDateRange` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/appointment_admin_repository.go:95` | `Create` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/appointment_admin_repository.go:102` | `SoftDelete` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/appointment_admin_repository.go:106` | `FindAllByCustomerID` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/appointment_admin_repository.go:121` | `FindByIDForNotify` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/appointment_admin_repository.go:135` | `CancelByID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go:42` | `CompleteForAccounting` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go:102` | `BackfillForMedicalRecord` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go:188` | `AssertMedicalRecordDoctorInClinic` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go:241` | `PrepareForMedicalRecordFinalization` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go:260` | `MarkNoShow` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go:266` | `MarkNoShowAt` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go:323` | `CreateForTrimming` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go:410` | `FindTrimmingByID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go:440` | `LockTrimmingByID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go:505` | `UpdateForTrimming` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go:591` | `DeleteForTrimming` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:248` | `FindAll` | A | pet filter `:277`、親相関 `:54-59` |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:292` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:296` | `FindByIDForClinics` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:335` | `Create` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:352` | `Delete` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:363` | `ExistsByReservationTypeID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:375` | `ExistsByStaffID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:387` | `FindClinicIDsByStaffID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:412` | `CountMedicalRecordsByReservationID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:428` | `AcquireBookingLock` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:444` | `LockAndFindByID` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:463` | `HasDoctorConflict` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:491` | `CountOnDutyDoctors` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:512` | `CountConflicts` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:545` | `CountByTypeAndStartTime` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:572` | `CountByTypeAndStartTimes` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:597` | `CountByCustomerAndDateRange` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:611` | `FindAllByCategory` | A | 親JOIN相関 `backend/internal/reservation/reservation_repository.go:620-624` |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:667` | `CountByDateAndSource` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:682` | `FindNoShowCandidates` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:688` | `FindNoShowCandidatesAt` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:730` | `AssertOwnerInClinic` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:749` | `FindPetOwnerInClinic` | D | pet_id read keyなし、またはwrite |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go:767` | `AssertLineCustomerInClinic` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/reservation_type_repository.go:36` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/reservation_type_repository.go:48` | `FindAllWithChildren` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/reservation_type_repository.go:64` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/reservation_type_repository.go:80` | `FindByIDWithChildren` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/reservation_type_repository.go:97` | `Create` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/reservation_type_repository.go:104` | `Update` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/reservation_type_repository.go:119` | `Delete` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/reservation_type_repository.go:131` | `CountUsageByReservationTypeID` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/reservation_type_repository.go:144` | `CountChildrenByParentID` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/reservation/reservation_type_repository.go:157` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:60` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:87` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:119` | `FindByIDInClinic` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:145` | `LockActiveByIDForUpdate` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:165` | `LockActiveByIDForUpdateInClinic` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:189` | `LockActiveByIDForShare` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:206` | `FindByAccountID` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:215` | `Create` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:225` | `Update` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:240` | `UpdatePrimaryClinicID` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:255` | `Delete` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:308` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:329` | `CountBlockingReferencesByStaffID` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:394` | `CreateForReservation` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:423` | `UpdateForReservation` | D | pet_id read keyなし、またはwrite |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go:430` | `SwapSortOrderForReservation` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_course_repository.go:34` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_course_repository.go:42` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_course_repository.go:54` | `Create` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_course_repository.go:61` | `Update` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_course_repository.go:68` | `Delete` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_course_repository.go:74` | `CountUsageByTrimmingCourseID` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_course_repository.go:87` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_option_repository.go:34` | `FindAll` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_option_repository.go:42` | `FindByID` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_option_repository.go:54` | `Create` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_option_repository.go:61` | `Update` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_option_repository.go:68` | `Delete` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_option_repository.go:72` | `Reorder` | D | pet_id read keyなし、またはwrite |
| `appointments` | `backend/internal/trimming/trimming_option_repository.go:78` | `CountUsageByTrimmingOptionID` | D | pet_id read keyなし、またはwrite |

##### per-file census equality (global unique files)

- helper-only direct-read 4 files are carried with `0 = 0`; exported receiver methodがないためclassification rowもない。

| tables | file | census | rows | result |
|---|---|---:|---:|---|
| `billings` | `backend/internal/billing/accounting_repository_actor.go` | 0 | 0 | PASS |
| `billings`<br>`medical_records` | `backend/internal/billing/accounting_repository_ltv.go` | 3 | 3 | PASS |
| `billings` | `backend/internal/billing/accounting_repository_reports_close.go` | 1 | 1 | PASS |
| `billings` | `backend/internal/billing/accounting_repository_reports_daily.go` | 1 | 1 | PASS |
| `billings` | `backend/internal/billing/accounting_repository_reports_monthly.go` | 2 | 2 | PASS |
| `billings` | `backend/internal/billing/accounting_repository_reports_shared.go` | 0 | 0 | PASS |
| `billings` | `backend/internal/billing/accounting_repository_unpaid.go` | 4 | 4 | PASS |
| `billings` | `backend/internal/billing/accounting_repository.go` | 9 | 9 | PASS |
| `medical_records` | `backend/internal/billing/billing_confirmation_repository.go` | 4 | 4 | PASS |
| `appointments`<br>`billings`<br>`medical_records` | `backend/internal/billing/billing_item_repository.go` | 11 | 11 | PASS |
| `billings` | `backend/internal/billing/payment_method_master_repository.go` | 7 | 7 | PASS |
| `billings` | `backend/internal/billing/refund_repository.go` | 4 | 4 | PASS |
| `appointments`<br>`billings`<br>`checkups`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vaccinations` | `backend/internal/clinic/clinic_repository.go` | 12 | 12 | PASS |
| `appointments`<br>`exams`<br>`medical_records`<br>`vaccinations`<br>`vital_records` | `backend/internal/csvimport/cutover_import.go` | 1 | 1 | PASS |
| `billings` | `backend/internal/csvimport/cutover_payment_target.go` | 0 | 0 | PASS |
| `billings` | `backend/internal/inventory/merchandise_item_repository.go` | 7 | 7 | PASS |
| `medical_records` | `backend/internal/inventory/repository.go` | 9 | 9 | PASS |
| `billings`<br>`checkups`<br>`medical_records`<br>`pet_chronic_conditions` | `backend/internal/lstep/checkup_sync_repository.go` | 1 | 1 | PASS |
| `medical_records` | `backend/internal/lstep/lstep_delivery_trigger_log_repository.go` | 11 | 11 | PASS |
| `hospitalizations` | `backend/internal/medicalrecord/cage_repository.go` | 7 | 7 | PASS |
| `hospitalizations` | `backend/internal/medicalrecord/care_plan_item_repository.go` | 5 | 5 | PASS |
| `checkups` | `backend/internal/medicalrecord/checkup_field_repository.go` | 4 | 4 | PASS |
| `checkups`<br>`medical_records` | `backend/internal/medicalrecord/checkup_repository.go` | 8 | 8 | PASS |
| `checkups` | `backend/internal/medicalrecord/checkup_type_repository.go` | 8 | 8 | PASS |
| `medical_records` | `backend/internal/medicalrecord/chief_complaint_repository.go` | 7 | 7 | PASS |
| `medical_records` | `backend/internal/medicalrecord/clinical_plan_repository.go` | 4 | 4 | PASS |
| `medical_records` | `backend/internal/medicalrecord/consultation_repository.go` | 8 | 8 | PASS |
| `vital_records` | `backend/internal/medicalrecord/daily_record_repository.go` | 6 | 6 | PASS |
| `medical_records` | `backend/internal/medicalrecord/diagnosis_name_repository.go` | 9 | 9 | PASS |
| `exams` | `backend/internal/medicalrecord/exam_type_repository.go` | 8 | 8 | PASS |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/examination_repository.go` | 10 | 10 | PASS |
| `hospitalizations` | `backend/internal/medicalrecord/hospitalization_repository.go` | 10 | 10 | PASS |
| `medical_records` | `backend/internal/medicalrecord/inquiry_repository.go` | 1 | 1 | PASS |
| `exams` | `backend/internal/medicalrecord/lab_import_repository.go` | 6 | 6 | PASS |
| `exams`<br>`medical_records` | `backend/internal/medicalrecord/medical_record_image_repository.go` | 4 | 4 | PASS |
| `medical_records` | `backend/internal/medicalrecord/medical_record_owner_visit_repository.go` | 8 | 8 | PASS |
| `billings`<br>`medical_records`<br>`vital_records` | `backend/internal/medicalrecord/medical_record_repository.go` | 12 | 12 | PASS |
| `hospitalizations` | `backend/internal/medicalrecord/medicine_repository.go` | 8 | 8 | PASS |
| `prescriptions` | `backend/internal/medicalrecord/prescription_repository.go` | 6 | 6 | PASS |
| `hospitalizations` | `backend/internal/medicalrecord/procedure_repository.go` | 8 | 8 | PASS |
| `billings`<br>`medical_records` | `backend/internal/medicalrecord/treatment_repository.go` | 9 | 9 | PASS |
| `medical_records`<br>`vaccinations` | `backend/internal/medicalrecord/vaccination_repository.go` | 8 | 8 | PASS |
| `vaccinations` | `backend/internal/medicalrecord/vaccine_repository.go` | 8 | 8 | PASS |
| `vital_records` | `backend/internal/medicalrecord/vital_repository.go` | 5 | 5 | PASS |
| `billings`<br>`medical_records` | `backend/internal/owner/ltv_repository.go` | 1 | 1 | PASS |
| `medical_records` | `backend/internal/persistence/scope.go` | 0 | 0 | PASS |
| `pet_chronic_conditions` | `backend/internal/pet/chronic_condition_repository.go` | 6 | 6 | PASS |
| `appointments` | `backend/internal/reservation/appointment_admin_repository.go` | 8 | 8 | PASS |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_intent_repository.go` | 11 | 11 | PASS |
| `appointments`<br>`medical_records` | `backend/internal/reservation/reservation_repository.go` | 24 | 24 | PASS |
| `appointments` | `backend/internal/reservation/reservation_type_repository.go` | 10 | 10 | PASS |
| `billings`<br>`exams`<br>`hospitalizations`<br>`medical_records`<br>`vital_records` | `backend/internal/staff/staff_repository.go` | 16 | 16 | PASS |
| `appointments` | `backend/internal/trimming/trimming_course_repository.go` | 7 | 7 | PASS |
| `appointments` | `backend/internal/trimming/trimming_option_repository.go` | 7 | 7 | PASS |



### BUG-430: stage-importの医院非限定DELETE

- CRITICAL。`backend/cmd/stage-import` のdeleteScopeが `owner_id >= 300000`（pets経由の継承含む）でclinic_id非限定。実行すると他院の高番ownerデータを削除し得る。`backend/cmd/stage-import/main_test.go:217-246` がこの挙動をテストで固定化している（cross-clinic保護テストは無い）。
- 対応方針変更（2026-07-25 DEC-20）: **stage-import退役で解消**（deleteScope修正には投資しない）。本番cutoverはrunbook既定の21表csv-import正式経路であり本ツールは本番使用禁止・local限定＋`--confirm-local-destroy`ガード既存。退役実装=cmd/stage-import削除またはビルド除外（#250再基準化転記とセット・USER承認後）。
- 出典: #251調査 Completion Report（2026-07-25）。テスト実測で確認済み。

### BUG-433: 生成FE型がGoドメインモデル由来のため、応答DTOに無いフィールドが型上は存在扱いになる

- HIGH（サイレント機能不全の生成器）。**S3/S2いずれにも属さない横断課題**。`frontend/src/types/generated/models.ts` は tygo が `backend/internal/model/` から生成しており（同ファイル冒頭コメント）、OpenAPI／応答DTOからではない。このため FE の型は *Goドメインモデル* を写し、HTTP が実際に返す *応答DTO* とは一致しない。DTOに無いフィールドは実行時 `undefined` なのに型検査は通る。
- 実害の実例: BUG-431（受付の危険度バッジが実APIで一度も点灯しなかった・`463e07424` で修正）は本ドリフトの1インスタンスに過ぎない。fixtureは型どおり作られるためテストでも検出されない。
- 実測された残存ギャップ: 生成 `Pet` は31プロパティ、修正後の予約pet DTOは9。残22フィールド（`clinic_id` `owner_id` `animal_species_id` `name_kana` `gender` `birth_date` `color` `blood_type` `microchip_number` `neutered_date` `acquisition_type` `food` `environment` `phone` `last_visit` `insurance_id` `remarks` `deceased_at` `deceased_reason` `created_at` `updated_at` `insurance`）は型上は利用可能だがワイヤに存在しない。他モデル（Owner/Reservation等）も同構造。
- 対応方針（未確定・要判断）: ①応答DTOからFE型を生成する経路へ切り替える ②生成型を「ドメインモデル」と明示リネームし、画面が使う型は応答DTO由来へ分離する ③現状維持で個別に埋める（BUG-431と同じ対症）。①②は生成基盤の変更を伴うため納品後が妥当。納品前は、新規に生成型のフィールドへ依存する実装を書くときに**そのフィールドが応答DTOに実在するかを都度確認する**運用で凌ぐ。
- 出典: BUG-431 修正時に判明したNew Work（2026-07-25・executorが残22フィールドを実測列挙）。

### BUG-432: 飼主生年月日がフォームから保存されない＋一覧列がpet値を表示

- HIGH。DB（`owners.birth_date`）・BE DTO・OpenAPI・DatePickerは実装済みだが、`frontend/src/features/owners/hooks/use-owner-form.ts` のcreate/update送信payloadにowner birthDateが含まれず、入力しても保存されない。さらに `OwnersListTable.tsx` の「生年月日」列はownerでなくpetの生年月日を表示している。
- 対応: payloadへbirth_date追加＋既存値を空へ戻す契約（JSON null vs 省略）の確定＋一覧列の正本（飼主DOB/ペットDOB）確定。#262の前提是正。
- 出典: #262調査 Completion Report（2026-07-25）。grepで整合確認済み。

2026-07-23に起票したBUG-421〜428、TEST-ROUTES-01、FMT-BE-01は2026-07-24のBE9実装でsource/testへ反映済みのため、本active listから削除した。release pending項目（fresh DB migration、remote CI/coverage、production deploy/ops rehearsal）は実装taskではないため本書へ再掲しない。
