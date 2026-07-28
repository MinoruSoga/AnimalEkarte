/**
 * M-10: 本番ビルドでデモアカウント欄を必ず非表示にするための判定ロジックの仕様。
 *
 * `frontend/.env.production` の `VITE_SHOW_DEMO_ACCOUNTS` は STG 環境（Vercel preview
 * デプロイ）向けの動作確認用フラグであり、意図的に "true" のまま維持される。しかし
 * この値だけに頼ると、設定ミスや preview 用の値が本番へ紛れ込んだ場合に本番でも
 * デモアカウントが露出してしまう。
 *
 * `VERCEL_ENV` は Vercel がビルド時に自動注入する値（"production" | "preview" |
 * "development"）で、Vercel Dashboard の環境変数設定ミスの影響を受けない。これを
 * vite.config.ts の `define` でビルド時定数として埋め込む。
 *
 * DEC-7 / SEC-DEMO-FAILCLOSED: deny-by-default の allowlist 形を使う。
 * - ローカル dev（isDev）→ 表示
 * - Vercel preview + VITE_SHOW_DEMO_ACCOUNTS=true → 表示
 * - それ以外（production / 未定義 "" / development 等）→ 非表示
 * 旧式 `vercelEnv !== "production"` は未定義時に fail-open するため禁止。
 *
 * 注意: `LoginForm.tsx` の `SHOW_DEMO` は本関数を呼ばず同じロジックをリテラル式で
 * インライン化している（関数呼び出しにすると Vite/esbuild が定数畳み込みできず、
 * 本番バンドルからデモ認証情報の配列が tree-shake されないため）。本関数はその
 * ロジックの単体テスト用リファレンス実装であり、`SHOW_DEMO` の式を変更した場合は
 * 本関数と `show-demo-accounts.test.ts` も合わせて更新すること。
 */
export function computeShowDemoAccounts(params: {
  vercelEnv: string;
  isDev: boolean;
  showDemoAccountsFlag: string | undefined;
}): boolean {
  const { vercelEnv, isDev, showDemoAccountsFlag } = params;
  return isDev || (vercelEnv === "preview" && showDemoAccountsFlag === "true");
}
