import { Suspense } from "react";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createMemoryRouter, RouterProvider, useLocation } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";

import type { Owner } from "@/types/owner";
import type { Pet } from "@/types";
import type { PetMutations } from "@/types/pet";
import { OwnerForm } from "@/features/owners";
import { ReceptionDetailModal, type ReceptionAppointment } from "@/features/reception";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canEdit: true,
    canCreate: true,
    canDelete: true,
    canView: true,
  }),
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { mainClinicId: "clinic-1", clinics: [{ id: "clinic-1", name: "本院" }] },
    currentClinicId: "clinic-1",
  }),
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPet: () => ({ data: undefined, isLoading: false, isError: false }),
}));

vi.mock("@/features/owners/hooks/use-animal-species", () => ({
  useAnimalSpecies: () => ({
    allSpecies: [{ id: 1, name: "犬" }],
    activeSpecies: [{ id: 1, name: "犬" }],
    isLoading: false,
  }),
}));

vi.mock("@/features/owners/api/get-insurances", () => ({
  useGetInsurances: () => ({ data: [], isLoading: false }),
}));

vi.mock("@/features/owners/api/get-owner-shared-pets", () => ({
  useGetOwnerSharedPets: () => ({
    data: { shared_pets: [] },
    isError: false,
    isLoading: false,
  }),
}));

vi.mock("@/features/owners/components/PetSubOwnersSection", () => ({
  PetSubOwnersSection: () => null,
}));

const RECEPTION_RETURN = "/?date=2026-05-29";

const pet = {
  id: "10",
  ownerId: "20",
  ownerName: "山田",
  petNumber: "P-10",
  name: "ポチ",
  species: "犬",
  status: "生存",
  gender: "雄",
} as unknown as Pet;

const owner = {
  id: "20",
  ownerName: "山田",
  ownerNameKana: "ヤマダ",
  company: "",
  postalCode: "",
  address1: "",
  address2: "",
  homePostalCode: "",
  homeAddress1: "",
  homeAddress2: "",
  phone: "",
  companyPhone: "",
  email: "",
  remarks: "",
  isDangerous: false,
  discountRate: 0,
  membershipType: "non_member",
  deliveryExcluded: false,
  deliveryCaution: false,
  isTransferred: false,
  lstepOptOut: false,
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
  pets: [pet],
} as Owner;

const petMutations: PetMutations = {
  createPetFn: vi.fn() as never,
  createPetMutate: vi.fn(),
  updatePetMutate: vi.fn(),
  deletePetMutate: vi.fn(),
  revokePetDeathMutate: vi.fn(),
};

const appointment = {
  id: "101",
  time: "09:45",
  visitDate: "2026-05-29",
  end: new Date(2026, 4, 29, 10, 15, 0),
  ownerName: "山田",
  petType: "犬",
  petName: "ポチ",
  visitType: "再診",
  reservationType: "一般診察",
  reservationTypeId: "1",
  reservationCategory: "general",
  isDesignated: false,
  doctor: "担当者A",
  doctorId: "33",
  petId: "10",
  ownerId: "20",
  status: "checked_in",
  notes: undefined,
  source: "manual",
} as ReceptionAppointment;

function LocationProbe() {
  const location = useLocation();
  const from = (location.state as { from?: string } | null)?.from ?? "";
  return (
    <>
      <output data-testid="here">{`${location.pathname}${location.search}`}</output>
      <output data-testid="from">{from}</output>
    </>
  );
}

function ownerElement(queryClient: QueryClient) {
  return (
    <QueryClientProvider client={queryClient}>
      <LocationProbe />
      <Suspense fallback={null}>
        <OwnerForm petMutations={petMutations} />
      </Suspense>
    </QueryClientProvider>
  );
}

describe("受付詳細からペット詳細へ進んで戻る", () => {
  it("ペット名から飼主詳細を開き、詳細を閉じたあと受付の日付へ戻る", async () => {
    const user = userEvent.setup();
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    const router = createMemoryRouter(
      [
        {
          path: "/",
          element: (
            <QueryClientProvider client={queryClient}>
              <LocationProbe />
              <ReceptionDetailModal
                isOpen
                onClose={() => {}}
                appointment={appointment}
                currentStatus="受付済"
                canCreateMedicalRecord
                canCreateAccounting
                canCreateHospitalization
                canViewVaccinations={false}
              />
            </QueryClientProvider>
          ),
        },
        { path: "/owners", element: <LocationProbe /> },
        {
          path: "/owners/:id",
          loader: () => ({ owner }),
          HydrateFallback: () => null,
          element: ownerElement(new QueryClient({ defaultOptions: { queries: { retry: false } } })),
        },
      ],
      { initialEntries: ["/"] },
    );
    render(<RouterProvider router={router} />);

    fireEvent.click(screen.getByRole("button", { name: /^ポチ$/ }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByTestId("here")).toHaveTextContent("/owners/20");
      expect(screen.getByTestId("here")).not.toHaveTextContent("pet=");
      expect(screen.getByTestId("from")).toHaveTextContent(RECEPTION_RETURN);
    });

    await user.keyboard("{Escape}");
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
    fireEvent.click(screen.getByRole("button", { name: "戻る" }));

    await waitFor(() => {
      expect(screen.getByTestId("here")).toHaveTextContent(RECEPTION_RETURN);
    });
  });

  it("存在しない pet を除いても受付の戻り先を残す", async () => {
    const router = createMemoryRouter(
      [
        { path: "/", element: <LocationProbe /> },
        { path: "/owners", element: <LocationProbe /> },
        {
          path: "/owners/:id",
          loader: () => ({ owner }),
          HydrateFallback: () => null,
          element: ownerElement(new QueryClient({ defaultOptions: { queries: { retry: false } } })),
        },
      ],
      {
        initialEntries: [
          {
            pathname: "/owners/20",
            search: "?pet=missing",
            state: { from: RECEPTION_RETURN },
          },
        ],
      },
    );
    render(<RouterProvider router={router} />);

    await waitFor(() => {
      expect(screen.getByTestId("here")).not.toHaveTextContent("pet=");
      expect(screen.getByTestId("from")).toHaveTextContent(RECEPTION_RETURN);
    });
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});
