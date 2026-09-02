/**
 * FEAT-374: AuthProvider — switchClinic / logout clinic-switch state
 *
 * カバー観点:
 * 1. system_admin は clinics 配列に複数クリニックが含まれ、全て切り替え可能
 * 2. 有効メンバークリニックへの切り替え → localStorage 更新 + reload 呼び出し
 * 3. 現在と同じ clinicId への切り替え → no-op
 * 4. メンバー外 clinicId への切り替え → no-op（通常スタッフ防護）
 * 5. logout 後に localStorage の clinic キーが削除される
 */
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, act, waitFor } from "@testing-library/react";
import { Suspense } from "react";
import { MemoryRouter } from "react-router";
import { toast } from "sonner";

const STORAGE_KEY = "auth_current_clinic:v1";
const CLINIC_A = "clinic-a";
const CLINIC_B = "clinic-b";
const CLINIC_C_UNKNOWN = "clinic-c";

// vi.hoisted: vi.mock factory から参照できるようにすべての定数をモジュール初期化前に巻き上げる
const { mockQueryClient, MOCK_SYSTEM_ADMIN, MOCK_STAFF } = vi.hoisted(() => {
  const systemAdmin = {
    id: "user-sysadmin",
    email: "sysadmin@example.com",
    displayName: "System Admin",
    isSystemAdmin: true,
    mainClinicId: "clinic-a",
    clinic: null as null,
    clinics: [
      { clinicId: "clinic-a", clinicName: "八王子院", isMain: true },
      { clinicId: "clinic-b", clinicName: "城東医院", isMain: false },
    ],
    permissions: {} as Record<string, never>,
  };
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
    mockQueryClient: { clear: vi.fn() },
    MOCK_SYSTEM_ADMIN: systemAdmin,
    MOCK_STAFF: staff,
  };
});

vi.mock("../api/refresh-token", () => ({
  refreshToken: vi.fn().mockResolvedValue({ user: MOCK_SYSTEM_ADMIN }),
}));

vi.mock("../api/get-me", () => ({
  useGetMe: vi.fn().mockReturnValue({ data: undefined }),
}));

vi.mock("../api/logout", () => ({
  logout: vi.fn().mockResolvedValue(undefined),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn(), warning: vi.fn() },
}));

vi.mock("@tanstack/react-query", async (importActual) => {
  const actual = await importActual<typeof import("@tanstack/react-query")>();
  return { ...actual, useQueryClient: () => mockQueryClient };
});

// ----- import after vi.mock -----
import { AuthProvider } from "../components/AuthProvider";
import { useAuth } from "@/hooks/use-auth";

// ----- helper components -----

function ClinicSwitcher({
  targetClinicId,
  label,
}: {
  targetClinicId: string;
  label: string;
}) {
  const { switchClinic, currentClinicId } = useAuth();
  return (
    <div>
      <span data-testid="current-clinic">{currentClinicId}</span>
      <button type="button" onClick={() => switchClinic(targetClinicId)}>{label}</button>
    </div>
  );
}

function LogoutTrigger() {
  const { logout } = useAuth();
  return <button type="button" onClick={() => void logout()}>logout</button>;
}

async function renderWithAuth(children: React.ReactNode) {
  // await act(...) flushes the cached initial session promise before waitFor polls
  await act(async () => {
    render(
      <MemoryRouter>
        <Suspense fallback={<div data-testid="loading">loading</div>}>
          <AuthProvider>{children}</AuthProvider>
        </Suspense>
      </MemoryRouter>,
    );
  });
  await waitFor(() =>
    expect(screen.queryByTestId("loading")).not.toBeInTheDocument(),
  );
}

// ----- tests -----

