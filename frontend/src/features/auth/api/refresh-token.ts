import { axios } from "@/lib/axios";
import type { AuthUser } from "../types";
import { type BackendMeResponse, mapMeToAuthUser } from "./transforms";

export interface RefreshResponse {
  accessToken: string;
  user: AuthUser;
}

/**
 * 起動時セッション復元: httpOnly Cookie が有効なら GET /me が成功する。
 * Cookie 未存在または期限切れの場合は null を返してログアウト扱いにする。
 * localStorage へのアクセスは不要（Cookie は自動送信される）。
 */
export async function refreshToken(): Promise<RefreshResponse | null> {
  try {
    const { data: meData } = await axios.get<BackendMeResponse>("/v1/me");
    const user = mapMeToAuthUser(meData);
    return { accessToken: "", user };
  } catch {
    // 401 / ネットワークエラー → セッションなし
    return null;
  }
}
