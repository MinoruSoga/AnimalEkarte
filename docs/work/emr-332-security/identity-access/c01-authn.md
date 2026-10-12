# U02-C01 — 保護経路の認証要求 (authentication required)

Campaign: emr-332-security-defense / Unit: U02-AUTHN-AUTHZ-CLINIC / Attempt: att-u02-1
Format: `docs/work/emr-332-security/baseline/case-verdict-template.md` の 12 フィールド構成。

## ケース判定レコード

```text
ケース ID:            C01
対象リビジョン:        de28e9765a29091339a872a85f1875a947e07e5b (U01 safety record §1 の target revision)
実施方式 (method):     static + simulated (staging-observed は未実施 — 下記「制約」)
正常条件:
  期待 — 有効な access JWT (Cookie または Authorization Bearer) と現在有効な
         staff/account/assignment/clinic authority を持つリクエストのみが
         protected ハンドラへ到達し、context に user_id/is_system_admin/
         clinic_id/clinic_ids/CurrentAccess が設定される。
  観測 — TestAuth/valid_HS256_token_passes_through_with_correct_context_values
         PASS、TestAuth_AccessTokenCookie PASS、TestAuth_AuthTokenCookieFallback
         PASS、TestAuth/Bearer_scheme_is_case-insensitive PASS
         (identity-access/evidence/auth-middleware.txt)。static: auth.go:143-152 で
         コンテキスト設定後に c.Next()。
拒否条件:
  期待 — 認証トークン不在・形式不正・検証失敗・nil claims のリクエストは
         401 で拒否され、ハンドラに到達しない。token service 未構成は
         503 fail-closed。
  観測 — TestAuth/{missing_Authorization_header_returns_401,
         malformed_header_without_Bearer_scheme_returns_401,
         invalid_token_string_returns_401,expired_token_returns_401,
         token_signed_with_wrong_secret_returns_401} PASS、
         TestAuth_FailsClosedWhenTokenVerifierReturnsNilClaims PASS、
         TestAuth_FailsClosedWhenTokenVerifierIsMissing PASS (503 +
         slog.Error)、TestAuth_SigningMethodMismatch PASS
         (identity-access/evidence/auth-middleware.txt)。static: auth.go:66-84
         (401/503 + return — c.Next() 未到達)、auth.go:164-183
         (extractToken は access_token Cookie → auth_token Cookie →
         Bearer 順で抽出)。
防御判定:            PASS (static+simulated の範囲)
副作用呼出回数:       0 — httptest 拒否ケースで handler/recorder 後段呼出
                     なし。拒否時の唯一の副作用は prev_clinic_id cookie 書換
                     経路の前に return するため 0 (auth.go:66-84 の早期 return;
                     clinic 切替 cookie は :390-406 で認証成功後のみ)。
                     notifier 呼出は一時的 staff 障害経路のみ (拒否自体は
                     副作用ではなく検出の一部)。実測 DB/queue/mail は
                     未検証 (simulated のみ)。
検出判定:            PASS (simulated 範囲) — 401 応答自体が拒否証跡。
                     token service 未構成・resolver 未構成・resolver 汎用
                     エラーは slog.ErrorContext で記録 (auth.go:72-76,
                     :200-206, :277-285)。一時的 staff 検証障害は
                     slog.WarnContext + StaffValidationFailureNotifier
                     (auth.go:244-255; TestAuth_StaffValidation が
                     notify 呼出回数=1 を検証)。ログにトークン値を
                     含まないことをソースで確認 (auth.go に tokenStr を
                     ログ出力する行なし)。
安全な証跡参照:       backend/internal/middleware/auth.go:66-153, :164-183,
                     :188-286 / backend/internal/auth/http_routes.go:88-93 /
                     backend/cmd/api/composition_runtime.go:517-537, :551-577 /
                     docs/work/emr-332-security/identity-access/evidence/auth-middleware.txt /
                     docs/work/emr-332-security/identity-access/evidence/route-gates.txt /
                     docs/work/emr-332-security/baseline/route-middleware-map.md §3-§6
制約・未検証範囲:     staging/production 実測は BLOCKED (baseline/safety-record:
                     稼働 revision 未照合・許可テナント未払出・認証済み
                     live check 禁止)。Worker エッジ層の転送動作は static のみ。
                     未検証: 実 HTTP での Cookie 属性送信、負荷下での挙動。
修復案:              none applied — 静的・模擬の範囲で防御ギャップを確認
                     できず (A6: ギャップ証拠なしのため製品コード変更なし)。
再検証結果:          修復なしのため未実施。simulated 再実行は全 PASS
                     (evidence-t01, t06)。
```

