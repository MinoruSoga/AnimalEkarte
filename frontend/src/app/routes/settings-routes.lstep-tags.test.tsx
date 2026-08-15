import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import {
  ResourceHospitalSettings,
  ResourceLstepAnalytics,
} from "@/types/generated/models";
import { settingsRoute } from "./settings-routes";

const CLINIC_ID = "clinic-test-1";

describe("/settings/lstep/tags — RBAC guard", () => {
  it("hospital-settings:view だけでは拒否し、lstep-analytics:view を要求する", () => {
    const hasPermission = vi.fn<AuthContextValue["hasPermission"]>(
      (resource, action) => resource === ResourceHospitalSettings && action === "view"
    );
    const authContext: AuthContextValue = {
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
    const lstepTagsRoute = settingsRoute.children?.find(
      (route) => route.path === "lstep/tags"
    );

    expect(lstepTagsRoute).toBeDefined();
    render(
      <AuthContext.Provider value={authContext}>
        {lstepTagsRoute?.element}
      </AuthContext.Provider>
    );

    expect(hasPermission).toHaveBeenCalledWith(ResourceLstepAnalytics, "view");
    expect(screen.getByText("アクセス権限がありません")).toBeInTheDocument();
  });
});
