# U06 — C15 判定レコード: 業務フロー順序 / 冪等性 (order & idempotency)

Campaign: emr-332-security-defense / Unit: U06-DETECT-BUSINESS-AVAIL / Attempt: att-u06-1 / Receiver: receiver-emr332-u06-01
Base doc: docs/work/emr-332-security/u01-case-verdict-template.md (12-field fixed format)

## 12 fields

1. **ケース ID**: `C15`

2. **対象リビジョン**: worktree `AnimalEkarte-emr-332-security-defense-w06` branch `feat/emr-332-security-defense-w06`, HEAD `fc6acc6b8d267eb51050a4848069e1a1b569e7df` + 本ユニットで追加した test-only ファイル (production code 差分なし — 下記「修復案」参照)。

3. **実施方式 (method)**:
   - `static` — 予約 create/update/check-in/no-show/tentative 遷移・medical-record auto-create のソースレビュー (行番号は下記証跡参照)。
   - `simulated` — `internal/reservation/emr332_u06` / `internal/medicalrecord/emr332_u06` の in-memory 単一スレッド test (httptest 不使用・DB 不使用)。`pg_try_advisory_xact_lock` busy は「即時スキップ (acquired=false → return)」としてモデル化 — serialize-and-wait ではない (medical_record_auto_create.go:103-116 と整合)。

4. **正常条件**:
   - 期待: 有効な一般予約 (owner+pet 整合・生存 pet・競合なし・capacity 内) は create を許可し、active tentative があれば新規行ではなく昇格 (reservation_service.go:274-279 `createOrPromoteTentative`)。medical-record auto-create は lock 取得→同日重複なし→文脈検証→appointment 行 lock→INSERT→create 監査→subrecord 確保の順で 1 回だけ作成する。
   - 観測 (simulated): `TestC15_AutoCreateCreatesOnceThenReplaySuppresses` が呼出し列 `tx.begin → AcquireAutoCreateLock → CountByPetAndDate → AssertOwnerInClinic → FindPetOwnerInClinic → BackfillForMedicalRecord → FindByID → AssertOwnerInClinic → FindPetOwnerInClinic → FindPetByIDInClinic → FindByAppointmentID → repo.Create → tx.commit → audit:create → plan.FindByMedicalRecordID → plan.Create` を厳密順序で記録。同一 reservation の replay (count=1) は `Acquire → Count` のみで停止し create/audit 追加 0 件。
   - 観測 (simulated): `TestC15_BookingConstraintEnforcementMatrix` — pending/confirmed/cancelled/no_show/tentative + line/phone/record_intent route は booking constraint enforcement = allow (予約時は実チェックを要求)、checked_in 以降と reception/exam_room/record_shortcut 経由は skip (slot を既に占有)。
   - 観測 (simulated): `TestC15_SlotConflictDeterministicScenarios` / `TestC15_TypeCapacityBoundary` — 空き slot・capacity 内・self-exclude (excludeID 転送) は pass。
   - 観測 (static): create は ambient tx 内で `AcquireBookingLock` (clinic 単位 advisory xact lock) → conflict/capacity チェック → insert/promote (reservation_service.go:358-390, 450-484)。guarded check-in は expected_updated_at CAS + FOR UPDATE 行 lock + 当日/owner/pet 検証 (reservation_guarded_checkin.go:20-26, reservation_repository_slots.go:96-135)。`MarkNoShowAt` は pending/confirmed → no_show の atomic 遷移で finalized カルテ持ちを除外 (reservation_intent_repository.go:279-304)。

5. **拒否条件**:
   - 期待: 順序違反・重複・不整合は副作用ゼロで拒否される — (a) 死亡ペットへの write、(b) 別 owner の pet・他 clinic owner、(c) 満員/競合 slot、(d) terminal status への mutation、(e) auto-create lock busy / 同日既存 / appointment 既存カルテ、(f) appointment binding 不一致、(g) tx 依存欠落。
   - 観測 (simulated): `TestC15_TerminalTrimmingMutationRejected` — completed/cancelled/no_show は Conflict 拒否、他 status は mutable。`TestC15_DeceasedPetWriteRejected` / `TestC15_OwnerPetLinkCrossTenantRejected` — InvalidInput/NotFound 拒否。`TestC15_AutoCreateLockBusySkipsCreate` — lock busy で `AcquireAutoCreateLock` のみ呼ばれ count/create/audit/plan 全て 0。`TestC15_AutoCreateSameDayExistingSkipsCreate` — 同日既存で create 0。`TestC15_AutoCreateExistingAppointmentRecordRepairsWithoutCreate` — appointment 既存カルテで `repo.Create`=0・create 監査=0、best-effort repair (plan.FindByMedicalRecordID) のみ。`TestC15_AutoCreateOrderViolationsRejectBeforeCreate` — 5 つの順序/整合違反が `tx.rollback` で終わり create/audit 0。`TestC15_AutoCreateWithoutTransactionDependencyIsFailClosed` — transactor nil で repo call ゼロ (fail-closed, withTx: medical_record_service.go:183-188)。
   - 観測 (static): tentative reservation は confirm まで medical record 割当不可; `completed`/`no_show` 遷移は専用 workflow 所有で汎用 update が拒否; in_consultation 遷移は medical record 必須 (reservation_service_update.go / reservation_intent_repository.go:469-495)。

