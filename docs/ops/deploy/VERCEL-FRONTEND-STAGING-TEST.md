# Vercel STG frontend verification

> **目的**: Vite frontendのdeployment、assets、cookie-auth API、settings routesを確認する。
> **対象**: `https://stg.noah-karte.com`。credential/cookie値を記録しない。

## 0. Release stop: prebuilt `/api` route

`.github/workflows/frontend-deploy.yml`は`.vercel/output/config.json`を手動生成し、現状は`frontend/vercel.json`のAPI rewriteを含めない。したがってprebuilt deployでは`/api/...`がSPA fallbackの`index.html`へ入る可能性がある。

API smokeの前にworkflowを修正し、filesystem/SPA fallbackより前に次のrouteをoutput configへencodeしてdeploymentで検証する。

```json
{"src":"/api/(.*)","dest":"https://api.stg.noah-karte.com/api/$1"}
```

`https://stg.noah-karte.com/api/...`がAPIのJSON/statusを返し、`index.html`を返さないことがrelease gate。source defectが直るまではこのrunbookのlogin/API smokeは **BLOCKED**。docsだけで解消済みにしない。

## 1. Deployment evidence

- reviewed `main -> staging` PRとfrontend path-filtered workflow runを確認する。
- Vercel deployment status、commit SHA、target domainを一致させる。
- frontend production workflowにはEnvironment approval gateが無い。gate実装・検証前に`production`へmerge/pushしない。
- backend healthだけをfrontend/API integrationの成功証跡にしない。

## 2. Public page and assets

1. `https://stg.noah-karte.com`をprivate browser sessionで開く。
2. Vite app shell/login formが表示される。
3. DevTools Networkでhashed `/assets/*.js` とCSSが`200/304`であることを確認する。
4. Consoleのuncaught error、asset 404、network 5xxを確認する。

これはViteであり、Next.js dev-server warningや固定`app.js`を期待値にしない。

## 3. Cookie-authenticated API

Approved provisioned accountでbrowser loginする。frontendはHttpOnly cookieと`withCredentials`を使う。`Authorization: Bearer` headerを期待しない。

DevTools Networkで、例えば次のsame-origin requestを確認する。

```http
GET https://stg.noah-karte.com/api/v1/clinics?scope=all
X-Requested-With: XMLHttpRequest
Cookie: <browser-managed HttpOnly cookies; value must not be copied>
```

- required permissionがあるsessionなら`200`、無ければcontractどおり`403`。
- response content-type/bodyがAPI JSONで、SPA HTMLでない。
- cookie/token/header valueをscreenshot、document、artifactへ残さない。

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
2. frontend same-origin `/api/...`がJSONかHTMLか確認する。
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
- prebuilt configに`/api` rewriteがあり、API JSON/statusを確認
- approved accountのHttpOnly cookie login
- exact settings routes success
- no unexpected console/network errors

source rewrite defect、account provisioning、external deployment state、approval gateのいずれかが未確認なら、該当項目を`BLOCKED`/`UNVERIFIED`と記録する。
