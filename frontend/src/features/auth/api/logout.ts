import { axios } from "@/lib/axios";

/**
 * ログアウト: サーバーに通知して httpOnly Cookie をクリアさせる。
 * Cookie は httpOnly のため JavaScript から直接削除できない。
 * バックエンドが Set-Cookie: auth_token=; Max-Age=0 を返してクリアする。
 */
export async function logout(): Promise<void> {
  await axios.post("/v1/logout");
}
