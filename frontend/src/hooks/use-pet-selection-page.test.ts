import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import type { Pet } from "@/types";

import { usePetSelectionPage } from "./use-pet-selection-page";

const navigate = vi.fn();
const locationState = {
  from: "/reservations",
  appointmentId: "88",
  visitDate: "2026-05-29",
};

vi.mock("react-router", () => ({
  useNavigate: () => navigate,
  useLocation: () => ({
    search: "?appointmentId=88&visitDate=2026-05-29",
    state: locationState,
  }),
}));

vi.mock("@/hooks/use-pet", () => ({
  useGetPets: () => ({ data: [] }),
}));

const pet = {
  id: "10",
  ownerId: "20",
  name: "ポチ",
  ownerName: "山田花子",
  species: "犬",
} as Pet;

describe("usePetSelectionPage", () => {
  it("ペット選択後も既存クエリと state を作成画面へ引き継ぐ", () => {
    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    act(() => {
      result.current.handleSelect(pet);
    });

    expect(navigate).toHaveBeenCalledWith(
      "/trimming/new?appointmentId=88&visitDate=2026-05-29&petId=10",
      { state: locationState },
    );
  });
});
