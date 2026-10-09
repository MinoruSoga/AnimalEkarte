# U02-C04 — clinic スコープ/コンテキスト導出 (clinic scope/context derivation)

Campaign: emr-332-security-defense / Unit: U02-AUTHN-AUTHZ-CLINIC / Attempt: att-u02-1
Format: `docs/work/emr-332-security/baseline/case-verdict-template.md` の 12 フィールド構成。

## ケース判定レコード

```text
ケース ID:            C04
対象リビジョン:        de28e9765a29091339a872a85f1875a947e07e5b
実施方式 (method):     static + simulated
正常条件:
  期待 — request context の clinic_id/clinic_ids/is_system_admin は
         JWT claims ではなく CurrentAccessResolver が DB から再解決した
         現在の active authority からのみ導出される。staff は active
         assignment の clinic のみ、system_admin は active clinic
         inventory のみ。
  観測 — TestAuth_HealthyRequestUsesOnlyCurrentAccessState PASS、
         TestAuth_CurrentSystemAdminClinicIdentityReplacesStaleTokenDefault
         PASS、TestCurrentAccessResolver_UsesCurrentAccountAndActive
         Assignments PASS、TestCurrentAccessResolver_RegularStaffUses
         ScopedActiveIDsWithoutListClinics PASS、TestCurrentAccessResolver_
         GraphPathRegularStaffUsesJoinedClinicActivity PASS
         (evidence-t01, t04)。static: auth.go:86-149 で context は
         currentAccess.ClinicIDs/IsSystemAdmin/MainClinicID 由来
         (:106-148) — JWT 内の ClinicID/ClinicIDs/IsSystemAdmin は
         requestClaims に clone された後、live authority で上書き
         される (:98-130)。X-Clinic-ID 切替も claims.ClinicIDs
         (=live authority) の subset 強制 (:363-371)。
拒否条件:
  期待 — 現在の authority 外の clinic (削除済み assignment, inactive
         clinic, 未割当 X-Clinic-ID, 権威なき clinic_ids 指定) は
         403/401 で拒否。authority 解決不能・空 authority も fail-closed。
  観測 — TestAuth_RejectsRemovedDefaultAndHeaderClinicAssignments
         (removed signed default / removed header / no active
         assignments) PASS、TestAuth_RegularStaffRejectsInactiveCurrentClinic
         (inactive header rejected, stale signed default→live main
         fallback) PASS、TestAuth_CurrentAdminStateControlsGlobalBypass
         (admin demotion / outside-authority clinic reject / zero clinic
         reject) PASS、TestAuth_PO005SystemAdminFallbackFailsClosed PASS、
         TestAuth_InvalidClinicIDHeader / TestAuth_NotAssignedClinicIDHeader
         PASS、TestCurrentAccessResolver_GraphPathStaffWithoutAssignments
         FailsClosed / _GraphPathFailsClosedWithoutClinicAuthority /
         _RegularStaffInactiveOnlyIsForbidden /
         _RegularStaffUsesOnlyActiveClinicInventory /
         _RegularStaffClinicAuthorityFailsClosed /
         _SystemAdminClinicAuthorityFailsClosed すべて PASS
         (evidence-t01, t04)。static: resolveClinicID は
         header/default が claims.ClinicIDs subset でなければ 403
         (CodeClinicSelectionUnavailable) (:334-371)、staff_id 非数値/
         0 → 401 (:194-197)。
防御判定:            PASS (static+simulated の範囲)
副作用呼出回数:       0 — 拒否は context 設定/c.Next() 前。clinic 切替
                     audit + prev_clinic_id cookie は許可時のみ
                     (:374-406)。実測 write 副作用は未検証 (simulated)。
検出判定:            PASS (simulated 範囲) — clinic 切替時 audit log
                     (best-effort, :386-406; TestAuth_ClinicSwitch_
                     AuditLog)、system_admin clinic authority 不正は
                     slog.ErrorContext (:121-129)、resolver 失敗は
                     WARN/ERROR + notifier (:244-284)。403 応答自体が
                     拒否証跡。
安全な証跡参照:       backend/internal/middleware/auth.go:86-152, :185-286,
                     :325-408 / backend/internal/auth/current_access_service.go
                     :74-213 (Resolve), :131-201 (loadCurrentAccessIdentity→
                     assignments→clinic access) /
                     backend/internal/httpapi/context.go:153-194, :202-218 /
                     docs/work/emr-332-security/identity-access/evidence/auth-middleware.txt /
                     docs/work/emr-332-security/identity-access/evidence/current-access.txt /
                     docs/work/emr-332-security/identity-access/evidence/domain-membership.txt
制約・未検証範囲:     staging/production 実測は BLOCKED。resolver → DB の
                     実 query 結果・cache 挙動 (CachedCurrentAccessResolver
                     TTL/error non-cache — cache test は memory-only で
                     PASS だが prod redis/DB 未実測)。repository 層の
                     ClinicScope/ClinicScopeIn の DB 実効果は TestRealDB_*
                     でのみ検証可能 → 本ユニットでは BLOCKED
                     (evidence-t07 で db 未到達の FAIL は製品不具合ではなく
                     実行制約)。未検証: cross-clinic list の実データ行数、
                     inactive clinic が実 DB で不活性化した直後の race。
修復案:              none applied — ギャップ証拠なし。
再検証結果:          修復なしのため未実施。simulated 再実行全 PASS。
```

