import { axios } from "@/lib/axios";

/**
 * ログアウト: sessionStorage から JWT トークンを削除し、サーバーに通知する。
 */
export async function logout(): Promise<void> {
  try {
    await axios.post("/v1/logout");
  } catch {
    // API 失敗時もローカル状態はクリアされるため無視
  } finally {
    // sessionStorage から認証トークンをクリア（JavaScript で直接削除可能）
    sessionStorage.removeItem("auth_token");
  }
}
