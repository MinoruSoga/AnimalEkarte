import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { AuthContext } from "@/hooks/auth-context";
import { RequirePermission } from "./RequirePermission";
import type { ResourceAction } from "@/types/auth";

const CLINIC_ID = "clinic-test-1";

function makeAuthCtx(hasPermissionFn: (r: string, a: ResourceAction) => boolean) {
  return {
    user: null,
    currentClinicId: CLINIC_ID,
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: hasPermissionFn,
    refreshPermissions: async () => {},
  };
}

describe("RequirePermission — 振る舞いテスト", () => {
  it("権限あり → children を描画する", () => {
    render(
      <AuthContext.Provider value={makeAuthCtx(() => true)}>
        <RequirePermission resource="accounting">
          <div data-testid="protected-content">protected</div>
        </RequirePermission>
      </AuthContext.Provider>
    );
    expect(screen.getByTestId("protected-content")).toBeInTheDocument();
    expect(screen.queryByText(/アクセス権限がありません/)).not.toBeInTheDocument();
  });

  it("権限なし → AccessDenied を描画する", () => {
    render(
      <AuthContext.Provider value={makeAuthCtx(() => false)}>
        <RequirePermission resource="accounting">
          <div data-testid="protected-content">protected</div>
        </RequirePermission>
      </AuthContext.Provider>
    );
    expect(
      screen.getByRole("heading", { level: 1, name: "アクセス権限がありません" }),
    ).toBeInTheDocument();
    expect(screen.queryByTestId("protected-content")).not.toBeInTheDocument();
  });

  it("権限なし + fallback あり → fallback を描画する", () => {
    render(
      <AuthContext.Provider value={makeAuthCtx(() => false)}>
        <RequirePermission
          resource="accounting"
          fallback={<div data-testid="custom-fallback">custom</div>}
        >
          <div data-testid="protected-content">protected</div>
        </RequirePermission>
      </AuthContext.Provider>
    );
    expect(screen.getByTestId("custom-fallback")).toBeInTheDocument();
    expect(screen.queryByText(/アクセス権限がありません/)).not.toBeInTheDocument();
    expect(screen.queryByTestId("protected-content")).not.toBeInTheDocument();
  });

  it("action 省略時は \"view\" が hasPermission に渡る", () => {
    let capturedAction: ResourceAction | undefined;
    render(
      <AuthContext.Provider
        value={makeAuthCtx((_r, action) => {
          capturedAction = action;
          return true;
        })}
      >
        <RequirePermission resource="accounting">
          <div />
        </RequirePermission>
      </AuthContext.Provider>
    );
    expect(capturedAction).toBe("view");
  });

  it("action=\"create\" 指定時は hasPermission に \"create\" が渡る", () => {
    let capturedAction: ResourceAction | undefined;
    render(
      <AuthContext.Provider
        value={makeAuthCtx((_r, action) => {
          capturedAction = action;
          return true;
        })}
      >
        <RequirePermission resource="accounting" action="create">
          <div />
        </RequirePermission>
      </AuthContext.Provider>
    );
    expect(capturedAction).toBe("create");
  });
});
