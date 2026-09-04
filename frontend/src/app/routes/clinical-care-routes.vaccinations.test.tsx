import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider, useParams } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { clinicalCareRoutes } from "./clinical-care-routes";

/** Wire resource id for vaccinations (avoids @/types/generated/models import). */
const VACCINATIONS_RESOURCE = "vaccinations";

const mockVaccinationChildMount = vi.hoisted(() => vi.fn());

vi.mock("@/features/vaccinations", () => ({
  VaccinationList: () => <div>予防接種一覧</div>,
  VaccinationPetSelection: () => {
    mockVaccinationChildMount("select-pet");
    return <div>予防接種ペット選択</div>;
  },
  VaccinationForm: () => {
    const { id } = useParams();
    mockVaccinationChildMount(id === undefined ? "new" : `detail:${id}`);
    return <div>{id === undefined ? "予防接種フォーム新規" : `予防接種フォーム:${id}`}</div>;
  },
}));

interface VaccinationPermissions {
  canView?: boolean;
  canCreate?: boolean;
}

function renderVaccinationRoute(
  path: string,
  { canView = false, canCreate = false }: VaccinationPermissions = {},
) {
  const hasPermission: AuthContextValue["hasPermission"] = vi.fn(
    (resource, action) =>
      (resource === VACCINATIONS_RESOURCE && action === "view" && canView) ||
      (resource === VACCINATIONS_RESOURCE && action === "create" && canCreate),
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
  const vaccinationsRoute = clinicalCareRoutes.find((route) => route.path === "/vaccinations");
  if (!vaccinationsRoute) throw new Error("/vaccinations route is missing");
  const router = createMemoryRouter([vaccinationsRoute], { initialEntries: [path] });

  render(
    <AuthContext.Provider value={auth}>
      <RouterProvider router={router} />
    </AuthContext.Provider>,
  );

  return { hasPermission };
}

describe("clinicalCareRoutes vaccination create path (BUG-501)", () => {
  beforeEach(() => {
    mockVaccinationChildMount.mockClear();
  });

  it.each(["/vaccinations/select-pet", "/vaccinations/new"])(
    "%s はvaccinations:create権限がなければmountしない",
    async (path) => {
      const { hasPermission } = renderVaccinationRoute(path, { canView: true });

      expect(
        await screen.findByRole("heading", {
          level: 1,
          name: "アクセス権限がありません",
        }),
      ).toBeInTheDocument();
      expect(hasPermission).toHaveBeenCalledWith(VACCINATIONS_RESOURCE, "create");
      expect(mockVaccinationChildMount).not.toHaveBeenCalled();
    },
  );

  it("/vaccinations/select-pet はvaccinations:createがあればペット選択をmountする", async () => {
    const { hasPermission } = renderVaccinationRoute("/vaccinations/select-pet", {
      canView: true,
      canCreate: true,
    });

    expect(await screen.findByText("予防接種ペット選択")).toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(VACCINATIONS_RESOURCE, "create");
    expect(mockVaccinationChildMount).toHaveBeenCalledWith("select-pet");
  });

  it("/vaccinations/new はselect-petへ常時リダイレクトせず VaccinationForm を新規としてmountする", async () => {
    const { hasPermission } = renderVaccinationRoute("/vaccinations/new?petId=pet-1", {
      canView: true,
      canCreate: true,
    });

    expect(await screen.findByText("予防接種フォーム新規")).toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(VACCINATIONS_RESOURCE, "create");
    expect(mockVaccinationChildMount).toHaveBeenCalledWith("new");
    expect(mockVaccinationChildMount).not.toHaveBeenCalledWith("select-pet");
    expect(mockVaccinationChildMount).not.toHaveBeenCalledWith("detail:new");
  });

  it("/vaccinations/new（petIdなし）も VaccinationForm をmountする（フォーム側がselect-petへ誘導）", async () => {
    renderVaccinationRoute("/vaccinations/new", {
      canView: true,
      canCreate: true,
    });

    expect(await screen.findByText("予防接種フォーム新規")).toBeInTheDocument();
    expect(mockVaccinationChildMount).toHaveBeenCalledWith("new");
    expect(mockVaccinationChildMount).not.toHaveBeenCalledWith("select-pet");
  });

  it("/vaccinations/abc は詳細としてVaccinationFormをmountする", async () => {
    renderVaccinationRoute("/vaccinations/abc", { canView: true });

    expect(await screen.findByText("予防接種フォーム:abc")).toBeInTheDocument();
    expect(mockVaccinationChildMount).toHaveBeenCalledWith("detail:abc");
  });
});
