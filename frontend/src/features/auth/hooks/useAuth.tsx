import {
  createContext,
  useContext,
  useState,
  useCallback,
  useEffect,
  useMemo,
} from "react";
import type { ReactNode } from "react";
import type { AuthContextValue, AuthUser, Permission } from "../types";
import { login as loginApi } from "../api/login";
import { logout as logoutApi } from "../api/logout";
import { refreshToken } from "../api/refresh-token";

/* client-localstorage-schema: バージョン付きキーでスキーマ変更時の破損を防ぐ */
const STORAGE_KEY_USER = "auth_user:v1";
const STORAGE_KEY_CLINIC = "auth_current_clinic:v1";

/** client-localstorage-schema: 最小フィールドのみ保存。権限情報はセッション復元時にモックDBから再取得 */
interface StoredUser {
  id: string;
  email: string;
  displayName: string;
  userType: string;
  mainClinicId: string;
}

function saveUserToStorage(user: StoredUser): void {
  try {
    localStorage.setItem(
      STORAGE_KEY_USER,
      JSON.stringify({
        id: user.id,
        email: user.email,
        displayName: user.displayName,
        userType: user.userType,
        mainClinicId: user.mainClinicId,
      }),
    );
  } catch {
    /* incognito / storage quota exceeded — 無視してセッションのみで継続 */
  }
}

function readStorageItem(key: string): string | null {
  try {
    return localStorage.getItem(key);
  } catch {
    return null;
  }
}

function removeStorageItems(...keys: string[]): void {
  try {
    keys.forEach((k) => localStorage.removeItem(k));
  } catch {
    /* ignore */
  }
}

function saveClinicToStorage(clinicId: string): void {
  try {
    localStorage.setItem(STORAGE_KEY_CLINIC, clinicId);
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
    refreshToken().then((result) => {
      if (cancelled) return;
      if (result) {
        setUser(result.user);
        const storedClinic = readStorageItem(STORAGE_KEY_CLINIC);
        const validClinic = result.user.clinics.some((c) => c.clinicId === storedClinic);
        setCurrentClinicId(validClinic ? storedClinic : result.user.mainClinicId);
      }
      setIsLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const result = await loginApi(email, password);
    setUser(result.user);
    setCurrentClinicId(result.user.mainClinicId);
    saveUserToStorage(result.user);
    saveClinicToStorage(result.user.mainClinicId);
  }, []);

  const logout = useCallback(async () => {
    await logoutApi();
    setUser(null);
    setCurrentClinicId(null);
    removeStorageItems(STORAGE_KEY_USER, STORAGE_KEY_CLINIC);
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
    (permission: Permission): boolean => {
      if (!user || !currentClinicId) return false;
      if (user.userType === "system_admin") return true;
      if (user.userType === "clinic_admin") {
        return user.clinics.some((c) => c.clinicId === currentClinicId);
      }
      const clinicPerms = user.permissions[currentClinicId];
      if (!clinicPerms) return false;
      return clinicPerms.includes(permission);
    },
    [user, currentClinicId],
  );

  const hasAnyPermission = useCallback(
    (permissions: readonly Permission[]): boolean => {
      return permissions.some((p) => hasPermission(p));
    },
    [hasPermission],
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
      hasAnyPermission,
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
      hasAnyPermission,
    ],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return ctx;
}
