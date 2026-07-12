import Axios from "axios";
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
 * - 401/403 → Cookie なし/期限切れ → null を返してログアウト扱いにする
 * - 5xx/ネットワークエラー → null を返す（コンソールに警告出力）
 */
export async function refreshToken(): Promise<RefreshResponse | null> {
  try {
    // withCredentials: true により Cookie が自動送信される
    const { data: meData } = await axios.get<BackendMeResponse>("/v1/me");
    return { user: mapMeToAuthUser(meData) };
  } catch (error) {
    if (Axios.isAxiosError(error) && error.response) {
      const status = error.response.status;
      if (status === 401 || status === 403) {
        // 正常なセッション失効
        return null;
      }
      // 5xx: バックエンド障害の可能性
      if (import.meta.env.DEV) {
        console.warn("[auth] refreshToken: unexpected server error", { status });
      }
    } else if (Axios.isAxiosError(error)) {
      // ネットワーク断
      if (import.meta.env.DEV) {
        console.warn("[auth] refreshToken: network error", { message: error.message });
      }
    }
    return null;
  }
}