6. **防御判定 (defense verdict)**: `PASS` — 記録した static+simulated 範囲で、順序ガード・重複抑止・fail-closed が全て成立し、拒否経路の副作用は実測 0。production code 差分なし。

7. **副作用呼出回数 (side-effect call count)**:
   - 拒否経路 (全ケース): `repo.Create`/`CountByPetAndDate`-後 write/audit/plan = **0** (fake カウンタ実測)。
   - 正常経路 (create 1 回目): repo.Create=1, audit.LogMedicalRecordChange=1, plan.Create=1 (期待値どおり各 1 回)。
   - replay: 追加分 0。
   - test 内に session/監査永続化/キャッシュ/キュー/ファイル/メール/外部テレメトリ実体なし (全 fake)。

8. **検出判定 (detection verdict)**: `PASS` — slog buffer キャプチャで `auto-create lock is busy` / `same-day record already exists` / `auto-create transaction failed` / `failed to acquire auto-create lock` イベントを観測 (medical_record_auto_create.go:105-130, 70-74)。auto-create 成功は `medical record created` + create 監査。lock busy が「スキップ=静寂」ではなく InfoContext で記録されることを確認。

9. **安全な証跡参照 (safe evidence ref)**:
   - Tests: `backend/internal/reservation/emr332_u06/order_idempotency_test.go` (TestC15_* 6 tests / 40 subcases), `backend/internal/medicalrecord/emr332_u06/auto_create_idempotency_test.go` (TestC15_* 8 tests / 20 subcases)。
   - Static anchors: `internal/reservation/reservation_service.go:274-279,358-390,450-484`, `reservation_service_validate.go:72,158-163,223-228`, `reservation_intent_repository.go:469-495,279-304`, `reservation_guarded_checkin.go:20-26`, `internal/medicalrecord/medical_record_auto_create.go:28-95,103-131`, `medical_record_appointment_context.go:14-70,91-133`, `medical_record_crud.go:76-140`, `internal/sharedkernel/pet_not_deceased.go:20-39`, `owner_pet_link.go:21-37`。
   - 実行出力 (verbatim、isolated docker、network none):
     ```text
     ok  	github.com/animal-ekarte/backend/internal/reservation/emr332_u06	0.003s
     ok  	github.com/animal-ekarte/backend/internal/medicalrecord/emr332_u06	0.003s
     ok  	github.com/animal-ekarte/backend/internal/middleware/emr332_u06	0.003s
     ok  	github.com/animal-ekarte/backend/internal/audit/emr332_u06	0.002s
     ```
     (コマンド: `docker run --rm --pull never --network none -v <w06>/backend:/app:ro -v ekarte-go-mod-cache:/go/pkg/mod -v ekarte-go-build-cache:/root/.cache/go-build -w /app golang:1.26-alpine go test -count=1 <4 pkgs>`; -v 全出力は本 attempt セッション内で全件 PASS を確認。)

10. **制約・未検証範囲 (constraints / unverified scope)**:
    - `pg_try_advisory_xact_lock` の実 DB 挙動は未検証 (DB 禁止)。busy ⇒ non-blocking skip の**モデル化**はソース契約 (acquired=false → 即 return, :110-116) に忠実。真の複数プロセス競合は単一スレッド test の範囲外 → `UNKNOWN`。
    - `MarkNoShowAt` FOR UPDATE SQL、`PromoteTentativeOnCheckin`/`AbsorbTentativeOnBooking` の実 SQL、guarded check-in の CAS `UPDATE ... WHERE updated_at = ?` は static 確認のみ (in-memory 再現せず)。
    - reservation handler → `AutoCreateFromReservation` の発火点 (reservation_handler.go:223,265,311,329 / reservation_intent_handler.go:143) は呼出経路の static 確認のみ。
    - 却下済み設計 (同一 pet/日付 unique 制約・clock seam) は backend/CLAUDE.md「却下済み提案」により判定対象外。manual 同時刻重複は正当業務であり欠陥としない。
    - `TestAutoCreateFromReservation_ConcurrentSameClinicPetJSTDayCreatesOneRecord` (既存 internal test) は実 PostgreSQL 前提のため本制約下では未実行 — DB-dependent として記録。

11. **修復案 (remediation)**: なし — owned paths 内で確認された gap は発見されず、product-code fix は未適用。候補 (将来・範囲外): 複数インスタンス環境での booking advisory-lock 到達確認は deploy 形態依存 (本件 C16 と同じ multi-instance 未検証枠)。

12. **再検証結果 (re-verification)**: 未実施 — 適用した fix がないため再検証対象なし。初回実行で全 test PASS (上記 verbatim 出力)。
