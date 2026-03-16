import { axios } from "@/lib/axios";
import type { AuthUser } from "../types";
import { type BackendMeResponse, mapMeToAuthUser } from "./transforms";

export interface RefreshResponse {
  accessToken: string;
  user: AuthUser;
}

/**
 * 起動時セッション復元: sessionStorage から JWT トークンを読み込む。
 * トークン存在 → GET /me で ユーザー情報を復元
 * トークン未存在または期限切れ → null を返してログアウト扱いにする
 */
export async function refreshToken(): Promise<RefreshResponse | null> {
  const token = sessionStorage.getItem("auth_token");

  // sessionStorage にトークンがない場合はセッションなし
  if (!token) {
    return null;
  }

  try {
    // Authorization Bearer で /me を呼び出す（axios インターセプターが自動設定）
    const { data: meData } = await axios.get<BackendMeResponse>("/v1/me");
    const user = mapMeToAuthUser(meData);
    return { accessToken: token, user };
  } catch {
    // 401 / ネットワークエラー → セッションなし（トークンをクリア）
    sessionStorage.removeItem("auth_token");
    return null;
  }
}
