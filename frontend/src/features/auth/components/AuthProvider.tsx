import { useState, useCallback, useMemo, useEffect, useRef } from "react";
import type { ReactNode } from "react";
import Axios from "axios";
import { useQueryClient } from "@tanstack/react-query";
import { useLocation, useNavigate } from "react-router";
import { toast } from "sonner";
import type { AuthContextValue, AuthUser, Resource, ResourceAction } from "@/types/auth";
import { AuthContext } from "@/hooks/auth-context";
import { paths } from "@/config/paths";
import { isLoginPublicPath, isPasswordRecoveryPublicPath } from "@/lib/auth-route-policy";
import {
  CURRENT_CLINIC_STORAGE_KEY,
  getStoredClinicId,
  setStoredClinicId,
} from "@/lib/current-clinic";
import { parseInternalPath } from "@/lib/internal-navigation";
import { ME_QUERY_KEY } from "@/lib/query-keys";
import { axios } from "@/lib/axios";
import { attachClinicSelectionInterceptors } from "@/lib/clinic-selection-axios";
import {
  areClinicWritesPaused,
  beginClinicSelectionLogout,
  cancelPendingClinicSelectionRecovery,
  clearClinicSelectionRecovery,
  getClinicSelectionBlockReason,
  rearmAutomaticClinicSelectionRecoveryAttempt,
  recoverClinicSelectionOnce,
} from "@/lib/clinic-selection-recovery";
import { login as loginApi } from "../api/login";
import { logout as logoutApi } from "../api/logout";
import { refreshToken } from "../api/refresh-token";
import { restoreSession as restoreSessionApi } from "../api/restore-session";
import type { SessionRestoreResult } from "../api/restore-session";
import { useGetMe } from "../api/get-me";
import { SessionPending } from "@/components/shared/auth/SessionPending";
import { ClinicSelectionBlockedScreen } from "./ClinicSelectionBlockedScreen";
import { SessionRestoreError } from "./SessionRestoreError";

attachClinicSelectionInterceptors(axios);

const RESTORE_DEADLINE_MS = 8000;

type RestorePhase = "pending" | "recovering" | "error" | "manual-login" | "ready";

/* セッション情報は httpOnly Cookie で管理するため localStorage への保存は不要。
 * 選択中のクリニック ID のみ localStorage に残す（権限情報ではないためリスク低） */
function saveClinicToStorage(clinicId: string): boolean {
  try {
    const ok = setStoredClinicId(clinicId);
    if (!ok && import.meta.env.DEV) {
      console.warn("[auth] failed to save clinic to localStorage");
    }
    return ok;
  } catch (error) {
    if (import.meta.env.DEV) {
      console.warn("[auth] failed to save clinic to localStorage", error);
    }
    return false;
  }
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
    <AuthProviderSession
      key={sessionKey}
      restoreSession={restoreSession}
      suppressBlockedScreen={passwordRecovery}
    >
      {children}
    </AuthProviderSession>
  );
}

interface AuthProviderSessionProps extends AuthProviderProps {
  restoreSession: boolean;
  suppressBlockedScreen: boolean;
}

