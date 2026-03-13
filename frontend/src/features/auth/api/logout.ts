import { axios } from "@/lib/axios";

/**
 * ログアウト: POST /logout でサーバー側の httpOnly Cookie を無効化する。
 * Cookie は HttpOnly のため JS から直接削除できない。
 */
export async function logout(): Promise<void> {
  try {
    await axios.post("/v1/logout");
  } catch {
    // API 失敗時もローカル状態はクリアされるため無視
  }
}
