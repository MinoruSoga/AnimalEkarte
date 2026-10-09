# U01 — Route × Middleware Coverage Map (ルート×ミドルウェア適用マップ)

Campaign: emr-332-security-defense / Unit: U01-SAFETY-BASELINE / Attempt: att-u01-1
Derivation basis: `feat/emr-332-security-defense` base `de28e9765a29091339a872a85f1875a947e07e5b` のコードを機械的に辿った結果 (推測ではなく登録箇所の列挙)。
Primary sources: `backend/cmd/api/main.go` (newRouter), `backend/cmd/api/composition_auth_routes.go`, `backend/cmd/api/composition_runtime.go`, `backend/cmd/api/base_routes.go`, `backend/cmd/api/batch_scheduler.go`, `backend/internal/*/routes.go` / `handler.go`, `backend/internal/auth/http_routes.go`, `backend/worker/index.ts` + `worker/scheduler-ops.ts`。

凡例 — middleware 列: `auth` (staff JWT 認証), `csrf` (RequireXRequestedWith), `cors`, `rate_limit`, `security_headers`, `logging` (RequestLoggingMiddleware)。`●` = 適用、`-` = 非適用、`△` = 条件付き/経路固有。

## 0. エッジ層 (backend/worker/index.ts) — Gin 到達前の分岐

全受信はまず frontend Worker / API Worker (`animalekarte-stg-api`) に入る。API Worker は以下を **Container (Gin) へ転送せず** Worker 内で処理する:

| Worker 経路 | 処理 | 認証・防御 (Worker 側) |
|-------------|------|------------------------|
| `POST /_internal/migrate` | Container 内 `/app/migrate` exec | `Authorization: Bearer <MIGRATE_RUN_SECRET>` (worker/index.ts:507-508, migrate-exec.ts)。失敗時 fail-closed |
| `/_internal/scheduler/*` (`SCHEDULER_OPS_PREFIX` = `/_internal/scheduler`) | scheduler status/control/runs | dispatch は worker/index.ts:511-529 → `handleSchedulerOpsRequest` が `authenticateSchedulerOpsRequest` を呼ぶ (呼出し箇所 scheduler-ops.ts:315)。認証実装は scheduler-access-auth.ts: まず ops Bearer 共有秘密を `isAuthorizedSchedulerOpsRequest` が `timingSafeEqual` で定数時間比較 (:350-363)、不成立なら `CF-Access-Jwt-Assertion` の Cloudflare Access 署名 JWT を team-domain JWKS / issuer / audience / exp / nbf で検証 (`authenticateSchedulerOpsRequest` :365-441; secrets 宣言 wrangler.jsonc:204-207) |
| その他の `/_internal/*` | — | `isInternalProxyPath` が 404 を返し Container へ転送しない (scheduler-ops.ts:83-107, index.ts:530-535)。`/_internal/scheduled-jobs/*` (Gin 側登録あり) も外部からはここで 404 |
| `OPTIONS *` (プリフライト) | edge 即応答 | `buildCorsPreflightResponse(request, env.CORS_ALLOWED_ORIGIN)` — Container を起こさない (index.ts:540-542) |
| 上記以外全て | Container (Gin) へ転送 | `CF-Connecting-IP` → `X-Forwarded-For` 転記 (無い場合は XFF 削除 — 偽装防止, index.ts:544-556)。`TRUSTED_PROXY_CIDR=10.1.0.0/32` (wrangler vars) で Go 側 proxy 信頼を固定 |

frontend Worker (`animalekarte-stg-frontend`) は静的アセット配信 + `/api/*` を service binding で API Worker へ転送 (frontend/wrangler.jsonc services.API)。

## 1. Gin グローバルチェーン (main.go:203-208)

`newRouter` が全 Gin 到達経路に適用 (登録順):

