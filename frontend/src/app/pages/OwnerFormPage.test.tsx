import { Suspense } from "react";
import { render, screen } from "@testing-library/react";
import { createMemoryRouter, RouterProvider } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { OwnerFormPage } from "./OwnerFormPage";

const { mockUseGetOwner } = vi.hoisted(() => ({
  mockUseGetOwner: vi.fn(() => ({ data: undefined })),
}));

vi.mock("@/features/owners", () => ({
  OwnerForm: () => <div>owner form</div>,
  LineIntegrationCard: () => null,
  LineSendPanel: () => null,
  useGetOwner: mockUseGetOwner,
}));
vi.mock("@/features/pets", () => ({
  createPet: vi.fn(),
  useCreatePet: () => ({ mutate: vi.fn() }),
  useUpdatePet: () => ({ mutate: vi.fn() }),
  useDeletePet: () => ({ mutate: vi.fn() }),
}));
vi.mock("@/features/accounting", () => ({
  OwnerAccountingHistory: () => null,
}));
vi.mock("@/features/line-reservation", () => ({
  LinkedLineCustomers: () => null,
}));
vi.mock("@/hooks/use-revoke-pet-death", () => ({
  useRevokePetDeath: () => ({ mutate: vi.fn() }),
}));
vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: { mainClinicId: "clinic-1" } }),
}));

const OWNER = {
  id: "owner-1",
  ownerName: "合成飼主",
};

describe("OwnerFormPage loader/query dedup", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("編集routeではloaderのownerを再利用し、同一owner queryを重ねない", async () => {
    const router = createMemoryRouter([
      {
        path: "/owners/:id",
        loader: () => ({ owner: OWNER }),
        HydrateFallback: () => null,
        element: (
          <Suspense fallback={null}>
            <OwnerFormPage />
          </Suspense>
        ),
      },
    ], { initialEntries: ["/owners/owner-1"] });

    render(<RouterProvider router={router} />);

    expect(await screen.findByText("owner form")).toBeInTheDocument();
    expect(mockUseGetOwner).not.toHaveBeenCalled();
  });
});
