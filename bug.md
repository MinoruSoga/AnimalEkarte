# AnimalEkarte バグ監査台帳

> **取扱区分**: Internal / Security-sensitive
> **更新日**: 2026-07-14
> **目的**: コード監査で発見した未修正バグの正本。修正状況、受入条件、検証結果をここへ集約する。
>
> **現状**: 優先表 AUD-001〜008 はすべて Closed。2026-07-14 Mode 3 再検証（実装証拠 + scoped test）で再確認済み。Open なし。

## 運用ルール

- 1件ずつTDDで修正し、複数AUDを同時に実装しない。
- 着手前に `git status` と対象ファイルの `git diff` を確認し、別作業の変更を保護する。
- 行番号は参考情報とし、実装時はシンボル名で再特定する。
- Backend/FrontendのコマンドはDocker内でスコープ限定して実行する。
- フルテスト、フルlint、build、type-check、DB reset、migration適用、直接SQLは自動実行しない。
- clinic分離に関する修正は `clinic-isolation-auditor` と `security-reviewer` のレビューを完了条件に含める。
- 修正前後で秘密情報・飼主情報・ペット情報・運用情報をログやエラー応答へ追加しない。
- 外部Issueへ転記する場合は、攻撃手順や実データを含めず、公開範囲を確認する。

## 優先順位

| 修正順 | ID | 概要 | Severity | Confidence | セキュリティ判定 | 状態 |
|---:|---|---|---|---|---|---|
| 1 | AUD-001 | 予約のOwner/Pet/LineCustomer clinic分離漏れ | High | High confidence | Approve | Closed |
| 2 | AUD-008 | カルテ本体のOwner/Pet clinic分離漏れ | High | High confidence | Approve | Closed |
| 3 | AUD-003 | DailyRecordの親clinic未検証とview GET書込み | High | High confidence | Approve | Closed |
| 4 | AUD-002 | 会計の関連FK clinic分離漏れ | High | High confidence | Approve | Closed |
| 5 | AUD-004 | 入院のOwner/Pet clinic分離漏れ | High | High confidence | Approve | Closed |
| 6 | AUD-005 | 見積の関連FK clinic分離漏れとcreated_by偽装 | High | High confidence | Approve | Closed |
| 7 | AUD-007 | 第2診断カテゴリのAPI/DB契約分裂 | Medium-High | High confidence | - | Closed |
| 8 | AUD-006 | 入院日誌バイタルのPetID欠落と部分コミット | Medium | High confidence | - | Closed |

---

## AUD-001: 予約のOwner/Pet/LineCustomer clinic分離漏れ

**状態**: Closed

**Severity**: High

**Confidence**: High confidence

**clinic_id境界監査**: Approve

### 症状

- 通常予約Create/Updateへ別clinicのOwner/Pet IDを指定できる。
- 管理予約Createへ別clinicのOwner/Pet/LineCustomer IDを指定できる。
- トリミング予約Create、既存予約へのトリミング詳細追加、トリミング予約Updateへ別clinicのPet IDを指定できる。
- カルテCreateがOwner/Pet未設定の予約へ、別clinicのOwner/Petを補完できる。
- 保存後の一覧・詳細取得で、別clinicの飼主名・ペット情報・LINE顧客情報を返し得る。

### 根本原因

- `reservationService.Create` / `reservationService.Update` がOwner/Petをclinic所有確認せず永続化する。
- `reservationAdminService.Create` がOwner/Pet/LineCustomerをclinic所有確認せず永続化する。
- `trimmingService.Create` / `createDetailForExistingAppointment` / `Update` がPetをclinic所有確認せず永続化する。
- `medicalRecordService.applyAppointmentContextForCreate` がrequest由来Owner/Petを所有確認せず予約へ補完する。
- 管理予約RepositoryのCreateがambient transactionへ参加せず、検証とwriteが別transactionになる。
- DBは関連IDの存在を確認する単純FKのみで、予約と関連行のclinic一致を保証しない。
- 予約のOwner/Pet/LineCustomer Preloadにclinic条件がない。

### 主な根拠