```go
router.Use(gin.Recovery())                              // panic 回収
router.Use(middleware.SecurityHeaders(cfg.GinMode == "release"))  // security_headers
router.Use(middleware.RequestID())                      // 相関 ID
router.Use(middleware.CORS(cfg.CORSAllowedOrigin))      // cors (exact-origin 照合 + Vary: Origin)
router.Use(middleware.RequestLoggingMiddleware())       // logging (構造化リクエストログ)
router.Use(middleware.SanitizeNullBytes())              // null byte 除去 + raw body 上限
```

**CSRF はグローバルチェーンに含まれない** — 保護グループと公開 auth 経路にのみ個別適用 (§3, §4)。
`release` モードで HSTS が有効 (security_headers.go:9-23)。レートリミットはグローバルではなく経路別。

## 2. 公開ベース経路 (base_routes.go + batch_scheduler.go)

| 経路 | auth | csrf | cors | rate_limit | security_headers | logging | 備考 |
|------|------|------|------|-----------|------------------|---------|------|
| `GET /health`, `GET /api/v1/health` | - | - | ● | - | ● | ● | healthOK (base_routes.go:77-78) |
| `GET /health/db` | - | - | ● | - | ● | ● | DB ping・応答は詳細を含まない (base_routes.go:79, 27-46) |
| `/uploads/*` (StaticFS) | - | - | ● | - | ● | ● | `STORAGE_TYPE != "s3"` の時のみマウント + `uploadsNoSniffHeader` (nosniff)。staging は s3 のため非マウント (base_routes.go:81-87) |
| `POST /_internal/scheduled-jobs/:jobAction` | △ | - | ● | - | ● | ● | `requireSchedulerInternalToken` — `X-Scheduler-Token` 共有秘密の定数時間比較・未設定時 fail-closed 401 (batch_scheduler.go:68-99)。外部到達は Worker が 404 にするため実質 Worker 内部経路 |

## 3. 公開 auth 経路 (`/api/v1` 直下・認証なし) — auth/http_routes.go:53-86

全て Gin グローバルチェーン適用済み (cors/security_headers/logging ●)。

| 経路 | auth | csrf | rate_limit | 備考 |
|------|------|------|-----------|------|
| `POST /api/v1/login` | - | ● | ● (login policy + FailureRateLimit per-account) | http_routes.go:53-58 |
| `POST /api/v1/logout` | - | ● | ● (logout_redirect policy) | http_routes.go:59-64 |
| `POST /api/v1/auth/refresh/logout` | - | ● | ● (logout policy) | http_routes.go:65-70 |
| `POST /api/v1/auth/refresh` | - | ● | ● (refresh policy) | http_routes.go:71-76 |
| `POST /api/v1/auth/forgot-password` | - | **-** | ● (password_reset policy) | http_routes.go:77-81 — CSRF 未適用 (公開パスワード再発行開始)。レート制限のみ |
| `POST /api/v1/auth/reset-password` | - | **-** | ● (password_reset policy) | http_routes.go:82-86 — CSRF 未適用 (token 所持が要件)。C08 評価時に組合せとして記録すること |

## 4. 保護グループ `/api/v1` (認証済み staff API)

構成: `api.Group("")` + `Use(Authenticate)` + `Use(CSRF)` (auth/http_routes.go:88-90)。`Authenticate` = `middleware.AuthWithStaffValidationFailureNotifier` (JWT 検証 + staff/clinic 再解決、composition_auth_routes.go:31-37)、`CSRF` = `middleware.RequireXRequestedWith()` (csrf.go:11-25 — 安全メソッド以外でカスタムヘッダ必須)。さらに `protected.Use(middleware.LimitRequestBody(middleware.DefaultJSONBodyMaxBytes))` (composition_runtime.go:533) で JSON body 上限を再適用。

以下の全グループに `auth ● / csrf ● / cors ● / rate_limit △ (経路固有のみ) / security_headers ● / logging ●` が適用される。RBAC は各経路の `requirePermission`/`RequirePermissionAllowingAssignedClinicGrant`/`requireSystemAdmin` で個別適用 (authz はミドルウェア列とは別にハンドラ前段で強制)。

