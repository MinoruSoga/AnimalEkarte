# U02-C03 — 同一テナント内認可 (same-tenant role/permission authorization)

Campaign: emr-332-security-defense / Unit: U02-AUTHN-AUTHZ-CLINIC / Attempt: att-u02-1
Format: `docs/work/emr-332-security/baseline/case-verdict-template.md` の 12 フィールド構成。

## ケース判定レコード

```text
ケース ID:            C03
対象リビジョン:        de28e9765a29091339a872a85f1875a947e07e5b
実施方式 (method):     static + simulated
正常条件:
  期待 — 選択/宛先 clinic での resource:action grant を持つ staff が
         許可され、context の権限判定が宛先 clinic ごとに評価される。
         membership AB + grant B で selected B は許可。
  観測 — TestHTTPHandler_RequirePermission_AttachesClinicChecker PASS、
         TestHTTPHandler_RequirePermissionAllowingAssignedClinicGrant_
         AllowsGETWhenAnotherAuthorizedClinicHasGrant PASS、
         TestListStaffs_/TestGetStaff_/TestGetStaffPermissionGroups_
         MembershipABGrantBSelectedB PASS、TestListMedicalRecords_
         MembershipABGrantsAB (clinic_ids=1,2 横断 / 未指定で選択医院
         のみ) PASS、TestGetLineReservationSetting_MembershipABGrantASelectedB
         /allows_path_clinic_A PASS (evidence-t03, t05)。
         static: RequirePermission は選択 clinic の grant を要求
         (http_permission.go:132-151)、GET/HEAD の assigned-clinic
         fallback は opt-in (RequirePermissionAllowingAssignedClinicGrant
         :124-130, isSafeHTTPMethod GET/HEAD のみ :195-205) で、宛先
         clinic 再検査は handler 側が Filter/Authorize で実施。
拒否条件:
  期待 — grant を持たない宛先 clinic・context 欠落・依存欠落・repo エラー
         で拒否 (fail-closed)。GET/HEAD でも他 clinic の grant では
         選択 clinic 要求を通さない。
  観測 — TestHTTPHandler_HasPermission_RepositoryErrorFailsClosed PASS、
         TestHTTPHandler_HasPermission_MissingContextDoesNotWriteResponse
         PASS、TestHTTPHandler_HasPermissionInClinic_UsesDestinationClinic
         PASS、TestHTTPHandler_RequirePermission_RejectsGETWhenOnlyAnother
         ClinicHasGrant PASS、TestGetLineReservationSetting_
         MembershipABGrantASelectedB/rejects_path_clinic_B_without_
         hospital-settings:view PASS、MembershipABGrantASelectedB 系
         "rejects_B_without_grant"/"rejects_explicit_B_clinic_ids" が
         owner/pet/medicalrecord/reservation/clinic/staff で PASS
         (evidence-t03, t05)。static: HasPermissionInClinic は
         is_system_admin 欠落→false (:69-72)、staffID 欠落/clinicID==0/
         EffectivePermissions nil → false (:77-80)、rules 解決失敗→false
         (:81-84)。requireClinicPermissionChecker 未設定は 403
         (clinic_permission.go:39-46)。
防御判定:            PASS (static+simulated の範囲)
副作用呼出回数:       0 — 403 拒否は requirePermission 内で c.Abort() する
                     ため handler/保存呼出に到達しない (http_permission.go
                     :148-150)。permission memo は request-local map のみ
                     (:34-58)。実測の write/audit 呼出回数は未検証。
検出判定:            PASS (simulated 範囲) — 403 応答が拒否証跡。
                     deny 専用ログは存在しないが 401/403 応答 + request
                     logging (global RequestLoggingMiddleware) が相関記録。
                     TestAuth_StaffValidation 系で通知 (notifier) 呼出回数
                     1 を検証。fail-closed 経路の ERROR/WARN ログが
                     evidence-t01 で確認済み。
安全な証跡参照:       backend/internal/auth/http_permission.go:17-22,
                     :34-58, :64-86, :111-151, :195-220 /
                     backend/internal/httpapi/clinic_permission.go:39-94,
                     :99-146, :150-173 / backend/internal/httpapi/
                     context.go:153-165, :175-194, :202-218 /
                     docs/work/emr-332-security/identity-access/evidence/http-permissions.txt /
                     docs/work/emr-332-security/identity-access/evidence/domain-membership.txt /
                     docs/work/emr-332-security/identity-access/evidence/route-gates.txt
制約・未検証範囲:     staging/production 実測・マルチロール live 検証は
                     BLOCKED (専用テストアカウント未払出 —
                     baseline/safety-record)。役割×リソース行列は
                     static/simulated で enumerated。DB 実データでの
                     repository フィルタ効果は TestRealDB_* が担当するが
                     本ユニットでは実行不可 (network none で db 未到達 —
                     evidence-t07)。未検証: 全 resource×action 組合せの
                     rules データ内容、object-level (行) 権限の DB 実証。
修復案:              none applied — ギャップ証拠なし。
再検証結果:          修復なしのため未実施。simulated 再実行全 PASS。
```