- `backend/internal/service/reservation_service.go` — `Create`, `Update`, `buildReservationUpdate`
- `backend/internal/service/appointment_admin_service.go` — `Create`
- `backend/internal/service/trimming_service.go` — `Create`, `createDetailForExistingAppointment`, `Update`
- `backend/internal/service/medical_record_crud.go` — `applyAppointmentContextForCreate`
- `backend/internal/repository/reservation_repository.go` — `reservationListPreloads`
- `backend/internal/repository/appointment_admin_repository.go` — 月次・日次一覧Preload
- `backend/migrations/001_init.sql` — `appointments` の関連FK、FORCEなしRLS

### 修正スコープ

- 通常予約Create/UpdateのOwner/Pet所有確認とOwner-Pet整合確認。
- 管理予約CreateのOwner/Pet/LineCustomer所有確認。
- トリミング予約の3書込み経路におけるPet所有確認。
- カルテ作成に伴う予約Owner/Pet補完の所有確認とOwner-Pet整合確認。
- 管理予約Createをambient transactionへ参加させる。
- LINE顧客のOwner.Pets Preloadをclinic scopeし、汚染関係の再伝播を防ぐ。
- 通常予約、管理予約、トリミング予約の対象書込みと検証を同一transaction境界で実施する。
- 既存不整合データからの漏洩を防ぐ防御的Preload条件。
- 初回修正ではDB migration、APIフィールド変更、他AUDの修正を行わない。

### 受入条件

- [x] Clinic Aの通常予約CreateにClinic BのOwnerを指定するとNotFound系エラーとなり、予約が作成されない。
- [x] Clinic Aの通常予約CreateにClinic BのPetを指定すると拒否される。
- [x] Clinic Aの管理予約CreateにClinic BのOwner/Pet/LineCustomerを個別に指定すると拒否される。
- [x] Clinic Aの予約UpdateにClinic BのOwner/Petを指定すると拒否され、既存値が変化しない。
- [x] Petが最終Ownerに属さないOwner-Pet組合せをCreate/Updateとも拒否する。
- [x] nil関連、Ownerのみ、Petのみの既存仕様を明示的にテストし、意図しない互換性変更を起こさない。
- [x] 同一clinicかつ整合したOwner/Pet/LineCustomerは従来どおり成功する。
- [x] 汚染済みFKを持つ予約を取得しても、別clinicのOwner/Pet/LineCustomer情報をレスポンスへPreloadしない。
- [x] repository errorを握りつぶさず、既存の`apperrors`規約で返す。
- [x] 対象テスト、`TestPreloadClinicScope`、`TestMasterFKWriteInventory`がPASSする。
- [x] `clinic-isolation-auditor` と `security-reviewer` のCRITICAL/HIGH指摘が解消されている。
- [x] Clinic Aのトリミング予約CreateにClinic BのPetを指定すると拒否され、予約・詳細が作成されない。
- [x] Clinic Aの既存予約へトリミング詳細を追加する際、Clinic BのPetを指定すると拒否され、予約・詳細が変化しない。
- [x] Clinic Aのトリミング予約UpdateにClinic BのPetを指定すると拒否され、既存値が変化しない。
- [x] トリミング3経路のPet検証と書込みが同一transactionに参加し、失敗時にrollbackされる。
- [x] 再監査で `clinic-isolation-auditor` と `security-reviewer` のCRITICAL/HIGH指摘が解消されている。
- [x] カルテCreateから予約へ別clinicのOwner/Petを補完できず、Owner-Pet不一致も拒否される。
- [x] 管理予約Createが所有確認と同じambient transactionへ参加し、rollbackされる。
- [x] LINE顧客の汚染Owner-Pet関係から別clinic Petを予約へ再付与しない。

### リスク

- Constructor/DIと既存mockへの影響がある。
- UpdateではPATCH値と現在値を統合した最終Owner/Petを検証しないと、片側だけ変更した場合に不整合が残る。
- UI選択肢が自院に限定されていてもraw JSONは改変可能なため、Handler/UIだけの検証では不十分。
- Appointment連携カルテの予約補完は異院IDをfail-fastで拒否するが、カルテCreateと予約Updateは単一transactionではない。競合・部分成功リスクはMediumの後続改善事項とする。
- AppointmentなしMedicalRecord本体のOwner/Pet分離漏れは予約FKとは別問題であり、AUD-008（Closed）で追跡・修正済み。

