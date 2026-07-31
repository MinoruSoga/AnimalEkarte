/**
 * M-10 / SEC-CS2-F01: デモアカウント欄をローカル Vite DEV のみに制限する判定ロジック。
 *
 * 以前は Vercel preview（STG）+ `VITE_SHOW_DEMO_ACCOUNTS=true` でも表示していたが、
 * 特権デモ資格情報（system_admin 等）が共有 STG / preview へ露出するリスクがあるため、
 * local DEV のみ許可する deny-by-default に変更した。
 *
 * - ローカル dev（isDev）→ 表示
 * - それ以外（Vercel preview / production / 未定義 / flag 有無を問わず）→ 非表示
 *
 * `VERCEL_ENV` と `VITE_SHOW_DEMO_ACCOUNTS` は後方互換のため引数に残すが、判定には使わない。
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
  // SEC-CS2-F01: local Vite DEV only. Preview/flag must never enable demo UI.
  void params.vercelEnv;
  void params.showDemoAccountsFlag;
  return params.isDev;
}
