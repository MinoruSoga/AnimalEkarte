import { axios } from "@/lib/axios";

/**
 * ログアウト: サーバーに通知して httpOnly Cookie をクリアさせる。
 * Cookie は httpOnly のため JavaScript から直接削除できない。
 * 旧 refresh cookie の Path に含まれる互換ルートを使い、新旧 token を同時に失効する。
 */
export async function logout(): Promise<void> {
  await axios.post("/v1/auth/refresh/logout");
}
