import { useState, useCallback, useMemo, useEffect, useRef } from "react";
import type { ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useLocation } from "react-router";
import { toast } from "sonner";
import type { AuthContextValue, AuthUser, Resource, ResourceAction } from "@/types/auth";
import { AuthContext } from "@/hooks/auth-context";
import { isPasswordRecoveryPublicPath } from "@/lib/auth-route-policy";
import {
  CURRENT_CLINIC_STORAGE_KEY,
  getStoredClinicId,
  setStoredClinicId,
} from "@/lib/current-clinic";
import { ME_QUERY_KEY } from "@/lib/query-keys";
import { login as loginApi } from "../api/login";
import { logout as logoutApi } from "../api/logout";
import { refreshToken } from "../api/refresh-token";
import { useGetMe } from "../api/get-me";

/* セッション情報は httpOnly Cookie で管理するため localStorage への保存は不要。
 * 選択中のクリニック ID のみ localStorage に残す（権限情報ではないためリスク低） */
function saveClinicToStorage(clinicId: string): boolean {
  const ok = setStoredClinicId(clinicId);
  if (!ok && import.meta.env.DEV) {
    console.warn("[auth] failed to save clinic to localStorage");
  }
  return ok;
}

function removeClinicFromStorage(): void {
  try {
    localStorage.removeItem(CURRENT_CLINIC_STORAGE_KEY);
  } catch (error) {
    if (import.meta.env.DEV) {
      console.warn("[auth] failed to remove clinic from localStorage", error);
    }
  }
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const { pathname } = useLocation();
  const passwordRecovery = isPasswordRecoveryPublicPath(pathname);
  // BUG-031: restore on `/login` (and all protected routes) so an existing
  // cookie session hydrates and LoginForm can redirect. Password-recovery
  // public routes still skip restore to avoid noisy 401s on cold entry.
  const restoreSession = !passwordRecovery;
  const sessionKey = passwordRecovery ? "password-recovery" : "session";

  return (
    <AuthProviderSession key={sessionKey} restoreSession={restoreSession}>
      {children}
    </AuthProviderSession>
  );
}

interface AuthProviderSessionProps extends AuthProviderProps {
  restoreSession: boolean;
}

