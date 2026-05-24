import { useState, useCallback, useMemo, use } from "react";
import type { ReactNode } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { AuthContextValue, AuthUser, Resource, ResourceAction } from "@/types/auth";
import { AuthContext } from "@/contexts/auth-context";
import { login as loginApi } from "../api/login";
import { logout as logoutApi } from "../api/logout";
import { refreshToken } from "../api/refresh-token";
import { useGetMe } from "../api/get-me";

// Re-export useAuth from shared hooks for backward compatibility within this feature
export { useAuth } from "@/hooks/use-auth";

/* セッション情報は httpOnly Cookie で管理するため localStorage への保存は不要。
 * 選択中のクリニック ID のみ localStorage に残す（権限情報ではないためリスク低） */
const STORAGE_KEY_CLINIC = "auth_current_clinic:v1";

function readClinicFromStorage(): string | null {
  try {
    return localStorage.getItem(STORAGE_KEY_CLINIC);
  } catch {
    return null;
  }
}

function saveClinicToStorage(clinicId: string): void {
  try {
    localStorage.setItem(STORAGE_KEY_CLINIC, clinicId);
  } catch (error) {
    console.warn("[auth] failed to save clinic to localStorage", error);
  }
}

function removeClinicFromStorage(): void {
  try {
    localStorage.removeItem(STORAGE_KEY_CLINIC);
  } catch (error) {
    console.warn("[auth] failed to remove clinic from localStorage", error);
  }
}

/**
 * Initial session restoration promise.
 * Module-level で一度だけ作成する（アプリ起動時に 1 回だけ /v1/me を呼ぶ）。
 * ログイン後は AuthContext の login() が setUser() を直接呼ぶため
 * フルリロードは不要。
 */
const initialAuthPromise = refreshToken().catch(() => null);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  // React 19 use() でセッション復元が完了するまでサスペンドする
  const initialResult = use(initialAuthPromise);

  const [user, setUser] = useState<AuthUser | null>(initialResult?.user ?? null);
  const [currentClinicId, setCurrentClinicId] = useState<string | null>(() => {
    if (!initialResult) return null;
    const storedClinic = readClinicFromStorage();
    const validClinic = initialResult.user.clinics.some((c) => c.clinicId === storedClinic);
    return validClinic ? storedClinic : initialResult.user.mainClinicId;
  });

  // isSwitchingClinic: クリニック切替はフルリロードで行うため常に false
  const isSwitchingClinic = false;

  // /me の定期ポーリング結果でユーザー情報（権限含む）を同期
  // 認証済みかつローディング完了後のみポーリングを有効化
  const { data: meData } = useGetMe(user !== null);
  const [prevMeData, setPrevMeData] = useState(meData);
  if (prevMeData !== meData) {
    setPrevMeData(meData);
    if (meData) {
      setUser(meData);
    }
  }

  const queryClient = useQueryClient();

  const login = useCallback(async (email: string, password: string) => {
    const result = await loginApi(email, password);
    setUser(result.user);
    setCurrentClinicId(result.user.mainClinicId);
    saveClinicToStorage(result.user.mainClinicId);
  }, []);

  const logout = useCallback(async () => {
    try {
      await logoutApi();
    } catch {
      // Cookie はサーバー側で未クリアの可能性があるため警告する。
      // ローカル状態は finally で必ずクリアするため UI は /login へ遷移する。
      toast.warning("ログアウト中にサーバーエラーが発生しました。ブラウザを閉じてセッションを終了することを推奨します。");
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
      saveClinicToStorage(clinicId);
      // 2. フルリロードで全データ（React Query + React Router loader）を新クリニックで再取得
      window.location.reload();
    },
    [user, currentClinicId],
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
      setUser(result.user);
    }
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      currentClinicId,
      isAuthenticated: user !== null,
      isLoading: false, // In React 19 use() pattern, if we are here, we are not loading.
      isSwitchingClinic,
      login,
      logout,
      switchClinic,
      hasPermission,
      refreshPermissions,
    }),
    [
      user,
      currentClinicId,
      isSwitchingClinic,
      login,
      logout,
      switchClinic,
      hasPermission,
      refreshPermissions,
    ],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
