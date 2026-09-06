# Vercel STG frontend verification

> **目的**: Vite frontendのdeployment、assets、cookie-auth API、settings routesを確認する。
> **対象**: `https://stg.noah-karte.com`。credential/cookie値を記録しない。

## 0. Build-time API target

現行 `frontend-deploy.yml` は `VERCEL_ENV=preview|production` を渡してビルドする。`frontend/vite.config.ts` はその値で `VITE_API_URL` を STG / production の絶対 API URL に固定し、`frontend/src/lib/axios.ts` が baseURL に使う。STG の通常 request は `https://api.stg.noah-karte.com/api/...` へ直接向かう。

prebuilt `.vercel/output/config.json` に `/api` rewrite はないが、上記の絶対 URL 経路に rewrite は不要。過去の「rewrite 欠落だけで login 全体 BLOCKED」という判定はこの build contract と一致しない。別の build path で same-origin `/api` を使う場合は、API JSON/status を返す rewrite が必要であり、SPA HTML への fallback は失敗とする。

設定の存在だけで疎通は証明しない。deployed artifact の API target と cookie/CORS を以下で確認する。

## 1. Deployment evidence

- reviewed `main -> staging` PRとfrontend path-filtered workflow runを確認する。
- Vercel deployment status、commit SHA、target domainを一致させる。
- frontend production job は `Production` Environment に bind し、production dispatch は production ref 以外を拒否する。外部 Required reviewers / branch protection と backend production gate が検証されるまで production delivery を進めない。
- backend healthだけをfrontend/API integrationの成功証跡にしない。

## 2. Public page and assets

1. `https://stg.noah-karte.com`をprivate browser sessionで開く。
2. Vite app shell/login formが表示される。
3. DevTools Networkでhashed `/assets/*.js` とCSSが`200/304`であることを確認する。
4. Consoleのuncaught error、asset 404、network 5xxを確認する。

これはViteであり、Next.js dev-server warningや固定`app.js`を期待値にしない。

## 3. Cookie-authenticated API

Approved provisioned accountでbrowser loginする。frontendはHttpOnly cookieと`withCredentials`を使う。`Authorization: Bearer` headerを期待しない。

DevTools Networkで、例えば次のcross-origin requestを確認する。

```http
GET https://api.stg.noah-karte.com/api/v1/clinics?scope=all
X-Requested-With: XMLHttpRequest
Cookie: <browser-managed HttpOnly cookies; value must not be copied>
```

- required permissionがあるsessionなら`200`、無ければcontractどおり`403`。
- response content-type/bodyがAPI JSONで、SPA HTMLでない。
- cookie/token/header valueをscreenshot、document、artifactへ残さない。

### 会計新規確定のCORS確認

新規会計確定は `Idempotency-Key` を付けてAPIへPOSTする。`backend/internal/middleware/cors.go` の固定許可ヘッダーにこれを含め、設定済みのoriginにだけアクセスを許可する。要求ヘッダーの無条件反映やwildcardで許可範囲を広げない。

OPTIONSの204やログイン成功だけで会計確定をPASSにしない。デプロイ後の承認済み検証では、origin・credentials・`Idempotency-Key` を含む要求ヘッダーの許可と、後続の会計POSTの到達を確認する。localのsame-origin proxy経由での成功は代用にならない（[会計精算仕様](../../spec/screens/11-accounting-detail.md#33-新規会計確定の再試行)）。

## 4. Exact settings routes

権限を持つapproved accountで次を直接開く。

- `/settings/clinic`
- `/settings/permission-groups`
- `/settings/staff`

各routeでVite navigation、API status、list/form render、tenant scopeを確認する。書込みは[CRUD smoke](./CRUD-SMOKE-TEST.md)のpayload/restore/cleanup contractに従う。

## 5. Troubleshooting

### Assets/build

- asset 404: Vercel output、base path、deployment SHAを確認する。
- type/build再現が必要な場合はproject Docker commandsを使う。例: `docker compose exec frontend pnpm type-check`、またはcurrent Make target。
- full install/build/type-checkがrepo policy上manual user-runならoperatorへexact commandを渡し、agentが自動実行しない。local bare `pnpm` / `npm`を案内しない。

### API

1. custom-domain `/health`とapproved workers.dev `/health`を比較する。
2. Network の実 API 接続先と JSON/status を確認する。same-origin build なら `/api/...` が SPA HTML でないことも確認する。
3. Cloudflare Workers Logs、Container、PlanetScale connectionを切り分ける。
4. CORS/cookie/`X-Requested-With` contractを確認する。
5. causeを修正後、Cloudflare/Vercel current artifactをrebuild/redeployする。

CloudWatch、ECS、CloudFrontはretiredでありrollback/triage pathに使わない。

### Cache

Vercel assetsはcontent-hashed。browser cacheをclearし、最新deploymentがtarget domainへpromoteされたか確認する。旧CloudFront invalidationを行わない。

## 6. Pass / block

PASSには次が全て必要。

- deployment SHA/domain一致
- hashed Vite assets success
- build-time API target が環境と一致し、cookie/CORS と API JSON/status を確認（same-origin build の場合は rewrite も確認）
- approved accountのHttpOnly cookie login
- exact settings routes success
- no unexpected console/network errors

API target / rewrite / CORS、account provisioning、external deployment state、approval gateのいずれかが未確認なら、該当項目を`BLOCKED`/`UNVERIFIED`と記録する。