| 登録元 (composition_runtime.go) | 登録関数 | 主なグループ | auth | csrf | cors | rate_limit | security_headers | logging |
|------------------|-----------|--------------|------|------|------|-----------|------------------|---------|
| :534-536 | auth `RegisterPermissionGroupRoutes(protected.Group("/masters"))` | `/masters` (permission groups) | ● | ● | ● | - | ● | ● |
| :551 | owner `Handler.RegisterRoutes` | `/owners` | ● | ● | ● | - | ● | ● |
| :552 | pet `Handler.RegisterRoutes` | `/pets`, `/owners/:id/report/pets`, `/pets/:id/sub-owners`, `/pets/:id/chronic-conditions`, `/masters` (pet 系) | ● | ● | ● | - | ● | ● |
| :553-556 | staff `Handler.RegisterRoutes` (内部で RegisterShiftRoutes / RegisterShiftTemplateRoutes) | `/masters` (staff/shift/shift-template 系) | ● | ● | ● | - | ● | ● |
| :557, :581-588 | clinic `RegisterClinicRoutes` / `RegisterClinicHolidayRoutes` / `RegisterCompanyRoutes` / `RegisterClosingSettingsRoutes` | `/clinics`, `/clinic-holidays`, `/company`, `/closing-settings` | ● | ● | ● | - | ● | ● |
| :596-600 | manualarticle `RegisterRoutes` | `/manual/articles` | ● | ● | ● | - | ● | ● |
| :601-611 | support `RegisterRoutes` | `/support` (chat, bug-reports) | ● | ● | ● | △ `POST /support/chat` のみ追加 RateLimit (20  req/min, burst 10 — main.go:25-26) | ● | ● |
| :612-620 | identitylink `RegisterRoutes` | `/identity-links` | ● | ● | ● | - | ● | ● |
| :621-625 | inventory `RegisterRoutes` | `/inventory`, `/masters` (inventory 系) | ● | ● | ● | - | ● | ● |
| :626-636 | trimming `RegisterRoutes` (+WithImageSupport) | `/trimmings`, `/masters` (trimming 系), 画像アップロード経路 | ● | ● | ● | - | ● | ● |
| :637-644 | medicalrecord `RegisterRoutes` | `/masters`, `/medical-records`, `/hospitalizations`, `/lab-device-item-masters`, `/lab-devices`, `/lab-device`, `/lab-imports`, `/lab-reports`, examinations/packages | ● | ● | ● | - | ● | ● |
| :562 | reservation `RegisterRoutes` | `/masters/reservation-type*`, `/clinics/:clinic_id/reservation-types|reservation-staffs|reservations|line-reservation-settings`, `/reservations` | ● | ● | ● | - | ● | ● |
| :567 | billing `Handler.RegisterRoutes` | `/masters` (billing 系), `/payment-methods`, `/estimates`, `/medical-records`, `/billing-items`, `/accountings`, `/cash-register`, `/reports` | ● | ● | ● | - | ● | ● |
| :569 | lstep `RegisterRoutes` (内部で RegisterLstepDeliveryMonitorRoutes / RegisterLstepTriggerPriorityRoutes / RegisterAggregationRoutes / RegisterLstepCsvImportRoutes / RegisterLstepAnalyticsRoutes / RegisterSharedFileRoutes) | `/lstep-settings`, `/clinics/:clinic_id/lstep-settings`, `/owners/:id/line*`, `/owners/:id/lstep*`, `/lstep/*`, `/clinics/:clinic_id/lstep*`, `/lstep-tag-*`, `/pets/:id/death`, `/clinics/:clinic_id/line-customers`, `/shared-files`, aggregation/csv-import/analytics/delivery-monitor/trigger-priority | ● | ● | ● | - | ● | ● |
| :91-92 (auth) | auth 保護内 `GET /me`, `PUT /users/me/password` | `/me`, `/users/me/password` | ● | ● | ● | - | ● | ● |