---

## AUD-002: 会計の関連FK clinic分離漏れ

**状態**: Closed

**Severity**: High

**Confidence**: High confidence

**clinic_id境界監査**: Approve

### 症状・原因

会計Create/Updateの `medical_record_id`, `hospitalization_id`, `owner_id`, `pet_id` を所有確認せず保存する。Billing本体だけがclinic scopeされ、Owner/Petは無条件Preloadされるため、別clinicの個人情報露出と会計・診療・入院データの不整合が成立する。

### 主な根拠

- `backend/internal/service/accounting_service_core.go` — `Create`, `Update`
- `backend/internal/service/accounting_service_builders.go` — `buildAccountingUpdate`
- `backend/internal/repository/accounting_repository.go` — Owner/Pet Preload
- `backend/migrations/001_init.sql` — `billings` の関連FK

### 再オープン理由（Mode 3）

- 2026-07-14 の Closed 記録は write-path FK 検証（`validateAccountingRelatedFKs` / DI / CrossClinic テスト）がコードに存在しない状態での誤クローズだった。実行可能証拠に基づき再オープンする。Preload `clinic_id` は実装済みのため本フォローの対象外。

### 修正内容

- Mode 3 write-path: `validateAccountingRelatedFKs` を Create/Update の persist 前に追加。Owner/Pet は `validateReservationOwnerPetLinks` 再利用。Update は最終 FK 合成後に検証。completed Create は同一 WithTx 内で検証。
- `NewAccountingService` に MedicalRecord/Hospitalization/Reservation repo を DI（`service.go`）。
- Named tests: `TestAccountingService_Create_RejectsCrossClinicRelatedFKs` / `Update_RejectsCrossClinicRelatedFKs` / `Create_CompletedValidatesInsideTx`。
- Preload `clinic_id` は Mode 3 以前に実装済みのため本フォローでは未変更。
- 2026-07-14 誤 Closed を再オープン後、上記実行可能証拠で再クローズ。

### 受入条件

- [x] 4関連IDをそれぞれ別clinicへ差し替えたCreate/Updateを拒否する。
- [x] Create失敗時は行が作られず、Update失敗時は既存値が変化しない。
- [x] MedicalRecord/Hospitalization/Owner/Pet相互の整合性を検証する。
- [x] 一覧・詳細から別clinicのOwner/Pet情報を返さない。

---

## AUD-003: DailyRecordの親clinic未検証とview GET書込み

**状態**: Closed

**Severity**: High

**Confidence**: High confidence

**clinic_id境界監査**: Approve

### 症状・原因

- `GET /hospitalizations/:id/daily-records/:date` がview権限で `FirstOrCreate` を実行する。
- DailyRecordServiceが親Hospitalizationのclinic所有を確認しない。
- DBは `clinic_id` と `hospitalization_id` の一致を保証しない。
- UNIQUEは日付単独ではなく `(hospitalization_id, date)` だがclinic_idを含まないため、別clinicが対象入院・対象日を先取りすると正常作成を妨害できる。

### 主な根拠

- `backend/internal/handler/daily_record_handler.go` — `GetDailyRecord`, route登録
- `backend/internal/service/daily_record_service.go` — `FindOrCreateByDate` 呼出し
- `backend/internal/repository/daily_record_repository.go` — `FirstOrCreate`
- `backend/migrations/001_init.sql` — `daily_records` FK/UNIQUE/RLS

### 修正スコープ

- 全 DailyRecord 経路（List/GetByDate/FindOrCreate/AddVital/AddCareLog/AddStaffNote）で親 Hospitalization の clinic 所有確認（`hospRepo.FindByID`）。
- view GET を `GetByDate`（Find のみ）に分離。未存在は NotFound。作成は create 権限の POST。
- OpenAPI `getDailyRecord` を自動作成なし・404 に更新。FE は既存 404→create CTA をスコープテストで固定。
- 初回修正では DB migration（UNIQUE への clinic_id 追加等）は行わない。

### 受入条件

