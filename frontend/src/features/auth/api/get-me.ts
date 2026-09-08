import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { ME_QUERY_KEY } from "@/lib/query-keys";
import { QUERY_STALE_TIMES } from "@/lib/react-query";
import { type BackendMeResponse, mapMeToAuthUser } from "./transforms";
import type { AuthUser } from "../types";

async function getMe(): Promise<AuthUser> {
  const { data } = await axios.get<BackendMeResponse>("/v1/me");
  return mapMeToAuthUser(data);
}

/** Shared so AuthProvider hydrate and tests pin the same cache policy. */
export const ME_QUERY_CACHE = {
  staleTime: QUERY_STALE_TIMES.SESSION,
  refetchOnWindowFocus: false as const,
  refetchInterval: false as const,
};

/**
 * /me は起動・ログイン・refreshToken の結果を ME_QUERY_KEY へ載せる。
 * staleTime 5分はキャッシュを fresh とみなす期間であり、5分後の自動再取得ではない。
 * 定期取得と window focus 再取得は無効。再取得はログイン、token refresh、
 * refreshPermissions、ME_QUERY_KEY 無効化、再マウント。自動ポーリングは追加しない。
 * UI の古い権限表示と BE の最終認可は別。パスワード変更は JWT epoch。
 *
 * @param enabled - 認証済みの場合のみ true を渡す（デフォルト: true）
 */
export function useGetMe(enabled = true) {
  return useQuery({
    queryKey: ME_QUERY_KEY,
    queryFn: getMe,
    enabled,
    staleTime: ME_QUERY_CACHE.staleTime,
    refetchInterval: ME_QUERY_CACHE.refetchInterval,
    refetchOnWindowFocus: ME_QUERY_CACHE.refetchOnWindowFocus,
    refetchIntervalInBackground: false,
  });
}
