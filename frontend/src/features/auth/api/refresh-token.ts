import { axios } from "@/lib/axios";
import type { AuthUser } from "../types";
import { type BackendMeResponse, mapMeToAuthUser } from "./transforms";

export interface RefreshResponse {
  user: AuthUser;
}

/**
 * 起動時セッション復元: httpOnly Cookie を使用するため sessionStorage チェック不要。
 * GET /me を呼び出してセッションの有効性を確認する。
 * - Cookie が有効 → ユーザー情報を返す
 * - Cookie がない/期限切れ → null を返してログアウト扱いにする
 */
export async function refreshToken(): Promise<RefreshResponse | null> {
  try {
    // withCredentials: true により Cookie が自動送信される
    const { data: meData } = await axios.get<BackendMeResponse>("/v1/me");
    const user = mapMeToAuthUser(meData);
    return { user };
  } catch {
    // 401 / ネットワークエラー → セッションなし
    return null;
  }
}
