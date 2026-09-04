import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider, useParams } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { AuthContext } from "@/hooks/auth-context";
import type { AuthContextValue } from "@/types/auth";
import { clinicalCareRoutes } from "./clinical-care-routes";

/** Wire resource id for examinations (avoids @/types/generated/models import; TASK-444-S1). */
const EXAMINATIONS_RESOURCE = "examinations";

const mockExaminationChildMount = vi.hoisted(() => vi.fn());

vi.mock("@/features/examinations", () => ({
  ExaminationsList: () => <div>検査一覧</div>,
  ExaminationPetSelection: () => {
    mockExaminationChildMount("select-pet");
    return <div>検査ペット選択</div>;
  },
  ExaminationForm: () => {
    const { id } = useParams();
    mockExaminationChildMount(id === undefined ? "new" : `detail:${id}`);
    return <div>{id === undefined ? "検査フォーム新規" : `検査フォーム:${id}`}</div>;
  },
}));

interface ExaminationPermissions {
  canView?: boolean;
  canCreate?: boolean;
}

function renderExaminationRoute(
  path: string,
  { canView = false, canCreate = false }: ExaminationPermissions = {},
) {
  const hasPermission: AuthContextValue["hasPermission"] = vi.fn(
    (resource, action) =>
      (resource === EXAMINATIONS_RESOURCE && action === "view" && canView) ||
      (resource === EXAMINATIONS_RESOURCE && action === "create" && canCreate),
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
  const examinationsRoute = clinicalCareRoutes.find((route) => route.path === "/examinations");
  if (!examinationsRoute) throw new Error("/examinations route is missing");
  const router = createMemoryRouter([examinationsRoute], { initialEntries: [path] });

  render(
    <AuthContext.Provider value={auth}>
      <RouterProvider router={router} />
    </AuthContext.Provider>,
  );

  return { hasPermission };
}

describe("clinicalCareRoutes examination create permission", () => {
  beforeEach(() => {
    mockExaminationChildMount.mockClear();
  });

  it.each(["/examinations/select-pet", "/examinations/new"])(
    "%s はexaminations:create権限がなければmountしない",
    async (path) => {
      const { hasPermission } = renderExaminationRoute(path, { canView: true });

      expect(
        await screen.findByRole("heading", {
          level: 1,
          name: "アクセス権限がありません",
        }),
      ).toBeInTheDocument();
      expect(hasPermission).toHaveBeenCalledWith(EXAMINATIONS_RESOURCE, "create");
      expect(mockExaminationChildMount).not.toHaveBeenCalled();
    },
  );

  it("/examinations/select-pet はexaminations:createがあればペット選択をmountする", async () => {
    const { hasPermission } = renderExaminationRoute("/examinations/select-pet", {
      canView: true,
      canCreate: true,
    });

    expect(await screen.findByText("検査ペット選択")).toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(EXAMINATIONS_RESOURCE, "create");
    expect(mockExaminationChildMount).toHaveBeenCalledWith("select-pet");
  });

  it("/examinations/new はexaminations:createでExaminationFormを新規としてmountし :id 扱いしない", async () => {
    const { hasPermission } = renderExaminationRoute("/examinations/new", {
      canView: true,
      canCreate: true,
    });

    expect(await screen.findByText("検査フォーム新規")).toBeInTheDocument();
    expect(hasPermission).toHaveBeenCalledWith(EXAMINATIONS_RESOURCE, "create");
    expect(mockExaminationChildMount).toHaveBeenCalledWith("new");
    expect(mockExaminationChildMount).not.toHaveBeenCalledWith("detail:new");
  });

  it("/examinations/abc は詳細としてExaminationFormをmountする", async () => {
    renderExaminationRoute("/examinations/abc", { canView: true });

    expect(await screen.findByText("検査フォーム:abc")).toBeInTheDocument();
    expect(mockExaminationChildMount).toHaveBeenCalledWith("detail:abc");
  });
});
