/**
 * M-10 / SEC-CS2-F01: デモアカウント欄の表示判定。
 *
 * - ローカル Vite DEV → 表示
 * - Vercel preview（STG、`stg.noah-karte.com`）→ 表示
 * - Vercel production / VERCEL_ENV 未定義の本番相当ビルド → 非表示
 *
 * `VITE_SHOW_DEMO_ACCOUNTS` は後方互換のため引数に残すが判定に使わない。
 * `frontend/.env.production` が flag=true のため、flag を信じると本番バンドルにも
 * デモ一覧が残る。
 *
 * 注意: `LoginForm.tsx` の `SHOW_DEMO` は本関数を呼ばず、
 * `import.meta.env.DEV || __VERCEL_ENV__ === "preview" || import.meta.env.VITE_VERCEL_ENV === "preview"`
 * をリテラルで置く（関数呼び出しにすると Vite/esbuild が定数畳み込みできず、
 * 本番バンドルからデモ認証情報の配列が tree-shake されないため）。本関数はその
 * ロジックの単体テスト用リファレンス実装であり、`SHOW_DEMO` の式を変更した場合は
 * 本関数と `show-demo-accounts.test.ts` も合わせて更新すること。
 */
export function computeShowDemoAccounts(params: {
  vercelEnv: string;
  isDev: boolean;
  showDemoAccountsFlag: string | undefined;
}): boolean {
  void params.showDemoAccountsFlag;
  return params.isDev || params.vercelEnv === "preview";
}
