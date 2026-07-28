import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { ResourceCheckups, ResourceMedicalRecords } from "@/types/generated/models";
import { clinicalCareRoutes } from "./clinical-care-routes";

const mockCheckupChildMount = vi.hoisted(() => vi.fn());

vi.mock("@/features/checkups", () => ({
  CheckupsList: () => <div>健診一覧</div>,
  CheckupPetSelection: () => {
    mockCheckupChildMount("select-pet");
    return <div>健診ペット選択</div>;
  },
  CheckupForm: () => {
    mockCheckupChildMount("new");
    return <div>健診フォーム</div>;
  },
}));

interface MedicalRecordPermissions {
  canView?: boolean;
  canCreate?: boolean;
  canEdit?: boolean;
}

function renderCheckupRoute(
  path: string,
  { canView = false, canCreate = false, canEdit = false }: MedicalRecordPermissions = {},
) {
  const hasPermission: AuthContextValue["hasPermission"] = vi.fn(
    (resource, action) =>
      (resource === ResourceCheckups && action === "view") ||
      (resource === ResourceMedicalRecords && action === "view" && canView) ||
      (resource === ResourceMedicalRecords && action === "create" && canCreate) ||
      (resource === ResourceMedicalRecords && action === "edit" && canEdit),
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
  const checkupsRoute = clinicalCareRoutes.find((route) => route.path === "/checkups");
  if (!checkupsRoute) throw new Error("/checkups route is missing");
  const router = createMemoryRouter([checkupsRoute], { initialEntries: [path] });

  render(
    <AuthContext.Provider value={auth}>
      <RouterProvider router={router} />
    </AuthContext.Provider>,
  );

  return { hasPermission };
}

describe("clinicalCareRoutes checkup create permission", () => {
  beforeEach(() => {
    mockCheckupChildMount.mockClear();
  });

  it.each(["/checkups/select-pet", "/checkups/new"])(
    "%s はmedical-records:create権限がなければmountしない",
    async (path) => {
      const { hasPermission } = renderCheckupRoute(path);

      expect(
        await screen.findByRole("heading", {
          level: 1,
          name: "アクセス権限がありません",
        }),
      ).toBeInTheDocument();
      expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "create");
      expect(mockCheckupChildMount).not.toHaveBeenCalled();
    },
  );

  it.each(["/checkups/select-pet", "/checkups/new"])(
    "%s はmedical-records:createのみではmountしない",
    async (path) => {
      const { hasPermission } = renderCheckupRoute(path, {
        canView: true,
        canCreate: true,
        canEdit: false,
      });

      expect(
        await screen.findByRole("heading", {
          level: 1,
          name: "アクセス権限がありません",
        }),
      ).toBeInTheDocument();
      expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "create");
      expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "edit");
      expect(mockCheckupChildMount).not.toHaveBeenCalled();
    },
  );

  it.each([
    ["/checkups/select-pet", "select-pet", "健診ペット選択"],
    ["/checkups/new", "new", "健診フォーム"],
  ])(
    "%s はmedical-records:create/editの両方があればmountする",
    async (path, expectedMount, expectedText) => {
      const { hasPermission } = renderCheckupRoute(path, {
        canView: true,
        canCreate: true,
        canEdit: true,
      });

      expect(await screen.findByText(expectedText)).toBeInTheDocument();
      expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "create");
      expect(hasPermission).toHaveBeenCalledWith(ResourceMedicalRecords, "edit");
      expect(mockCheckupChildMount).toHaveBeenCalledWith(expectedMount);
    },
  );
});
