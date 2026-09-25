import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { describe, expect, it, vi } from "vitest";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { clinicalCareRoutes } from "./clinical-care-routes";

/** Wire resource id for lab-import (avoids @/types/generated/models import; TASK-444-S1). */
const LAB_IMPORT_RESOURCE = "lab-import";

vi.mock("@/features/lab-device", () => ({
  LabDeviceBoard: () => <div>検査受信ボード</div>,
}));

interface LabDevicePermissions {
  canView?: boolean;
  canCreate?: boolean;
}

function renderLabDeviceRoute({ canView = false, canCreate = false }: LabDevicePermissions = {}) {
  const hasPermission: AuthContextValue["hasPermission"] = vi.fn(
    (resource, action) =>
      (resource === LAB_IMPORT_RESOURCE && action === "view" && canView) ||
      (resource === LAB_IMPORT_RESOURCE && action === "create" && canCreate),
  );
  const auth: AuthContextValue = {
    user: null,
    currentClinicId: "clinic-1",
    isAuthenticated: true,
    isLoading: false,
    login: async () => undefined,
    logout: async () => undefined,
    switchClinic: () => undefined,
    hasPermission,
    refreshPermissions: async () => undefined,
  };
  const labDeviceRoute = clinicalCareRoutes.find((route) => route.path === "/lab-device");
  if (!labDeviceRoute) throw new Error("/lab-device route is missing");
  const router = createMemoryRouter([labDeviceRoute], { initialEntries: ["/lab-device"] });

  render(
    <AuthContext.Provider value={auth}>
      <RouterProvider router={router} />
    </AuthContext.Provider>,
  );

  return { hasPermission };
}

describe("clinicalCareRoutes /lab-device permission", () => {
  it("lab-import:view が無ければボードを mount せず AccessDenied を出す", async () => {
    const { hasPermission } = renderLabDeviceRoute({ canView: false, canCreate: false });

    expect(
      await screen.findByRole("heading", {
        level: 1,
        name: "アクセス権限がありません",
      }),
    ).toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(LAB_IMPORT_RESOURCE, "view");
    expect(screen.queryByText("検査受信ボード")).not.toBeInTheDocument();
  });

  it("view のみで create が無くてもボードを mount する（受信不可はボード内バナーが説明する）", async () => {
    const { hasPermission } = renderLabDeviceRoute({ canView: true, canCreate: false });

    expect(await screen.findByText("検査受信ボード")).toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(LAB_IMPORT_RESOURCE, "view");
    expect(screen.queryByRole("heading", { name: "アクセス権限がありません" })).toBeNull();
  });

  it("view+create があればボードを mount する", async () => {
    renderLabDeviceRoute({ canView: true, canCreate: true });

    expect(await screen.findByText("検査受信ボード")).toBeInTheDocument();
  });
});
