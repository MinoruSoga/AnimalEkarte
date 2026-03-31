import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  useMemo,
  use,
} from "react";
import type { ReactNode } from "react";
import type { AuthContextValue, AuthUser, Resource, ResourceAction } from "../types";
import { login as loginApi } from "../api/login";
import { logout as logoutApi } from "../api/logout";
import { refreshToken } from "../api/refresh-token";
import { useGetMe } from "../api/get-me";

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
  } catch {
    /* ignore */
  }
}

function removeClinicFromStorage(): void {
  try {
    localStorage.removeItem(STORAGE_KEY_CLINIC);
  } catch {
    /* ignore */
  }
}

const AuthContext = createContext<AuthContextValue | null>(null);

/**
 * Initial session restoration promise.
 * Module-level で一度だけ作成する。ログイン後は window.location.replace() で
 * フルリロードするため、ログイン後のマウント時には最新の Cookie でセッション復元される。
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
  
  const [isSwitchingClinic, setIsSwitchingClinic] = useState(false);

  // /me の定期ポーリング結果でユーザー情報（権限含む）を同期
  // 認証済みかつローディング完了後のみポーリングを有効化
  const { data: meData } = useGetMe(user !== null);
  useEffect(() => {
    if (meData) {
      setUser(meData);
    }
  }, [meData]);

  const login = useCallback(async (email: string, password: string) => {
    const result = await loginApi(email, password);
    setUser(result.user);
    setCurrentClinicId(result.user.mainClinicId);
    saveClinicToStorage(result.user.mainClinicId);
  }, []);

  const logout = useCallback(async () => {
    await logoutApi();
    setUser(null);
    setCurrentClinicId(null);
    removeClinicFromStorage();
  }, []);

  const switchClinic = useCallback(
    async (clinicId: string) => {
      if (!user) return;
      if (clinicId === currentClinicId) return;
      const isMember =
        user.userType === "system_admin" ||
        user.clinics.some((c) => c.clinicId === clinicId);
      if (!isMember) return;
      setIsSwitchingClinic(true);
      await new Promise<void>((resolve) => setTimeout(resolve, 800));
      setCurrentClinicId(clinicId);
      saveClinicToStorage(clinicId);
      setIsSwitchingClinic(false);
    },
    [user, currentClinicId],
  );

  const hasPermission = useCallback(
    (resource: Resource, action: ResourceAction): boolean => {
      if (!user) return false;
      // system_admin / clinic_admin はバイパス（BEも全権限 true で返すが念のため）
      if (user.userType === "system_admin" || user.userType === "clinic_admin") return true;
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

// eslint-disable-next-line react-refresh/only-export-components -- AuthProvider と useAuth は同一コンテキストを共有するため同一ファイルで定義
export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
