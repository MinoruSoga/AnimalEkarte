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

const mockUseGetPets = vi.fn(() => ({ data: [] as Pet[] }));

vi.mock("@/hooks/use-pet", () => ({
  useGetPets: () => mockUseGetPets(),
}));

const katakanaOwnerPet = {
  id: "10",
  ownerId: "20",
  name: "ポチ",
  ownerName: "ヤマダハナコ",
  ownerNameKana: "ヤマダハナコ",
  petNameKana: "ポチ",
  species: "犬",
} as unknown as Pet;

const hiraganaOwnerPet = {
  id: "11",
  ownerId: "21",
  name: "たろう",
  ownerName: "さとうけんじ",
  ownerNameKana: "さとうけんじ",
  petNameKana: "たろう",
  species: "猫",
} as unknown as Pet;

describe("usePetSelectionPage", () => {
  it("ペット選択後も既存クエリと state を作成画面へ引き継ぐ", () => {
    mockUseGetPets.mockReturnValue({ data: [katakanaOwnerPet] });
    const { result } = renderHook(() =>
      usePetSelectionPage({
        selectPath: "/trimming/new",
        backPath: "/trimming",
      }),
    );

    act(() => {
      result.current.handleSelect(katakanaOwnerPet);
    });

    expect(navigate).toHaveBeenCalledWith(
      "/trimming/new?appointmentId=88&visitDate=2026-05-29&petId=10",
      { state: locationState },
    );
  });

  describe("かな正規化フィルタ", () => {
    const config = { selectPath: "/trimming/new", backPath: "/trimming" };
    const pets = [katakanaOwnerPet, hiraganaOwnerPet];

    function setup() {
      mockUseGetPets.mockReturnValue({ data: pets });
      return renderHook(() => usePetSelectionPage(config));
    }

    it("ひらがな入力でカタカナ ownerName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({ ...result.current.searchParams, ownerName: "やまだ" });
      });
      expect(result.current.filteredPets).toEqual([katakanaOwnerPet]);
    });

    it("カタカナ入力でひらがな ownerName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({ ...result.current.searchParams, ownerName: "サトウ" });
      });
      expect(result.current.filteredPets).toEqual([hiraganaOwnerPet]);
    });

    it("ひらがな入力でカタカナ petName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({ ...result.current.searchParams, petName: "ぽち" });
      });
      expect(result.current.filteredPets).toEqual([katakanaOwnerPet]);
    });

    it("カタカナ入力でひらがな petName がヒットする", () => {
      const { result } = setup();
      act(() => {
        result.current.setSearchParams({ ...result.current.searchParams, petName: "タロウ" });
      });
      expect(result.current.filteredPets).toEqual([hiraganaOwnerPet]);
    });
  });
});