function AuthProviderSession({
  children,
  restoreSession,
  suppressBlockedScreen,
}: AuthProviderSessionProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [user, setUser] = useState<AuthUser | null>(null);
  const [currentClinicId, setCurrentClinicId] = useState<string | null>(null);
  const [isInitialized, setIsInitialized] = useState(!restoreSession);
  const [restorePhase, setRestorePhase] = useState<RestorePhase>(
    restoreSession ? "pending" : "ready",
  );
  const [restoreErrorKind, setRestoreErrorKind] = useState<"transport" | "restricted">("transport");
  const restorePhaseRef = useRef<RestorePhase>(restoreSession ? "pending" : "ready");
  const restoreGenerationRef = useRef(0);
  const restoreControllerRef = useRef<AbortController | null>(null);
  const flightRef = useRef<Promise<void> | null>(null);
  const flightGenerationRef = useRef<number | null>(null);
  const deadlineTimerRef = useRef<number | null>(null);
  const applyEnabledRef = useRef(true);
  const mountIdRef = useRef(0);

  const hydrateUser = useCallback(
    (next: AuthUser) => {
      queryClient.setQueryData(ME_QUERY_KEY, next);
      setUser(next);
    },
    [queryClient],
  );

  const clearDeadline = useCallback(() => {
    if (deadlineTimerRef.current !== null) {
      window.clearTimeout(deadlineTimerRef.current);
      deadlineTimerRef.current = null;
    }
  }, []);

  const invalidateRestore = useCallback(() => {
    restoreGenerationRef.current += 1;
    restoreControllerRef.current?.abort();
    restoreControllerRef.current = null;
    flightRef.current = null;
    flightGenerationRef.current = null;
    clearDeadline();
    cancelPendingClinicSelectionRecovery();
  }, [clearDeadline]);

  const applyRestoreResult = useCallback(
    (generation: number, result: SessionRestoreResult) => {
      if (!applyEnabledRef.current) return;
      if (generation !== restoreGenerationRef.current) return;
      if (result.kind === "cancel") return;
      if (result.kind === "verified200") {
        clearDeadline();
        const storedClinic = getStoredClinicId();
        const validClinic = result.user.clinics.some((clinic) => clinic.clinicId === storedClinic);
        const nextClinicId =
          validClinic && storedClinic !== null ? storedClinic : result.user.mainClinicId;
        if (areClinicWritesPaused() && !validClinic && !saveClinicToStorage(nextClinicId)) {
          restorePhaseRef.current = "error";
          setRestoreErrorKind("transport");
          setRestorePhase("error");
          return;
        }
        clearClinicSelectionRecovery();
        hydrateUser(result.user);
        setCurrentClinicId(nextClinicId);
        restorePhaseRef.current = "ready";
        setRestorePhase("ready");
        setIsInitialized(true);
        return;
      }
      if (result.kind === "anonymous401") {
        clearDeadline();
        setUser(null);
        setCurrentClinicId(null);
        restorePhaseRef.current = "ready";
        setRestorePhase("ready");
        setIsInitialized(true);
        return;
      }
      if (result.kind === "restricted403") {
        if (result.reason === "clinic_selection_unavailable") {
          // Keep the original 8s wall budget for at-most-one clinic recovery.
          restorePhaseRef.current = "recovering";
          setRestorePhase("recovering");
          void recoverClinicSelectionOnce();
          return;
        }
        clearDeadline();
        restorePhaseRef.current = "error";
        setRestoreErrorKind("restricted");
        setRestorePhase("error");
        return;
      }
      clearDeadline();
      restorePhaseRef.current = "error";
      setRestoreErrorKind("transport");
      setRestorePhase("error");
    },
    [clearDeadline, hydrateUser],
  );

  const runRestore = useCallback(() => {
    if (!restoreSession) return;
    const generation = restoreGenerationRef.current;
    if (flightRef.current !== null && flightGenerationRef.current === generation) {
      return;
    }
    const controller = new AbortController();
    restoreControllerRef.current = controller;
    flightGenerationRef.current = generation;
    deadlineTimerRef.current = window.setTimeout(() => {
      if (generation !== restoreGenerationRef.current) return;
      if (
        restorePhaseRef.current === "ready" ||
        restorePhaseRef.current === "error" ||
        restorePhaseRef.current === "manual-login"
      ) {
        return;
      }
      // Terminal clinic block already has usable restriction UI + logout; do not
      // replace it with transport SessionRestoreError. Hung recovery (block none)
      // still fails closed to the recoverable error shell.
      if (getClinicSelectionBlockReason() !== "none") {
        deadlineTimerRef.current = null;
        return;
      }
      restoreGenerationRef.current += 1;
      controller.abort();
      flightRef.current = null;
      flightGenerationRef.current = null;
      deadlineTimerRef.current = null;
      cancelPendingClinicSelectionRecovery();
      restorePhaseRef.current = "error";
      setRestoreErrorKind("transport");
      setRestorePhase("error");
    }, RESTORE_DEADLINE_MS);

    const flight = restoreSessionApi({ signal: controller.signal }).then((result) => {
      applyRestoreResult(generation, result);
    });
    flightRef.current = flight;
  }, [applyRestoreResult, restoreSession]);

  // recovery/session 境界をまたぐと key により remount され、最新の Cookie 状態を
  // 取得する。同一 mount では StrictMode の effect 再実行時も 1 回だけ呼び出す。
  // True unmount (no remount microtask) hard-invalidates so late results cannot write
  // the shared query cache or start clinic recovery after password-recovery navigation.
  useEffect(() => {
    if (!restoreSession) return;
    const mountId = mountIdRef.current + 1;
    mountIdRef.current = mountId;
    applyEnabledRef.current = true;
    runRestore();
    return () => {
      applyEnabledRef.current = false;
      queueMicrotask(() => {
        if (mountIdRef.current === mountId) {
          invalidateRestore();
        }
      });
    };
  }, [invalidateRestore, restoreSession, runRestore]);

  // /me のキャッシュ（起動時 hydrate）でユーザー情報を同期する。
  // staleTime 経過だけでは再取得しない。定期ポーリングはしない。
  // 権限変更の反映は refreshPermissions / ME_QUERY_KEY 無効化。
  const { data: meData } = useGetMe(user !== null);
  const currentUser = user === null ? null : (meData ?? user);

  // FE5-2: マルチタブ穴 — 他タブでクリニックが切り替わった場合、
  // このタブは storage イベントを検知するまで旧クリニック画面のまま。
  // axios はリクエスト毎に localStorage の clinic を読むため、以後の書き込みが
  // 誤テナントで永続化されうる。他タブ由来の変更を検知したらフルリロードで揃える。
  useEffect(() => {
    function handleStorage(event: StorageEvent): void {
      if (event.key === CURRENT_CLINIC_STORAGE_KEY) {
        invalidateRestore();
        window.location.reload();
      }
    }
    window.addEventListener("storage", handleStorage);
    return () => window.removeEventListener("storage", handleStorage);
  }, [invalidateRestore]);

  const login = useCallback(
    async (email: string, password: string) => {
      invalidateRestore();
      const generation = restoreGenerationRef.current;
      const result = await loginApi(email, password);
      if (!applyEnabledRef.current || generation !== restoreGenerationRef.current) {
        throw new Axios.CanceledError("stale login result");
      }
      const stored = saveClinicToStorage(result.user.mainClinicId);
      if (!stored) {
        toast.error("クリニックの切替に失敗しました。ブラウザのストレージ設定を確認してください。");
        throw new Error("failed to save clinic selection");
      }
      hydrateUser(result.user);
      setCurrentClinicId(result.user.mainClinicId);
      clearClinicSelectionRecovery();
      restorePhaseRef.current = "ready";
      setRestorePhase("ready");
      setIsInitialized(true);
    },
    [hydrateUser, invalidateRestore],
  );

  const logout = useCallback(async () => {
    invalidateRestore();
    beginClinicSelectionLogout();
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
      clearClinicSelectionRecovery();
      queryClient.clear();
      restorePhaseRef.current = "ready";
      setRestorePhase("ready");
      setIsInitialized(true);
    }
  }, [invalidateRestore, queryClient]);

  const switchClinic = useCallback(
    (clinicId: string) => {
      if (!currentUser) return;
      if (clinicId === currentClinicId) return;
      const isMember = currentUser.clinics.some((c) => c.clinicId === clinicId);
      if (!isMember) return;
      invalidateRestore();
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
    [currentUser, currentClinicId, queryClient, invalidateRestore],
  );

  const hasPermission = useCallback(
    (resource: Resource, action: ResourceAction): boolean => {
      if (!currentUser) return false;
      // system_admin はバイパス（BEも全権限 true で返すが念のため）
      if (currentUser.isSystemAdmin) return true;
      const resourcePerms = currentUser.permissions[resource];
      if (!resourcePerms) return false;
      return resourcePerms[action] === true;
    },
    [currentUser],
  );

  const refreshPermissions = useCallback(async () => {
    const result = await refreshToken();
    if (result) {
      hydrateUser(result.user);
    }
  }, [hydrateUser]);

  const handleRetryRestore = useCallback(() => {
    if (restorePhaseRef.current === "pending") {
      return;
    }
    invalidateRestore();
    rearmAutomaticClinicSelectionRecoveryAttempt();
    restorePhaseRef.current = "pending";
    setRestorePhase("pending");
    setIsInitialized(false);
    applyEnabledRef.current = true;
    runRestore();
  }, [invalidateRestore, runRestore]);

  const handleSwitchToLogin = useCallback(() => {
    invalidateRestore();
    restorePhaseRef.current = "manual-login";
    setRestorePhase("manual-login");
    const loginPath = paths.auth.login.getHref();
    const sanitized = parseInternalPath(`${location.pathname}${location.search}`);
    const fromQuery =
      sanitized !== null && sanitized !== loginPath && !isLoginPublicPath(location.pathname)
        ? `?from=${encodeURIComponent(sanitized)}`
        : "";
    navigate(`${loginPath}${fromQuery}`, { replace: true });
  }, [invalidateRestore, location.pathname, location.search, navigate]);

  const value = useMemo<AuthContextValue>(
    () => ({
      user: currentUser,
      currentClinicId,
      isAuthenticated: currentUser !== null,
      isLoading: !isInitialized,
      login,
      logout,
      switchClinic,
      hasPermission,
      refreshPermissions,
    }),
    [
      currentUser,
      currentClinicId,
      isInitialized,
      login,
      logout,
      switchClinic,
      hasPermission,
      refreshPermissions,
    ],
  );

  if (restorePhase === "error") {
    return (
      <SessionRestoreError
        kind={restoreErrorKind}
        onRetry={handleRetryRestore}
        onSwitchToLogin={handleSwitchToLogin}
      />
    );
  }

  if (restorePhase === "manual-login") {
    return (
      <AuthContext.Provider value={value}>
        {isLoginPublicPath(location.pathname) ? children : null}
      </AuthContext.Provider>
    );
  }

  const blockedScreen =
    suppressBlockedScreen || restorePhase === "manual-login" ? null : (
      <ClinicSelectionBlockedScreen onLogout={logout} />
    );

  if (!isInitialized) {
    return (
      <AuthContext.Provider value={value}>
        {blockedScreen}
        <SessionPending />
      </AuthContext.Provider>
    );
  }

  return (
    <AuthContext.Provider value={value}>
      {blockedScreen}
      {children}
    </AuthContext.Provider>
  );
}
