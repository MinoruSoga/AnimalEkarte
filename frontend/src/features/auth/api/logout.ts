/**
 * ログアウト: サーバーサイドセッションは現状なし（JWT はステートレス）。
 * localStorage から token を削除するのみ。
 */
export function logout(): Promise<void> {
  localStorage.removeItem("token");
  return Promise.resolve();
}
