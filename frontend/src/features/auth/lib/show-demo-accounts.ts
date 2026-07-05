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
 * vite.config.ts の `define` でビルド時定数として埋め込み、Vercel Production
 * デプロイでは他の設定に関わらず常に false にする一次ガードとして使う。
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
  const isVercelProduction = vercelEnv === "production";
  if (isVercelProduction) return false;
  return isDev || showDemoAccountsFlag === "true";
}
