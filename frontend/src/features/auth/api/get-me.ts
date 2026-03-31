import Axios from "axios";
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { type BackendMeResponse, mapMeToAuthUser } from "./transforms";
import type { AuthUser } from "../types";

export const ME_QUERY_KEY = ["me"] as const;

export async function getMe(): Promise<AuthUser> {
  const { data } = await axios.get<BackendMeResponse>("/v1/me");
  return mapMeToAuthUser(data);
}

/**
 * /me を定期ポーリングして権限を最新に保つ。
 * - staleTime: 4分（ほぼ毎回 refetch を避けつつ fresh に保つ）
 * - refetchInterval: 5分（バックグラウンドでは停止）
 * - refetchOnWindowFocus: true（タブアクティブ時に再取得）
 * - 401 時はポーリング停止（ログインページへの無限リダイレクト防止）
 *
 * @param enabled - 認証済みの場合のみ true を渡す（デフォルト: true）
 */
export function useGetMe(enabled = true) {
  return useQuery({
    queryKey: ME_QUERY_KEY,
    queryFn: getMe,
    enabled,
    staleTime: 4 * 60 * 1000,
    refetchInterval: (query) => {
      if (
        query.state.error !== null &&
        Axios.isAxiosError(query.state.error) &&
        query.state.error.response?.status === 401
      ) {
        return false;
      }
      return 5 * 60 * 1000;
    },
    refetchOnWindowFocus: true,
    refetchIntervalInBackground: false,
  });
}
