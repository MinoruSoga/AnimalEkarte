import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { describe, expect, it, vi } from "vitest";

import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { ResourceHospitalSettings, ResourceLstepAnalytics } from "@/types/generated/models";

import { operationsRoutes } from "./operations-routes";

const { deliveryApiMock } = vi.hoisted(() => ({
  deliveryApiMock: vi.fn(),
}));

vi.mock("@/features/lstep", () => ({
  LstepDeliveryMonitorPage: () => {
    deliveryApiMock();
    return <p>delivery monitor page</p>;
  },
}));

function auth(hasPermission: AuthContextValue["hasPermission"]): AuthContextValue {
  return {
    user: null,
    currentClinicId: "clinic-test-1",
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission,
    refreshPermissions: async () => {},
  };
}

describe("/lstep/delivery-monitor — RBAC guard", () => {
  it("hospital-settings:viewだけでは拒否し、pageをmountしない", async () => {
    const hasPermission = vi.fn<AuthContextValue["hasPermission"]>(
      (resource, action) => resource === ResourceHospitalSettings && action === "view",
    );
    const router = createMemoryRouter(operationsRoutes, {
      initialEntries: ["/lstep/delivery-monitor"],
    });

    render(
      <AuthContext.Provider value={auth(hasPermission)}>
        <RouterProvider router={router} />
      </AuthContext.Provider>,
    );

    expect(await screen.findByText("アクセス権限がありません")).toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(ResourceLstepAnalytics, "view");
    expect(deliveryApiMock).not.toHaveBeenCalled();
  });

  it("lstep-analytics:viewのみで page を mount する（親 HospitalSettings AND を課さない）", async () => {
    const hasPermission = vi.fn<AuthContextValue["hasPermission"]>(
      (resource, action) => resource === ResourceLstepAnalytics && action === "view",
    );
    const router = createMemoryRouter(operationsRoutes, {
      initialEntries: ["/lstep/delivery-monitor"],
    });

    render(
      <AuthContext.Provider value={auth(hasPermission)}>
        <RouterProvider router={router} />
      </AuthContext.Provider>,
    );

    expect(await screen.findByText("delivery monitor page")).toBeInTheDocument();
    expect(deliveryApiMock).toHaveBeenCalled();
    expect(hasPermission).toHaveBeenCalledWith(ResourceLstepAnalytics, "view");
    expect(hasPermission).not.toHaveBeenCalledWith(ResourceHospitalSettings, "view");
  });
});