## 実装された function/object/field レベルの権限構造 (A3 レビュー範囲)

- **function level**: `RequirePermission`/`RequirePermissionAny` が middleware で選択 clinic の grant を強制 (`http_permission.go:120-151, 156-193`)。GET/HEAD の assigned-clinic fallback は `RequirePermissionAllowingAssignedClinicGrant` opt-in のみ (`:124-130` + `isSafeHTTPMethod` GET/HEAD 限定 `:195-205`)。
- **object level**: `HasPermissionInClinic` が宛先 clinic の effective rules を評価 (`:64-86`)。cross-clinic list/detail は handler 側で `FilterClinicIDsForPermission`/`AuthorizeClinicIDsForPermission`/`ResolveListClinicIDsForPermission`/`ResolveAllClinicIDsForPermission` が宛先 clinic ごとに再評価 (`clinic_permission.go:69-173`)。1件でも grant 外/宛先外なら 403 (fail-closed, partial 許可なし)。
- **field level**: permission rule は `rule.CanView/CanCreate/CanEdit/CanDelete` の action 別 boolean で評価 (`permissionRulesAllow` :88-109)。割引系は `RequireDiscountEditFloat`/`RequireDiscountCreateFloat` など action+対象別の個別判定 (`:225-254`)。
- **system_admin**: 明示 bypass (`:73-75`, `clinic_permission.go:80-82, :100-102`) — ただし membership/subset 検査は bypass 後も維持 (`AuthorizeClinicIDsForPermission` は先に `AuthorizeClinicIDs` で所属 subset を強制 `:77-79`; `FilterClinicIDsForPermission` の admin 経路は membership compact 集合を返す `:123-130`)。
- **route composition**: GET/HEAD 分類ゲート (clinic-fixed 159 / cross-clinic 21 / liff 9 / shared-master 9 / public 6 / self 3 / global 3 = 210 route) を `get_head_permission_*_test.go` が機械検査、`TestGETHEADCrossClinicAllowingDoesNotGrantSelectedClinicB` が cross-clinic GET/HEAD の「選択 B では grant B は効かない」を route ごと検証 (evidence-t06)。

## 役割×リソース行列 (simulated — live multi-role staging は BLOCKED)

| 主体 | 選択 clinic | 宛先 clinic | grant | 操作例 | 期待→観測 |
|------|------------|------------|-------|--------|----------|
| regular staff (membership A,B) | A (selected) | A | hospital-settings:view on A | `GET /clinics/A/line-reservation-settings` | 許可→PASS (evidence-t05 reservation) |
| regular staff (membership A,B) | A | B | hospital-settings:view on A のみ | 同上で path=B | 403→PASS (evidence-t05 `rejects_path_clinic_B_without_hospital-settings:view`) |
| regular staff (membership A,B) | A | A,B mixed list | grant on A のみ | `GET /owners?clinic_ids=A,B` | B は filter で除去 or 403→PASS (`filters_mixed_clinic_ids_to_A...` / `rejects_explicit_B_clinic_ids`, owner/pet/medicalrecord/reservation/clinic) |
| regular staff (membership A,B) | A | B | grant on B のみ | `GET /medical-records?clinic_ids=B` while A selected | 宛先 B grant で許可→PASS (`TestListMedicalRecords_MembershipABGrantsAB`) |
| regular staff | A | A | 当該 resource:action grant 無し | 任意 protected write | 403→PASS (`TestHTTPHandler_RequirePermission_*`) |
| system_admin | 任意 active clinic | 同 | (bypass) | `GET /pets` 等 | 許可→PASS (admin bypass 分岐 :73-75; ただし membership subset は維持) |
| 認証なし / context 欠落 | — | — | — | 任意 | 401→PASS (TestAuth missing header) |

（「PASS」は対応する simulated テストが期待どおりの応答を返したこと。live 多役割 staging 検証は専用アカウント未払出のため BLOCKED — 上表は模擬マトリクスであり staging-observed ではない。）

## 検証コマンド

```sh
docker run --rm --pull never --network none \
  -v <worktree>/backend:/app:ro -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-go-build-cache:/root/.cache/go-build -w /app golang:1.26-alpine \
  go test -count=1 -v -run 'TestHTTPHandler_(HasPermission|RequirePermission)' ./internal/auth/
# => PASS / ok .../internal/auth 0.009s (identity-access/evidence/http-permissions.txt)

docker run --rm --pull never --network none \
  -v <worktree>/backend:/app:ro -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-go-build-cache:/root/.cache/go-build -w /app golang:1.26-alpine \
  sh -c "go test -count=1 -v -run 'MembershipAB' ./internal/medicalrecord/ ./internal/reservation/ ./internal/staff/ ./internal/clinic/ && \
         go test -count=1 -v -run '^Test(ListOwners|GetOwner)_MembershipAB' ./internal/owner/ && \
         go test -count=1 -v -run '^Test(ListPets|GetPet|ListOwnerReportPets)_MembershipAB' ./internal/pet/"
# => 全 6 パッケージ PASS (identity-access/evidence/domain-membership.txt)
```