## clinic 導出チェーン (static)

1. JWT claims の `UserID` を `resolveCurrentAccess` へ (`auth.go:86-94`)。claims の `ClinicID`/`ClinicIDs`/`IsSystemAdmin` はこの時点では信頼しない。
2. `CurrentAccessResolver.Resolve` が現在有効な staff・account・active assignment・active clinic inventory を DB から解決 (`current_access_service.go:131-211`)。staff 不在/無効・account 不整合・assignment/clinic 読取失敗・authority 空は全て error 返却で fail-closed (`:136-201`)。
3. 成功時のみ `requestClaims.IsSystemAdmin = currentAccess.IsSystemAdmin`、`requestClaims.ClinicIDs = currentAccess.ClinicIDs` を clone した JWT claims に上書き (`auth.go:98-110`)。system_admin 経路は `currentAccess.MainClinicID` が ClinicIDs に含まれることを再検証、含まれなければ 503 (:111-130)。
4. `TokenMatchesAccountEpoch` で現在 account epoch と JWT epoch を照合、不一致 → 401 (`:99-105`, `token_service.go:318-329`)。
5. `resolveClinicID` が `X-Clinic-ID`/signed default を `claims.ClinicIDs` (= live authority) subset として検証。外なら 403、数値不正は 400、0 は 400/403 (:325-371)。
6. 最終的に `clinic_id` (選択医院), `clinic_ids` (権威集合), `is_system_admin`, `user_id`, `CurrentAccess` を context に設定 (`:143-150`)。
7. handler 層は `ExtractClinicID`/`AuthorizeClinicIDs`/`ResolveListClinicIDs`/`ResolveAllClinicIDs` (httpapi/context.go:107-218) と `*ForPermission` 系 (clinic_permission.go:72-173) で宛先 clinic の membership+grant を再強制。repository 層は `ClinicScope`/`ClinicScopeIn`/JOIN 親テーブル clinic 述語で DB スコープ (U01 §4 参照)。

## repository predicate サンプル (A4: >=3 domain packages, read/write/delete/count + background)

`clinic_id` は必ず middleware context (`ExtractClinicID`/`ExtractClinicIDs`/`Resolve*ClinicIDs`) から service→repository へ引数で渡され、ユーザー入力 body/path から直接 repository スコープを構成しない。サンプル grep 証跡 (read は evidence-t05 の MembershipAB 系と対応):

