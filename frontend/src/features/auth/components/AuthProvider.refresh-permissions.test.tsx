/**
 * EMR-532 F2b: AuthProvider.refreshPermissions の契約
 *
 * refreshToken() は /v1/me の失敗（401/403/5xx/通信断）を null に丸める。
 * 権限キャッシュを更新できなかったのに resolve すると、呼出側は古い権限を
 * 「現在の権限」として扱ってしまう（fail-open）。更新できないときは reject する。
 */
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act, waitFor } from "@testing-library/react";
import { Suspense, useEffect } from "react";
import { MemoryRouter } from "react-router";

const { mockQueryClient, MOCK_STAFF, FRESH_STAFF, restoreSessionMock, refreshTokenMock } =
  vi.hoisted(() => {
    const staff = {
      id: "user-staff",
      email: "staff@example.com",
      displayName: "田中 太郎",
      isSystemAdmin: false,
      mainClinicId: "clinic-a",
      clinic: null as null,
      clinics: [{ clinicId: "clinic-a", clinicName: "八王子院", isMain: true }],
      permissions: {} as Record<string, never>,
    };
    return {
      mockQueryClient: { clear: vi.fn(), setQueryData: vi.fn() },
      MOCK_STAFF: staff,
      FRESH_STAFF: { ...staff, displayName: "更新後" },
      restoreSessionMock: vi.fn(),
      refreshTokenMock: vi.fn(),
    };
  });

vi.mock("../api/refresh-token", () => ({ refreshToken: refreshTokenMock }));
vi.mock("../api/restore-session", () => ({ restoreSession: restoreSessionMock }));
vi.mock("../api/get-me", () => ({ useGetMe: vi.fn().mockReturnValue({ data: undefined }) }));
vi.mock("../api/logout", () => ({ logout: vi.fn().mockResolvedValue(undefined) }));
vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() } }));
vi.mock("@tanstack/react-query", async (importActual) => {
  const actual = await importActual<typeof import("@tanstack/react-query")>();
  return { ...actual, useQueryClient: () => mockQueryClient };
});

// ----- import after vi.mock -----
import { AuthProvider } from "./AuthProvider";
import { useAuth } from "@/hooks/use-auth";
import type { AuthContextValue } from "@/types/auth";

let latestAuth: AuthContextValue | null = null;
const captureAuth = (auth: AuthContextValue): void => {
  latestAuth = auth;
};

function AuthProbe({ onAuth }: { onAuth: (auth: AuthContextValue) => void }) {
  const auth = useAuth();
  useEffect(() => {
    onAuth(auth);
  }, [auth, onAuth]);
  return <span data-testid="probe-ready">ready</span>;
}

async function renderProvider() {
  await act(async () => {
    render(
      <MemoryRouter>
        <Suspense fallback={<div data-testid="loading">loading</div>}>
          <AuthProvider>
            <AuthProbe onAuth={captureAuth} />
          </AuthProvider>
        </Suspense>
      </MemoryRouter>,
    );
  });
  await waitFor(() => expect(screen.getByTestId("probe-ready")).toBeInTheDocument());
}

beforeEach(() => {
  latestAuth = null;
  mockQueryClient.clear.mockClear();
  mockQueryClient.setQueryData.mockClear();
  restoreSessionMock.mockReset().mockResolvedValue({ kind: "verified200", user: MOCK_STAFF });
  refreshTokenMock.mockReset();
});

describe("EMR-532: refreshPermissions", () => {
  it("/me を取得できなかったとき（refreshToken が null）は reject し、user を更新しない", async () => {
    refreshTokenMock.mockResolvedValue(null);
    await renderProvider();
    mockQueryClient.setQueryData.mockClear(); // 起動時 hydrate を除外する。

    await expect(latestAuth!.refreshPermissions()).rejects.toThrow();

    expect(refreshTokenMock).toHaveBeenCalledTimes(1);
    expect(mockQueryClient.setQueryData).not.toHaveBeenCalled();
  });

  it("/me を取得できたときは resolve し、ユーザーを hydrate する", async () => {
    refreshTokenMock.mockResolvedValue({ user: FRESH_STAFF });
    await renderProvider();
    mockQueryClient.setQueryData.mockClear();

    await act(async () => {
      await latestAuth!.refreshPermissions();
    });

    expect(mockQueryClient.setQueryData).toHaveBeenCalledTimes(1);
    expect(mockQueryClient.setQueryData.mock.calls[0]![1]).toEqual(FRESH_STAFF);
  });
});
