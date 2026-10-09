# U02-C17 — 閉鎖失敗 (fail-closed rejection)

Campaign: emr-332-security-defense / Unit: U02-AUTHN-AUTHZ-CLINIC / Attempt: att-u02-1
Format: `docs/work/emr-332-security/baseline/case-verdict-template.md` の 12 フィールド構成。

## ケース判定レコード

```text
ケース ID:            C17
対象リビジョン:        de28e9765a29091339a872a85f1875a947e07e5b
実施方式 (method):     static + simulated
正常条件:
  期待 — 依存 (token service / current access resolver / staff notifier /
         clinic inventory / permission checker / scheduler secret / LIFF
         upstream) が健全な時のみ処理が進み、正常系で 200 を返す。
  観測 — evidence-t01/t03/t04/t05 で正常系サブテスト全 PASS
         (TestAuth/valid_HS256..., TestLiffAuth/正常系,
         TestHTTPHandler_RequirePermission 正常系,
         TestCurrentAccessResolver_UsesCurrentAccountAndActiveAssignments,
         MembershipAB* の許可系)。
拒否条件:
  期待 — 以下の全 fail-closed トリガーが 401/403/503/500 で拒否され、
         ハンドラ・後段・副作用に到達しない:
         (a) token 不在/不正/期限切れ/alg 違い/purpose 不一致/claims 欠落
         (b) token service 未構成 / 検証器 nil claims
         (c) staff identity 非数値/0 / staff 非存在・非アクティブ
         (d) resolver 未構成 / resolve エラー (一時 DB 障害, 汎用 error,
             not-found 以外の StaffLookupError)
         (e) staff 検証一時障害 (bad conn / SQLSTATE 08xxx,53300,57P0x)
             → notify + 503 (通知失敗でも拒否は維持)
         (f) 現在 authority 空 / epoch 不一致 / invalid clinic identity /
             X-Clinic-ID 未割当
         (g) permission 依存欠落 (EffectivePermissions nil, checker
             未設定, clinicID==0) / repo error
         (h) scheduler 共有秘密未設定・不一致
         (i) LIFF: header 不在/形式不正, clinicId 非数値/0, 設定 DB
             エラー, LIFF 未構成, LINE verify error, sub 空, customer
             解決失敗, mock モード lookup 失敗
         (j) UAT surface: env/DB host/HTTP host/bearer 不適合
  観測 — 上記すべてを PASS で確認:
         (a,b,f) evidence-t01: TestAuth/*401/403*, TestAuth_FailsClosed
         WhenTokenVerifier{IsMissing,ReturnsNilClaims}, TestAuth_
         RejectsAccountEpochMismatchAndResolverFailure, TestAuth_
         StaffValidationConfigurationFailsClosed (nil validator→503,
         malformed staff_id→401), TestAuth_RejectsRemovedDefaultAnd
         HeaderClinicAssignments, TestAuth_CurrentAdminStateControls
         GlobalBypass, TestAuth_InvalidClinicIDHeader / NotAssigned
         ClinicIDHeader, TestAuth_RejectsRefreshTokenSubject,
         TestAuth_SigningMethodMismatch
         (c,d,e) evidence-t01: TestAuth_StaffValidation (inactive/
         missing→deny, bad connection→notify×1+deny, cannot-connect-now
         →deny, generic repo error / SQLSTATE 42601→deny w/o notify,
         notifier 失敗でも deny), TestAuth_AccountAndAssignmentFailures
         NeverUsePO005 (account/assignment read error→deny),
         TestAuth_PO005SystemAdminFallbackFailsClosed
         evidence-t04: TestCurrentAccessResolver_*FailsClosed*,
         _RegularStaffInactiveOnlyIsForbidden, _GraphPathStaffWithout
         AssignmentsFailsClosed, TestCurrentAccessStaffReader_FailsClosed
         WithoutDatabase, TestStaffLookupError_PreservesCause
         (g) evidence-t03: TestHTTPHandler_HasPermission_RepositoryError
         FailsClosed, _MissingContextDoesNotWriteResponse,
         HasPermissionInClinic_*; static: http_permission.go:69-85,
         clinic_permission.go:39-46, :87-93
         (h) static: batch_scheduler.go:84-99 — expected 空は常に 401、
             長さ不一致/値不一致は定数時間比較で 401。unknown IRoutes
             shape は登録拒否 (:72-80)
         (i) evidence-t01: TestLiffAuth 全 11 サブテスト PASS +
             TestLiffAuth_MockModeLookupFailureIsFailClosed +
             _MockModeInvalidClinicID; static: liff_auth.go:59,75,87,92,
             101,113,117,129,136
         (j) static: synthetic_closing_http.go:55-59,112-155,
             synthetic_closing_env.go:39-77 — 不適合は 404
防御判定:            PASS (static+simulated の範囲)
副作用呼出回数:       0 — すべての拒否経路は c.Next()/handler 未到達。
                     StaffValidation の notifier は「検出」であり副作用
                     ではない。audit write は best-effort かつ許可時のみ。
                     TestAuth_ClinicSwitch_AuditLog の
                     continues_request_even_when_audit_fails は許可後の
                     audit 失敗許容であり拒否経路の副作用ではない。
                     実測 write/queue/mail 呼出回数は未検証。
検出判定:            PASS (simulated 範囲) — fail-closed 経路の slog
                     Warn/Error ("access token validation unavailable",
                     "current access validation failed closed",
                     "failed to verify staff validity; denying request
                     (fail-closed)", "resolver is not configured",
                     "LIFF_MOCK customer lookup failed") が evidence-t01
                     に記録されている。notifier 呼出も検証。ただし
                     401/403 の一般拒否はログ非対象 — 応答コードが証跡。
安全な証跡参照:       backend/internal/middleware/auth.go:66-153, :188-319 /
                     backend/internal/auth/token_service.go:77-93, :227-293 /
                     backend/internal/auth/current_access_service.go:131-211 /
                     backend/internal/auth/http_permission.go:64-220 /
                     backend/internal/httpapi/clinic_permission.go:39-146 /
                     backend/internal/middleware/liff_auth.go:50-151 /
                     backend/cmd/api/batch_scheduler.go:68-99 /
                     backend/internal/billing/synthetic_closing_http.go:44-155 /
                     backend/internal/billing/synthetic_closing_env.go:39-77 /
                     docs/work/emr-332-security/identity-access/evidence/auth-middleware.txt /
                     docs/work/emr-332-security/identity-access/evidence/http-permissions.txt /
                     docs/work/emr-332-security/identity-access/evidence/current-access.txt
制約・未検証範囲:     staging/production 実測・実 DB/redis 障害注入は
                     BLOCKED。simulated は mock/stub による依存注入で
                     検証。panic recovery・network timeout・外部 LINE API
                     実接続の挙動は未検証。worker/edge 層の fail-close
                     は U01 記録のみ (本ユニットでは static 引用)。
修復案:              none applied — ギャップ証拠なし。
再検証結果:          修復なしのため未実施。simulated 再実行全 PASS。
```