- [x] すべてのGET/POST経路でHospitalizationのclinic所有を確認する。
- [x] view権限のGETはDBを書き換えない。
- [x] 未存在日のGET挙動を既存FE・元仕様・OpenAPI間で統一する。
- [x] Clinic AからClinic BのHospitalization IDを指定してもDailyRecordを作成できない。
- [x] 拒否後もClinic Bが同日DailyRecordを正常作成できる。

---

## AUD-004: 入院のOwner/Pet clinic分離漏れ

**状態**: Closed

**Severity**: High

**Confidence**: High confidence

**clinic_id境界監査**: Approve

### 症状・原因

Hospitalization Create/UpdateはCageを検証する一方、Owner/Petを所有確認せず保存する。Owner/Pet Preloadもclinic無条件であり、退院会計へ汚染された関連IDが伝播する。

### 主な根拠

- `backend/internal/service/hospitalization_service.go` — `Create`, `Update`, `DischargeWithBilling`
- `backend/internal/repository/hospitalization_repository.go` — Owner/Pet Preload
- `backend/migrations/001_init.sql` — `hospitalizations` の関連FK

### 修正内容

- Create/Update で `validateReservationOwnerPetLinks`（AUD-001再利用）により Owner/Pet clinic所有と Owner-Pet 整合を persist 前に検証。
- Update は最終マージ後の Owner/Pet を検証。DischargeWithBilling は会計作成前に再検証。
- Owner/Pet Preload へ clinic 条件。commit `6ee2e419`。migration なし。

### 受入条件

- [x] 別clinicのOwner/Petを指定したCreate/Updateを拒否する。
- [x] Owner-Pet不一致を拒否する。
- [x] 拒否時に行未作成・既存値不変を保証する。
- [x] 退院会計へ異院関連IDが伝播しない。

---

## AUD-005: 見積の関連FK clinic分離漏れとcreated_by偽装

**状態**: Closed

**Severity**: High

**Confidence**: High confidence

**clinic_id境界監査**: Approve

### 症状・原因

見積Createが `medical_record_id`, `owner_id`, `created_by` をrequest bodyから受け取り、所有確認せず保存する。`created_by` は認証主体ではないため、同一clinic内でも任意スタッフを偽装できる。

### 主な根拠

- `backend/internal/handler/estimate_request.go` — `createEstimateRequest`
- `backend/internal/handler/estimate_handler.go` — `CreateEstimate`
- `backend/internal/service/estimate_service.go` — `Create`
- `backend/internal/repository/estimate_repository.go` — Owner/CreatedStaff Preload

### 修正内容

- `validateEstimateRelatedFKs` + MR/Reservation DI。`created_by` は extractStaffID のみ（body削除）。Owner Preload clinic_id。commit `e5a571f6`。migration なし。

### 受入条件

- [x] 別clinicのMedicalRecord/Owner/Staffを指定したCreateを拒否する。
- [x] `created_by` は認証済みactorから設定し、request bodyで偽装できない。
- [x] API契約変更が必要な場合はcontract正本とFrontend利用箇所を同時に更新する。
- [x] 作成直後・一覧・詳細で別clinicのOwner情報を返さない。

---

## AUD-006: 入院日誌バイタルのPetID欠落と部分コミット

**状態**: Closed

**Severity**: Medium

**Confidence**: High confidence

### 症状・原因

`AddVitalRecord` が必須の `VitalRecord.PetID` を設定しないためINSERTが制約違反になる。未作成日のDailyRecord作成とVital作成は別autocommitで、Vital失敗時にDailyRecordだけが残る。

### 主な根拠

- `backend/internal/service/daily_record_service.go` — `AddVitalRecord`
- `backend/internal/model/vital.go` — 必須PetID
- `backend/internal/repository/daily_record_repository.go` — 別autocommit
- `backend/migrations/001_init.sql` — `vital_records.pet_id` NOT NULL/FK

### 修正内容

- `AddVitalRecord` で `loadHospitalization`（AUD-003所有確認再利用）から `PetID` を解決して VitalRecord へ設定。
- `Transactor.WithTx` + repo `dbOrTx` で FindOrCreateByDate と CreateVitalRecord を同一 transaction 化。
- Named tests: PetID 解決、tx atomicity（commit/rollback）。migration なし。

