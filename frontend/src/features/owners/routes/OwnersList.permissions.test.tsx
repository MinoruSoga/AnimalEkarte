import { act, useLayoutEffect, useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { AuthContext } from "@/hooks/auth-context";
import type { AuthUser, ResourceAction } from "@/types/auth";
import type { Pet } from "@/types";
import { OwnersList } from "./OwnersList";

const permissionMocks = vi.hoisted(() => ({
  canDelete: true,
  confirmDelete: undefined as (() => void) | undefined,
  deleteOwner: vi.fn(),
  revokeDeleteAndConfirm: undefined as (() => void) | undefined,
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: vi.fn((resource: string) => ({
    canView: true,
    canCreate: resource === "owners",
    canEdit: resource === "owners",
    canDelete: resource === "owners" && permissionMocks.canDelete,
  })),
}));

vi.mock("../api/delete-owner", () => ({
  deleteOwner: permissionMocks.deleteOwner,
}));

vi.mock("@/hooks/use-animal-species", () => ({
  useAnimalSpecies: vi.fn(() => ({
    activeSpecies: [{ id: 1, name: "犬" }],
    allSpecies: [{ id: 1, name: "犬" }],
  })),
}));

vi.mock("@/components/shared/ConfirmDialog/ConfirmDialog", () => ({
  ConfirmDialog: ({ open, onConfirm }: { open: boolean; onConfirm: () => void }) => {
    permissionMocks.confirmDelete = onConfirm;
    return open ? <button onClick={onConfirm}>確認削除</button> : null;
  },
}));

const CLINIC_ID = "1";

function makeAuthContext() {
  const user: AuthUser = {
    id: "10",
    email: "staff@example.com",
    displayName: "テストスタッフ",
    isSystemAdmin: false,
    mainClinicId: CLINIC_ID,
    clinic: null,
    clinics: [{ clinicId: CLINIC_ID, clinicName: "本院", isMain: true }],
    permissions: {},
  };

  return {
    user,
    currentClinicId: CLINIC_ID,
    isAuthenticated: true,
    isLoading: false,
    login: async () => {},
    logout: async () => {},
    switchClinic: () => {},
    hasPermission: (_resource: string, _action: ResourceAction) => true,
    refreshPermissions: async () => {},
  };
}

function makePet(): Pet {
  return {
    id: "pet-1",
    ownerId: "owner-1",
    ownerName: "山田太郎",
    ownerNumber: 1,
    name: "ポチ",
    species: "犬",
    clinicId: CLINIC_ID,
  } as unknown as Pet;
}

function DeleteRevocationHarness() {
  const [confirmAfterRender, setConfirmAfterRender] = useState(false);

  useLayoutEffect(() => {
    permissionMocks.revokeDeleteAndConfirm = () => {
      permissionMocks.canDelete = false;
      setConfirmAfterRender(true);
    };
  }, []);

  useLayoutEffect(() => {
    if (confirmAfterRender) {
      permissionMocks.confirmDelete?.();
    }
  }, [confirmAfterRender]);

  return <OwnersList />;
}

function renderOwnersList() {
  const router = createMemoryRouter(
    [
      {
        path: "/owners",
        element: <DeleteRevocationHarness />,
        loader: () => ({ pets: [makePet()], page: 1, limit: 20, total: 1 }),
      },
    ],
    { initialEntries: ["/owners"] },
  );

  return render(
    <AuthContext.Provider value={makeAuthContext()}>
      <RouterProvider router={router} />
    </AuthContext.Provider>,
  );
}

beforeEach(() => {
  permissionMocks.canDelete = true;
  permissionMocks.confirmDelete = undefined;
  permissionMocks.deleteOwner.mockReset();
  permissionMocks.revokeDeleteAndConfirm = undefined;
});

describe("OwnersList mutation permission boundary", () => {
  it("削除確認中にdelete権限を失い同じcommitのlayout phaseで確定してもdelete APIを発行しない", async () => {
    const user = userEvent.setup();
    renderOwnersList();

    await user.click(
      await screen.findByRole("button", {
        name: /飼主.*山田太郎.*ペット.*ポチ.*操作/,
      }),
    );
    await user.click(await screen.findByRole("menuitem", { name: "削除" }));
    expect(screen.getByRole("button", { name: "確認削除" })).toBeInTheDocument();

    act(() => {
      permissionMocks.revokeDeleteAndConfirm?.();
    });

    expect(permissionMocks.deleteOwner).not.toHaveBeenCalled();
  });
});
