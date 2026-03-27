import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  useMemo,
} from "react";
import type { ReactNode } from "react";
import type { AuthContextValue, AuthUser, ResourceAction } from "../types";
import { login as loginApi } from "../api/login";
import { logout as logoutApi } from "../api/logout";
import { refreshToken } from "../api/refresh-token";

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

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [currentClinicId, setCurrentClinicId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isSwitchingClinic, setIsSwitchingClinic] = useState(false);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const result = await refreshToken();
        if (cancelled) return;
        if (result) {
          setUser(result.user);
          const storedClinic = readClinicFromStorage();
          const validClinic = result.user.clinics.some((c) => c.clinicId === storedClinic);
          setCurrentClinicId(validClinic ? storedClinic : result.user.mainClinicId);
        }
      } catch {
        /* refreshToken は内部で catch して null を返すが、念のため */
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

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
    (resource: string, action: ResourceAction): boolean => {
      if (!user || !currentClinicId) return false;
      // system_admin / clinic_admin はバイパス（BEも全権限 true で返すが念のため）
      if (user.userType === "system_admin" || user.userType === "clinic_admin") return true;
      const clinicPerms = user.permissions[currentClinicId];
      if (!clinicPerms) return false;
      const resourcePerms = clinicPerms[resource];
      if (!resourcePerms) return false;
      return resourcePerms[action] === true;
    },
    [user, currentClinicId],
  );

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      currentClinicId,
      isAuthenticated: user !== null,
      isLoading,
      isSwitchingClinic,
      login,
      logout,
      switchClinic,
      hasPermission,
    }),
    [
      user,
      currentClinicId,
      isLoading,
      isSwitchingClinic,
      login,
      logout,
      switchClinic,
      hasPermission,
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