### 受入条件

- [x] 親HospitalizationからPetIDを解決し、VitalRecordへ設定する。
- [x] DailyRecord作成とVital作成を同一transactionに含める。
- [x] Vital作成失敗時に新規DailyRecordもrollbackされる。
- [x] AUD-003の所有確認を前提とし、重複実装を作らない。

---

## AUD-007: 第2診断カテゴリのAPI/DB契約分裂

**状態**: Closed

**Severity**: Medium-High

**Confidence**: High confidence

### 症状・原因

- FEは `diagnosis_2_type_id` を送るが、Handlerは `diagnosis_2_category_id` を待つため第2分類がsilent dropされる。
- 第2病名だけが保存され、不整合状態を作り得る。
- `diagnosis_2_category_id` を直接送るとServiceが存在しない列をUPDATEし、本番PostgreSQLでは400になる見込み。
- Response DTO/Preloadが第2診断を返さない。
- DB/model/generated型は `diagnosis_2_type_id` であり、これをcanonical名とするのが自然。

### 主な根拠

- `frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts`
- `backend/internal/handler/clinical_plan_request.go`
- `backend/internal/service/clinical_plan_service.go`
- `backend/internal/repository/clinical_plan_repository.go`
- `backend/internal/model/clinical_plan.go`
- `backend/internal/handler/clinical_plan_response.go`
- `backend/docs/api.yaml`, `docs/openapi.yaml`

### 修正内容

- Handler request を `diagnosis_2_type_id` に統一。JSON null クリアは `nullableUint64RequestField`（未送信と null を区別）。
- Service update map を実DB列 `diagnosis_2_type_id` に修正。Go フィールドを `Diagnosis2TypeID` にリネーム（**uint64）。
- Model/Response/Preload に第2診断を追加。第2 type↔name 整合を Update で検証。
- CreateSubRecords 経路の誤列名も同修正。OpenAPI UpdateClinicalPlanRequest と FE clinical-plan/save-action を同期。
- migration なし。

### 受入条件

- [x] Request/Service update map/DB/model/Response/OpenAPI/Frontendを `diagnosis_2_type_id` に統一する。
- [x] 第2診断分類・病名の設定、変更、クリアを実DB契約を通すテストで確認する。
- [x] 保存後レスポンスと再取得レスポンスに第2診断が含まれる。
- [x] 第2分類と第2病名の整合性を検証する。
- [x] 誤列名を正解として固定しているmockテストを修正する。

---

## AUD-008: カルテ本体のOwner/Pet clinic分離漏れ

**状態**: Closed

**Severity**: High

**Confidence**: High confidence

**clinic_id境界監査**: Approve

### 症状・原因

- Appointmentを指定しないMedicalRecord Create/Updateがrequest由来Owner/Petをclinic所有確認せず保存する。
- MedicalRecordのOwner/Pet Preloadにclinic条件がなく、汚染FKから別clinicの個人情報を返し得る。
- AUD-001の「予約へOwner/Petを補完する経路」とは異なり、カルテ本体のFK分離問題である。

### 主な根拠

- `backend/internal/service/medical_record_crud.go` — `Create`, `Update`
- `backend/internal/service/medical_record_builders.go` — Owner/Petフィールド構築
- `backend/internal/repository/medical_record_repository.go` — Owner/Pet Preload

### 修正スコープ

- Appointment有無を問わず Create/Update で `validateReservationOwnerPetLinks`（AUD-001再利用）により Owner/Pet clinic所有と Owner-Pet 整合を検証する。
- Update は PATCH 後の最終 Owner/Pet 状態で検証する。
- 検証と MedicalRecord write を `Transactor.WithTx` + repo `dbOrTx` の同一 transaction に含める。
- MedicalRecord の Owner/Pet Preload に clinic 条件を追加する。
- 初回修正では DB migration、APIフィールド変更、他AUDの修正を行わない。

### 受入条件

- [x] Appointment有無にかかわらず別clinicのOwner/Petを指定したCreate/Updateを拒否する。
- [x] Owner-Pet不一致を拒否する。
- [x] 検証とMedicalRecord writeを同一transactionへ含める。
- [x] 汚染FKを持つカルテから別clinicのOwner/PetをPreloadしない。
- [x] 異院拒否時にカルテ・予約のどちらも変化しない。