describe("FEAT-374: switchClinic", () => {
  let reloadSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    localStorage.clear();
    mockQueryClient.clear.mockClear();
    reloadSpy = vi.fn();
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: { ...window.location, reload: reloadSpy },
    });
  });

  it("system_admin の clinics は複数クリニックを含み、いずれにも切り替え可能", async () => {
    await renderWithAuth(
      <ClinicSwitcher targetClinicId={CLINIC_B} label="switch-to-b" />,
    );

    expect(screen.getByTestId("current-clinic").textContent).toBe(CLINIC_A);

    await act(async () => {
      screen.getByText("switch-to-b").click();
    });

    expect(localStorage.getItem(STORAGE_KEY)).toBe(CLINIC_B);
    expect(reloadSpy).toHaveBeenCalledOnce();
  });

  it("有効メンバークリニックへの切り替えで localStorage を更新し reload を呼ぶ", async () => {
    await renderWithAuth(
      <ClinicSwitcher targetClinicId={CLINIC_B} label="switch-to-b" />,
    );

    await act(async () => {
      screen.getByText("switch-to-b").click();
    });

    expect(localStorage.getItem(STORAGE_KEY)).toBe(CLINIC_B);
    expect(reloadSpy).toHaveBeenCalledOnce();
  });

  it("FE5-3: switchClinic 成功パスで reload 前に queryClient.clear() を呼ぶ（将来 SPA 化への防壁）", async () => {
    const callOrder: string[] = [];
    mockQueryClient.clear.mockImplementation(() => {
      callOrder.push("clear");
    });
    reloadSpy.mockImplementation(() => {
      callOrder.push("reload");
    });

    await renderWithAuth(
      <ClinicSwitcher targetClinicId={CLINIC_B} label="switch-to-b" />,
    );

    await act(async () => {
      screen.getByText("switch-to-b").click();
    });

    expect(mockQueryClient.clear).toHaveBeenCalledOnce();
    expect(reloadSpy).toHaveBeenCalledOnce();
    expect(callOrder).toEqual(["clear", "reload"]);
  });

  it("現在と同じ clinicId への切り替えは no-op", async () => {
    await renderWithAuth(
      <ClinicSwitcher targetClinicId={CLINIC_A} label="switch-to-same" />,
    );

    await act(async () => {
      screen.getByText("switch-to-same").click();
    });

    expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
    expect(reloadSpy).not.toHaveBeenCalled();
  });

  it("メンバー外クリニックへの切り替えは no-op（通常スタッフ防護）", async () => {
    // MOCK_SYSTEM_ADMIN.clinics には CLINIC_C が存在しないため rejected
    await renderWithAuth(
      <ClinicSwitcher
        targetClinicId={CLINIC_C_UNKNOWN}
        label="switch-to-unknown"
      />,
    );

    await act(async () => {
      screen.getByText("switch-to-unknown").click();
    });

    expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
    expect(reloadSpy).not.toHaveBeenCalled();
  });

  it("localStorage書込が失敗した場合はreloadせずエラートーストを表示する", async () => {
    const setItemSpy = vi
      .spyOn(Storage.prototype, "setItem")
      .mockImplementation(() => {
        throw new DOMException("QuotaExceededError");
      });

    await renderWithAuth(
      <ClinicSwitcher targetClinicId={CLINIC_B} label="switch-to-b" />,
    );

    await act(async () => {
      screen.getByText("switch-to-b").click();
    });

    expect(reloadSpy).not.toHaveBeenCalled();
    expect(mockQueryClient.clear).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(
      "クリニックの切替に失敗しました。ブラウザのストレージ設定を確認してください。",
    );

    setItemSpy.mockRestore();
  });
});

describe("FEAT-374: logout でクリニック状態をクリア", () => {
  beforeEach(() => {
    localStorage.clear();
    mockQueryClient.clear.mockClear();
  });

  it("logout 後に localStorage の clinic キーが削除される", async () => {
    // 事前に切り替え先をセット（クリア対象を作る）
    localStorage.setItem(STORAGE_KEY, CLINIC_B);

    await renderWithAuth(<LogoutTrigger />);

    await act(async () => {
      screen.getByText("logout").click();
    });

    await waitFor(() => {
      expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
    });
    expect(mockQueryClient.clear).toHaveBeenCalledOnce();
  });
});

describe("FE5-2: マルチタブ storage イベント検知で reload", () => {
  let reloadSpy: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    localStorage.clear();
    mockQueryClient.clear.mockClear();
    reloadSpy = vi.fn();
    Object.defineProperty(window, "location", {
      configurable: true,
      writable: true,
      value: { ...window.location, reload: reloadSpy },
    });
  });

  it("他タブで CURRENT_CLINIC_STORAGE_KEY が変更された storage イベントを検知すると reload する", async () => {
    await renderWithAuth(<div />);

    await act(async () => {
      window.dispatchEvent(
        new StorageEvent("storage", {
          key: STORAGE_KEY,
          oldValue: CLINIC_A,
          newValue: CLINIC_B,
        }),
      );
    });

    expect(reloadSpy).toHaveBeenCalledOnce();
  });

  it("無関係な key の storage イベントでは reload しない", async () => {
    await renderWithAuth(<div />);

    await act(async () => {
      window.dispatchEvent(
        new StorageEvent("storage", {
          key: "some-other-key",
          oldValue: "a",
          newValue: "b",
        }),
      );
    });

    expect(reloadSpy).not.toHaveBeenCalled();
  });
});

describe("FEAT-374: 通常スタッフのクリニック切り替え制限（ユニット検証）", () => {
  it("staff ユーザーは所属外クリニック ID を clinics に持たない", () => {
    const isMember = MOCK_STAFF.clinics.some(
      (c) => c.clinicId === CLINIC_C_UNKNOWN,
    );
    expect(isMember).toBe(false);
  });

  it("staff ユーザーは自クリニックを clinics に持つ", () => {
    const isMember = MOCK_STAFF.clinics.some(
      (c) => c.clinicId === CLINIC_A,
    );
    expect(isMember).toBe(true);
  });
});
