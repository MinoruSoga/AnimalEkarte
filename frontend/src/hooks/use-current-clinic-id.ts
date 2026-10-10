import { useContext } from "react";

import { AuthContext } from "@/hooks/auth-context";
import { getStoredClinicId } from "@/lib/current-clinic";

/**
 * 認証コンテキストの現在クリニック ID を返す。
 * Provider 外（Storybook/テスト等）では localStorage の値にフォールバック。
 *
 * Codex PR606 #10: clinic 切替で未解決の fetch が同一 query key へ書き込む
 * 混在を防ぐため、clinic スコープの query key にはこの値を含める。
 */
export function useCurrentClinicId(): string | null {
  const ctx = useContext(AuthContext);
  return ctx?.currentClinicId ?? getStoredClinicId();
}