### リスク

- `reservationRepo == nil`（ユニットテスト用）のとき検証はスキップされる。本番DIでは常に注入される。
- Appointment連携の予約補完とカルテCreateを同一txにしたことで部分成功リスクは減るが、監査ログは引き続きtx外best-effort。

---
## 修正履歴

| 日付 | ID | 変更 | 検証 | 状態 |
|---|---|---|---|---|
| 2026-07-14 | AUD-001〜008 | Mode 3 再検証: 台帳 Closed を鵜呑みにせず、受入条件ごとに実装+Named/isolation テストで再判定。全件 PASS。任意後続（受入外）: CareLog/StaffNote の WithTx 非対称、Update 未変更時の汚染行再検証なし、FE 内部変数名 diagnosis2CategoryId | 静的 grep + scoped `go test`（service/repository/handler: CrossClinic・DailyRecord・ClinicalPlan・Estimate AuthStaffIDWins・vital tx atomicity 等）PASS | Closed |
| 2026-07-14 | AUD-005 | 台帳再クローズ: 実装は e5a571f6 済み。clinic-isolation-auditor Approve CRITICAL=0 HIGH=0。受入条件チェック完了 | scoped Estimate CrossClinic/CreateEstimate/Owner Preload tests PASS | Closed |
| 2026-07-14 | AUD-004 | 台帳再クローズ: 実装は 6ee2e419 済み。clinic-isolation-auditor Approve CRITICAL=0 HIGH=0。受入条件チェック完了 | scoped Hospitalization CrossClinic/Discharge/Owner Pet Preload tests PASS | Closed |
| 2026-07-14 | AUD-006 | AddVitalRecordへ親入院PetID設定+WithTxでFindOrCreate/CreateVital同一tx化。dbOrTx参加。AUD-003 loadHospitalization再利用。migrationなし | `go test ./internal/service/ -run TestDailyRecord` PASS; `go test ./internal/repository/ -run 'TestDailyRecord|TestDBOrTxInventory'` PASS | Closed |
| 2026-07-14 | AUD-007 | Request/Service/Model/Response/Preload/OpenAPI/FEを diagnosis_2_type_id に統一。nullableUint64 で nullクリア。第2 type↔name 整合。CreateSubRecords 誤列名修正。migrationなし | `go test ./internal/service/ -run 'TestBuildClinicalPlanUpdate|TestClinicalPlanService_|TestMedicalRecordService_CreateSubRecords|RejectsCrossClinicDiagnosis|TestMasterFKWriteInventory'` PASS; `go test ./internal/handler/ -run 'TestUpdateClinicalPlan|TestToClinicalPlanResponse|TestNullableUint64'` PASS; `go test ./internal/repository/ -run 'TestClinicalPlanRepository_|TestPreloadClinicScope'` PASS | Closed |
| 2026-07-14 | AUD-002 | Mode 3 write-path完了: validateAccountingRelatedFKs + MR/Hosp/Reservation DI + Named CrossClinic Create/Update + CompletedValidatesInsideTx。誤Closed再オープン後に証拠付き再クローズ。Preload未再作業 | `go test ./internal/service/ -run 'TestAccountingService_.*(CrossClinic|Reject|FK)'` PASS（Create/Update RejectsCrossClinic* 実行）; `-run CompletedValidatesInsideTx` PASS; `go build ./internal/service/` PASS; go-reviewer CRITICAL=0 HIGH=0 | Closed |
| 2026-07-14 | AUD-002 | Mode 3: 誤Closedを再オープン。write-path FK検証・DI・Named CrossClinicテストがコード上未存在（Preload clinic_idは実装済み・対象外） | 静的監査: validateAccountingRelatedFKs 未定義、NewAccountingService に MR/Hosp/Reservation 未注入、CrossClinic Create/Update テスト 0件 | Open |
| 2026-07-14 | AUD-002 | 会計Create/Updateへ関連FK clinic所有確認+MR/Hosp-Owner/Pet相互整合。Owner/PetはvalidateReservationOwnerPetLinks再利用。Updateは最終FK検証。completed CreateはWithTx内検証。Owner/Pet Preloadへclinic条件。変更: accounting_service*.go, accounting_repository.go(+unpaid), service.go DI, isolation tests。migrationなし | `go test ./internal/service/ -run 'TestAccounting.*Clinic|TestAccountingService.*Cross|TestAccountingService_Create|TestAccountingService_Update'` PASS; `go test ./internal/repository/ -run TestAccountingRepository_` PASS; clinic-isolation/security/go/healthcare Approve（CRITICAL/HIGHなし） | Closed |
| 2026-07-14 | AUD-003 | DailyRecord全経路へ親Hospitalization clinic所有確認。GETをGetByDate(Findのみ)に分離しview書込み除去。OpenAPI getDailyRecordを404契約に更新。FE 404→create CTAスコープテスト追加。変更: daily_record_service.go(+hospRepo DI), daily_record_handler.go, 関連テスト, openapi.yaml, DailyRecordsTab.test.tsx。migrationなし | `go test ./internal/service/ -run TestDailyRecord` PASS; `go test ./internal/handler/ -run 'Test.*DailyRecord|TestGetDailyRecord|TestCreateDailyRecord'` PASS; `go test ./internal/repository/ -run TestDailyRecord` PASS; `npx vitest run .../DailyRecordsTab.test.tsx` PASS; clinic-isolation/security/go-reviewer Approve（CRITICAL/HIGHなし） | Closed |
| 2026-07-14 | AUD-008 | MedicalRecord Create/UpdateへOwner/Pet clinic所有確認+最終Owner-Pet整合。WithTx+dbOrTxで検証とwriteを同一tx化。Owner/Pet Preloadへclinic条件。findMedicalRecordByIDもdbOrTx化（tx内read-your-writes）。変更: medical_record_crud.go, medical_record_service.go, medical_record_repository.go, service.go DI, isolation tests | `go test ./internal/service/ -run TestMedicalRecord` PASS; `go test ./internal/repository/ -run 'TestMedicalRecordRepository_|TestDBOrTxInventory_MatchesAllowlist'` PASS; clinic-isolation/security Approve（CRITICAL/HIGHなし）。go-reviewer CRITICAL（Update後FindByIDのtx非参加）を修正済み | Closed |
| 2026-07-14 | AUD-001 | トリミング3書込み経路へPet clinic所有確認とtransaction/lockを追加。管理予約Createをambient transactionへ参加。カルテからの予約Owner/Pet補完を検証。LINE顧客のOwner.Petsをclinic scopeし、汚染関係をfail-closed化 | Service/Repository scoped test、`go vet`、`go build`、`gofmt -l` PASS。clinic-isolation/security/GoレビューはいずれもApprove、CRITICAL/HIGHなし | Closed |
| 2026-07-14 | AUD-008 | AUD-001再監査中に、AppointmentなしMedicalRecord本体のOwner/Pet clinic分離漏れを新規登録 | 静的監査。実装は未着手 | Open |
| 2026-07-14 | AUD-001 | 再監査でトリミング予約のPet clinic所有確認漏れ3経路を検出し、修正を再開 | TDD実施前 | In Progress |
| 2026-07-14 | AUD-001 | Service境界でOwner/Pet/LineCustomer clinic所有確認+Owner-Pet整合。Preloadへclinic条件。Assert*はdbOrTx。Updateは最終状態検証+tx内Lock。変更: reservation_service.go, appointment_admin_service.go, reservation_repository.go, appointment_admin_repository.go, 関連テスト/モック | `go test ./internal/service/ -run 'TestReservationService_(Create\|Update)\|TestReservationAdminService_Create'` PASS; `go test ./internal/repository/ -run 'TestReservationRepository_(FindAll\|FindByID_ClinicIsolation)\|TestPreloadClinicScope\|AssertOwnerPet\|DoesNotPreload'` PASS; `TestMasterFKWriteInventory` PASS; `TestDBOrTxInventory` PASS。tdd/go/security/clinic-isolation レビューのCRITICAL/HIGH解消 | Closed |
| 2026-07-13 | AUD-001〜007 | 初回監査と独立再検証結果を記録 | 静的監査。動的異院データ再現なし | Open |