RBAC 注記: `requireSystemAdmin()` (lstep/routes.go:52-66) が tag-config mutation を system_admin に限定。`RequirePermissionAllowingAssignedClinicGrant` は assigned-clinic grant を許容するが、GET/HEAD は返却対象 clinic ごとの filter/authorize が caller の義務 (SECURITY.md Caller 義務)。support の bug-reports 一覧・作成は全医院共有 (owner 承認済み製品決定) で変更操作のみ報告元 clinic スコープ。

## 5. 非保護公開面 (protected グループ外)

| 経路 | auth | csrf | cors | rate_limit | security_headers | logging | 登録・防御 |
|------|------|------|------|-----------|------------------|---------|-----------|
| `GET /api/liff/:clinicId/settings` | - | - | ● | ● (30 req/min) | ● | ● | LiffBodyLimit 適用 (reservation/routes.go:178,181) |
| `POST /api/liff/:clinicId/link` | △ | - | ● | ● (10 req/min) | ● | ● | link_token + line_id_token による自己認証 (routes.go:183) |
| `GET|POST|DELETE /api/liff/:clinicId/{profile,courses,trimming-courses,trimming-options,staffs,available-dates,available-times,reservations,my-reservations,my-reservations/:id,health-card}` | △ LiffAuth | - | ● | ● (60 req/min + my-reservations GET 30 req/min) | ● | ● | `authed` グループに `LiffRateLimit(60)` + `LiffAuth` (LINE ID トークン→clinic/customer 解決) (routes.go:186-199) |
| `POST /api/line/webhook` | △ HMAC | - | ● | ● (50 rps / burst 200) | ● | ● | JWT 認証なし・HMAC-SHA256 署名検証 (lstep/routes.go:226-233, composition_runtime.go:570-577, main.go:20-21) |
| `POST|DELETE /api/v1/uat/synthetic-closings*` | △ | - | ● | - | ● | ● | `RegisterUATRoutes(router.Group("/api/v1"))` — protected 外。`allowHTTP`: APP_ENV/DB_HOST/Host ゲート + `Authorization: Bearer <UAT_SYNTHETIC_CLOSING_SECRET>` (定数時間比較・未設定 fail-closed)、不適合は 404 (synthetic_closing_http.go:44-155)。staging/production では 404 |

## 6. GET/HEAD 在庫 (参照分類)

`backend/cmd/api/testdata/get_head_permissions.json` は 210 件の GET/HEAD 経路を class 別に保持 (clinic-fixed 159 / cross-clinic 21 / liff 9 / shared-master 9 / public 6 / self 3 / global 3)。`get_head_permission_*_test.go` がこの分類の機械ゲート。後段ユニットが未認証 GET 対象を選ぶ際の正本として使う。

## 7. 観測された適用の特記 (結論ではなく記録)

- CSRF (RequireXRequestedWith) は保護グループと login/logout/refresh 系公開経路に適用されるが、`forgot-password` / `reset-password` / LIFF / webhook / UAT / health / `_internal` には適用されない — 各々が独自の認証・ゲートを持つ設計。C08 での評価は「適用されている経路」と「別防御を持つ経路」を分けて行うこと。
- rate_limit はグローバル適用ではなく経路別 (auth 系・LIFF・webhook・support chat)。全て process-local store — multi-instance 共有 quota ではない (SECURITY.md LIMIT-1 / rate_limit.go)。
- `protected` 外で `/api/v1` にぶら下がる唯一の surface は UAT (環境+Bearer ゲート)。それ以外の `/api/v1` ドメイン経路は全て protected グループ経由。
- Worker 層が `/_internal/*` を外部から遮断するため、Gin の `/_internal/scheduled-jobs/:jobAction` は Worker coordinator の内部 containerFetch 経路であり、かつ X-Scheduler-Token を独立に要求する (defense-in-depth)。
