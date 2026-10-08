# U04 — Case Verdict Record: C08 (CSRF custom-header enforcement)

Campaign: emr-332-security-defense / Unit: U04-BROWSER-EDGE / Attempt: att-u04-1
Template: ../u01-case-verdict-template.md (12 fields)

| # | Field | 記録 |
|---|-------|------|
| 1 | ケース ID | `C08` — カスタムヘッダ CSRF 強制 (`X-Requested-With`) |
| 2 | 対象リビジョン | 静的・模擬対象: task branch `feat/emr-332-security-defense` base `de28e9765a29091339a872a85f1875a947e07e5b` (u01-safety-record header) + w04 worktree 追加物 (`backend/internal/middleware/emr332_u04/` 新規テストのみ)。**staging 稼働リビジョン: undetermined** — 本ユニットの U01 §2 照合試行の結果、repo 内に稼働リビジョンを特定する deploy 記録・Worker バージョン刻印・既知差分証跡は存在しない (SECURITY.md:91 は 2026-10-03 の疎通確認であり commit 対応の証明ではない; deploys は GitHub Actions `staging` Environment 経由で run 記録は外部)。u01-safety-record §2 規則により staging-observed は BLOCKED。 |
| 3 | 実施方式 (method) | `simulated` (httptest 経由で実 `RequireXRequestedWith` middleware を実行) + `static` (適用範囲・組合せ評価)。staging-observed 行は BLOCKED (上記)。 |
| 4 | 正常条件 | 期待: safe method (GET/HEAD/OPTIONS) はヘッダ無しで通過; non-safe method は `X-Requested-With` 非空で通過し downstream handler が呼ばれる。観測: GET/HEAD/OPTIONS ヘッダ無し → 200 + downstream 呼出 1 回; POST + `X-Requested-With: XMLHttpRequest` → 200 + downstream 呼出 1 回 (TestEMR332C08_*)。 |
| 5 | 拒否条件 | 期待: non-safe method で `X-Requested-With` 欠落 → 403 で handler 未到達。観測: POST/PUT/PATCH/DELETE いずれもヘッダ無し → 403 `{"error":"X-Requested-With header required for state-changing requests"}` + **downstream 呼出回数 = 0** (spy counter、cors_test.go:44-46 の先例パターン)。csrf.go:20-22 `AbortWithStatusJSON` で abort。 |
| 6 | 防御判定 | **PASS** (simulated 範囲: 適用経路上での middleware 実動作)。staging-observed 判定は BLOCKED — 稼働リビジョン照合未確定のため未送信。 |
| 7 | 副作用呼出回数 | 0 — 拒否経路で downstream handler の呼出回数実測 0 (全 4 non-safe method)。正常経路の handler 呼出は制御群として計測 (1 回)。session/監査/DB 等の副作用は本模擬内に存在しない。 |
| 8 | 検出判定 | **UNKNOWN** — CSRF 拒否のログ捕捉証跡なし。`logging.go` は U06 所有 (read-only) であり、marker-capture による検出証拠は本ユニットでは取得していない。403 応答自体の送出は確認済みだが観測・検出の成立を意味しない。 |
| 9 | 安全な証跡参照 | `backend/internal/middleware/csrf.go:11-25` (実装); `backend/internal/middleware/emr332_u04/browser_edge_test.go` `TestEMR332C08_CSRFBlocksStateChangingRequestsWithoutHeader`; `docs/work/emr-332-security/u04/evidence-t01-middleware-tests.txt` (verbatim go test 出力); 適用範囲 `backend/internal/auth/http_routes.go:53-93`, `u01-route-middleware-map.md` §3-5。 |
| 10 | 制約・未検証範囲 | **適用範囲 (u01-route-middleware-map §3-5 より)**: CSRF は保護グループ全体 (`protected.Use(Authenticate)` → `protected.Use(CSRF)` — 認証が先に走る) と公開 auth 4 経路 (`login`/`logout`/`auth/refresh`/`auth/refresh/logout`, http_routes.go:53-76) に適用。**未適用は別防御を持つ経路**: `forgot-password`/`reset-password` (:77-86 — rate limit のみ、token 所持前提で ambient cookie 非依存), LIFF 面 (LiffAuth = LINE ID token), `/api/line/webhook` (HMAC-SHA256 署名), UAT (env+Host+Bearer, 404 fail-closed), health, `/_internal/*` (edge 404)。**組合せ評価**: release では cookie が `SameSite=None; Secure` (http_types.go:88-93) で cross-site 送信可能 — ゆえに non-simple ヘッダ `X-Requested-With` 必須化が実効 CSRF 防御。cross-origin での当該ヘッダ送信は preflight 許可が前提で、許可は設定 origin のみ (cors.go:18-29 + edge 複製 worker/index.ts:388-427)。non-browser クライアントはヘッダを常に付けられるため本防御は browser-forgery 対策であり認可境界ではない (protected 群では Authenticate が先行)。**未検証**: staging-observed (BLOCKED)、各 cookie 属性の live 観測 (login 禁止により恒久的に static-only)。 |
| 11 | 修復案 | **適用なし (none applied)** — owned_paths 内に gap 証拠を確認できず。csrf.go の方式は既定方針どおりの動作を実測で確認。設計上の注記 (修復ではない): `forgot-password`/`reset-password` は CSRF 未適用だが匿名経路で cookie 認証に依存しないため ambient-auth 偽造の対象外 — 現状は整合的。token-mode 再設計は本ケースの評価範囲外で要求もしない。 |
| 12 | 再検証結果 | 修復なしのため再検証不要。初回検証自体が scoped Docker で PASS: `go test -count=1 ./internal/middleware/...` → `ok internal/middleware 0.069s` + `ok internal/middleware/emr332_u04 0.008s` (evidence-t01 verbatim)。 |

## 補足: 稼働リビジョン照合の試行記録 (U01 §2 gate)

照合方法は u01-safety-record §2 の3経路をすべて試行: (a) deploy 記録 — repo 内に存在しない (`.github/workflows/backend-deploy.yml` は `staging` push 時の Actions deploy で run 記録は GitHub 外部、`reports/` に deploy stamp なし); (b) Worker バージョン識別 — repo 内に version stamp ファイルなし (`wrangler containers list`/Cloudflare API は本ユニットのネットワーク bounds 外で実行不可); (c) 既知差分証跡 — U01 系ドキュメントは base `de28e976` 時点の機械導出で、staging-served 側との差分証拠は存在しない。**結論: undetermined → staging-observed は全行 BLOCKED、live 送信 0 回。**
