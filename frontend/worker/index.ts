// EMR-255: frontend 配信用 Cloudflare Worker。Vercel(frontend/vercel.json)が担っていた
// rewrite / header 規則を Workers Static Assets + service binding へ移植したもの。
//
// 経路対応:
//   /api/*          -> service binding で backend Worker (animalekarte-{stg,prod}-api) へ
//                      同一オリジン中継。line-reserve / liff は相対パス /api/* を叩くため
//                      本経路が必須(Vercel 時代は /api/:path* の外部 rewrite が担っていた)。
//   静的資産         -> ASSETS binding。content-hash 付き /assets/* の immutable cache と
//                      CSP 等の security header は dist/_headers(frontend/public/_headers)が付与。
//   /line-reserve/* -> ASSETS 404 時に /line-reserve/index.html へ fallback
//                      (vercel.json rewrites 相当。/{clinicId}/ 形式の LIFF 予約導線)。
//   /liff/*         -> ASSETS 404 時に /liff/index.html へ fallback。
//   その他           -> ASSETS 404 時に /index.html へ SPA fallback。
//
// vercel.json の { handle: filesystem } + rewrites との対応は「まず ASSETS.fetch で実在
// ファイルを優先し、404 なら各 index.html へ fallback」で等価になる。

interface BindingFetcher {
  fetch(request: Request): Promise<Response>;
}

export interface FrontendWorkerEnv {
  ASSETS: BindingFetcher;
  API: BindingFetcher;
}

// Workers Static Assets は /foo/index.html への fetch をディレクトリ URL へ
// 307 リダイレクトする(pretty URL 正規化)。SPA fallback でリダイレクトを返すと
// ブラウザの URL が書き換わり deep link が壊れるため、fallback 先は
// index.html ではなくディレクトリパス(/, /line-reserve/ 等)を指定する。
const SPA_ENTRYPOINTS: ReadonlyArray<readonly [prefix: string, entrypoint: string]> = [
  ["/line-reserve/", "/line-reserve/"],
  ["/liff/", "/liff/"],
];

function fallbackEntrypoint(pathname: string): string {
  for (const [prefix, entrypoint] of SPA_ENTRYPOINTS) {
    if (pathname.startsWith(prefix)) return entrypoint;
  }
  return "/";
}

async function handleRequest(request: Request, env: FrontendWorkerEnv): Promise<Response> {
  const url = new URL(request.url);

  if (url.pathname === "/api" || url.pathname.startsWith("/api/")) {
    return env.API.fetch(request);
  }

  const asset = await env.ASSETS.fetch(request);
  if (asset.status !== 404) return asset;
  if (request.method !== "GET" && request.method !== "HEAD") return asset;

  return env.ASSETS.fetch(new Request(new URL(fallbackEntrypoint(url.pathname), url), request));
}

// eslint-disable-next-line no-restricted-syntax -- Workers runtime は default export の fetch handler を必須とする(プラットフォーム契約)
export default {
  fetch: handleRequest,
};
