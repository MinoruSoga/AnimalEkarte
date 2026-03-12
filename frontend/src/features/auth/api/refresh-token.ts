import { axios } from "@/lib/axios";
import type { AuthUser } from "../types";
import { type BackendMeResponse, mapMeToAuthUser } from "./transforms";

export interface RefreshResponse {
  accessToken: string;
  user: AuthUser;
}

/**
 * 起動時セッション復元: localStorage の JWT で GET /me を呼び出す。
 * トークン不在または 401 の場合は null を返してログアウト扱いにする。
 */
export async function refreshToken(): Promise<RefreshResponse | null> {
  const token = localStorage.getItem("token");
  if (!token) return null;

  try {
    const { data: meData } = await axios.get<BackendMeResponse>("/v1/me");
    const user = mapMeToAuthUser(meData);
    return { accessToken: token, user };
  } catch {
    // 401 など認証失敗 — axios インターセプターが token を削除してリダイレクト済みのケースもある
    localStorage.removeItem("token");
    return null;
  }
}