## 備考

- `isTemporaryStaffValidationError` (auth.go:288-319) は型付き DB 可用性エラー (driver.ErrBadConn, pgconn.ErrConnClosed, SQLSTATE 08xxx/53300/57P01-03) のみ「一時的」とみなし notifier を呼ぶが、**拒否自体は fail-closed で変わらない**。programming error (SQLSTATE 42601) や汎用 repo error は notify せず 503 で拒否 — TestAuth_StaffValidation が両分岐を検証。
- `rejectResolveCurrentAccessError` (auth.go:230-286) は not-found→403 / 一時障害→notify+503 / Forbidden→403 (ClinicSelectionUnavailable code 付き分岐) / Unauthorized→401 / その他→503 を厳密に分岐。JWT 権限を継続利用する経路は存在しない。
- `registerScheduledJobRoutes` (batch_scheduler.go:68-81) は `*gin.Engine` / `*gin.RouterGroup` 以外の IRoutes 実装を拒否 (fail-closed 登録)。
- LIFF mock モード (`LIFF_MOCK`) でも lookup 失敗は 503 fail-closed で、Authorization ヘッダ省略による bypass は mock 時のみの明示的挙動 (liff_auth.go:59-85, TestLiffAuth_MockModeLookupFailureIsFailClosed/_MockModeInvalidClinicID)。本番パスは header 不在→401 (:87)。

## 検証コマンド

上記各証跡ファイルに対応するコマンドを c01-c04 の verdict に記載。追加なし。