- **owner** (`internal/owner/repository.go`): read `FindAll` :105 `ClinicScopeIn(clinicIDs)` + preload `Pets`/`Pets.Insurance`/`Pets.Owner` に `clinic_id IN ?` (:140-181); `findOwnerByIDWithDB` :181; write `UpdateAndFindApplying` :335-394 (ClinicScope + ambient tx)、`CreateWithPets` :236; delete `Delete` :414-424 → `persistence.DeleteScopedByID(clinicID, id)`; count `CountPetsByOwnerID` :426-441 (ClinicScope); lock `LockByIDForUpdate` :319-334 / `LockForDelete` :442-462 (ClinicScope + ambient tx 必須で fail-closed)。
- **pet** (`internal/pet/repository.go`): read `pets.clinic_id IN ?` :142/:169, `ClinicScopeIn` :301; **JOIN 相関** `owners.clinic_id = pets.clinic_id AND owners.clinic_id IN ?` :165-171 (BUG-454: 認可集合に両院があっても owner/pet の clinic 一致を強制), `INNER JOIN owners ... owners.clinic_id = pets.clinic_id` + `pets.clinic_id = ?` :269-272; write/update/delete 系 :339-543 は `ClinicScope(clinicID)` 適用。
- **medicalrecord** (`internal/medicalrecord/`): addendum `JOIN medical_records ... medical_records.clinic_id = medical_record_addenda.clinic_id` + `medical_record_addenda.clinic_id = ?` (medical_record_addendum_repository.go:41-54); image `medical_records.clinic_id = ?` + sub-select `(SELECT id FROM medical_records WHERE clinic_id = ? AND deleted_at IS NULL)` (medical_record_image_repository.go:43-67); owner_visit `JOIN pets AS current_owner_pet ... clinic_id = mr.clinic_id` + `JOIN owners ... o.clinic_id = mr.clinic_id` + `mr.clinic_id = ?` (medical_record_owner_visit_repository.go:79-81)。
- **reservation** (`internal/reservation/line_reservation_setting_handler.go:25-33`): `pathAuthorizedClinicID` が path param `clinic_id` を `AuthorizeClinicIDsForPermission` で宛先 clinic grant+membership 検査 — path 指定医院 B で grant 無しは拒否 (TestGetLineReservationSetting_MembershipABGrantASelectedB/rejects_path_clinic_B_without_hospital-settings:view PASS, evidence-t05)。
- **background/batch** (`cmd/api/batch_scheduler.go:102-158`): scheduled jobs は `RunNoShowCheckAllClinicsAt` / `RunDeliveryTriggerBatchAllClinicsAt` / `RunDormantDetectionAllClinicsAt` の設計上全医院走査 — これは認証コンテキストを持たない worker-coordinator 内部経路 (`X-Scheduler-Token` 定数時間比較 + Worker edge が外部 404; :68-99) であり、リクエスト起因の clinic コンテキスト導出の対象外。ジョブ内部の各医院処理は医院単位でイテレートする設計 (all-clinic by name)。
- **clinic scope helper**: `persistence.ClinicScope`/`ClinicScopeIn`/`DeleteScopedByID` が統一スコーププリミティブ (`internal/persistence/`)。handler から user-input clinic_id が直接 repository に渡る経路は grep で確認できず (全て context 由来)。

## 検証コマンド

```sh
docker run --rm --pull never --network none \
  -v <worktree>/backend:/app:ro -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-go-build-cache:/root/.cache/go-build -w /app golang:1.26-alpine \
  go test -count=1 -v -run 'TestCurrentAccessResolver_|TestCurrentAccessStaffReader_FailsClosedWithoutDatabase|TestStaffLookupError' ./internal/auth/
# => PASS / ok .../internal/auth 0.016s (identity-access/evidence/current-access.txt)
# (middleware 側は evidence-t01, domain は evidence-t05 を参照)
```