## 経路グループ列挙 (A1: route coverage)

`protected` グループは `auth.RegisterRoutes` が `api.Group("") + Use(Authenticate) + Use(CSRF)` で構築し (http_routes.go:88-90)、`composition_runtime.go:517-537` がこれを全ドメイン登録に渡す。U01 マップ §4 の 15 登録元行すべてが `protected` 経由であることを再確認した (masters/permission-groups, owners, pets+sub-owners+chronic-conditions, staff/shift/shift-template masters, clinics/clinic-holidays/company/closing-settings, manual/articles, support, identity-links, inventory, trimmings, medical-records+hospitalizations+lab*, reservations+reservation masters, billing/accountings/cash-register/reports, lstep 系, /me+/users/me/password)。

protected 外に登録される `/api*` surface とその防御 (baseline/route-middleware-map.md §5 と一致):

| Surface | 登録 | 防御 |
|---------|------|------|
| `/api/liff/:clinicId/settings` (GET) | reservation/routes.go | rate_limit のみ (公開設定) |
| `/api/liff/:clinicId/link` (POST) | reservation/routes.go | link_token + line_id_token 自己認証 |
| `/api/liff/:clinicId/*` 認証済みグループ | reservation/routes.go | LiffRateLimit + LiffAuth (liff_auth.go:50-151) |
| `POST /api/line/webhook` | composition_runtime.go:570-577 | rate_limit + HMAC 署名検証 (lstep) |
| `POST|DELETE /api/v1/uat/synthetic-closings*` | composition_runtime.go:568 | allowHTTP 三重ゲート + Bearer 秘密 (synthetic_closing_http.go:44-155) → 不適合は 404。`AllowUATSyntheticClosing` が APP_ENV∈{test,development,local,dev} かつ DB_HOST∈{db,localhost,127.0.0.1} 以外を fail-closed 拒否 (synthetic_closing_env.go:39-77) |
| `POST /_internal/scheduled-jobs/:jobAction` | batch_scheduler.go:68-81 | requireSchedulerInternalToken 定数時間比較・未設定 fail-closed 401 (batch_scheduler.go:84-99) + Worker 層が外部 404 |
| `GET /health*` `/uploads/*` | base_routes.go | 公開 (uploads は s3 時非マウント) |

## 検証コマンド (verbatim 出力は証跡ファイル)

```sh
docker run --rm --pull never --network none \
  -v <worktree>/backend:/app:ro -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-go-build-cache:/root/.cache/go-build -w /app golang:1.26-alpine \
  go test -count=1 -v -run 'TestAuth|TestLiffAuth|TestVerifyLiffIDToken|TestLiffChannelIDExtraction|TestExtractLiffCustomerID' ./internal/middleware/...
# => PASS / ok github.com/animal-ekarte/backend/internal/middleware 0.019s
#    (identity-access/evidence/auth-middleware.txt, exit 0)

docker run --rm --pull never --network none \
  -v <worktree>/backend:/app:ro -v ekarte-go-mod-cache:/go/pkg/mod \
  -v ekarte-go-build-cache:/root/.cache/go-build -w /app golang:1.26-alpine \
  go test -count=1 -v -run 'TestGETHEAD|TestRouteCompositionSmoke' ./cmd/api/
# => PASS / ok github.com/animal-ekarte/backend/cmd/api 0.210s
#    (identity-access/evidence/route-gates.txt, exit 0 — TestGETHEADInventoryRejectsRouteDrift
#     が登録経路と inventory の乖離をゲート, TestRouteCompositionSmoke_
#     TargetGraphRegistersEverySurface が全 surface 登録をゲート)
```
