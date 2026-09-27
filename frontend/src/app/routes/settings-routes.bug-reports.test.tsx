import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { settingsRoute } from "./settings-routes";

const CLINIC_ID = "clinic-test-1";

function buildAuthContext(hasPermission: AuthContextValue["hasPermission"]): AuthContextValue {
  return {
    user: null,
    currentClinicId: CLINIC_ID,
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission,
    refreshPermissions: async () => {},
  };
}

describe("/settings/bug-reports — RBAC guard", () => {
  it("hospital-settings:view がない場合は拒否する", () => {
    const hasPermission = vi.fn<AuthContextValue["hasPermission"]>(
      (resource, action) => resource === "master-staff" && action === "view",
    );
    const bugReportsRoute = settingsRoute.children?.find((route) => route.path === "bug-reports");

    expect(bugReportsRoute).toBeDefined();
    render(
      <AuthContext.Provider value={buildAuthContext(hasPermission)}>
        {bugReportsRoute?.element}
      </AuthContext.Provider>,
    );

    expect(hasPermission).toHaveBeenCalledWith("hospital-settings", "view");
    expect(screen.getByText("アクセス権限がありません")).toBeInTheDocument();
  });
});