function AuthProviderSession({ children, restoreSession }: AuthProviderSessionProps) {
  const queryClient = useQueryClient();
  const [user, setUser] = useState<AuthUser | null>(null);
  const [currentClinicId, setCurrentClinicId] = useState<string | null>(null);
  const [isInitialized, setIsInitialized] = useState(!restoreSession);
  const initialAuthPromiseRef = useRef<ReturnType<typeof refreshToken> | null>(null);

  const hydrateUser = useCallback(
    (next: AuthUser) => {
      queryClient.setQueryData(ME_QUERY_KEY, next);
      setUser(next);
    },
    [queryClient],
  );

  // recovery/session 境界をまたぐと key により remount され、最新の Cookie 状態を
  // 取得する。同一 mount では StrictMode の effect 再実行時も 1 回だけ呼び出す。
  useEffect(() => {
    if (!restoreSession) return;

    let active = true;
    initialAuthPromiseRef.current ??= refreshToken().catch(() => null);

    void initialAuthPromiseRef.current.then((result) => {
      if (!active) return;

      if (result) {
        const storedClinic = getStoredClinicId();
        const validClinic = result.user.clinics.some((clinic) => clinic.clinicId === storedClinic);
        hydrateUser(result.user);
        setCurrentClinicId(validClinic ? storedClinic : result.user.mainClinicId);
      } else {
        setUser(null);
        setCurrentClinicId(null);
      }
      setIsInitialized(true);
    });

    return () => {
      active = false;
    };
  }, [hydrateUser, restoreSession]);

  // /me のキャッシュ（起動時 hydrate）でユーザー情報を同期する。
  // 定期ポーリングはしない。権限変更は refreshPermissions。
  const { data: meData } = useGetMe(user !== null);
  const [prevMeData, setPrevMeData] = useState(meData);
  if (prevMeData !== meData) {
    setPrevMeData(meData);
    if (meData) {
      setUser(meData);
    }
  }

  // FE5-2: マルチタブ穴 — 他タブでクリニックが切り替わった場合、
  // このタブは storage イベントを検知するまで旧クリニック画面のまま。
  // axios はリクエスト毎に localStorage の clinic を読むため、以後の書き込みが
  // 誤テナントで永続化されうる。他タブ由来の変更を検知したらフルリロードで揃える。
  useEffect(() => {
    function handleStorage(event: StorageEvent): void {
      if (event.key === CURRENT_CLINIC_STORAGE_KEY) {
        window.location.reload();
      }
    }
    window.addEventListener("storage", handleStorage);
    return () => window.removeEventListener("storage", handleStorage);
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      const result = await loginApi(email, password);
      hydrateUser(result.user);
      setCurrentClinicId(result.user.mainClinicId);
      saveClinicToStorage(result.user.mainClinicId);
    },
    [hydrateUser],
  );

  const logout = useCallback(async () => {
    try {
      await logoutApi();
    } catch {
      // Cookie はサーバー側で未クリアの可能性があるため警告する。
      // ローカル状態は finally で必ずクリアするため UI は /login へ遷移する。
      toast.warning(
        "ログアウト中にサーバーエラーが発生しました。ブラウザを閉じてセッションを終了することを推奨します。",
      );
    } finally {
      setUser(null);
      setCurrentClinicId(null);
      removeClinicFromStorage();
      queryClient.clear();
    }
  }, [queryClient]);

  const switchClinic = useCallback(
    (clinicId: string) => {
      if (!user) return;
      if (clinicId === currentClinicId) return;
      const isMember = user.clinics.some((c) => c.clinicId === clinicId);
      if (!isMember) return;
      // 1. localStorage 更新（リロード後に axios interceptor が新 clinic_id を送信する）
      // FE6-2: 書込失敗時はここで打ち切る。続行して reload すると旧クリニックIDのまま
      // 復帰し、ユーザーが切替成功と誤認する無音失敗になるため。
      if (!saveClinicToStorage(clinicId)) {
        toast.error("クリニックの切替に失敗しました。ブラウザのストレージ設定を確認してください。");
        return;
      }
      // 2. FE5-3: reload 前に React Query キャッシュを破棄する。
      //    現状は reload 1 行が安全性の全てを担っており、将来切替を SPA 化した際に
      //    clinic id を含まないクエリキー（accountings/medical-records 等）が
      //    旧クリニックのキャッシュを漏らす防壁として先に明示しておく。
      queryClient.clear();
      // 3. フルリロードで全データ（React Query + React Router loader）を新クリニックで再取得
      window.location.reload();
    },
    [user, currentClinicId, queryClient],
  );

  const hasPermission = useCallback(
    (resource: Resource, action: ResourceAction): boolean => {
      if (!user) return false;
      // system_admin はバイパス（BEも全権限 true で返すが念のため）
      if (user.isSystemAdmin) return true;
      const resourcePerms = user.permissions[resource];
      if (!resourcePerms) return false;
      return resourcePerms[action] === true;
    },
    [user],
  );

  const refreshPermissions = useCallback(async () => {
    const result = await refreshToken();
    if (result) {
      hydrateUser(result.user);
    }
  }, [hydrateUser]);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      currentClinicId,
      isAuthenticated: user !== null,
      isLoading: !isInitialized,
      login,
      logout,
      switchClinic,
      hasPermission,
      refreshPermissions,
    }),
    [
      user,
      currentClinicId,
      isInitialized,
      login,
      logout,
      switchClinic,
      hasPermission,
      refreshPermissions,
    ],
  );

  // セッション復元中は子を描画せず、復元前の一瞬だけ匿名 UI が見えることを防ぐ。
  if (!isInitialized) return null;

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
